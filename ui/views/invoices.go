package views

import (
	"fmt"
	"time"

	"github.com/belegbuddy/belegbuddy/internal/db"
	"github.com/sirupsen/logrus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// InvoiceListItem stellt eine Rechnung in der Übersicht dar
type InvoiceListItem struct {
	ID          uint
	Date        time.Time
	Supplier    string
	Amount      float64
	Status      string
}

// NewInvoicesView erstellt die Rechnungsübersicht
func NewInvoicesView() fyne.CanvasObject {
	// Leerer Container für die Detailansicht
	detailContainer := container.NewVBox(
		widget.NewLabel("Wählen Sie eine Rechnung aus, um Details anzuzeigen"),
	)
	
	// Daten aus der Datenbank laden
	var tableData []InvoiceListItem
	invoices, err := db.GetAllInvoices()
	if err != nil {
		logrus.Error("Fehler beim Laden der Rechnungen: ", err)
		// Fallback zu Dummy-Daten
		tableData = []InvoiceListItem{
			{ID: 1, Date: time.Now().AddDate(0, 0, -5), Supplier: "Beispiel GmbH", Amount: 120.50, Status: "Bezahlt"},
			{ID: 2, Date: time.Now().AddDate(0, 0, -12), Supplier: "Muster AG", Amount: 450.75, Status: "Offen"},
		}
	} else {
		// Daten aus der Datenbank in das ListItem-Format konvertieren
		for _, invoice := range invoices {
			tableData = append(tableData, InvoiceListItem{
				ID:       invoice.ID,
				Date:     invoice.Date,
				Supplier: invoice.Supplier.Name,
				Amount:   invoice.TotalAmount,
				Status:   invoice.Status,
			})
		}
		
		if len(tableData) == 0 {
			// Keine Daten in der Datenbank, Dummy-Daten anzeigen
			tableData = []InvoiceListItem{
				{ID: 0, Date: time.Now(), Supplier: "Keine Rechnungen vorhanden", Amount: 0, Status: "-"},
			}
		}
	}
	
	// Tabelle erstellen
	table := createInvoiceTable(tableData, func(id uint) {
		// Diese Funktion wird aufgerufen, wenn eine Rechnung ausgewählt wird
		showInvoiceDetails(id, detailContainer)
	})
	
	// Suchfeld
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Suche nach Rechnungen...")
	
	// Filter-Buttons
	allButton := widget.NewButton("Alle", nil)
	allButton.Importance = widget.HighImportance
	
	openButton := widget.NewButton("Offen", nil)
	paidButton := widget.NewButton("Bezahlt", nil)
	
	filterContainer := container.NewHBox(
		allButton,
		openButton,
		paidButton,
	)
	
	// Aktions-Buttons
	addButton := widget.NewButtonWithIcon("Neu", theme.ContentAddIcon(), nil)
	exportButton := widget.NewButtonWithIcon("Exportieren", theme.DocumentSaveIcon(), nil)
	
	actionContainer := container.NewHBox(
		addButton,
		exportButton,
	)
	
	// Layout zusammenstellen
	headerContainer := container.NewBorder(
		nil, nil,
		filterContainer,
		actionContainer,
		searchEntry,
	)
	
	// Tabelle und Detailcontainer in Split-View
	split := container.NewHSplit(
		table,
		container.NewPadded(detailContainer),
	)
	split.Offset = 0.4 // 40% für die Tabelle, 60% für die Details
	
	return container.NewBorder(
		container.NewPadded(headerContainer),
		nil, nil, nil,
		split,
	)
}

// createInvoiceTable erstellt eine Tabelle mit Rechnungsdaten
func createInvoiceTable(data []InvoiceListItem, onSelect func(id uint)) *widget.Table {
	table := widget.NewTable(
		// Callback für Anzahl Zeilen/Spalten
		func() (int, int) {
			return len(data) + 1, 5 // +1 für die Kopfzeile
		},
		// Callback für Zellinhalt
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		// Callback für Zellinhalt-Update
		func(i widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			
			// Kopfzeile
			if i.Row == 0 {
				label.TextStyle = fyne.TextStyle{Bold: true}
				switch i.Col {
				case 0:
					label.SetText("Datum")
				case 1:
					label.SetText("Lieferant")
				case 2:
					label.SetText("Betrag")
				case 3:
					label.SetText("Status")
				case 4:
					label.SetText("Aktionen")
				}
				return
			}
			
			// Daten-Zeilen
			dataIndex := i.Row - 1
			if dataIndex < len(data) {
				item := data[dataIndex]
				switch i.Col {
				case 0:
					label.SetText(item.Date.Format("02.01.2006"))
				case 1:
					label.SetText(item.Supplier)
				case 2:
					label.SetText(formatAmount(item.Amount))
				case 3:
					label.SetText(item.Status)
				case 4:
					label.SetText("Details")
				}
			}
		},
	)
	
	// Spaltenbreiten
	table.SetColumnWidth(0, 120)
	table.SetColumnWidth(1, 200)
	table.SetColumnWidth(2, 120)
	table.SetColumnWidth(3, 120)
	table.SetColumnWidth(4, 120)
	
	// Zeilen auswählbar machen
	table.OnSelected = func(id widget.TableCellID) {
		// Ignoriere Klicks auf die Kopfzeile
		if id.Row == 0 {
			return
		}
		
		// Rufe die übergebene Funktion mit der Rechnungs-ID auf
		dataIndex := id.Row - 1
		if dataIndex < len(data) {
			invoiceID := data[dataIndex].ID
			onSelect(invoiceID)
		}
		
		// Setze die Auswahl zurück (optional, je nach gewünschtem Verhalten)
		table.UnselectAll()
	}
	
	return table
}

// formatAmount formatiert einen Betrag als Währungsangabe
func formatAmount(amount float64) string {
	return fmt.Sprintf("%.2f €", amount)
}

// showInvoiceDetails zeigt die Details einer Rechnung in einem Container an
func showInvoiceDetails(invoiceID uint, detailContainer *fyne.Container) {
	// Container leeren
	detailContainer.RemoveAll()
	
	if invoiceID == 0 {
		message := widget.NewLabelWithStyle(
			"Keine Rechnung ausgewählt",
			fyne.TextAlignCenter,
			fyne.TextStyle{Italic: true},
		)
		detailContainer.Add(container.NewCenter(message))
		detailContainer.Refresh()
		return
	}
	
	// Lade die Rechnungsdetails aus der Datenbank
	invoice, err := db.GetInvoiceByID(invoiceID)
	if err != nil {
		errorMsg := widget.NewLabelWithStyle(
			"Fehler beim Laden der Rechnungsdetails: "+err.Error(),
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		)
		detailContainer.Add(container.NewCenter(errorMsg))
		detailContainer.Refresh()
		return
	}
	
	// Hauptinfos anzeigen
	title := widget.NewLabelWithStyle(
		fmt.Sprintf("Rechnung #%d - %s", invoice.ID, invoice.Supplier.Name),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	
	// ---- Abschnitt: Allgemeine Informationen ----
	infoSection := container.NewVBox(
		widget.NewLabelWithStyle("Allgemeine Informationen", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
	)
	
	// Infokarte mit allgemeinen Daten
	infoGrid := container.NewGridWithColumns(2,
		widget.NewLabelWithStyle("Lieferant:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(invoice.Supplier.Name),
		
		widget.NewLabelWithStyle("Adresse:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(invoice.Supplier.Address),
		
		widget.NewLabelWithStyle("Rechnungsdatum:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(invoice.Date.Format("02.01.2006")),
		
		widget.NewLabelWithStyle("Fälligkeitsdatum:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(invoice.DueDate.Format("02.01.2006")),
	)
	
	// Betragsinfos
	amountsGrid := container.NewGridWithColumns(2,
		widget.NewLabelWithStyle("Gesamtbetrag:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(formatAmount(invoice.TotalAmount), fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		
		widget.NewLabelWithStyle("MwSt-Betrag:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(formatAmount(invoice.VatAmount), fyne.TextAlignTrailing, fyne.TextStyle{}),
		
		widget.NewLabelWithStyle("Status:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(invoice.Status),
		
		widget.NewLabelWithStyle("Kategorie:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(invoice.Category),
	)
	
	infoSection.Add(container.NewPadded(infoGrid))
	infoSection.Add(widget.NewSeparator())
	infoSection.Add(container.NewPadded(amountsGrid))
	
	// ---- Abschnitt: Rechnungspositionen ----
	positionsSection := container.NewVBox(
		widget.NewLabelWithStyle("Rechnungspositionen", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
	)
	
	// Rechnungspositionen als Tabelle
	if len(invoice.InvoiceItems) > 0 {
		itemsTable := widget.NewTable(
			// Anzahl Zeilen und Spalten
			func() (int, int) {
				return len(invoice.InvoiceItems) + 1, 4 // +1 für Kopfzeile
			},
			// Zellenelement
			func() fyne.CanvasObject {
				return widget.NewLabel("Template")
			},
			// Zellinhalt
			func(id widget.TableCellID, obj fyne.CanvasObject) {
				label := obj.(*widget.Label)
				
				// Kopfzeile
				if id.Row == 0 {
					label.TextStyle = fyne.TextStyle{Bold: true}
					label.Alignment = fyne.TextAlignCenter
					
					switch id.Col {
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
				
				// Inhalt
				item := invoice.InvoiceItems[id.Row-1]
				
				switch id.Col {
				case 0:
					label.SetText(item.Description)
					label.Alignment = fyne.TextAlignLeading
				case 1:
					label.SetText(fmt.Sprintf("%.2f", item.Quantity))
					label.Alignment = fyne.TextAlignTrailing
				case 2:
					label.SetText(formatAmount(item.SinglePrice))
					label.Alignment = fyne.TextAlignTrailing
				case 3:
					label.SetText(formatAmount(item.TotalPrice))
					label.Alignment = fyne.TextAlignTrailing
				}
			},
		)
		
		// Spaltenbreiten
		itemsTable.SetColumnWidth(0, 300)
		itemsTable.SetColumnWidth(1, 80)
		itemsTable.SetColumnWidth(2, 120)
		itemsTable.SetColumnWidth(3, 120)
		
		positionsSection.Add(container.NewPadded(itemsTable))
	} else {
		noItemsLabel := widget.NewLabelWithStyle(
			"Keine Rechnungspositionen vorhanden", 
			fyne.TextAlignCenter, 
			fyne.TextStyle{Italic: true},
		)
		positionsSection.Add(container.NewPadded(noItemsLabel))
	}
	
	// ---- Abschnitt: Dateien ----
	filesSection := container.NewVBox(
		widget.NewLabelWithStyle("Zugehörige Dateien", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
	)
	
	if len(invoice.FileRefs) > 0 {
		filesList := container.NewVBox()
		
		for _, fileRef := range invoice.FileRefs {
			fileItem := container.NewHBox(
				widget.NewIcon(theme.FileIcon()),
				widget.NewLabel(fileRef.Filename),
			)
			filesList.Add(fileItem)
		}
		
		filesSection.Add(container.NewPadded(filesList))
	} else {
		noFilesLabel := widget.NewLabelWithStyle(
			"Keine Dateien vorhanden", 
			fyne.TextAlignCenter, 
			fyne.TextStyle{Italic: true},
		)
		filesSection.Add(container.NewPadded(noFilesLabel))
	}
	
	// ---- Abschnitt: Aktionen ----
	actionsSection := container.NewVBox(
		widget.NewLabelWithStyle("Aktionen", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
	)
	
	// Aktions-Buttons
	editButton := widget.NewButtonWithIcon("Bearbeiten", theme.DocumentCreateIcon(), func() {
		logrus.Info("Rechnung mit ID ", invoice.ID, " soll bearbeitet werden")
	})
	editButton.Importance = widget.HighImportance
	
	deleteButton := widget.NewButtonWithIcon("Löschen", theme.DeleteIcon(), func() {
		dialog.ShowConfirm(
			"Rechnung löschen",
			"Möchten Sie die Rechnung wirklich löschen?",
			func(confirm bool) {
				if confirm {
					logrus.Info("Rechnung mit ID ", invoice.ID, " soll gelöscht werden")
				}
			},
			fyne.CurrentApp().Driver().AllWindows()[0],
		)
	})
	deleteButton.Importance = widget.MediumImportance
	
	actionsContainer := container.NewHBox(
		container.NewPadded(editButton),
		container.NewPadded(deleteButton),
	)
	
	actionsSection.Add(container.NewCenter(actionsContainer))
	
	// Zusammenstellen aller Komponenten
	scrollContainer := container.NewVScroll(container.NewVBox(
		container.NewPadded(title),
		widget.NewSeparator(),
		infoSection,
		widget.NewSeparator(),
		positionsSection,
		widget.NewSeparator(),
		filesSection,
		widget.NewSeparator(),
		actionsSection,
	))
	
	detailContainer.Add(scrollContainer)
	
	// Container aktualisieren
	detailContainer.Refresh()
}