// Package dialogs provides dialog windows for the DNP3 Workbench.
package dialogs

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
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
	// Create icon label with styling
	iconLabel := widget.NewLabelWithStyle("⬡", fyne.TextAlignCenter, widget.TextStyle{
		Size: 48,
	})

	// Application name with bold styling
	nameLabel := widget.NewLabelWithStyle(AppInfo.Name, fyne.TextAlignCenter, widget.TextStyle{
		Bold: true,
		Size: 20,
	})

	// Version
	versionLabel := widget.NewLabelWithStyle("Version "+AppInfo.Version, fyne.TextAlignCenter, widget.TextStyle{
		Italic: true,
	})

	// Description
	descLabel := widget.NewLabel(AppInfo.Description)
	descLabel.Alignment = fyne.TextAlignCenter

	// Framework info
	frameworkLabel := widget.NewLabel("Built with " + AppInfo.Framework)
	frameworkLabel.Alignment = fyne.TextAlignCenter
	frameworkLabel.TextStyle.Italic = true
	frameworkLabel.TextStyle.Color = theme.DisabledColor()

	// Copyright
	copyrightLabel := widget.NewLabel("© " + AppInfo.Copyright + " All rights reserved.")
	copyrightLabel.Alignment = fyne.TextAlignCenter
	copyrightLabel.TextStyle.Italic = true
	copyrightLabel.TextStyle.Color = theme.DisabledColor()

	// Layout with padding
	content := container.NewVBox(
		iconLabel,
		widget.NewLabel(""),
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
	content.Add(widget.NewLabelWithStyle("Keyboard Shortcuts", fyne.TextAlignCenter, widget.TextStyle{Bold: true}))
	content.Add(widget.NewLabel(""))

	for _, s := range shortcuts {
		row := container.NewHBox(
			widget.NewLabel(s.shortcut),
		)
		row.Add(widget.NewLabel("                    " + s.action))
		content.Add(row)
	}

	dialog.ShowCustom("Keyboard Shortcuts", "Close", content, s.parent)
}
