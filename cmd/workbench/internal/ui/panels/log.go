package panels

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"dnp3/cmd/workbench/internal/session"
)

// LogPanel displays the communication log.
type LogPanel struct {
	container *fyne.Container
	list      *widget.List
	entries   []LogEntry
	mu        sync.RWMutex

	OnClear func()

	txBytes int
	rxBytes int
	txLabel *widget.Label
	rxLabel *widget.Label
}

// LogEntry represents a log entry.
type LogEntry struct {
	Timestamp time.Time
	Direction string // "TX", "RX", "INFO", "ERROR"
	Message   string
}

// NewLogPanel creates a new log panel.
func NewLogPanel() *LogPanel {
	p := &LogPanel{
		entries: make([]LogEntry, 0),
	}

	title := widget.NewLabel("COMMUNICATION LOG")
	title.TextStyle.Bold = true

	p.list = widget.NewList(
		func() int {
			p.mu.RLock()
			defer p.mu.RUnlock()
			return len(p.entries)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Entry")
		},
		func(i int, obj fyne.CanvasObject) {
			p.mu.RLock()
			defer p.mu.RUnlock()
			if i < len(p.entries) {
				obj.(*widget.Label).SetText(p.formatEntry(p.entries[i]))
			}
		},
	)

	p.list.Resize(fyne.NewSize(800, 150))

	scroll := container.NewScroll(p.list)
	scroll.SetMinSize(fyne.NewSize(800, 150))

	clearBtn := widget.NewButton("Clear Log", func() {
		p.Clear()
		if p.OnClear != nil {
			p.OnClear()
		}
	})

	p.txLabel = widget.NewLabel("TX: 0 bytes")
	p.rxLabel = widget.NewLabel("RX: 0 bytes")

	statsBox := container.NewHBox(p.txLabel, p.rxLabel, layout.NewSpacer(), clearBtn)

	p.container = container.NewVBox(
		title,
		scroll,
		statsBox,
	)

	return p
}

// Container returns the panel container.
func (p *LogPanel) Container() *fyne.Container {
	return p.container
}

// Append adds a log entry.
func (p *LogPanel) Append(t time.Time, direction, message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.entries = append(p.entries, LogEntry{
		Timestamp: t,
		Direction: direction,
		Message:   message,
	})

	// Limit log size
	if len(p.entries) > 1000 {
		p.entries = p.entries[len(p.entries)-1000:]
	}

	// Update stats
	switch direction {
	case "TX":
		p.txBytes += len(message)
		p.txLabel.SetText(fmt.Sprintf("TX: %d bytes", p.txBytes))
	case "RX":
		p.rxBytes += len(message)
		p.rxLabel.SetText(fmt.Sprintf("RX: %d bytes", p.rxBytes))
	}

	p.list.Refresh()
}

// UpdateWithResponse updates the log with a DNP3 response.
func (p *LogPanel) UpdateWithResponse(resp *session.Response) {
	if resp == nil {
		return
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	// Log entry count by type
	txCount := 0
	rxCount := 0
	for _, e := range p.entries {
		switch e.Direction {
		case "TX":
			txCount++
		case "RX":
			rxCount++
		}
	}

	p.txLabel.SetText(fmt.Sprintf("TX: %d messages", txCount))
	p.rxLabel.SetText(fmt.Sprintf("RX: %d messages", rxCount))
}

// Clear clears the log.
func (p *LogPanel) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.entries = make([]LogEntry, 0)
	p.txBytes = 0
	p.rxBytes = 0
	p.txLabel.SetText("TX: 0 bytes")
	p.rxLabel.SetText("RX: 0 bytes")
	p.list.Refresh()
}

func (p *LogPanel) formatEntry(entry LogEntry) string {
	ts := entry.Timestamp.Format("15:04:05.000")
	return fmt.Sprintf("[%s] %s → %s", ts, entry.Direction, entry.Message)
}
