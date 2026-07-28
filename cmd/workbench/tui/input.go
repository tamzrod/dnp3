package tui

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// Key represents keyboard keys.
type Key int

const (
	// Special keys (values above ASCII range)
	KeyUp Key = iota + 256
	KeyDown
	KeyLeft
	KeyRight
	KeyEnter
	KeyEscape
	KeyBackspace
	KeyTab
	KeySpace
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
	KeyDelete
	KeyInsert

	// Ctrl keys
	KeyCtrlA
	KeyCtrlB
	KeyCtrlC
	KeyCtrlD
	KeyCtrlE
	KeyCtrlF
	KeyCtrlG
	KeyCtrlH
	KeyCtrlI
	KeyCtrlJ
	KeyCtrlK
	KeyCtrlL
	KeyCtrlM
	KeyCtrlN
	KeyCtrlO
	KeyCtrlP
	KeyCtrlQ
	KeyCtrlR
	KeyCtrlS
	KeyCtrlT
	KeyCtrlU
	KeyCtrlV
	KeyCtrlW
	KeyCtrlX
	KeyCtrlY
	KeyCtrlZ

	// Rune keys (rune values)
	KeyRune Key = 256 + 26 + 26 // Start of rune range
)

// Event represents an input event.
type Event struct {
	Type EventType
	Key  Key
	Rune rune
}

// EventType represents the type of input event.
type EventType int

const (
	EventKey EventType = iota
	EventResize
	EventQuit
)

// Input handles keyboard input.
type Input struct {
	oldState *term.State
	done     chan struct{}
}

// NewInput creates a new input handler.
func NewInput() *Input {
	return &Input{
		done: make(chan struct{}),
	}
}

// EnableRawMode enables raw mode for the terminal.
func (i *Input) EnableRawMode() error {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	i.oldState = oldState
	return nil
}

// DisableRawMode restores the terminal to its previous state.
func (i *Input) DisableRawMode() {
	if i.oldState != nil {
		term.Restore(int(os.Stdin.Fd()), i.oldState)
		i.oldState = nil
	}
}

// Events returns a channel of input events.
func (i *Input) Events() <-chan Event {
	events := make(chan Event)

	go func() {
		defer close(events)

		// Set up signal handler for resize
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGWINCH)

		for {
			select {
			case <-i.done:
				return
			case <-sigChan:
				events <- Event{Type: EventResize}
			default:
				// Read a single character with non-blocking read
				b := make([]byte, 1)
				fd := int(os.Stdin.Fd())
				oldState, _ := term.GetState(fd)
				
				// Switch to non-blocking
				term.MakeRaw(fd)
				n, err := os.Stdin.Read(b)
				term.Restore(fd, oldState)
				
				if err != nil || n == 0 {
					continue
				}

				// Handle ANSI escape sequences
				if b[0] == 27 { // ESC
					// Read remaining bytes with timeout
					remaining := make([]byte, 2)
					term.MakeRaw(fd)
					for j := 0; j < 2; j++ {
						tmp := make([]byte, 1)
						os.Stdin.Read(tmp)
						if tmp[0] >= 64 && tmp[0] <= 126 {
							remaining[j] = tmp[0]
						} else {
							break
						}
					}
					term.Restore(fd, oldState)
					
					if len(remaining) >= 2 && remaining[0] == '[' {
						key := i.parseRawSequence(remaining[1:])
						if key != 0 {
							events <- Event{Type: EventKey, Key: key}
							continue
						}
					}
					events <- Event{Type: EventKey, Key: KeyEscape}
					continue
				}

				// Handle regular keys
				key := i.parseKey(b[0])
				if key != 0 {
					events <- Event{Type: EventKey, Key: key}
				} else {
					events <- Event{Type: EventKey, Rune: rune(b[0])}
				}
			}
		}
	}()

	return events
}

