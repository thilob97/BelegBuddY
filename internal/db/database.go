package db

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/belegbuddy/belegbuddy/internal/models"
	"github.com/belegbuddy/belegbuddy/internal/ocr"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB ist die globale Datenbankverbindung
var DB *gorm.DB

// InitDB initialisiert die Datenbankverbindung
func InitDB(dbPath string) error {
	var err error

	// Logger für GORM konfigurieren
	gormLogger := logger.New(
		logrus.StandardLogger(),
		logger.Config{
			SlowThreshold: time.Second,     // Ab wann eine Abfrage als langsam gilt
			LogLevel:      logger.Info,     // Log-Level
			Colorful:      false,           // Keine Farben im CLI
		},
	)

	// SQLite-Datenbank öffnen
	logrus.Info("Öffne Datenbankverbindung zu: ", dbPath)
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("fehler beim Öffnen der Datenbank: %v", err)
	}

	// Auto-Migration für Datenbank-Schemen aktivieren
	logrus.Info("Führe Datenbank-Migrationen durch...")
	err = DB.AutoMigrate(&models.Invoice{}, &models.Supplier{}, &models.InvoiceItem{}, &models.FileReference{})
	if err != nil {
		return fmt.Errorf("fehler bei der Datenbank-Migration: %v", err)
	}

	logrus.Info("Datenbank erfolgreich initialisiert!")
	return nil
}

// GetAllInvoices lädt alle Rechnungen mit zugehörigen Daten aus der Datenbank
func GetAllInvoices() ([]models.Invoice, error) {
	var invoices []models.Invoice

	result := DB.
		Preload("Supplier").
		Preload("InvoiceItems").
		Preload("FileRefs").
		Order("date DESC").
		Find(&invoices)

	if result.Error != nil {
		return nil, result.Error
	}

	return invoices, nil
}

// GetDashboardData liefert Statistiken für das Dashboard
func GetDashboardData() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	
	// Alle Rechnungen laden mit allen Beziehungen
	var invoices []models.Invoice
	result := DB.Preload("Supplier").Preload("InvoiceItems").Find(&invoices)
	if result.Error != nil {
		return nil, result.Error
	}
	
	// Debug-Ausgabe der Rohwerte
	logrus.Infof("Gefundene Rechnungen für Dashboard: %d", len(invoices))
	
	// Gesamtanzahl der Rechnungen
	data["totalCount"] = len(invoices)
	
	// Gesamtsumme aller Rechnungen
	var totalAmount float64
	for _, invoice := range invoices {
		totalAmount += invoice.TotalAmount
	}
	data["totalAmount"] = totalAmount
	
	// Aktuelle Monatsstatistik
	currentMonth := time.Now().Month()
	currentYear := time.Now().Year()
	var monthlyAmount float64
	var monthlyCount int
	
	for _, invoice := range invoices {
		if invoice.Date.Month() == currentMonth && invoice.Date.Year() == currentYear {
			monthlyAmount += invoice.TotalAmount
			monthlyCount++
		}
	}
	
	data["currentMonthAmount"] = monthlyAmount
	data["currentMonthCount"] = monthlyCount
	
	// Durchschnittlicher Rechnungsbetrag
	var averageAmount float64
	if len(invoices) > 0 {
		averageAmount = totalAmount / float64(len(invoices))
	}
	data["averageAmount"] = averageAmount
	
	// Größte Rechnung
	var maxAmount float64
	var maxInvoice models.Invoice
	for _, invoice := range invoices {
		if invoice.TotalAmount > maxAmount {
			maxAmount = invoice.TotalAmount
			maxInvoice = invoice
		}
	}
	data["maxAmount"] = maxAmount
	if maxAmount > 0 {
		data["maxInvoiceSupplier"] = maxInvoice.Supplier.Name
		data["maxInvoiceDate"] = maxInvoice.Date
	}
	
	// Top 5 Lieferanten nach Gesamtbetrag
	type SupplierSummary struct {
		Name   string
		Amount float64
		Count  int
	}
	
	supplierMap := make(map[uint]*SupplierSummary)
	
	for _, invoice := range invoices {
		if _, ok := supplierMap[invoice.SupplierID]; !ok {
			supplierMap[invoice.SupplierID] = &SupplierSummary{
				Name: invoice.Supplier.Name,
			}
		}
		supplierMap[invoice.SupplierID].Amount += invoice.TotalAmount
		supplierMap[invoice.SupplierID].Count++
	}
	
	var topSuppliers []SupplierSummary
	for _, summary := range supplierMap {
		topSuppliers = append(topSuppliers, *summary)
	}
	
	// Nach Betrag sortieren
	sort.Slice(topSuppliers, func(i, j int) bool {
		return topSuppliers[i].Amount > topSuppliers[j].Amount
	})
	
	// Nur die Top 5 behalten
	if len(topSuppliers) > 5 {
		topSuppliers = topSuppliers[:5]
	}
	
	data["topSuppliers"] = topSuppliers
	
	// Monatliche Statistiken für die letzten 12 Monate
	monthlyStats := make(map[string]float64)
	now := time.Now()
	
	// Iteriere über die letzten 12 Monate
	for i := 0; i < 12; i++ {
		targetMonth := now.AddDate(0, -i, 0)
		monthKey := targetMonth.Format("01/2006") // MM/YYYY
		monthlyStats[monthKey] = 0
	}
	
	// Summiere die Beträge pro Monat
	for _, invoice := range invoices {
		monthKey := invoice.Date.Format("01/2006")
		if _, exists := monthlyStats[monthKey]; exists {
			monthlyStats[monthKey] += invoice.TotalAmount
		}
	}
	
	data["monthlyStats"] = monthlyStats
	
	return data, nil
}

