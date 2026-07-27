// Package dialogs provides dialog windows for the DNP3 Workbench.
package dialogs

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"dnp3/cmd/workbench/internal/config"
)

// SettingsDialog provides a dialog for user preferences.
type SettingsDialog struct {
	dialog    *dialog.CustomDialog
	container *fyne.Container
	
	cfg       *config.Config
	
	// Settings widgets
	autoScroll  *widget.Check
	confirmDisc *widget.Check
	themeSelect *widget.Select
}

// NewSettingsDialog creates a new settings dialog.
func NewSettingsDialog(parent fyne.Window, cfg *config.Config) *SettingsDialog {
	s := &SettingsDialog{
		cfg: cfg,
	}

	// Appearance section
	appearanceLabel := widget.NewLabel("Appearance")
	appearanceLabel.TextStyle.Bold = true

	themeLabel := widget.NewLabel("Theme:")
	s.themeSelect = widget.NewSelect([]string{"Light", "Dark"}, func(selected string) {
		s.cfg.Appearance.Theme = selected
		// Apply theme immediately
		a := app.Current()
		if selected == "Dark" {
			a.Settings().SetTheme(theme.DarkTheme())
		} else {
			a.Settings().SetTheme(theme.LightTheme())
		}
	})
	if cfg.Appearance.Theme == "" {
		s.themeSelect.Selected = "Light"
	} else {
		s.themeSelect.Selected = cfg.Appearance.Theme
	}

	// Behavior section
	behaviorLabel := widget.NewLabel("")
	behaviorLabel.TextStyle.Bold = true

	s.autoScroll = widget.NewCheck("Auto-scroll log on new entries", func(checked bool) {
		s.cfg.Behavior.AutoScroll = checked
	})
	s.autoScroll.SetChecked(s.cfg.Behavior.AutoScroll)

	s.confirmDisc = widget.NewCheck("Confirm before disconnecting", func(checked bool) {
		s.cfg.Behavior.ConfirmDisconnect = checked
	})
	s.confirmDisc.SetChecked(s.cfg.Behavior.ConfirmDisconnect)

	// Buttons
	saveBtn := widget.NewButton("Save", func() {
		s.cfg.Save()
		s.dialog.Hide()
	})
	cancelBtn := widget.NewButton("Cancel", func() {
		s.dialog.Hide()
	})

	buttonBox := container.NewHBox(
		widget.NewLabel(""),
		widget.NewLabel(""),
		widget.NewLabel(""),
		widget.NewLabel(""),
		saveBtn,
		cancelBtn,
	)

	s.container = container.NewVBox(
		widget.NewLabel("Settings"),
		widget.NewLabel(""),
		appearanceLabel,
		container.NewHBox(themeLabel, s.themeSelect),
		widget.NewLabel(""),
		widget.NewLabel("Behavior"),
		s.autoScroll,
		s.confirmDisc,
		widget.NewLabel(""),
		buttonBox,
	)

	s.dialog = dialog.NewCustom("Settings", "Close", s.container, parent)
	return s
}

// Show displays the settings dialog.
func (s *SettingsDialog) Show() {
	s.dialog.Show()
}

// Hide hides the settings dialog.
func (s *SettingsDialog) Hide() {
	s.dialog.Hide()
}
