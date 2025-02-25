#\!/bin/bash

# Backup erstellen
cp "/Users/thilobarth/Developer/ClaudeCode/BelegBuddY/internal/ocr/claude_ocr.go" "/Users/thilobarth/Developer/ClaudeCode/BelegBuddY/internal/ocr/claude_ocr.go.bak2"

# Ersetze die alten Zeilen mit dem neuen Code
sed -i '' '/\/\/ Rechnungspositionen aus dem strukturierten Claude-Format extrahieren/,/return result, nil/c\
	// Versuche, JSON-Antwort zu parsen\
	var jsonResponse map[string]interface{}\
	jsonStartIdx := strings.Index(responseText, "{")\
	jsonEndIdx := strings.LastIndex(responseText, "}")\
	\
	if jsonStartIdx >= 0 && jsonEndIdx > jsonStartIdx {\
		// Extrahiere möglichen JSON-Teil\
		jsonText := responseText[jsonStartIdx:jsonEndIdx+1]\
		logrus.Infof("Mögliches JSON gefunden: %s", jsonText[:minInt(200, len(jsonText))])\
		\
		err := json.Unmarshal([]byte(jsonText), &jsonResponse)\
		if err == nil {\
			logrus.Info("Claude-Antwort erfolgreich als JSON geparst")\
			\
			// Extrahiere Rechnungspositionen aus JSON\
			if positionen, ok := jsonResponse["positionen"].([]interface{}); ok {\
				logrus.Infof("JSON enthält %d Rechnungspositionen", len(positionen))\
				\
				for i, pos := range positionen {\
					if posMap, ok := pos.(map[string]interface{}); ok {\
						beschreibung, _ := posMap["beschreibung"].(string)\
						menge, _ := posMap["menge"].(string)\
						einzelpreis, _ := posMap["einzelpreis"].(string)\
						gesamtpreis, _ := posMap["gesamtpreis"].(string)\
						\
						// Füge € hinzu, falls nicht vorhanden\
						if einzelpreis \!= "" && \!strings.Contains(einzelpreis, "€") {\
							einzelpreis += " €"\
						}\
						if gesamtpreis \!= "" && \!strings.Contains(gesamtpreis, "€") {\
							gesamtpreis += " €"\
						}\
						\
						item := LineItem{\
							Description: beschreibung,\
							Quantity:    menge,\
							UnitPrice:   einzelpreis,\
							TotalPrice:  gesamtpreis,\
						}\
						result.LineItems = append(result.LineItems, item)\
						logrus.Infof("Position %d: %s, %s x %s = %s", i+1, item.Description, item.Quantity, item.UnitPrice, item.TotalPrice)\
					}\
				}\
			}\
			\
			// Überschreibe auch die anderen Felder, falls vorhanden\
			if supplier, ok := jsonResponse["lieferant"].(string); ok && supplier \!= "" {\
				result.Supplier = supplier\
			}\
			\
			if datum, ok := jsonResponse["datum"].(string); ok && datum \!= "" {\
				result.PossibleDate = datum\
			}\
			\
			if betrag, ok := jsonResponse["gesamtbetrag"].(string); ok && betrag \!= "" {\
				result.PossibleSum = betrag\
			}\
			\
			logrus.Infof("Extrahierte Rechnungspositionen aus JSON: %d", len(result.LineItems))\
			return result, nil\
		} else {\
			logrus.Warnf("Fehler beim Parsen des JSON: %v", err)\
		}\
	}\
	\
	// Fallback: Alte Methode zur Extraktion verwenden\
	logrus.Warn("Kein gültiges JSON gefunden, verwende Fallback-Extraktion")\
	\
	// Rechnungspositionen aus dem strukturierten Claude-Format extrahieren\
	positionsRegex := regexp.MustCompile(`(?s)POSITIONEN:\\s*\\n(.*?)(?:\\n\\n|$)`)\
	positionsMatch := positionsRegex.FindStringSubmatch(responseText)\
	\
	if len(positionsMatch) > 1 {\
		// Einzelne Positionen suchen\
		lineItemsText := positionsMatch[1]\
		lineItemRegex := regexp.MustCompile(`(?m)^\\d+\\.\\s+(.+?)(?:,)?\\s+(\\d+(?:\\.\\d+)?)\\s*[×xX]\\s*(\\d+(?:[,.]\d+)?)\\s*€?\\s*=\\s*(\\d+(?:[,.]\d+)?)\\s*€?`)\
		lineItemMatches := lineItemRegex.FindAllStringSubmatch(lineItemsText, -1)\
		\
		for _, match := range lineItemMatches {\
			if len(match) >= 5 {\
				item := LineItem{\
					Description: match[1],\
					Quantity:    match[2],\
					UnitPrice:   match[3] + " €",\
					TotalPrice:  match[4] + " €",\
				}\
				result.LineItems = append(result.LineItems, item)\
			}\
		}\
		\
		logrus.Infof("Extrahierte Rechnungspositionen: %d", len(result.LineItems))\
	} else {\
		// Fallback zur allgemeinen Extraktion\
		result.LineItems = ExtractLineItems(responseText)\
	}\
\
	return result, nil\
' "/Users/thilobarth/Developer/ClaudeCode/BelegBuddY/internal/ocr/claude_ocr.go"

echo "Änderung durchgeführt\!"
