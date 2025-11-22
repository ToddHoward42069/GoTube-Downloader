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
	"subs_embed":   "Embed Subtitles",
	"subs_auto":    "Auto-Generated",
	"subs_lang":    "Language:",

	// Playlist Selector
	"pl_select_btn":  "Select Videos",
	"pl_title":       "Select Videos to Download",
	"pl_selected":    "Selected: %d",
	"pl_confirm":     "Confirm",
	"pl_select_all":  "Select All",
	"pl_select_none": "Select None",
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
	"subs_embed":   "Untertitel einbetten",
	"subs_auto":    "Automatisch generiert",
	"subs_lang":    "Sprache:",

	// Playlist Selector
	"pl_select_btn":  "Videos auswählen",
	"pl_title":       "Videos auswählen",
	"pl_selected":    "Ausgewählt: %d",
	"pl_confirm":     "Bestätigen",
	"pl_select_all":  "Alle",
	"pl_select_none": "Keine",
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
	if !ok {
		return key
	}
	return val
}
