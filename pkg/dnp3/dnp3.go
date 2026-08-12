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

	// ErrCRC indicates a received link frame failed CRC validation (DNP3-043).
	// A corrupted frame is transient line noise; the request is retryable. This
	// public sentinel is attached at the API boundary when an internal CRC
	// failure propagates out, so callers can distinguish it without importing
	// internal packages.
	ErrCRC = errors.New("frame CRC error")

	// ErrRequestOutstanding indicates a request to the same outstation is
	// already in flight (DNP3-040/043). A DNP3 link permits at most one
	// outstanding master request per outstation for MVP.
	ErrRequestOutstanding = errors.New("request already outstanding for outstation")
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

// ErrorCode is a stable, comparable classification of a DNP3 error returned by
// the public API (DNP3-043). It lets callers distinguish failure categories
// without inspecting error message strings or importing internal packages.
//
// Use [ClassifyError] to derive the code from an error. The category reflects
// the underlying failure source, not the operation that surfaced it: for
// example a Read that fails because the link died reports ErrorCodeDisconnect.
type ErrorCode int

const (
	// ErrorCodeUnknown is the fallback for errors the library does not
	// recognize (including caller-supplied errors). Callers should treat it as
	// "unclassified" rather than a specific failure.
	ErrorCodeUnknown ErrorCode = iota
	// ErrorCodeTimeout indicates a response/confirmation did not arrive within
	// the configured timeout (DNP3-009/034).
	ErrorCodeTimeout
	// ErrorCodeCRC indicates a received link frame failed CRC validation
	// (DNP3-034/043). Transient line noise; retryable.
	ErrorCodeCRC
	// ErrorCodeSequence indicates an application-layer sequence mismatch
	// (confirmation or response SEQ did not match the outstanding request)
	// (DNP3-009/010).
	ErrorCodeSequence
	// ErrorCodeUnsupported indicates the caller requested an object group,
	// function, or option outside the v0 supported profile (DNP3-029/030).
	ErrorCodeUnsupported
	// ErrorCodeDisconnect indicates the transport/peer closed the connection
	// (or the session was closed/idle-timed-out). The link is dead; not
	// retryable without reconnecting (DNP3-031/042).
	ErrorCodeDisconnect
	// ErrorCodeConfiguration indicates an invalid configuration was supplied
	// (DNP3-041).
	ErrorCodeConfiguration
	// ErrorCodeCanceled indicates the caller's context was cancelled while the
	// operation was outstanding (SAFE-03).
	ErrorCodeCanceled
	// ErrorCodeBusy indicates a request to the same outstation is already in
	// flight (DNP3-040).
	ErrorCodeBusy
	// ErrorCodeInvalid indicates a malformed request or response that is not
	// covered by a more specific category.
	ErrorCodeInvalid
)

// String returns a human-readable name for the error code.
func (c ErrorCode) String() string {
	switch c {
	case ErrorCodeTimeout:
		return "timeout"
	case ErrorCodeCRC:
		return "crc"
	case ErrorCodeSequence:
		return "sequence"
	case ErrorCodeUnsupported:
		return "unsupported"
	case ErrorCodeDisconnect:
		return "disconnect"
	case ErrorCodeConfiguration:
		return "configuration"
	case ErrorCodeCanceled:
		return "canceled"
	case ErrorCodeBusy:
		return "busy"
	case ErrorCodeInvalid:
		return "invalid"
	default:
		return "unknown"
	}
}

// ClassifyError maps an error returned by the public API to its [ErrorCode]
// category (DNP3-043). A nil error returns ErrorCodeUnknown. The classifier
// walks the error chain with errors.Is/errors.As so wrapped errors (e.g.
// "read failed: %w") are still recognized.
//
// The order is deliberate: cancellation and configuration are caller-side
// conditions checked first; the protocol-specific classes (CRC, sequence,
// timeout, unsupported) follow; disconnect is a terminal transport condition;
// invalid is the catch-all for malformed responses.
func ClassifyError(err error) ErrorCode {
	if err == nil {
		return ErrorCodeUnknown
	}
	if errors.Is(err, ErrContextCanceled) {
		return ErrorCodeCanceled
	}
	var ce *ConfigurationError
	if errors.As(err, &ce) || errors.Is(err, ErrConfiguration) {
		return ErrorCodeConfiguration
	}
	if errors.Is(err, ErrUnsupportedFunction) ||
		errors.Is(err, ErrUnsupportedGroup) ||
		errors.Is(err, ErrUnsupportedOption) {
		return ErrorCodeUnsupported
	}
	if errors.Is(err, ErrCRC) {
		return ErrorCodeCRC
	}
	if errors.Is(err, ErrSequenceError) {
		return ErrorCodeSequence
	}
	if errors.Is(err, ErrTimeout) {
		return ErrorCodeTimeout
	}
	if errors.Is(err, ErrRequestOutstanding) {
		return ErrorCodeBusy
	}
	if errors.Is(err, ErrNotConnected) || errors.Is(err, ErrClosed) {
		return ErrorCodeDisconnect
	}
	if errors.Is(err, ErrInvalidRequest) || errors.Is(err, ErrInvalidResponse) {
		return ErrorCodeInvalid
	}
	return ErrorCodeUnknown
}
