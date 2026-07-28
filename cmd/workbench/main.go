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
	modeStr := flag.String("mode", "", "Operating mode: master, outstation, or leave empty for wizard")
	address := flag.String("address", "127.0.0.1", "Remote address (Master mode)")
	port := flag.Int("port", 20000, "Port number")
	flag.Parse()

	var mode tui.Mode
	var addr string
	var p int

	// Check if mode was specified on command line
	if *modeStr != "" {
		// Use command-line arguments
		mode = tui.Mode(strings.ToLower(*modeStr))
		if mode != tui.ModeMaster && mode != tui.ModeOutstation {
			fmt.Fprintf(os.Stderr, "Invalid mode: %s (use 'master' or 'outstation')\n", *modeStr)
			os.Exit(1)
		}
		addr = *address
		p = *port
	} else {
		// Run wizard to select mode
		wizard := tui.NewWizard()
		mode, addr, p = wizard.Run()
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
		setupMaster(app, addr, p)
	} else {
		setupOutstation(app, addr, p)
	}

	// Handle quit
	app.OnQuit = func() {
		app.LogInfo("Shutting down...")
	}

	// Run the application
	if err := app.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

// setupMaster sets up Master mode callbacks.
func setupMaster(app *tui.App, address string, port int) {
	app.LogInfo(fmt.Sprintf("Connecting to %s:%d", address, port))

	// Create logger
	log := logger.New()

	// Create controller
	ctrl := masterctrl.NewController(log)

	// Connect callback
	app.OnConnect = func() {
		app.LogInfo(fmt.Sprintf("Connecting to %s:%d...", address, port))
		if err := ctrl.Connect(address, port); err != nil {
			app.LogError(fmt.Sprintf("Connection failed: %v", err))
		} else {
			app.LogInfo("Connected!")
			app.SetConnection("Connected", fmt.Sprintf("%s:%d", address, port))
		}
	}

	// Disconnect callback
	app.OnDisconnect = func() {
		app.LogInfo("Disconnecting...")
		if err := ctrl.Disconnect(); err != nil {
			app.LogError(fmt.Sprintf("Disconnect failed: %v", err))
		} else {
			app.LogInfo("Disconnected")
			app.SetConnection("Disconnected", "")
		}
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

	// Start polling for state updates
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				state := ctrl.State()
				updateData(app, state)
			case <-updateCh:
				return
			}
		}
	}()
}

// setupOutstation sets up Outstation mode callbacks.
func setupOutstation(app *tui.App, address string, port int) {
	app.LogInfo(fmt.Sprintf("Starting server on %s:%d", address, port))

	// Create logger
	log := logger.New()

	// Create controller
	ctrl := outstationctrl.NewController(log)

	// Start server callback (not wired to keyboard yet, just for reference)
	_ = func() {
		app.LogInfo(fmt.Sprintf("Starting server on %s:%d...", address, port))
		if err := ctrl.StartServer(address, port); err != nil {
			app.LogError(fmt.Sprintf("Server start failed: %v", err))
		} else {
			app.LogInfo("Server started!")
			app.SetConnection("Listening", fmt.Sprintf("%s:%d", address, port))
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

		for _, bi := range resp.BinaryInputs {
			quality := qualityString(bi.Quality)
			rows = append(rows, tui.Row{Cells: []string{
				"BI",
				fmt.Sprintf("%d", bi.Index),
				fmt.Sprintf("%v", bi.Value),
				quality,
				time.Now().Format("15:04:05"),
			}})
		}

		for _, ai := range resp.AnalogInputs {
			quality := qualityString(ai.Quality)
			rows = append(rows, tui.Row{Cells: []string{
				"AI",
				fmt.Sprintf("%d", ai.Index),
				fmt.Sprintf("%.2f", ai.Value),
				quality,
				time.Now().Format("15:04:05"),
			}})
		}

		for _, c := range resp.Counters {
			quality := qualityString(c.Quality)
			rows = append(rows, tui.Row{Cells: []string{
				"CTR",
				fmt.Sprintf("%d", c.Index),
				fmt.Sprintf("%d", c.Value),
				quality,
				time.Now().Format("15:04:05"),
			}})
		}
	}

	app.UpdateData(rows)
}

// updateOutstationData updates the TUI with controller state (Outstation mode).
func updateOutstationData(app *tui.App, ctrl *outstationctrl.Controller) {
	// Build data rows from simulator
	var rows []tui.Row

	binary := ctrl.GetBinaryInputs()
	for _, bi := range binary {
		quality := qualityString(bi.Quality)
		rows = append(rows, tui.Row{Cells: []string{
			"BI",
			fmt.Sprintf("%d", bi.Index),
			fmt.Sprintf("%v", bi.Value),
			quality,
			time.Now().Format("15:04:05"),
		}})
	}

	analog := ctrl.GetAnalogInputs()
	for _, ai := range analog {
		quality := qualityString(ai.Quality)
		rows = append(rows, tui.Row{Cells: []string{
			"AI",
			fmt.Sprintf("%d", ai.Index),
			fmt.Sprintf("%.2f", ai.Value),
			quality,
			time.Now().Format("15:04:05"),
		}})
	}

	counters := ctrl.GetCounters()
	for _, c := range counters {
		quality := qualityString(c.Quality)
		rows = append(rows, tui.Row{Cells: []string{
			"CTR",
			fmt.Sprintf("%d", c.Index),
			fmt.Sprintf("%d", c.Value),
			quality,
			time.Now().Format("15:04:05"),
		}})
	}

	app.UpdateData(rows)
}

// qualityString converts quality flags to a string.
func qualityString(q types.QualityFlags) string {
	if q.IsGood() {
		return "ONLINE"
	}
	return "OFFLINE"
}
