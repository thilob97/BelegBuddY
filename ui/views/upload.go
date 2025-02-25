package views

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/belegbuddy/belegbuddy/internal/config"
	"github.com/belegbuddy/belegbuddy/internal/ocr"
	"github.com/belegbuddy/belegbuddy/ui/components"
	"github.com/sirupsen/logrus"
)

// UploadViewCallbacks enthält Callback-Funktionen für die Upload-Ansicht
type UploadViewCallbacks struct {
	OnFileProcessed func(filepath string, result *ocr.OCRResult)
}

// NewUploadView erstellt die Upload-Ansicht
func NewUploadView(window fyne.Window, uploadsDir string, callbacks UploadViewCallbacks) fyne.CanvasObject {
	// Status-Anzeige
	statusLabel := widget.NewLabel("")
	
	// Progress-Anzeige
	progress := widget.NewProgressBar()
	progress.Hide()
	
	// OCR-Methode auswählen
	ocrMethodLabel := widget.NewLabelWithStyle("OCR-Methode:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	ocrMethodSelect := widget.NewSelect([]string{"Tesseract (lokal)", "Claude AI"}, nil)
	ocrMethodSelect.SetSelected("Tesseract (lokal)")
	
	// Funktion für das Verarbeiten einer Datei
	processFile := func(filePath string) {
		// Status aktualisieren
		statusLabel.SetText(fmt.Sprintf("Verarbeite: %s", filepath.Base(filePath)))
		progress.Show()
		
		// OCR-Verarbeitung starten
		go func() {
			// Ausgewählte OCR-Methode ermitteln
			useClaudeOCR := ocrMethodSelect.Selected == "Claude AI"
			
			progressDialog := dialog.NewProgress("Verarbeite Dokument", 
				fmt.Sprintf("OCR-Verarbeitung mit %s läuft...", ocrMethodSelect.Selected), window)
			progressDialog.Show()
			
			var result *ocr.OCRResult
			var err error
			
			// Je nach ausgewählter Methode verarbeiten
			if useClaudeOCR {
				// Config auslesen
				appConfig := &config.Config{}
				configFilePath := filepath.Join(uploadsDir, "..", "config.json")
				if fileData, err := os.ReadFile(configFilePath); err == nil {
					json.Unmarshal(fileData, appConfig)
				}
				
				// API-Key prüfen
				if appConfig.ClaudeAPIKey == "" {
					err = errors.New("Claude API-Schlüssel nicht konfiguriert. Bitte in den Einstellungen hinterlegen")
				} else {
					// Claude API verwenden
					result, err = ocr.ProcessWithClaude(filePath, appConfig.ClaudeAPIKey)
				}
			} else {
				// Tesseract OCR verwenden
				result, err = ocr.ProcessFile(filePath, "deu") // Sprache könnte aus Config kommen
			}
			
			// Dialog schließen
			progressDialog.Hide()
			
			if err != nil {
				logrus.Error("Fehler bei der OCR-Verarbeitung: ", err)
				dialog.ShowError(err, window)
				statusLabel.SetText(fmt.Sprintf("Fehler: %v", err))
				progress.Hide()
				return
			}
			
			// Status aktualisieren
			statusLabel.SetText(fmt.Sprintf("Verarbeitung von %s abgeschlossen", filepath.Base(filePath)))
			progress.Hide()
			
			// Callback aufrufen
			if callbacks.OnFileProcessed != nil {
				callbacks.OnFileProcessed(filePath, result)
			}
		}()
	}
	
	// Drag & Drop Bereich
	handleFileDrop := func(uris []fyne.URI) {
		// Nur eine Datei gleichzeitig bearbeiten
		if len(uris) > 0 {
			uri := uris[0]
			
			// Prüfen ob lokale Datei
			if uri.Scheme() != "file" {
				dialog.ShowError(fmt.Errorf("nur lokale Dateien werden unterstützt"), window)
				return
			}
			
			// Zieldatei erstellen
			destFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(uri.Path()))
			destPath := filepath.Join(uploadsDir, destFileName)
			
			// Datei lesen und kopieren
			reader, err := storage.Reader(uri)
			if err != nil {
				logrus.Error("Fehler beim Lesen der Datei: ", err)
				dialog.ShowError(err, window)
				return
			}
			defer reader.Close()
			
			dest, err := os.Create(destPath)
			if err != nil {
				logrus.Error("Fehler beim Erstellen der Zieldatei: ", err)
				dialog.ShowError(err, window)
				return
			}
			defer dest.Close()
			
			if _, err := io.Copy(dest, reader); err != nil {
				logrus.Error("Fehler beim Kopieren der Datei: ", err)
				dialog.ShowError(err, window)
				return
			}
			
			// Datei verarbeiten
			processFile(destPath)
		}
	}
	
	// Button zum manuellen Auswählen von Dateien
	uploadButton := widget.NewButton("Datei auswählen", nil)
	uploadButton.OnTapped = components.GetUploadButtonCallback(handleFileDrop, window)
	
	dropArea := components.DragDropArea(handleFileDrop)
	
	// OCR-Methoden-Selector
	ocrMethodContainer := container.NewHBox(
		ocrMethodLabel,
		ocrMethodSelect,
	)
	
	// Layout zusammenstellen
	statusContainer := container.NewVBox(
		statusLabel,
		progress,
	)
	
	return container.NewBorder(
		ocrMethodContainer, 
		container.NewVBox(
			container.NewCenter(uploadButton),
			statusContainer,
		),
		nil, nil,
		dropArea,
	)
}