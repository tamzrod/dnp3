// Package ui provides the user interface for the DNP3 Engineering Workbench.
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Toolbar provides quick access to frequent actions (UX Standard Section 5.1-5.4).
type Toolbar struct {
	container *fyne.Container

	// Buttons (UX Standard: Icons with tooltips)
	connectBtn    *widget.Button
	disconnectBtn *widget.Button
	readClass0Btn *widget.Button
	clearBtn      *widget.Button

	// Callbacks
	OnConnect     func()
	OnDisconnect func()
	OnReadClass0 func()
	OnClear      func()
}

// NewToolbar creates a new toolbar.
func NewToolbar() *Toolbar {
	t := &Toolbar{}

	// Connect button
	t.connectBtn = widget.NewButtonWithIcon("Connect (F5)", theme.ConfirmIcon(), func() {
		if t.OnConnect != nil {
			t.OnConnect()
		}
	})
	t.connectBtn.Importance = widget.HighImportance

	// Disconnect button
	t.disconnectBtn = widget.NewButtonWithIcon("Disconnect (F6)", theme.CancelIcon(), func() {
		if t.OnDisconnect != nil {
			t.OnDisconnect()
		}
	})
	t.disconnectBtn.Importance = widget.MediumImportance
	t.disconnectBtn.Disable()

	// Read Class 0 button
	t.readClass0Btn = widget.NewButtonWithIcon("Read All", theme.SearchIcon(), func() {
		if t.OnReadClass0 != nil {
			t.OnReadClass0()
		}
	})
	t.readClass0Btn.Disable()

	// Clear button
	t.clearBtn = widget.NewButtonWithIcon("Clear Log", theme.ContentClearIcon(), func() {
		if t.OnClear != nil {
			t.OnClear()
		}
	})

	// Layout with separators (UX Standard: Group related actions)
	sep1 := widget.NewSeparator()
	sep2 := widget.NewSeparator()

	t.container = container.NewHBox(
		t.connectBtn,
		t.disconnectBtn,
		sep1,
		t.readClass0Btn,
		sep2,
		t.clearBtn,
	)

	return t
}

// Container returns the toolbar container.
func (t *Toolbar) Container() *fyne.Container {
	return t.container
}

// SetConnected enables/disables buttons based on connection state.
func (t *Toolbar) SetConnected(connected bool) {
	if connected {
		t.connectBtn.Disable()
		t.connectBtn.Importance = widget.MediumImportance
		t.disconnectBtn.Enable()
		t.disconnectBtn.Importance = widget.HighImportance
		t.readClass0Btn.Enable()
	} else {
		t.connectBtn.Enable()
		t.connectBtn.Importance = widget.HighImportance
		t.disconnectBtn.Disable()
		t.disconnectBtn.Importance = widget.MediumImportance
		t.readClass0Btn.Disable()
	}
	t.connectBtn.Refresh()
	t.disconnectBtn.Refresh()
	t.readClass0Btn.Refresh()
}

// SetReading enables/disables buttons during a read operation.
func (t *Toolbar) SetReading(reading bool) {
	if reading {
		t.connectBtn.Disable()
		t.disconnectBtn.Disable()
		t.readClass0Btn.Disable()
	} else {
		// Re-enable based on connection - will be updated by SetConnected
	}
}
