// DNP3 Engineering Workbench
// A desktop application for validating and debugging the native Go DNP3 library.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"

	"dnp3/cmd/workbench/internal/config"
	"dnp3/cmd/workbench/internal/logger"
	masterctrl "dnp3/cmd/workbench/internal/master"
	outstationctrl "dnp3/cmd/workbench/internal/outstation"
	"dnp3/cmd/workbench/internal/shared/types"
	"dnp3/cmd/workbench/internal/ui"
	"dnp3/cmd/workbench/internal/ui/dialogs"
)

// Window size constraints
const (
	MinWindowWidth  = 800
	MinWindowHeight = 600
	DefaultWidth    = 1200
	DefaultHeight   = 800
)

func main() {
	// Parse command-line flags
	modeStr := flag.String("mode", "select", "Operating mode: master, outstation, or select (default)")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "DNP3 Engineering Workbench\n\nUsage: %s [options]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// Validate mode
	mode := types.Mode(*modeStr)
	if err := mode.Validate(); err != nil {
		log.Fatalf("%v", err)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Failed to load config: %v, using defaults", err)
		cfg = config.Default()
	}

	// Create Fyne application
	a := app.New()

	// Apply saved theme
	if cfg.Appearance.Theme == "Dark" {
		a.Settings().SetTheme(theme.DarkTheme())
	} else {
		a.Settings().SetTheme(theme.LightTheme())
	}

	// Create logger
	log := logger.New()

	// Route to appropriate mode
	switch mode {
	case types.ModeMaster:
		runMaster(a, cfg, log)
	case types.ModeOutstation:
		runOutstation(a, cfg, log)
	case types.ModeSelect:
		runModeSelection(a, cfg, log)
	}
}

// runMaster runs the application in Master mode.
func runMaster(a fyne.App, cfg *config.Config, log *logger.Logger) {
	log.Info("Starting DNP3 Master mode")

	// Create Master controller
	ctrl := masterctrl.NewController(log)

	// Create Master window
	window := ui.NewMasterWindow(a, ctrl, cfg)

	// Set window properties
	window.Resize(fyne.NewSize(DefaultWidth, DefaultHeight))
	window.SetTitle("DNP3 Master - Connect to Outstation")
	window.CenterOnScreen()

	// Create menu with window controls
	window.SetMainMenu(createMasterMenu(a, window, ctrl, cfg))

	// Register keyboard shortcuts
	registerMasterShortcuts(window, ctrl)

	// Start controller
	if err := ctrl.Start(); err != nil {
		log.Error("Failed to start controller: %v", err)
	}

	// Show window
	window.Show()

	// Run event loop
	a.Run()

	// Cleanup
	cfg.Save()
	ctrl.Stop()
}

// runOutstation runs the application in Outstation mode.
func runOutstation(a fyne.App, cfg *config.Config, log *logger.Logger) {
	log.Info("Starting DNP3 Outstation mode")

	// Create Outstation controller
	ctrl := outstationctrl.NewController(log)

	// Create Outstation window
	window := ui.NewOutstationWindow(a, ctrl, cfg)

	// Set window properties
	window.Resize(fyne.NewSize(DefaultWidth, DefaultHeight))
	window.SetTitle("DNP3 Outstation - Simulate Data")
	window.CenterOnScreen()

	// Create menu with window controls
	window.SetMainMenu(createOutstationMenu(a, window, ctrl, cfg))

	// Register keyboard shortcuts
	registerOutstationShortcuts(window, ctrl)

	// Start controller
	if err := ctrl.Start(); err != nil {
		log.Error("Failed to start controller: %v", err)
	}

	// Show window
	window.Show()

	// Run event loop
	a.Run()

	// Cleanup
	cfg.Save()
	ctrl.Stop()
}

// runModeSelection shows the mode selection dialog.
func runModeSelection(a fyne.App, cfg *config.Config, log *logger.Logger) {
	log.Info("Showing mode selection dialog")

	// Create a temporary window for the dialog
	dialogs.ShowModeSelection(a, func(mode types.Mode) {
		// User selected a mode - restart with selected mode
		switch mode {
		case types.ModeMaster:
			runMaster(a, cfg, log)
		case types.ModeOutstation:
			runOutstation(a, cfg, log)
		default:
			log.Error("Invalid mode selected")
			a.Quit()
		}
	})
}

