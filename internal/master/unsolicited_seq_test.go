package master

import (
	"testing"

	"dnp3/internal/al"
)

func TestUnsolicitedDuplicateAndGap(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 1000})
	tr := &seqRecorderTransport{}
	m.SetTransport(tr)
	var got int
	m.SetUnsolicitedCallback(func([]byte) { got++ })

	app := al.NewUnsolicited(3, []byte{0x00, 0x00}).Encode()
	if !m.consumeUnsolicited(app, 1024) || got != 1 {
		t.Fatalf("first delivery got=%d", got)
	}
	if !m.consumeUnsolicited(app, 1024) || got != 1 {
		t.Fatalf("duplicate delivered, got=%d sends=%d", got, len(tr.sent))
	}
	gap := al.NewUnsolicited(5, []byte{0x00, 0x00}).Encode()
	if !m.consumeUnsolicited(gap, 1024) || got != 1 {
		t.Fatalf("gap changed delivery, got=%d", got)
	}
	next := al.NewUnsolicited(4, []byte{0x00, 0x00}).Encode()
	if !m.consumeUnsolicited(next, 1024) || got != 2 {
		t.Fatalf("in-order follow-up got=%d", got)
	}
	if len(tr.sent) < 2 {
		t.Fatalf("confirms sent = %d", len(tr.sent))
	}
}
