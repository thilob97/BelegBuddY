package views

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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
	// Dummy-Daten für die Vorschau
	// In der vollständigen Implementierung würden diese aus der Datenbank kommen
	dummyData := []InvoiceListItem{
		{ID: 1, Date: time.Now().AddDate(0, 0, -5), Supplier: "Beispiel GmbH", Amount: 120.50, Status: "Bezahlt"},
		{ID: 2, Date: time.Now().AddDate(0, 0, -12), Supplier: "Muster AG", Amount: 450.75, Status: "Offen"},
	}
	
	// Tabelle erstellen
	table := createInvoiceTable(dummyData)
	
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
	
	return container.NewBorder(
		container.NewPadded(headerContainer),
		nil, nil, nil,
		table,
	)
}

// createInvoiceTable erstellt eine Tabelle mit Rechnungsdaten
func createInvoiceTable(data []InvoiceListItem) *widget.Table {
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
					label.SetText("Bearbeiten")
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
	
	return table
}

// formatAmount formatiert einen Betrag als Währungsangabe
func formatAmount(amount float64) string {
	return fmt.Sprintf("%.2f €", amount)
}