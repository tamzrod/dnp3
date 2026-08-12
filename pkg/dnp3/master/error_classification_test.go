package master

import (
	"context"
	"errors"
	"sync"
	"testing"

	"dnp3/internal/master"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
)

// TestPublicErrorClassification verifies the public client attaches the correct
// public error sentinel at the API boundary for each failure category, so
// callers can classify failures via dnp3.ClassifyError without importing
// internal packages (DNP3-043). The internal error chain is preserved so
// existing errors.Is checks keep working.
func TestPublicErrorClassification(t *testing.T) {
	// CRC: a corrupted DLL frame fails CRC validation in processReceivedBytes,
	// surfaces as internal master.ErrCRCError, and is wrapped with dnp3.ErrCRC.
	t.Run("crc", func(t *testing.T) {
		cc := newConnectedClientWithTransport(t, &badCRCTransport{})
		_, err := cc.Read(context.Background(), &types.ReadRequest{
			Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
		})
		if err == nil {
			t.Fatal("expected error for CRC failure")
		}
		if !errors.Is(err, dnp3.ErrCRC) {
			t.Fatalf("error = %v, want dnp3.ErrCRC in chain", err)
		}
		if !errors.Is(err, master.ErrCRCError) {
			t.Fatalf("error = %v, want internal ErrCRCError preserved in chain", err)
		}
		if got := dnp3.ClassifyError(err); got != dnp3.ErrorCodeCRC {
			t.Fatalf("ClassifyError = %v, want ErrorCodeCRC", got)
		}
	})

	// Disconnect: a peer close (io.EOF) surfaces as a disconnect; the public
	// ErrNotConnected sentinel is attached while the internal
	// ErrTransportDisconnected sentinel is preserved.
	t.Run("disconnect", func(t *testing.T) {
		cc := newConnectedClientWithTransport(t, &peerCloseTransport{})
		_, err := cc.Read(context.Background(), &types.ReadRequest{
			Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
		})
		if err == nil {
			t.Fatal("expected error on peer close")
		}
		if !errors.Is(err, dnp3.ErrNotConnected) {
			t.Fatalf("error = %v, want dnp3.ErrNotConnected in chain", err)
		}
		if !errors.Is(err, master.ErrTransportDisconnected) {
			t.Fatalf("error = %v, want internal ErrTransportDisconnected preserved", err)
		}
		if got := dnp3.ClassifyError(err); got != dnp3.ErrorCodeDisconnect {
			t.Fatalf("ClassifyError = %v, want ErrorCodeDisconnect", got)
		}
	})

	// Timeout: a Receive error that is NOT a disconnect surfaces as internal
	// ErrConfirmTimeout and is wrapped with dnp3.ErrTimeout.
	t.Run("timeout", func(t *testing.T) {
		cc := newConnectedClientWithTransport(t, &timeoutTransport{})
		_, err := cc.Read(context.Background(), &types.ReadRequest{
			Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
		})
		if err == nil {
			t.Fatal("expected error on receive timeout")
		}
		if !errors.Is(err, dnp3.ErrTimeout) {
			t.Fatalf("error = %v, want dnp3.ErrTimeout in chain", err)
		}
		if got := dnp3.ClassifyError(err); got != dnp3.ErrorCodeTimeout {
			t.Fatalf("ClassifyError = %v, want ErrorCodeTimeout", got)
		}
	})

	// Sequence mismatch: a response whose SEQ does not echo the request's SEQ
	// surfaces as internal ErrResponseSeqMismatch and is wrapped with
	// dnp3.ErrSequenceError.
	t.Run("sequence", func(t *testing.T) {
		cc := newConnectedClientWithTransport(t, &seqMismatchTransport{})
		_, err := cc.Read(context.Background(), &types.ReadRequest{
			Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
		})
		if err == nil {
			t.Fatal("expected error on sequence mismatch")
		}
		if !errors.Is(err, dnp3.ErrSequenceError) {
			t.Fatalf("error = %v, want dnp3.ErrSequenceError in chain", err)
		}
		if !errors.Is(err, master.ErrResponseSeqMismatch) {
			t.Fatalf("error = %v, want internal ErrResponseSeqMismatch preserved", err)
		}
		if got := dnp3.ClassifyError(err); got != dnp3.ErrorCodeSequence {
			t.Fatalf("ClassifyError = %v, want ErrorCodeSequence", got)
		}
	})
}

