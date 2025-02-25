# BelegBuddY

BelegBuddY ist eine Desktop-Anwendung zur Rechnungsdigitalisierung, die lokale OCR-Verarbeitung und Datenspeicherung bietet.

## Features

- Drag-and-Drop-Upload von Rechnungs-PDFs und Bildern
- OCR-Erkennung mit Tesseract
- Extraktion von Rechnungsinformationen (Datum, Betrag, Lieferant, etc.)
- Manuelle Korrekturmöglichkeit
- Lokale SQLite-Datenbank
- Einfaches Dashboard mit Ausgabenübersicht

## Technologie-Stack

- Go (Golang)
- Fyne GUI-Framework
- Tesseract OCR via gosseract
- SQLite mit GORM
- Logrus für Logging

## Entwicklung

```bash
# Abhängigkeiten installieren
go mod tidy

# Anwendung bauen
go build -o belegbuddy ./cmd/belegbuddy

# Anwendung ausführen
go run ./cmd/belegbuddy
```