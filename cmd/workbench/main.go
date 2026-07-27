// DNP3 Engineering Workbench
// A desktop application for validating and debugging the native Go DNP3 library.
package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"

	"dnp3/cmd/workbench/internal/controller"
	"dnp3/cmd/workbench/internal/ui"
)

func main() {
	// Create Fyne application
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())

	// Create controller
	ctrl := controller.New(nil)

	// Create main window with controller
	window := ui.NewMainWindow(a, ctrl)

	// Set window properties
	window.Resize(fyne.NewSize(1200, 800))
	window.SetTitle("DNP3 Engineering Workbench")
	window.CenterOnScreen()

	// Start controller
	if err := ctrl.Start(); err != nil {
		log.Printf("Failed to start controller: %v", err)
	}

	// Show window
	window.Show()

	// Run the Fyne event loop - this blocks until the app terminates
	a.Run()

	// Cleanup (executed after window is closed)
	ctrl.Stop()
}

// init registers the application
func init() {
	log.SetFlags(log.Lshortfile | log.Ltime)
}
