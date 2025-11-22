package utils

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"fyne.io/fyne/v2"
)

// FetchResource downloads an image and converts YouTube WEBP links to JPG
func FetchResource(urlStr string) (fyne.Resource, error) {
	// FIX: Fyne cannot decode WEBP images by default.
	// YouTube links usually look like: https://i.ytimg.com/vi_webp/ID/maxresdefault.webp
	// We rewrite them to:            https://i.ytimg.com/vi/ID/maxresdefault.jpg
	if strings.Contains(urlStr, "ytimg.com") {
		urlStr = strings.Replace(urlStr, "/vi_webp/", "/vi/", 1)
		urlStr = strings.Replace(urlStr, ".webp", ".jpg", 1)
	}

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: %s", resp.Status)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// We name it .jpg so Fyne's decoder knows what to expect
	return fyne.NewStaticResource("thumbnail.jpg", bytes), nil
}
