package updater

import (
	"fmt"
	"gotube/internal/utils"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type BinaryManager struct {
	ConfigDir string
}

func NewBinaryManager() *BinaryManager {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config", "gotube", "bin")
	os.MkdirAll(configPath, 0755)
	return &BinaryManager{ConfigDir: configPath}
}

func (bm *BinaryManager) GetYtDlpPath() string {
	localName := utils.GetExecutableName("yt-dlp", runtime.GOOS)
	localPath := filepath.Join(bm.ConfigDir, localName)
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}
	if path, err := exec.LookPath("yt-dlp"); err == nil {
		return path
	}
	return localPath
}

func (bm *BinaryManager) UpdateBinary(progress func(string)) error {
	url := "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp"
	if runtime.GOOS == "windows" { url += ".exe" }

	localName := utils.GetExecutableName("yt-dlp", runtime.GOOS)
	destPath := filepath.Join(bm.ConfigDir, localName)

	progress(fmt.Sprintf("Fetching from: %s", url))
	resp, err := http.Get(url)
	if err != nil { return fmt.Errorf("network error: %v", err) }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { return fmt.Errorf("bad status: %s", resp.Status) }
	out, err := os.Create(destPath)
	if err != nil { return fmt.Errorf("file create error: %v", err) }
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil { return fmt.Errorf("write error: %v", err) }

	if runtime.GOOS != "windows" {
		if err := os.Chmod(destPath, 0755); err != nil { return fmt.Errorf("chmod error: %v", err) }
	}
	progress(fmt.Sprintf("Updated successfully to: %s", destPath))
	return nil
}
