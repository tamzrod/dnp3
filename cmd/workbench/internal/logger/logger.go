// Package logger provides logging functionality for the workbench.
package logger

import (
	"fmt"
	"sync"
	"time"
)

// Level represents log level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Entry represents a log entry.
type Entry struct {
	Timestamp time.Time
	Level    Level
	Message  string
	Source   string
}

// LogCallback is called when a new log entry is created.
type LogCallback func(*Entry)

// Logger provides thread-safe logging with a circular buffer.
type Logger struct {
	mu       sync.RWMutex
	entries  []Entry
	buffer   *Buffer
	callback LogCallback
	minLevel Level
	source   string
}

// Buffer is a circular buffer for log entries.
type Buffer struct {
	entries []Entry
	size   int
	head   int
	count  int
	mu     sync.RWMutex
}

// New creates a new logger.
func New() *Logger {
	return &Logger{
		entries:  make([]Entry, 0, 1000),
		buffer:   NewBuffer(1000),
		minLevel: LevelInfo,
		source:   "app",
	}
}

// NewBuffer creates a new circular buffer.
func NewBuffer(size int) *Buffer {
	return &Buffer{
		entries: make([]Entry, size),
		size:    size,
	}
}

// SetCallback sets the callback for log entries.
func (l *Logger) SetCallback(cb LogCallback) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.callback = cb
}

// SetLevel sets the minimum log level.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

// SetSource sets the log source.
func (l *Logger) SetSource(source string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.source = source
}

// Debug logs a debug message.
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs an info message.
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error logs an error message.
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Errorf logs an error with format.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// log creates a log entry.
func (l *Logger) log(level Level, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.minLevel {
		return
	}

	entry := Entry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   fmt.Sprintf(format, args...),
		Source:   l.source,
	}

	l.entries = append(l.entries, entry)
	l.buffer.Push(entry)

	if l.callback != nil {
		// Create copy for callback
		cbEntry := entry
		go l.callback(&cbEntry)
	}
}

// Entries returns all log entries.
func (l *Logger) Entries() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	result := make([]Entry, len(l.entries))
	copy(result, l.entries)
	return result
}

// Recent returns the most recent entries.
func (l *Logger) Recent(n int) []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if n > len(l.entries) {
		n = len(l.entries)
	}
	
	start := len(l.entries) - n
	result := make([]Entry, n)
	copy(result, l.entries[start:])
	return result
}

// Clear removes all log entries.
func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = make([]Entry, 0, 1000)
	l.buffer.Clear()
}

// Buffer returns the circular buffer.
func (l *Logger) Buffer() *Buffer {
	return l.buffer
}

// Push adds an entry to the buffer.
func (b *Buffer) Push(entry Entry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.entries[b.head] = entry
	b.head = (b.head + 1) % b.size
	if b.count < b.size {
		b.count++
	}
}

// Entries returns all buffer entries in order.
func (b *Buffer) Entries() []Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.count == 0 {
		return nil
	}

	result := make([]Entry, b.count)
	
	if b.count < b.size {
		// Buffer not full yet
		copy(result, b.entries[:b.count])
	} else {
		// Buffer is full, head points to oldest
		for i := 0; i < b.count; i++ {
			idx := (b.head + i) % b.size
			result[i] = b.entries[idx]
		}
	}
	
	return result
}

// Clear removes all entries from the buffer.
func (b *Buffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.head = 0
	b.count = 0
}

// Size returns the buffer size.
func (b *Buffer) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size
}

// Count returns the number of entries in the buffer.
func (b *Buffer) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.count
}
