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
	"dnp3/cmd/workbench/internal/controller"
	"dnp3/cmd/workbench/internal/session"
	"dnp3/cmd/workbench/internal/ui/panels"
)

// MainWindow represents the main application window.
type MainWindow struct {
	app    fyne.App
	window fyne.Window
	ctrl   *controller.Controller

	// Panels
	modePanel       *panels.ModePanel
	connectionPanel *panels.ConnectionPanel
	commandPanel    *panels.CommandPanel
	dataPanel       *panels.DataPanel
	protocolPanel   *panels.ProtocolPanel
	logPanel        *panels.LogPanel
	statusBar       *panels.StatusBar

	// Toolbar (UX Standard Section 5.1-5.4)
	toolbar *Toolbar

	// Layout state (UX Standard Section 6.3)
	sidebarVisible  bool
	logPanelVisible bool
	fullscreen      bool

	// Search state
	searchEntry *widget.Entry
	searchOpen  bool

	// State bindings
	state        binding.String
	connectionBinding binding.String
	iinBinding       binding.String

	mu     sync.RWMutex
	closed bool
}

// NewMainWindow creates a new main window.
func NewMainWindow(app fyne.App, ctrl *controller.Controller, cfg *config.Config) *MainWindow {
	w := &MainWindow{
		app:    app,
		window:  app.NewWindow("DNP3 Engineering Workbench"),
		ctrl:    ctrl,
		state:   binding.NewString(),
		connectionBinding: binding.NewString(),
		iinBinding: binding.NewString(),
		
		// Initialize visibility states from config (UX Standard: collapsible panels)
		sidebarVisible:  cfg.Layout.SidebarVisible,
		logPanelVisible: cfg.Layout.LogPanelVisible,
		fullscreen:      cfg.Window.Full,
		searchOpen:      false,
	}

	w.state.Set("Disconnected")
	w.connectionBinding.Set("Not Connected")
	w.iinBinding.Set("0x0000")

	w.setupUI()
	w.setupEventHandlers()

	return w
}

// setupUI creates the UI components.
func (w *MainWindow) setupUI() {
	// Create panels
	w.modePanel = panels.NewModePanel()
	w.connectionPanel = panels.NewConnectionPanel(w.ctrl)
	w.commandPanel = panels.NewCommandPanel()
	w.dataPanel = panels.NewDataPanel()
	w.protocolPanel = panels.NewProtocolPanel()
	w.logPanel = panels.NewLogPanel()
	w.statusBar = panels.NewStatusBar(w.state, w.connectionBinding, w.iinBinding)

	// Create toolbar (UX Standard Section 5.1-5.4)
	w.toolbar = NewToolbar()

	// Left sidebar
	leftSidebar := container.NewVBox(
		w.modePanel.Container(),
		w.connectionPanel.Container(),
		w.commandPanel.Container(),
	)

	// Main content area - split between data/protocol and sidebar
	mainContent := container.NewHSplit(
		leftSidebar,
		container.NewVBox(
			w.dataPanel.Container(),
			w.protocolPanel.Container(),
		),
	)

	// Complete layout with toolbar at top (UX Standard Section 5.1)
	content := container.NewBorder(
		w.toolbar.Container(), // top - toolbar
		w.statusBar.Container(), // bottom - status bar
		nil, // left
		nil, // right
		container.NewVBox(
			mainContent,
			w.logPanel.Container(),
		),
	)

	w.window.SetContent(content)
}

// setupEventHandlers sets up event handling between panels.
func (w *MainWindow) setupEventHandlers() {
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

	// Toolbar events (UX Standard Section 5.4)
	w.toolbar.OnConnect = func() {
		w.ctrl.Connect(w.ctrl.State().Address, w.ctrl.State().Port)
	}

	w.toolbar.OnDisconnect = func() {
		w.ctrl.Disconnect()
	}

	w.toolbar.OnReadClass0 = func() {
		w.ctrl.ReadClass(0)
	}

	w.toolbar.OnClear = func() {
		w.ctrl.Logger().Clear()
		w.logPanel.Clear()
	}

	// Status bar toggle callbacks (UX Standard Section 6.3)
	w.statusBar.OnSidebarToggle = func() {
		w.ToggleSidebar()
	}

	w.statusBar.OnLogPanelToggle = func() {
		w.ToggleLogPanel()
	}
}

// Show shows the window and starts the state polling goroutine.
func (w *MainWindow) Show() {
	go w.pollState()
	w.window.Show()
}

// Resize resizes the window.
func (w *MainWindow) Resize(size fyne.Size) {
	w.window.Resize(size)
}

// SetTitle sets the window title.
func (w *MainWindow) SetTitle(title string) {
	w.window.SetTitle(title)
}

// CenterOnScreen centers the window on screen.
func (w *MainWindow) CenterOnScreen() {
	w.window.CenterOnScreen()
}

// SetMainMenu sets the main menu for the window.
func (w *MainWindow) SetMainMenu(menu *fyne.MainMenu) {
	w.window.SetMainMenu(menu)
}

