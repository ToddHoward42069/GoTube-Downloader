package locales

var currentLang = "en"

var en = map[string]string{
	"tab_download": "Download",
	"tab_history":  "History",
	"tab_system":   "System",
	"placeholder":  "Paste YouTube Link...",
	"check":        "Check",
	"quality":      "Quality",
	"mode":         "Mode",
	"playlist":     "Playlist Mode",
	"save_to":      "Save To",
	"adv_options":  "Advanced Options",
	"trim_start":   "Trim Start:",
	"trim_end":     "Trim End:",
	"client":       "Client:",
	"auth":         "Auth:",
	"cookies":      "Load Cookies",
	"sponsor":      "SponsorBlock",
	"safe_mode":    "Safe Mode",
	"view_logs":    "View Logs",
	"btn_download": "Download",
	"update_btn":   "Check for Updates",
	"ready":        "Ready",
	"fetching":     "Fetching info...",
	"meta_loaded":  "Metadata loaded",
	"success":      "Download Complete",
	"failed":       "Failed",
	"format_video": "Video (MP4)",
	"format_audio": "Audio",
}

var de = map[string]string{
	"tab_download": "Herunterladen",
	"tab_history":  "Verlauf",
	"tab_system":   "System",
	"placeholder":  "YouTube Link einfügen...",
	"check":        "Prüfen",
	"quality":      "Qualität",
	"mode":         "Modus",
	"playlist":     "Playlist-Modus",
	"save_to":      "Speichern unter",
	"adv_options":  "Erweiterte Optionen",
	"trim_start":   "Startzeit:",
	"trim_end":     "Endzeit:",
	"client":       "Klient:",
	"auth":         "Auth:",
	"cookies":      "Cookies laden",
	"sponsor":      "SponsorBlock",
	"safe_mode":    "Sicherer Modus",
	"view_logs":    "Protokolle",
	"btn_download": "Starten",
	"update_btn":   "Auf Updates prüfen",
	"ready":        "Bereit",
	"fetching":     "Lade Infos...",
	"meta_loaded":  "Metadaten geladen",
	"success":      "Download abgeschlossen",
	"failed":       "Fehlgeschlagen",
	"format_video": "Video (MP4)",
	"format_audio": "Audio",
}

func SetLanguage(lang string) {
	if lang == "German" || lang == "de" {
		currentLang = "de"
	} else {
		currentLang = "en"
	}
}

func Get(key string) string {
	dict := en
	if currentLang == "de" {
		dict = de
	}
	val, ok := dict[key]
	if !ok { return key }
	return val
}
