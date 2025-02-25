//go:build !tesseract

package ocr

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ProcessFile ist eine Mock-Version der OCR-Verarbeitung für Systeme ohne Tesseract
func ProcessFile(filePath string, language string) (*OCRResult, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// Prüfe, ob das Dateiformat unterstützt wird
	supportedExts := map[string]bool{
		".pdf": true,
		".png": true,
		".jpg": true,
		".jpeg": true,
		".tiff": true,
		".tif": true,
	}

	if !supportedExts[ext] {
		return nil, fmt.Errorf("nicht unterstütztes Dateiformat: %s", ext)
	}
	
	// Mock-Daten je nach Dateityp
	var text string
	if ext == ".pdf" {
		text = generateMockPdfText(filePath)
	} else {
		text = generateMockImageText(filePath)
	}
	
	// Ergebnis mit extrahierten Daten erstellen
	result := &OCRResult{
		FullText:     text,
		PossibleDate: extractDate(text),
		PossibleSum:  extractAmount(text),
		Supplier:     extractSupplier(text),
	}
	
	logrus.Info("Mock OCR-Verarbeitung abgeschlossen für: ", filePath)
	return result, nil
}

// ProcessPDF ist eine Mock-Version der PDF-Verarbeitung
func ProcessPDF(pdfPath, tempDir, language string) (*OCRResult, error) {
	// Direkt die ProcessFile-Funktion aufrufen, da wir im Mock-Modus sind
	return ProcessFile(pdfPath, language)
}

// generateMockPdfText erzeugt Mock-OCR-Text für PDFs
func generateMockPdfText(filePath string) string {
	dateStr := time.Now().AddDate(0, -1, -15).Format("02.01.2006")
	
	text := fmt.Sprintf("Dies ist ein OCR-Platzhalter für die PDF-Datei: %s\n\n", filePath)
	text += "Beispiel GmbH\n"
	text += "Musterstraße 123\n"
	text += "12345 Musterstadt\n\n"
	text += "Rechnung Nr.: 2023-1234\n"
	text += fmt.Sprintf("Rechnungsdatum: %s\n", dateStr)
	text += "Lieferant: Beispiel GmbH\n\n"
	text += "Pos.  Bezeichnung                  Anzahl     Einzelpreis     Gesamtpreis\n"
	text += "-------------------------------------------------------------------------------\n"
	text += "1     Produkt A                      2         59,99 €         119,98 €\n"
	text += "2     Dienstleistung B               3         45,00 €         135,00 €\n"
	text += "3     Produkt C                      1         84,50 €          84,50 €\n"
	text += "-------------------------------------------------------------------------------\n"
	text += "                                                Nettobetrag:    339,48 €\n"
	text += "                                                MwSt. 19%:       64,50 €\n"
	text += "                                                -------------------------\n"
	text += "                                                Gesamtbetrag:   403,98 €\n\n"
	text += "Zahlbar bis: 15.04.2023\n"
	text += "Bankverbindung: DE12 3456 7890 1234 5678 90\n"
	
	return text
}

// generateMockImageText erzeugt Mock-OCR-Text für Bilder
func generateMockImageText(filePath string) string {
	dateStr := time.Now().AddDate(0, 0, -5).Format("02.01.2006")
	
	text := fmt.Sprintf("Dies ist ein OCR-Platzhalter für die Bilddatei: %s\n\n", filePath)
	text += "Lieferanten KG\n"
	text += "Beispielweg 42\n"
	text += "10115 Berlin\n\n"
	text += "Quittung Nr.: Q-5678\n"
	text += fmt.Sprintf("Datum: %s\n", dateStr)
	text += "Kunde: BelegBuddY-Nutzer\n\n"
	text += "Artikel                                      Preis\n"
	text += "---------------------------------------------------\n"
	text += "Büromaterial                               45,75 €\n"
	text += "Fachliteratur                              28,99 €\n"
	text += "---------------------------------------------------\n"
	text += "Gesamtbetrag                               74,74 €\n\n"
	text += "Vielen Dank für Ihren Einkauf!\n"
	
	return text
}