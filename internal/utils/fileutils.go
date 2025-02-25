package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// CopyFile kopiert eine Datei von Quelle zu Ziel
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		logrus.Errorf("Fehler beim Öffnen der Quelldatei: %v", err)
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		logrus.Errorf("Fehler beim Erstellen der Zieldatei: %v", err)
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		logrus.Errorf("Fehler beim Kopieren der Datei: %v", err)
		return err
	}

	return nil
}

// HashFile erzeugt einen SHA-256-Hash der Datei
func HashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		logrus.Errorf("Fehler beim Öffnen der Datei für Hashing: %v", err)
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		logrus.Errorf("Fehler beim Hashing der Datei: %v", err)
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// EnsureDir stellt sicher, dass ein Verzeichnis existiert
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// GetFileExtension gibt die Dateiendung zurück
func GetFileExtension(path string) string {
	return filepath.Ext(path)
}

// IsImageFile prüft, ob es sich um eine Bilddatei handelt
func IsImageFile(path string) bool {
	ext := filepath.Ext(path)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".tiff", ".tif", ".bmp":
		return true
	default:
		return false
	}
}

// IsPDFFile prüft, ob es sich um eine PDF-Datei handelt
func IsPDFFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".pdf"
}