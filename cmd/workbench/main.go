// DNP3 Engineering Workbench
// A desktop application for validating and debugging the native Go DNP3 library.
package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"dnp3/cmd/workbench/internal/ui"
	"dnp3/cmd/workbench/internal/session"
)

func main() {
	// Create Fyne application
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())

	// Create session manager
	manager := session.NewManager()

	// Create main window
	window := ui.NewMainWindow(a, manager)

	// Set window properties
	window.Resize(fyne.NewSize(1200, 800))
	window.SetTitle("DNP3 Engineering Workbench")
	window.CenterOnScreen()

	// Show and run
	window.ShowAndRun()
}

// init registers the application
func init() {
	log.SetFlags(log.Lshortfile | log.Ltime)
}
