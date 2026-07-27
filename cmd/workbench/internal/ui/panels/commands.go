package panels

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CommandPanel provides DNP3 command buttons.
type CommandPanel struct {
	container *fyne.Container

	OnReadClass  func(class int)
	OnOperate   func(index uint16, value bool)
}

// NewCommandPanel creates a new command panel.
func NewCommandPanel() *CommandPanel {
	p := &CommandPanel{}

	title := widget.NewLabel("COMMAND PANEL")
	title.TextStyle.Bold = true

	// Read buttons
	readLabel := widget.NewLabel("Read Commands:")
	readLabel.TextStyle.Italic = true

	btnReadClass0 := widget.NewButton("Read Class 0", func() {
		if p.OnReadClass != nil {
			p.OnReadClass(0)
		}
	})

	btnReadClass1 := widget.NewButton("Read Class 1", func() {
		if p.OnReadClass != nil {
			p.OnReadClass(1)
		}
	})

	btnReadClass2 := widget.NewButton("Read Class 2", func() {
		if p.OnReadClass != nil {
			p.OnReadClass(2)
		}
	})

	btnReadClass3 := widget.NewButton("Read Class 3", func() {
		if p.OnReadClass != nil {
			p.OnReadClass(3)
		}
	})

	// Control buttons
	controlLabel := widget.NewLabel("Control Commands:")
	controlLabel.TextStyle.Italic = true

	btnOperateOn := widget.NewButton("Operate ON (index 0)", func() {
		if p.OnOperate != nil {
			p.OnOperate(0, true)
		}
	})

	btnOperateOff := widget.NewButton("Operate OFF (index 0)", func() {
		if p.OnOperate != nil {
			p.OnOperate(0, false)
		}
	})

	p.container = container.NewVBox(
		title,
		widget.NewLabel(""),
		readLabel,
		btnReadClass0,
		btnReadClass1,
		btnReadClass2,
		btnReadClass3,
		widget.NewLabel(""),
		controlLabel,
		btnOperateOn,
		btnOperateOff,
	)

	return p
}

// Container returns the panel container.
func (p *CommandPanel) Container() *fyne.Container {
	return p.container
}

// SetEnabled enables or disables all command buttons.
func (p *CommandPanel) SetEnabled(enabled bool) {
	// This would be implemented with proper widget state management
	_ = enabled
}
