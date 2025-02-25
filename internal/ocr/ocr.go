//go:build tesseract

package ocr

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/otiai10/gosseract/v2"
	"github.com/sirupsen/logrus"
)

// ProcessFile verarbeitet eine Datei mit OCR
func ProcessFile(filePath string, language string) (*OCRResult, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	supportedExts := map[string]bool{
		".pdf": true,
		".png": true,
		".jpg": true,
		".jpeg": true,
		".tiff": true,
		".tif": true,
	}

	if !supportedExts[ext] {
		return nil, errors.New("nicht unterstütztes Dateiformat: " + ext)
	}

	// Tesseract-Client initialisieren
	client := gosseract.NewClient()
	defer client.Close()

	// Sprache setzen
	if err := client.SetLanguage(language); err != nil {
		logrus.Error("Fehler beim Setzen der OCR-Sprache: ", err)
		return nil, err
	}

	// PDF-Handling
	if strings.ToLower(filepath.Ext(filePath)) == ".pdf" {
		// Temporäres Verzeichnis für die extrahierten Bilder
		tempDir := filepath.Join(filepath.Dir(filePath), "temp_ocr_images")
		
		// PDF verarbeiten
		result, err := ProcessPDF(filePath, tempDir, language)
		if err != nil {
			logrus.Error("Fehler bei der PDF-Verarbeitung: ", err)
			return nil, err
		}
		
		// Erfolgreiche PDF-Verarbeitung
		return result, nil
	}

	// Bild für OCR setzen
	if err := client.SetImage(filePath); err != nil {
		logrus.Error("Fehler beim Laden des Bildes: ", err)
		return nil, err
	}

	// OCR ausführen
	text, err := client.Text()
	if err != nil {
		logrus.Error("Fehler bei der OCR-Verarbeitung: ", err)
		return nil, err
	}

	logrus.Info("OCR-Verarbeitung erfolgreich durchgeführt, Textlänge: ", len(text))

	// Einfache Extraktion von Informationen (in einer vollständigen Implementierung 
	// würde hier eine komplexere Logik zum Erkennen von Datums- und Geldbeträgen stehen)
	result := &OCRResult{
		FullText: text,
		// Die folgenden Felder sollten mit intelligenten Parsing-Algorithmen gefüllt werden
		PossibleDate: extractDate(text),
		PossibleSum: extractAmount(text),
		Supplier: extractSupplier(text),
		LineItems: ExtractLineItems(text),
	}

	logrus.Info("OCR-Verarbeitung abgeschlossen für: ", filePath)
	return result, nil
}

