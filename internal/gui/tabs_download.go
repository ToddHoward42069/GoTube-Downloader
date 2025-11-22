package gui

import (
	"fmt"
	"gotube/internal/locales"
	"gotube/internal/models"
	"gotube/internal/utils"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Returns: Content, ActionButton, UpdateFunc
func buildMainTab(ctx *AppContext) (fyne.CanvasObject, *widget.Button, func()) {
	// Components
	urlEntry := widget.NewEntry()
	previewImage := createPreviewImage()

	var currentTitle string = "Unknown Video"
	var currentPlEntries []models.PlaylistEntry
	var selectedPlIndices []string
	isPlMode := false

	previewTitle := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	previewTitle.Wrapping = fyne.TextWrapWord
	previewInfo := widget.NewLabel(locales.Get("ready"))
	previewInfo.Alignment = fyne.TextAlignCenter

	formatSelect, detailSelect := createFormatSelectors()

	playlistBtn := widget.NewButton(locales.Get("pl_select_btn"), nil)
	playlistBtn.Disable()

	pathEntry := widget.NewEntry()
	pathEntry.SetText(ctx.Settings.LastSavePath)
	pathEntry.Disable()
	pathBtn := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if uri != nil {
				pathEntry.SetText(uri.Path())
				ctx.Settings.LastSavePath = uri.Path()
				ctx.DB.SaveSetting("LastSavePath", uri.Path())
			}
		}, ctx.Win)
	})
	pathContainer := container.NewBorder(nil, nil, nil, pathBtn, pathEntry)

	// Advanced
	trimStart := widget.NewEntry()
	trimStart.SetPlaceHolder("00:00:00")
	trimEnd := widget.NewEntry()
	trimEnd.SetPlaceHolder("00:00:00")

	clientSelect := widget.NewSelect([]string{"Web", "Android", "iOS"}, nil)
	clientSelect.Selected = "Web"
	if ctx.Settings.ClientSpoof != "" {
		clientSelect.Selected = ctx.Settings.ClientSpoof
	}

	checkSponsor := widget.NewCheck("", nil)
	checkSafe := widget.NewCheck("", nil)

	checkEmbed := widget.NewCheck("", nil)
	checkAuto := widget.NewCheck("", nil)
	checkAuto.Disable()
	checkEmbed.OnChanged = func(b bool) {
		if b {
			checkAuto.Enable()
		} else {
			checkAuto.Disable()
		}
	}

	subLang := widget.NewSelect([]string{"en", "de", "all"}, nil)
	subLang.Selected = "en"

	cookieBtn := widget.NewButton("", func() {
		dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
			if r != nil {
				ctx.Settings.CookiesPath = r.URI().Path()
				ctx.DB.SaveSetting("CookiesPath", ctx.Settings.CookiesPath)
			}
		}, ctx.Win)
	})

	// Dynamic Labels
	labelQuality := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	labelSaveTo := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	labelTrimStart := widget.NewLabel("")
	labelTrimEnd := widget.NewLabel("")
	labelClient := widget.NewLabel("")
	labelSubLang := widget.NewLabel("")

	// Logic
	playlistBtn.OnTapped = func() {
		if len(currentPlEntries) == 0 {
			return
		}
		selectionState := make(map[int]bool)
		for i := range currentPlEntries {
			selectionState[i] = true
		}

		labelCount := widget.NewLabel(fmt.Sprintf(locales.Get("pl_selected"), len(currentPlEntries)))

		vList := widget.NewList(
			func() int { return len(currentPlEntries) },
			func() fyne.CanvasObject { return widget.NewCheck("Video Title", nil) },
			func(i int, o fyne.CanvasObject) {
				check := o.(*widget.Check)
				check.Text = fmt.Sprintf("%d. %s", i+1, currentPlEntries[i].Title)
				check.Checked = selectionState[i]
				check.OnChanged = func(b bool) {
					selectionState[i] = b
					count := 0
					for _, v := range selectionState {
						if v {
							count++
						}
					}
					labelCount.SetText(fmt.Sprintf(locales.Get("pl_selected"), count))
				}
				check.Refresh()
			},
		)

		btnAll := widget.NewButton(locales.Get("pl_select_all"), func() {
			for i := range currentPlEntries {
				selectionState[i] = true
			}
			vList.Refresh()
			labelCount.SetText(fmt.Sprintf(locales.Get("pl_selected"), len(currentPlEntries)))
		})
		btnNone := widget.NewButton(locales.Get("pl_select_none"), func() {
			for i := range currentPlEntries {
				selectionState[i] = false
			}
			vList.Refresh()
			labelCount.SetText(fmt.Sprintf(locales.Get("pl_selected"), 0))
		})

		content := container.NewBorder(container.NewVBox(labelCount, container.NewHBox(btnAll, btnNone)), nil, nil, nil, vList)

		d := dialog.NewCustomConfirm(locales.Get("pl_title"), locales.Get("pl_confirm"), "Cancel", content, func(b bool) {
			if b {
				selectedPlIndices = []string{}
				count := 0
				for i, selected := range selectionState {
					if selected {
						selectedPlIndices = append(selectedPlIndices, fmt.Sprintf("%d", i+1))
						count++
					}
				}
				playlistBtn.SetText(fmt.Sprintf("%s (%d)", locales.Get("pl_select_btn"), count))
			}
		}, ctx.Win)
		d.Resize(fyne.NewSize(400, 600))
		d.Show()
	}

	var fetchTimer *time.Timer
	performFetch := func(url string) {
		if url == "" || !strings.HasPrefix(url, "http") {
			return
		}
		ctx.Status.Set(locales.Get("fetching"))
		go func() {
			meta, err := ctx.Engine.GetMetadata(url)
			if err != nil {
				ctx.Status.Set("Error: " + err.Error())
				return
			}
			currentTitle = meta.Title
			ctx.Status.Set(locales.Get("meta_loaded"))
			previewTitle.SetText(meta.Title)

			if meta.Type == "playlist" {
				isPlMode = true
				currentPlEntries = meta.Entries
				previewInfo.SetText(fmt.Sprintf("Playlist • %d Videos", meta.EntryCount))
				selectedPlIndices = []string{}
				for i := 1; i <= meta.EntryCount; i++ {
					selectedPlIndices = append(selectedPlIndices, fmt.Sprintf("%d", i))
				}
				playlistBtn.Enable()
				playlistBtn.SetText(fmt.Sprintf("%s (%d)", locales.Get("pl_select_btn"), meta.EntryCount))
			} else {
				isPlMode = false
				currentPlEntries = nil
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
	checkBtn := widget.NewButtonWithIcon("", theme.SearchIcon(), func() { performFetch(urlEntry.Text) })

	var downloadBtn *widget.Button
	downloadBtn = widget.NewButtonWithIcon("", theme.DownloadIcon(), func() {
		if urlEntry.Text == "" {
			return
		}

		mode := "Video"
		if formatSelect.Selected == locales.Get("format_audio") {
			mode = "Audio"
		}
		idxStr := ""
		if isPlMode && len(selectedPlIndices) > 0 {
			idxStr = strings.Join(selectedPlIndices, ",")
		}

		req := models.DownloadConfig{
			URL:             urlEntry.Text,
			OutputPath:      ctx.Settings.LastSavePath,
			DownloadMode:    mode,
			Quality:         detailSelect.Selected,
			TrimStart:       trimStart.Text,
			TrimEnd:         trimEnd.Text,
			UseSponsorBlock: checkSponsor.Checked,
			Client:          clientSelect.Selected,
			CookiesPath:     ctx.Settings.CookiesPath,
			SafeMode:        checkSafe.Checked,
			IsPlaylist:      isPlMode,
			PlaylistItems:   idxStr,
			EmbedSubs:       checkEmbed.Checked,
			AutoSubs:        checkAuto.Checked,
			SubLanguage:     subLang.Selected,
		}
		ctx.DB.SaveSetting("ClientSpoof", clientSelect.Selected)

		downloadBtn.Disable()
		ctx.Progress.Set(0.0)
		ctx.Logger.Clear()
		ctx.Logger.Write("Starting download...")

		go func() {
			if currentTitle == "Unknown Video" {
				if meta, err := ctx.Engine.GetMetadata(req.URL); err == nil {
					currentTitle = meta.Title
				}
			}

			err := ctx.Engine.Download(req, func(update models.ProgressUpdate) {
				if update.Percent > 0 {
					ctx.Progress.Set(update.Percent)
				}
				ctx.Status.Set(update.Stage + "...")
				ctx.Logger.Write(update.Text)
			})

			if err != nil {
				ctx.Status.Set(locales.Get("failed"))
				ctx.Logger.Write("ERROR: " + err.Error())
				dialog.ShowError(err, ctx.Win)
			} else {
				ctx.Status.Set(locales.Get("success"))
				ctx.Progress.Set(1.0)
				ctx.Logger.Write("SUCCESS: Download finished.")
				ctx.DB.SaveHistory(currentTitle, req.URL, req.OutputPath)
				currentTitle = "Unknown Video"
			}
			downloadBtn.Enable()
		}()
	})
	downloadBtn.Importance = widget.HighImportance

	// Layout Assembly
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
		playlistBtn,
	)
	unifiedCard := widget.NewCard("", "", unifiedContent)

	advContent := container.NewVBox(
		container.NewGridWithColumns(2, labelTrimStart, trimStart),
		container.NewGridWithColumns(2, labelTrimEnd, trimEnd),
		container.NewGridWithColumns(2, labelClient, clientSelect),
		container.NewGridWithColumns(2, labelSubLang, subLang),
		container.NewGridWithColumns(2, checkEmbed, checkAuto),
		container.NewGridWithColumns(2, cookieBtn, container.NewHBox(checkSponsor, checkSafe)),
	)
	advExpander := widget.NewAccordion(widget.NewAccordionItem("", advContent))

	formContent := container.NewVBox(
		topRow,
		unifiedCard,
		advExpander,
		layout.NewSpacer(),
	)

	// Return content without the button (button is returned separately)
	content := container.NewVScroll(container.NewPadded(formContent))

	// Updater closure
	updateText := func() {
		urlEntry.SetPlaceHolder(locales.Get("placeholder"))
		labelQuality.SetText(locales.Get("quality"))
		labelSaveTo.SetText(locales.Get("save_to"))
		if isPlMode {
			playlistBtn.SetText(fmt.Sprintf("%s (%d)", locales.Get("pl_select_btn"), len(selectedPlIndices)))
		} else {
			playlistBtn.SetText(locales.Get("pl_select_btn"))
		}
		advExpander.Items[0].Title = locales.Get("adv_options")
		advExpander.Refresh()
		labelTrimStart.SetText(locales.Get("trim_start"))
		labelTrimEnd.SetText(locales.Get("trim_end"))
		labelClient.SetText(locales.Get("client"))
		cookieBtn.SetText(locales.Get("cookies"))
		checkSponsor.SetText(locales.Get("sponsor"))
		checkSafe.SetText(locales.Get("safe_mode"))
		downloadBtn.SetText(locales.Get("btn_download"))
		checkEmbed.SetText(locales.Get("subs_embed"))
		checkAuto.SetText(locales.Get("subs_auto"))
		labelSubLang.SetText(locales.Get("subs_lang"))

		// Select Options
		formatSelect.Options = []string{locales.Get("format_video"), locales.Get("format_audio")}
		formatSelect.Selected = locales.Get("format_video")
		formatSelect.Refresh()
	}

	return content, downloadBtn, updateText
}
