package tui

import (
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

// Mode represents the operating mode.
type Mode string

const (
	ModeMaster    Mode = "master"
	ModeOutstation Mode = "outstation"
)

// App represents the main TUI application.
type App struct {
	Mode    Mode
	Layout  *Layout
	Screen  *Screen
	Input   *Input
	Table   *Table
	Log     *Log
	Status  *StatusBar

	// Data
	dataMu    sync.RWMutex
	dataRows  []Row

	// Callbacks
	OnConnect    func()
	OnDisconnect func()
	OnStart      func()
	OnStop       func()
	OnReadClass  func(class int)
	OnOperate    func(index int, value bool)
	OnQuit       func()

	// State
	running bool
	done    chan struct{}
}

// NewApp creates a new TUI application.
func NewApp(mode Mode) *App {
	width, height, _ := term.GetSize(int(os.Stdout.Fd()))
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	layout := NewLayout(width, height)
	screen := NewScreen(width, height)

	app := &App{
		Mode:   mode,
		Layout: layout,
		Screen: screen,
		Input:  NewInput(),
		Log:    NewLog(layout.LogBounds()),
		Status: NewStatusBar(),
		done:   make(chan struct{}),
	}

	// Set up table
	tableBounds := layout.TableBounds()
	app.Table = NewTable(tableBounds)
	app.Table.SetColumns([]Column{
		{Title: "Type", Width: 6},
		{Title: "Index", Width: 6},
		{Title: "Value", Width: 12},
		{Title: "Quality", Width: 10},
		{Title: "Timestamp", Width: 12},
	})

	// Set mode in status bar
	if mode == ModeMaster {
		app.Status.SetMode("MASTER")
	} else {
		app.Status.SetMode("OUTSTATION")
	}

	return app
}

// Run starts the application.
func (a *App) Run() error {
	a.running = true

	// Enable raw mode
	if err := a.Input.EnableRawMode(); err != nil {
		return err
	}
	defer a.Input.DisableRawMode()

	// Hide cursor
	os.Stdout.WriteString(HideCursor)
	defer os.Stdout.WriteString(ShowCursor)

	// Clear screen
	a.Screen.Clear()
	a.Screen.Flush()

	// Start input handling
	events := a.Input.Events()

	// Start render loop
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Initial draw
	a.draw()
	a.Screen.Flush()

	// Main loop
	for a.running {
		select {
		case <-a.done:
			return nil
		case <-ticker.C:
			// Periodic redraw (for live data updates)
			a.draw()
			a.Screen.Flush()
		case event, ok := <-events:
			if !ok {
				return nil
			}
			if a.handleEvent(event) {
				a.draw()
				a.Screen.Flush()
			}
		}
	}

	return nil
}

// Stop stops the application.
func (a *App) Stop() {
	a.running = false
	close(a.done)
	a.Input.Stop()
}

// handleEvent handles an input event.
func (a *App) handleEvent(event Event) bool {
	switch event.Type {
	case EventResize:
		if width, height, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
			a.Layout.Resize(width, height)
			a.Screen.width = width
			a.Screen.height = height
		}
		return true
	case EventKey:
		return a.handleKey(event.Key, event.Rune)
	}
	return false
}

// handleKey handles a key press.
func (a *App) handleKey(key Key, r rune) bool {
	switch key {
	case KeyEscape:
		// Quit
		a.Log.Info("Exiting...")
		a.Stop()
		if a.OnQuit != nil {
			a.OnQuit()
		}
		return true
	case KeyUp:
		a.Table.MoveUp()
		return true
	case KeyDown:
		a.Table.MoveDown()
		return true
	case KeyEnter:
		a.Table.Select()
		selected := a.Table.GetSelected()
		if selected >= 0 && a.OnOperate != nil {
			a.OnOperate(selected, true)
		}
		return true
	}

	// Handle character keys
	switch r {
	case 'q', 'Q':
		a.Log.Info("Exiting...")
		a.Stop()
		if a.OnQuit != nil {
			a.OnQuit()
		}
		return true
	case 'm', 'M':
		a.toggleMode()
		return true
	case 's', 'S':
		if a.OnStart != nil {
			a.OnStart()
		}
		return true
	case 'x', 'X':
		if a.OnStop != nil {
			a.OnStop()
		}
		return true
	case 'r', 'R':
		if a.OnReadClass != nil {
			a.OnReadClass(0)
		}
		return true
	case '1':
		if a.OnReadClass != nil {
			a.OnReadClass(1)
		}
		return true
	case '2':
		if a.OnReadClass != nil {
			a.OnReadClass(2)
		}
		return true
	case '3':
		if a.OnReadClass != nil {
			a.OnReadClass(3)
		}
		return true
	case 'l', 'L':
		a.Log.Clear()
		return true
	case 'h', 'H', '?':
		a.showHelp()
		return true
	}
	return false
}

