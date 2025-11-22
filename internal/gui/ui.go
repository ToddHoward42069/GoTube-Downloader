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
	w := a.NewWindow("GoTube")
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

	// Shared Footer Components
	viewLogsBtn := widget.NewButton("", func() { showLogs(ctx) })
	statusLabel := widget.NewLabelWithData(ctx.Status)
	statusLabel.Alignment = fyne.TextAlignCenter
	progressContainer := container.NewPadded(widget.NewProgressBarWithData(ctx.Progress))

	// --- Custom Footer for Main Tab ---
	// Layout: [View Logs] [Download]
	footer1 := container.NewVBox(
		widget.NewSeparator(),
		statusLabel,
		progressContainer,
		container.NewGridWithColumns(2, viewLogsBtn, mainBtn),
	)

	// --- Custom Footer for Batch Tab ---
	// Same layout, different action button
	// Note: We can't reuse the exact same widget instance in two places in Fyne safely if switching rapidly
	// But viewLogsBtn is a button, so it can only be in one parent.
	// We need a second Logs button for the second tab.
	viewLogsBtn2 := widget.NewButton("", func() { showLogs(ctx) })

	footer2 := container.NewVBox(
		widget.NewSeparator(),
		statusLabel, // Labels/Bars can be reused usually as they are just renderers of data
		progressContainer,
		container.NewGridWithColumns(2, viewLogsBtn2, batchBtn),
	)

	// Wrap content
	t1Content := container.NewBorder(nil, container.NewPadded(footer1), nil, nil, mainTab)
	t2Content := container.NewBorder(nil, container.NewPadded(footer2), nil, nil, batchTab)

	t1 := container.NewTabItemWithIcon(locales.Get("tab_download"), theme.DownloadIcon(), t1Content)
	t2 := container.NewTabItemWithIcon("Batch", theme.ListIcon(), t2Content)
	t3 := container.NewTabItemWithIcon(locales.Get("tab_history"), theme.HistoryIcon(), historyTab)
	t4 := container.NewTabItemWithIcon(locales.Get("tab_system"), theme.SettingsIcon(), settingsTab)

	tabs := container.NewAppTabs(t1, t2, t3, t4)

	// Global Update
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

	// Log Ticker
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

	t4.Content = buildSettingsTabWithCallback(ctx, updateAllTexts)

	w.SetContent(tabs)
	w.ShowAndRun()
}

// Helper Builders
func buildSettingsTabWithCallback(ctx *AppContext, updateFunc func()) fyne.CanvasObject {
	langSelect := widget.NewSelect([]string{"English", "German"}, nil)
	if ctx.Settings.Language == "German" {
		langSelect.Selected = "German"
	} else {
		langSelect.Selected = "English"
	}

	updateBtn := widget.NewButton("Update Core", func() {
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

	langSelect.OnChanged = func(s string) {
		locales.SetLanguage(s)
		ctx.DB.SaveSetting("Language", s)
		updateFunc()
	}

	return container.NewPadded(widget.NewCard(locales.Get("tab_system"), "", container.NewVBox(
		widget.NewLabel("Language"), langSelect,
		widget.NewSeparator(),
		widget.NewLabel("Core: "+ctx.BinMgr.GetYtDlpPath()),
		updateBtn,
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
