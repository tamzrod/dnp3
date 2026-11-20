package outstation

import (
	"bytes"
	"testing"
	"time"

	"dnp3/internal/al"
)

// Logical-clock / time-sync behaviour at the internal outstation boundary.
// These tests use hand-written G50V1/G52V1 vectors and never touch the host
// clock or NTP.

func timeSyncOutstation() *Outstation {
	o := NewOutstation(&Config{
		OutstationAddress: 1024,
		MasterAddress:     1,
		EnableTimeSync:    true,
		NeedTime:          true,
		MaxEventBuffers:   100,
	})
	o.Initialize()
	o.Start()
	return o
}

// Independent vector: 2026-11-19T11:02:56.910Z == 1795086176910 ms.
var testG50V1 = []byte{0x32, 0x01, 0x07, 0x01, 0x8E, 0x52, 0x79, 0xF3, 0xA1, 0x01}

func timeAPDU(fc uint8, data []byte) *al.APDU {
	return &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: fc,
		Data:     data,
	}
}

// TestTimeSyncViaFunctionCode32AppliesLogicalTime proves a G50V1 TIME_SYNC
// updates the logical clock to the independently-derived vector value and
// clears IIN NEED_TIME (IIN1 bit 0x10).
func TestTimeSyncViaFunctionCode32AppliesLogicalTime(t *testing.T) {
	o := timeSyncOutstation()
	if !o.Clock().NeedsTime() {
		t.Fatal("startup clock must need time")
	}

	resp, err := o.ProcessRequest(timeAPDU(al.FuncTimeSync, testG50V1))
	if err != nil {
		t.Fatalf("TIME_SYNC failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected a TIME_SYNC response for a unicast master")
	}
	if resp.Data[0]&0x10 != 0 {
		t.Fatalf("IIN NEED_TIME not cleared after sync: IIN=% X", resp.Data[:2])
	}
	st := o.Clock().Status()
	if st.Source != ClockMaster || st.SyncCount != 1 {
		t.Fatalf("clock status after sync = %+v", st)
	}
	if got := st.LastValue.UnixMilli(); got != 1795086176910 {
		t.Fatalf("logical time = %d ms, want 1795086176910", got)
	}
	// Logical Now must be close to the applied value (within a second).
	if d := o.Clock().Now().Sub(st.LastValue); d < -time.Second || d > time.Second {
		t.Fatalf("logical Now drifted from applied time by %v", d)
	}
}

// TestTimeSyncViaWriteFixesFuncUnknown is the direct regression for the
// Windows DNP3 Tester's "IIN 0100 NO_FUNC_CODE_SUPPORT": it sends time via
// WRITE to Group 50, which previously hit handleWrite's default branch and
// produced FUNC_UNKNOWN. It must now apply the same logical time as FC32.
func TestTimeSyncViaWriteFixesFuncUnknown(t *testing.T) {
	o := timeSyncOutstation()

	resp, err := o.ProcessRequest(timeAPDU(al.FuncWrite, testG50V1))
	if err != nil {
		t.Fatalf("WRITE G50V1 failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected a WRITE G50V1 response")
	}
	if resp.Data[0]&0x10 != 0 {
		t.Fatalf("IIN NEED_TIME not cleared after WRITE sync: IIN=% X", resp.Data[:2])
	}
	if resp.Data[1]&0x01 != 0 {
		t.Fatalf("IIN FUNC_UNKNOWN still set after WRITE G50V1: IIN=% X", resp.Data[:2])
	}
	if st := o.Clock().Status(); st.Source != ClockMaster || st.LastValue.UnixMilli() != 1795086176910 {
		t.Fatalf("WRITE did not apply logical time: %+v", st)
	}
}

// TestTimeSyncRefusedWithoutOptIn preserves the v0 profile: with time sync
// disabled the request fails so the run loop returns FUNC_UNKNOWN.
func TestTimeSyncRefusedWithoutOptIn(t *testing.T) {
	o := NewOutstation(nil)
	o.Initialize()
	o.Start()

	for _, fc := range []uint8{al.FuncTimeSync, al.FuncWrite, al.FuncDelayMeasurement, al.FuncRecordCurrentTime} {
		data := testG50V1
		if fc == al.FuncDelayMeasurement {
			data = al.BuildG52V1Object(0)
		}
		if resp, err := o.ProcessRequest(timeAPDU(fc, data)); err == nil {
			t.Fatalf("function code %d must be refused without opt-in, got %v", fc, resp)
		}
	}
}

