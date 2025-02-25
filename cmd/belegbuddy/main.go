package main

import (
	"github.com/belegbuddy/belegbuddy/internal/app"
	"github.com/sirupsen/logrus"
)

func main() {
	// Erstelle und starte die Anwendung
	belegBuddyApp := app.New()
	
	// Anwendung ausführen (blockiert bis Fenster geschlossen wird)
	belegBuddyApp.Run()
	
	logrus.Info("BelegBuddY beendet")
}