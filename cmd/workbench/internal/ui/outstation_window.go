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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"dnp3/cmd/workbench/internal/config"
	outstationctrl "dnp3/cmd/workbench/internal/outstation"
	"dnp3/cmd/workbench/internal/ui/panels"
	"dnp3/pkg/dnp3/types"
)

// OutstationWindow represents the Outstation mode window.
type OutstationWindow struct {
	app    fyne.App
	window fyne.Window
	ctrl   *outstationctrl.Controller

	// Panels
	serverPanel      *ServerPanel
	simulationPanel  *SimulationPanel
	dataPointsPanel  *DataPointsPanel
	logPanel         *panels.LogPanel
	statusBar        *panels.StatusBar

	// State bindings
	state             binding.String
	serverBinding     binding.String
	mastersBinding    binding.String
	simulationBinding binding.String

	mu     sync.RWMutex
	closed bool
}

// ServerPanel represents the server configuration panel.
type ServerPanel struct {
	container *fyne.Container
	address   *widget.Entry
	port      *widget.Entry
	onStart   func(address string, port int)
	onStop    func()
}

func NewServerPanel() *ServerPanel {
	p := &ServerPanel{}

	title := widget.NewLabel("SERVER")
	title.TextStyle.Bold = true

	p.address = widget.NewEntry()
	p.address.SetText("0.0.0.0")
	p.address.SetPlaceHolder("Listen address")

	p.port = widget.NewEntry()
	p.port.SetText("20000")
	p.port.SetPlaceHolder("Port")

	startBtn := widget.NewButton("Start Server", func() {
		// Will be set by event handler
	})
	startBtn.Importance = widget.HighImportance

	stopBtn := widget.NewButton("Stop Server", func() {
		if p.onStop != nil {
			p.onStop()
		}
	})

	p.container = container.NewVBox(
		title,
		widget.NewLabel("Address:"),
		p.address,
		widget.NewLabel("Port:"),
		p.port,
		layout.NewSpacer(),
		startBtn,
		stopBtn,
	)

	// Set default onStart to change button state
	p.onStart = func(addr string, port int) {
		startBtn.SetText("Starting...")
		startBtn.Disable()
	}

	return p
}

func (p *ServerPanel) SetOnStart(f func(address string, port int)) {
	p.onStart = f
}

func (p *ServerPanel) SetOnStop(f func()) {
	p.onStop = f
}

// SimulationPanel represents the simulation configuration panel.
type SimulationPanel struct {
	container      *fyne.Container
	enabled        *widget.Check
	updateRate     *widget.Entry
	binaryRate     *widget.Entry
	analogVariance *widget.Entry
}

func NewSimulationPanel() *SimulationPanel {
	p := &SimulationPanel{}

	title := widget.NewLabel("SIMULATION")
	title.TextStyle.Bold = true

	p.enabled = widget.NewCheck("Enable Simulation", func(enabled bool) {
		// Will be set by event handler
	})
	p.enabled.SetChecked(true)

	p.updateRate = widget.NewEntry()
	p.updateRate.SetText("1.0")
	p.updateRate.SetPlaceHolder("Update rate (seconds)")

	p.binaryRate = widget.NewEntry()
	p.binaryRate.SetText("0.5")
	p.binaryRate.SetPlaceHolder("Binary toggle rate (Hz)")

	p.analogVariance = widget.NewEntry()
	p.analogVariance.SetText("10.0")
	p.analogVariance.SetPlaceHolder("Analog variance (±)")

	p.container = container.NewVBox(
		title,
		p.enabled,
		widget.NewLabel("Update Rate (sec):"),
		p.updateRate,
		widget.NewLabel("Binary Rate (Hz):"),
		p.binaryRate,
		widget.NewLabel("Analog Variance:"),
		p.analogVariance,
	)

	return p
}

// DataPointsPanel represents the data points display panel.
type DataPointsPanel struct {
	container    *fyne.Container
	binaryLabel  *widget.Label
	analogLabel  *widget.Label
	counterLabel *widget.Label
}

func NewDataPointsPanel() *DataPointsPanel {
	p := &DataPointsPanel{}

	title := widget.NewLabel("DATA POINTS")
	title.TextStyle.Bold = true

	p.binaryLabel = widget.NewLabel("Binary Inputs: No data")
	p.analogLabel = widget.NewLabel("Analog Inputs: No data")
	p.counterLabel = widget.NewLabel("Counters: No data")

	p.container = container.NewVBox(
		title,
		widget.NewLabel(""),
		widget.NewLabel("Binary Inputs (8):"),
		p.binaryLabel,
		widget.NewLabel(""),
		widget.NewLabel("Analog Inputs (4):"),
		p.analogLabel,
		widget.NewLabel(""),
		widget.NewLabel("Counters (4):"),
		p.counterLabel,
	)

	return p
}

func (p *DataPointsPanel) Update(binary []*types.BinaryInput, analog []*types.AnalogInput, counters []*types.Counter) {
	// Build binary string
	binaryStr := ""
	for i, b := range binary {
		if i > 0 {
			binaryStr += ", "
		}
		binaryStr += fmt.Sprintf("BI%d:%v", i, b.Value)
	}
	if binaryStr == "" {
		binaryStr = "No data"
	}
	p.binaryLabel.SetText(binaryStr)

	// Build analog string
	analogStr := ""
	for i, a := range analog {
		if i > 0 {
			analogStr += ", "
		}
		analogStr += fmt.Sprintf("AI%d:%.1f", i, a.Value)
	}
	if analogStr == "" {
		analogStr = "No data"
	}
	p.analogLabel.SetText(analogStr)

	// Build counter string
	counterStr := ""
	for i, c := range counters {
		if i > 0 {
			counterStr += ", "
		}
		counterStr += fmt.Sprintf("C%d:%d", i, c.Value)
	}
	if counterStr == "" {
		counterStr = "No data"
	}
	p.counterLabel.SetText(counterStr)
}

