package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
)

// showMessage displays an information dialog
func (d *DatabaseApp) showMessage(title, message string) {
	dialog.ShowInformation(title, message, d.window)
}

// contains checks if a substring exists in a string
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// createThumbnailFromBytes creates a small canvas image from raw bytes
func createThumbnailFromBytes(imageData []byte) (*canvas.Image, error) {
	if len(imageData) == 0 {
		return nil, fmt.Errorf("пустые данные изображения")
	}

	// Decode image
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования изображения: %v", err)
	}

	// Create canvas.Image
	canvasImg := canvas.NewImageFromImage(img)

	// Set fill mode and size
	canvasImg.FillMode = canvas.ImageFillContain
	canvasImg.SetMinSize(fyne.NewSize(30, 30))
	// Note: The second SetMinSize in the original code overwrote the first.
	// Assuming 100x100 is preferred for visibility.
	canvasImg.SetMinSize(fyne.NewSize(100, 100))

	return canvasImg, nil
}
