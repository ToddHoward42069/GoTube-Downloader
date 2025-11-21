package utils

import (
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

func SanitizeFilename(name string) string {
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	safe := reg.ReplaceAllString(name, "_")
	return strings.TrimSpace(safe)
}

func GetExecutableName(name string, osName string) string {
	if osName == "windows" {
		if !strings.HasSuffix(name, ".exe") {
			return name + ".exe"
		}
	}
	return name
}

func OpenFolder(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	cmd.Start()
}
