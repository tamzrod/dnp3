package dialogs

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// AppInfo contains application metadata.
var AppInfo = struct {
	Name        string
	Version     string
	Description string
	Copyright   string
	Framework   string
}{
	Name:        "DNP3 Engineering Workbench",
	Version:     "0.1.0",
	Description: "A desktop application for validating and debugging the native Go DNP3 library.",
	Copyright:   "2026",
	Framework:   "Fyne",
}

// ShowAbout displays the About dialog.
func ShowAbout(parent fyne.Window) {
	nameLabel := widget.NewLabel(AppInfo.Name)
	nameLabel.Alignment = fyne.TextAlignCenter

	versionLabel := widget.NewLabel("Version " + AppInfo.Version)
	versionLabel.Alignment = fyne.TextAlignCenter

	descLabel := widget.NewLabel(AppInfo.Description)
	descLabel.Alignment = fyne.TextAlignCenter

	frameworkLabel := widget.NewLabel("Built with " + AppInfo.Framework)
	frameworkLabel.Alignment = fyne.TextAlignCenter

	copyrightLabel := widget.NewLabel("Copyright " + AppInfo.Copyright)
	copyrightLabel.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		nameLabel,
		versionLabel,
		widget.NewLabel(""),
		descLabel,
		widget.NewLabel(""),
		frameworkLabel,
		widget.NewLabel(""),
		copyrightLabel,
	)

	dialog.ShowCustom("About", "Close", content, parent)
}

// ShortcutsDialog displays keyboard shortcuts help.
type ShortcutsDialog struct {
	parent fyne.Window
}

// NewShortcutsDialog creates a new shortcuts dialog.
func NewShortcutsDialog(parent fyne.Window) *ShortcutsDialog {
	return &ShortcutsDialog{parent: parent}
}

// Show displays the shortcuts dialog.
func (s *ShortcutsDialog) Show() {
	shortcuts := []struct {
		shortcut string
		action   string
	}{
		{"F5", "Connect to outstation"},
		{"F6", "Disconnect from outstation"},
		{"F11", "Toggle fullscreen"},
		{"Ctrl+F", "Find in log"},
		{"Escape", "Exit fullscreen / Close dialog"},
		{"Ctrl+N", "New session"},
		{"Ctrl+O", "Open configuration"},
		{"Ctrl+S", "Save configuration"},
	}

	content := container.NewVBox()
	content.Add(widget.NewLabel("Keyboard Shortcuts"))
	content.Add(widget.NewLabel(""))

	for _, sc := range shortcuts {
		row := container.NewHBox(
			widget.NewLabel(sc.shortcut),
			widget.NewLabel("  " + sc.action),
		)
		content.Add(row)
	}

	dialog.ShowCustom("Keyboard Shortcuts", "Close", content, s.parent)
}
