package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Wizard runs an interactive mode selection wizard.
type Wizard struct {
	screen  *Screen
	reader  *bufio.Reader
	choice  int
	address string
	port    int
}

// NewWizard creates a new wizard.
func NewWizard() *Wizard {
	return &Wizard{
		choice: 0,
		address: "127.0.0.1",
		port: 20000,
	}
}

// Run displays the wizard and returns the selected mode.
func (w *Wizard) Run() (Mode, string, int) {
	// Print banner
	fmt.Println()
	fmt.Println(Cyan + Bold + "╔═══════════════════════════════════════════════════════════╗" + Reset)
	fmt.Println(Cyan + Bold + "║" + Reset + BrightWhite + Bold + "           DNP3 Engineering Workbench" + Reset + Cyan + Bold + "                    ║" + Reset)
	fmt.Println(Cyan + Bold + "╚═══════════════════════════════════════════════════════════╝" + Reset)
	fmt.Println()
	fmt.Println(BrightWhite + "Select operating mode:" + Reset)
	fmt.Println()

	for {
		// Print mode options
		if w.choice == 0 {
			fmt.Println(BrightGreen + "  ► " + Reset + BrightWhite + "1. Master Mode" + Reset + Dim + " (connect to outstation)" + Reset)
		} else {
			fmt.Println("  " + BrightBlack + "1. Master Mode" + Reset + Dim + " (connect to outstation)" + Reset)
		}

		if w.choice == 1 {
			fmt.Println(BrightGreen + "  ► " + Reset + BrightWhite + "2. Outstation Mode" + Reset + Dim + " (run as server)" + Reset)
		} else {
			fmt.Println("  " + BrightBlack + "2. Outstation Mode" + Reset + Dim + " (run as server)" + Reset)
		}

		fmt.Println()
		fmt.Println(Dim + "Use 1/2 to select, Enter to confirm, q to quit" + Reset)

		// Create a simple reader
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "1", "master":
			w.choice = 0
		case "2", "outstation":
			w.choice = 1
		case "q", "quit", "exit":
			fmt.Println("Exiting...")
			os.Exit(0)
		case "":
			// Enter pressed - confirm selection
			break
		}

		// Move cursor up and redraw
		if input != "" && (input == "1" || input == "2" || input == "master" || input == "outstation") {
			// Just update the choice
		}
	}

	// If Master mode, ask for address and port
	if w.choice == 0 {
		fmt.Println()
		fmt.Println(BrightWhite + "Master Mode Configuration:" + Reset)

		// Get address
		fmt.Printf("Remote address ["+BrightCyan+"%s"+Reset+"] (press Enter for default): ", w.address)
		addrReader := bufio.NewReader(os.Stdin)
		addrInput, _ := addrReader.ReadString('\n')
		addrInput = strings.TrimSpace(addrInput)
		if addrInput != "" {
			w.address = addrInput
		}

		// Get port
		fmt.Printf("Port ["+BrightCyan+"%d"+Reset+"] (press Enter for default): ", w.port)
		portReader := bufio.NewReader(os.Stdin)
		portInput, _ := portReader.ReadString('\n')
		portInput = strings.TrimSpace(portInput)
		if portInput != "" {
			fmt.Sscanf(portInput, "%d", &w.port)
		}
	}

	// Clear screen for the main app
	fmt.Print(ClearScreen)
	fmt.Print(MoveTo(1, 1))

	if w.choice == 0 {
		return ModeMaster, w.address, w.port
	}
	return ModeOutstation, "", 0
}
