package master

import (
	"bytes"
	"errors"
	"sync"
	"testing"

	"dnp3/internal/dll/frame"
	"dnp3/internal/testutils/golden"
)

// MEXT-017 — Link handshake external frame vectors.
//
// These tests lock the IEEE 1815 wire shape of the master's link-layer
// handshake (Reset Link Stations + Request Link Status requests) against
// external-style golden byte vectors, and prove Connect requires BOTH
// exchanges (ACK then Link Status) — any mismatch fails Connect.

const (
	vecMasterAddr   uint16 = 0x0003
	vecOutstationID uint16 = 0x0004
)

// scriptedTransport serves a queued sequence of response frames to Receive()
// and records every frame the master Send()s. Each Receive() call pops the
// next queued response (cyclically if exhausted). A nil entry surfaces a
// transport error so a "missing exchange" scenario fails the handshake.
type scriptedTransport struct {
	mu      sync.Mutex
	sent    [][]byte
	resps   [][]byte
	idx     int
	timeout int
}

func (t *scriptedTransport) Send(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	cp := make([]byte, len(data))
	copy(cp, data)
	t.sent = append(t.sent, cp)
	return nil
}

func (t *scriptedTransport) Receive() ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.idx >= len(t.resps) {
		return nil, errors.New("no scripted response: transport closed")
	}
	r := t.resps[t.idx]
	t.idx++
	if r == nil {
		return nil, errors.New("scripted transport error")
	}
	return r, nil
}

func (t *scriptedTransport) SetTimeout(ms int) { t.timeout = ms }

func (t *scriptedTransport) sentBytes() [][]byte {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([][]byte, len(t.sent))
	copy(out, t.sent)
	return out
}

// newVecMaster builds a master wired to vecMasterAddr/vecOutstationID (matching
// the golden fixtures) with a NACK-retry budget of 0 so a handshake NACK fails
// immediately rather than exercising the retry loop.
func newVecMaster(t *testing.T, tr *scriptedTransport) *Master {
	t.Helper()
	m := NewMaster(&Config{
		MasterAddress: vecMasterAddr,
		Timeout:       500,
		MaxRetries:    3,
	})
	m.AddOutstation(vecOutstationID, "vec")
	m.SetTransport(tr)
	// NACKRetries=0: a single NACK terminates the handshake immediately.
	m.SetRetryPolicy(&RetryPolicy{NACKRetries: 0})
	return m
}

// TestLinkHandshakeRequestVectors asserts the master emits spec-correct IEEE
// 1815 wire bytes for both handshake request frames (external-style golden
// vectors in active_work/testdata). Connect drives the full handshake, so the
// first two sent frames are the Reset Link Stations and Request Link Status
// requests.
func TestLinkHandshakeRequestVectors(t *testing.T) {
	tr := &scriptedTransport{
		resps: [][]byte{
			mustLoadGolden(t, "link-secondary-ack.hex"),
			mustLoadGolden(t, "link-secondary-link-status.hex"),
		},
	}
	m := newVecMaster(t, tr)
	if err := m.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if m.State() != StateActive {
		t.Fatalf("state = %v, want Active", m.State())
	}

	sent := tr.sentBytes()
	if len(sent) < 2 {
		t.Fatalf("expected >=2 sent handshake frames, got %d", len(sent))
	}

	wantReset, err := golden.LoadHex("link-reset-link-stations.hex")
	if err != nil {
		t.Fatalf("load reset golden: %v", err)
	}
	if !bytes.Equal(sent[0], wantReset) {
		t.Fatalf("reset-link-stations request = % X, want golden % X", sent[0], wantReset)
	}
	// Decode-level field assertion (not just byte equality): control/func/addrs.
	decoded, err := frame.Decode(sent[0])
	if err != nil {
		t.Fatalf("decode reset request: %v", err)
	}
	if got := decoded.Control.ToByte(); got != 0xC0 {
		t.Errorf("reset control byte = 0x%02X, want 0xC0 (DIR|PRM|func0)", got)
	}
	if decoded.DestAddr != vecOutstationID || decoded.SrcAddr != vecMasterAddr {
		t.Errorf("reset addrs = dest 0x%04X src 0x%04X, want dest 0x%04X src 0x%04X",
			decoded.DestAddr, decoded.SrcAddr, vecOutstationID, vecMasterAddr)
	}

	wantReq, err := golden.LoadHex("link-request-link-status.hex")
	if err != nil {
		t.Fatalf("load link-status-request golden: %v", err)
	}
	if !bytes.Equal(sent[1], wantReq) {
		t.Fatalf("request-link-status request = % X, want golden % X", sent[1], wantReq)
	}
	decoded2, err := frame.Decode(sent[1])
	if err != nil {
		t.Fatalf("decode link-status request: %v", err)
	}
	if got := decoded2.Control.ToByte(); got != 0xC9 {
		t.Errorf("link-status control byte = 0x%02X, want 0xC9 (DIR|PRM|func9)", got)
	}
	if decoded2.DestAddr != vecOutstationID || decoded2.SrcAddr != vecMasterAddr {
		t.Errorf("link-status addrs = dest 0x%04X src 0x%04X, want dest 0x%04X src 0x%04X",
			decoded2.DestAddr, decoded2.SrcAddr, vecOutstationID, vecMasterAddr)
	}
}

