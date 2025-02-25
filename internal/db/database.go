package db

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/belegbuddy/belegbuddy/internal/models"
	"github.com/belegbuddy/belegbuddy/internal/ocr"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB ist der Datenbankverbindungspool
var DB *gorm.DB

// Init initialisiert die Datenbankverbindung und erstellt die Tabellen
func Init(dbPath string) error {
	var err error
	
	logrus.Info("Verbinde mit Datenbank: ", dbPath)
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		logrus.Error("Fehler beim Verbinden zur Datenbank: ", err)
		return err
	}

	// Automatische Migrationsdurchführung für alle Modelle
	logrus.Info("Führe Datenbankmigrationen durch")
	err = DB.AutoMigrate(
		&models.Invoice{},
		&models.InvoiceItem{},
		&models.Supplier{},
		&models.FileReference{},
	)
	if err != nil {
		logrus.Error("Fehler bei der Datenbankaktualisierung: ", err)
		return err
	}

	logrus.Info("Datenbankinitialisierung erfolgreich")
	return nil
}

// GetDB gibt die aktuelle Datenbankverbindung zurück
func GetDB() *gorm.DB {
	return DB
}

// GetAllInvoices gibt alle Rechnungen aus der Datenbank zurück
func GetAllInvoices() ([]models.Invoice, error) {
	var invoices []models.Invoice
	result := DB.Preload("Supplier").Find(&invoices)
	if result.Error != nil {
		logrus.Error("Fehler beim Abrufen der Rechnungen: ", result.Error)
		return nil, result.Error
	}
	return invoices, nil
}

// GetInvoiceByID gibt eine Rechnung mit allen Details zurück
func GetInvoiceByID(id uint) (*models.Invoice, error) {
	var invoice models.Invoice
	result := DB.Preload("Supplier").
		Preload("InvoiceItems").
		Preload("FileRefs").
		First(&invoice, id)
	
	if result.Error != nil {
		logrus.Error("Fehler beim Abrufen der Rechnung mit ID ", id, ": ", result.Error)
		return nil, result.Error
	}
	
	return &invoice, nil
}

// SaveInvoice speichert eine erkannte Rechnung in der Datenbank
func SaveInvoice(date, amount, supplierName, fullText, filePath string) error {
	// Debug-Ausgaben
	logrus.WithFields(logrus.Fields{
		"date":         date,
		"amount":       amount,
		"supplierName": supplierName,
		"filePath":     filePath,
	}).Info("Speichere Rechnung")

	// OCR-Ergebnis für die Positionsextraktion
	ocrResult := ocr.OCRResult{
		FullText:     fullText,
		PossibleDate: date,
		PossibleSum:  amount,
		Supplier:     supplierName,
		LineItems:    ocr.ExtractLineItems(fullText),
	}
	// Validierung
	if date == "" || amount == "" || supplierName == "" {
		return errors.New("Datum, Betrag und Lieferant sind Pflichtfelder")
	}

	// Parse des Datums (flexible Formate zulassen)
	parsedDate, err := parseDate(date)
	if err != nil {
		logrus.Error("Fehler beim Parsen des Datums: ", err)
		return err
	}

	// Parse des Betrags
	parsedAmount, err := parseAmount(amount)
	if err != nil {
		logrus.Error("Fehler beim Parsen des Betrags: ", err)
		return err
	}

	// Transaktion starten
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Lieferanten finden oder erstellen
	var supplier models.Supplier
	dbResult := tx.Where("name = ?", supplierName).FirstOrCreate(&supplier, models.Supplier{
		Name: supplierName,
	})
	if dbResult.Error != nil {
		tx.Rollback()
		logrus.Error("Fehler beim Erstellen/Finden des Lieferanten: ", dbResult.Error)
		return dbResult.Error
	}

	// Rechnung erstellen
	invoice := models.Invoice{
		Date:        parsedDate,
		TotalAmount: parsedAmount,
		SupplierID:  supplier.ID,
		Currency:    "EUR", // Standard-Währung
		Status:      "Neu",
	}

	if err := tx.Create(&invoice).Error; err != nil {
		tx.Rollback()
		logrus.Error("Fehler beim Speichern der Rechnung: ", err)
		return err
	}

	// Konvertiere OCR LineItems in Datenbank InvoiceItems
	if len(ocrResult.LineItems) > 0 {
		logrus.Infof("Gefundene LineItems: %d", len(ocrResult.LineItems))
		for i, lineItem := range ocrResult.LineItems {
			logrus.Infof("Verarbeite Position %d: %s", i+1, lineItem.Description)
			
			// Parse Menge und Preise
			quantity := 1.0
			if lineItem.Quantity != "" {
				if q, err := parseAmount(lineItem.Quantity); err == nil {
					quantity = q
				} else {
					logrus.Warnf("Fehler beim Parsen der Menge '%s': %s", lineItem.Quantity, err)
				}
			}
			
			// Parse Einzelpreis
			unitPrice := 0.0
			if lineItem.UnitPrice != "" {
				if up, err := parseAmount(lineItem.UnitPrice); err == nil {
					unitPrice = up
				} else {
					logrus.Warnf("Fehler beim Parsen des Einzelpreises '%s': %s", lineItem.UnitPrice, err)
				}
			}
			
			// Parse Gesamtpreis
			totalPrice := 0.0
			if lineItem.TotalPrice != "" {
				if tp, err := parseAmount(lineItem.TotalPrice); err == nil {
					totalPrice = tp
				} else {
					logrus.Warnf("Fehler beim Parsen des Gesamtpreises '%s': %s", lineItem.TotalPrice, err)
				}
			}
			
			// Wenn kein Einzelpreis aber Menge und Gesamtpreis vorhanden, berechne Einzelpreis
			if unitPrice == 0.0 && quantity > 0 && totalPrice > 0 {
				unitPrice = totalPrice / quantity
			}
			
			// Wenn kein Gesamtpreis aber Einzelpreis und Menge vorhanden, berechne Gesamtpreis
			if totalPrice == 0.0 && unitPrice > 0 {
				totalPrice = unitPrice * quantity
			}
			
			// Erstelle Rechnungsposition
			invoiceItem := models.InvoiceItem{
				InvoiceID:   invoice.ID,
				Description: lineItem.Description,
				Quantity:    quantity,
				SinglePrice: unitPrice,
				TotalPrice:  totalPrice,
			}
			
			logrus.Infof("Speichere Position: %+v", invoiceItem)
			if err := tx.Create(&invoiceItem).Error; err != nil {
				tx.Rollback()
				logrus.Error("Fehler beim Speichern der Rechnungsposition: ", err)
				return err
			}
		}
	} else {
		logrus.Info("Keine LineItems gefunden, erstelle Standardposition")
		// Fallback: Erstelle eine Standardposition, wenn keine Positionen erkannt wurden
		invoiceItem := models.InvoiceItem{
			InvoiceID:   invoice.ID,
			Description: fmt.Sprintf("Artikel aus Rechnung vom %s", date),
			Quantity:    1.0,
			SinglePrice: parsedAmount,
			TotalPrice:  parsedAmount,
		}

		if err := tx.Create(&invoiceItem).Error; err != nil {
			tx.Rollback()
			logrus.Error("Fehler beim Speichern der Rechnungsposition: ", err)
			return err
		}
	}

	// Dateireferenz speichern, wenn ein Dateipfad angegeben ist
	if filePath != "" {
		fileRef := models.FileReference{
			InvoiceID:    invoice.ID,
			OriginalPath: filePath,
			Filename:     filepath.Base(filePath),
		}

		if err := tx.Create(&fileRef).Error; err != nil {
			tx.Rollback()
			logrus.Error("Fehler beim Speichern der Dateireferenz: ", err)
			return err
		}
	}

	// Transaktion abschließen
	if err := tx.Commit().Error; err != nil {
		logrus.Error("Fehler beim Commit der Transaktion: ", err)
		return err
	}

	logrus.Info("Rechnung erfolgreich in Datenbank gespeichert. ID: ", invoice.ID)
	return nil
}

