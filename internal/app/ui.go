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
	// OCR-Ergebnis-Dialog anzeigen
	textEntry := widget.NewMultiLineEntry()
	textEntry.SetText(result.FullText)
	textEntry.SetMinRowsVisible(10)
	
	// Extrahierte Daten anzeigen
	dateEntry := widget.NewEntry()
	dateEntry.SetText(result.PossibleDate)
	
	amountEntry := widget.NewEntry()
	amountEntry.SetText(result.PossibleSum)
	
	supplierEntry := widget.NewEntry()
	supplierEntry.SetText(result.Supplier)
	
	// Titel für den Dialog
	title := "OCR-Ergebnis"
	if result.IsDemo {
		title = "Demo-Beispiel Ergebnis"
	}
	
	// Info-Label
	infoLabel := widget.NewLabel("OCR-Verarbeitung abgeschlossen")
	if result.IsDemo {
		infoLabel.SetText("Demo-Beispiel: Dies sind künstlich generierte Daten für Testzwecke")
	}
	
	// Formular für extrahierte Daten
	extractedForm := container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewLabel("Datum:"),
			dateEntry,
			widget.NewLabel("Betrag:"),
			amountEntry,
			widget.NewLabel("Lieferant:"),
			supplierEntry,
		),
	)
	
	// Button zum Speichern
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
	
	// Vollständiges Layout
	content := container.NewVBox(
		infoLabel,
		container.NewAppTabs(
			container.NewTabItem("Extrahierte Daten", container.NewVBox(
				container.NewPadded(extractedForm),
				container.NewPadded(saveButton),
			)),
			container.NewTabItem("Erkannter Text", container.NewVBox(
				container.NewPadded(textEntry),
			)),
		),
	)
	
	// Dialog anzeigen
	dialog.ShowCustom(title, "Schließen", content, a.MainWindow)

	// Log-Eintrag
	logrus.Info("OCR-Ergebnis für Datei angezeigt: ", filePath)
}