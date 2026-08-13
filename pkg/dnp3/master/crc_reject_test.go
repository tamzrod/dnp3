package master

import (
	"context"
	"strings"
	"testing"
	"time"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
	"dnp3/pkg/dnp3/types"
)

// badCRCTransport returns a single DLL frame whose header CRC has been
// corrupted, so the master receive path must reject it (DNP3-026).
type badCRCTransport struct{ sent bool }

func (t *badCRCTransport) Send(data []byte) error { t.sent = true; return nil }
func (t *badCRCTransport) SetTimeout(ms int)      {}

func (t *badCRCTransport) Receive() ([]byte, error) {
	// Build a valid response frame, then corrupt the header CRC bytes (the
	// two bytes immediately following the 8-byte header prefix).
	good := buildPubReadResponse(0)
	if len(good) >= 11 {
		good[10] ^= 0xFF
	}
	return good, nil
}

// TestReadRejectsInvalidCRC asserts the master receive path surfaces a CRC
// validation failure as an error and returns no points (DNP3-026).
func TestReadRejectsInvalidCRC(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &badCRCTransport{})

	resp, err := cc.Read(context.Background(), &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
	})
	if err == nil {
		t.Fatalf("expected error for invalid CRC frame, got nil")
	}
	if resp != nil {
		t.Fatalf("expected nil response on CRC failure; got %+v", resp)
	}
}

// TestReadRejectsInvalidCRCIsNoPartial re-asserts that no partial points reach
// the caller on a CRC failure (the response is nil).
func TestReadRejectsInvalidCRCIsNoPartial(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &badCRCTransport{})
	resp, err := cc.Read(context.Background(), &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
	})
	if err == nil {
		t.Fatal("expected error for invalid CRC frame")
	}
	if resp != nil {
		t.Fatalf("expected nil response (no partial points); got %+v", resp)
	}
}

// TestReadBadCRCNoHangDeadline asserts the master receive path returns a
// bounded CRC error within an explicit deadline (no deadlock) when the peer
// keeps emitting bad-CRC frames (MEXT-026). The transport always returns a
// header-CRC-corrupted frame; the master must reject it (not block).
func TestReadBadCRCNoHangDeadline(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &badCRCTransport{})

	done := make(chan error, 1)
	go func() {
		_, err := cc.Read(context.Background(), &types.ReadRequest{
			Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
		})
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected a bounded CRC error, got nil")
		}
		if !strings.Contains(err.Error(), "CRC") {
			t.Fatalf("error did not name the CRC failure: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Read deadlocked on a bad-CRC frame (no bounded error within 5s)")
	}
}

// TestALRejectsInvalidCRC is a direct unit-level check that frame.Decode fails
// on a corrupted header CRC and a corrupted data-block CRC, so no points can be
// parsed from a bad-CRC frame.
func TestALRejectsInvalidCRC(t *testing.T) {
	t.Run("header_crc", func(t *testing.T) {
		good := buildPubReadResponse(0)
		bad := make([]byte, len(good))
		copy(bad, good)
		// Header CRC sits at offset 8 (2 bytes) after the 8-byte header prefix.
		bad[8] ^= 0xFF
		_, err := frame.Decode(bad)
		if err == nil {
			t.Fatal("expected header CRC validation error")
		}
		if !strings.Contains(err.Error(), "CRC") {
			t.Fatalf("expected CRC error, got: %v", err)
		}
	})

	t.Run("data_crc", func(t *testing.T) {
		apdu := &al.APDU{
			Control:  al.AppControl{FIR: true, FIN: true, Seq: 0},
			FuncCode: al.FuncResponse,
			Data:     make([]byte, 16), // one full 16-byte block
		}
		frag := tl.Fragment{FIR: true, FIN: true, Data: apdu.Encode()}
		tlData := tl.EncodeFragment(frag)
		dllFrame := &frame.Frame{
			Control:  frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncConfirmedUserDataR},
			DestAddr: 1, SrcAddr: 2, Data: tlData,
		}
		good, _ := frame.Encode(dllFrame)
		// Layout: 8 header + 2 headerCRC + 16 data + 2 dataCRC = 28; data CRC
		// occupies the last 2 bytes.
		bad := make([]byte, len(good))
		copy(bad, good)
		bad[len(bad)-1] ^= 0xFF
		_, err := frame.Decode(bad)
		if err == nil {
			t.Fatal("expected data CRC validation error")
		}
		if !strings.Contains(err.Error(), "CRC") {
			t.Fatalf("expected CRC error, got: %v", err)
		}
	})
}

