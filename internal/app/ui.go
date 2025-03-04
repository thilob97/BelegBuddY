package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	
	"github.com/belegbuddy/belegbuddy/internal/db"
	"github.com/belegbuddy/belegbuddy/internal/ocr"
	"github.com/belegbuddy/belegbuddy/ui/views"
	"github.com/sirupsen/logrus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// minInt gibt das Minimum zweier Ganzzahlen zurück
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// buildUI erstellt die Hauptbenutzeroberfläche
func buildUI(a *App) {
	// Callbacks für Upload-View
	uploadCallbacks := views.UploadViewCallbacks{
		OnFileProcessed: func(filePath string, result *ocr.OCRResult) {
			showOCRResultDialog(a, result, filePath)
		},
	}

	// Tabs für verschiedene Ansichten
	tabs := container.NewAppTabs(
		container.NewTabItem("Upload", views.NewUploadView(a.MainWindow, a.Config.UploadsDir, uploadCallbacks)),
		container.NewTabItem("Rechnungen", views.NewInvoicesView()),
		container.NewTabItem("Dashboard", views.NewDashboardView()),
		container.NewTabItem("Einstellungen", createSettingsView(a)),
	)
	
	// Tab-Wechsel-Event hinzufügen
	tabs.OnChanged = func(tab *container.TabItem) {
		// Dashboard aktualisieren, wenn es ausgewählt wird
		if tab.Text == "Dashboard" {
			// Immer ein komplett neues Dashboard erstellen
			newDashboard := views.NewDashboardView()
			tab.Content = newDashboard
			
			// Explizit das UI aktualisieren
			tab.Content.Refresh()
			tabs.Refresh()
			
			logrus.Info("Dashboard wurde komplett neu erstellt")
		}
	}
	
	tabs.SetTabLocation(container.TabLocationTop)
	
	// Menü aufbauen
	a.MainWindow.SetMainMenu(buildMainMenu(a))
	
	// Content setzen
	a.MainWindow.SetContent(tabs)

	logrus.Info("UI wurde erfolgreich aufgebaut")
}

// buildMainMenu erstellt das Hauptmenü der Anwendung
func buildMainMenu(a *App) *fyne.MainMenu {
	// Datei-Menü
	fileMenu := fyne.NewMenu("Datei",
		fyne.NewMenuItem("Einstellungen", func() {
			showSettingsDialog(a)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Beenden", func() {
			a.MainWindow.Close()
		}),
	)
	
	// Hilfe-Menü
	helpMenu := fyne.NewMenu("Hilfe",
		fyne.NewMenuItem("Über", func() {
			dialog.ShowInformation("Über BelegBuddY", 
				"BelegBuddY v0.1\nEine Anwendung zur Rechnungsdigitalisierung", 
				a.MainWindow)
		}),
	)
	
	return fyne.NewMainMenu(
		fileMenu,
		helpMenu,
	)
}

