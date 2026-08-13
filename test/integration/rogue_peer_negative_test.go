package integration

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/master"
	"dnp3/pkg/dnp3/types"
)

// MEXT-026 — Negative: bad CRC / wrong address must not hang (no deadlock).
//
// The master receive path must reject a corrupted-CRC frame or a link-layer
// frame addressed to a different master with a BOUNDED error — it must never
// deadlock. These cases run against a RAW TCP "rogue peer" (not the in-repo
// outstation, which always emits well-formed, correctly-addressed frames) so
// the malformed/wrong-address frames are the only thing the master sees on the
// wire.

// readLinkFrame reads exactly one DNP3 link frame from r by syncing on the
// 0x05 0x64 start bytes and using the length byte to size the read. It returns
// the raw frame bytes (sync..final CRC).
func readLinkFrame(r io.Reader) ([]byte, error) {
	// Sync on 0x05 0x64.
	var hdr [3]byte // sync1, sync2, length
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, err
	}
	for !(hdr[0] == frame.SyncByte1 && hdr[1] == frame.SyncByte2) {
		// Shift left and read one more byte until we find the sync pair.
		hdr[0], hdr[1] = hdr[1], hdr[2]
		if _, err := io.ReadFull(r, hdr[2:]); err != nil {
			return nil, err
		}
	}
	lengthByte := int(hdr[2])
	dataLen := lengthByte - 5
	if dataLen < 0 {
		dataLen = 0
	}
	crcBytes := 2 + ((dataLen+15)/16)*2
	restSize := 5 + dataLen + crcBytes // control(1)+dest(2)+src(2) + data + crcs
	rest := make([]byte, restSize)
	if _, err := io.ReadFull(r, rest); err != nil {
		return nil, err
	}
	out := make([]byte, 3+restSize)
	out[0], out[1], out[2] = hdr[0], hdr[1], hdr[2]
	copy(out[3:], rest)
	return out, nil
}

// encodeLinkFrame encodes a no-data secondary link frame (ACK / Link Status)
// addressed from src to dest.
func encodeLinkFrame(funcCode uint8, destAddr, srcAddr uint16) ([]byte, error) {
	f := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: funcCode},
		DestAddr: destAddr,
		SrcAddr:  srcAddr,
		Data:     nil,
	}
	return frame.Encode(f)
}

// encodeReadResponse builds a well-formed DLL user-data response frame carrying
// a minimal AL Read-Response (FIR/FIN, FuncResponse, clean IIN, no objects),
// addressed from the outstation to the master.
func encodeReadResponse(masterAddr, outstationAddr uint16, seq uint8) ([]byte, error) {
	iin := al.IIN{}
	iinBytes := iin.Bytes()
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: seq},
		FuncCode: al.FuncResponse,
		Data:     iinBytes[:],
	}
	frag := tl.Fragment{FIR: true, FIN: true, Data: apdu.Encode()}
	tlData := tl.EncodeFragment(frag)
	f := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncConfirmedUserDataR},
		DestAddr: masterAddr,
		SrcAddr:  outstationAddr,
		Data:     tlData,
	}
	return frame.Encode(f)
}

// startRoguePeer listens on a free port, accepts one master connection, and
// hands the accepted conn + listener address to the handler. The handler runs
// in a goroutine and speaks raw DNP3 link frames to/with the master.
func startRoguePeer(t *testing.T, handle func(net.Conn)) (port int) {
	t.Helper()
	port = getFreePort(t)
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("rogue listen: %v", err)
	}
	go func() {
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		handle(conn)
	}()
	return port
}

// rogueLoopACK handles a master that may retry the reset-link exchange: it
// reads reset-link frames in a loop and re-ACKs each one with an ACK addressed
// to wrongMaster (instead of the real master address). The master therefore
// exhausts its retry budget on the address mismatch and surfaces a bounded
// "invalid reset link ACK: ... destination address" error rather than
// silently accepting the wrong-address frame (MEXT-026).
func rogueLoopACK(t *testing.T, conn net.Conn, wrongMaster, outstationAddr uint16) {
	for {
		f, err := readLinkFrame(conn)
		if err != nil {
			return
		}
		_ = f
		ack, err := encodeLinkFrame(frame.FuncAck, wrongMaster, outstationAddr)
		if err != nil {
			return
		}
		if _, err := conn.Write(ack); err != nil {
			return
		}
	}
}

