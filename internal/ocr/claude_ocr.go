package ocr

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// ClaudeOCRResponse stellt die Antwort der Claude API dar
type ClaudeOCRResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

// ProcessWithClaude verarbeitet eine Datei mit der Claude API
func ProcessWithClaude(filePath, apiKey string) (*OCRResult, error) {
	if apiKey == "" {
		return nil, errors.New("Claude API-Schlüssel nicht konfiguriert")
	}

	// Dateiformat prüfen und ggf. konvertieren
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// Bei PDF-Dateien müssen wir diese zuerst in ein Bild konvertieren
	var imageFilePath string
	if ext == ".pdf" {
		// Temporäres Verzeichnis für die konvertierte Datei
		tempDir := os.TempDir()
		imageFilePath = filepath.Join(tempDir, "claude_temp_"+filepath.Base(filePath)+".png")
		
		// PDF zu Bild konvertieren mit ImageMagick (erste Seite)
		logrus.Info("Konvertiere PDF zu Bild für Claude API: ", filePath)
		cmd := exec.Command("convert", "-density", "300", filePath+"[0]", "-quality", "100", "-flatten", "-background", "white", imageFilePath)
		stderr := &strings.Builder{}
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("fehler bei der Konvertierung von PDF zu Bild: %v, stderr: %s", err, stderr.String())
		}
		
		// Prüfen ob die Datei existiert und eine vernünftige Größe hat
		fileInfo, err := os.Stat(imageFilePath)
		if err != nil {
			return nil, fmt.Errorf("konvertierte Bilddatei kann nicht gefunden werden: %v", err)
		}
		if fileInfo.Size() < 100 {
			return nil, fmt.Errorf("konvertierte Bilddatei scheint fehlerhaft zu sein (zu klein: %d bytes)", fileInfo.Size())
		}
		
		logrus.Infof("PDF erfolgreich in Bild konvertiert: %s (%d bytes)", imageFilePath, fileInfo.Size())
		
		// Statt PDF-Datei das Bild verwenden
		filePath = imageFilePath
		ext = ".png"
		
		// Cleanup am Ende
		defer os.Remove(imageFilePath)
	} else if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		return nil, fmt.Errorf("nicht unterstütztes Dateiformat für Claude: %s", ext)
	}

	// Datei öffnen
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Öffnen der Datei: %v", err)
	}
	defer file.Close()

	// Medientyp bestimmen
	mediaType := "image/png"
	if ext == ".jpg" || ext == ".jpeg" {
		mediaType = "image/jpeg"
	} else if ext == ".pdf" {
		mediaType = "application/pdf"
	}

	// JSON-orientierter Prompt für die Rechnungsextraktion
	prompt := "Du bist ein Rechnungsanalyse-API-Endpunkt, der Rechnungsdaten extrahiert und als JSON zurückgibt.\n" +
		"AUFGABE:\n" +
		"1. Analysiere die Rechnung im beigefügten Bild\n" +
		"2. Extrahiere die folgenden Informationen:\n" +
		"   - Lieferant/Absender (Name und Anschrift)\n" +
		"   - Rechnungsdatum\n" +
		"   - Rechnungsnummer\n" +
		"   - Gesamtbetrag (MwSt. inkl.)\n" +
		"   - Einzelne Rechnungspositionen mit Beschreibung, Menge, Einzelpreis und Gesamtpreis\n\n" +
		"WICHTIG: Gib die extrahierten Daten AUSSCHLIESSLICH im folgenden JSON-Format zurück:\n\n" +
		"{\n" +
		"  \"lieferant\": \"Name des Lieferanten\",\n" +
		"  \"datum\": \"TT.MM.JJJJ\",\n" +
		"  \"rechnungsnummer\": \"XXX\",\n" +
		"  \"gesamtbetrag\": \"XX,XX\",\n" +
		"  \"positionen\": [\n" +
		"    {\n" +
		"      \"beschreibung\": \"Artikelname\",\n" +
		"      \"menge\": \"X.XX\",\n" +
		"      \"einzelpreis\": \"XX,XX\",\n" +
		"      \"gesamtpreis\": \"XX,XX\"\n" +
		"    }\n" +
		"  ],\n" +
		"  \"rohtext\": \"Der erkannte vollständige Text der Rechnung\"\n" +
		"}\n\n" +
		"Wenn eine Information nicht gefunden werden konnte, setze sie auf einen leeren String oder leeres Array. " +
		"Achte auf korrekte JSON-Syntax. Gib AUSSCHLIESSLICH JSON zurück, keine Erklärungen, Einleitungen oder Ergänzungen."

	// Anfrage-JSON
	requestJSON := fmt.Sprintf(`{
		"model": "claude-3-haiku-20240307",
		"max_tokens": 2000,
		"messages": [
			{
				"role": "user",
				"content": [
					{
						"type": "text",
						"text": %q
					},
					{
						"type": "image",
						"source": {
							"type": "base64",
							"media_type": "%s",
							"data": ""
						}
					}
				]
			}
		]
	}`, prompt, mediaType)

	// Datei einlesen und als base64 kodieren
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Lesen der Datei: %v", err)
	}
	
	// Anfrage-JSON mit base64-Daten
	base64Data := base64.StdEncoding.EncodeToString(fileData)
	requestJSON = strings.Replace(requestJSON, `"data": ""`, fmt.Sprintf(`"data": %q`, base64Data), 1)
	
	// HTTP-Anfrage erstellen (direkt JSON senden, kein Multipart)
	request, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", strings.NewReader(requestJSON))
	if err != nil {
		return nil, fmt.Errorf("fehler beim Erstellen der HTTP-Anfrage: %v", err)
	}

	// Header setzen
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-api-key", apiKey)
	request.Header.Set("anthropic-version", "2023-06-01")

	// Anfrage senden
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Senden der Anfrage: %v", err)
	}
	defer response.Body.Close()

	// Antwort prüfen
	if response.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("API-Fehler (%d): %s", response.StatusCode, string(responseBody))
	}

	// Antwort verarbeiten
	var claudeResponse ClaudeOCRResponse
	if err := json.NewDecoder(response.Body).Decode(&claudeResponse); err != nil {
		return nil, fmt.Errorf("fehler beim Dekodieren der API-Antwort: %v", err)
	}

	// Prüfen, ob wir eine Antwort erhalten haben
	if len(claudeResponse.Content) == 0 || claudeResponse.Content[0].Text == "" {
		return nil, errors.New("keine Antwort von Claude erhalten")
	}

	// Text der Antwort extrahieren
	responseText := claudeResponse.Content[0].Text
	logrus.Infof("Claude API Antwort erhalten: %d Zeichen", len(responseText))
	logrus.Infof("Claude API Antwort Text: %s", responseText)

	// OCR-Ergebnis erstellen
	result := &OCRResult{
		FullText: responseText,
		IsDemo:   false,
	}

	// Datum extrahieren
	datePattern := regexp.MustCompile(`(?i)(?:rechnungsdatum|datum|date):\s*(\d{1,2}[.-]\d{1,2}[.-]\d{2,4})`)
	dateMatches := datePattern.FindStringSubmatch(responseText)
	if len(dateMatches) > 1 {
		result.PossibleDate = dateMatches[1]
	}

	// Betrag extrahieren
	amountPattern := regexp.MustCompile(`(?i)(?:gesamtbetrag|summe|total):\s*([0-9.]+,\d{2}|\d+[,.]\d{2})(?:\s*€|\s*EUR)?`)
	amountMatches := amountPattern.FindStringSubmatch(responseText)
	if len(amountMatches) > 1 {
		result.PossibleSum = amountMatches[1]
	}

	// Lieferant extrahieren
	supplierPattern := regexp.MustCompile(`(?i)(?:lieferant|absender|firma|company):\s*([^\n]+)`)
	supplierMatches := supplierPattern.FindStringSubmatch(responseText)
	if len(supplierMatches) > 1 {
		result.Supplier = supplierMatches[1]
	}

	// Versuche, JSON-Antwort zu parsen
	var jsonResponse map[string]interface{}
	
	// Suche nach dem JSON in der Claude-Antwort mit vorheriger Entfernung aller Markdown-Formatierungen
	cleanText := responseText
	
	// Entferne Markdown Code Block-Formatierungen
	cleanText = strings.ReplaceAll(cleanText, "```json", "")
	cleanText = strings.ReplaceAll(cleanText, "```", "")
	
	// Suche nach dem JSON - präziser matching für valides JSON pattern
	jsonRegExp := regexp.MustCompile(`(?s)\{[\s\S]*"lieferant"[\s\S]*"datum"[\s\S]*"positionen"[\s\S]*\}`)
	jsonMatch := jsonRegExp.FindString(cleanText)
	
	if jsonMatch != "" {
		// Extrahiere möglichen JSON-Teil
		jsonText := jsonMatch
		logrus.Infof("Mögliches JSON gefunden: %s", jsonText[:minInt(200, len(jsonText))])
		
		// Debugging-Ausgabe vor der Bereinigung
		logrus.Infof("Rohes JSON: %s", jsonText[:minInt(500, len(jsonText))])
		
		// Bereinige JSON für bessere Verarbeitung
		
		// 1. Problem: Deutsche Zahlenformate (z.B. 1.056,50 -> 1056.50)
		// Für mehrteilige Zahlen mit Punkt als Tausendertrennzeichen und Komma als Dezimaltrennzeichen
		germanNumberPattern := regexp.MustCompile(`"(gesamtbetrag|einzelpreis|gesamtpreis|menge)":\s*"(\d{1,3})\.(\d{3}),(\d{2})"`)
		jsonText = germanNumberPattern.ReplaceAllString(jsonText, `"$1": "$2$3.$4"`)
		
		// 2. Problem: Deutsche Zahlenformate ohne Tausendertrennzeichen (z.B. 56,50 -> 56.50)
		simpleGermanNumberPattern := regexp.MustCompile(`"(gesamtbetrag|einzelpreis|gesamtpreis|menge)":\s*"(\d+),(\d{2})"`)
		jsonText = simpleGermanNumberPattern.ReplaceAllString(jsonText, `"$1": "$2.$3"`)
		
		// 3. Problem: Amerikanische Zahlenformate (z.B. 1,400.00 -> 1400.00)
		usNumberPattern := regexp.MustCompile(`"(gesamtbetrag|einzelpreis|gesamtpreis|menge)":\s*"(\d{1,3}),(\d{3}\.\d{2})"`)
		jsonText = usNumberPattern.ReplaceAllString(jsonText, `"$1": "$2$3"`)
		
		// 4. Problem: Entferne abgeschnittene Inhalte am Ende
		if strings.Contains(jsonText, "...") {
			lastBraceIdx := strings.LastIndex(jsonText, "}")
			if lastBraceIdx > 0 {
				jsonText = jsonText[:lastBraceIdx+1]
			}
		}
		
		// 5. Versuche unvollständige JSONs zu reparieren
		if !strings.HasSuffix(jsonText, "}") {
			jsonText = jsonText + "}"
		}
		
		// Entferne Whitespace und Steuerzeichen, die JSON ungültig machen könnten
		jsonText = strings.Replace(jsonText, "\r", "", -1)
		jsonText = strings.Replace(jsonText, "\t", " ", -1)
		
		// Debugging-Ausgabe nach der Bereinigung
		logrus.Infof("Bereinigtes JSON: %s", jsonText[:minInt(500, len(jsonText))])
		
		// Erstelle einen Direktzugriffsmechanismus für die OCR-Ergebnisse, unabhängig vom JSON-Parsing
		// Dies fungiert als Sicherheitsnetz, falls das JSON-Parsing fehlschlägt
		captureJsonValues(jsonText, result)
		
		// Versuche das bereinigte JSON zu parsen
		err := json.Unmarshal([]byte(jsonText), &jsonResponse)
		if err == nil {
			logrus.Info("Claude-Antwort erfolgreich als JSON geparst")
			
			// Extrahiere Rechnungspositionen aus JSON
			if positionen, ok := jsonResponse["positionen"].([]interface{}); ok {
				logrus.Infof("JSON enthält %d Rechnungspositionen", len(positionen))
				
				// Leere vorherige Positionen, um Duplikate zu vermeiden
				result.LineItems = []LineItem{}
				
				for i, pos := range positionen {
					if posMap, ok := pos.(map[string]interface{}); ok {
						logrus.Infof("Verarbeite Position: %+v", posMap)
						
						// Extrahiere Werte mit robuster Typkonvertierung
						var beschreibung, menge, einzelpreis, gesamtpreis string
						
						if b, ok := posMap["beschreibung"]; ok {
							beschreibung = fmt.Sprintf("%v", b)
						}
						
						if m, ok := posMap["menge"]; ok {
							menge = fmt.Sprintf("%v", m)
						}
						
						if e, ok := posMap["einzelpreis"]; ok {
							einzelpreis = fmt.Sprintf("%v", e)
						}
						
						if g, ok := posMap["gesamtpreis"]; ok {
							gesamtpreis = fmt.Sprintf("%v", g)
						}
						
						// Füge € hinzu, falls nicht vorhanden
						if einzelpreis != "" && !strings.Contains(einzelpreis, "€") {
							einzelpreis += " €"
						}
						if gesamtpreis != "" && !strings.Contains(gesamtpreis, "€") {
							gesamtpreis += " €"
						}
						
						item := LineItem{
							Description: beschreibung,
							Quantity:    menge,
							UnitPrice:   einzelpreis,
							TotalPrice:  gesamtpreis,
						}
						result.LineItems = append(result.LineItems, item)
						logrus.Infof("Position %d: %s, %s x %s = %s", i+1, item.Description, item.Quantity, item.UnitPrice, item.TotalPrice)
					}
				}
			}
			
			// Überschreibe auch die anderen Felder, falls vorhanden
			if supplier, ok := jsonResponse["lieferant"].(string); ok && supplier != "" {
				result.Supplier = supplier
			}
			
			if datum, ok := jsonResponse["datum"].(string); ok && datum != "" {
				result.PossibleDate = datum
			}
			
			if betrag, ok := jsonResponse["gesamtbetrag"].(string); ok && betrag != "" {
				result.PossibleSum = betrag
				// € hinzufügen, falls nicht vorhanden
				if !strings.Contains(result.PossibleSum, "€") {
					result.PossibleSum += " €"
				}
			}
			
			if number, ok := jsonResponse["rechnungsnummer"].(string); ok && number != "" {
				// Speichere die Rechnungsnummer in einem Kommentar im Supplier-Feld, falls dieses nicht genutzt wird
				if result.Supplier == "" {
					result.Supplier = "Rechnungsnummer: " + number
				}
			}
			
			logrus.Infof("Extrahierte Rechnungspositionen aus JSON: %d", len(result.LineItems))
			
			// Validiere, dass wir tatsächlich Daten extrahiert haben
			if result.Supplier == "" && result.PossibleDate == "" && result.PossibleSum == "" && len(result.LineItems) == 0 {
				logrus.Warn("JSON wurde geparst, aber keine Daten extrahiert. Verwende Fallback.")
			} else {
				return result, nil
			}
		} else {
			logrus.Warnf("Fehler beim Parsen des JSON: %v", err)
		}
	}
	
	// Fallback: Alte Methode zur Extraktion verwenden
	logrus.Warn("Kein gültiges JSON gefunden, verwende Fallback-Extraktion")
	
	// Rechnungspositionen aus dem strukturierten Claude-Format extrahieren
	positionsRegex := regexp.MustCompile(`(?s)POSITIONEN:\s*\n(.*?)(?:\n\n|$)`)
	positionsMatch := positionsRegex.FindStringSubmatch(responseText)
	
	if len(positionsMatch) > 1 {
		// Einzelne Positionen suchen
		lineItemsText := positionsMatch[1]
		lineItemRegex := regexp.MustCompile(`(?m)^\d+\.\s+(.+?)(?:,)?\s+(\d+(?:\.\d+)?)\s*[×xX]\s*(\d+(?:[,.]\d+)?)\s*€?\s*=\s*(\d+(?:[,.]\d+)?)\s*€?`)
		lineItemMatches := lineItemRegex.FindAllStringSubmatch(lineItemsText, -1)
		
		for _, match := range lineItemMatches {
			if len(match) >= 5 {
				item := LineItem{
					Description: match[1],
					Quantity:    match[2],
					UnitPrice:   match[3] + " €",
					TotalPrice:  match[4] + " €",
				}
				result.LineItems = append(result.LineItems, item)
			}
		}
		
		logrus.Infof("Extrahierte Rechnungspositionen: %d", len(result.LineItems))
	} else {
		// Fallback zur allgemeinen Extraktion
		result.LineItems = ExtractLineItems(responseText)
	}

	return result, nil
}

