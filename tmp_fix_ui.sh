#\!/bin/bash

# Backup erstellen
cp "/Users/thilobarth/Developer/ClaudeCode/BelegBuddY/internal/app/ui.go" "/Users/thilobarth/Developer/ClaudeCode/BelegBuddY/internal/app/ui.go.bak"

# Die Zeile mit minInt einfügen, falls sie noch nicht existiert
grep -q "func minInt" "/Users/thilobarth/Developer/ClaudeCode/BelegBuddY/internal/app/ui.go" || 
sed -i '' '/^package app/a\
// minInt gibt das Minimum zweier Ganzzahlen zurück\
func minInt(a, b int) int {\
	if a < b {\
		return a\
	}\
	return b\
}\
' "/Users/thilobarth/Developer/ClaudeCode/BelegBuddY/internal/app/ui.go"

# Ersetze die FormContent-Definition
sed -i '' '/\/\/ Formular für extrahierte Daten/,/container.NewPadded(saveButton),/c\
		// Formular für extrahierte Daten\
		\
		// Tabelle für Rechnungspositionen\
		positionsLabel := widget.NewLabelWithStyle("Rechnungspositionen:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})\
		\
		// Tabelle für Positionen anlegen\
		positionsTable := widget.NewTable(\
			func() (int, int) {\
				// Zeilen (Positionen + Überschrift), Spalten (Beschreibung, Menge, Einzelpreis, Gesamtpreis)\
				return len(result.LineItems) + 1, 4\
			},\
			func() fyne.CanvasObject {\
				return widget.NewLabel("Template")\
			},\
			func(i widget.TableCellID, o fyne.CanvasObject) {\
				label := o.(*widget.Label)\
				label.TextStyle = fyne.TextStyle{}\
				\
				// Überschriften in der ersten Zeile\
				if i.Row == 0 {\
					label.TextStyle.Bold = true\
					switch i.Col {\
					case 0:\
						label.SetText("Beschreibung")\
					case 1:\
						label.SetText("Menge")\
					case 2:\
						label.SetText("Einzelpreis")\
					case 3:\
						label.SetText("Gesamtpreis")\
					}\
					return\
				}\
				\
				// Positionsdaten in den weiteren Zeilen\
				lineItem := result.LineItems[i.Row-1]\
				switch i.Col {\
				case 0:\
					label.SetText(lineItem.Description)\
				case 1:\
					label.SetText(lineItem.Quantity)\
				case 2:\
					label.SetText(lineItem.UnitPrice)\
				case 3:\
					label.SetText(lineItem.TotalPrice)\
				}\
			},\
		)\
		\
		// Tabellenhöhe anpassen\
		tableHeight := minInt(300, 30*(len(result.LineItems)+1))\
		positionsTable.SetMinSize(fyne.NewSize(500, float32(tableHeight)))\
		\
		formContent := container.NewVBox(\
			container.NewPadded(dateLabel),\
			container.NewPadded(dateEntry),\
			container.NewPadded(amountLabel),\
			container.NewPadded(amountEntry),\
			container.NewPadded(supplierLabel),\
			container.NewPadded(supplierEntry),\
			widget.NewSeparator(),\
			container.NewPadded(positionsLabel),\
			container.NewPadded(positionsTable),\
			widget.NewSeparator(),\
			container.NewPadded(saveButton),\
		)\
' "/Users/thilobarth/Developer/ClaudeCode/BelegBuddY/internal/app/ui.go"

echo "Änderung durchgeführt\!"
