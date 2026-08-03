// Package ui provides the user interface for the DNP3 Engineering Workbench.
package ui

import (
	"fmt"
	"io"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"dnp3/cmd/workbench/internal/config"
	masterctrl "dnp3/cmd/workbench/internal/master"
	"dnp3/cmd/workbench/internal/session"
	"dnp3/cmd/workbench/internal/ui/panels"
)

// MasterWindow represents the Master mode window.
type MasterWindow struct {
	app    fyne.App
	window fyne.Window
	ctrl   *masterctrl.Controller

	// Panels
	connectionPanel *panels.ConnectionPanel
	commandPanel   *panels.CommandPanel
	dataTablePanel *panels.DataTablePanel
	controlPanel   *panels.ControlPanel
	logPanel       *panels.LogPanel
	statusBar      *panels.StatusBar

	// State bindings
	state             binding.String
	connectionBinding binding.String
	iinBinding        binding.String

	mu     sync.RWMutex
	closed bool
}

// NewMasterWindow creates a new Master window.
func NewMasterWindow(app fyne.App, ctrl *masterctrl.Controller, cfg *config.Config) *MasterWindow {
	w := &MasterWindow{
		app:               app,
		window:            app.NewWindow("DNP3 Master"),
		ctrl:              ctrl,
		state:             binding.NewString(),
		connectionBinding: binding.NewString(),
		iinBinding:        binding.NewString(),
	}

	w.state.Set("Disconnected")
	w.connectionBinding.Set("Not Connected")
	w.iinBinding.Set("0x0000")

	w.setupUI()
	w.setupEventHandlers()

	return w
}

// setupUI creates the UI components.
func (w *MasterWindow) setupUI() {
	// Create panels
	w.connectionPanel = panels.NewConnectionPanel(nil) // Master-specific
	w.commandPanel = panels.NewCommandPanel()
	w.dataTablePanel = panels.NewDataTablePanel()
	w.controlPanel = panels.NewControlPanel()
	w.logPanel = panels.NewLogPanel()
	w.statusBar = panels.NewStatusBar(w.state, w.connectionBinding, w.iinBinding)

	// Left sidebar - connection, commands
	leftSidebar := container.NewVBox(
		w.connectionPanel.Container(),
		w.commandPanel.Container(),
	)

	// Right side - data table and control panel
	rightContent := container.NewVBox(
		w.dataTablePanel.Container(),
		w.controlPanel.Container(),
	)

	// Make the split pane resizable
	mainContent := container.NewHSplit(
		leftSidebar,
		rightContent,
	)
	mainContent.Offset = 0.25 // 25% for sidebar

	// Complete layout - status bar at bottom
	content := container.NewBorder(
		nil,                      // Top
		w.statusBar.Container(),  // bottom
		nil,                      // left
		nil,                      // right
		container.NewVBox(
			mainContent,
			w.logPanel.Container(),
		),
	)

	w.window.SetContent(content)
}

// setupEventHandlers sets up event handling between panels.
func (w *MasterWindow) setupEventHandlers() {
	// Connection panel events
	w.connectionPanel.OnConnect = func(address string, port int) {
		w.ctrl.Connect(address, port)
	}

	w.connectionPanel.OnDisconnect = func() {
		w.ctrl.Disconnect()
	}

	// Command panel events
	w.commandPanel.OnReadClass = func(class int) {
		w.ctrl.ReadClass(class)
	}

	w.commandPanel.OnOperate = func(index uint16, value bool) {
		w.ctrl.Operate(index, value)
	}

	// Log panel events
	w.logPanel.OnClear = func() {
		w.ctrl.Logger().Clear()
		w.logPanel.Clear()
	}

	// Status bar toggle callbacks
	w.statusBar.OnSidebarToggle = func() {
		w.ToggleSidebar()
	}

	w.statusBar.OnLogPanelToggle = func() {
		w.ToggleLogPanel()
	}

	// Data table panel events
	w.dataTablePanel.OnPointSelected = func(pointType panels.PointType, index uint16, selected bool) {
		if selected {
			w.controlPanel.SelectPoint(pointType, index)
		} else {
			w.controlPanel.DeselectPoint(pointType, index)
		}
	}

	w.dataTablePanel.SetOnReadAll(func() {
		w.ctrl.ReadClass(0)
	})

	// Control panel events
	w.controlPanel.OnOperate = func(pointType panels.PointType, index uint16, value interface{}) {
		switch v := value.(type) {
		case bool:
			w.ctrl.Operate(index, v)
		case string:
			// Handle analog output
		}
	}

	// Start state polling
	go w.pollState()
}