// showOCRResultDialog zeigt einen Dialog mit den OCR-Ergebnissen an
func showOCRResultDialog(a *App, result *ocr.OCRResult, filePath string) {
	// Titel für den Dialog
	title := "OCR-Ergebnis"
	if result.IsDemo {
		title = "Demo-Beispiel Ergebnis"
	}

	// Info-Label mit Statustext
	infoLabel := widget.NewLabel("OCR-Verarbeitung abgeschlossen")
	if result.IsDemo {
		infoLabel.SetText("Demo-Beispiel: Dies sind künstlich generierte Daten für Testzwecke")
	}
	infoLabel.Alignment = fyne.TextAlignCenter
	infoLabel.TextStyle = fyne.TextStyle{Bold: true}
	
	// Eingabefelder für extrahierte Daten
	dateLabel := widget.NewLabelWithStyle("Datum:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	dateEntry := widget.NewEntry()
	dateEntry.SetText(result.PossibleDate)
	dateEntry.SetPlaceHolder("TT.MM.JJJJ")
	
	amountLabel := widget.NewLabelWithStyle("Betrag:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	amountEntry := widget.NewEntry()
	amountEntry.SetText(result.PossibleSum)
	amountEntry.SetPlaceHolder("0,00")
	
	supplierLabel := widget.NewLabelWithStyle("Lieferant:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	supplierEntry := widget.NewEntry()
	supplierEntry.SetText(result.Supplier)
	supplierEntry.SetPlaceHolder("Lieferantenname")
	
	// Textfeld für den erkannten Text
	textLabel := widget.NewLabelWithStyle("Erkannter Text:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	textEntry := widget.NewMultiLineEntry()
	textEntry.SetText(result.FullText)
	textEntry.SetMinRowsVisible(10)
	
	// Speichern-Button mit prominentem Styling
	saveButton := widget.NewButton("In Datenbank speichern", func() {
		err := db.SaveInvoiceFromOCR(
			result,
			filePath,
		)
		
		if err != nil {
			dialog.ShowError(err, a.MainWindow)
			return
		}
		
		dialog.ShowInformation("Erfolg", "Rechnung wurde erfolgreich in der Datenbank gespeichert", a.MainWindow)
		
		// Aktualisiere die Rechnungsliste, falls die Ansicht bereits geöffnet ist
		if views.RefreshInvoiceList != nil {
			views.RefreshInvoiceList()
		}
		
		// Aktualisiere auch das Dashboard, falls die Ansicht initialisiert wurde
		if views.RefreshDashboard != nil {
			views.RefreshDashboard()
		}
	})
	saveButton.Importance = widget.HighImportance

	// Tabelle für Rechnungspositionen
	positionsLabel := widget.NewLabelWithStyle("Rechnungspositionen:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	
	// Tabelle für Positionen anlegen
	positionsTable := widget.NewTable(
		func() (int, int) {
			// Zeilen (Positionen + Überschrift), Spalten (Beschreibung, Menge, Einzelpreis, Gesamtpreis)
			return len(result.LineItems) + 1, 4
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			label.TextStyle = fyne.TextStyle{}
			
			// Überschriften in der ersten Zeile
			if i.Row == 0 {
				label.TextStyle.Bold = true
				switch i.Col {
				case 0:
					label.SetText("Beschreibung")
				case 1:
					label.SetText("Menge")
				case 2:
					label.SetText("Einzelpreis")
				case 3:
					label.SetText("Gesamtpreis")
				}
				return
			}
			
			// Positionsdaten in den weiteren Zeilen
			lineItem := result.LineItems[i.Row-1]
			switch i.Col {
			case 0:
				label.SetText(lineItem.Description)
			case 1:
				label.SetText(lineItem.Quantity)
			case 2:
				label.SetText(lineItem.UnitPrice)
			case 3:
				label.SetText(lineItem.TotalPrice)
			}
		},
	)
	
	// Tabellengröße anpassen
	tableHeight := minInt(300, 30*(len(result.LineItems)+1))
	positionsTableContainer := container.NewVScroll(positionsTable)
	positionsTableContainer.SetMinSize(fyne.NewSize(500, float32(tableHeight)))
	
	// Formular für extrahierte Daten
	formContent := container.NewVBox(
		container.NewPadded(dateLabel),
		container.NewPadded(dateEntry),
		container.NewPadded(amountLabel),
		container.NewPadded(amountEntry),
		container.NewPadded(supplierLabel),
		container.NewPadded(supplierEntry),
		widget.NewSeparator(),
		container.NewPadded(positionsLabel),
		container.NewPadded(positionsTableContainer),
		widget.NewSeparator(),
		container.NewPadded(saveButton),
	)

	// Textbereich
	textContent := container.NewVBox(
		container.NewPadded(textLabel),
		container.NewPadded(textEntry),
	)
	
	// Tabs für verschiedene Ansichten
	tabs := container.NewAppTabs(
		container.NewTabItem("Rechnungsdaten", formContent),
		container.NewTabItem("Erkannter Text", textContent),
	)

	// Überschrift mit Info-Label
	header := container.NewVBox(
		widget.NewSeparator(),
		container.NewPadded(infoLabel),
		widget.NewSeparator(),
	)
	
	// Gesamtlayout
	content := container.NewBorder(
		header,
		nil,
		nil, 
		nil,
		tabs,
	)
	
	// Dialog anzeigen
	customDialog := dialog.NewCustom(title, "Schließen", content, a.MainWindow)
	customDialog.Resize(fyne.NewSize(600, 500))
	customDialog.Show()

	// Log-Eintrag
	logrus.Info("OCR-Ergebnis für Datei angezeigt: ", filePath)
}

// showSettingsDialog zeigt den Einstellungsdialog an
func showSettingsDialog(a *App) {
	// Claude API-Key Eingabefeld
	apiKeyLabel := widget.NewLabelWithStyle("Claude API-Schlüssel:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetPlaceHolder("sk-ant-api-...")
	if a.Config.ClaudeAPIKey != "" {
		apiKeyEntry.SetText(a.Config.ClaudeAPIKey)
	}
	
	// OCR-Provider Auswahl
	providerLabel := widget.NewLabelWithStyle("Standard OCR-Provider:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	providerSelect := widget.NewSelect([]string{"tesseract", "claude"}, nil)
	providerSelect.SetSelected(a.Config.OCRProvider)
	
	// OCR-Sprache für Tesseract
	languageLabel := widget.NewLabelWithStyle("OCR-Sprache (Tesseract):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	languageEntry := widget.NewEntry()
	languageEntry.SetText(a.Config.TesseractLang)
	languageEntry.SetPlaceHolder("deu")
	
	// Formular für Einstellungen
	formContent := container.NewVBox(
		container.NewPadded(apiKeyLabel),
		container.NewPadded(apiKeyEntry),
		widget.NewSeparator(),
		container.NewPadded(providerLabel),
		container.NewPadded(providerSelect),
		widget.NewSeparator(),
		container.NewPadded(languageLabel),
		container.NewPadded(languageEntry),
	)
	
	// Speichern-Button mit prominentem Styling
	saveButton := widget.NewButton("Einstellungen speichern", func() {
		// Werte speichern
		a.Config.ClaudeAPIKey = apiKeyEntry.Text
		a.Config.OCRProvider = providerSelect.Selected
		a.Config.TesseractLang = languageEntry.Text
		
		// Konfiguration in Datei speichern
		saveSettings(a)
		
		dialog.ShowInformation("Erfolg", "Einstellungen wurden gespeichert", a.MainWindow)
	})
	saveButton.Importance = widget.HighImportance
	
	// Dialog erstellen
	content := container.NewBorder(
		nil,
		container.NewPadded(saveButton),
		nil,
		nil,
		container.NewPadded(formContent),
	)
	
	// Dialog anzeigen
	customDialog := dialog.NewCustom("Einstellungen", "Abbrechen", content, a.MainWindow)
	customDialog.Resize(fyne.NewSize(400, 300))
	customDialog.Show()
}

// createSettingsView erstellt die Ansicht für Einstellungen
func createSettingsView(a *App) fyne.CanvasObject {
	// Claude API-Key Eingabefeld
	apiKeyLabel := widget.NewLabelWithStyle("Claude API-Schlüssel:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetPlaceHolder("sk-ant-api-...")
	if a.Config.ClaudeAPIKey != "" {
		apiKeyEntry.SetText(a.Config.ClaudeAPIKey)
	}
	
	// OCR-Provider Auswahl
	providerLabel := widget.NewLabelWithStyle("Standard OCR-Provider:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	providerSelect := widget.NewSelect([]string{"tesseract", "claude"}, nil)
	providerSelect.SetSelected(a.Config.OCRProvider)
	
	// OCR-Sprache für Tesseract
	languageLabel := widget.NewLabelWithStyle("OCR-Sprache (Tesseract):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	languageEntry := widget.NewEntry()
	languageEntry.SetText(a.Config.TesseractLang)
	languageEntry.SetPlaceHolder("deu")
	
	// Speichern-Button mit prominentem Styling
	saveButton := widget.NewButton("Einstellungen speichern", func() {
		// Werte speichern
		a.Config.ClaudeAPIKey = apiKeyEntry.Text
		a.Config.OCRProvider = providerSelect.Selected
		a.Config.TesseractLang = languageEntry.Text
		
		// Konfiguration in Datei speichern
		saveSettings(a)
		
		dialog.ShowInformation("Erfolg", "Einstellungen wurden gespeichert", a.MainWindow)
	})
	saveButton.Importance = widget.HighImportance
	
	// Formular für Einstellungen
	formContent := container.NewVBox(
		container.NewPadded(widget.NewLabel("API-Einstellungen")),
		container.NewPadded(apiKeyLabel),
		container.NewPadded(apiKeyEntry),
		widget.NewSeparator(),
		container.NewPadded(widget.NewLabel("OCR-Einstellungen")),
		container.NewPadded(providerLabel),
		container.NewPadded(providerSelect),
		container.NewPadded(languageLabel),
		container.NewPadded(languageEntry),
		widget.NewSeparator(),
		container.NewPadded(saveButton),
	)
	
	return container.NewScroll(container.NewPadded(formContent))
}

// saveSettings speichert die Einstellungen in einer JSON-Datei
func saveSettings(a *App) {
	// Config-Datei-Pfad
	configPath := filepath.Join(a.Config.AppDir, "config.json")
	
	// Als JSON speichern
	file, err := os.Create(configPath)
	if err != nil {
		logrus.Error("Fehler beim Erstellen der Konfigurationsdatei: ", err)
		return
	}
	defer file.Close()
	
	// Als JSON kodieren
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(a.Config); err != nil {
		logrus.Error("Fehler beim Speichern der Konfiguration: ", err)
		return
	}
	
	logrus.Info("Konfiguration wurde in Datei gespeichert: ", configPath)
}