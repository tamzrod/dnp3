package panels

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// CommandPanel provides DNP3 command buttons with state management.
type CommandPanel struct {
	container *fyne.Container

	// Button references for state management (UX Standard Section 6.5)
	btnReadClass0  *widget.Button
	btnReadClass1  *widget.Button
	btnReadClass2  *widget.Button
	btnReadClass3  *widget.Button
	btnOperateOn   *widget.Button
	btnOperateOff  *widget.Button

	OnReadClass func(class int)
	OnOperate  func(index uint16, value bool)
}

// NewCommandPanel creates a new command panel.
func NewCommandPanel() *CommandPanel {
	p := &CommandPanel{}

	title := widget.NewLabel("COMMAND PANEL")
	title.TextStyle.Bold = true

	// Read buttons
	readLabel := widget.NewLabel("Read Commands:")
	readLabel.TextStyle.Italic = true

	p.btnReadClass0 = widget.NewButtonWithIcon("Read Class 0", theme.SearchIcon(), func() {
		if p.OnReadClass != nil {
			p.OnReadClass(0)
		}
	})
	p.btnReadClass0.Disable() // Disabled by default until connected

	p.btnReadClass1 = widget.NewButtonWithIcon("Read Class 1", theme.SearchIcon(), func() {
		if p.OnReadClass != nil {
			p.OnReadClass(1)
		}
	})
	p.btnReadClass1.Disable()

	p.btnReadClass2 = widget.NewButtonWithIcon("Read Class 2", theme.SearchIcon(), func() {
		if p.OnReadClass != nil {
			p.OnReadClass(2)
		}
	})
	p.btnReadClass2.Disable()

	p.btnReadClass3 = widget.NewButtonWithIcon("Read Class 3", theme.SearchIcon(), func() {
		if p.OnReadClass != nil {
			p.OnReadClass(3)
		}
	})
	p.btnReadClass3.Disable()

	// Control buttons
	controlLabel := widget.NewLabel("Control Commands:")
	controlLabel.TextStyle.Italic = true

	p.btnOperateOn = widget.NewButtonWithIcon("Operate ON", theme.CheckButtonIcon(), func() {
		if p.OnOperate != nil {
			p.OnOperate(0, true)
		}
	})
	p.btnOperateOn.Disable()

	p.btnOperateOff = widget.NewButtonWithIcon("Operate OFF", theme.CancelIcon(), func() {
		if p.OnOperate != nil {
			p.OnOperate(0, false)
		}
	})
	p.btnOperateOff.Disable()

	p.container = container.NewVBox(
		title,
		widget.NewLabel(""),
		readLabel,
		p.btnReadClass0,
		p.btnReadClass1,
		p.btnReadClass2,
		p.btnReadClass3,
		widget.NewLabel(""),
		controlLabel,
		p.btnOperateOn,
		p.btnOperateOff,
	)

	return p
}

// Container returns the panel container.
func (p *CommandPanel) Container() *fyne.Container {
	return p.container
}

// SetConnected enables or disables command buttons based on connection state.
// (UX Standard Section 6.5: Disable controls when action unavailable)
func (p *CommandPanel) SetConnected(connected bool) {
	if connected {
		p.btnReadClass0.Enable()
		p.btnReadClass1.Enable()
		p.btnReadClass2.Enable()
		p.btnReadClass3.Enable()
		p.btnOperateOn.Enable()
		p.btnOperateOff.Enable()
	} else {
		p.btnReadClass0.Disable()
		p.btnReadClass1.Disable()
		p.btnReadClass2.Disable()
		p.btnReadClass3.Disable()
		p.btnOperateOn.Disable()
		p.btnOperateOff.Disable()
	}
}

// SetOperationInProgress shows a loading state during command execution.
func (p *CommandPanel) SetOperationInProgress(inProgress bool) {
	if inProgress {
		// Could show a spinner or disable all buttons during operation
		p.btnReadClass0.Disable()
		p.btnReadClass1.Disable()
		p.btnReadClass2.Disable()
		p.btnReadClass3.Disable()
	} else {
		// Re-enable based on connection state
		// This is a simplified version; a full implementation would track connection state
	}
}
