#\!/bin/bash

# Datei ansehen und nach dem Muster suchen
grep -n "lineItemRegex :=" "/Users/thilobarth/Developer/ClaudeCode/BelegBuddY/internal/ocr/claude_ocr.go"

# Zeile mit dem Regex ersetzen
sed -i '' '217s/`(?m)^\d+\\.\s+(.+?),\s+(\d+(?:\\.\d+)?)\s*×\s*(\d+(?:[,.]\d+)?)\s*€?\s*=\s*(\d+(?:[,.]\d+)?)\s*€?`/`(?m)^\d+\\.\s+(.+?)(?:,)?\s+(\d+(?:\\.\d+)?)\s*[×xX]\s*(\d+(?:[,.]\d+)?)\s*€?\s*=\s*(\d+(?:[,.]\d+)?)\s*€?`/' "/Users/thilobarth/Developer/ClaudeCode/BelegBuddY/internal/ocr/claude_ocr.go"

# Erfolg bestätigen
echo "Änderung durchgeführt\!"
