// Package session provides DNP3 session management.
package session

import (
	"context"
	"sync"
)

// Manager manages DNP3 sessions.
type Manager struct {
	mu       sync.RWMutex
	session  Session
	logger   *ManagerLogger
}

// NewManager creates a new session manager.
func NewManager() *Manager {
	return &Manager{
		logger: &ManagerLogger{},
	}
}

// CreateMasterSession creates a new Master session.
func (m *Manager) CreateMasterSession() *MasterSession {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close existing session if any
	if m.session != nil {
		m.session.Close()
	}

	session := NewMasterSession(m.logger)
	m.session = session
	return session
}

// GetSession returns the current session.
func (m *Manager) GetSession() Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.session
}

// Close closes the current session.
func (m *Manager) Close(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.session != nil {
		return m.session.Close()
	}
	return nil
}

// Logger returns the manager logger.
func (m *Manager) Logger() *ManagerLogger {
	return m.logger
}

// ManagerLogger logs messages for the workbench.
type ManagerLogger struct {
	logs []LogEntry
	mu   sync.RWMutex
}

// LogEntry represents a log entry.
type LogEntry struct {
	Timestamp string
	Level     string
	Message   string
}

// Info logs an info message.
func (l *ManagerLogger) Info(format string, args ...interface{}) {
	l.addLog("INFO", format, args...)
}

// Error logs an error message.
func (l *ManagerLogger) Error(format string, args ...interface{}) {
	l.addLog("ERROR", format, args...)
}

// Debug logs a debug message.
func (l *ManagerLogger) Debug(format string, args ...interface{}) {
	l.addLog("DEBUG", format, args...)
}

func (l *ManagerLogger) addLog(level, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	// In a real implementation, this would use a proper logging library
	_ = format
	_ = args
	_ = level
	l.logs = append(l.logs, LogEntry{
		Level:   level,
		Message: format,
	})
}

// GetLogs returns all log entries.
func (l *ManagerLogger) GetLogs() []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	logs := make([]LogEntry, len(l.logs))
	copy(logs, l.logs)
	return logs
}

// Clear clears the log.
func (l *ManagerLogger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = nil
}
