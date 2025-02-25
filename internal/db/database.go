package db

import (
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