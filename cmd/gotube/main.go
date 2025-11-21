package main

import (
	"gotube/internal/gui"
	"os"
	"path/filepath"
)

func main() {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config", "gotube")
	os.MkdirAll(configPath, 0755)
	gui.StartApp()
}
