package outstation

import (
	"bytes"
	"testing"

	"dnp3/internal/al"
)

// Group 21 is "Frozen Counter" — a static snapshot group — not a counter
// event group. These tests pin that a G21 READ returns the frozen snapshot
// captured by FREEZE, in the IEEE 1815 flags-first little-endian layout, and
// never the live counter values.

// frozenCounterHandler keeps distinct live and frozen slices so a regression
// to reading live counters is immediately visible.
type frozenCounterHandler struct {
	DefaultDataHandler
	live   []Counter
	frozen []Counter
}

func (h *frozenCounterHandler) GetCounters() []Counter       { return h.live }
func (h *frozenCounterHandler) GetFrozenCounters() []Counter { return h.frozen }
func (h *frozenCounterHandler) FreezeCounters(clear bool) error {
	h.frozen = append([]Counter(nil), h.live...)
	if clear {
		for i := range h.live {
			h.live[i].Value = 0
		}
	}
	return nil
}

func readGroup21Request(variation uint8) *al.APDU {
	return &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncRead,
		// 16-bit range 0..0 for group 21, requested variation.
		Data: []byte{21, variation, al.QualRange16, 0, 0, 0, 0},
	}
}

func newFrozenCounterOutstation(h *frozenCounterHandler) *Outstation {
	o := NewOutstation(nil)
	o.Initialize()
	o.Start()
	o.SetDataHandler(h)
	return o
}

// TestFrozenCounterReadUsesSnapshotNotLive verifies a G21V1 READ returns the
// frozen values after FREEZE_CLEAR, with an independent byte-exact wire vector:
// header 15 01 07 02 then flags+LE value for each point.
func TestFrozenCounterReadUsesSnapshotNotLive(t *testing.T) {
	h := &frozenCounterHandler{
		live: []Counter{
			{Value: 0x01020304, Quality: CounterQualityOnline},
			{Value: 0xAABBCCDD, Quality: CounterQualityOnline | CounterQualityRollOver},
		},
	}
	o := newFrozenCounterOutstation(h)

	if _, err := o.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncFreezeClear,
	}); err != nil {
		t.Fatalf("FREEZE_CLEAR failed: %v", err)
	}

	resp, err := o.ProcessRequest(readGroup21Request(1))
	if err != nil {
		t.Fatalf("G21 READ failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected a G21 READ response")
	}

	// IIN(2) + independent G21V1 vector: flags 0x01, value 0x01020304 LE;
	// flags 0x09 (online|rollover), value 0xAABBCCDD LE.
	want := []byte{
		0x00, 0x00, // IIN
		0x15, 0x01, 0x07, 0x02, // Group 21 V1, 8-bit count, 2 points
		0x01, 0x04, 0x03, 0x02, 0x01,
		0x09, 0xDD, 0xCC, 0xBB, 0xAA,
	}
	if !bytes.Equal(resp.Data[2:], want[2:]) {
		t.Fatalf("G21V1 wire mismatch\n got: % X\nwant: % X", resp.Data[2:], want[2:])
	}

	// Live counters were cleared; the frozen read must still carry old values.
	if h.live[0].Value != 0 || h.live[1].Value != 0 {
		t.Fatalf("live counters not cleared: %+v", h.live)
	}
}

// TestFrozenCounterReadIgnoresLaterLiveChanges verifies the snapshot is stable
// between freezes: changing the live counter after FREEZE must not affect a
// subsequent G21 READ (the old live-counter bug would fail here).
func TestFrozenCounterReadIgnoresLaterLiveChanges(t *testing.T) {
	h := &frozenCounterHandler{
		live: []Counter{{Value: 100, Quality: CounterQualityOnline}},
	}
	o := newFrozenCounterOutstation(h)

	if _, err := o.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncFreeze,
	}); err != nil {
		t.Fatalf("FREEZE failed: %v", err)
	}
	h.live[0].Value = 999 // change after the freeze

	resp, err := o.ProcessRequest(readGroup21Request(1))
	if err != nil || resp == nil {
		t.Fatalf("G21 READ failed: err=%v resp=%v", err, resp)
	}
	// flags 0x01 + LE 100 (0x64)
	want := []byte{0x15, 0x01, 0x07, 0x01, 0x01, 0x64, 0x00, 0x00, 0x00}
	if !bytes.Equal(resp.Data[2:], want) {
		t.Fatalf("G21V1 did not use the frozen snapshot\n got: % X\nwant: % X", resp.Data[2:], want)
	}
}

// TestFrozenCounterV5Carries48BitTime verifies the time-tagged frozen counter
// variation uses a 6-byte little-endian DNP time derived from the snapshot's
// stored Time field (ms since the 1970 IEEE 1815 epoch).
func TestFrozenCounterV5Carries48BitTime(t *testing.T) {
	// Independent vector: 2026-11-19T11:02:56.910Z == 1795086176910 ms,
	// LSB-first octets 8E 52 79 F3 A1 01.
	h := &frozenCounterHandler{
		frozen: []Counter{{Value: 7, Quality: CounterQualityOnline, Time: 1795086176910}},
	}
	o := newFrozenCounterOutstation(h)

	resp, err := o.ProcessRequest(readGroup21Request(5))
	if err != nil || resp == nil {
		t.Fatalf("G21V5 READ failed: err=%v resp=%v", err, resp)
	}
	want := []byte{
		0x15, 0x05, 0x07, 0x01,
		0x01, 0x07, 0x00, 0x00, 0x00, // flags + LE value 7
		0x8E, 0x52, 0x79, 0xF3, 0xA1, 0x01, // 48-bit DNP time
	}
	if !bytes.Equal(resp.Data[2:], want) {
		t.Fatalf("G21V5 wire mismatch\n got: % X\nwant: % X", resp.Data[2:], want)
	}
}
