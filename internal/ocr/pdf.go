//go:build tesseract
// +build tesseract

package ocr

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// ConvertPDFToImages konvertiert ein PDF in Bilder mit Hilfe von ImageMagick
func ConvertPDFToImages(pdfPath, outputDir string) ([]string, error) {
	// Prüfen, ob ImageMagick installiert ist
	_, err := exec.LookPath("convert")
	if err != nil {
		return nil, errors.New("ImageMagick 'convert' nicht gefunden - bitte installieren Sie ImageMagick")
	}

	// Stellen Sie sicher, dass das Ausgabeverzeichnis existiert
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("fehler beim Erstellen des Ausgabeverzeichnisses: %v", err)
	}

	// Basisname des PDFs ohne Erweiterung
	baseName := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
	
	// Ausgabepfad-Muster
	outputPattern := filepath.Join(outputDir, fmt.Sprintf("%s-page-%%d.png", baseName))

	// ImageMagick-Befehl ausführen
	cmd := exec.Command("convert", "-density", "300", pdfPath, "-quality", "100", outputPattern)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("Fehler bei der PDF-Konvertierung: %v, Ausgabe: %s", err, string(output))
		return nil, fmt.Errorf("fehler bei der PDF-Konvertierung: %v", err)
	}

	// Überprüfen, welche Dateien erstellt wurden
	pattern := filepath.Join(outputDir, fmt.Sprintf("%s-page-*.png", baseName))
	imagePaths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Suchen der erzeugten Bilder: %v", err)
	}

	if len(imagePaths) == 0 {
		return nil, errors.New("keine Bilder wurden aus dem PDF erzeugt")
	}

	return imagePaths, nil
}

// ProcessPDF verarbeitet ein PDF-Dokument, indem es zunächst in Bilder umgewandelt wird
func ProcessPDF(pdfPath, tempDir, language string) (*OCRResult, error) {
	// PDF in Bilder umwandeln
	imagePaths, err := ConvertPDFToImages(pdfPath, tempDir)
	if err != nil {
		return nil, err
	}

	// Ergebnisse für alle Seiten zusammenführen
	var fullText strings.Builder
	var dates, amounts, suppliers []string

	// Jedes Bild mit OCR verarbeiten
	for i, imagePath := range imagePaths {
		logrus.Infof("Verarbeite PDF-Seite %d von %d: %s", i+1, len(imagePaths), imagePath)
		
		// OCR für die einzelne Seite durchführen
		pageResult, err := ProcessFile(imagePath, language)
		if err != nil {
			logrus.Errorf("Fehler bei der Verarbeitung von Seite %d: %v", i+1, err)
			continue
		}

		// Text dieser Seite hinzufügen
		if i > 0 {
			fullText.WriteString("\n\n--- Seite " + fmt.Sprintf("%d", i+1) + " ---\n\n")
		}
		fullText.WriteString(pageResult.FullText)

		// Extrahierte Daten sammeln
		if pageResult.PossibleDate != "" {
			dates = append(dates, pageResult.PossibleDate)
		}
		if pageResult.PossibleSum != "" {
			amounts = append(amounts, pageResult.PossibleSum)
		}
		if pageResult.Supplier != "" {
			suppliers = append(suppliers, pageResult.Supplier)
		}

		// Optional: Zwischenbilder löschen
		// os.Remove(imagePath)
	}

	// Kombiniertes Ergebnis erstellen
	result := &OCRResult{
		FullText: fullText.String(),
	}

	// Die erste gefundene Information nehmen (könnte erweitert werden zu intelligenterer Auswahl)
	if len(dates) > 0 {
		result.PossibleDate = dates[0]
	}
	if len(amounts) > 0 {
		result.PossibleSum = amounts[0]
	}
	if len(suppliers) > 0 {
		result.Supplier = suppliers[0]
	}

	return result, nil
}