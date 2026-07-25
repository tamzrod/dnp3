// Package ui provides the user interface for the DNP3 Engineering Workbench.
package ui

import (
	"context"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"dnp3/cmd/workbench/internal/session"
	"dnp3/cmd/workbench/internal/ui/panels"
	"dnp3/pkg/dnp3/types"
)

// MainWindow represents the main application window.
type MainWindow struct {
	app      fyne.App
	window   fyne.Window
	manager  *session.Manager

	// Panels
	modePanel       *panels.ModePanel
	connectionPanel *panels.ConnectionPanel
	commandPanel    *panels.CommandPanel
	dataPanel       *panels.DataPanel
	protocolPanel   *panels.ProtocolPanel
	logPanel        *panels.LogPanel
	statusBar       *panels.StatusBar

	// State
	state        binding.String
	connectionBinding binding.String
	iinBinding       binding.String

	mu     sync.RWMutex
	closed bool
}

// NewMainWindow creates a new main window.
func NewMainWindow(app fyne.App, manager *session.Manager) *MainWindow {
	w := &MainWindow{
		app:      app,
		window:    app.NewWindow("DNP3 Engineering Workbench"),
		manager:   manager,
		state:     binding.NewString(),
		connectionBinding: binding.NewString(),
		iinBinding: binding.NewString(),
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
	w.connectionPanel = panels.NewConnectionPanel(w.manager)
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

	// Main content area
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
		w.handleConnect(address, port)
	}

	w.connectionPanel.OnDisconnect = func() {
		w.handleDisconnect()
	}

	// Command panel events
	w.commandPanel.OnReadClass = func(class int) {
		w.handleReadClass(class)
	}

	w.commandPanel.OnOperate = func(index uint16, value bool) {
		w.handleOperate(index, value)
	}

	// Log panel events
	w.logPanel.OnClear = func() {
		w.logPanel.Clear()
	}
}

// handleConnect handles connection requests.
func (w *MainWindow) handleConnect(address string, port int) {
	w.state.Set("Connecting...")
	w.connectionBinding.Set("Connecting...")
	w.logPanel.Append(time.Now(), "TX", "Connecting to "+address+":"+string(rune(port)))

	// Create master session
	masterSession := w.manager.CreateMasterSession()

	// Connect in goroutine
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := masterSession.Connect(ctx, address, port)
		if err != nil {
			fyne.NewRunnable(func() {
				w.state.Set("Error")
				w.connectionBinding.Set("Connection failed")
				w.logPanel.Append(time.Now(), "ERROR", "Connection failed: "+err.Error())
			})
			return
		}

		fyne.NewRunnable(func() {
			w.state.Set("Connected")
			w.connectionBinding.Set(address + ":" + itoa(port))
			w.iinBinding.Set("0x0000")
			w.logPanel.Append(time.Now(), "INFO", "Connected successfully")
			w.connectionPanel.SetConnected(true)
		})

		// Handle session events
		for {
			select {
			case event := <-masterSession.Events():
				w.handleSessionEvent(event)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// handleDisconnect handles disconnection.
func (w *MainWindow) handleDisconnect() {
	w.state.Set("Disconnecting...")
	w.logPanel.Append(time.Now(), "TX", "Disconnecting")

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		w.manager.Close(ctx)

		fyne.NewRunnable(func() {
			w.state.Set("Disconnected")
			w.connectionBinding.Set("Not Connected")
			w.iinBinding.Set("0x0000")
			w.logPanel.Append(time.Now(), "INFO", "Disconnected")
			w.connectionPanel.SetConnected(false)
			w.dataPanel.Clear()
		})
	}()
}

// handleReadClass handles read class commands.
func (w *MainWindow) handleReadClass(class int) {
	session := w.manager.GetSession()
	if session == nil || session.State() != session.StateConnected {
		w.logPanel.Append(time.Now(), "ERROR", "Not connected")
		return
	}

	// Build read request based on class
	var groups []types.GroupRequest
	switch class {
	case 0:
		groups = []types.GroupRequest{{Group: 60, Variation: 1}} // All static data
	case 1:
		groups = []types.GroupRequest{{Group: 60, Variation: 2}} // Class 1 events
	case 2:
		groups = []types.GroupRequest{{Group: 60, Variation: 3}} // Class 2 events
	case 3:
		groups = []types.GroupRequest{{Group: 60, Variation: 4}} // Class 3 events
	}

	cmd := &session.ReadCommand{Groups: groups}
	w.logPanel.Append(time.Now(), "TX", "Read Class "+itoa(class))

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := session.SendCommand(ctx, cmd)
		if err != nil {
			fyne.NewRunnable(func() {
				w.logPanel.Append(time.Now(), "ERROR", "Read failed: "+err.Error())
			})
			return
		}

		fyne.NewRunnable(func() {
			w.logPanel.Append(time.Now(), "RX", "Response received")
			w.iinBinding.Set(formatIIN(resp.IIN))
			w.dataPanel.Update(resp)
			w.protocolPanel.Update(resp)
		})
	}()
}

// handleOperate handles operate commands.
func (w *MainWindow) handleOperate(index uint16, value bool) {
	session := w.manager.GetSession()
	if session == nil || session.State() != session.StateConnected {
		w.logPanel.Append(time.Now(), "ERROR", "Not connected")
		return
	}

	cmd := &session.OperateCommand{
		Group:      12, // Binary Output
		Variation:  1,
		Index:      index,
		Value:      value,
		SelectThenOperate: true,
	}

	w.logPanel.Append(time.Now(), "TX", "Operate DO Index="+itoa(int(index))+" Value="+boolToString(value))

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := session.SendCommand(ctx, cmd)
		if err != nil {
			fyne.NewRunnable(func() {
				w.logPanel.Append(time.Now(), "ERROR", "Operate failed: "+err.Error())
			})
			return
		}

		fyne.NewRunnable(func() {
			w.logPanel.Append(time.Now(), "RX", "Operate response received")
			w.iinBinding.Set(formatIIN(resp.IIN))
		})
	}()
}

// handleSessionEvent handles session events.
func (w *MainWindow) handleSessionEvent(event session.SessionEvent) {
	fyne.NewRunnable(func() {
		switch event.Type {
		case "response":
			if resp, ok := event.Data.(*session.Response); ok {
				w.dataPanel.Update(resp)
				w.protocolPanel.Update(resp)
			}
		case "error":
			w.logPanel.Append(time.Now(), "ERROR", event.Data.(string))
		}
	})
}

// ShowAndRun shows the window and runs the application.
func (w *MainWindow) ShowAndRun() {
	w.window.Show()
}

// Helper functions
func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func formatIIN(iin [2]byte) string {
	return fmt.Sprintf("0x%02X%02X", iin[0], iin[1])
}
