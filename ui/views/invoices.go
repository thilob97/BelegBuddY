package views

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/belegbuddy/belegbuddy/internal/db"
	"github.com/belegbuddy/belegbuddy/internal/models"
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
// Globale Variable für den Bereich, der die Liste der Rechnungen enthält
var invoiceListContainer *fyne.Container

// Globale Variable für den Detailbereich
var globalDetailContainer *fyne.Container

// Callback-Funktion zur Aktualisierung der Tabellenansicht
var onSelectInvoiceCallback func(id uint)

// RefreshInvoiceList ist eine exportierte Funktion, die von anderen Paketen aufgerufen werden kann
var RefreshInvoiceList func()

// refreshInvoiceList aktualisiert die Rechnungsliste 
func refreshInvoiceList(table *widget.Table) {
	// Wenn keine spezifische Tabelle angegeben ist, neu laden
	if invoiceListContainer != nil {
		// Bestehenden Inhalt entfernen
		invoiceListContainer.RemoveAll()
		
		// UI neu erzeugen
		invoices, err := db.GetAllInvoices()
		if err != nil {
			logrus.Error("Fehler beim Aktualisieren der Rechnungsliste: ", err)
			return
		}
		
		// Daten für die Tabelle vorbereiten
		var tableData []InvoiceListItem
		for _, invoice := range invoices {
			tableData = append(tableData, InvoiceListItem{
				ID:       invoice.ID,
				Date:     invoice.Date,
				Supplier: invoice.Supplier.Name,
				Amount:   invoice.TotalAmount,
				Status:   invoice.Status,
			})
		}
		
		// Neue Tabelle erstellen mit dem gespeicherten Callback
		newTable := createInvoiceTable(tableData, onSelectInvoiceCallback)
		invoiceListContainer.Add(newTable)
		
		// Container aktualisieren
		invoiceListContainer.Refresh()
		logrus.Info("Rechnungsliste aktualisiert, ", len(tableData), " Einträge gefunden")
	}
}

