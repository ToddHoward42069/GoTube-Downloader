package gui

import (
	"gotube/internal/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func buildHistoryTab(ctx *AppContext) fyne.CanvasObject {
	return widget.NewList(
		func() int { return len(ctx.DB.GetHistory()) },
		func() fyne.CanvasObject {
			icon := widget.NewIcon(theme.MediaPlayIcon())
			title := widget.NewLabel("Title")
			title.TextStyle = fyne.TextStyle{Bold: true}
			title.Truncation = fyne.TextTruncateEllipsis
			btn := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), nil)

			// Card look for history items
			content := container.NewBorder(nil, nil, icon, btn, title)
			return widget.NewCard("", "", content)
		},
		func(i int, o fyne.CanvasObject) {
			h := ctx.DB.GetHistory()[i]

			card := o.(*widget.Card)
			border := card.Content.(*fyne.Container)

			// [0]=Label, [1]=Icon(Left), [2]=Button(Right)
			border.Objects[0].(*widget.Label).SetText(h.Title)
			border.Objects[2].(*widget.Button).OnTapped = func() {
				utils.OpenFolder(h.FilePath)
			}
		},
	)
}