// createMasterMenu creates the menu for Master window with File menu controls.
func createMasterMenu(a fyne.App, window *ui.MasterWindow, ctrl *masterctrl.Controller, cfg *config.Config) *fyne.MainMenu {
	// File Menu with window controls
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Minimize", func() {
			window.Window().Minimize()
		}),
		fyne.NewMenuItem("Maximize", func() {
			window.Maximize()
		}),
		fyne.NewMenuItem("Restore", func() {
			window.Restore()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Close", func() {
			a.Quit()
		}),
	)

	// Edit Menu
	editMenu := fyne.NewMenu("Edit",
		fyne.NewMenuItem("Find in Log", func() {
			window.ShowLogSearch()
		}),
	)

	// Session Menu
	sessionMenu := fyne.NewMenu("Session",
		fyne.NewMenuItem("Connect", func() {
			ctrl.Connect(ctrl.State().Address, ctrl.State().Port)
		}),
		fyne.NewMenuItem("Disconnect", func() {
			ctrl.Disconnect()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Read Class 0", func() {
			ctrl.ReadClass(0)
		}),
		fyne.NewMenuItem("Read Class 1", func() {
			ctrl.ReadClass(1)
		}),
		fyne.NewMenuItem("Read Class 2", func() {
			ctrl.ReadClass(2)
		}),
		fyne.NewMenuItem("Read Class 3", func() {
			ctrl.ReadClass(3)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Clear Log", func() {
			ctrl.Logger().Clear()
			window.ClearLog()
		}),
	)

	// Help Menu
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Keyboard Shortcuts", func() {
			showShortcutsDialog(window.Window())
		}),
		fyne.NewMenuItem("About DNP3 Workbench", func() {
			dialogs.ShowAbout(window.Window())
		}),
	)

	return fyne.NewMainMenu(fileMenu, editMenu, sessionMenu, helpMenu)
}

// createOutstationMenu creates the menu for Outstation window with File menu controls.
func createOutstationMenu(a fyne.App, window *ui.OutstationWindow, ctrl *outstationctrl.Controller, cfg *config.Config) *fyne.MainMenu {
	// File Menu with window controls
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Minimize", func() {
			window.Window().Minimize()
		}),
		fyne.NewMenuItem("Maximize", func() {
			window.Maximize()
		}),
		fyne.NewMenuItem("Restore", func() {
			window.Restore()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Close", func() {
			a.Quit()
		}),
	)

	// Edit Menu
	editMenu := fyne.NewMenu("Edit",
		fyne.NewMenuItem("Find in Log", func() {
			window.ShowLogSearch()
		}),
	)

	// Session Menu
	sessionMenu := fyne.NewMenu("Session",
		fyne.NewMenuItem("Start Server", func() {
			ctrl.StartServer(ctrl.State().ListenAddress, ctrl.State().ListenPort)
		}),
		fyne.NewMenuItem("Stop Server", func() {
			ctrl.Stop()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Clear Log", func() {
			ctrl.Logger().Clear()
			window.ClearLog()
		}),
	)

	// Help Menu
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Keyboard Shortcuts", func() {
			showShortcutsDialog(window.Window())
		}),
		fyne.NewMenuItem("About DNP3 Workbench", func() {
			dialogs.ShowAbout(window.Window())
		}),
	)

	return fyne.NewMainMenu(fileMenu, editMenu, sessionMenu, helpMenu)
}

// registerMasterShortcuts sets up keyboard shortcuts for Master window.
func registerMasterShortcuts(window *ui.MasterWindow, ctrl *masterctrl.Controller) {
	// Shortcuts are handled via menu accelerators in Fyne.
	_ = window
	_ = ctrl
}

// registerOutstationShortcuts sets up keyboard shortcuts for Outstation window.
func registerOutstationShortcuts(window *ui.OutstationWindow, ctrl *outstationctrl.Controller) {
	// Shortcuts are handled via menu accelerators in Fyne.
	_ = window
	_ = ctrl
}

// showShortcutsDialog displays keyboard shortcuts help.
func showShortcutsDialog(parent fyne.Window) {
	content := dialogs.NewShortcutsDialog(parent)
	content.Show()
}

func init() {
	log.SetFlags(log.Lshortfile | log.Ltime)
}
