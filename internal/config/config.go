package config

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// Config enthält die Anwendungskonfiguration
type Config struct {
	// Pfade
	AppDir         string
	DatabasePath   string
	UploadsDir     string

	// OCR-Optionen
	TesseractLang  string
	
	// UI-Optionen
	DefaultTheme   string
	WindowWidth    int
	WindowHeight   int
}

// DefaultConfig erstellt eine Standardkonfiguration
func DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatal("Fehler beim Ermitteln des Home-Verzeichnisses: ", err)
	}

	appDir := filepath.Join(homeDir, ".belegbuddy")
	
	// Erstelle App-Verzeichnisse falls nötig
	dirs := []string{
		appDir,
		filepath.Join(appDir, "uploads"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			logrus.Fatal("Fehler beim Erstellen des Verzeichnisses: ", err)
		}
	}

	return &Config{
		AppDir:        appDir,
		DatabasePath:  filepath.Join(appDir, "belegbuddy.db"),
		UploadsDir:    filepath.Join(appDir, "uploads"),
		TesseractLang: "deu",
		DefaultTheme:  "light",
		WindowWidth:   1024,
		WindowHeight:  768,
	}
}