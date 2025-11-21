package main

import (
	"gotube/internal/gui"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/app"
)

func main() {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config", "gotube")
	os.MkdirAll(configPath, 0755)

	// Create app here to apply settings before GUI starts
	a := app.NewWithID("com.github.gotube.downloader")

	// Apply the Custom Theme
	a.Settings().SetTheme(&gui.CustomTheme{}) // We need to export the struct in gui package

	gui.StartApp(a) // Pass the app instance
}