func NewInvoicesView() fyne.CanvasObject {
	// Leerer Container für die Detailansicht - volle Größe und zentriert
	detailContainer := container.NewMax(
		container.NewCenter(
			widget.NewLabelWithStyle("Wählen Sie eine Rechnung aus, um Details anzuzeigen", 
				fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		),
	)
	
	// Globale Variable setzen
	globalDetailContainer = detailContainer
	
	// Container für die Rechnungsliste initialisieren
	invoiceListContainer = container.NewMax()
	
	// Exportierte Aktualisierungsfunktion initialisieren
	RefreshInvoiceList = func() {
		refreshInvoiceList(nil)
	}
	
	// Daten aus der Datenbank laden
	var tableData []InvoiceListItem
	invoices, err := db.GetAllInvoices()
	if err != nil {
		logrus.Error("Fehler beim Laden der Rechnungen: ", err)
		// Keine künstlichen Daten mehr verwenden
		tableData = []InvoiceListItem{}
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
			logrus.Warn("Keine Rechnungen in der Datenbank gefunden")
		}
	}
	
	// Callback-Funktion speichern
	onSelectInvoiceCallback = func(id uint) {
		// Diese Funktion wird aufgerufen, wenn eine Rechnung ausgewählt wird
		showInvoiceDetails(id, detailContainer)
	}
	
	// Tabelle erstellen
	table := createInvoiceTable(tableData, onSelectInvoiceCallback)
	
	// Tabelle zum Container hinzufügen
	invoiceListContainer.Add(table)
	
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
	
	// Tabelle und Detailcontainer in Split-View - Detailansicht über die volle Höhe
	split := container.NewHSplit(
		invoiceListContainer,
		container.NewPadded(container.NewMax(detailContainer)),  // Max-Container nutzt volle Höhe
	)
	split.Offset = 0.35 // 35% für die Tabelle, 65% für die Details
	
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

// showInvoiceEditDialog zeigt einen Dialog zum Bearbeiten einer Rechnung an
func showInvoiceEditDialog(invoice *models.Invoice, detailContainer *fyne.Container) {
	// Formularelemente erstellen
	mainWindow := fyne.CurrentApp().Driver().AllWindows()[0]
	
	// Einfache Formularfelder für die wichtigsten Daten
	dateField := widget.NewEntry()
	dateField.SetText(invoice.Date.Format("02.01.2006"))
	
	supplierField := widget.NewEntry()
	supplierField.SetText(invoice.Supplier.Name)
	
	addressField := widget.NewEntry()
	addressField.SetText(invoice.Supplier.Address)
	addressField.MultiLine = true
	
	amountField := widget.NewEntry()
	amountField.SetText(fmt.Sprintf("%.2f", invoice.TotalAmount))
	
	statusOptions := []string{"Neu", "Offen", "Bezahlt", "Storniert"}
	statusSelect := widget.NewSelect(statusOptions, nil)
	statusSelect.Selected = invoice.Status
	
	categoryField := widget.NewEntry()
	categoryField.SetText(invoice.Category)
	
	// Rechnungspositionen bearbeitbar machen
	lineItemsContainer := container.NewVBox()
	
	// Überschrift für Rechnungspositionen
	lineItemsContainer.Add(widget.NewLabelWithStyle("Rechnungspositionen", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
	lineItemsContainer.Add(widget.NewSeparator())
	
	// Ein Array für die Eingabefelder der Rechnungspositionen
	var lineItemFields []*struct {
		Description *widget.Entry
		Quantity    *widget.Entry
		Price       *widget.Entry
		TotalPrice  *widget.Entry
	}
	
	// Bestehende Positionen anzeigen
	for _, item := range invoice.InvoiceItems {
		desc := widget.NewEntry()
		desc.SetText(item.Description)
		
		qty := widget.NewEntry()
		qty.SetText(fmt.Sprintf("%.2f", item.Quantity))
		
		price := widget.NewEntry()
		price.SetText(fmt.Sprintf("%.2f", item.SinglePrice))
		
		total := widget.NewEntry()
		total.SetText(fmt.Sprintf("%.2f", item.TotalPrice))
		
		// In einem Line-Container anordnen
		row := container.NewGridWithColumns(4, desc, qty, price, total)
		lineItemsContainer.Add(row)
		
		// Eingabefelder speichern
		lineItemFields = append(lineItemFields, &struct {
			Description *widget.Entry
			Quantity    *widget.Entry
			Price       *widget.Entry
			TotalPrice  *widget.Entry
		}{
			Description: desc,
			Quantity:    qty,
			Price:       price,
			TotalPrice:  total,
		})
	}
	
	// Scrollbare Ansicht der Positionen
	lineItemsScroll := container.NewVScroll(lineItemsContainer)
	lineItemsScroll.SetMinSize(fyne.NewSize(600, 200))
	
	// Formular erstellen
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Datum", Widget: dateField},
			{Text: "Lieferant", Widget: supplierField},
			{Text: "Adresse", Widget: addressField},
			{Text: "Gesamtbetrag", Widget: amountField},
			{Text: "Status", Widget: statusSelect},
			{Text: "Kategorie", Widget: categoryField},
			{Text: "Positionen", Widget: lineItemsScroll},
		},
		OnSubmit: func() {
			// Werte parsen und validieren
			parsedDate, err := time.Parse("02.01.2006", dateField.Text)
			if err != nil {
				dialog.ShowError(errors.New("Ungültiges Datumsformat. Bitte im Format TT.MM.JJJJ angeben."), mainWindow)
				return
			}
			
			parsedAmount, err := strconv.ParseFloat(strings.ReplaceAll(amountField.Text, ",", "."), 64)
			if err != nil {
				dialog.ShowError(errors.New("Ungültiger Betrag. Bitte eine Zahl eingeben."), mainWindow)
				return
			}
			
			// Rechnung aktualisieren
			invoice.Date = parsedDate
			invoice.TotalAmount = parsedAmount
			invoice.Status = statusSelect.Selected
			invoice.Category = categoryField.Text
			
			// Lieferant aktualisieren
			invoice.Supplier.Name = supplierField.Text
			invoice.Supplier.Address = addressField.Text
			
			// Rechnungspositionen aktualisieren
			var updatedLineItems []models.InvoiceItem
			
			for _, field := range lineItemFields {
				// Werte parsen
				qty, err := strconv.ParseFloat(strings.ReplaceAll(field.Quantity.Text, ",", "."), 64)
				if err != nil {
					dialog.ShowError(errors.New("Ungültige Menge. Bitte eine Zahl eingeben."), mainWindow)
					return
				}
				
				price, err := strconv.ParseFloat(strings.ReplaceAll(field.Price.Text, ",", "."), 64)
				if err != nil {
					dialog.ShowError(errors.New("Ungültiger Einzelpreis. Bitte eine Zahl eingeben."), mainWindow)
					return
				}
				
				total, err := strconv.ParseFloat(strings.ReplaceAll(field.TotalPrice.Text, ",", "."), 64)
				if err != nil {
					dialog.ShowError(errors.New("Ungültiger Gesamtpreis. Bitte eine Zahl eingeben."), mainWindow)
					return
				}
				
				// Position aktualisieren
				lineItem := models.InvoiceItem{
					InvoiceID:   invoice.ID,
					Description: field.Description.Text,
					Quantity:    qty,
					SinglePrice: price,
					TotalPrice:  total,
				}
				
				updatedLineItems = append(updatedLineItems, lineItem)
			}
			
			// Neue Line Items Liste zuweisen
			invoice.InvoiceItems = updatedLineItems
			
			// In der Datenbank speichern
			err = db.UpdateInvoice(invoice)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Fehler beim Speichern: %v", err), mainWindow)
				return
			}
			
			// Dialog schließen
			dialog.ShowInformation("Gespeichert", "Die Rechnung wurde erfolgreich aktualisiert.", mainWindow)
			
			// Tabellenansicht aktualisieren
			refreshInvoiceList(nil)
			
			// Details neu laden
			showInvoiceDetails(invoice.ID, detailContainer)
		},
		OnCancel: func() {
			// Nichts tun, Dialog wird geschlossen
		},
	}
	
	// Dialog anzeigen
	dialog.ShowCustom("Rechnung bearbeiten", "Abbrechen", form, mainWindow)
}

