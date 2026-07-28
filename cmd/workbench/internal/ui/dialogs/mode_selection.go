package dialogs

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"dnp3/cmd/workbench/internal/shared/types"
)

// ShowModeSelection displays a mode selection dialog and calls the callback with the selected mode.
func ShowModeSelection(a fyne.App, onSelect func(types.Mode)) {
	// Create a temporary window for the dialog
	dialogWindow := a.NewWindow("DNP3 Engineering Workbench - Select Mode")
	dialogWindow.Resize(fyne.NewSize(400, 300))
	dialogWindow.SetFixedSize(true)
	dialogWindow.CenterOnScreen()

	// Title
	title := widget.NewLabel("DNP3 Engineering Workbench")
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = widget.RichTextStyleHeading1

	subtitle := widget.NewLabel("Choose Operating Mode:")
	subtitle.Alignment = fyne.TextAlignCenter

	// Master mode card
	masterIcon := widget.NewLabel("🔌")
	masterIcon.Alignment = fyne.TextAlignCenter

	masterTitle := widget.NewLabel("Master Mode")
	masterTitle.Alignment = fyne.TextAlignCenter
	masterTitle.TextStyle.Bold = true

	masterDesc := widget.NewLabel("Connect to remote outstations\nRead/write data points")
	masterDesc.Alignment = fyne.TextAlignCenter

	masterCard := widget.NewCard("", "", container.NewVBox(masterIcon, masterTitle, masterDesc))
	masterCard.OnTapped = func() {
		dialogWindow.Close()
		onSelect(types.ModeMaster)
	}

	// Outstation mode card
	outstationIcon := widget.NewLabel("🖥️")
	outstationIcon.Alignment = fyne.TextAlignCenter

	outstationTitle := widget.NewLabel("Outstation Mode")
	outstationTitle.Alignment = fyne.TextAlignCenter
	outstationTitle.TextStyle.Bold = true

	outstationDesc := widget.NewLabel("Run simulated DNP3 server\nGenerate random data")
	outstationDesc.Alignment = fyne.TextAlignCenter

	outstationCard := widget.NewCard("", "", container.NewVBox(outstationIcon, outstationTitle, outstationDesc))
	outstationCard.OnTapped = func() {
		dialogWindow.Close()
		onSelect(types.ModeOutstation)
	}

	// Cards container
	cardsContainer := container.NewGridWithColumns(2,
		masterCard,
		outstationCard,
	)

	// Cancel button
	cancelBtn := widget.NewButton("Cancel", func() {
		dialogWindow.Close()
	})
	cancelBtn.Importance = widget.LowImportance

	// Layout
	content := container.NewVBox(
		layout.NewSpacer(),
		title,
		widget.NewLabel(""),
		subtitle,
		widget.NewLabel(""),
		cardsContainer,
		widget.NewLabel(""),
		cancelBtn,
		layout.NewSpacer(),
	)

	dialogWindow.SetContent(content)
	dialogWindow.Show()
}
