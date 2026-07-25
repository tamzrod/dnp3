package panels

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// StatusBar displays connection status information.
type StatusBar struct {
	container *fyne.Container
	state     binding.String
	connection binding.String
	iin       binding.String
}

// NewStatusBar creates a new status bar.
func NewStatusBar(state, connection, iin binding.String) *StatusBar {
	p := &StatusBar{
		state:      state,
		connection: connection,
		iin:        iin,
	}

	stateLabel := widget.NewLabel("State:")
	stateLabel.TextStyle.Bold = true
	stateValue := widget.NewLabelWithData(state)

	connLabel := widget.NewLabel("Connection:")
	connLabel.TextStyle.Bold = true
	connValue := widget.NewLabelWithData(connection)

	iinLabel := widget.NewLabel("IIN:")
	iinLabel.TextStyle.Bold = true
	iinValue := widget.NewLabelWithData(iin)

	p.container = container.NewHBox(
		stateLabel,
		stateValue,
		layout.NewSpacer(),
		connLabel,
		connValue,
		layout.NewSpacer(),
		iinLabel,
		iinValue,
		layout.NewSpacer(),
		widget.NewLabel("DNP3 Engineering Workbench v0.1.0"),
	)

	return p
}

// Container returns the status bar container.
func (p *StatusBar) Container() *fyne.Container {
	return p.container
}