// NewOutstationWindow creates a new Outstation window.
func NewOutstationWindow(app fyne.App, ctrl *outstationctrl.Controller, cfg *config.Config) *OutstationWindow {
	w := &OutstationWindow{
		app:               app,
		window:            app.NewWindow("DNP3 Outstation"),
		ctrl:              ctrl,
		state:             binding.NewString(),
		serverBinding:     binding.NewString(),
		mastersBinding:    binding.NewString(),
		simulationBinding: binding.NewString(),
	}

	w.state.Set("Stopped")
	w.serverBinding.Set("Not listening")
	w.mastersBinding.Set("0")
	w.simulationBinding.Set("Enabled")

	w.setupUI()
	w.setupEventHandlers()

	return w
}

// setupUI creates the UI components.
func (w *OutstationWindow) setupUI() {
	// Create panels
	w.serverPanel = NewServerPanel()
	w.simulationPanel = NewSimulationPanel()
	w.dataPointsPanel = NewDataPointsPanel()
	w.logPanel = panels.NewLogPanel()
	w.statusBar = panels.NewStatusBar(w.state, w.serverBinding, w.mastersBinding)

	// Left sidebar - server, simulation
	leftSidebar := container.NewVBox(
		w.serverPanel.container,
		w.simulationPanel.container,
	)

	// Right side - data points
	rightContent := container.NewVBox(
		w.dataPointsPanel.container,
	)

	// Make the split pane resizable
	mainContent := container.NewHSplit(
		leftSidebar,
		rightContent,
	)
	mainContent.Offset = 0.3 // 30% for sidebar

	// Complete layout - status bar at bottom
	content := container.NewBorder(
		nil,                     // Top
		w.statusBar.Container(), // bottom
		nil,                     // left
		nil,                     // right
		container.NewVBox(
			mainContent,
			w.logPanel.Container(),
		),
	)

	w.window.SetContent(content)
}

// setupEventHandlers sets up event handling between panels.
func (w *OutstationWindow) setupEventHandlers() {
	// Server panel events
	w.serverPanel.SetOnStart(func(address string, port int) {
		w.ctrl.StartServer(address, port)
	})

	w.serverPanel.SetOnStop(func() {
		w.ctrl.Stop()
	})

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

	// Start state polling
	go w.pollState()
}

// pollState periodically checks controller state and updates UI.
func (w *OutstationWindow) pollState() {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	var lastState *outstationctrl.State
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
func (w *OutstationWindow) updateUI(state *outstationctrl.State) {
	// Update state display
	if state.Running {
		w.state.Set("Listening")
		w.serverBinding.Set(fmt.Sprintf("%s:%d", state.ListenAddress, state.ListenPort))
	} else {
		w.state.Set("Stopped")
		w.serverBinding.Set("Not listening")
	}

	w.mastersBinding.Set(fmt.Sprintf("%d", state.ConnectedMasters))

	if state.SimulationEnabled {
		w.simulationBinding.Set("Enabled")
	} else {
		w.simulationBinding.Set("Disabled")
	}

	// Update data points
	w.dataPointsPanel.Update(
		w.ctrl.GetBinaryInputs(),
		w.ctrl.GetAnalogInputs(),
		w.ctrl.GetCounters(),
	)

	// Update status bar
	w.updateStatusBar(state)
}

// updateStatusBar updates the status bar with connection state.
func (w *OutstationWindow) updateStatusBar(state *outstationctrl.State) {
	if state.Running {
		w.statusBar.SetConnectionState(panels.ConnectionStateConnected, "Listening")
		w.statusBar.ClearError()
	} else {
		w.statusBar.SetConnectionState(panels.ConnectionStateDisconnected, "Stopped")
	}

	if state.LastError != "" {
		w.statusBar.ShowError(state.LastError)
	}
}

// Window returns the underlying Fyne window.
func (w *OutstationWindow) Window() fyne.Window {
	return w.window
}

// Show shows the window.
func (w *OutstationWindow) Show() {
	w.window.Show()
}

// Resize resizes the window.
func (w *OutstationWindow) Resize(size fyne.Size) {
	w.window.Resize(size)
}

// SetTitle sets the window title.
func (w *OutstationWindow) SetTitle(title string) {
	w.window.SetTitle(title)
}

// CenterOnScreen centers the window on screen.
func (w *OutstationWindow) CenterOnScreen() {
	w.window.CenterOnScreen()
}

// SetMainMenu sets the main menu for the window.
func (w *OutstationWindow) SetMainMenu(menu *fyne.MainMenu) {
	w.window.SetMainMenu(menu)
}

// Maximize maximizes the window.
func (w *OutstationWindow) Maximize() {
	w.window.SetFullScreen(true)
}

// Restore restores the window.
func (w *OutstationWindow) Restore() {
	w.window.SetFullScreen(false)
}

// ToggleSidebar shows/hides the sidebar panel.
func (w *OutstationWindow) ToggleSidebar() {
	// Placeholder for future implementation
}

// ToggleLogPanel shows/hides the log panel.
func (w *OutstationWindow) ToggleLogPanel() {
	// Placeholder for future implementation
}

// ShowLogSearch shows the log search bar.
func (w *OutstationWindow) ShowLogSearch() {
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
func (w *OutstationWindow) ExportLog(writer fyne.URIWriteCloser) {
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
func (w *OutstationWindow) ClearLog() {
	w.logPanel.Clear()
}
