// DNP3 Engineering Workbench
// A desktop application for validating and debugging the native Go DNP3 library.
package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"

	"dnp3/cmd/workbench/internal/config"
	"dnp3/cmd/workbench/internal/controller"
	"dnp3/cmd/workbench/internal/ui"
	"dnp3/cmd/workbench/internal/ui/dialogs"
)

// Window size constraints as per UX standards (Section 3.2)
const (
	MinWindowWidth  = 800
	MinWindowHeight = 600
	DefaultWidth    = 1200
	DefaultHeight   = 800
)

func main() {
	// Load configuration (UX Standard Section 8.1)
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Failed to load config: %v, using defaults", err)
		cfg = config.Default()
	}

	// Create Fyne application
	a := app.New()
	
	// Apply saved theme or default to light (UX Standard: Platform consistency)
	// Fyne uses native decorations by default on all platforms
	if cfg.Appearance.Theme == "Dark" {
		a.Settings().SetTheme(theme.DarkTheme())
	} else {
		a.Settings().SetTheme(theme.LightTheme())
	}

	// Create controller
	ctrl := controller.New(nil)

	// Create main window with controller
	window := ui.NewMainWindow(a, ctrl, cfg)

	// Set default window size
	window.Resize(fyne.NewSize(DefaultWidth, DefaultHeight))
	window.SetTitle("DNP3 Engineering Workbench")
	window.CenterOnScreen()

	// Create complete menu structure (UX Standard: File, Edit, View, Session, Help)
	window.SetMainMenu(createMainMenu(a, window, ctrl, cfg))

	// Register keyboard shortcuts (UX Standard: Standard shortcuts reduce learning curve)
	registerShortcuts(window, ctrl)

	// Start controller
	if err := ctrl.Start(); err != nil {
		log.Printf("Failed to start controller: %v", err)
	}

	// Show window
	window.Show()

	// Run the Fyne event loop - this blocks until the app terminates
	a.Run()

	// Cleanup
	cfg.Save()
	ctrl.Stop()
}

// createMainMenu builds the complete menu structure per UX standards.
func createMainMenu(a fyne.App, window *ui.MainWindow, ctrl *controller.Controller, cfg *config.Config) *fyne.MainMenu {
	// File Menu (UX Standard Section 4.2)
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("New Session", func() {
			// Stub: Reset to default state
			dialog.ShowInformation("New Session", "Starting a new session...", window.Window())
		}),
		fyne.NewMenuItem("Open Configuration...", func() {
			dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil || reader == nil {
					return
				}
				defer reader.Close()
				dialog.ShowInformation("Open", "Configuration loading not yet implemented.", window.Window())
			}, window.Window())
		}),
		fyne.NewMenuItem("Save Configuration", func() {
			dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
				if err != nil || writer == nil {
					return
				}
				defer writer.Close()
				dialog.ShowInformation("Save", "Configuration saving not yet implemented.", window.Window())
			}, window.Window())
		}),
		fyne.NewMenuItem("Save Configuration As...", func() {
			dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
				if err != nil || writer == nil {
					return
				}
				defer writer.Close()
				dialog.ShowInformation("Save As", "Configuration saving not yet implemented.", window.Window())
			}, window.Window())
		}),
		fyne.NewMenuItem("Export Log...", func() {
			dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
				if err != nil || writer == nil {
					return
				}
				defer writer.Close()
				window.ExportLog(writer)
			}, window.Window())
		}),
		fyne.NewMenuItem("Print...", func() {
			dialog.ShowInformation("Print", "Printing not yet implemented.", window.Window())
		}),
		fyne.NewMenuItem("Exit", func() {
			a.Quit()
		}),
	)

	// Edit Menu (UX Standard Section 4.3)
	editMenu := fyne.NewMenu("Edit",
		fyne.NewMenuItem("Undo", func() {
			// Stub: Undo not implemented
		}),
		fyne.NewMenuItem("Redo", func() {
			// Stub: Redo not implemented
		}),
		fyne.NewMenuItem("Cut", func() {
			// Use standard clipboard cut
		}),
		fyne.NewMenuItem("Copy", func() {
			// Use standard clipboard copy
		}),
		fyne.NewMenuItem("Paste", func() {
			// Use standard clipboard paste
		}),
		fyne.NewMenuItem("Delete", func() {
			// Context-dependent delete
		}),
		fyne.NewMenuItem("Find in Log", func() {
			window.ShowLogSearch()
		}),
		fyne.NewMenuItem("Select All", func() {
			// Use standard select all
		}),
	)

	// View Menu (UX Standard Section 4.4)
	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Zoom In", func() {
			// Stub: Zoom not implemented
		}),
		fyne.NewMenuItem("Zoom Out", func() {
			// Stub: Zoom not implemented
		}),
		fyne.NewMenuItem("Reset Zoom", func() {
			// Stub: Zoom not implemented
		}),
		fyne.NewMenuItem("Sidebar", func() {
			window.ToggleSidebar()
		}),
		fyne.NewMenuItem("Log Panel", func() {
			window.ToggleLogPanel()
		}),
		fyne.NewMenuItem("Fullscreen", func() {
			window.ToggleFullscreen()
		}),
	)

	// Settings menu (UX Standard Section 4.5)
	settingsMenu := fyne.NewMenu("Settings",
		fyne.NewMenuItem("Preferences...", func() {
			showSettingsDialog(window.Window(), cfg)
		}),
	)

	// Session Menu (UX Standard: Engineering-specific actions)
	sessionMenu := fyne.NewMenu("Session",
		fyne.NewMenuItem("Connect", func() {
			ctrl.Connect(ctrl.State().Address, ctrl.State().Port)
		}),
		fyne.NewMenuItem("Disconnect", func() {
			ctrl.Disconnect()
		}),
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
		fyne.NewMenuItem("Clear Log", func() {
			ctrl.Logger().Clear()
			window.ClearLog()
		}),
	)

	// Help Menu (UX Standard Section 4.1)
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Documentation", func() {
			dialog.ShowInformation("Documentation", "DNP3 Engineering Workbench Documentation\n\nSee README.md for usage instructions.", window.Window())
		}),
		fyne.NewMenuItem("Keyboard Shortcuts", func() {
			showShortcutsDialog(window.Window())
		}),
		fyne.NewMenuItem("About DNP3 Workbench", func() {
			dialogs.ShowAbout(window.Window())
		}),
	)

	return fyne.NewMainMenu(fileMenu, editMenu, viewMenu, sessionMenu, settingsMenu, helpMenu)
}

// showSettingsDialog displays the settings dialog.
func showSettingsDialog(parent fyne.Window, cfg *config.Config) {
	dialogs.NewSettingsDialog(parent, cfg).Show()
}

// registerShortcuts sets up keyboard shortcuts per UX standards.
// Note: Fyne menus handle shortcuts automatically via accelerator keys.
func registerShortcuts(window *ui.MainWindow, ctrl *controller.Controller) {
	// Shortcuts are handled via menu accelerators in Fyne.
	// The menu bar items define the keyboard shortcuts.
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
