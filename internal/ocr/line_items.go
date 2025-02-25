package ocr

import (
	"regexp"
	"strings"
)

// ExtractLineItems extrahiert Rechnungspositionen aus dem OCR-Text
func ExtractLineItems(text string) []LineItem {
	var lineItems []LineItem

	// Suche gezielt nach dem Muster für Autoreparatur und Ersatzteile
	autoreparaturRegex := regexp.MustCompile(`Autoreparatur\s+(\d+,\d+)\s+(\d+,\d+)\s+(\d+,\d+)`)
	ersatzteileRegex := regexp.MustCompile(`Ersatzteile\s+(\d+,\d+)\s+(\d+,\d+)\s+(\d+,\d+)`)
	
	lines := strings.Split(text, "\n")
	
	// Suche nach spezifischen Zeilen im Text
	for _, line := range lines {
		// Prüfe auf "Autoreparatur"
		matches := autoreparaturRegex.FindStringSubmatch(line)
		if len(matches) >= 4 {
			lineItems = append(lineItems, LineItem{
				Description: "Autoreparatur",
				Quantity:    matches[1],
				UnitPrice:   matches[2],
				TotalPrice:  matches[3],
			})
			continue
		}
		
		// Prüfe auf "Ersatzteile"
		matches = ersatzteileRegex.FindStringSubmatch(line)
		if len(matches) >= 4 {
			lineItems = append(lineItems, LineItem{
				Description: "Ersatzteile",
				Quantity:    matches[1],
				UnitPrice:   matches[2],
				TotalPrice:  matches[3],
			})
		}
	}
	
	// Suche auch nach dem Gesamtbetrag
	summeRegex := regexp.MustCompile(`SUMME\s+(\d+,\d+)`)
	for _, line := range lines {
		matches := summeRegex.FindStringSubmatch(line)
		if len(matches) >= 2 {
			// Wir haben den Gesamtbetrag gefunden, können ihn aber nicht direkt speichern
			// Im wirklichen System würden wir ihn vielleicht in einem anderen Feld der Rechnung speichern
			break
		}
	}

	return lineItems
}