// Hilfsfunktion zum Parsen von Datumsangaben
func parseDate(dateStr string) (time.Time, error) {
	// Versuche verschiedene Datumsformate
	formats := []string{
		"02.01.2006",
		"2006-01-02",
		"02/01/2006",
		"01/02/2006", // US Format
		"Jan 02, 2006",
		"2 Jan 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	// Wenn kein Format passt, gib Fehler zurück
	return time.Time{}, errors.New("kann Datumsformat nicht erkennen")
}

// Hilfsfunktion zum Parsen von Geldbeträgen
func parseAmount(amountStr string) (float64, error) {
	// Debug-Ausgabe für Fehleranalyse
	logrus.Debugf("parseAmount wird aufgerufen mit: '%s'", amountStr)
	
	// Entferne alle Zeichen außer Ziffern, Punkten und Kommas
	sanitized := regexp.MustCompile(`[^0-9.,]`).ReplaceAllString(amountStr, "")
	logrus.Debugf("Nach Entfernung von Sonderzeichen: '%s'", sanitized)
	
	// Ersetze Komma durch Punkt für die Konvertierung
	sanitized = strings.Replace(sanitized, ",", ".", -1)
	logrus.Debugf("Nach Ersetzung von Kommas: '%s'", sanitized)
	
	// Wenn der Wert leer ist, gib Fehler zurück
	if sanitized == "" {
		return 0.0, errors.New("leerer Betrag nach Bereinigung")
	}
	
	// Wenn mehrere Punkte vorhanden sind, behalte nur den letzten
	parts := strings.Split(sanitized, ".")
	if len(parts) > 2 {
		sanitized = strings.Join(parts[:len(parts)-1], "") + "." + parts[len(parts)-1]
		logrus.Debugf("Nach Korrektur mehrerer Punkte: '%s'", sanitized)
	}
	
	// Parse als float64
	result, err := strconv.ParseFloat(sanitized, 64)
	if err != nil {
		logrus.Warnf("Fehler beim Parsen des Betrags '%s': %s", sanitized, err)
		return 0.0, err
	}
	
	logrus.Debugf("Erfolgreich geparst: %.2f", result)
	return result, nil
}