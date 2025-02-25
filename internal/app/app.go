package app

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/belegbuddy/belegbuddy/internal/config"
	"github.com/belegbuddy/belegbuddy/internal/db"
	"github.com/sirupsen/logrus"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// App repräsentiert die BelegBuddY-Anwendung
type App struct {
	Config     *config.Config
	FyneApp    fyne.App
	MainWindow fyne.Window
}

// New erstellt eine neue App-Instanz
func New() *App {
	// Konfiguration initialisieren
	cfg := config.DefaultConfig()
	
	// Konfigurationsdatei laden, falls vorhanden
	configPath := filepath.Join(cfg.AppDir, "config.json")
	if _, err := os.Stat(configPath); err == nil {
		if file, err := os.Open(configPath); err == nil {
			defer file.Close()
			if err := json.NewDecoder(file).Decode(cfg); err != nil {
				logrus.Warn("Fehler beim Laden der Konfigurationsdatei: ", err)
			} else {
				logrus.Info("Konfiguration aus Datei geladen: ", configPath)
			}
		}
	}
	
	// Logging einrichten
	setupLogging(cfg)

	// Datenbank initialisieren
	if err := db.InitDB(cfg.DatabasePath); err != nil {
		logrus.Fatal("Fehler beim Initialisieren der Datenbank: ", err)
	}

	// Fyne-App erstellen
	fyneApp := app.New()
	mainWindow := fyneApp.NewWindow("BelegBuddY - Rechnungsdigitalisierung")
	mainWindow.Resize(fyne.NewSize(float32(cfg.WindowWidth), float32(cfg.WindowHeight)))

	return &App{
		Config:     cfg,
		FyneApp:    fyneApp,
		MainWindow: mainWindow,
	}
}

// Run startet die Anwendung
func (a *App) Run() {
	// UI aufbauen
	buildUI(a)
	
	// Hauptfenster anzeigen und Anwendungsschleife starten
	a.MainWindow.ShowAndRun()
}

// setupLogging richtet das Logging ein
func setupLogging(cfg *config.Config) {
	logFile := filepath.Join(cfg.AppDir, "belegbuddy.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatal("Fehler beim Öffnen der Logdatei: ", err)
	}

	// Multi-Writer für Log-Ausgabe sowohl in Datei als auch auf Konsole
	mw := io.MultiWriter(os.Stdout, file)
	logrus.SetOutput(mw)
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logrus.Info("BelegBuddY gestartet")
}