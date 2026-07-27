// Package panels provides UI panel components for the workbench.
package panels

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ModePanel provides mode selection (Master/Outstation).
type ModePanel struct {
	container    *fyne.Container
	masterRB     *widget.RadioGroup
	isMaster     bool
	onModeChange func(bool)
}

// NewModePanel creates a new mode panel.
func NewModePanel() *ModePanel {
	p := &ModePanel{
		isMaster: true, // Master is default for MVP
	}

	title := widget.NewLabel("MODE SELECTION")
	title.TextStyle.Bold = true

	p.masterRB = widget.NewRadioGroup([]string{"Master Mode", "Outstation Mode"}, func(selected string) {
		p.isMaster = (selected == "Master Mode")
		if p.onModeChange != nil {
			p.onModeChange(p.isMaster)
		}
	})
	p.masterRB.Selected = "Master Mode"
	p.masterRB.Horizontal = false

	note := widget.NewLabel("MVP: Master Mode Only")
	note.TextStyle.Italic = true

	p.container = container.NewVBox(
		title,
		widget.NewLabel(""),
		p.masterRB,
		note,
	)

	return p
}

// Container returns the panel container.
func (p *ModePanel) Container() *fyne.Container {
	return p.container
}

// IsMaster returns true if Master mode is selected.
func (p *ModePanel) IsMaster() bool {
	return p.isMaster
}

// SetMaster sets the mode to Master.
func (p *ModePanel) SetMaster(master bool) {
	if master {
		p.masterRB.Selected = "Master Mode"
		p.isMaster = true
	} else {
		p.masterRB.Selected = "Outstation Mode"
		p.isMaster = false
	}
}

// SetOnModeChange sets a callback for mode changes.
func (p *ModePanel) SetOnModeChange(callback func(bool)) {
	p.onModeChange = callback
}
