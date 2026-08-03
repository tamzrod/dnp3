// DNP3 Engineering Workbench
// A terminal-based application for validating and debugging the native Go DNP3 library.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"dnp3/cmd/workbench/internal/logger"
	masterctrl "dnp3/cmd/workbench/internal/master"
	outstationctrl "dnp3/cmd/workbench/internal/outstation"
	"dnp3/cmd/workbench/internal/session"
	"dnp3/cmd/workbench/tui"
	"dnp3/pkg/dnp3/types"
)

// Global state for data updates
var (
	updateMu sync.RWMutex
	updateCh chan struct{}
)



func main() {
	// Parse command-line flags
	modeStr := flag.String("mode", "master", "Operating mode: master or outstation")
	address := flag.String("address", "127.0.0.1", "Remote address (Master mode)")
	port := flag.Int("port", 20000, "Port number")
	flag.Parse()

	// Validate mode
	mode := tui.Mode(strings.ToLower(*modeStr))
	if mode != tui.ModeMaster && mode != tui.ModeOutstation {
		fmt.Fprintf(os.Stderr, "Invalid mode: %s (use 'master' or 'outstation')\n", *modeStr)
		os.Exit(1)
	}

	// Create channel for updates
	updateCh = make(chan struct{}, 1)

	// Create TUI application
	app := tui.NewApp(mode)

	// Set up logging
	app.LogInfo("DNP3 Engineering Workbench starting...")
	app.LogInfo(fmt.Sprintf("Mode: %s", mode))

	// Set up callbacks based on mode
	if mode == tui.ModeMaster {
		setupMaster(app, *address, *port)
	} else {
		setupOutstation(app, *address, *port)
	}

	// Handle quit - clear terminal and restore cursor
	app.OnQuit = func() {
		app.LogInfo("Shutting down...")
		// Clear screen and show cursor on exit
		os.Stdout.WriteString(tui.ClearScreen)
		os.Stdout.WriteString(tui.ShowCursor)
	}

	// Run the application
	if err := app.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

// setupMaster sets up Master mode callbacks.
func setupMaster(app *tui.App, address string, port int) {
	// Create logger
	log := logger.New()

	// Create controller
	ctrl := masterctrl.NewController(log)

	// Start the controller (enables auto-poll and auto-write handlers)
	if err := ctrl.Start(); err != nil {
		app.LogError(fmt.Sprintf("Failed to start controller: %v", err))
	}

	// Start callback (connect)
	app.OnStart = func() {
		app.LogInfo(fmt.Sprintf("Connecting to %s:%d...", address, port))
		if err := ctrl.Connect(address, port); err != nil {
			app.LogError(fmt.Sprintf("Connection failed: %v", err))
		} else {
			app.LogInfo("Connected!")
			app.SetConnection("Connected", fmt.Sprintf("%s:%d", address, port))
			// Auto-read Class 0 on connect
			app.LogInfo("Auto-reading Class 0...")
			if err := ctrl.ReadClass(0); err != nil {
				app.LogError(fmt.Sprintf("Initial read failed: %v", err))
			}
		}
	}

	// Stop callback (disconnect)
	app.OnStop = func() {
		app.LogInfo("Disconnecting...")
		// Disable simulation mode on disconnect
		if ctrl.IsSimulationModeEnabled() {
			ctrl.EnableSimulationMode(false)
		} else {
			if ctrl.IsAutoPollEnabled() {
				ctrl.EnableAutoPoll(false)
			}
			if ctrl.IsAutoWriteEnabled() {
				ctrl.EnableAutoWrite(false)
			}
		}
		if err := ctrl.Disconnect(); err != nil {
			app.LogError(fmt.Sprintf("Disconnect failed: %v", err))
		} else {
			app.LogInfo("Disconnected")
			app.SetConnection("Disconnected", "")
		}
		// Update TUI status
		app.SetAutoRead(false)
		app.SetAutoWrite(false)
	}

	// Read class callback
	app.OnReadClass = func(class int) {
		app.LogSend(fmt.Sprintf("Read Class %d", class))
		if err := ctrl.ReadClass(class); err != nil {
			app.LogError(fmt.Sprintf("Read failed: %v", err))
		}
	}

	// Operate callback
	app.OnOperate = func(index int, value bool) {
		app.LogSend(fmt.Sprintf("Operate BO%d=%v", index, value))
		if err := ctrl.Operate(uint16(index), value); err != nil {
			app.LogError(fmt.Sprintf("Operate failed: %v", err))
		}
	}

	// Auto-poll toggle callback (auto-read)
	app.OnAutoPollToggle = func() {
		if ctrl.IsAutoPollEnabled() {
			ctrl.EnableAutoPoll(false)
			app.LogInfo("Auto-read DISABLED")
			app.SetAutoRead(false)
		} else {
			// Disable simulation mode if active
			if ctrl.IsSimulationModeEnabled() {
				ctrl.EnableSimulationMode(false)
			}
			ctrl.EnableAutoPoll(true)
			app.LogInfo("Auto-read ENABLED (1s)")
			app.SetAutoRead(true)
		}
	}

	// Auto-write toggle callback
	app.OnAutoWriteToggle = func() {
		state := ctrl.State()
		if state.Connection != session.StateConnected {
			app.LogError("Not connected - cannot enable auto-write")
			return
		}
		
		if ctrl.IsAutoWriteEnabled() {
			ctrl.EnableAutoWrite(false)
			app.LogInfo("Auto-write DISABLED")
			app.SetAutoWrite(false)
		} else {
			// Disable simulation mode if active
			if ctrl.IsSimulationModeEnabled() {
				ctrl.EnableSimulationMode(false)
			}
			ctrl.EnableAutoWrite(true)
			app.LogInfo("Auto-write ENABLED (random operate)")
			app.SetAutoWrite(true)
		}
	}

	// Simulation mode toggle callback
	app.OnSimulationModeToggle = func() {
		state := ctrl.State()
		if state.Connection != session.StateConnected {
			app.LogError("Not connected - cannot enable simulation mode")
			return
		}
		
		if ctrl.IsSimulationModeEnabled() {
			ctrl.EnableSimulationMode(false)
			app.LogInfo("Simulation mode DISABLED")
			app.SetAutoRead(false)
			app.SetAutoWrite(false)
		} else {
			ctrl.EnableSimulationMode(true)
			app.LogInfo("Simulation mode ENABLED")
			app.SetAutoRead(true)
			app.SetAutoWrite(true)
		}
	}

	// Start polling for state updates
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				state := ctrl.State()
				updateData(app, state)
				// Update auto status in TUI
				app.SetAutoRead(state.AutoPollEnabled)
				app.SetAutoWrite(state.AutoWriteEnabled)
			case <-updateCh:
				return
			}
		}
	}()
}

