// Package ui provides the user interface for the DNP3 Engineering Workbench.
package ui

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"

	"dnp3/cmd/workbench/internal/controller"
	"dnp3/cmd/workbench/internal/logger"
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

	// State bindings
	state        binding.String
	connectionBinding binding.String
	iinBinding       binding.String

	mu     sync.RWMutex
	closed bool
}

// NewMainWindow creates a new main window.
func NewMainWindow(app fyne.App, ctrl *controller.Controller) *MainWindow {
	w := &MainWindow{
		app:    app,
		window:  app.NewWindow("DNP3 Engineering Workbench"),
		ctrl:    ctrl,
		state:   binding.NewString(),
		connectionBinding: binding.NewString(),
		iinBinding: binding.NewString(),
	}

	w.state.Set("Disconnected")
	w.connectionBinding.Set("Not Connected")
	w.iinBinding.Set("0x0000")

	w.setupUI()
	w.setupEventHandlers()
	w.setupControllerCallbacks()

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

	// Complete layout
	content := container.NewBorder(
		nil, // top
		w.statusBar.Container(), // bottom
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
}

// setupControllerCallbacks sets up callbacks for controller events.
func (w *MainWindow) setupControllerCallbacks() {
	// State change callback
	w.ctrl.OnStateChange = func(state *controller.AppState) {
		fyne.NewRunnable(func() {
			w.state.Set(state.Connection.String())
			
			if state.Address != "" {
				w.connectionBinding.Set(fmt.Sprintf("%s:%d", state.Address, state.Port))
			}
			
			if state.LastResponse != nil {
				w.iinBinding.Set(fmt.Sprintf("0x%02X%02X", state.LastResponse.IIN[0], state.LastResponse.IIN[1]))
				w.dataPanel.Update(state.LastResponse)
			}
			
			// Update connection panel state
			w.connectionPanel.SetConnected(state.Connection == controller.StateConnected)
		})
	}

	// Log entry callback
	w.ctrl.OnLogEntry = func(entry *logger.Entry) {
		fyne.NewRunnable(func() {
			direction := "INFO"
			if entry.Level == logger.LevelError {
				direction = "ERROR"
			} else if entry.Level == logger.LevelDebug {
				direction = "DEBUG"
			}
			w.logPanel.Append(entry.Timestamp, direction, entry.Message)
		})
	}
}

// ShowAndRun shows the window and runs the application.
func (w *MainWindow) ShowAndRun() {
	w.window.Show()
}
