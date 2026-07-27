// Package panels provides UI panel components for the workbench.
package panels

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// WorkbenchMode represents the workbench operating mode.
type WorkbenchMode int

const (
	// ModePollOutstation - Master mode: connect to and poll an outstation
	ModePollOutstation WorkbenchMode = iota
	// ModeSimulateOutstation - Outstation mode: act as a DNP3 server
	ModeSimulateOutstation
)

// String returns a string representation of the mode.
func (m WorkbenchMode) String() string {
	switch m {
	case ModePollOutstation:
		return "Poll Outstation"
	case ModeSimulateOutstation:
		return "Simulate Outstation"
	default:
		return "Unknown"
	}
}

// ModePanel provides mode selection (Poll Outstation / Simulate Outstation).
type ModePanel struct {
	container    *fyne.Container
	modeRB       *widget.RadioGroup
	mode         WorkbenchMode
	onModeChange func(WorkbenchMode)
}

// NewModePanel creates a new mode panel.
func NewModePanel() *ModePanel {
	p := &ModePanel{
		mode: ModePollOutstation, // Default to master/poll mode
	}

	title := widget.NewLabel("WORKBENCH MODE")
	title.TextStyle.Bold = true

	// Clear, action-oriented labels
	p.modeRB = widget.NewRadioGroup([]string{
		"Poll Outstation",
		"Simulate Outstation",
	}, func(selected string) {
		switch selected {
		case "Poll Outstation":
			p.mode = ModePollOutstation
		case "Simulate Outstation":
			p.mode = ModeSimulateOutstation
		}
		if p.onModeChange != nil {
			p.onModeChange(p.mode)
		}
	})
	p.modeRB.Selected = "Poll Outstation"
	p.modeRB.Horizontal = false

	// Help text
	helpText := widget.NewLabel("Poll: Connect to remote outstation\nSimulate: Act as DNP3 server")
	helpText.TextStyle.Italic = true
	helpText.Wrapping = fyne.TextWrapWord

	p.container = container.NewVBox(
		title,
		widget.NewLabel(""),
		p.modeRB,
		widget.NewLabel(""),
		helpText,
	)

	return p
}

// Container returns the panel container.
func (p *ModePanel) Container() *fyne.Container {
	return p.container
}

// Mode returns the current mode.
func (p *ModePanel) Mode() WorkbenchMode {
	return p.mode
}

// IsPollMode returns true if in poll outstation mode.
func (p *ModePanel) IsPollMode() bool {
	return p.mode == ModePollOutstation
}

// IsSimulateMode returns true if in simulate outstation mode.
func (p *ModePanel) IsSimulateMode() bool {
	return p.mode == ModeSimulateOutstation
}

// SetMode sets the workbench mode.
func (p *ModePanel) SetMode(mode WorkbenchMode) {
	p.mode = mode
	switch mode {
	case ModePollOutstation:
		p.modeRB.Selected = "Poll Outstation"
	case ModeSimulateOutstation:
		p.modeRB.Selected = "Simulate Outstation"
	}
}

// SetOnModeChange sets a callback for mode changes.
func (p *ModePanel) SetOnModeChange(callback func(WorkbenchMode)) {
	p.onModeChange = callback
}