// TestLinkHandshakeRequiresBothExchanges proves Connect needs BOTH the ACK and
// the Link Status: a valid ACK followed by a valid Link Status succeeds.
func TestLinkHandshakeRequiresBothExchanges(t *testing.T) {
	tr := &scriptedTransport{
		resps: [][]byte{
			mustLoadGolden(t, "link-secondary-ack.hex"),
			mustLoadGolden(t, "link-secondary-link-status.hex"),
		},
	}
	m := newVecMaster(t, tr)
	if err := m.Connect(); err != nil {
		t.Fatalf("Connect should succeed with ACK + Link Status: %v", err)
	}
	if m.State() != StateActive {
		t.Fatalf("state = %v, want Active", m.State())
	}
}

// TestLinkHandshakeNACKFailsConnect proves a secondary NACK on the Reset Link
// Stations exchange fails Connect (mismatch). NACKRetries=0 makes it terminal
// on the first NACK.
func TestLinkHandshakeNACKFailsConnect(t *testing.T) {
	tr := &scriptedTransport{
		resps: [][]byte{
			mustLoadGolden(t, "link-secondary-nack.hex"),
			mustLoadGolden(t, "link-secondary-link-status.hex"),
		},
	}
	m := newVecMaster(t, tr)
	err := m.Connect()
	if err == nil {
		t.Fatal("Connect should fail on NACK, got nil")
	}
	if m.State() != StateError {
		t.Errorf("state = %v, want Error after handshake failure", m.State())
	}
}

// TestLinkHandshakeWrongFuncOnLinkStatusFailsConnect proves the second exchange
// is validated: an ACK (func 0) where a Link Status (func 2) is expected fails
// Connect even though the first exchange (ACK) was valid.
func TestLinkHandshakeWrongFuncOnLinkStatusFailsConnect(t *testing.T) {
	tr := &scriptedTransport{
		resps: [][]byte{
			mustLoadGolden(t, "link-secondary-ack.hex"),
			// ACK again instead of a Link Status response.
			mustLoadGolden(t, "link-secondary-ack.hex"),
		},
	}
	m := newVecMaster(t, tr)
	if err := m.Connect(); err == nil {
		t.Fatal("Connect should fail when Link Status response has wrong func, got nil")
	}
	if m.State() != StateError {
		t.Errorf("state = %v, want Error", m.State())
	}
}

// TestLinkHandshakeMissingLinkStatusFailsConnect proves Connect requires the
// second exchange at all: a transport that closes after the ACK (no Link Status
// response) fails Connect.
func TestLinkHandshakeMissingLinkStatusFailsConnect(t *testing.T) {
	tr := &scriptedTransport{
		resps: [][]byte{
			mustLoadGolden(t, "link-secondary-ack.hex"),
		},
	}
	m := newVecMaster(t, tr)
	if err := m.Connect(); err == nil {
		t.Fatal("Connect should fail when the Link Status exchange is missing, got nil")
	}
	if m.State() != StateError {
		t.Errorf("state = %v, want Error", m.State())
	}
}

// TestLinkHandshakeGoldenResponseDecode asserts the secondary golden response
// fixtures decode to the IEEE 1815 fields the validators expect (external-style
// vector round-trip), independent of the master's encoder.
func TestLinkHandshakeGoldenResponseDecode(t *testing.T) {
	cases := []struct {
		name     string
		fixture  string
		ctrlByte byte
		funcCode uint8
	}{
		{"secondary ACK", "link-secondary-ack.hex", 0x00, frame.FuncAck},
		{"secondary Link Status", "link-secondary-link-status.hex", 0x02, frame.FuncLinkStatus},
		{"secondary NACK", "link-secondary-nack.hex", 0x01, frame.FuncNack},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			raw, err := golden.LoadHex(c.fixture)
			if err != nil {
				t.Fatalf("load %s: %v", c.fixture, err)
			}
			f, err := frame.Decode(raw)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got := f.Control.ToByte(); got != c.ctrlByte {
				t.Errorf("control byte = 0x%02X, want 0x%02X", got, c.ctrlByte)
			}
			if f.Control.FuncCode != c.funcCode {
				t.Errorf("func code = %d, want %d", f.Control.FuncCode, c.funcCode)
			}
			if f.Control.DIR || f.Control.PRM {
				t.Errorf("DIR=%v PRM=%v, want both false (secondary)", f.Control.DIR, f.Control.PRM)
			}
			if f.SrcAddr != vecOutstationID || f.DestAddr != vecMasterAddr {
				t.Errorf("addrs = src 0x%04X dest 0x%04X, want src 0x%04X dest 0x%04X",
					f.SrcAddr, f.DestAddr, vecOutstationID, vecMasterAddr)
			}
		})
	}
}

func mustLoadGolden(t *testing.T, name string) []byte {
	t.Helper()
	b, err := golden.LoadHex(name)
	if err != nil {
		t.Fatalf("load golden %s: %v", name, err)
	}
	return b
}
