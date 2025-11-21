package models

type DownloadConfig struct {
	URL             string
	OutputPath      string
	DownloadMode    string
	Quality         string
	TrimStart       string
	TrimEnd         string
	UseSponsorBlock bool
	Client          string
	CookiesPath     string
	SafeMode        bool
	IsPlaylist      bool
}

type VideoMetadata struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Uploader     string `json:"uploader"`
	Duration     int    `json:"duration"`
	ThumbnailURL string `json:"thumbnail"`
	Type         string `json:"_type"`
	EntryCount   int    `json:"playlist_count"`
}

type ProgressUpdate struct {
	Percent float64
	Text    string
	Stage   string
}

type AppSettings struct {
	LastSavePath string
	YtDlpPath    string
	CookiesPath  string
	ClientSpoof  string
	Language     string
}

type HistoryEntry struct {
	ID        int
	Title     string
	URL       string
	FilePath  string
	Timestamp int64
}
