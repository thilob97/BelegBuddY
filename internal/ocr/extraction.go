package ocr

import (
	"regexp"
	"strings"
)

// extractDate versucht ein Datum im Text zu finden
func extractDate(text string) string {
	// Typische deutsche Datumsformate
	// Format: DD.MM.YYYY oder DD.MM.YY
	datePattern := regexp.MustCompile("\\b(0[1-9]|[12][0-9]|3[01])[.](0[1-9]|1[012])[.](19|20)?\\d\\d\\b")
	
	// Suche nach Datum mit Schlüsselwörtern in der Nähe
	dateLines := extractLinesWithKeywords(text, []string{"datum", "date", "vom", "ausstellungsdatum", "rechnungsdatum"})
	
	// Suche in den relevanten Zeilen
	for _, line := range dateLines {
		matches := datePattern.FindStringSubmatch(line)
		if len(matches) > 0 {
			return matches[0]
		}
	}
	
	// Falls nichts gefunden wurde, suche im gesamten Text
	matches := datePattern.FindStringSubmatch(text)
	if len(matches) > 0 {
		return matches[0]
	}
	
	return ""
}

// extractAmount versucht einen Geldbetrag im Text zu finden
func extractAmount(text string) string {
	// Spezifischer Summenausdruck für die gegebene Rechnung
	sumPattern := regexp.MustCompile("SUMME\\s+(\\d+[,\\.]\\d+)")
	matches := sumPattern.FindStringSubmatch(text)
	if len(matches) >= 2 {
		return matches[1]
	}
	
	// Typische deutsche Geldbetragsformate mit Komma
	// Matches: 123,45€, 123,45 €, EUR 123,45, 123.45 EUR usw.
	amountPattern := regexp.MustCompile("\\b(\\d{1,3}(?:\\.\\d{3})*|\\d+),\\d{2}\\s*(?:€|EUR|Euro)?|\\b(?:€|EUR|Euro)\\s*(\\d{1,3}(?:\\.\\d{3})*|\\d+),\\d{2}\\b")
	
	// Suche nach Beträgen mit Schlüsselwörtern in der Nähe
	amountLines := extractLinesWithKeywords(text, []string{"summe", "betrag", "total", "gesamtbetrag", "rechnungsbetrag", "zu zahlen"})
	
	// Suche in den relevanten Zeilen
	for _, line := range amountLines {
		matches := amountPattern.FindStringSubmatch(line)
		if len(matches) > 0 {
			for _, match := range matches {
				if match != "" {
					return match
				}
			}
		}
	}
	
	// Falls nichts gefunden wurde, suche im gesamten Text
	matches = amountPattern.FindStringSubmatch(text)
	if len(matches) > 0 {
		for _, match := range matches {
			if match != "" {
				return match
			}
		}
	}
	
	return ""
}

// extractSupplier versucht den Lieferantennamen zu extrahieren
func extractSupplier(text string) string {
	// Typische Positionen für Lieferantennamen
	supplierLines := extractLinesWithKeywords(text, []string{"absender", "lieferant", "von", "firma", "rechnungssteller", "verkäufer"})
	
	// Wenn Zeilen mit Schlüsselwörtern gefunden wurden, nimm die erste
	if len(supplierLines) > 0 {
		// Entferne die Schlüsselwörter, um nur den Namen zu extrahieren
		cleanLine := removeKeywords(supplierLines[0], []string{"absender", "lieferant", "von", "firma", "rechnungssteller", "verkäufer"})
		return strings.TrimSpace(cleanLine)
	}
	
	// Alternative: Nimm die ersten Zeilen des Dokuments, da dort oft der Absender steht
	lines := strings.Split(text, "\n")
	if len(lines) > 0 {
		// Rückgabe der ersten nicht-leeren Zeile
		for _, line := range lines[:min(5, len(lines))] {
			if len(strings.TrimSpace(line)) > 0 {
				return strings.TrimSpace(line)
			}
		}
	}
	
	return ""
}

// Hilfsfunktionen

// extractLinesWithKeywords findet Zeilen, die bestimmte Schlüsselwörter enthalten
func extractLinesWithKeywords(text string, keywords []string) []string {
	var matchedLines []string
	lines := strings.Split(text, "\n")
	
	for _, line := range lines {
		lowercaseLine := strings.ToLower(line)
		for _, keyword := range keywords {
			if strings.Contains(lowercaseLine, strings.ToLower(keyword)) {
				matchedLines = append(matchedLines, line)
				break
			}
		}
	}
	
	return matchedLines
}

// removeKeywords entfernt Schlüsselwörter aus einer Zeile
func removeKeywords(line string, keywords []string) string {
	lowercaseLine := strings.ToLower(line)
	result := line
	
	for _, keyword := range keywords {
		lowercaseKeyword := strings.ToLower(keyword)
		if idx := strings.Index(lowercaseLine, lowercaseKeyword); idx >= 0 {
			// Entferne das Schlüsselwort und alles davor
			if idx+len(keyword) < len(line) {
				result = line[idx+len(keyword):]
			} else {
				result = ""
			}
			
			// Entferne führende Sonderzeichen wie :, -, etc.
			result = strings.TrimLeft(result, " :-\t")
			break
		}
	}
	
	return result
}

// minInt gibt das Minimum zweier Ganzzahlen zurück
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