// TestRoguePeerWrongAddressNoHang asserts a link ACK addressed to a DIFFERENT
// master is rejected with a bounded error and Connect returns promptly (no
// deadlock) (MEXT-026). The master detects the address mismatch on each retry
// and exhausts its retry budget on it, so the surfaced error names the
// address violation (not a hang).
func TestRoguePeerWrongAddressNoHang(t *testing.T) {
	const (
		masterAddr     = uint16(0xFFFF)
		outstationAddr = uint16(1024)
		wrongMaster    = uint16(0x0001) // ACK dest != our master address
	)
	port := startRoguePeer(t, func(conn net.Conn) {
		rogueLoopACK(t, conn, wrongMaster, outstationAddr)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client, err := master.NewClient(master.NewConfig(
		master.WithMasterAddress(masterAddr),
		master.WithOutstationAddress(outstationAddr),
		master.WithTransport(dnp3.TCP, "localhost", port),
		master.WithTimeout(1500*time.Millisecond),
		master.WithRetry(0, 0),
	))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()

	done := make(chan error, 1)
	go func() { done <- client.Connect(ctx) }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("Connect unexpectedly succeeded against a wrong-address ACK (must be rejected)")
		}
		// No-deadlock guarantee: Connect returned a bounded error. The error
		// must name the address violation the master detected (not a hang).
		t.Logf("wrong-address Connect error (bounded, no hang): %v", err)
		if !strings.Contains(err.Error(), "destination address") &&
			!strings.Contains(err.Error(), "address") {
			t.Fatalf("Connect error did not name the address mismatch: %v", err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("Connect deadlocked on a wrong-address ACK (no bounded error within 8s)")
	}
}

// rogueBadCRCHandler completes a valid link handshake, then for every master
// read-request frame replies with a corrupted-CRC user-data response. The
// master exhausts its CRC-retry budget on the bad frame and surfaces a bounded
// CRC error rather than hanging (MEXT-026).
func rogueBadCRCHandler(t *testing.T, conn net.Conn, masterAddr, outstationAddr uint16) {
	// 1) Reset Link Stations → valid ACK (loop-tolerant for retries).
	if _, err := readLinkFrame(conn); err != nil {
		return
	}
	ack, _ := encodeLinkFrame(frame.FuncAck, masterAddr, outstationAddr)
	if _, err := conn.Write(ack); err != nil {
		return
	}
	// 2) Request Link Status → valid Link Status (loop-tolerant).
	if _, err := readLinkFrame(conn); err != nil {
		return
	}
	ls, _ := encodeLinkFrame(frame.FuncLinkStatus, masterAddr, outstationAddr)
	if _, err := conn.Write(ls); err != nil {
		return
	}
	// 3) For each subsequent master frame (read requests / retries), reply
	//    with a BAD-CRC user-data response.
	for {
		if _, err := readLinkFrame(conn); err != nil {
			return
		}
		good, err := encodeReadResponse(masterAddr, outstationAddr, 0)
		if err != nil {
			return
		}
		if len(good) >= 11 {
			good[10] ^= 0xFF // corrupt a header-CRC byte
		}
		if _, err := conn.Write(good); err != nil {
			return
		}
	}
}

// TestRoguePeerBadCRCNoHang asserts a corrupted-CRC response frame is rejected
// with a bounded CRC error and Read returns promptly (no deadlock) (MEXT-026).
func TestRoguePeerBadCRCNoHang(t *testing.T) {
	const (
		masterAddr     = uint16(0xFFFF)
		outstationAddr = uint16(1024)
	)
	port := startRoguePeer(t, func(conn net.Conn) {
		rogueBadCRCHandler(t, conn, masterAddr, outstationAddr)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client, err := master.NewClient(master.NewConfig(
		master.WithMasterAddress(masterAddr),
		master.WithOutstationAddress(outstationAddr),
		master.WithTransport(dnp3.TCP, "localhost", port),
		master.WithTimeout(1500*time.Millisecond),
		master.WithRetry(0, 0),
	))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect against rogue (valid handshake) failed: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		_, e := client.Read(ctx, types.NewReadRequest(types.GroupRequest{Group: 1, Variation: 0}))
		done <- e
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("Read unexpectedly succeeded against a bad-CRC frame (must be rejected)")
		}
		// No-deadlock guarantee: Read returned a bounded error naming the CRC
		// failure (not a hang).
		t.Logf("bad-CRC Read error (bounded, no hang): %v", err)
		if !strings.Contains(err.Error(), "CRC") {
			t.Fatalf("Read error did not name the CRC failure: %v", err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("Read deadlocked on a bad-CRC frame (no bounded error within 8s)")
	}
}
