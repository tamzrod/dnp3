package panels

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ConnectionState represents the visual state of the connection.
type ConnectionState int

const (
	ConnectionStateDisconnected ConnectionState = iota
	ConnectionStateConnecting
	ConnectionStateConnected
	ConnectionStateError
)

// StatusBar displays connection status information.
type StatusBar struct {
	container *fyne.Container
	state     binding.String
	connection binding.String
	iin       binding.String
	
	connectionIndicator *widget.Label
	errorLabel         *widget.Label
	progressBar        *widget.ProgressBarInfinite
	sidebarToggle      *widget.Button
	logPanelToggle     *widget.Button
	
	OnSidebarToggle func()
	OnLogPanelToggle func()
	
	currentState ConnectionState
}

// NewStatusBar creates a new status bar.
func NewStatusBar(state, connection, iin binding.String) *StatusBar {
	p := &StatusBar{
		state:      state,
		connection: connection,
		iin:        iin,
		currentState: ConnectionStateDisconnected,
	}

	p.connectionIndicator = widget.NewLabel("Disconnected")
	p.connectionIndicator.TextStyle.Bold = true

	p.errorLabel = widget.NewLabel("")
	p.errorLabel.TextStyle.Italic = true
	p.errorLabel.Hide()

	p.progressBar = widget.NewProgressBarInfinite()
	p.progressBar.Hide()

	stateLabel := widget.NewLabel("State:")
	stateLabel.TextStyle.Bold = true
	stateValue := widget.NewLabelWithData(state)

	connLabel := widget.NewLabel("Connection:")
	connLabel.TextStyle.Bold = true
	connValue := widget.NewLabelWithData(connection)

	iinLabel := widget.NewLabel("IIN:")
	iinLabel.TextStyle.Bold = true
	iinValue := widget.NewLabelWithData(iin)

	p.sidebarToggle = widget.NewButtonWithIcon("", theme.ViewRestoreIcon(), func() {
		if p.OnSidebarToggle != nil {
			p.OnSidebarToggle()
		}
	})
	p.sidebarToggle.Importance = widget.MediumImportance
	
	p.logPanelToggle = widget.NewButtonWithIcon("", theme.ViewBottomSheetIcon(), func() {
		if p.OnLogPanelToggle != nil {
			p.OnLogPanelToggle()
		}
	})
	p.logPanelToggle.Importance = widget.MediumImportance

	p.container = container.NewHBox(
		p.connectionIndicator,
		layout.NewSpacer(),
		stateLabel,
		stateValue,
		layout.NewSpacer(),
		connLabel,
		connValue,
		layout.NewSpacer(),
		iinLabel,
		iinValue,
		layout.NewSpacer(),
		p.errorLabel,
		layout.NewSpacer(),
		p.sidebarToggle,
		p.logPanelToggle,
	)

	return p
}

// Container returns the status bar container.
func (p *StatusBar) Container() *fyne.Container {
	return p.container
}

// SetConnectionState updates the visual connection indicator.
func (p *StatusBar) SetConnectionState(state ConnectionState, text string) {
	p.currentState = state
	
	switch state {
	case ConnectionStateConnected:
		p.connectionIndicator.SetText("Connected")
		p.errorLabel.Hide()
		p.progressBar.Hide()
	case ConnectionStateConnecting:
		p.connectionIndicator.SetText("Connecting...")
		p.errorLabel.Hide()
		p.progressBar.Show()
	case ConnectionStateError:
		p.connectionIndicator.SetText("Error")
		p.progressBar.Hide()
	case ConnectionStateDisconnected:
		p.connectionIndicator.SetText("Disconnected")
		p.errorLabel.Hide()
		p.progressBar.Hide()
	}
	
	p.connectionIndicator.Refresh()
}

// ShowError displays an error message in the status bar.
func (p *StatusBar) ShowError(message string) {
	p.errorLabel.SetText(message)
	p.errorLabel.Show()
	p.errorLabel.Refresh()
}

// ClearError hides the error message.
func (p *StatusBar) ClearError() {
	p.errorLabel.Hide()
	p.errorLabel.Refresh()
}

// ShowProgress shows a progress indicator.
func (p *StatusBar) ShowProgress() {
	p.progressBar.Show()
}

// HideProgress hides the progress indicator.
func (p *StatusBar) HideProgress() {
	p.progressBar.Hide()
}

// SetSidebarToggleChecked sets the sidebar toggle visual state.
func (p *StatusBar) SetSidebarToggleChecked(checked bool) {
	if checked {
		p.sidebarToggle.Importance = widget.MediumImportance
	} else {
		p.sidebarToggle.Importance = widget.LowImportance
	}
	p.sidebarToggle.Refresh()
}

// SetLogPanelToggleChecked sets the log panel toggle visual state.
func (p *StatusBar) SetLogPanelToggleChecked(checked bool) {
	if checked {
		p.logPanelToggle.Importance = widget.MediumImportance
	} else {
		p.logPanelToggle.Importance = widget.LowImportance
	}
	p.logPanelToggle.Refresh()
}
