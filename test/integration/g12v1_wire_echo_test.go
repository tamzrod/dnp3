package integration

import (
	"context"
	"encoding/hex"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/outstation"
	"dnp3/pkg/dnp3/types"
)

// exact wire bytes captured passively from the Windows DNP3 Tester on
// 2026-11-19 (evidence/protocols/dnp3-crob-interop in the EG628 workspace).
// Master 1 -> outstation 1024, FC=05 DIRECT OPERATE, G12V1 qualifier 0x17,
// count=1, index=2, Latch Off, on=1000 ms, off=0, status=0. Header CRC and
// the two 16-octet data-block CRCs are included, as they are on the wire.
const capturedWindowsCROBDirectOperateFrame = "056418c4000401005b61d5c6050c011701020401e8030000000042c6000000ffff"

type g12WireCommandHandler struct {
	calls []*types.ControlOutput
}

func (h *g12WireCommandHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	h.calls = append(h.calls, cmd)
	s := types.ControlSuccess
	return &s, nil
}

func (h *g12WireCommandHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlNotSupported
	return &s, nil
}

// TestCapturedWindowsCROBFrameReturnsStatusEcho replays the exact request that
// previously produced an IIN-only response with IIN2.1 FUNC_UNKNOWN (which the
// Windows tester reports as "the outstation returned no command statuses").
// The corrected outstation must dispatch the command and echo G12V1 with a
// trailing SUCCESS status octet. The test is entirely in-process TCP; no
// physical output is written.
func TestCapturedWindowsCROBFrameReturnsStatusEcho(t *testing.T) {
	port := getFreePort(t)
	cmd := &g12WireCommandHandler{}
	server, err := outstation.NewServer(outstation.NewConfig(
		outstation.WithAddress(1024),
		outstation.WithMasterAddress(1),
		outstation.WithTransport(dnp3.TCP, "localhost", port),
	))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	server.SetCommandHandler(cmd)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := server.Start(ctx); err != nil {
		t.Fatalf("server.Start: %v", err)
	}
	defer server.Stop(context.Background())

	conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", strconv.Itoa(port)), 2*time.Second)
	if err != nil {
		t.Fatalf("dial outstation: %v", err)
	}
	defer conn.Close()

	request, err := hex.DecodeString(capturedWindowsCROBDirectOperateFrame)
	if err != nil {
		t.Fatalf("decode captured frame: %v", err)
	}
	if _, err := conn.Write(request); err != nil {
		t.Fatalf("send captured frame: %v", err)
	}

	response, err := readOneDNP3Frame(conn)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	t.Logf("captured request = %s", strings.ToUpper(hex.EncodeToString(request)))
	t.Logf("response frame  = %s", strings.ToUpper(hex.EncodeToString(response)))

	if len(cmd.calls) != 1 {
		t.Fatalf("command handler calls = %d, want 1", len(cmd.calls))
	}
	if cmd.calls[0].Index != 2 {
		t.Fatalf("dispatched index = %d, want 2", cmd.calls[0].Index)
	}

	// APPLICATION: C0 (transport) C6 (AC) 81 (RESPONSE) 0000 (IIN)
	// then the echoed G12V1 object. The last octet is command status SUCCESS.
	wantObject := "0C011701020401E803000000000000"
	app := applicationData(t, response)
	if !strings.Contains(strings.ToUpper(hex.EncodeToString(app)), wantObject) {
		t.Fatalf("response application data does not contain G12V1 status echo %s: %s",
			wantObject, strings.ToUpper(hex.EncodeToString(app)))
	}
}

// capturedWindowsCROBLatchOnFrame is the exact Direct Operate frame from the
// post-deploy Windows tester retry (t=1795068514.403978): index 1, Latch On
// (code 0x03) carrying the tester's default OnTime=1000 ms. It previously
// produced a G12V1 echo with command status 0x04 (NOT_SUPPORTED).
const capturedWindowsCROBLatchOnFrame = "056418c4000401005b61e7c8050c011701010301e80300000000e232000000ffff"

// v0GuardWireCommandHandler models the EG628 command policy: a single latch
// write is accepted, but pulse repetition and non-zero on/off times are
// refused with a definite NOT_SUPPORTED status.
type v0GuardWireCommandHandler struct {
	calls   int
	onTime  uint32
	offTime uint32
}

