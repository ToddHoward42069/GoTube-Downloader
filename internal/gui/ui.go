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
	"time"

	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//go:embed icon.svg
var iconData []byte

type AppContext struct {
	App      fyne.App
	Win      fyne.Window
	DB       *database.DB
	Engine   *downloader.Engine
	BinMgr   *updater.BinaryManager
	Settings models.AppSettings
	Status   binding.String
	Progress binding.Float
	Logger   *utils.LogBuffer
	Console  *widget.Entry
}

func StartApp(a fyne.App) {
	if len(iconData) > 0 {
		a.SetIcon(fyne.NewStaticResource("icon.svg", iconData))
	}
	w := a.NewWindow("GoTube " + models.AppVersion) // Show version in title
	w.Resize(fyne.NewSize(500, 730))

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

	ctx := &AppContext{
		App:      a,
		Win:      w,
		DB:       db,
		Engine:   engine,
		BinMgr:   binMgr,
		Settings: settings,
		Status:   binding.NewString(),
		Progress: binding.NewFloat(),
		Logger:   utils.NewLogBuffer(300),
	}
	ctx.Status.Set(locales.Get("ready"))

	// Build Tabs
	mainTab, mainBtn, mainUpdate := buildMainTab(ctx)
	batchTab, batchBtn, batchUpdate := buildBatchTab(ctx)
	historyTab := buildHistoryTab(ctx)
	settingsTab := buildSettingsTab(ctx)

	// Footer
	viewLogsBtn := widget.NewButton("", func() { showLogs(ctx) })
	statusLabel := widget.NewLabelWithData(ctx.Status)
	statusLabel.Alignment = fyne.TextAlignCenter
	progressContainer := container.NewPadded(widget.NewProgressBarWithData(ctx.Progress))

	footer1 := container.NewVBox(widget.NewSeparator(), statusLabel, progressContainer, container.NewGridWithColumns(2, viewLogsBtn, mainBtn))

	// Button for Batch Footer (Needs clone)
	viewLogsBtn2 := widget.NewButton("", func() { showLogs(ctx) })
	footer2 := container.NewVBox(widget.NewSeparator(), statusLabel, progressContainer, container.NewGridWithColumns(2, viewLogsBtn2, batchBtn))

	t1Content := container.NewBorder(nil, container.NewPadded(footer1), nil, nil, mainTab)
	t2Content := container.NewBorder(nil, container.NewPadded(footer2), nil, nil, batchTab)

	t1 := container.NewTabItemWithIcon(locales.Get("tab_download"), theme.DownloadIcon(), t1Content)
	t2 := container.NewTabItemWithIcon("Batch", theme.ListIcon(), t2Content)
	t3 := container.NewTabItemWithIcon(locales.Get("tab_history"), theme.HistoryIcon(), historyTab)
	t4 := container.NewTabItemWithIcon(locales.Get("tab_system"), theme.SettingsIcon(), settingsTab)

	tabs := container.NewAppTabs(t1, t2, t3, t4)

	updateAllTexts := func() {
		mainUpdate()
		batchUpdate()
		t1.Text = locales.Get("tab_download")
		t3.Text = locales.Get("tab_history")
		t4.Text = locales.Get("tab_system")
		viewLogsBtn.SetText(locales.Get("view_logs"))
		viewLogsBtn2.SetText(locales.Get("view_logs"))
		tabs.Refresh()
	}
	updateAllTexts()

	ctx.App.Metadata().Custom["updateTexts"] = "true"

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		ctx.Console = widget.NewMultiLineEntry()
		ctx.Console.TextStyle = fyne.TextStyle{Monospace: true}
		for range ticker.C {
			if ctx.Logger.HasChanged() {
				text := ctx.Logger.String()
				ctx.Logger.MarkRead()
				ctx.Console.SetText(text)
				ctx.Console.CursorRow = len(text)
				ctx.Console.Refresh()
			}
		}
	}()

	// Inject Updater into Settings Tab
	t4.Content = buildSettingsTabWithCallback(ctx, updateAllTexts)

	// --- AUTO UPDATE CHECK ON STARTUP ---
	go func() {
		// Wait a second for UI to render
		time.Sleep(1 * time.Second)
		newVer, downloadUrl, err := updater.CheckAppUpdate()
		if err == nil && newVer != "" {
			dialog.ShowConfirm("Update Available",
				"Version "+newVer+" is available. Update now?",
				func(b bool) {
					if b {
						performAppUpdate(ctx, downloadUrl)
					}
				}, w)
		}
	}()

	w.SetContent(tabs)
	w.ShowAndRun()
}