// setupOutstation sets up Outstation mode callbacks.
func setupOutstation(app *tui.App, address string, port int) {
	// Create logger
	log := logger.New()

	// Create controller
	ctrl := outstationctrl.NewController(log)

	// Start callback
	app.OnStart = func() {
		app.LogInfo(fmt.Sprintf("Starting server on %s:%d...", address, port))
		if err := ctrl.StartServer(address, port); err != nil {
			app.LogError(fmt.Sprintf("Server start failed: %v", err))
		} else {
			app.LogInfo("Server started!")
			app.SetConnection("Listening", fmt.Sprintf("%s:%d", address, port))
		}
	}

	// Stop callback
	app.OnStop = func() {
		app.LogInfo("Stopping server...")
		if err := ctrl.Stop(); err != nil {
			app.LogError(fmt.Sprintf("Server stop failed: %v", err))
		} else {
			app.LogInfo("Server stopped")
			app.SetConnection("Stopped", "")
		}
	}

	// Start polling for state updates
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				updateOutstationData(app, ctrl)
			case <-updateCh:
				return
			}
		}
	}()
}

// updateData updates the TUI with controller state (Master mode).
func updateData(app *tui.App, state *masterctrl.State) {
	// Build data rows from response
	var rows []tui.Row

	if state.LastResponse != nil {
		resp := state.LastResponse
		respTime := resp.Timestamp

		// Check if read returned any points
		totalPoints := len(resp.BinaryInputs) + len(resp.AnalogInputs) +
			len(resp.Counters) + len(resp.BinaryOutputs) + len(resp.AnalogOutputs)
		if totalPoints == 0 {
			app.LogError(fmt.Sprintf("Read returned 0 points: BI=%d AI=%d CTR=%d BO=%d AO=%d",
				len(resp.BinaryInputs), len(resp.AnalogInputs), len(resp.Counters),
				len(resp.BinaryOutputs), len(resp.AnalogOutputs)))
		}

		// RX Time is when the response was received (always same for all points)
		rxTimeStr := respTime.Format("15:04:05")

		for _, bi := range resp.BinaryInputs {
			quality := qualityString(bi.Quality)
			// Point Time: only show if timestamp is available
			pointTimeStr := formatPointTime(bi.Time)
			rows = append(rows, tui.Row{Cells: []string{
				"BI",
				fmt.Sprintf("%d", bi.Index),
				fmt.Sprintf("%v", bi.Value),
				quality,
				rxTimeStr,
				pointTimeStr,
			}})
		}

		// Add Binary Outputs - BO doesn't have Point Time in DNP3 objects
		for _, bo := range resp.BinaryOutputs {
			quality := qualityString(bo.Quality)
			rows = append(rows, tui.Row{Cells: []string{
				"BO",
				fmt.Sprintf("%d", bo.Index),
				fmt.Sprintf("%v", bo.Value),
				quality,
				rxTimeStr,
				"-", // BO has no Point Time in DNP3 objects
			}})
		}

		for _, ai := range resp.AnalogInputs {
			quality := qualityString(ai.Quality)
			pointTimeStr := formatPointTime(ai.Time)
			rows = append(rows, tui.Row{Cells: []string{
				"AI",
				fmt.Sprintf("%d", ai.Index),
				fmt.Sprintf("%.2f", ai.Value),
				quality,
				rxTimeStr,
				pointTimeStr,
			}})
		}

		// Add Analog Outputs - AO doesn't have Point Time in DNP3 objects
		for _, ao := range resp.AnalogOutputs {
			quality := qualityString(ao.Quality)
			rows = append(rows, tui.Row{Cells: []string{
				"AO",
				fmt.Sprintf("%d", ao.Index),
				fmt.Sprintf("%.2f", ao.Value),
				quality,
				rxTimeStr,
				"-", // AO has no Point Time in DNP3 objects
			}})
		}

		for _, c := range resp.Counters {
			quality := qualityString(c.Quality)
			pointTimeStr := formatPointTime(c.Time)
			rows = append(rows, tui.Row{Cells: []string{
				"CTR",
				fmt.Sprintf("%d", c.Index),
				fmt.Sprintf("%d", c.Value),
				quality,
				rxTimeStr,
				pointTimeStr,
			}})
		}
	}

	// Use UpdateDataIfChanged and SignalRedraw to avoid flicker
	if app.UpdateDataIfChanged(rows) {
		app.SignalRedraw()
	}
}

