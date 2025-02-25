//go:build tesseract

package ocr

// #cgo CFLAGS: -I/opt/homebrew/Cellar/leptonica/1.85.0/include -I/opt/homebrew/Cellar/tesseract/5.5.0/include
// #cgo LDFLAGS: -L/opt/homebrew/Cellar/leptonica/1.85.0/lib -L/opt/homebrew/Cellar/tesseract/5.5.0/lib -lleptonica -ltesseract
import "C"
