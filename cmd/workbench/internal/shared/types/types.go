// Package types provides shared types for Master and Outstation controllers.
package types

import "errors"

// Mode represents the operating mode of the workbench.
type Mode string

const (
	// ModeMaster indicates Master mode (connect to outstation).
	ModeMaster Mode = "master"
	// ModeOutstation indicates Outstation mode (run as server).
	ModeOutstation Mode = "outstation"
	// ModeSelect indicates mode selection dialog should be shown.
	ModeSelect Mode = "select"
)

// String returns the string representation of the mode.
func (m Mode) String() string {
	return string(m)
}

// IsMaster returns true if the mode is Master.
func (m Mode) IsMaster() bool {
	return m == ModeMaster
}

// IsOutstation returns true if the mode is Outstation.
func (m Mode) IsOutstation() bool {
	return m == ModeOutstation
}

// Validate checks if the mode is valid.
func (m Mode) Validate() error {
	switch m {
	case ModeMaster, ModeOutstation, ModeSelect:
		return nil
	default:
		return errors.New("invalid mode: must be 'master', 'outstation', or 'select'")
	}
}

// WindowConfig holds window configuration.
type WindowConfig struct {
	Title     string
	Width     int
	Height    int
	MinWidth  int
	MinHeight int
}

// DefaultMasterWindowConfig returns default configuration for Master window.
func DefaultMasterWindowConfig() *WindowConfig {
	return &WindowConfig{
		Title:     "DNP3 Master - Connect to Outstation",
		Width:     1200,
		Height:    800,
		MinWidth:  800,
		MinHeight: 600,
	}
}

// DefaultOutstationWindowConfig returns default configuration for Outstation window.
func DefaultOutstationWindowConfig() *WindowConfig {
	return &WindowConfig{
		Title:     "DNP3 Outstation - Simulate Data",
		Width:     1200,
		Height:    800,
		MinWidth:  800,
		MinHeight: 600,
	}
}
