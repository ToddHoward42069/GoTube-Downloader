package gui

import (
	"fmt"
	"gotube/internal/locales"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// createFormatSelectors returns the pair of dropdowns for Video/Audio selection
func createFormatSelectors() (*widget.Select, *widget.Select) {
	formatSelect := widget.NewSelect([]string{locales.Get("format_video"), locales.Get("format_audio")}, nil)
	formatSelect.Selected = locales.Get("format_video")

	detailSelect := widget.NewSelect([]string{"Best", "4k", "1080p", "720p"}, nil)
	detailSelect.Selected = "Best"

	formatSelect.OnChanged = func(s string) {
		if s == locales.Get("format_audio") {
			detailSelect.Options = []string{"Best", "mp3", "m4a"}
			detailSelect.Selected = "mp3"
		} else {
			detailSelect.Options = []string{"Best", "4k", "1080p", "720p"}
			detailSelect.Selected = "Best"
		}
		detailSelect.Refresh()
	}
	return formatSelect, detailSelect
}

// createPreviewImage returns a standard configured image canvas
func createPreviewImage() *canvas.Image {
	img := canvas.NewImageFromResource(theme.FileImageIcon())
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(200, 120))
	return img
}

// formatDuration formats seconds into readable string
func formatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	seconds = seconds % 60
	if minutes < 60 {
		return fmt.Sprintf("%dm %02ds", minutes, seconds)
	}
	hours := minutes / 60
	minutes = minutes % 60
	return fmt.Sprintf("%dh %02dm %02ds", hours, minutes, seconds)
}