// showInvoiceDetails zeigt die Details einer Rechnung in einem Container an
func showInvoiceDetails(invoiceID uint, detailContainer *fyne.Container) {
	// Container leeren
	detailContainer.RemoveAll()
	
	if invoiceID == 0 {
		// Bei keiner ausgewählten Rechnung die Anzeige zentrieren und über volle Höhe
		message := widget.NewLabelWithStyle(
			"Keine Rechnung ausgewählt",
			fyne.TextAlignCenter,
			fyne.TextStyle{Italic: true, Bold: true},
		)
		detailContainer.RemoveAll()
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
	
	// Container für die Info-Karten mit Abstand
	infoCardsContainer := container.NewPadded(
		container.NewVBox(
			// Infokarte mit allgemeinen Daten - klar formatiert mit festen Breiten
			container.NewHBox(
				container.NewVBox(
					widget.NewLabelWithStyle("Lieferant:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
					widget.NewLabelWithStyle("Adresse:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
					widget.NewLabelWithStyle("Rechnungsdatum:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
					widget.NewLabelWithStyle("Fälligkeitsdatum:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				),
				container.NewVBox(
					widget.NewLabel(invoice.Supplier.Name),
					widget.NewLabel(invoice.Supplier.Address),
					widget.NewLabel(invoice.Date.Format("02.01.2006")),
					widget.NewLabel(invoice.DueDate.Format("02.01.2006")),
				),
			),
			
			widget.NewSeparator(),
			
			// Betragsinfos - klar formatiert mit festen Spaltenbreiten
			container.NewHBox(
				container.NewVBox(
					widget.NewLabelWithStyle("Gesamtbetrag:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
					widget.NewLabelWithStyle("MwSt-Betrag:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
					widget.NewLabelWithStyle("Status:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
					widget.NewLabelWithStyle("Kategorie:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				),
				container.NewVBox(
					widget.NewLabelWithStyle(formatAmount(invoice.TotalAmount), fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
					widget.NewLabelWithStyle(formatAmount(invoice.VatAmount), fyne.TextAlignTrailing, fyne.TextStyle{}),
					widget.NewLabel(invoice.Status),
					widget.NewLabel(invoice.Category),
				),
			),
		),
	)
	
	infoSection.Add(infoCardsContainer)
	
	// ---- Abschnitt: Rechnungspositionen ----
	positionsSection := container.NewVBox(
		widget.NewLabelWithStyle("Rechnungspositionen", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
	)
	
	// Rechnungspositionen als besseres Layout
	if len(invoice.InvoiceItems) > 0 {
		// Header-Zeile für die Tabelle
		headerRow := container.NewGridWithColumns(4,
			widget.NewLabelWithStyle("Beschreibung", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Menge", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Einzelpreis", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Gesamtpreis", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		)
		
		// Besser strukturierter Container für alle Positionen
		itemsContainer := container.NewVBox(
			headerRow,
			widget.NewSeparator(),
		)
		
		// Jede Position als eigene Zeile
		for _, item := range invoice.InvoiceItems {
			// Stelle sicher, dass die Werte korrekt formatiert werden
			// und verwende direkt die Strings aus den Rechnungspositionen wenn möglich
			description := item.Description
			
			// Prüfe, ob die Menge ein gültiger Float-Wert ist und formatiere entsprechend
			quantity := fmt.Sprintf("%.2f", item.Quantity)
			if item.Quantity == 0 && len(item.QuantityStr) > 0 {
				// Verwende den String-Wert, falls verfügbar
				quantity = item.QuantityStr
			}
			
			// Einheitspreis korrekt formatieren
			unitPrice := formatAmount(item.SinglePrice)
			if item.SinglePrice == 0 && len(item.SinglePriceStr) > 0 {
				// Verwende den String-Wert, falls verfügbar
				unitPrice = item.SinglePriceStr
				// Füge € hinzu, falls nicht vorhanden
				if !strings.Contains(unitPrice, "€") {
					unitPrice += " €"
				}
			}
			
			// Gesamtpreis korrekt formatieren
			totalPrice := formatAmount(item.TotalPrice)
			if item.TotalPrice == 0 && len(item.TotalPriceStr) > 0 {
				// Verwende den String-Wert, falls verfügbar
				totalPrice = item.TotalPriceStr
				// Füge € hinzu, falls nicht vorhanden
				if !strings.Contains(totalPrice, "€") {
					totalPrice += " €"
				}
			}
			
			row := container.NewGridWithColumns(4,
				widget.NewLabelWithStyle(description, fyne.TextAlignLeading, fyne.TextStyle{}),
				widget.NewLabelWithStyle(quantity, fyne.TextAlignTrailing, fyne.TextStyle{}),
				widget.NewLabelWithStyle(unitPrice, fyne.TextAlignTrailing, fyne.TextStyle{}),
				widget.NewLabelWithStyle(totalPrice, fyne.TextAlignTrailing, fyne.TextStyle{}),
			)
			
			// Einheitliche Zeilenhöhe für alle Positionen
			itemsContainer.Add(container.NewPadded(row))
		}
		
		// Trennlinie vor der Gesamtsumme
		itemsContainer.Add(widget.NewSeparator())
		
		// Gesamtsumme als Zusammenfassung am Ende
		totalRow := container.NewGridWithColumns(4,
			widget.NewLabel(""),
			widget.NewLabel(""),
			widget.NewLabelWithStyle("Gesamtsumme:", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle(formatAmount(invoice.TotalAmount), fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		)
		
		itemsContainer.Add(container.NewPadded(totalRow))
		
		// Umgeben mit Padding für bessere Lesbarkeit
		positionsSection.Add(container.NewPadded(itemsContainer))
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
	
	// Dateien-Container mit klarem Layout und mehr Abstand
	filesContainer := container.NewPadded(container.NewVBox())
	
	if len(invoice.FileRefs) > 0 {
		for _, fileRef := range invoice.FileRefs {
			// Jede Datei mit Icon und besserem Abstand
			fileItem := container.NewPadded(
				container.NewHBox(
					widget.NewIcon(theme.FileIcon()),
					widget.NewLabel(fileRef.Filename),
				),
			)
			filesContainer.Add(fileItem)
		}
	} else {
		noFilesLabel := widget.NewLabelWithStyle(
			"Keine Dateien vorhanden", 
			fyne.TextAlignCenter, 
			fyne.TextStyle{Italic: true},
		)
		filesContainer.Add(noFilesLabel)
	}
	
	filesSection.Add(filesContainer)
	
	// ---- Abschnitt: Aktionen ----
	actionsSection := container.NewVBox(
		widget.NewLabelWithStyle("Aktionen", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
	)
	
	// Aktions-Buttons mit besserer Strukturierung und mehr Abstand
	editButton := widget.NewButtonWithIcon("Bearbeiten", theme.DocumentCreateIcon(), func() {
		logrus.Info("Rechnung mit ID ", invoice.ID, " wird bearbeitet")
		showInvoiceEditDialog(invoice, detailContainer)
	})
	editButton.Importance = widget.HighImportance
	editButton.Resize(fyne.NewSize(150, editButton.MinSize().Height))
	
	deleteButton := widget.NewButtonWithIcon("Löschen", theme.DeleteIcon(), func() {
		dialog.ShowConfirm(
			"Rechnung löschen",
			"Möchten Sie die Rechnung wirklich löschen? Diese Aktion kann nicht rückgängig gemacht werden.",
			func(confirm bool) {
				if confirm {
					logrus.Info("Rechnung mit ID ", invoice.ID, " wird gelöscht")
					err := db.DeleteInvoice(invoice.ID)
					if err != nil {
						dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
						return
					}
					
					// Zurück zur Listenansicht
					detailContainer.RemoveAll()
					detailContainer.Add(container.NewCenter(
						widget.NewLabelWithStyle("Rechnung erfolgreich gelöscht", 
							fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
					))
					detailContainer.Refresh()
					
					// Aktualisiere die Rechnungsliste
					refreshInvoiceList(nil)
				}
			},
			fyne.CurrentApp().Driver().AllWindows()[0],
		)
	})
	deleteButton.Importance = widget.MediumImportance
	deleteButton.Resize(fyne.NewSize(150, deleteButton.MinSize().Height))
	
	// Größerer Abstand zwischen den Buttons
	actionsContainer := container.NewHBox(
		container.NewPadded(editButton),
		widget.NewLabel("    "),  // Extra Abstand
		container.NewPadded(deleteButton),
	)
	
	// Aktionen zentriert und mit Abstand
	actionsSection.Add(container.NewPadded(container.NewCenter(actionsContainer)))
	
	// Zusammenstellen aller Komponenten mit mehr Abstand und besserer Lesbarkeit
	contentBox := container.NewVBox(
		// Titel mit mehr Abstand
		container.NewPadded(container.NewPadded(title)),
		widget.NewSeparator(),
		
		// Sections mit Abständen dazwischen für bessere Übersicht
		container.NewPadded(infoSection),
		widget.NewSeparator(),
		container.NewPadded(positionsSection),
		widget.NewSeparator(),
		container.NewPadded(filesSection),
		widget.NewSeparator(),
		container.NewPadded(actionsSection),
	)
	
	// Scroll-Container mit Padding umgeben
	scrollContainer := container.NewVScroll(contentBox)
	
	// Größeren Inhaltsbereich nutzen - füllt die komplette Höhe aus
	detailWrapper := container.NewMax(
		container.NewPadded(scrollContainer),
	)
	
	detailContainer.Add(detailWrapper)
	
	// Container aktualisieren
	detailContainer.Refresh()
}