func (h *v0GuardWireCommandHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	h.calls++
	h.onTime, h.offTime = cmd.OnTime, cmd.OffTime
	if cmd.Count > 1 || cmd.OnTime != 0 || cmd.OffTime != 0 {
		s := types.ControlNotSupported
		return &s, nil
	}
	s := types.ControlSuccess
	return &s, nil
}

func (h *v0GuardWireCommandHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlNotSupported
	return &s, nil
}

// TestCapturedLatchOnWithDefaultOnTimeEchoesSuccess replays the captured
// Latch On frame end to end over loopback TCP. A master's non-zero default
// OnTime is not applicable to a latch code and must be normalised by the
// bridge, so the v0-guard handler accepts the command and the outstation
// echoes G12V1 with status 0x00 instead of 0x04. No physical output is touched.
func TestCapturedLatchOnWithDefaultOnTimeEchoesSuccess(t *testing.T) {
	port := getFreePort(t)
	cmd := &v0GuardWireCommandHandler{}
	server, err := outstation.NewServer(outstation.NewConfig(
		outstation.WithAddress(1024),
		outstation.WithMasterAddress(1),
		outstation.WithTransport(dnp3.TCP, "localhost", port),
	))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	server.SetCommandHandler(cmd)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := server.Start(ctx); err != nil {
		t.Fatalf("server.Start: %v", err)
	}
	defer server.Stop(context.Background())

	conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", strconv.Itoa(port)), 2*time.Second)
	if err != nil {
		t.Fatalf("dial outstation: %v", err)
	}
	defer conn.Close()

	request, err := hex.DecodeString(capturedWindowsCROBLatchOnFrame)
	if err != nil {
		t.Fatalf("decode captured frame: %v", err)
	}
	if _, err := conn.Write(request); err != nil {
		t.Fatalf("send captured frame: %v", err)
	}
	response, err := readOneDNP3Frame(conn)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	if cmd.calls != 1 {
		t.Fatalf("command handler calls = %d, want 1", cmd.calls)
	}
	if cmd.onTime != 0 || cmd.offTime != 0 {
		t.Fatalf("handler saw times (%d,%d), want (0,0)", cmd.onTime, cmd.offTime)
	}

	// group/var/qual/count/index + CROB fields + trailing status SUCCESS.
	wantObject := "0C01170101" + "0301E8030000" + "00000000" + "00"
	app := applicationData(t, response)
	if !strings.Contains(strings.ToUpper(hex.EncodeToString(app)), wantObject) {
		t.Fatalf("response application data does not contain G12V1 SUCCESS echo %s: %s",
			wantObject, strings.ToUpper(hex.EncodeToString(app)))
	}
}

// applicationData strips the header and the per-16-octet data-block CRCs so
// the application fragment can be compared with the expected object bytes.
func applicationData(t *testing.T, frame []byte) []byte {
	t.Helper()
	dataLen := int(frame[2]) - 5
	app := make([]byte, 0, dataLen)
	for off, remaining := 10, dataLen; remaining > 0; {
		n := min(16, remaining)
		app = append(app, frame[off:off+n]...)
		off += n + 2
		remaining -= n
	}
	return app
}

// readOneDNP3Frame reads exactly one self-delimiting DNP3 link frame,
// accounting for one CRC per 16-octet data block.
func readOneDNP3Frame(conn net.Conn) ([]byte, error) {
	// A malformed request is silently ignored by the outstation; bound the
	// wait so a bad test vector fails fast instead of hanging the suite.
	if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return nil, err
	}
	header := make([]byte, 10)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}
	if header[0] != 0x05 || header[1] != 0x64 {
		return nil, io.ErrUnexpectedEOF
	}
	dataLen := int(header[2]) - 5
	if dataLen < 0 {
		return nil, io.ErrUnexpectedEOF
	}
	blocks := (dataLen + 15) / 16
	frame := make([]byte, 8+2+blocks*2+dataLen)
	copy(frame, header)
	if _, err := io.ReadFull(conn, frame[10:]); err != nil {
		return nil, err
	}
	return frame, nil
}
