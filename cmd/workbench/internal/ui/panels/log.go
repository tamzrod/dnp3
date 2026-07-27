package panels

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"dnp3/cmd/workbench/internal/session"
)

// LogLevel represents the severity level of a log entry.
type LogLevel int

const (
	LogLevelAll LogLevel = iota
	LogLevelTX
	LogLevelRX
	LogLevelInfo
	LogLevelError
)

func (l LogLevel) String() string {
	switch l {
	case LogLevelAll:
		return "All"
	case LogLevelTX:
		return "TX Only"
	case LogLevelRX:
		return "RX Only"
	case LogLevelInfo:
		return "Info Only"
	case LogLevelError:
		return "Errors Only"
	default:
		return "All"
	}
}

// LogPanel displays the communication log.
type LogPanel struct {
	container *fyne.Container
	list      *widget.List
	entries   []LogEntry
	filtered  []int // indices of filtered entries for search
	levelFilter LogLevel
	searchText  string
	mu        sync.RWMutex

	OnClear func()

	txBytes int
	rxBytes int
	txLabel *widget.Label
	rxLabel *widget.Label
	
	autoScroll bool
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
		entries:    make([]LogEntry, 0),
		filtered:   nil, // nil means show all
		levelFilter: LogLevelAll,
		autoScroll: true,
	}

	title := widget.NewLabel("COMMUNICATION LOG")
	title.TextStyle.Bold = true

	p.list = widget.NewList(
		func() int {
			p.mu.RLock()
			defer p.mu.RUnlock()
			if p.filtered != nil {
				return len(p.filtered)
			}
			return len(p.entries)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Entry")
		},
		func(i int, obj fyne.CanvasObject) {
			p.mu.RLock()
			defer p.mu.RUnlock()
			var entry LogEntry
			if p.filtered != nil {
				if i < len(p.filtered) {
					entry = p.entries[p.filtered[i]]
				}
			} else {
				if i < len(p.entries) {
					entry = p.entries[i]
				}
			}
			obj.(*widget.Label).SetText(p.formatEntry(entry))
		},
	)

	p.list.Resize(fyne.NewSize(800, 150))

	scroll := container.NewScroll(p.list)
	scroll.SetMinSize(fyne.NewSize(800, 150))

	// Filter controls (UX Standard Section 7.3)
	filterLabel := widget.NewLabel("Filter:")
	filterSelect := widget.NewSelect([]string{"All", "TX Only", "RX Only", "Info Only", "Errors Only"}, func(selected string) {
		p.SetLevelFilter(selected)
	})
	filterSelect.Selected = "All"

	clearBtn := widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), func() {
		p.Clear()
		if p.OnClear != nil {
			p.OnClear()
		}
	})

	p.txLabel = widget.NewLabel("TX: 0 bytes")
	p.rxLabel = widget.NewLabel("RX: 0 bytes")

	statsBox := container.NewHBox(p.txLabel, p.rxLabel, layout.NewSpacer(), filterLabel, filterSelect, clearBtn)

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

	p.applyFilters()
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
	p.filtered = nil // Clear filter too
	p.searchText = ""
	p.txBytes = 0
	p.rxBytes = 0
	p.txLabel.SetText("TX: 0 bytes")
	p.rxLabel.SetText("RX: 0 bytes")
	p.list.Refresh()
}

// SetLevelFilter sets the log level filter (UX Standard Section 7.3).
func (p *LogPanel) SetLevelFilter(level string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch level {
	case "All":
		p.levelFilter = LogLevelAll
	case "TX Only":
		p.levelFilter = LogLevelTX
	case "RX Only":
		p.levelFilter = LogLevelRX
	case "Info Only":
		p.levelFilter = LogLevelInfo
	case "Errors Only":
		p.levelFilter = LogLevelError
	default:
		p.levelFilter = LogLevelAll
	}

	p.applyFilters()
	p.list.Refresh()
}

// Search searches the log for the given text (UX Standard Section 7.3).
func (p *LogPanel) Search(text string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.searchText = text
	p.applyFilters()
	p.list.Refresh()
}

// applyFilters applies both level filter and search filter.
func (p *LogPanel) applyFilters() {
	if p.levelFilter == LogLevelAll && p.searchText == "" {
		p.filtered = nil
		return
	}

	searchLower := strings.ToLower(p.searchText)
	p.filtered = make([]int, 0)

	for i, entry := range p.entries {
		// Apply level filter
		if !p.matchesLevel(entry.Direction) {
			continue
		}

		// Apply search filter
		if searchLower != "" {
			if !strings.Contains(strings.ToLower(entry.Message), searchLower) &&
				!strings.Contains(strings.ToLower(entry.Direction), searchLower) {
				continue
			}
		}

		p.filtered = append(p.filtered, i)
	}
}

// matchesLevel checks if an entry matches the current level filter.
func (p *LogPanel) matchesLevel(direction string) bool {
	switch p.levelFilter {
	case LogLevelAll:
		return true
	case LogLevelTX:
		return direction == "TX"
	case LogLevelRX:
		return direction == "RX"
	case LogLevelInfo:
		return direction == "INFO"
	case LogLevelError:
		return direction == "ERROR"
	default:
		return true
	}
}

// GetEntries returns all log entries.
func (p *LogPanel) GetEntries() []LogEntry {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	result := make([]LogEntry, len(p.entries))
	copy(result, p.entries)
	return result
}

// SetAutoScroll enables or disables auto-scrolling.
func (p *LogPanel) SetAutoScroll(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.autoScroll = enabled
}

func (p *LogPanel) formatEntry(entry LogEntry) string {
	ts := entry.Timestamp.Format("15:04:05.000")
	return fmt.Sprintf("[%s] %s → %s", ts, entry.Direction, entry.Message)
}
