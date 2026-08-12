// Package dnp3 provides the public API for the DNP3 protocol.
//
// This package is the official entry point for external applications
// that want to use the DNP3 protocol. It provides high-level interfaces
// for both Master clients and Outstation servers.
//
// # Package Structure
//
// The package is organized into the following subpackages:
//
//   - dnp3/types   - Common data types (BinaryInput, AnalogInput, Counter, etc.)
//   - dnp3/master  - Master client API
//   - dnp3/outstation - Outstation server API
//
// # Usage Example
//
// Create a Master client:
//
//	client, err := dnp3/master.NewClient(dnp3/master.WithAddress(1024))
//
// Create an Outstation server:
//
//	server, err := dnp3/outstation.NewServer(dnp3/outstation.WithAddress(1024))
//
// # Architecture
//
// This package provides a public facade layer that wraps the internal
// protocol implementations. The internal packages (internal/*) contain
// the detailed protocol logic and are not part of the public API.
//
// External consumers should only import from pkg/dnp3/* and should
// not directly import internal/* packages.
package dnp3

import "errors"

// Version information
const (
	// Version is the DNP3 library version
	Version = "0.1.0"

	// ProtocolVersion is the supported DNP3 protocol version
	ProtocolVersion = "3.0"
)

// TransportType represents the type of network transport
type TransportType int

const (
	// TCP transport over IP networks
	TCP TransportType = iota
	// TLS transport over IP networks with encryption
	TLS
)

// String returns the transport type name
func (t TransportType) String() string {
	switch t {
	case TCP:
		return "TCP"
	case TLS:
		return "TLS"
	default:
		return "Unknown"
	}
}

// ConnectionState represents the state of a connection
type ConnectionState int

const (
	// StateDisconnected - no connection established
	StateDisconnected ConnectionState = iota
	// StateConnecting - connection in progress
	StateConnecting
	// StateConnected - connection established
	StateConnected
	// StateInitialized - DNP3 session initialized
	StateInitialized
	// StateActive - actively exchanging data
	StateActive
	// StateError - error condition
	StateError
)

// String returns the connection state name
func (s ConnectionState) String() string {
	switch s {
	case StateDisconnected:
		return "Disconnected"
	case StateConnecting:
		return "Connecting"
	case StateConnected:
		return "Connected"
	case StateInitialized:
		return "Initialized"
	case StateActive:
		return "Active"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// Package-level error definitions
var (
	// ErrNotConnected indicates the client/server is not connected
	ErrNotConnected = errors.New("not connected")

	// ErrTimeout indicates an operation timed out
	ErrTimeout = errors.New("operation timeout")

	// ErrClosed indicates the connection has been closed
	ErrClosed = errors.New("connection closed")

	// ErrInvalidResponse indicates an invalid response was received
	ErrInvalidResponse = errors.New("invalid response")

	// ErrMaxRetries indicates maximum retry attempts exceeded
	ErrMaxRetries = errors.New("maximum retries exceeded")

	// ErrOutstationNotFound indicates the outstation was not found
	ErrOutstationNotFound = errors.New("outstation not found")

	// ErrInvalidRequest indicates an invalid request was made
	ErrInvalidRequest = errors.New("invalid request")

	// ErrConfiguration indicates a configuration error
	ErrConfiguration = errors.New("configuration error")

	// ErrUnsupportedFunction indicates unsupported DNP3 function
	ErrUnsupportedFunction = errors.New("unsupported function")

	// ErrSequenceError indicates a sequence number error
	ErrSequenceError = errors.New("sequence error")

	// ErrContextCanceled indicates the operation was aborted because the
	// caller's context was cancelled (SAFE-03).
	ErrContextCanceled = errors.New("context canceled")

	// ErrUnsupportedGroup indicates a read request asked for an object
	// group/variation outside the v0 supported profile (DNP3-029). The v0
	// profile reads only Binary Input (G1), Counter (G20), and Analog Input
	// (G30); variation 0 ("any") is accepted for these groups.
	ErrUnsupportedGroup = errors.New("unsupported object group/variation")

	// ErrUnsupportedOption indicates the caller used a public API option that
	// is outside the v0 supported profile (DNP3-030): TLS transport, unsolicited
	// responses, select-before-operate, and direct-operate-no-response.
	ErrUnsupportedOption = errors.New("unsupported option for v0 profile")
)

// ConfigurationError represents a configuration validation error
type ConfigurationError struct {
	Field   string
	Message string
}

func (e *ConfigurationError) Error() string {
	return "configuration error on " + e.Field + ": " + e.Message
}

// NewConfigurationError creates a new ConfigurationError
func NewConfigurationError(field, message string) *ConfigurationError {
	return &ConfigurationError{
		Field:   field,
		Message: message,
	}
}

// ProtocolError represents a DNP3 protocol error
type ProtocolError struct {
	FuncCode    uint8
	IIN         [2]byte
	Description string
}

func (e *ProtocolError) Error() string {
	return e.Description
}

// NewProtocolError creates a new ProtocolError
func NewProtocolError(funcCode uint8, iin [2]byte, description string) *ProtocolError {
	return &ProtocolError{
		FuncCode:    funcCode,
		IIN:         iin,
		Description: description,
	}
}
