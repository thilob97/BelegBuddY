package ocr

// OCRResult enthält Ergebnisse der OCR-Verarbeitung
type OCRResult struct {
	FullText     string
	PossibleDate string
	PossibleSum  string
	Supplier     string
}