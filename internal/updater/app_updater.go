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
	"strings"
)

// UPDATE THIS URL TO YOUR PUBLIC REPO
const repoURL = "https://api.github.com/repos/ToddHoward42069/GoTube-Downloader/releases/latest"

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// isAppImage checks if we are running inside an AppImage
func isAppImage() bool {
	return os.Getenv("APPIMAGE") != ""
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

	if rel.TagName == models.AppVersion || rel.TagName == "" {
		return "", "", nil
	}

	// Determine which file to download
	var targetName string

	if isAppImage() {
		// Look for the .AppImage file
		// build.sh names it: GoTube-VERSION-x86_64.AppImage
		// We match loosely by suffix to be safe
		targetName = ".AppImage"
	} else if runtime.GOOS == "windows" {
		targetName = "gotube-windows-amd64.exe"
	} else {
		targetName = "gotube-linux-amd64"
	}

	for _, asset := range rel.Assets {
		if isAppImage() {
			if strings.HasSuffix(asset.Name, ".AppImage") {
				return rel.TagName, asset.BrowserDownloadURL, nil
			}
		} else {
			if asset.Name == targetName {
				return rel.TagName, asset.BrowserDownloadURL, nil
			}
		}
	}

	return "", "", fmt.Errorf("no compatible binary found in release")
}

// DoAppUpdate downloads the new binary/AppImage and replaces the current one
func DoAppUpdate(url string, progress func(float64)) error {
	// 1. Determine Target Path
	var targetPath string
	if isAppImage() {
		// In AppImage mode, we replace the AppImage file itself, not the internal binary
		targetPath = os.Getenv("APPIMAGE")
	} else {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		targetPath = exe
	}

	// 2. Download to .new file
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	size := resp.ContentLength
	newPath := targetPath + ".new"

	out, err := os.Create(newPath)
	if err != nil {
		return fmt.Errorf("cannot create update file: %v", err)
	}

	buf := make([]byte, 1024*32) // 32KB buffer
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

	// 3. Make Executable
	if runtime.GOOS != "windows" {
		if err := os.Chmod(newPath, 0755); err != nil {
			return err
		}
	}

	// 4. Swap Files
	if runtime.GOOS == "windows" {
		oldPath := targetPath + ".old"
		os.Rename(targetPath, oldPath)
		if err := os.Rename(newPath, targetPath); err != nil {
			return err
		}
	} else {
		if err := os.Rename(newPath, targetPath); err != nil {
			return err
		}
	}

	return nil
}

// RestartApp attempts to restart the application
func RestartApp() {
	var cmd *exec.Cmd

	if isAppImage() {
		// Relaunch the AppImage file
		appImage := os.Getenv("APPIMAGE")
		cmd = exec.Command(appImage)
	} else {
		// Relaunch binary
		executable, _ := os.Executable()
		cmd = exec.Command(executable)
	}

	cmd.Start()
	os.Exit(0)
}