// TestPublicOperateErrorClassification verifies the Operate boundary attaches the
// public error sentinels symmetrically with Read (DNP3-043).
func TestPublicOperateErrorClassification(t *testing.T) {
	t.Run("crc", func(t *testing.T) {
		cc := newConnectedClientWithTransport(t, &badCRCTransport{})
		_, err := cc.Operate(context.Background(), &types.ControlOutput{
			Group: 12, Variation: 1, Index: 0,
			CommandType: types.DirectOperate,
			Value:       &types.BinaryCommandValue{Value: true},
		})
		if err == nil {
			t.Fatal("expected error for CRC failure")
		}
		if !errors.Is(err, dnp3.ErrCRC) {
			t.Fatalf("error = %v, want dnp3.ErrCRC in chain", err)
		}
		if got := dnp3.ClassifyError(err); got != dnp3.ErrorCodeCRC {
			t.Fatalf("ClassifyError = %v, want ErrorCodeCRC", got)
		}
	})

	t.Run("disconnect", func(t *testing.T) {
		cc := newConnectedClientWithTransport(t, &peerCloseTransport{})
		_, err := cc.Operate(context.Background(), &types.ControlOutput{
			Group: 12, Variation: 1, Index: 0,
			CommandType: types.DirectOperate,
			Value:       &types.BinaryCommandValue{Value: true},
		})
		if err == nil {
			t.Fatal("expected error on peer close")
		}
		if !errors.Is(err, dnp3.ErrNotConnected) {
			t.Fatalf("error = %v, want dnp3.ErrNotConnected in chain", err)
		}
		if !errors.Is(err, master.ErrTransportDisconnected) {
			t.Fatalf("error = %v, want internal ErrTransportDisconnected preserved", err)
		}
		if got := dnp3.ClassifyError(err); got != dnp3.ErrorCodeDisconnect {
			t.Fatalf("ClassifyError = %v, want ErrorCodeDisconnect", got)
		}
	})
}

// TestPublicErrorClassificationNotConnected verifies the pre-flight NotConnected
// check (no wire traffic) classifies as a disconnect (DNP3-043).
func TestPublicErrorClassificationNotConnected(t *testing.T) {
	c, err := NewClient(NewConfig())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cc := c.(*client)
	// Disconnected is the default state; Read returns ErrNotConnected pre-flight.
	_, rerr := cc.Read(context.Background(), &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
	})
	if !errors.Is(rerr, dnp3.ErrNotConnected) {
		t.Fatalf("Read error = %v, want ErrNotConnected", rerr)
	}
	if got := dnp3.ClassifyError(rerr); got != dnp3.ErrorCodeDisconnect {
		t.Fatalf("ClassifyError = %v, want ErrorCodeDisconnect", got)
	}

	// Operate pre-flight NotConnected.
	_, oerr := cc.Operate(context.Background(), &types.ControlOutput{
		Group: 12, Variation: 1, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if !errors.Is(oerr, dnp3.ErrNotConnected) {
		t.Fatalf("Operate error = %v, want ErrNotConnected", oerr)
	}
	if got := dnp3.ClassifyError(oerr); got != dnp3.ErrorCodeDisconnect {
		t.Fatalf("ClassifyError = %v, want ErrorCodeDisconnect", got)
	}
}

// TestPublicErrorClassificationUnsupported verifies the unsupported-group and
// unsupported-option rejections classify as ErrorCodeUnsupported (DNP3-043).
func TestPublicErrorClassificationUnsupported(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &pubReadEchoTransport{})

	_, err := cc.Read(context.Background(), &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 40, Variation: 1}},
	})
	if !errors.Is(err, dnp3.ErrUnsupportedGroup) {
		t.Fatalf("Read error = %v, want ErrUnsupportedGroup", err)
	}
	if got := dnp3.ClassifyError(err); got != dnp3.ErrorCodeUnsupported {
		t.Fatalf("ClassifyError = %v, want ErrorCodeUnsupported", got)
	}
}

// timeoutTransport returns a non-disconnect error from Receive, which the master
// surfaces as ErrConfirmTimeout (a response/confirmation timeout).
type timeoutTransport struct{}

func (t *timeoutTransport) Send(data []byte) error   { return nil }
func (t *timeoutTransport) SetTimeout(ms int)        {}
func (t *timeoutTransport) Receive() ([]byte, error) { return nil, errors.New("i/o timeout") }

// seqMismatchTransport echoes a fixed SEQ (5) that does not match the master's
// request SEQ, so processResponse rejects it with ErrResponseSeqMismatch. The
// dedicated-confirm path in waitForConfirmation accepts a dedicated confirm
// only if its SEQ matches; a mismatch there surfaces ErrConfirmSeqMismatch.
// To exercise the response-seq path specifically, this transport returns a full
// Read response carrying a constant SEQ different from the request's.
type seqMismatchTransport struct {
	mu      sync.Mutex
	lastSeq uint8
}

func (t *seqMismatchTransport) Send(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastSeq = extractPubRequestSeq(data)
	return nil
}

func (t *seqMismatchTransport) SetTimeout(ms int) {}

func (t *seqMismatchTransport) Receive() ([]byte, error) {
	// Always respond with SEQ 5, regardless of the request SEQ (which starts at
	// 0). The mismatch triggers ErrResponseSeqMismatch / ErrConfirmSeqMismatch.
	return buildPubReadResponse(5), nil
}
