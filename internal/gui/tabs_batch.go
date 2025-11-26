package gui

import (
	"fmt"
	"gotube/internal/locales"
	"gotube/internal/models"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func buildBatchTab(ctx *AppContext) (fyne.CanvasObject, *widget.Button, func()) {
	// Components
	batchEntry := widget.NewMultiLineEntry()
	batchEntry.SetPlaceHolder("https://youtube.com/video1\nhttps://youtube.com/video2\n...")
	batchEntry.SetMinRowsVisible(8)

	formatSelect, detailSelect := createFormatSelectors()

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

	// Advanced (Batch Specific)
	clientSelect := widget.NewSelect([]string{"Web", "Android", "iOS"}, nil)
	clientSelect.Selected = "Web"
	checkSponsor := widget.NewCheck("", nil)
	checkSafe := widget.NewCheck("", nil)

	cookieBtn := widget.NewButton("", func() {
		dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
			if r != nil {
				ctx.Settings.CookiesPath = r.URI().Path()
				ctx.DB.SaveSetting("CookiesPath", r.URI().Path())
			}
		}, ctx.Win)
	})

	// Batch specific start button
	batchBtn := widget.NewButtonWithIcon("Start Batch", theme.MediaPlayIcon(), nil)
	batchBtn.Importance = widget.HighImportance

	// Logic
	batchBtn.OnTapped = func() {
		raw := batchEntry.Text
		if raw == "" {
			return
		}
		lines := strings.Split(raw, "\n")
		var urls []string
		for _, l := range lines {
			if strings.TrimSpace(l) != "" {
				urls = append(urls, strings.TrimSpace(l))
			}
		}
		if len(urls) == 0 {
			return
		}

		batchBtn.Disable()
		ctx.Progress.Set(0.0)

		mode := "Video"
		if formatSelect.Selected == locales.Get("format_audio") {
			mode = "Audio"
		}

		baseReq := models.DownloadConfig{
			OutputPath:      ctx.Settings.LastSavePath,
			DownloadMode:    mode,
			Quality:         detailSelect.Selected,
			Client:          clientSelect.Selected,
			CookiesPath:     ctx.Settings.CookiesPath,
			SafeMode:        checkSafe.Checked,
			IsPlaylist:      false,
			EmbedSubs:       false, // Simplified for batch
			AutoSubs:        false,
			UseSponsorBlock: checkSponsor.Checked,
		}

		go func() {
			total := float64(len(urls))
			for i, u := range urls {
				ctx.Status.Set(fmt.Sprintf("Batch: %d/%d", i+1, int(total)))

				req := baseReq
				req.URL = u

				title := u
				if meta, err := ctx.Engine.GetMetadata(u); err == nil {
					title = meta.Title
				}

				ctx.Engine.Download(req, func(update models.ProgressUpdate) {
					ctx.Logger.Write(fmt.Sprintf("[%d/%d] %s", i+1, int(total), update.Text))
				})

				ctx.DB.SaveHistory(title, u, req.OutputPath)
				ctx.Progress.Set(float64(i+1) / total)
			}
			ctx.Status.Set("Batch Complete")
			batchBtn.Enable()
		}()
	}

	// Layout
	configCard := widget.NewCard("Batch Settings", "", container.NewVBox(
		container.NewGridWithColumns(2, formatSelect, detailSelect),
		widget.NewSeparator(),
		widget.NewLabelWithStyle(locales.Get("save_to"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		pathContainer,
	))

	advContent := container.NewVBox(
		container.NewGridWithColumns(2, widget.NewLabel("Client:"), clientSelect),
		container.NewGridWithColumns(2, cookieBtn, container.NewHBox(checkSponsor, checkSafe)),
	)
	advExpander := widget.NewAccordion(widget.NewAccordionItem(locales.Get("adv_options"), advContent))

	inputCard := widget.NewCard("Input List", "", container.NewPadded(batchEntry))

	scrollContainer := container.NewVScroll(container.NewPadded(container.NewVBox(
		inputCard,
		configCard,
		advExpander,
	)))

	// Monitor accordion state to reset scroll when closed
	go func() {
		wasOpen := false
		for {
			time.Sleep(100 * time.Millisecond)
			isOpen := len(advExpander.Items) > 0 && advExpander.Items[0].Open
			if wasOpen && !isOpen {
				// Accordion just closed, reset scroll to top
				scrollContainer.Offset = fyne.NewPos(0, 0)
				scrollContainer.Refresh()
			}
			wasOpen = isOpen
		}
	}()

	content := scrollContainer

	// Updater
	updateText := func() {
		cookieBtn.SetText(locales.Get("cookies"))
		checkSponsor.SetText(locales.Get("sponsor"))
		checkSafe.SetText(locales.Get("safe_mode"))
		advExpander.Items[0].Title = locales.Get("adv_options")
		advExpander.Refresh()

		formatSelect.Options = []string{locales.Get("format_video"), locales.Get("format_audio")}
		formatSelect.Selected = locales.Get("format_video")
		formatSelect.Refresh()
	}

	return content, batchBtn, updateText
}
