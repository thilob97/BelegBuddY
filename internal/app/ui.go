package app

import (
	"github.com/belegbuddy/belegbuddy/internal/db"
	"github.com/belegbuddy/belegbuddy/internal/ocr"
	"github.com/belegbuddy/belegbuddy/ui/views"
	"github.com/sirupsen/logrus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

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
	)
	
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
			dialog.ShowInformation("Einstellungen", 
				"Einstellungsdialog - noch nicht implementiert", 
				a.MainWindow)
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
		err := db.SaveInvoice(
			dateEntry.Text,
			amountEntry.Text,
			supplierEntry.Text,
			textEntry.Text,
			filePath,
		)
		
		if err != nil {
			dialog.ShowError(err, a.MainWindow)
			return
		}
		
		dialog.ShowInformation("Erfolg", "Rechnung wurde erfolgreich in der Datenbank gespeichert", a.MainWindow)
	})
	saveButton.Importance = widget.HighImportance

	// Formular für extrahierte Daten
	formContent := container.NewVBox(
		container.NewPadded(dateLabel),
		container.NewPadded(dateEntry),
		container.NewPadded(amountLabel),
		container.NewPadded(amountEntry),
		container.NewPadded(supplierLabel),
		container.NewPadded(supplierEntry),
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