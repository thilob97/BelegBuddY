package ocr

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ProcessDemoFile erstellt ein Demo-OCR-Ergebnis für die Testdatei
func ProcessDemoFile(filePath string) (*OCRResult, error) {
	logrus.Info("Verarbeite Demo-Datei: ", filePath)
	
	// Basisdaten für die Demo
	supplier := "Beispiel GmbH"
	date := "15.02.2023"
	amount := "343,91 €"
	
	// Anpassen des Namens basierend auf Dateiname
	fileName := filepath.Base(filePath)
	if strings.Contains(strings.ToLower(fileName), "rechnung") {
		supplier = "Beispiel GmbH"
		date = "15.02.2023"
		amount = "343,91 €"
	} else if strings.Contains(strings.ToLower(fileName), "quittung") {
		supplier = "Einzelhandel KG"
		date = time.Now().AddDate(0, 0, -5).Format("02.01.2006")
		amount = "74,74 €"
	}
	
	// Demo-OCR-Ergebnis erstellen
	result := &OCRResult{
		FullText: generateDemoText(supplier, date, amount),
		PossibleDate: date,
		PossibleSum: amount,
		Supplier: supplier,
		IsDemo: true,
	}
	
	return result, nil
}

// generateDemoText erzeugt einen Demo-Text basierend auf den übergebenen Werten
func generateDemoText(supplier, date, amount string) string {
	var text string
	
	if supplier == "Beispiel GmbH" {
		text = fmt.Sprintf("%s\nMusterstraße 123\n12345 Musterstadt\n\n", supplier)
		text += "Rechnungsnummer: R-2023-1234\n"
		text += fmt.Sprintf("Rechnungsdatum: %s\n\n", date)
		text += "Position    Beschreibung                Anzahl    Einzelpreis    Gesamtpreis\n"
		text += "---------------------------------------------------------------------------\n"
		text += "1           Premium Service              1         199,00 €       199,00 €\n"
		text += "2           Zusatzleistung              2          45,00 €        90,00 €\n\n"
		text += "Zwischensumme:                                                  289,00 €\n"
		text += "Mehrwertsteuer 19%:                                              54,91 €\n"
		text += fmt.Sprintf("Gesamtbetrag:                                                   %s\n", amount)
	} else {
		text = fmt.Sprintf("%s\nHauptstraße 78\n10559 Berlin\n\n", supplier)
		text += "Quittung Nr.: Q-5678\n"
		text += fmt.Sprintf("Datum: %s\n", date)
		text += "Kunde: Barzahler\n\n"
		text += "Artikel                                      Preis\n"
		text += "---------------------------------------------------\n"
		text += "Bürobedarf                               45,75 €\n"
		text += "Fachliteratur                            28,99 €\n"
		text += "---------------------------------------------------\n"
		text += fmt.Sprintf("Gesamtbetrag                             %s\n\n", amount)
		text += "Vielen Dank für Ihren Einkauf!\n"
	}
	
	return text
}