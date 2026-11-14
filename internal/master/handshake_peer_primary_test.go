package master

import (
	"strings"
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

// encodeSecondaryHandshake builds a DIR=0 PRM=0 frame with the given function.
func encodeSecondaryHandshake(t *testing.T, fc uint8, dest, src uint16) []byte {
	t.Helper()
	raw, err := frame.Encode(&frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: fc},
		DestAddr: dest,
		SrcAddr:  src,
	})
	if err != nil {
		t.Fatalf("encode secondary: %v", err)
	}
	return raw
}

// encodePeerNullUnsolicited builds the primary unconfirmed user-data frame
// OpenDNP3 emits for a null unsolicited response (control 0x44).
func encodePeerNullUnsolicited(t *testing.T, dest, src uint16) []byte {
	t.Helper()
	apdu := al.NewUnsolicited(0, []byte{0x00, 0x00})
	frag := tl.Fragment{FIR: true, FIN: true, Data: apdu.Encode()}
	raw, err := frame.Encode(&frame.Frame{
		Control:  frame.Control{DIR: false, PRM: true, FuncCode: frame.FuncUnconfirmedUserData},
		DestAddr: dest,
		SrcAddr:  src,
		Data:     tl.EncodeFragment(frag),
	})
	if err != nil {
		t.Fatalf("encode primary: %v", err)
	}
	decoded, err := frame.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Control.ToByte() != 0x44 {
		t.Fatalf("primary control = 0x%02X, want 0x44", decoded.Control.ToByte())
	}
	return raw
}

func confirmFromSent(t *testing.T, raw []byte) *al.APDU {
	t.Helper()
	f, err := frame.Decode(raw)
	if err != nil {
		t.Fatalf("decode sent: %v", err)
	}
	frag, err := tl.DecodeFragment(f.Data)
	if err != nil {
		t.Fatalf("decode transport: %v", err)
	}
	apdu, err := al.Decode(frag.Data)
	if err != nil {
		t.Fatalf("decode apdu: %v", err)
	}
	return apdu
}

// TestLinkHandshakeSkipsPeerPrimaryThenConfirms proves a correctly addressed
// outstation-primary frame before the secondary ACK does not fail Reset Link,
// and that the null unsolicited fragment is confirmed with UNS set.
func TestLinkHandshakeSkipsPeerPrimaryThenConfirms(t *testing.T) {
	primary := encodePeerNullUnsolicited(t, vecMasterAddr, vecOutstationID)
	tr := &scriptedTransport{resps: [][]byte{
		primary,
		encodeSecondaryHandshake(t, frame.FuncAck, vecMasterAddr, vecOutstationID),
		encodeSecondaryHandshake(t, frame.FuncLinkStatus, vecMasterAddr, vecOutstationID),
	}}
	m := newVecMaster(t, tr)
	if err := m.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	sent := tr.sentBytes()
	if len(sent) != 3 {
		t.Fatalf("sent %d frames, want reset, link status, and unsolicited confirm", len(sent))
	}
	reset, err := frame.Decode(sent[0])
	if err != nil {
		t.Fatal(err)
	}
	if reset.Control.ToByte() != 0xC0 {
		t.Fatalf("first frame control = 0x%02X, want reset 0xC0", reset.Control.ToByte())
	}
	status, err := frame.Decode(sent[1])
	if err != nil {
		t.Fatal(err)
	}
	if status.Control.FuncCode != frame.FuncRequestLinkStatus {
		t.Fatalf("second frame func = %d, want request link status", status.Control.FuncCode)
	}
	apdu := confirmFromSent(t, sent[2])
	if apdu.FuncCode != al.FuncConfirm || !apdu.Control.UNS || apdu.Control.CON || apdu.Control.Seq != 0 {
		t.Fatalf("confirm = func %d UNS %v CON %v seq %d, want confirm UNS seq 0",
			apdu.FuncCode, apdu.Control.UNS, apdu.Control.CON, apdu.Control.Seq)
	}
}

// TestLinkHandshakePeerPrimaryBoundDoesNotRetryReset proves extra primary
// frames are not treated as a failed ACK that resends Reset Link Stations.
func TestLinkHandshakePeerPrimaryBoundDoesNotRetryReset(t *testing.T) {
	primary := encodePeerNullUnsolicited(t, vecMasterAddr, vecOutstationID)
	resps := make([][]byte, maxHandshakePeerFrames+1)
	for i := range resps {
		resps[i] = primary
	}
	tr := &scriptedTransport{resps: resps}
	m := newVecMaster(t, tr)
	err := m.Connect()
	if err == nil {
		t.Fatal("Connect succeeded with no secondary ACK")
	}
	if !strings.Contains(err.Error(), "primary frames") {
		t.Fatalf("error = %v, want the primary-frame bound", err)
	}
	if got := len(tr.sentBytes()); got != 1 {
		t.Fatalf("reset sends = %d, want 1", got)
	}
}

// TestLinkHandshakeMisaddressedPrimaryIsNotSkipped proves a primary frame
// from the wrong outstation still fails the handshake immediately.
func TestLinkHandshakeMisaddressedPrimaryIsNotSkipped(t *testing.T) {
	bad := encodePeerNullUnsolicited(t, vecMasterAddr, 0x0999)
	tr := &scriptedTransport{resps: [][]byte{
		bad,
		encodeSecondaryHandshake(t, frame.FuncAck, vecMasterAddr, vecOutstationID),
		encodeSecondaryHandshake(t, frame.FuncLinkStatus, vecMasterAddr, vecOutstationID),
	}}
	m := newVecMaster(t, tr)
	err := m.Connect()
	if err == nil {
		t.Fatal("Connect succeeded after a misaddressed primary frame")
	}
	if !strings.Contains(err.Error(), "PRM=1") {
		t.Fatalf("error = %v, want the secondary PRM check", err)
	}
	if got := len(tr.sentBytes()); got != 1 {
		t.Fatalf("reset sends = %d, want 1", got)
	}
}

// TestLinkHandshakeBadACKAfterPrimaryStillFails proves a real secondary ACK
// with the wrong destination is not accepted just because a primary frame
// preceded it, and that no unsolicited confirm is sent.
func TestLinkHandshakeBadACKAfterPrimaryStillFails(t *testing.T) {
	primary := encodePeerNullUnsolicited(t, vecMasterAddr, vecOutstationID)
	badACK := encodeSecondaryHandshake(t, frame.FuncAck, 0x0200, vecOutstationID)
	tr := &scriptedTransport{resps: [][]byte{primary, badACK}}
	m := newVecMaster(t, tr)
	err := m.Connect()
	if err == nil {
		t.Fatal("Connect succeeded with a misaddressed ACK")
	}
	if !strings.Contains(err.Error(), "destination address") {
		t.Fatalf("error = %v, want destination address", err)
	}
	if got := len(tr.sentBytes()); got != 1 {
		t.Fatalf("sent %d frames, want only the reset", got)
	}
}
