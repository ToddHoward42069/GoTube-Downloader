package downloader

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gotube/internal/models"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Engine struct {
	BinaryPath string
}

func NewEngine(binaryPath string) *Engine {
	return &Engine{BinaryPath: binaryPath}
}

func (e *Engine) GetMetadata(url string) (*models.VideoMetadata, error) {
	// --flat-playlist gives us the list of entries (ID + Title) very quickly
	cmd := exec.Command(e.BinaryPath, "--dump-single-json", "--flat-playlist", url)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var meta models.VideoMetadata
	if err := json.Unmarshal(output, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %v", err)
	}
	return &meta, nil
}

func (e *Engine) Download(config models.DownloadConfig, callback func(models.ProgressUpdate)) error {
	maxRetries := 3
	retryDelay := 5 * time.Second
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		args := e.buildArgs(config)
		err := e.executeCommand(args, callback)
		if err == nil {
			return nil
		}
		lastErr = err
		errMsg := err.Error()
		if strings.Contains(errMsg, "HTTP Error 429") {
			callback(models.ProgressUpdate{Text: "Rate limited. Waiting 30s...", Stage: "Retrying"})
			time.Sleep(30 * time.Second)
			continue
		}
		if strings.Contains(errMsg, "Sign in required") {
			return fmt.Errorf("authentication required: please import cookies")
		}
		if strings.Contains(errMsg, "fragment not found") {
			callback(models.ProgressUpdate{Text: "Fragment missing, retrying...", Stage: "Retrying"})
			time.Sleep(retryDelay)
			continue
		}
		callback(models.ProgressUpdate{Text: fmt.Sprintf("Error: %v. Retrying...", err), Stage: "Retrying"})
		time.Sleep(retryDelay)
	}
	return fmt.Errorf("failed after %d attempts: %v", maxRetries, lastErr)
}

func (e *Engine) buildArgs(config models.DownloadConfig) []string {
	if config.SafeMode {
		return []string{config.URL, "-o", filepath.Join(config.OutputPath, "safe_%(title)s.%(ext)s"), "-f", "best"}
	}
	args := []string{
		config.URL,
		"-o", filepath.Join(config.OutputPath, "%(title)s.%(ext)s"),
		"--no-mtime",
		"--newline",
		"--add-metadata", "--embed-thumbnail",
	}

	// PLAYLIST LOGIC
	if config.IsPlaylist {
		args = append(args, "--yes-playlist")
		// If user selected specific videos (e.g. "1,3,5")
		if config.PlaylistItems != "" {
			args = append(args, "--playlist-items", config.PlaylistItems)
		}
	} else {
		args = append(args, "--no-playlist")
	}

	if config.EmbedSubs {
		args = append(args, "--embed-subs")
		args = append(args, "--convert-subs", "srt") // Ensure embedding works in mp4
		if config.AutoSubs {
			args = append(args, "--write-auto-subs")
		}
		lang := "en.*"
		if config.SubLanguage == "de" {
			lang = "de.*,en.*"
		}
		if config.SubLanguage == "all" {
			lang = "all"
		}
		args = append(args, "--sub-langs", lang)
	}

	if config.DownloadMode == "Audio" {
		args = append(args, "-x")
		switch config.Quality {
		case "mp3":
			args = append(args, "--audio-format", "mp3", "--audio-quality", "0")
		case "m4a":
			args = append(args, "--audio-format", "m4a")
		default:
			args = append(args, "--audio-format", "best")
		}
	} else {
		args = append(args, "--merge-output-format", "mp4")
		switch config.Quality {
		case "4k":
			args = append(args, "-f", "bestvideo[height<=2160]+bestaudio/best")
		case "1080p":
			args = append(args, "-f", "bestvideo[height<=1080]+bestaudio/best")
		case "720p":
			args = append(args, "-f", "bestvideo[height<=720]+bestaudio/best")
		default:
			args = append(args, "-f", "bestvideo+bestaudio/best")
		}
	}

	if config.TrimStart != "" {
		section := fmt.Sprintf("*%s-%s", config.TrimStart, config.TrimEnd)
		if config.TrimEnd == "" {
			section = fmt.Sprintf("*%s-inf", config.TrimStart)
		}
		args = append(args, "--download-sections", section, "--force-keyframes-at-cuts")
	}
	if config.UseSponsorBlock {
		args = append(args, "--sponsorblock-remove", "all")
	}
	if config.Client != "" && config.Client != "Web" {
		args = append(args, "--extractor-args", fmt.Sprintf("youtube:player_client=%s", strings.ToUpper(config.Client)))
	}
	if config.CookiesPath != "" {
		args = append(args, "--cookies", config.CookiesPath)
	}
	return args
}

func (e *Engine) executeCommand(args []string, callback func(models.ProgressUpdate)) error {
	cmd := exec.Command(e.BinaryPath, args...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return err
	}

	progressRegex := regexp.MustCompile(`\[download\]\s+(\d+\.?\d*)%`)
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		matches := progressRegex.FindStringSubmatch(line)
		var percent float64
		if len(matches) > 1 {
			p, _ := strconv.ParseFloat(matches[1], 64)
			percent = p / 100.0
		}
		callback(models.ProgressUpdate{Percent: percent, Text: line, Stage: "Downloading"})
	}
	errScanner := bufio.NewScanner(stderr)
	var errOutput string
	for errScanner.Scan() {
		errOutput += errScanner.Text() + "\n"
	}
	if err := cmd.Wait(); err != nil {
		if errOutput != "" {
			return fmt.Errorf("%v | %s", err, errOutput)
		}
		return err
	}
	return nil
}
