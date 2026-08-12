package dnp3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
)

// TestClassifyErrorMapsSentinels verifies ClassifyError recognizes every public
// error sentinel and the typed ConfigurationError (DNP3-043).
func TestClassifyErrorMapsSentinels(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want ErrorCode
	}{
		{"nil", nil, ErrorCodeUnknown},
		{"timeout", ErrTimeout, ErrorCodeTimeout},
		{"crc", ErrCRC, ErrorCodeCRC},
		{"sequence", ErrSequenceError, ErrorCodeSequence},
		{"unsupported_function", ErrUnsupportedFunction, ErrorCodeUnsupported},
		{"unsupported_group", ErrUnsupportedGroup, ErrorCodeUnsupported},
		{"unsupported_option", ErrUnsupportedOption, ErrorCodeUnsupported},
		{"not_connected", ErrNotConnected, ErrorCodeDisconnect},
		{"closed", ErrClosed, ErrorCodeDisconnect},
		{"request_outstanding", ErrRequestOutstanding, ErrorCodeBusy},
		{"canceled", ErrContextCanceled, ErrorCodeCanceled},
		{"configuration_sentinel", ErrConfiguration, ErrorCodeConfiguration},
		{"invalid_request", ErrInvalidRequest, ErrorCodeInvalid},
		{"invalid_response", ErrInvalidResponse, ErrorCodeInvalid},
		{"configuration_typed", NewConfigurationError("Timeout", "zero"), ErrorCodeConfiguration},
		{"unknown_foreign", errors.New("something else"), ErrorCodeUnknown},
		{"unknown_io", io.EOF, ErrorCodeUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClassifyError(tc.err); got != tc.want {
				t.Fatalf("ClassifyError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

// TestClassifyErrorUnwrapsChain verifies ClassifyError walks the wrapped error
// chain (errors.Is) so a boundary-wrapped error like "read failed: %w" is still
// recognized by category (DNP3-043). The API boundary attaches exactly one public
// sentinel per internal error, so each case carries a single category sentinel.
func TestClassifyErrorUnwrapsChain(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want ErrorCode
	}{
		{"wrapped_crc", fmt.Errorf("read failed: %w", ErrCRC), ErrorCodeCRC},
		{"wrapped_timeout", fmt.Errorf("read failed: %w", ErrTimeout), ErrorCodeTimeout},
		{"wrapped_sequence", fmt.Errorf("operate failed: %w", ErrSequenceError), ErrorCodeSequence},
		{"wrapped_disconnect", fmt.Errorf("read failed: %w", ErrNotConnected), ErrorCodeDisconnect},
		{"wrapped_busy", fmt.Errorf("read failed: %w", ErrRequestOutstanding), ErrorCodeBusy},
		{"wrapped_unsupported", fmt.Errorf("read: %w", ErrUnsupportedGroup), ErrorCodeUnsupported},
		{"wrapped_canceled", fmt.Errorf("read: %w", ErrContextCanceled), ErrorCodeCanceled},
		{"double_wrapped_crc", fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", ErrCRC)), ErrorCodeCRC},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClassifyError(tc.err); got != tc.want {
				t.Fatalf("ClassifyError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

// TestErrorCodeString verifies the human-readable names (DNP3-043).
func TestErrorCodeString(t *testing.T) {
	cases := []struct {
		code ErrorCode
		want string
	}{
		{ErrorCodeUnknown, "unknown"},
		{ErrorCodeTimeout, "timeout"},
		{ErrorCodeCRC, "crc"},
		{ErrorCodeSequence, "sequence"},
		{ErrorCodeUnsupported, "unsupported"},
		{ErrorCodeDisconnect, "disconnect"},
		{ErrorCodeConfiguration, "configuration"},
		{ErrorCodeCanceled, "canceled"},
		{ErrorCodeBusy, "busy"},
		{ErrorCodeInvalid, "invalid"},
		{ErrorCode(999), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.code.String(); got != tc.want {
			t.Errorf("ErrorCode(%d).String() = %q, want %q", tc.code, got, tc.want)
		}
	}
}

// TestClassifyErrorPrecedence verifies the deliberate classification order
// (DNP3-043): a caller-side cancellation is reported as canceled even when the
// operation was aborted mid-flight (the public client wraps ctx-cancel as
// "%w: %v" with ErrContextCanceled, and context.Canceled is not itself a
// disconnect sentinel).
func TestClassifyErrorPrecedence(t *testing.T) {
	// Realistic ctx-cancel boundary error: ErrContextCanceled wrapping
	// context.Canceled. context.Canceled is not a disconnect sentinel, so this
	// must classify as canceled, not unknown.
	canceled := fmt.Errorf("%w: %w", ErrContextCanceled, context.Canceled)
	if got := ClassifyError(canceled); got != ErrorCodeCanceled {
		t.Fatalf("ctx-canceled boundary = %v, want canceled", got)
	}

	// Canceled is checked before disconnect: even if a disconnect sentinel were
	// present in the same chain, the caller-side cancellation wins.
	canceledWithDisconnect := fmt.Errorf("%w: %w", ErrContextCanceled, ErrNotConnected)
	if got := ClassifyError(canceledWithDisconnect); got != ErrorCodeCanceled {
		t.Fatalf("canceled+disconnect chain = %v, want canceled", got)
	}
}
