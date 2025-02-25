package main

import (
	"encoding/json"
	"fmt"
	"github.com/belegbuddy/belegbuddy/internal/db"
	"github.com/sirupsen/logrus"
)

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
	
	// Wenn es ein Slice ist, die einzelnen Elemente ausgeben
	if suppliers, ok := topSuppliers.([]interface{}); ok {
		fmt.Printf("\nAnzahl Lieferanten: %d\n", len(suppliers))
		
		for i, supplier := range suppliers {
			fmt.Printf("\nLieferant %d:\n", i)
			fmt.Printf("  Typ: %T\n", supplier)
			
			// Als JSON ausgeben
			supplierJSON, _ := json.MarshalIndent(supplier, "  ", "  ")
			fmt.Printf("  Daten: %s\n", string(supplierJSON))
		}
	}
}
