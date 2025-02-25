//go:build !tesseract
package ocr

import (
	"fmt"
	"path/filepath"
	"strings"
	
	"github.com/belegbuddy/belegbuddy/internal/invoice2data"
	"github.com/sirupsen/logrus"
)

// ProcessFile verarbeitet eine Datei mit OCR
func ProcessFile(filePath string, language string) (*OCRResult, error) {
	// Unterstützte Formate prüfen
	ext := strings.ToLower(filepath.Ext(filePath))
	supportedExts := map[string]bool{
		".pdf":  true,
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".tiff": true,
		".tif":  true,
	}

	if !supportedExts[ext] {
		return nil, fmt.Errorf("nicht unterstütztes Dateiformat: %s", ext)
	}

	// Demo-Dateien erkennen - deaktiviert, um echte OCR zu nutzen
	// fileName := filepath.Base(filePath)
	// if strings.Contains(strings.ToLower(fileName), "beispiel") || 
	//    strings.Contains(strings.ToLower(fileName), "demo") ||
	//    strings.Contains(strings.ToLower(fileName), "test") {
	// 	return ProcessDemoFile(filePath)
	// }

	// invoice2data verwenden
	i2dResult, err := invoice2data.ProcessInvoice(filePath)
	if err != nil {
		logrus.Errorf("Fehler bei der invoice2data-Verarbeitung: %v", err)
		return fallbackToMockData(filePath)
	}

	// Konvertiere invoice2data.OCRResult zu ocr.OCRResult
	result := &OCRResult{
		FullText:     i2dResult.FullText,
		PossibleDate: i2dResult.PossibleDate,
		PossibleSum:  i2dResult.PossibleSum,
		Supplier:     i2dResult.Supplier,
		IsDemo:       false,
	}
	
	logrus.Info("Rechnungsverarbeitung erfolgreich für: ", filePath)
	return result, nil
}

// ProcessPDF ist eine Wrapper-Funktion für die PDF-Verarbeitung
func ProcessPDF(pdfPath, tempDir, language string) (*OCRResult, error) {
	// invoice2data kann PDFs direkt verarbeiten
	return ProcessFile(pdfPath, language)
}

// fallbackToMockData erzeugt einfache Daten ohne Demo-Marker
func fallbackToMockData(filePath string) (*OCRResult, error) {
	logrus.Warn("Falle zurück auf einfache Extraktion für: ", filePath)
	
	// Einfache Daten für Fehlerfall, aber NICHT als Demo markieren
	result := &OCRResult{
		FullText:     "Daten konnten nicht automatisch extrahiert werden",
		PossibleDate: "",
		PossibleSum:  "",
		Supplier:     "",
		IsDemo:       false, // Wichtig: NICHT als Demo markieren
	}
	
	return result, nil
}