// toggleMode toggles between Master and Outstation mode.
func (a *App) toggleMode() {
	a.Log.Info("Switching mode...")
	
	// Stop current operation
	if a.OnStop != nil {
		a.OnStop()
	}
	
	// Toggle mode
	if a.Mode == ModeMaster {
		a.Mode = ModeOutstation
		a.Log.Info("Switched to Outstation mode")
	} else {
		a.Mode = ModeMaster
		a.Log.Info("Switched to Master mode")
	}
	
	// Update status
	a.Status.Mode = string(a.Mode)
	
	// Clear data table
	a.Table.Clear()
	a.UpdateData(nil)
}

// draw redraws the entire screen.
func (a *App) draw() {
	s := a.Screen

	// Clear screen
	s.Clear()

	// Draw header
	a.drawHeader()

	// Draw table
	a.drawTable()

	// Draw footer
	a.drawFooter()
}

// drawHeader draws the header bar.
func (a *App) drawHeader() {
	s := a.Screen
	width := s.width

	// Mode indicator
	modeStr := "[" + string(a.Mode) + "]"
	s.PrintStyled(1, 2, modeStr, "brightcyan", "bold")

	// Title
	title := " DNP3 Engineering Workbench "
	s.Print(1, 2+len(modeStr)+2, title)

	// Connection status on right
	connStr := a.Status.Connection
	if a.Status.Address != "" {
		connStr += " " + a.Status.Address
	}
	s.Print(1, width-len(connStr)-2, connStr)

	// Draw separator
	s.DrawSeparator(2, "─")
}

// drawTable draws the data table.
func (a *App) drawTable() {
	s := a.Screen

	// Draw table title
	tableBounds := a.Layout.TableBounds()
	s.PrintStyled(tableBounds.Top, 2, "DATA POINTS", "cyan", "bold")

	// Draw table
	a.dataMu.RLock()
	a.Table.SetRows(a.dataRows)
	a.dataMu.RUnlock()
	a.Table.DrawSimple(s, tableBounds.Top+2)
}

// drawLog draws the log panel.
func (a *App) drawLog() {
	s := a.Screen
	logBounds := a.Layout.LogBounds()

	// Draw log title
	s.PrintStyled(logBounds.Top, 2, "LOG", "cyan", "bold")

	// Draw log entries
	a.Log.DrawSimple(s, logBounds.Top+2, logBounds.Bottom-logBounds.Top)
}

// drawFooter draws the footer/controls.
func (a *App) drawFooter() {
	s := a.Screen
	height := s.height

	// Draw separator
	s.DrawSeparator(height-1, "─")

	// Draw controls
	controls := []string{
		"[m]ode",
		"[s]tart",
		"[x]stop",
		"[r]ead",
		"[1-3] class",
		"[↑↓] nav",
		"[l]og",
		"[h]elp",
		"[q]uit",
	}

	// Draw controls
	ctrlStr := " " + joinControls(controls) + " "
	s.Print(height, 1, ctrlStr)
}

// joinControls joins control strings with separator.
func joinControls(controls []string) string {
	result := controls[0]
	for _, c := range controls[1:] {
		result += " │ " + c
	}
	return result
}

// showHelp shows the help screen.
func (a *App) showHelp() {
	s := a.Screen
	width := s.width

	// Draw help overlay
	s.FillRect(5, 10, 20, width-10, " ", "black", "white")

	// Draw box
	s.DrawBox(5, 10, 20, width-10, "HELP", "cyan")

	help := []string{
		"q, Esc    Quit the application",
		"c         Connect (Master mode)",
		"d         Disconnect",
		"r         Read Class 0",
		"1-3       Read Class 1-3",
		"↑, k      Move cursor up",
		"↓, j      Move cursor down",
		"Enter     Select/Operate",
		"l         Clear log",
		"h, ?      Show this help",
	}

	for i, line := range help {
		s.Print(6+i, 12, line)
	}

	s.Flush()

	// Wait for key press
	events := a.Input.Events()
	<-events
}

// UpdateData updates the table data.
func (a *App) UpdateData(rows []Row) {
	a.dataMu.Lock()
	defer a.dataMu.Unlock()
	a.dataRows = rows
}

// AddDataPoint adds a data point to the table.
func (a *App) AddDataPoint(pointType string, index int, value string, quality string, timestamp string) {
	a.dataMu.Lock()
	defer a.dataMu.Unlock()

	row := Row{Cells: []string{
		pointType,
		fmt.Sprintf("%d", index),
		value,
		quality,
		timestamp,
	}}
	a.dataRows = append(a.dataRows, row)
}

// SetConnection updates the connection status.
func (a *App) SetConnection(status string, address string) {
	a.Status.SetConnection(status, address)
}

// SetError updates the error status.
func (a *App) SetError(err string) {
	a.Status.SetError(err)
}

// LogInfo logs an info message.
func (a *App) LogInfo(msg string) {
	a.Log.Info(msg)
}

// LogSend logs a sent message.
func (a *App) LogSend(msg string) {
	a.Log.Send(msg)
}

// LogRecv logs a received message.
func (a *App) LogRecv(msg string) {
	a.Log.Recv(msg)
}

// LogError logs an error message.
func (a *App) LogError(msg string) {
	a.Log.Error(msg)
}
