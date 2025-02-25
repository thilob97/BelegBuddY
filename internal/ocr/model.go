package ocr

// OCRResult enthält Ergebnisse der OCR-Verarbeitung
type OCRResult struct {
	FullText     string
	PossibleDate string
	PossibleSum  string
	Supplier     string
	LineItems    []LineItem // Erkannte Rechnungspositionen
	IsDemo       bool       // Gibt an, ob es sich um Demo-Daten handelt
}

// LineItem repräsentiert eine Rechnungsposition aus dem OCR-Ergebnis
type LineItem struct {
	Description string
	Quantity    string
	UnitPrice   string
	TotalPrice  string
}