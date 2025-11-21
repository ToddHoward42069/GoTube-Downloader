package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type CustomTheme struct{}

var _ fyne.Theme = (*CustomTheme)(nil)

// --- Professional Dark Palette ---
var (
	colBackground = color.RGBA{R: 24, G: 24, B: 24, A: 255}    // #181818 (Deep Grey)
	colSurface    = color.RGBA{R: 37, G: 37, B: 38, A: 255}    // #252526 (Lighter Grey)
	colPrimary    = color.RGBA{R: 53, G: 132, B: 228, A: 255}  // #3584e4 (Standard Blue)
	colText       = color.RGBA{R: 240, G: 240, B: 240, A: 255} // #f0f0f0
	colSubText    = color.RGBA{R: 150, G: 150, B: 150, A: 255} // #969696
	colInput      = color.RGBA{R: 45, G: 45, B: 45, A: 255}    // #2d2d2d
)

func (m CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return colBackground
	case theme.ColorNameButton, theme.ColorNameOverlayBackground:
		return colSurface
	case theme.ColorNamePrimary, theme.ColorNameHyperlink, theme.ColorNameFocus:
		return colPrimary
	case theme.ColorNameForeground:
		return colText
	case theme.ColorNamePlaceHolder, theme.ColorNameDisabled:
		return colSubText
	case theme.ColorNameInputBackground:
		return colInput
	case theme.ColorNameScrollBar:
		return color.RGBA{R: 255, G: 255, B: 255, A: 20}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (m CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6 // Standard padding
	case theme.SizeNameScrollBar:
		return 8
	case theme.SizeNameText:
		return 13
	case theme.SizeNameInputRadius:
		return 4 // Sharper corners
	}
	return theme.DefaultTheme().Size(name)
}