// Helper to run the update process with UI feedback
func performAppUpdate(ctx *AppContext, url string) {
	p := dialog.NewProgress("Updating App", "Downloading...", ctx.Win)
	p.Show()

	go func() {
		err := updater.DoAppUpdate(url, func(f float64) {
			p.SetValue(f)
		})
		p.Hide()

		if err != nil {
			dialog.ShowError(err, ctx.Win)
		} else {
			dialog.ShowInformation("Success", "Update complete. The app will restart.", ctx.Win)
			time.Sleep(2 * time.Second)
			updater.RestartApp()
		}
	}()
}

func buildSettingsTabWithCallback(ctx *AppContext, updateFunc func()) fyne.CanvasObject {
	langSelect := widget.NewSelect([]string{"English", "German"}, nil)
	if ctx.Settings.Language == "German" {
		langSelect.Selected = "German"
	} else {
		langSelect.Selected = "English"
	}

	// Button to update yt-dlp (Core)
	updateCoreBtn := widget.NewButton("Update Core (yt-dlp)", func() {
		p := dialog.NewProgressInfinite("Updating", "Checking GitHub...", ctx.Win)
		p.Show()
		go func() {
			err := ctx.BinMgr.UpdateBinary(func(msg string) { fmt.Println(msg) })
			p.Hide()
			if err != nil {
				dialog.ShowError(err, ctx.Win)
			} else {
				dialog.ShowInformation("Success", "Core updated.", ctx.Win)
			}
		}()
	})

	// Button to update GoTube (App)
	updateAppBtn := widget.NewButton("Check for App Updates", func() {
		p := dialog.NewProgressInfinite("Checking", "Contacting GitHub...", ctx.Win)
		p.Show()
		go func() {
			newVer, downloadUrl, err := updater.CheckAppUpdate()
			p.Hide()
			if err != nil {
				dialog.ShowError(err, ctx.Win)
				return
			}
			if newVer == "" {
				dialog.ShowInformation("Up to Date", "You are using the latest version ("+models.AppVersion+")", ctx.Win)
				return
			}

			dialog.ShowConfirm("Update Available", "Version "+newVer+" is available. Download now?", func(b bool) {
				if b {
					performAppUpdate(ctx, downloadUrl)
				}
			}, ctx.Win)
		}()
	})

	langSelect.OnChanged = func(s string) {
		locales.SetLanguage(s)
		ctx.DB.SaveSetting("Language", s)
		updateFunc()
	}

	return container.NewPadded(widget.NewCard(locales.Get("tab_system"), "", container.NewVBox(
		widget.NewLabel("Language"), langSelect,
		widget.NewSeparator(),
		widget.NewLabel("Core: "+ctx.BinMgr.GetYtDlpPath()),
		updateCoreBtn,
		widget.NewSeparator(),
		widget.NewLabel("App Version: "+models.AppVersion),
		updateAppBtn,
	)))
}

func buildSettingsTab(ctx *AppContext) fyne.CanvasObject {
	return buildSettingsTabWithCallback(ctx, func() {})
}

func showLogs(ctx *AppContext) {
	d := dialog.NewCustom("Logs", "Close", container.NewPadded(ctx.Console), ctx.Win)
	d.Resize(fyne.NewSize(700, 500))
	d.Show()
}
