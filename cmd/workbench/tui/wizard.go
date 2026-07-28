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

	for {
		// Print banner
		fmt.Println()
		fmt.Println(Cyan + Bold + "╔═══════════════════════════════════════════════════════════╗" + Reset)
		fmt.Println(Cyan + Bold + "║" + Reset + BrightWhite + Bold + "           DNP3 Engineering Workbench" + Reset + Cyan + Bold + "                    ║" + Reset)
		fmt.Println(Cyan + Bold + "╚═══════════════════════════════════════════════════════════╝" + Reset)
		fmt.Println()
		fmt.Println(BrightWhite + "Select operating mode:" + Reset)
		fmt.Println()

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
		fmt.Println(Dim + "Type 1 or 2 and press Enter, q to quit" + Reset)

		// Read input
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "q" || input == "quit" || input == "exit" {
			fmt.Println("Exiting...")
			os.Exit(0)
		}

		if input == "1" {
			w.choice = 0
			break
		} else if input == "2" {
			w.choice = 1
			break
		}
	}

	// If Master mode, ask for address and port
	if w.choice == 0 {
		fmt.Println()
		fmt.Println(BrightWhite + "Master Mode Configuration:" + Reset)

		// Get address
		fmt.Printf("Remote address ["+BrightCyan+"%s"+Reset+"] (Enter for default): ", w.address)
		addrReader := bufio.NewReader(os.Stdin)
		addrInput, _ := addrReader.ReadString('\n')
		addrInput = strings.TrimSpace(addrInput)
		if addrInput != "" {
			w.address = addrInput
		}

		// Get port
		fmt.Printf("Port ["+BrightCyan+"%d"+Reset+"] (Enter for default): ", w.port)
		portReader := bufio.NewReader(os.Stdin)
		portInput, _ := portReader.ReadString('\n')
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
