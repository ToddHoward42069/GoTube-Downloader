package gui

import (
	"fmt"
	"gotube/internal/database"
	"gotube/internal/downloader"
	"gotube/internal/locales"
	"gotube/internal/models"
	"gotube/internal/updater"
	"gotube/internal/utils"
	"os"
	"strings"
	"time"

	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//go:embed icon.svg
var iconData []byte

func formatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	seconds = seconds % 60
	if minutes < 60 {
		return fmt.Sprintf("%dm %02ds", minutes, seconds)
	}
	hours := minutes / 60
	minutes = minutes % 60
	return fmt.Sprintf("%dh %02dm %02ds", hours, minutes, seconds)
}

func StartApp(a fyne.App) {
	if len(iconData) > 0 {
		a.SetIcon(fyne.NewStaticResource("icon.svg", iconData))
	}

	w := a.NewWindow("GoTube")
	w.Resize(fyne.NewSize(500, 730))

	// --- Init ---
	db, _ := database.InitDB()
	binMgr := updater.NewBinaryManager()
	engine := downloader.NewEngine(binMgr.GetYtDlpPath())
	settings := db.LoadSettings()
	if settings.LastSavePath == "" {
		settings.LastSavePath, _ = os.Getwd()
	}
	if settings.Language != "" {
		locales.SetLanguage(settings.Language)
	} else {
		locales.SetLanguage("English")
	}

	// --- Bindings & State ---
	progressBind := binding.NewFloat()
	statusBind := binding.NewString()
	statusBind.Set(locales.Get("ready"))
	logger := utils.NewLogBuffer(300)

	// Playlist State
	var currentPlaylistEntries []models.PlaylistEntry
	var selectedPlaylistIndices []string // Stored as "1", "2", etc.
	isPlaylistMode := false

	// --- UI COMPONENTS ---

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://youtube.com/...")

	previewImage := canvas.NewImageFromResource(theme.FileImageIcon())
	previewImage.FillMode = canvas.ImageFillContain
	previewImage.SetMinSize(fyne.NewSize(200, 120))
	var currentTitle string = "Unknown Video"

	previewTitle := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	previewTitle.Wrapping = fyne.TextWrapWord
	previewInfo := widget.NewLabel(locales.Get("ready"))
	previewInfo.Alignment = fyne.TextAlignCenter

	formatSelect := widget.NewSelect([]string{locales.Get("format_video"), locales.Get("format_audio")}, nil)
	formatSelect.Selected = locales.Get("format_video")
	detailSelect := widget.NewSelect([]string{"Best", "4k", "1080p", "720p"}, nil)
	detailSelect.Selected = "Best"
	formatSelect.OnChanged = func(s string) {
		if s == locales.Get("format_audio") {
			detailSelect.Options = []string{"Best", "mp3", "m4a"}
			detailSelect.Selected = "mp3"
		} else {
			detailSelect.Options = []string{"Best", "4k", "1080p", "720p"}
			detailSelect.Selected = "Best"
		}
		detailSelect.Refresh()
	}

	// PLAYLIST SELECTOR BUTTON
	playlistBtn := widget.NewButton(locales.Get("pl_select_btn"), nil)
	playlistBtn.Disable() // Enabled only when playlist detected

	// Playlist Dialog Logic
	playlistBtn.OnTapped = func() {
		if len(currentPlaylistEntries) == 0 {
			return
		}

		// Temporary state for the dialog
		selectionState := make(map[int]bool)
		for i := range currentPlaylistEntries {
			selectionState[i] = true
		} // Default select all

		labelCount := widget.NewLabel(fmt.Sprintf(locales.Get("pl_selected"), len(currentPlaylistEntries)))

		// Scrollable List of Checkboxes
		// We use a VBox inside a Scroll, populated manually for simplicity since lists inside dialogs can be tricky
		// For massive playlists (500+), a virtual List is better, but for <100, a VBox is fine.
		// Let's use a Virtual List for performance.

		vList := widget.NewList(
			func() int { return len(currentPlaylistEntries) },
			func() fyne.CanvasObject {
				return widget.NewCheck("Video Title", nil)
			},
			func(i int, o fyne.CanvasObject) {
				check := o.(*widget.Check)
				check.Text = fmt.Sprintf("%d. %s", i+1, currentPlaylistEntries[i].Title)
				check.Checked = selectionState[i]
				check.OnChanged = func(b bool) {
					selectionState[i] = b
					// Update count label
					count := 0
					for _, v := range selectionState {
						if v {
							count++
						}
					}
					labelCount.SetText(fmt.Sprintf(locales.Get("pl_selected"), count))
				}
				check.Refresh() // Critical for recycling
			},
		)

		// Dialog Buttons
		btnAll := widget.NewButton(locales.Get("pl_select_all"), func() {
			for i := range currentPlaylistEntries {
				selectionState[i] = true
			}
			vList.Refresh()
			labelCount.SetText(fmt.Sprintf(locales.Get("pl_selected"), len(currentPlaylistEntries)))
		})
		btnNone := widget.NewButton(locales.Get("pl_select_none"), func() {
			for i := range currentPlaylistEntries {
				selectionState[i] = false
			}
			vList.Refresh()
			labelCount.SetText(fmt.Sprintf(locales.Get("pl_selected"), 0))
		})

		content := container.NewBorder(
			container.NewVBox(labelCount, container.NewHBox(btnAll, btnNone)),
			nil, nil, nil,
			vList,
		)

		d := dialog.NewCustomConfirm(locales.Get("pl_title"), locales.Get("pl_confirm"), "Cancel", content, func(b bool) {
			if b {
				// Commit Selection
				selectedPlaylistIndices = []string{}
				count := 0
				for i, selected := range selectionState {
					if selected {
						selectedPlaylistIndices = append(selectedPlaylistIndices, fmt.Sprintf("%d", i+1)) // 1-based index
						count++
					}
				}
				// Update Main Button Text
				playlistBtn.SetText(fmt.Sprintf("%s (%d)", locales.Get("pl_select_btn"), count))
			}
		}, w)
		d.Resize(fyne.NewSize(400, 600))
		d.Show()
	}

	pathEntry := widget.NewEntry()
	pathEntry.SetText(settings.LastSavePath)
	pathEntry.Disable()
	pathBtn := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if uri != nil {
				pathEntry.SetText(uri.Path())
				settings.LastSavePath = uri.Path()
				db.SaveSetting("LastSavePath", settings.LastSavePath)
			}
		}, w)
	})
	pathContainer := container.NewBorder(nil, nil, nil, pathBtn, pathEntry)

	trimStart := widget.NewEntry()
	trimStart.SetPlaceHolder("00:00:00")
	trimEnd := widget.NewEntry()
	trimEnd.SetPlaceHolder("00:00:00")
	clientSelect := widget.NewSelect([]string{"Web", "Android", "iOS"}, nil)
	clientSelect.Selected = "Web"
	if settings.ClientSpoof != "" {
		clientSelect.Selected = settings.ClientSpoof
	}
	checkSponsor := widget.NewCheck("", nil)
	checkSafe := widget.NewCheck("", nil)

	checkEmbedSubs := widget.NewCheck("", nil)
	checkAutoSubs := widget.NewCheck("", nil)
	checkAutoSubs.Disable()
	checkEmbedSubs.OnChanged = func(b bool) {
		if b {
			checkAutoSubs.Enable()
		} else {
			checkAutoSubs.Disable()
		}
	}
	subLangSelect := widget.NewSelect([]string{"en", "de", "all"}, nil)
	subLangSelect.Selected = "en"

	cookieBtn := widget.NewButton("", func() {
		dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
			if r != nil {
				settings.CookiesPath = r.URI().Path()
				db.SaveSetting("CookiesPath", settings.CookiesPath)
			}
		}, w)
	})

	consoleEntry := widget.NewMultiLineEntry()
	consoleEntry.Disable()
	consoleEntry.TextStyle = fyne.TextStyle{Monospace: true}
	consoleEntry.SetPlaceHolder("Waiting for download...")

	labelQuality := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	labelSaveTo := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	labelTrimStart := widget.NewLabel("")
	labelTrimEnd := widget.NewLabel("")
	labelClient := widget.NewLabel("")
	labelSubLang := widget.NewLabel("")

	checkBtn := widget.NewButtonWithIcon("", theme.SearchIcon(), nil)
	downloadBtn := widget.NewButtonWithIcon("", theme.DownloadIcon(), nil)
	downloadBtn.Importance = widget.HighImportance

	viewLogsBtn := widget.NewButton("", func() {
		d := dialog.NewCustom("Live Console", "Close", container.NewPadded(consoleEntry), w)
		d.Resize(fyne.NewSize(700, 500))
		d.Show()
	})
	viewLogsBtn.Importance = widget.LowImportance

	updateBtn := widget.NewButton("", nil)

	advItem := widget.NewAccordionItem("", nil)
	langSelect := widget.NewSelect([]string{"English", "German"}, nil)
	if settings.Language == "German" {
		langSelect.Selected = "German"
	} else {
		langSelect.Selected = "English"
	}

	updateTexts := func() {
		urlEntry.SetPlaceHolder(locales.Get("placeholder"))
		labelQuality.SetText(locales.Get("quality"))
		labelSaveTo.SetText(locales.Get("save_to"))

		// Playlist button text
		if isPlaylistMode {
			playlistBtn.SetText(fmt.Sprintf("%s (%d)", locales.Get("pl_select_btn"), len(selectedPlaylistIndices)))
		} else {
			playlistBtn.SetText(locales.Get("pl_select_btn"))
		}

		advItem.Title = locales.Get("adv_options")
		labelTrimStart.SetText(locales.Get("trim_start"))
		labelTrimEnd.SetText(locales.Get("trim_end"))
		labelClient.SetText(locales.Get("client"))
		cookieBtn.SetText(locales.Get("cookies"))
		checkSponsor.SetText(locales.Get("sponsor"))
		checkSafe.SetText(locales.Get("safe_mode"))
		downloadBtn.SetText(locales.Get("btn_download"))
		updateBtn.SetText(locales.Get("update_btn"))
		viewLogsBtn.SetText(locales.Get("view_logs"))
		checkEmbedSubs.SetText(locales.Get("subs_embed"))
		checkAutoSubs.SetText(locales.Get("subs_auto"))
		labelSubLang.SetText(locales.Get("subs_lang"))
	}

	langSelect.OnChanged = func(s string) {
		locales.SetLanguage(s)
		db.SaveSetting("Language", s)
		updateTexts()
		formatSelect.Options = []string{locales.Get("format_video"), locales.Get("format_audio")}
		formatSelect.Selected = locales.Get("format_video")
		formatSelect.Refresh()
		advItem.Detail.Refresh()
	}
	updateTexts()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		for range ticker.C {
			if logger.HasChanged() {
				text := logger.String()
				logger.MarkRead()
				consoleEntry.SetText(text)
				consoleEntry.CursorRow = len(text)
				consoleEntry.Refresh()
			}
		}
	}()

	var fetchTimer *time.Timer
	performFetch := func(url string) {
		if url == "" || !strings.HasPrefix(url, "http") {
			return
		}
		statusBind.Set(locales.Get("fetching"))
		go func() {
			meta, err := engine.GetMetadata(url)
			if err != nil {
				statusBind.Set("Error: " + err.Error())
				return
			}
			currentTitle = meta.Title
			statusBind.Set(locales.Get("meta_loaded"))
			previewTitle.SetText(meta.Title)

			// Display info
			if meta.Type == "playlist" {
				isPlaylistMode = true
				currentPlaylistEntries = meta.Entries
				previewInfo.SetText(fmt.Sprintf("Playlist • %d Videos", meta.EntryCount))

				// Reset Selection
				selectedPlaylistIndices = []string{}
				for i := 1; i <= meta.EntryCount; i++ {
					selectedPlaylistIndices = append(selectedPlaylistIndices, fmt.Sprintf("%d", i))
				}

				playlistBtn.Enable()
				playlistBtn.SetText(fmt.Sprintf("%s (%d)", locales.Get("pl_select_btn"), meta.EntryCount))
			} else {
				isPlaylistMode = false
				currentPlaylistEntries = nil
				previewInfo.SetText(fmt.Sprintf("%s • %s", meta.Uploader, formatDuration(meta.Duration)))
				playlistBtn.Disable()
				playlistBtn.SetText(locales.Get("pl_select_btn"))
			}

			if meta.ThumbnailURL != "" {
				if res, err := utils.FetchResource(meta.ThumbnailURL); err == nil {
					previewImage.Resource = res
					previewImage.Refresh()
				}
			}
		}()
	}

	urlEntry.OnChanged = func(s string) {
		if fetchTimer != nil {
			fetchTimer.Stop()
		}
		fetchTimer = time.AfterFunc(500*time.Millisecond, func() { performFetch(s) })
	}
	checkBtn.OnTapped = func() { performFetch(urlEntry.Text) }

	downloadBtn.OnTapped = func() {
		if urlEntry.Text == "" {
			return
		}
		mode := "Video"
		if formatSelect.Selected == locales.Get("format_audio") {
			mode = "Audio"
		}

		// Build Index String
		idxStr := ""
		if isPlaylistMode && len(selectedPlaylistIndices) > 0 {
			idxStr = strings.Join(selectedPlaylistIndices, ",")
		}

		req := models.DownloadConfig{
			URL:             urlEntry.Text,
			OutputPath:      settings.LastSavePath,
			DownloadMode:    mode,
			Quality:         detailSelect.Selected,
			TrimStart:       trimStart.Text,
			TrimEnd:         trimEnd.Text,
			UseSponsorBlock: checkSponsor.Checked,
			Client:          clientSelect.Selected,
			CookiesPath:     settings.CookiesPath,
			SafeMode:        checkSafe.Checked,
			IsPlaylist:      isPlaylistMode,
			PlaylistItems:   idxStr,
			EmbedSubs:       checkEmbedSubs.Checked,
			AutoSubs:        checkAutoSubs.Checked,
			SubLanguage:     subLangSelect.Selected,
		}
		db.SaveSetting("ClientSpoof", clientSelect.Selected)
		downloadBtn.Disable()
		progressBind.Set(0.0)

		logger.Clear()
		logger.Write("Starting download...")

		go func() {
			if currentTitle == "Unknown Video" {
				if meta, err := engine.GetMetadata(req.URL); err == nil {
					currentTitle = meta.Title
				}
			}

			err := engine.Download(req, func(update models.ProgressUpdate) {
				if update.Percent > 0 {
					progressBind.Set(update.Percent)
				}
				statusBind.Set(update.Stage + "...")
				logger.Write(update.Text)
			})

			if err != nil {
				statusBind.Set(locales.Get("failed"))
				logger.Write("ERROR: " + err.Error())
				dialog.ShowError(err, w)
			} else {
				statusBind.Set(locales.Get("success"))
				progressBind.Set(1.0)
				logger.Write("SUCCESS: Download finished.")
				db.SaveHistory(currentTitle, req.URL, req.OutputPath)
				currentTitle = "Unknown Video"
			}
			downloadBtn.Enable()
		}()
	}
	updateBtn.OnTapped = func() {
		p := dialog.NewProgressInfinite("Updating", "Checking GitHub...", w)
		p.Show()
		go func() {
			err := binMgr.UpdateBinary(func(msg string) { fmt.Println(msg) })
			p.Hide()
			if err != nil {
				dialog.ShowError(err, w)
			} else {
				dialog.ShowInformation("Success", "Core updated.", w)
			}
		}()
	}

	topRow := container.NewBorder(nil, nil, nil, checkBtn, urlEntry)

	unifiedContent := container.NewVBox(
		container.NewCenter(previewImage),
		previewTitle,
		previewInfo,
		widget.NewSeparator(),
		labelQuality,
		container.NewGridWithColumns(2, formatSelect, detailSelect),
		labelSaveTo,
		pathContainer,
		// Playlist Btn instead of Check
		playlistBtn,
	)
	unifiedCard := widget.NewCard("", "", unifiedContent)

	advContent := container.NewVBox(
		container.NewGridWithColumns(2, labelTrimStart, trimStart),
		container.NewGridWithColumns(2, labelTrimEnd, trimEnd),
		container.NewGridWithColumns(2, labelClient, clientSelect),
		container.NewGridWithColumns(2, labelSubLang, subLangSelect),
		container.NewGridWithColumns(2, checkEmbedSubs, checkAutoSubs),
		container.NewGridWithColumns(2, cookieBtn, container.NewHBox(checkSponsor, checkSafe)),
	)
	advItem.Detail = advContent
	advExpander := widget.NewAccordion(advItem)

	formContent := container.NewVBox(
		topRow,
		unifiedCard,
		advExpander,
		layout.NewSpacer(),
	)

	statusLabel := widget.NewLabelWithData(statusBind)
	statusLabel.Alignment = fyne.TextAlignCenter
	progressContainer := container.NewPadded(widget.NewProgressBarWithData(progressBind))

	footer := container.NewVBox(
		widget.NewSeparator(),
		statusLabel,
		progressContainer,
		container.NewGridWithColumns(2, viewLogsBtn, downloadBtn),
	)

	mainLayout := container.NewBorder(
		nil,
		container.NewPadded(footer),
		nil, nil,
		container.NewVScroll(container.NewPadded(formContent)),
	)

	historyList := widget.NewList(
		func() int { return len(db.GetHistory()) },
		func() fyne.CanvasObject {
			icon := widget.NewIcon(theme.MediaPlayIcon())
			title := widget.NewLabel("Title")
			title.TextStyle = fyne.TextStyle{Bold: true}
			title.Truncation = fyne.TextTruncateEllipsis
			btn := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), nil)
			content := container.NewBorder(nil, nil, icon, btn, title)
			return widget.NewCard("", "", content)
		},
		func(i int, o fyne.CanvasObject) {
			h := db.GetHistory()[i]
			card := o.(*widget.Card)
			border := card.Content.(*fyne.Container)
			border.Objects[0].(*widget.Label).SetText(h.Title)
			border.Objects[2].(*widget.Button).OnTapped = func() { utils.OpenFolder(h.FilePath) }
		},
	)

	settingsContent := container.NewVBox(
		widget.NewCard(locales.Get("tab_system"), "", container.NewVBox(
			widget.NewLabel("Language / Sprache"),
			langSelect,
			widget.NewSeparator(),
			widget.NewLabel("Core Binary: "+binMgr.GetYtDlpPath()),
			updateBtn,
		)),
	)

	t1 := container.NewTabItemWithIcon(locales.Get("tab_download"), theme.DownloadIcon(), mainLayout)
	t2 := container.NewTabItemWithIcon(locales.Get("tab_history"), theme.HistoryIcon(), historyList)
	t3 := container.NewTabItemWithIcon(locales.Get("tab_system"), theme.SettingsIcon(), container.NewPadded(settingsContent))

	tabs := container.NewAppTabs(t1, t2, t3)

	originalUpdateTexts := updateTexts
	updateTexts = func() {
		originalUpdateTexts()
		t1.Text = locales.Get("tab_download")
		t2.Text = locales.Get("tab_history")
		t3.Text = locales.Get("tab_system")
		tabs.Refresh()
	}

	w.SetContent(tabs)
	w.ShowAndRun()
}
