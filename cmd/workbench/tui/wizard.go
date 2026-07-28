package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Wizard runs an interactive mode selection wizard.
type Wizard struct {
	choice  int
	address string
	port    int
}

// NewWizard creates a new wizard.
func NewWizard() *Wizard {
	return &Wizard{
		choice:  0,
		address: "127.0.0.1",
		port:    20000,
	}
}

// Run displays the wizard and returns the selected mode.
func (w *Wizard) Run() (Mode, string, int) {
	fmt.Print(ClearScreen)
	fmt.Println()
	fmt.Println(Cyan + Bold + "╔═══════════════════════════════════════════════════════════╗" + Reset)
	fmt.Println(Cyan + Bold + "║" + Reset + BrightWhite + Bold + "           DNP3 Engineering Workbench" + Reset + Cyan + Bold + "                    ║" + Reset)
	fmt.Println(Cyan + Bold + "╚═══════════════════════════════════════════════════════════╝" + Reset)
	fmt.Println()
	fmt.Println(BrightWhite + "Select operating mode:" + Reset)
	fmt.Println()

	// Create a single reader for all input
	reader := bufio.NewReader(os.Stdin)

	for {
		// Print mode options
		if w.choice == 0 {
			fmt.Println(BrightGreen + "  >>> [1] Master Mode" + Reset + Dim + " - connect to outstation" + Reset)
		} else {
			fmt.Println("      [1] Master Mode" + Dim + " - connect to outstation" + Reset)
		}

		if w.choice == 1 {
			fmt.Println(BrightGreen + "  >>> [2] Outstation Mode" + Reset + Dim + " - run as server" + Reset)
		} else {
			fmt.Println("      [2] Outstation Mode" + Dim + " - run as server" + Reset)
		}

		fmt.Println()
		fmt.Print(Dim + "Type 1 or 2, press Enter to confirm, or q to quit: " + Reset)

		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "q" || input == "quit" || input == "exit" {
			fmt.Println("\nExiting...")
			os.Exit(0)
		}

		if input == "1" || input == "master" {
			w.choice = 0
			break
		} else if input == "2" || input == "outstation" {
			w.choice = 1
			break
		}

		// Invalid input, just show error and loop
		fmt.Println(Red + "Invalid selection. Please type 1 or 2." + Reset)
	}

	// If Master mode, ask for address and port
	if w.choice == 0 {
		fmt.Println()
		fmt.Println(BrightWhite + "Master Mode Configuration:" + Reset)

		// Get address
		fmt.Printf("Remote address ["+BrightCyan+"%s"+Reset+"] (Enter for default): ", w.address)
		addrInput, _ := reader.ReadString('\n')
		addrInput = strings.TrimSpace(addrInput)
		if addrInput != "" {
			w.address = addrInput
		}

		// Get port
		fmt.Printf("Port ["+BrightCyan+"%d"+Reset+"] (Enter for default): ", w.port)
		portInput, _ := reader.ReadString('\n')
		portInput = strings.TrimSpace(portInput)
		if portInput != "" {
			fmt.Sscanf(portInput, "%d", &w.port)
		}
	}

	// Clear screen for the main app
	fmt.Print(ClearScreen)

	if w.choice == 0 {
		return ModeMaster, w.address, w.port
	}
	return ModeOutstation, "", 0
}
