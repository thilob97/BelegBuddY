package main

import (
	"encoding/json"
	"fmt"
	"github.com/belegbuddy/belegbuddy/internal/db"
	"github.com/sirupsen/logrus"
)

// SupplierSummary gleiche Struktur wie in der Datenbank
type SupplierSummary struct {
	Name   string
	Amount float64
	Count  int
}

func main() {
	// Datenbank initialisieren
	err := db.InitDB("/Users/thilobarth/.belegbuddy/belegbuddy.db")
	if err != nil {
		logrus.Fatalf("Fehler beim Initialisieren der Datenbank: %v", err)
	}

	// Dashboard-Daten laden
	data, err := db.GetDashboardData()
	if err != nil {
		logrus.Fatalf("Fehler beim Laden der Dashboard-Daten: %v", err)
	}

	// topSuppliers extrahieren
	topSuppliers, ok := data["topSuppliers"]
	if !ok {
		logrus.Fatal("Keine topSuppliers in den Dashboard-Daten gefunden")
	}

	// Als JSON ausgeben
	jsonData, err := json.MarshalIndent(topSuppliers, "", "  ")
	if err != nil {
		logrus.Fatalf("Fehler beim Konvertieren zu JSON: %v", err)
	}
	
	fmt.Println("Typ von topSuppliers:", fmt.Sprintf("%T", topSuppliers))
	fmt.Println("JSON-Struktur von topSuppliers:")
	fmt.Println(string(jsonData))
	
	// Versuchen, direkt als []interface{} zu konvertieren
	if suppliers, ok := topSuppliers.([]interface{}); ok {
		fmt.Printf("\nAnzahl Lieferanten: %d\n", len(suppliers))
		
		// Jeden Lieferanten im Array durchgehen
		for i, supplier := range suppliers {
			fmt.Printf("\nLieferant %d:\n", i)
			fmt.Printf("  Typ: %T\n", supplier)
			
			// Als JSON ausgeben
			supplierJSON, _ := json.MarshalIndent(supplier, "  ", "  ")
			fmt.Printf("  Daten: %s\n", string(supplierJSON))
			
			// Konvertierung in unser Modell versuchen
			if s, ok := supplier.(map[string]interface{}); ok {
				fmt.Println("  Konnte als Map extrahieren")
				
				// Felder auslesen
				name, ok1 := s["Name"].(string)
				amount, ok2 := s["Amount"].(float64)
				count, ok3 := s["Count"].(float64)
				
				if ok1 && ok2 && ok3 {
					fmt.Printf("  Erfolgreiche Extraktion: Name=%s, Amount=%.2f, Count=%d\n", 
						name, amount, int(count))
				} else {
					fmt.Println("  Nicht alle Felder konnten extrahiert werden:", ok1, ok2, ok3)
				}
			}
		}
	} else {
		fmt.Println("Konnte topSuppliers nicht als []interface{} konvertieren")
		
		// Versuchen, direkt als []SupplierSummary zu konvertieren
		var suppliers []SupplierSummary
		jsonStr, _ := json.Marshal(topSuppliers)
		if err := json.Unmarshal(jsonStr, &suppliers); err != nil {
			fmt.Println("Konvertierung über JSON zu []SupplierSummary fehlgeschlagen:", err)
		} else {
			fmt.Println("\nKonnte über JSON konvertieren:")
			for i, s := range suppliers {
				fmt.Printf("  Lieferant %d: Name=%s, Amount=%.2f, Count=%d\n", 
					i, s.Name, s.Amount, s.Count)
			}
		}
	}
}
