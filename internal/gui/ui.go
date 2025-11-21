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

// Helper to format seconds into "55s" or "4m 20s" or "1h 05m 20s"
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

	// --- Bindings ---
	progressBind := binding.NewFloat()
	statusBind := binding.NewString()
	logData := ""
	statusBind.Set(locales.Get("ready"))

	// --- UI COMPONENTS ---

	// 1. INPUT
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://youtube.com/...")

	// 2. PREVIEW IMAGE
	previewImage := canvas.NewImageFromResource(theme.FileImageIcon())
	previewImage.FillMode = canvas.ImageFillContain
	previewImage.SetMinSize(fyne.NewSize(200, 120))

	var currentTitle string = "Unknown Video"

	previewTitle := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	previewTitle.Wrapping = fyne.TextWrapWord
	previewInfo := widget.NewLabel(locales.Get("ready"))
	previewInfo.Alignment = fyne.TextAlignCenter

	// 3. OPTIONS
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

	playlistCheck := widget.NewCheck("", nil)
	playlistCheck.Disable()

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

	// 4. ADVANCED
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

	cookieBtn := widget.NewButton("", func() {
		dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
			if r != nil {
				settings.CookiesPath = r.URI().Path()
				db.SaveSetting("CookiesPath", settings.CookiesPath)
			}
		}, w)
	})

	viewLogsBtn := widget.NewButton("", func() {
		content := widget.NewMultiLineEntry()
		content.SetText(logData)
		content.Disable()
		content.SetMinRowsVisible(15)
		d := dialog.NewCustom("Logs", "Close", container.NewPadded(content), w)
		d.Resize(fyne.NewSize(600, 400))
		d.Show()
	})
	viewLogsBtn.Importance = widget.LowImportance

	// --- DYNAMIC TEXTS ---
	labelQuality := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	labelSaveTo := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	labelTrimStart := widget.NewLabel("")
	labelTrimEnd := widget.NewLabel("")
	labelClient := widget.NewLabel("")

	checkBtn := widget.NewButtonWithIcon("", theme.SearchIcon(), nil)
	downloadBtn := widget.NewButtonWithIcon("", theme.DownloadIcon(), nil)
	downloadBtn.Importance = widget.HighImportance
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
		playlistCheck.SetText(locales.Get("playlist"))
		advItem.Title = locales.Get("adv_options")
		labelTrimStart.SetText(locales.Get("trim_start"))
		labelTrimEnd.SetText(locales.Get("trim_end"))
		labelClient.SetText(locales.Get("client"))
		cookieBtn.SetText(locales.Get("cookies"))
		checkSponsor.SetText(locales.Get("sponsor"))
		checkSafe.SetText(locales.Get("safe_mode"))
		viewLogsBtn.SetText(locales.Get("view_logs"))
		downloadBtn.SetText(locales.Get("btn_download"))
		updateBtn.SetText(locales.Get("update_btn"))
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

	// --- LOGIC ---
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

			// UPDATED DURATION FORMATTING
			previewInfo.SetText(fmt.Sprintf("%s • %s", meta.Uploader, formatDuration(meta.Duration)))

			if meta.ThumbnailURL != "" {
				if res, err := utils.FetchResource(meta.ThumbnailURL); err == nil {
					previewImage.Resource = res
					previewImage.Refresh()
				}
			}
			if meta.Type == "playlist" {
				playlistCheck.Enable()
				playlistCheck.SetChecked(true)
			} else {
				playlistCheck.Disable()
				playlistCheck.SetChecked(false)
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
			IsPlaylist:      playlistCheck.Checked,
		}
		db.SaveSetting("ClientSpoof", clientSelect.Selected)
		downloadBtn.Disable()
		progressBind.Set(0.0)
		logData = "Starting..."
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
				logData += update.Text + "\n"
			})
			if err != nil {
				statusBind.Set(locales.Get("failed"))
				dialog.ShowError(err, w)
			} else {
				statusBind.Set(locales.Get("success"))
				progressBind.Set(1.0)
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

	// --- LAYOUT ASSEMBLY ---

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
		playlistCheck,
	)
	unifiedCard := widget.NewCard("", "", unifiedContent)

	advContent := container.NewVBox(
		container.NewGridWithColumns(2, labelTrimStart, trimStart),
		container.NewGridWithColumns(2, labelTrimEnd, trimEnd),
		container.NewGridWithColumns(2, labelClient, clientSelect),
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

	// --- HISTORY ---
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

	// --- SETTINGS ---
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
