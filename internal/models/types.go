package models

// Set by build flags (e.g., -ldflags "-X gotube/internal/models.AppVersion=v1.5.0")
var AppVersion = "v0.0.0-dev"

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
	PlaylistItems   string
	EmbedSubs       bool
	AutoSubs        bool
	SubLanguage     string
}

// ... (Rest of the file remains the same: VideoMetadata, ProgressUpdate, etc.)
type VideoMetadata struct {
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	Uploader     string          `json:"uploader"`
	Duration     int             `json:"duration"`
	ThumbnailURL string          `json:"thumbnail"`
	Type         string          `json:"_type"`
	EntryCount   int             `json:"playlist_count"`
	Entries      []PlaylistEntry `json:"entries"`
}

type PlaylistEntry struct {
	ID    string `json:"id"`
	Title string `json:"title"`
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
