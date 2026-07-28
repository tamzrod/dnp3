package tui

import (
	"bufio"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/term"
)

// Key represents keyboard keys.
type Key int

const (
	KeyUp Key = iota + 256
	KeyDown
	KeyLeft
	KeyRight
	KeyEnter
	KeyEscape
	KeyBackspace
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
	// Check if stdin is a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil // Skip raw mode for non-TTY (e.g., piped input)
	}
	
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

		reader := bufio.NewReader(os.Stdin)

		for {
			select {
			case <-i.done:
				return
			case <-sigChan:
				events <- Event{Type: EventResize}
			default:
				// Set a read timeout
				os.Stdin.SetReadDeadline(time.Now().Add(50 * time.Millisecond))

				b, err := reader.ReadByte()
				if err != nil {
					continue
				}

				// Handle ANSI escape sequences
				if b == 27 {
					// Peek at next byte
					os.Stdin.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
					next1, err1 := reader.ReadByte()
					if err1 == nil && next1 == '[' {
						os.Stdin.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
						next2, err2 := reader.ReadByte()
						if err2 == nil {
							switch next2 {
							case 'A':
								events <- Event{Type: EventKey, Key: KeyUp}
								continue
							case 'B':
								events <- Event{Type: EventKey, Key: KeyDown}
								continue
							case 'C':
								events <- Event{Type: EventKey, Key: KeyRight}
								continue
							case 'D':
								events <- Event{Type: EventKey, Key: KeyLeft}
								continue
							}
							// Put back the bytes we don't recognize
							reader.UnreadByte()
						}
						reader.UnreadByte()
					}
					events <- Event{Type: EventKey, Key: KeyEscape}
					continue
				}

				// Handle regular keys
				switch b {
				case 13, 10: // Enter
					events <- Event{Type: EventKey, Key: KeyEnter}
				case 127, 8: // Backspace
					events <- Event{Type: EventKey, Key: KeyBackspace}
				default:
					events <- Event{Type: EventKey, Rune: rune(b)}
				}
			}
		}
	}()

	return events
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
		return "Up"
	case KeyDown:
		return "Down"
	case KeyLeft:
		return "Left"
	case KeyRight:
		return "Right"
	case KeyEnter:
		return "Enter"
	case KeyEscape:
		return "Esc"
	case KeyBackspace:
		return "Backspace"
	default:
		return string(rune(k))
	}
}