// formatPointTime formats a timestamp from a DNP3 object.
// Returns "-" if the timestamp is nil or null.
func formatPointTime(ts *types.Timestamp) string {
	if ts != nil && !ts.IsNull() {
		return ts.Time().Format("15:04:05.000")
	}
	return "-"
}

// formatTimestamp formats a timestamp for display.
// 1. If point.Time is set → show as HH:MM:SS.mmm (matching outstation format)
// 2. Else if respTime is set → show "RX HH:MM:SS" (labeled receive time)
// 3. Else → show "—"
func formatTimestamp(ts *types.Timestamp, respTime time.Time) string {
	if ts != nil && !ts.IsNull() {
		return ts.Time().Format("15:04:05.000")
	}
	if !respTime.IsZero() {
		return "RX " + respTime.Format("15:04:05")
	}
	return "—"
}

// formatTimestampSimple formats a timestamp for display (outstation mode).
// Returns "-" if the timestamp is nil or null.
func formatTimestampSimple(ts *types.Timestamp) string {
	if ts != nil && !ts.IsNull() {
		return ts.Time().Format("15:04:05")
	}
	return "-"
}

// updateOutstationData updates the TUI with controller state (Outstation mode).
func updateOutstationData(app *tui.App, ctrl *outstationctrl.Controller) {
	// Build data rows from simulator
	var rows []tui.Row

	binary := ctrl.GetBinaryInputs()
	for _, bi := range binary {
		quality := qualityString(bi.Quality)
		// Use formatTimestamp to display point's timestamp
		timeStr := formatTimestampSimple(bi.Time)
		rows = append(rows, tui.Row{Cells: []string{
			"BI",
			fmt.Sprintf("%d", bi.Index),
			fmt.Sprintf("%v", bi.Value),
			quality,
			timeStr,
		}})
	}

	// Add Binary Outputs
	binaryOut := ctrl.GetBinaryOutputs()
	for _, bo := range binaryOut {
		quality := qualityString(bo.Quality)
		rows = append(rows, tui.Row{Cells: []string{
			"BO",
			fmt.Sprintf("%d", bo.Index),
			fmt.Sprintf("%v", bo.Value),
			quality,
			"-",
		}})
	}

	analog := ctrl.GetAnalogInputs()
	for _, ai := range analog {
		quality := qualityString(ai.Quality)
		timeStr := formatTimestampSimple(ai.Time)
		rows = append(rows, tui.Row{Cells: []string{
			"AI",
			fmt.Sprintf("%d", ai.Index),
			fmt.Sprintf("%.2f", ai.Value),
			quality,
			timeStr,
		}})
	}

	// Add Analog Outputs
	analogOut := ctrl.GetAnalogOutputs()
	for _, ao := range analogOut {
		quality := qualityString(ao.Quality)
		rows = append(rows, tui.Row{Cells: []string{
			"AO",
			fmt.Sprintf("%d", ao.Index),
			fmt.Sprintf("%.2f", ao.Value),
			quality,
			"-",
		}})
	}

	counters := ctrl.GetCounters()
	for _, c := range counters {
		quality := qualityString(c.Quality)
		timeStr := formatTimestampSimple(c.Time)
		rows = append(rows, tui.Row{Cells: []string{
			"CTR",
			fmt.Sprintf("%d", c.Index),
			fmt.Sprintf("%d", c.Value),
			quality,
			timeStr,
		}})
	}

	// Use UpdateDataIfChanged and SignalRedraw to avoid flicker
	if app.UpdateDataIfChanged(rows) {
		app.SignalRedraw()
	}
}

// qualityString converts quality flags to a string.
func qualityString(q types.QualityFlags) string {
	if q.IsGood() {
		return "ONLINE"
	}
	return "OFFLINE"
}
