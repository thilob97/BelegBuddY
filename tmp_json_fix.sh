#\!/bin/bash

# Backup erstellen
cp "/Users/thilobarth/Developer/ClaudeCode/BelegBuddY/internal/ocr/claude_ocr.go" "/Users/thilobarth/Developer/ClaudeCode/BelegBuddY/internal/ocr/claude_ocr.go.bak3"

# Code für Backtick-Entfernung hinzufügen
sed -i '' '/jsonText = strings.Replace(jsonText, "\\t", "", -1)/a\
			\
			// Entferne Code-Block-Markup, falls vorhanden\
			jsonText = strings.TrimPrefix(jsonText, "```json")\
			jsonText = strings.TrimPrefix(jsonText, "```")\
			jsonText = strings.TrimSuffix(jsonText, "```")' "/Users/thilobarth/Developer/ClaudeCode/BelegBuddY/internal/ocr/claude_ocr.go"

echo "Änderung durchgeführt\!"