// GetInvoiceByID lädt eine Rechnung anhand ihrer ID
func GetInvoiceByID(id uint) (*models.Invoice, error) {
	var invoice models.Invoice

	result := DB.
		Preload("Supplier").
		Preload("InvoiceItems").
		Preload("FileRefs").
		First(&invoice, id)

	if result.Error != nil {
		return nil, result.Error
	}

	return &invoice, nil
}

// UpdateInvoice aktualisiert eine bestehende Rechnung
func UpdateInvoice(invoice *models.Invoice) error {
	// Transaktion starten
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Suche existierende Rechnung
	var existingInvoice models.Invoice
	if err := tx.First(&existingInvoice, invoice.ID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Aktualisiere die Lieferanten-Daten, falls sie geändert wurden
	var supplier models.Supplier
	if err := tx.First(&supplier, invoice.Supplier.ID).Error; err == nil {
		supplier.Name = invoice.Supplier.Name
		supplier.Address = invoice.Supplier.Address
		if err := tx.Save(&supplier).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// Lieferant existiert nicht, erstelle einen neuen
		supplier = invoice.Supplier
		supplier.ID = 0 // Setze ID zurück, damit ein neuer Eintrag erstellt wird
		if err := tx.Create(&supplier).Error; err != nil {
			tx.Rollback()
			return err
		}
		
		// Setze die neue Lieferanten-ID für die Rechnung
		invoice.SupplierID = supplier.ID
	}

	// Lösche bestehende Rechnungspositionen
	if err := tx.Where("invoice_id = ?", invoice.ID).Delete(&models.InvoiceItem{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Speichere aktualisierte Rechnung
	if err := tx.Save(invoice).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Speichere Rechnungspositionen neu
	for i := range invoice.InvoiceItems {
		// Stelle sicher, dass die Rechungs-ID korrekt gesetzt ist
		invoice.InvoiceItems[i].InvoiceID = invoice.ID
		// Reset PrimaryKey für Neuanlage
		invoice.InvoiceItems[i].ID = 0
		
		if err := tx.Create(&invoice.InvoiceItems[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Transaktion abschließen
	return tx.Commit().Error
}

// DeleteInvoice löscht eine Rechnung und alle zugehörigen Daten
func DeleteInvoice(id uint) error {
	// Transaktion starten
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Lösche Rechnungspositionen
	if err := tx.Where("invoice_id = ?", id).Delete(&models.InvoiceItem{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Lösche Dateireferenzen
	if err := tx.Where("invoice_id = ?", id).Delete(&models.FileReference{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Lösche die Rechnung selbst
	if err := tx.Delete(&models.Invoice{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Transaktion abschließen
	return tx.Commit().Error
}

// SaveInvoiceFromOCR speichert eine Rechnung basierend auf OCR-Ergebnissen
func SaveInvoiceFromOCR(ocrResult *ocr.OCRResult, filePath string) error {
	if ocrResult == nil {
		return errors.New("OCR-Ergebnis ist nil")
	}

	// Extrahiere die wichtigsten Informationen aus dem OCR-Ergebnis
	supplierName := ocrResult.Supplier
	if supplierName == "" {
		supplierName = "Unbekannter Lieferant"
	}

	date := ocrResult.PossibleDate
	if date == "" {
		date = time.Now().Format("02.01.2006")
	}

	amount := ocrResult.PossibleSum
	if amount == "" {
		amount = "0.00"
	}

	// Debug-Ausgaben
	logrus.Infof("Speichere Rechnung mit folgenden Daten:")
	logrus.Infof("Lieferant: %s", supplierName)
	logrus.Infof("Datum: %s", date)
	logrus.Infof("Betrag: %s", amount)

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
			
			// Erstelle Rechnungsposition mit Original-Strings aus OCR
			invoiceItem := models.InvoiceItem{
				InvoiceID:      invoice.ID,
				Description:    lineItem.Description,
				Quantity:       quantity,
				QuantityStr:    lineItem.Quantity,
				SinglePrice:    unitPrice,
				SinglePriceStr: lineItem.UnitPrice,
				TotalPrice:     totalPrice,
				TotalPriceStr:  lineItem.TotalPrice,
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
			InvoiceID:      invoice.ID,
			Description:    fmt.Sprintf("Artikel aus Rechnung vom %s", date),
			Quantity:       1.0,
			QuantityStr:    "1",
			SinglePrice:    parsedAmount,
			SinglePriceStr: amount,
			TotalPrice:     parsedAmount,
			TotalPriceStr:  amount,
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
	
	// Parse Float
	value, err := parseFloat(sanitized)
	if err != nil {
		return 0.0, err
	}
	
	logrus.Debugf("Parsed Amount: %.2f", value)
	return value, nil
}

// parseFloat ist eine Hilfsfunktion zum Parsen von Fließkommazahlen
func parseFloat(s string) (float64, error) {
	// Versuche direkt zu parsen
	value, err := parseStringToFloat(s)
	if err == nil {
		return value, nil
	}
	
	// Entferne weitere Formatierungen, z.B. bei 1.234.567,89
	s = cleanupNumberFormat(s)
	
	// Versuche erneut zu parsen
	return parseStringToFloat(s)
}

// parseStringToFloat konvertiert einen String in eine Gleitkommazahl
func parseStringToFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

// cleanupNumberFormat bereinigt Zahlenformate weiter
func cleanupNumberFormat(s string) string {
	// Zähle die Punkte
	dotCount := strings.Count(s, ".")
	
	// Wenn mehrere Punkte, entferne alle bis auf den letzten (dieser wird als Dezimaltrenner interpretiert)
	if dotCount > 1 {
		lastDotPos := strings.LastIndex(s, ".")
		s = strings.ReplaceAll(s[:lastDotPos], ".", "") + s[lastDotPos:]
	}
	
	return s
}