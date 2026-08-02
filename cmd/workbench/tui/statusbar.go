package tui

import (
	"strings"
	"time"
)

// StatusBar represents the status bar at the bottom of the screen.
type StatusBar struct {
	Mode        string
	Connection  string
	Address     string
	IIN         string
	Error       string
	AutoRead    bool
	AutoWrite   bool
}

// NewStatusBar creates a new status bar.
func NewStatusBar() *StatusBar {
	return &StatusBar{
		Mode:       "N/A",
		Connection: "Disconnected",
		Address:    "",
		IIN:        "",
		AutoRead:   false,
		AutoWrite:  false,
	}
}

// SetMode sets the mode (Master/Outstation).
func (s *StatusBar) SetMode(mode string) {
	s.Mode = strings.ToUpper(mode)
}

// SetConnection sets the connection status.
func (s *StatusBar) SetConnection(status string, address string) {
	s.Connection = status
	s.Address = address
}

// SetIIN sets the IIN (Interrupt/Information flag).
func (s *StatusBar) SetIIN(iin string) {
	s.IIN = iin
}

// SetError sets the error message.
func (s *StatusBar) SetError(err string) {
	s.Error = err
}

// ClearError clears the error message.
func (s *StatusBar) ClearError() {
	s.Error = ""
}

// SetAutoRead sets the auto-read status.
func (s *StatusBar) SetAutoRead(enabled bool) {
	s.AutoRead = enabled
}

// SetAutoWrite sets the auto-write status.
func (s *StatusBar) SetAutoWrite(enabled bool) {
	s.AutoWrite = enabled
}

// Draw renders the status bar.
func (s *StatusBar) Draw(scr *Screen, width int) {
	y := scr.height

	// Build status text
	status := ""
	if s.Mode != "" {
		status += "[" + s.Mode + "] "
	}
	if s.Connection != "" {
		status += s.Connection
	}
	if s.Address != "" {
		status += " " + s.Address
	}
	if s.IIN != "" {
		status += " | IIN: " + s.IIN
	}

	// Draw background
	scr.DrawHeader(y-1, strings.Repeat(" ", width), "black", "white")

	// Draw status
	scr.Print(y-1, 2, status)

	// Draw AutoRead/AutoWrite status in the middle area
	autoStatus := ""
	if s.AutoRead {
		autoStatus += "[AutoR] "
	}
	if s.AutoWrite {
		autoStatus += "[AutoW] "
	}
	if autoStatus != "" {
		scr.PrintStyled(y-1, width/2-len(autoStatus)/2, autoStatus, "yellow")
	}

	// Draw error if present
	if s.Error != "" {
		scr.PrintStyled(y-1, width-len(s.Error)-2, "⚠ "+s.Error, "red")
	}

	// Draw timestamp
	timeStr := time.Now().Format("15:04:05")
	scr.Print(y-1, width-len(timeStr)-1, timeStr)
}

// DrawControls renders the controls line.
func DrawControls(scr *Screen, width int, controls []string) {
	y := scr.height

	// Draw separator
	scr.DrawSeparator(y, "─")

	// Build controls string
	ctrlStr := " " + strings.Join(controls, " │ ") + " "

	// Draw controls
	scr.Print(y, 1, ctrlStr)
}
