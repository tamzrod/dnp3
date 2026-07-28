package tui

import (
	"sync"
	"time"
)

// LogEntry represents a single log entry.
type LogEntry struct {
	Timestamp time.Time
	Direction string // "→" (sent), "←" (received), "●" (info), "⚠" (warning), "✗" (error)
	Message   string
	Color     string // ANSI color name
}

// Log represents the log display.
type Log struct {
	mu       sync.RWMutex
	entries  []LogEntry
	maxSize int
	bounds  Rect
}

// NewLog creates a new log display.
func NewLog(bounds Rect) *Log {
	return &Log{
		entries:  []LogEntry{},
		maxSize:  100,
		bounds:   bounds,
	}
}

// Add adds a log entry.
func (l *Log) Add(direction, message, color string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now(),
		Direction: direction,
		Message:   message,
		Color:     color,
	}

	l.entries = append(l.entries, entry)

	// Trim if too large
	if len(l.entries) > l.maxSize {
		l.entries = l.entries[len(l.entries)-l.maxSize:]
	}
}

// Info adds an info log entry.
func (l *Log) Info(message string) {
	l.Add("●", message, "white")
}

// Send adds a sent message entry.
func (l *Log) Send(message string) {
	l.Add("→", message, "cyan")
}

// Recv adds a received message entry.
func (l *Log) Recv(message string) {
	l.Add("←", message, "green")
}

// Warn adds a warning entry.
func (l *Log) Warn(message string) {
	l.Add("⚠", message, "yellow")
}

// Error adds an error entry.
func (l *Log) Error(message string) {
	l.Add("✗", message, "red")
}

// Clear clears all log entries.
func (l *Log) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = []LogEntry{}
}

// GetEntries returns all log entries.
func (l *Log) GetEntries() []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.entries
}

// Draw renders the log display.
func (l *Log) Draw(scr *Screen) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	height := l.bounds.Bottom - l.bounds.Top + 1

	// Draw border
	scr.Print(l.bounds.Top, l.bounds.Left, "┌─LOG─────────────────────────────────────────────────────┐")

	// Draw entries
	visible := height - 3
	start := len(l.entries) - visible
	if start < 0 {
		start = 0
	}

	for i := start; i < len(l.entries) && i < start+visible; i++ {
		y := l.bounds.Top + 1 + (i - start)
		entry := l.entries[i]

		// Format: [HH:MM:SS.mmm] DIRECTION Message
		timeStr := entry.Timestamp.Format("15:04:05.000")
		text := "[" + timeStr + "] " + entry.Direction + " " + entry.Message

		// Truncate to fit
		text = Truncate(text, scr.width-2)

		// Draw with color
		if entry.Color != "" {
			scr.PrintStyled(y, 2, text, entry.Color)
		} else {
			scr.Print(y, 2, text)
		}
	}

	// Draw bottom border
	scr.Print(l.bounds.Bottom, l.bounds.Left, "└─────────────────────────────────────────────────────────────┘")
}

// DrawSimple renders the log without borders.
func (l *Log) DrawSimple(scr *Screen, startY, height int) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	visible := height - 1

	for i := 0; i < visible && i < len(l.entries); i++ {
		idx := len(l.entries) - visible + i
		if idx < 0 {
			continue
		}
		entry := l.entries[idx]

		// Format: [HH:MM:SS.mmm] DIRECTION Message
		timeStr := entry.Timestamp.Format("15:04:05")
		text := "[" + timeStr + "] " + entry.Direction + " " + entry.Message

		// Truncate to fit
		text = Truncate(text, scr.width-2)

		// Draw with color
		if entry.Color != "" {
			scr.PrintStyled(startY+i, 1, text, entry.Color)
		} else {
			scr.Print(startY+i, 1, text)
		}
	}
}
