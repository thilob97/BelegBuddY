package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/belegbuddy/belegbuddy/internal/ocr"
)

// OCRRequest ist das Anfrage-Format für die OCR-Verarbeitung
type OCRRequest struct {
	FilePath string `json:"file_path"`
	Language string `json:"language"`
}

// OCRResponse ist das Antwort-Format für die OCR-Verarbeitung
type OCRResponse struct {
	Error      string      `json:"error,omitempty"`
	OCRResult  *ocr.OCRResult `json:"result,omitempty"`
}

func main() {
	// Prüfen, ob genügend Argumente übergeben wurden
	if len(os.Args) < 2 {
		fmt.Println(`{"error": "Keine JSON-Anfrage übergeben"}`)
		os.Exit(1)
	}

	// JSON-Anfrage parsen
	requestJSON := os.Args[1]
	var request OCRRequest
	if err := json.Unmarshal([]byte(requestJSON), &request); err != nil {
		fmt.Printf(`{"error": "Fehler beim Parsen der JSON-Anfrage: %s"}`, err)
		os.Exit(1)
	}

	// Datei prüfen
	fileInfo, err := os.Stat(request.FilePath)
	if err != nil {
		fmt.Printf(`{"error": "Datei nicht gefunden: %s"}`, err)
		os.Exit(1)
	}

	if fileInfo.IsDir() {
		fmt.Printf(`{"error": "Der angegebene Pfad ist ein Verzeichnis, keine Datei"}`)
		os.Exit(1)
	}

	// Dateiendung prüfen
	fileExt := strings.ToLower(filepath.Ext(request.FilePath))
	fileExt = strings.TrimPrefix(fileExt, ".")
	supportedExts := map[string]bool{
		"pdf":  true,
		"png":  true,
		"jpg":  true,
		"jpeg": true,
		"tiff": true,
		"tif":  true,
	}

	if !supportedExts[fileExt] {
		fmt.Printf(`{"error": "Nicht unterstütztes Dateiformat: %s"}`, fileExt)
		os.Exit(1)
	}

	// OCR mit Mock-Implementierung ausführen
	result, err := ocr.ProcessFile(request.FilePath, request.Language)
	
	response := OCRResponse{}
	if err != nil {
		response.Error = fmt.Sprintf("OCR-Fehler: %s", err)
	} else {
		response.OCRResult = result
	}

	// Antwort als JSON ausgeben
	responseJSON, err := json.Marshal(response)
	if err != nil {
		fmt.Printf(`{"error": "Fehler beim Generieren der JSON-Antwort: %s"}`, err)
		os.Exit(1)
	}

	fmt.Println(string(responseJSON))
}