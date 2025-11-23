package updater

import (
	"encoding/json"
	"fmt"
	"gotube/internal/models"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

const repoURL = "https://api.github.com/repos/ToddHoward42069/GoTube-Downloader/releases/latest"

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// CheckAppUpdate returns the new version tag and download URL if an update exists
func CheckAppUpdate() (string, string, error) {
	resp, err := http.Get(repoURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("github api returned %d", resp.StatusCode)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", "", err
	}

	// Simple string comparison. Ideally use semantic versioning library.
	// Remove 'v' prefix for comparison if needed, but assuming strictly v1.0.0 format
	if rel.TagName == models.AppVersion || rel.TagName == "" {
		return "", "", nil // No update
	}

	// Find matching asset for current OS
	targetName := "gotube-linux-amd64"
	if runtime.GOOS == "windows" {
		targetName = "gotube-windows-amd64.exe"
	}

	for _, asset := range rel.Assets {
		if asset.Name == targetName {
			return rel.TagName, asset.BrowserDownloadURL, nil
		}
	}

	return "", "", fmt.Errorf("no compatible binary found in release")
}

// DoAppUpdate downloads the new binary and replaces the current one
func DoAppUpdate(url string, progress func(float64)) error {
	// 1. Download to temp
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Calculate progress
	size := resp.ContentLength

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	// New file path
	newPath := executable + ".new"

	out, err := os.Create(newPath)
	if err != nil {
		return err
	}

	// Copy with progress
	buf := make([]byte, 1024)
	var downloaded int64
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
			downloaded += int64(n)
			if size > 0 {
				progress(float64(downloaded) / float64(size))
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			out.Close()
			return err
		}
	}
	out.Close()

	// 2. Replace
	if runtime.GOOS == "windows" {
		// Windows can't delete running exe.
		// We rename running to .old, move .new to .exe
		oldPath := executable + ".old"
		os.Rename(executable, oldPath)
		if err := os.Rename(newPath, executable); err != nil {
			return err
		}
		// Note: User needs to restart manually or we trigger a restart command
	} else {
		// Linux allows overwriting running binaries (mostly)
		if err := os.Chmod(newPath, 0755); err != nil {
			return err
		}
		if err := os.Rename(newPath, executable); err != nil {
			return err
		}
	}

	return nil
}

// RestartApp attempts to restart the application
func RestartApp() {
	executable, _ := os.Executable()
	cmd := exec.Command(executable)
	cmd.Start()
	os.Exit(0)
}
