package ocr

import (
	"regexp"
	"strings"
	
	"github.com/sirupsen/logrus"
)

// intMin gibt das Minimum zweier Ganzzahlen zurück
func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ExtractLineItems extrahiert Rechnungspositionen aus dem OCR-Text
func ExtractLineItems(text string) []LineItem {
	var lineItems []LineItem

	// Ausgabe des zu analysierenden Textes für Debugging
	if len(text) > 0 {
		previewLen := intMin(100, len(text))
		logrus.Debugf("Extrahiere LineItems aus Text (erste %d Zeichen): %s", previewLen, text[:previewLen])
	} else {
		logrus.Warn("ExtractLineItems: Leerer Text übergeben")
		return lineItems // Leere Liste zurückgeben, wenn Text leer ist
	}

	// Suche gezielt nach dem Muster für Autoreparatur und Ersatzteile
	autoreparaturRegex := regexp.MustCompile(`Autoreparatur\s+(\d+[,\.]\d+)\s+(\d+[,\.]\d+)\s+(\d+[,\.]\d+)`)
	ersatzteileRegex := regexp.MustCompile(`Ersatzteile\s+(\d+[,\.]\d+)\s+(\d+[,\.]\d+)\s+(\d+[,\.]\d+)`)
	
	lines := strings.Split(text, "\n")
	logrus.Debugf("Text aufgeteilt in %d Zeilen", len(lines))
	
	foundItems := 0
	
	// Suche nach spezifischen Zeilen im Text
	for i, line := range lines {
		// Überspringe leere Zeilen
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		// Debug-Ausgabe für aktuelle Zeile
		logrus.Debugf("Analysiere Zeile %d: %s", i, line)
		
		// Prüfe auf "Autoreparatur"
		matches := autoreparaturRegex.FindStringSubmatch(line)
		if len(matches) >= 4 {
			logrus.Infof("Autoreparatur gefunden: %s %s %s", matches[1], matches[2], matches[3])
			lineItems = append(lineItems, LineItem{
				Description: "Autoreparatur",
				Quantity:    matches[1],
				UnitPrice:   matches[2],
				TotalPrice:  matches[3],
			})
			foundItems++
			continue
		}
		
		// Prüfe auf "Ersatzteile"
		matches = ersatzteileRegex.FindStringSubmatch(line)
		if len(matches) >= 4 {
			logrus.Infof("Ersatzteile gefunden: %s %s %s", matches[1], matches[2], matches[3])
			lineItems = append(lineItems, LineItem{
				Description: "Ersatzteile",
				Quantity:    matches[1],
				UnitPrice:   matches[2],
				TotalPrice:  matches[3],
			})
			foundItems++
		}
	}
	
	// Suche auch nach dem Gesamtbetrag
	summeRegex := regexp.MustCompile(`SUMME\s+(\d+[,\.]\d+)`)
	for _, line := range lines {
		matches := summeRegex.FindStringSubmatch(line)
		if len(matches) >= 2 {
			logrus.Infof("Gesamtbetrag gefunden: %s", matches[1])
			break
		}
	}
	
	// Wenn keine Items gefunden wurden, erstelle ein Fallback-Item
	if foundItems == 0 {
		logrus.Warn("Keine Rechnungspositionen gefunden, erstelle Standardposition")
		lineItems = append(lineItems, LineItem{
			Description: "Nicht zugeordnete Position",
			Quantity:    "1",
			UnitPrice:   "0",
			TotalPrice:  "0",
		})
	}

	logrus.Infof("Insgesamt %d Rechnungspositionen gefunden", len(lineItems))
	return lineItems
}
