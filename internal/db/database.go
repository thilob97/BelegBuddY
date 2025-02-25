package db

import (
	"errors"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/belegbuddy/belegbuddy/internal/models"
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
	result := tx.Where("name = ?", supplierName).FirstOrCreate(&supplier, models.Supplier{
		Name: supplierName,
	})
	if result.Error != nil {
		tx.Rollback()
		logrus.Error("Fehler beim Erstellen/Finden des Lieferanten: ", result.Error)
		return result.Error
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
	// Entferne alle Zeichen außer Ziffern, Punkten und Kommas
	sanitized := regexp.MustCompile(`[^0-9.,]`).ReplaceAllString(amountStr, "")
	
	// Ersetze Komma durch Punkt für die Konvertierung
	sanitized = strings.Replace(sanitized, ",", ".", -1)
	
	// Wenn mehrere Punkte vorhanden sind, behalte nur den letzten
	parts := strings.Split(sanitized, ".")
	if len(parts) > 2 {
		sanitized = strings.Join(parts[:len(parts)-1], "") + "." + parts[len(parts)-1]
	}
	
	// Parse als float64
	return strconv.ParseFloat(sanitized, 64)
}