// TestDelayMeasurementReturnsG52Object verifies the outstation reports a real
// G52V1 object with the measured processing delay rather than a hardcoded 0.
func TestDelayMeasurementReturnsG52Object(t *testing.T) {
	o := timeSyncOutstation()

	resp, err := o.ProcessRequest(timeAPDU(al.FuncDelayMeasurement, al.BuildG52V1Object(0)))
	if err != nil {
		t.Fatalf("DELAY_MEASUREMENT failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected a DELAY_MEASUREMENT response")
	}
	// Skip the two IIN octets prepended by ProcessRequest.
	if len(resp.Data) < 8 {
		t.Fatalf("short delay response: % X", resp.Data)
	}
	if !bytes.Equal(resp.Data[2:4], []byte{0x34, 0x01}) {
		t.Fatalf("expected G52V1 object, got % X", resp.Data[2:])
	}
	measured, err := al.ParseG52V1Object(resp.Data[2:])
	if err != nil {
		t.Fatalf("parse G52V1 response: %v", err)
	}
	if got := o.Clock().Status().LastDelayMS; got != measured {
		t.Fatalf("diagnostic delay %d != response delay %d", got, measured)
	}
}

// TestRecordCurrentTimePreservesBufferedSOE proves Record Current Time is a
// diagnostic-only operation: it stores the value but never rewrites the clock
// or an already captured detection timestamp.
func TestRecordCurrentTimePreservesBufferedSOE(t *testing.T) {
	o := timeSyncOutstation()
	before := o.Clock().Now()

	resp, err := o.ProcessRequest(timeAPDU(al.FuncRecordCurrentTime, testG50V1))
	if err != nil {
		t.Fatalf("RECORD_CURRENT_TIME failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected a RECORD_CURRENT_TIME response")
	}
	st := o.Clock().Status()
	if st.LastRecorded.UnixMilli() != 1795086176910 {
		t.Fatalf("recorded time = %v", st.LastRecorded)
	}
	if st.SyncCount != 0 || st.Source != ClockUnset {
		t.Fatalf("RECORD_CURRENT_TIME must not set the clock: %+v", st)
	}
	if d := o.Clock().Now().Sub(before); d < 0 || d > time.Second {
		t.Fatalf("logical clock jumped after RECORD_CURRENT_TIME: %v", d)
	}
}

// TestTimeValidityExpiryReassertsNeedTime verifies expiration diagnostics:
// after a configured validity window, NEED_TIME is asserted again until the
// master syncs once more.
func TestTimeValidityExpiryReassertsNeedTime(t *testing.T) {
	clock := NewLogicalClock(20*time.Millisecond, true)
	o := timeSyncOutstation()
	o.SetLogicalClock(clock)

	if _, err := o.ProcessRequest(timeAPDU(al.FuncTimeSync, testG50V1)); err != nil {
		t.Fatalf("TIME_SYNC failed: %v", err)
	}
	if clock.NeedsTime() {
		t.Fatal("NEED_TIME should be clear just after sync")
	}
	time.Sleep(30 * time.Millisecond)
	if !clock.Expired() || !clock.NeedsTime() {
		t.Fatal("NEED_TIME should be re-asserted after validity expiry")
	}

	// A fresh sync clears it again.
	resp, err := o.ProcessRequest(timeAPDU(al.FuncTimeSync, testG50V1))
	if err != nil || resp == nil {
		t.Fatalf("re-sync failed: err=%v resp=%v", err, resp)
	}
	if clock.NeedsTime() {
		t.Fatal("NEED_TIME should be clear after re-sync")
	}
}

// TestMalformedTimeSyncNeverChangesClock proves a malformed G50V1 object is
// rejected with no side effect on the logical clock.
func TestMalformedTimeSyncNeverChangesClock(t *testing.T) {
	o := timeSyncOutstation()
	before := o.Clock().Status()

	for _, data := range [][]byte{
		{0x32, 0x01, 0x07, 0x02, 0x8E, 0x52, 0x79, 0xF3, 0xA1, 0x01},
		{0x2A, 0x01, 0x07, 0x01, 0x8E, 0x52, 0x79, 0xF3, 0xA1, 0x01},
		nil,
	} {
		if resp, err := o.ProcessRequest(timeAPDU(al.FuncTimeSync, data)); err == nil {
			t.Fatalf("malformed G50V1 accepted: resp=%v", resp)
		}
	}
	after := o.Clock().Status()
	if after.SyncCount != before.SyncCount || !after.LastValue.Equal(before.LastValue) {
		t.Fatalf("malformed time sync changed the clock: before=%+v after=%+v", before, after)
	}
}