// pollState periodically checks controller state and updates UI.
func (w *MasterWindow) pollState() {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	var lastState *masterctrl.State
	for !w.closed {
		<-ticker.C

		state := w.ctrl.State()
		if state != lastState {
			lastState = state
			w.updateUI(state)
		}
	}
}

// updateUI updates UI elements from the app state.
func (w *MasterWindow) updateUI(state *masterctrl.State) {
	// Update state display
	w.state.Set(state.Connection.String())

	// Update connection info
	if state.Address != "" {
		w.connectionBinding.Set(fmt.Sprintf("%s:%d", state.Address, state.Port))
	}

	// Update IIN and data table panel
	if state.LastResponse != nil {
		w.iinBinding.Set(fmt.Sprintf("0x%02X%02X", state.LastResponse.IIN[0], state.LastResponse.IIN[1]))
		w.dataTablePanel.Update(state.LastResponse, state.LastResponse.Timestamp)
	}

	// Update connection panel
	connected := state.Connection == session.StateConnected
	w.connectionPanel.SetConnected(connected)
	w.commandPanel.SetConnected(connected)

	// Update status bar
	w.updateStatusBar(state)
}

// updateStatusBar updates the status bar with connection state.
func (w *MasterWindow) updateStatusBar(state *masterctrl.State) {
	switch state.Connection {
	case session.StateConnected:
		w.statusBar.SetConnectionState(panels.ConnectionStateConnected, "Connected")
		w.statusBar.ClearError()
	case session.StateConnecting:
		w.statusBar.SetConnectionState(panels.ConnectionStateConnecting, "Connecting...")
	case session.StateError:
		w.statusBar.SetConnectionState(panels.ConnectionStateError, "Error")
		if state.Error != "" {
			w.statusBar.ShowError(state.Error)
		}
	default:
		w.statusBar.SetConnectionState(panels.ConnectionStateDisconnected, "Disconnected")
	}
}

// Window returns the underlying Fyne window.
func (w *MasterWindow) Window() fyne.Window {
	return w.window
}

// Show shows the window.
func (w *MasterWindow) Show() {
	w.window.Show()
}

// Resize resizes the window.
func (w *MasterWindow) Resize(size fyne.Size) {
	w.window.Resize(size)
}

// SetTitle sets the window title.
func (w *MasterWindow) SetTitle(title string) {
	w.window.SetTitle(title)
}

// CenterOnScreen centers the window on screen.
func (w *MasterWindow) CenterOnScreen() {
	w.window.CenterOnScreen()
}

// SetMainMenu sets the main menu for the window.
func (w *MasterWindow) SetMainMenu(menu *fyne.MainMenu) {
	w.window.SetMainMenu(menu)
}

// Maximize maximizes the window.
func (w *MasterWindow) Maximize() {
	w.window.SetFullScreen(true)
}

// Restore restores the window.
func (w *MasterWindow) Restore() {
	w.window.SetFullScreen(false)
}

// ToggleSidebar shows/hides the sidebar panel.
func (w *MasterWindow) ToggleSidebar() {
	// Placeholder for future implementation
}

// ToggleLogPanel shows/hides the log panel.
func (w *MasterWindow) ToggleLogPanel() {
	// Placeholder for future implementation
}

// ShowLogSearch shows the log search bar.
func (w *MasterWindow) ShowLogSearch() {
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search log...")

	searchEntry.OnSubmitted = func(text string) {
		w.logPanel.Search(text)
	}

	dialog.ShowCustom("Find in Log", "Close", container.NewHBox(
		searchEntry,
		widget.NewButton("Find", func() {
			w.logPanel.Search(searchEntry.Text)
		}),
	), w.window)

	searchEntry.FocusGained()
}

// ExportLog exports the log to the provided writer.
func (w *MasterWindow) ExportLog(writer fyne.URIWriteCloser) {
	if writer == nil {
		return
	}
	defer writer.Close()

	entries := w.logPanel.GetEntries()
	for _, entry := range entries {
		line := fmt.Sprintf("[%s] %s %s\n",
			entry.Timestamp.Format("2006-01-02 15:04:05.000"),
			entry.Direction,
			entry.Message)
		io.WriteString(writer, line)
	}
}

// ClearLog clears the log panel.
func (w *MasterWindow) ClearLog() {
	w.logPanel.Clear()
}
