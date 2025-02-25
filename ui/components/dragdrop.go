package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

// DragDropCallback ist eine Funktion, die aufgerufen wird, wenn Dateien per Drag & Drop abgelegt werden
type DragDropCallback func([]fyne.URI)

// DragDropArea erstellt einen Container, der Drag & Drop unterstützt
func DragDropArea(callback DragDropCallback) *fyne.Container {
	// Icon und Text
	icon := widget.NewIcon(theme.FolderOpenIcon())
	icon.Resize(fyne.NewSize(64, 64))
	
	label := widget.NewLabel("Rechnungen hier ablegen oder klicken zum Auswählen")
	
	// Container für visuelles Feedback
	dropBorder := canvas.NewRectangle(color.NRGBA{R: 200, G: 200, B: 230, A: 255})
	dropBorder.StrokeWidth = 2
	dropBorder.StrokeColor = color.NRGBA{R: 100, G: 100, B: 180, A: 255}
	
	// Content aufbauen
	content := container.NewCenter(
		container.NewVBox(
			container.NewCenter(icon),
			container.NewCenter(label),
		),
	)
	
	// Container mit Border und Content
	dropArea := container.NewMax(dropBorder, content)
	
	// Button für Dateiauswahl einrichten
	uploadButton := widget.NewButton("Datei auswählen", nil)
	
	// Button zum Callback hinzufügen
	if callback != nil {
		uploadButton.OnTapped = func() {
			// Callback wird über die Upload-View implementiert
		}
	}
	
	content.Add(container.NewCenter(uploadButton))
	
	return dropArea
}

// GetUploadButtonCallback gibt eine Funktion zurück, die einen Dateiauswahldialog öffnet
func GetUploadButtonCallback(callback DragDropCallback, window fyne.Window) func() {
	return func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()
			
			// URI an den Callback übergeben
			if callback != nil {
				callback([]fyne.URI{reader.URI()})
			}
		}, window)
		
		// Filter für unterstützte Dateitypen setzen
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".pdf", ".png", ".jpg", ".jpeg", ".tiff", ".tif"}))
		fd.Show()
	}
}