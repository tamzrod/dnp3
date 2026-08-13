package master

import (
	"errors"
	"io"
	"net"
	"syscall"
	"testing"

	internalmaster "dnp3/internal/master"
)

// MEXT-025 — IsDisconnectError must recognize the TCP errno signals of a peer
// drop (broken-pipe write, connection reset/aborted, timeout) in addition to
// the canonical close sentinels. Without this, a peer drop mid-session leaves
// the master stuck in a connected state (the public state never flips to
// Disconnected), so a subsequent recovery/reconnect cannot proceed.
func TestIsDisconnectErrorPeerDropErrnos(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"transport_disconnected", internalmaster.ErrTransportDisconnected, true},
		{"eof", io.EOF, true},
		{"unexpected_eof", io.ErrUnexpectedEOF, true},
		{"net_closed", net.ErrClosed, true},
		// MEXT-025: peer-drop write/read errnos (wrapped as they arrive from the
		// transport: "...: TCP send failed: write tcp ...: write: broken pipe").
		{"broken_pipe_wrapped", errors.New("TCP send failed: write tcp 127.0.0.1:a->127.0.0.1:b: write: broken pipe"), false},
		{"econnreset_errno", syscall.ECONNRESET, true},
		{"epipe_errno", syscall.EPIPE, true},
		{"econnaborted_errno", syscall.ECONNABORTED, true},
		{"etimedout_errno", syscall.ETIMEDOUT, true},
		// A bare "broken pipe" string is NOT treated as a disconnect (we match the
		// errno, not the message) — this documents that string-matching is not the
		// mechanism and guards against accidental over-matching.
		{"plain_broken_pipe_string", errors.New("broken pipe"), false},
		// An errno nested inside a wrapped net.OpError-style chain (the real shape
		// produced by the net package on a dropped peer) must be recognized.
		{"wrapped_op_errno", &net.OpError{Op: "write", Net: "tcp", Err: &net.OpError{Op: "write", Err: syscall.EPIPE}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := internalmaster.IsDisconnectError(tc.err); got != tc.want {
				t.Fatalf("IsDisconnectError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
