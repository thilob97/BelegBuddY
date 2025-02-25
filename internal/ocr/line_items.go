package ocr

// Keine Imports benötigt

// ExtractLineItems extrahiert Rechnungspositionen aus dem OCR-Text
func ExtractLineItems(text string) []LineItem {
	var lineItems []LineItem

	// Beispiel-LineItems für die Demo
	lineItems = append(lineItems, LineItem{
		Description: "Produkt A",
		Quantity:    "2",
		UnitPrice:   "59,99",
		TotalPrice:  "119,98",
	})
	
	lineItems = append(lineItems, LineItem{
		Description: "Dienstleistung B",
		Quantity:    "3",
		UnitPrice:   "45,00",
		TotalPrice:  "135,00",
	})
	
	lineItems = append(lineItems, LineItem{
		Description: "Produkt C",
		Quantity:    "1",
		UnitPrice:   "84,50",
		TotalPrice:  "84,50",
	})

	return lineItems
}
