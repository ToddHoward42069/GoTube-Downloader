package utils

import (
	"fyne.io/fyne/v2"
	"io"
	"net/http"
	"time"
)

func FetchResource(urlStr string) (fyne.Resource, error) {
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(urlStr)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil { return nil, err }

	return fyne.NewStaticResource("thumbnail.jpg", bytes), nil
}
