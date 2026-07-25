// Package panels provides UI panel components for the workbench.
package panels

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ModePanel provides mode selection (Master/Outstation).
type ModePanel struct {
	container *fyne.Container
	masterRB  *widget.RadioButton
	outstnRB  *widget.RadioButton
}

// NewModePanel creates a new mode panel.
func NewModePanel() *ModePanel {
	p := &ModePanel{}

	title := widget.NewLabel("MODE SELECTION")
	title.TextStyle.Bold = true

	p.masterRB = widget.NewRadioButton("Master Mode", func(string) {})
	p.masterRB.Selected = true // Master is default for MVP
	p.masterRB.Disable()      // Only Master for MVP

	p.outstnRB = widget.NewRadioButton("Outstation Mode", func(string) {})
	p.outstnRB.Disable() // Only Master for MVP

	note := widget.NewLabel("MVP: Master Mode Only")
	note.TextStyle.Italic = true

	p.container = container.NewVBox(
		title,
		layout.NewSpacer().MinSize().Height,
		p.masterRB,
		p.outstnRB,
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
	return p.masterRB.Selected
}

// SetMaster sets the mode to Master.
func (p *ModePanel) SetMaster(master bool) {
	if master {
		p.masterRB.SetSelected("Master Mode")
	} else {
		p.outstnRB.SetSelected("Outstation Mode")
	}
}
