package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

// NewDashboardView erstellt die Dashboard-Ansicht
func NewDashboardView() fyne.CanvasObject {
	// Statistik-Karten erstellen
	totalCard := createStatisticCard("Rechnungen gesamt", "0")
	totalAmountCard := createStatisticCard("Gesamtsumme", "0,00 €")
	thisMonthCard := createStatisticCard("Dieser Monat", "0,00 €")
	averageCard := createStatisticCard("Durchschnitt", "0,00 €")
	
	// Charts-Platzhalter
	monthlySummaryChart := createChartPlaceholder("Monatliche Ausgaben")
	supplierChart := createChartPlaceholder("Top Lieferanten")
	
	// Layout zusammenstellen
	statsContainer := container.NewGridWithColumns(4,
		totalCard,
		totalAmountCard,
		thisMonthCard,
		averageCard,
	)
	
	chartsContainer := container.NewGridWithColumns(2,
		container.NewPadded(monthlySummaryChart),
		container.NewPadded(supplierChart),
	)
	
	return container.NewBorder(
		container.NewPadded(statsContainer),
		nil, nil, nil,
		container.NewPadded(chartsContainer),
	)
}

// createStatisticCard erstellt eine Karte mit statistischen Informationen
func createStatisticCard(title, value string) fyne.CanvasObject {
	titleLabel := widget.NewLabel(title)
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	
	valueLabel := widget.NewLabel(value)
	valueLabel.TextStyle = fyne.TextStyle{Bold: true}
	valueLabel.Alignment = fyne.TextAlignCenter
	
	// Rechteck für Card-Hintergrund
	background := canvas.NewRectangle(color.NRGBA{R: 240, G: 240, B: 250, A: 255})
	
	// Layout
	content := container.NewVBox(
		titleLabel,
		container.NewCenter(valueLabel),
	)
	
	return container.NewPadded(
		container.NewMax(background, content),
	)
}

// createChartPlaceholder erstellt einen Platzhalter für Diagramme
func createChartPlaceholder(title string) fyne.CanvasObject {
	titleLabel := widget.NewLabel(title)
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	
	chartPlaceholder := canvas.NewRectangle(color.NRGBA{R: 230, G: 230, B: 240, A: 255})
	chartPlaceholder.SetMinSize(fyne.NewSize(300, 200))
	
	placeholderText := widget.NewLabel("Diagramm - noch nicht implementiert")
	placeholderText.Alignment = fyne.TextAlignCenter
	
	chartContent := container.NewCenter(placeholderText)
	
	return container.NewBorder(
		titleLabel,
		nil, nil, nil,
		container.NewMax(chartPlaceholder, chartContent),
	)
}