// pollState periodically checks controller state and updates UI.
func (w *MainWindow) pollState() {
	ticker := time.NewTicker(100 * time.Millisecond) // Update every 100ms
	defer ticker.Stop()

	var lastState *controller.AppState
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
func (w *MainWindow) updateUI(state *controller.AppState) {
	// Update state display
	w.state.Set(state.Connection.String())

	// Update connection info
	if state.Address != "" {
		w.connectionBinding.Set(fmt.Sprintf("%s:%d", state.Address, state.Port))
	}

	// Update IIN and data panel
	if state.LastResponse != nil {
		w.iinBinding.Set(fmt.Sprintf("0x%02X%02X", state.LastResponse.IIN[0], state.LastResponse.IIN[1]))
		w.dataPanel.Update(state.LastResponse)
	}

	// Update all panels based on connection state
	connected := state.Connection == session.StateConnected
	
	// Update connection panel
	w.connectionPanel.SetConnected(connected)
	
	// Update command panel (UX Standard: disable commands when disconnected)
	w.commandPanel.SetConnected(connected)
	
	// Update toolbar (UX Standard Section 5.4: disable when unavailable)
	w.toolbar.SetConnected(connected)
	
	// Update status bar with visual indicator
	w.updateStatusBar(state)
}

// updateStatusBar updates the status bar with connection state.
func (w *MainWindow) updateStatusBar(state *controller.AppState) {
	switch state.Connection {
	case session.StateConnected:
		w.statusBar.SetConnectionState(panels.ConnectionStateConnected, "Connected")
		w.statusBar.ClearError()
	case session.StateConnecting:
		w.statusBar.SetConnectionState(panels.ConnectionStateConnecting, "Connecting...")
	case session.StateError:
		w.statusBar.SetConnectionState(panels.ConnectionStateError, "Error")
		if state.ConnectionError != "" {
			w.statusBar.ShowError(state.ConnectionError)
		}
	default:
		w.statusBar.SetConnectionState(panels.ConnectionStateDisconnected, "Disconnected")
	}
}

// Window returns the underlying Fyne window.
func (w *MainWindow) Window() fyne.Window {
	return w.window
}

// ToggleSidebar shows/hides the sidebar panel (UX Standard Section 6.3).
func (w *MainWindow) ToggleSidebar() {
	w.mu.Lock()
	w.sidebarVisible = !w.sidebarVisible
	visible := w.sidebarVisible
	w.mu.Unlock()
	
	if visible {
		w.statusBar.SetSidebarToggleChecked(true)
	} else {
		w.statusBar.SetSidebarToggleChecked(false)
	}
	// Note: Full layout rebuild would be needed for proper toggle
	// This is a placeholder for future implementation
}

// ToggleLogPanel shows/hides the log panel (UX Standard Section 6.3).
func (w *MainWindow) ToggleLogPanel() {
	w.mu.Lock()
	w.logPanelVisible = !w.logPanelVisible
	visible := w.logPanelVisible
	w.mu.Unlock()
	
	if visible {
		w.statusBar.SetLogPanelToggleChecked(true)
	} else {
		w.statusBar.SetLogPanelToggleChecked(false)
	}
	// Note: Full layout rebuild would be needed for proper toggle
	// This is a placeholder for future implementation
}

// ToggleFullscreen toggles fullscreen mode (UX Standard Section 4.4).
func (w *MainWindow) ToggleFullscreen() {
	w.mu.Lock()
	w.fullscreen = !w.fullscreen
	fullscreen := w.fullscreen
	w.mu.Unlock()
	
	if fullscreen {
		w.window.SetFullScreen(true)
	} else {
		w.window.SetFullScreen(false)
	}
}

// ShowLogSearch shows the log search bar (UX Standard Section 7.3).
func (w *MainWindow) ShowLogSearch() {
	if w.searchOpen {
		return
	}
	
	w.searchOpen = true
	w.searchEntry = widget.NewEntry()
	w.searchEntry.SetPlaceHolder("Search log...")
	
	w.searchEntry.OnSubmitted = func(text string) {
		w.logPanel.Search(text)
	}
	
	// Create a simple dialog for search
	dialog.ShowCustom("Find in Log", "Close", container.NewHBox(
		w.searchEntry,
		widget.NewButton("Find", func() {
			w.logPanel.Search(w.searchEntry.Text)
		}),
	), w.window)
	
	w.searchEntry.FocusGained()
}

// HandleEscape handles escape key press (UX Standard).
func (w *MainWindow) HandleEscape() {
	if w.fullscreen {
		w.ToggleFullscreen()
		return
	}
	
	if w.searchOpen {
		w.searchOpen = false
		return
	}
}

// ExportLog exports the log to the provided writer (UX Standard Section 7.3).
func (w *MainWindow) ExportLog(writer fyne.URIWriteCloser) {
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
func (w *MainWindow) ClearLog() {
	w.logPanel.Clear()
}

// IsFullscreen returns whether the window is in fullscreen mode.
func (w *MainWindow) IsFullscreen() bool {
	return w.fullscreen
}