// parseANSISequence parses ANSI escape sequences.
func (i *Input) parseRawSequence(b []byte) Key {
	if len(b) == 0 {
		return 0
	}

	switch b[0] {
	case 'A':
		return KeyUp
	case 'B':
		return KeyDown
	case 'C':
		return KeyRight
	case 'D':
		return KeyLeft
	case 'H':
		return KeyHome
	case 'F':
		return KeyEnd
	}

	return 0
}

// parseKey converts a byte to a Key.
func (i *Input) parseKey(b byte) Key {
	switch b {
	case 13:
		return KeyEnter
	case 127:
		return KeyBackspace
	case 9:
		return KeyTab
	case 32:
		return KeySpace
	case 1:
		return KeyCtrlA
	case 2:
		return KeyCtrlB
	case 3:
		return KeyCtrlC
	case 4:
		return KeyCtrlD
	case 5:
		return KeyCtrlE
	case 6:
		return KeyCtrlF
	case 7:
		return KeyCtrlG
	case 8:
		return KeyCtrlH
	case 11:
		return KeyCtrlK
	case 12:
		return KeyCtrlL
	case 14:
		return KeyCtrlN
	case 16:
		return KeyCtrlP
	case 18:
		return KeyCtrlR
	case 19:
		return KeyCtrlS
	case 20:
		return KeyCtrlT
	case 21:
		return KeyCtrlU
	case 22:
		return KeyCtrlV
	case 23:
		return KeyCtrlW
	case 24:
		return KeyCtrlX
	case 25:
		return KeyCtrlY
	case 26:
		return KeyCtrlZ
	case 17:
		return KeyCtrlQ
	}
	return 0
}

// Stop stops the input handler.
func (i *Input) Stop() {
	close(i.done)
	i.DisableRawMode()
}

// KeyName returns a human-readable name for a key.
func KeyName(k Key) string {
	switch k {
	case KeyUp:
		return "↑"
	case KeyDown:
		return "↓"
	case KeyLeft:
		return "←"
	case KeyRight:
		return "→"
	case KeyEnter:
		return "Enter"
	case KeyEscape:
		return "Esc"
	case KeyBackspace:
		return "⌫"
	case KeyTab:
		return "Tab"
	case KeySpace:
		return "Space"
	case KeyHome:
		return "Home"
	case KeyEnd:
		return "End"
	case KeyPageUp:
		return "PgUp"
	case KeyPageDown:
		return "PgDn"
	case KeyDelete:
		return "Del"
	case KeyInsert:
		return "Ins"
	case KeyCtrlA:
		return "Ctrl+A"
	case KeyCtrlB:
		return "Ctrl+B"
	case KeyCtrlC:
		return "Ctrl+C"
	case KeyCtrlD:
		return "Ctrl+D"
	case KeyCtrlE:
		return "Ctrl+E"
	case KeyCtrlF:
		return "Ctrl+F"
	case KeyCtrlG:
		return "Ctrl+G"
	case KeyCtrlH:
		return "Ctrl+H"
	case KeyCtrlK:
		return "Ctrl+K"
	case KeyCtrlL:
		return "Ctrl+L"
	case KeyCtrlN:
		return "Ctrl+N"
	case KeyCtrlP:
		return "Ctrl+P"
	case KeyCtrlQ:
		return "Ctrl+Q"
	case KeyCtrlR:
		return "Ctrl+R"
	case KeyCtrlS:
		return "Ctrl+S"
	case KeyCtrlT:
		return "Ctrl+T"
	case KeyCtrlU:
		return "Ctrl+U"
	case KeyCtrlV:
		return "Ctrl+V"
	case KeyCtrlW:
		return "Ctrl+W"
	case KeyCtrlX:
		return "Ctrl+X"
	case KeyCtrlY:
		return "Ctrl+Y"
	case KeyCtrlZ:
		return "Ctrl+Z"
	default:
		return string(rune(k))
	}
}