// captureJsonValues extrahiert Werte direkt aus dem JSON-Text per Regex, als Fallback Methode
func captureJsonValues(jsonText string, result *OCRResult) {
	// Extrahiere Lieferant
	lieferantPattern := regexp.MustCompile(`"lieferant":\s*"([^"]+)"`)
	lieferantMatches := lieferantPattern.FindStringSubmatch(jsonText)
	if len(lieferantMatches) > 1 && lieferantMatches[1] != "" {
		result.Supplier = lieferantMatches[1]
		logrus.Infof("Direkt extrahierter Lieferant: %s", result.Supplier)
	}
	
	// Extrahiere Datum
	datumPattern := regexp.MustCompile(`"datum":\s*"([^"]+)"`)
	datumMatches := datumPattern.FindStringSubmatch(jsonText)
	if len(datumMatches) > 1 && datumMatches[1] != "" {
		result.PossibleDate = datumMatches[1]
		logrus.Infof("Direkt extrahiertes Datum: %s", result.PossibleDate)
	}
	
	// Extrahiere Gesamtbetrag
	betragPattern := regexp.MustCompile(`"gesamtbetrag":\s*"([^"]+)"`)
	betragMatches := betragPattern.FindStringSubmatch(jsonText)
	if len(betragMatches) > 1 && betragMatches[1] != "" {
		result.PossibleSum = betragMatches[1]
		if !strings.Contains(result.PossibleSum, "€") {
			result.PossibleSum += " €"
		}
		logrus.Infof("Direkt extrahierter Gesamtbetrag: %s", result.PossibleSum)
	}
	
	// Extrahiere einzelne Positionen
	positionenPattern := regexp.MustCompile(`\{\s*"beschreibung":\s*"([^"]+)",\s*"menge":\s*"([^"]+)",\s*"einzelpreis":\s*"([^"]+)",\s*"gesamtpreis":\s*"([^"]+)"\s*\}`)
	positionenMatches := positionenPattern.FindAllStringSubmatch(jsonText, -1)
	if len(positionenMatches) > 0 {
		// Erstelle Line Items aus gefundenen Positionen
		for i, match := range positionenMatches {
			if len(match) >= 5 {
				beschreibung := match[1]
				menge := match[2]
				einzelpreis := match[3]
				gesamtpreis := match[4]
				
				// Füge € hinzu, falls nicht vorhanden
				if einzelpreis != "" && !strings.Contains(einzelpreis, "€") {
					einzelpreis += " €"
				}
				if gesamtpreis != "" && !strings.Contains(gesamtpreis, "€") {
					gesamtpreis += " €"
				}
				
				item := LineItem{
					Description: beschreibung,
					Quantity:    menge,
					UnitPrice:   einzelpreis,
					TotalPrice:  gesamtpreis,
				}
				result.LineItems = append(result.LineItems, item)
				logrus.Infof("Direkt extrahierte Position %d: %s, %s x %s = %s", i+1, item.Description, item.Quantity, item.UnitPrice, item.TotalPrice)
			}
		}
	}
}