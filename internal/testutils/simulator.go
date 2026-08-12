package testutils

import (
	"encoding/binary"
	"fmt"
	"math"
	"sync"
	"time"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
	"dnp3/pkg/dnp3/types"
)

// MVPOutstationSimulator is a deterministic, in-memory outstation for the v0
// MVP profile (DNP3-036). It implements the public transport.Handler interface
// so it can be injected into a public master client (via NewClientWithTransport)
// and driven through the full Connect → Read → Operate flow with NO network
// I/O and NO real outstation process.
//
// The simulator answers:
//   - link-layer handshake frames (Reset Link Stations → ACK; Request Link
//     Status → Link Status), so the master's Connect succeeds (DNP3-006/007);
//   - Class-0 Read requests (G1/G20/G30) with golden static data configured via
//     the Set* accessors;
//   - G12V1 DirectOperate (CROB) requests with a configurable per-point command
//     status (default ControlSuccess).
//
// All responses echo the request's application sequence (DNP3-010) and carry a
// clean IIN. The simulator is concurrency-safe.
type MVPOutstationSimulator struct {
	mu             sync.Mutex
	outstationAddr uint16
	masterAddr     uint16

	// Golden static data (MVP profile).
	binaryInputs []*types.BinaryInput
	analogInputs []*types.AnalogInput
	counters     []*types.Counter

	// Per-point command status returned for CROB DirectOperate (default 0 =
	// ControlSuccess). Configurable via SetCommandStatus.
	commandStatus byte

	// IIN returned in every response (default all-clear).
	iin [2]byte

	// DNP3-053 test hook: a one-shot IIN applied to the next application
	// response only, then reverted to the sticky iin above. Set with
	// SetNextResponseIIN (consumed by the first application-layer response —
	// link handshake responses do not touch it).
	nextIIN [2]byte
	hasNext bool

	// DNP3-053 test hook: ordered list of application Read group numbers the
	// simulator has handled, for asserting follow-on integrity polls.
	readGroups []uint8

	// Recorded frames the master sent (post-Decode summaries), for assertions.
	sent []*frame.Frame

	// Response queue: each Send enqueues at most one response frame; Receive
	// pops it. The master's handshake + request path is strictly ping-pong
	// (send one, receive one), so a depth-1 queue suffices.
	pending [][]byte

	closed bool
}

// NewMVPOutstationSimulator creates a simulator addressed as outstationAddr
// talking to a master at masterAddr. Golden data defaults to a single online
// binary input (true) so a bare Read of G1 returns one point out of the box.
func NewMVPOutstationSimulator(outstationAddr, masterAddr uint16) *MVPOutstationSimulator {
	return &MVPOutstationSimulator{
		outstationAddr: outstationAddr,
		masterAddr:     masterAddr,
		binaryInputs: []*types.BinaryInput{
			{Index: 0, Value: true, Quality: types.QualityOnline},
		},
		commandStatus: byte(types.ControlSuccess),
	}
}

// SetBinaryInputs configures the golden binary input data returned for G1.
func (s *MVPOutstationSimulator) SetBinaryInputs(points []*types.BinaryInput) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.binaryInputs = append([]*types.BinaryInput(nil), points...)
}

// SetAnalogInputs configures the golden analog input data returned for G30.
func (s *MVPOutstationSimulator) SetAnalogInputs(points []*types.AnalogInput) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.analogInputs = append([]*types.AnalogInput(nil), points...)
}

// SetCounters configures the golden counter data returned for G20.
func (s *MVPOutstationSimulator) SetCounters(points []*types.Counter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters = append([]*types.Counter(nil), points...)
}

// SetCommandStatus sets the per-point command status returned for CROB
// DirectOperate (default 0 = ControlSuccess).
func (s *MVPOutstationSimulator) SetCommandStatus(status types.ControlStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commandStatus = byte(status)
}

// SetIIN sets the IIN octets returned in every response (default all-clear).
func (s *MVPOutstationSimulator) SetIIN(iin [2]byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.iin = iin
}

// SetNextResponseIIN sets a one-shot IIN applied to the NEXT application-layer
// response only (DNP3-053 test hook). After that response the simulator reverts
// to the sticky IIN set by SetIIN. Link-layer handshake responses do not
// consume it, so the one-shot lands on the first application Read/Operate.
func (s *MVPOutstationSimulator) SetNextResponseIIN(iin [2]byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextIIN = iin
	s.hasNext = true
}

// ReadGroups returns the ordered list of application Read group numbers the
// simulator has handled (DNP3-053 test hook), for asserting that an integrity
// poll followed a trigger.
func (s *MVPOutstationSimulator) ReadGroups() []uint8 {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]uint8, len(s.readGroups))
	copy(out, s.readGroups)
	return out
}

// SentFrames returns a snapshot of the decoded frames the master sent, in send
// order, for test assertions.
func (s *MVPOutstationSimulator) SentFrames() []*frame.Frame {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*frame.Frame, len(s.sent))
	copy(out, s.sent)
	return out
}

// --- transport.Handler (public interface) ---

// Listen is a no-op: the in-memory simulator has no listener.
func (s *MVPOutstationSimulator) Listen() error { return nil }

// Connect establishes the (simulated) transport connection. It resets any prior
// closed state so the same simulator instance can serve a Close → Connect cycle
// — mirroring a real outstation that accepts a fresh link session after the
// master disconnects (DNP3-050 lifecycle testing).
func (s *MVPOutstationSimulator) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = false
	s.pending = s.pending[:0]
	return nil
}

// Accept is a no-op: the in-memory simulator has no incoming connections.
func (s *MVPOutstationSimulator) Accept() error { return nil }

// Close marks the simulator closed; further Send/Receive return ErrTransportClosed.
func (s *MVPOutstationSimulator) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return nil
}

// SetTimeout is a no-op: the simulator responds synchronously, so timeouts are
// not exercised.
func (s *MVPOutstationSimulator) SetTimeout(ms int) {}

// Send decodes the master's DLL frame and enqueues the deterministic response
// (link handshake ACK/LinkStatus, or an application Read/Operate response). The
// master drives a strict send-then-receive ping-pong, so exactly one response
// is queued per Send.
func (s *MVPOutstationSimulator) Send(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrTransportClosed
	}

	f, err := frame.Decode(data)
	if err != nil {
		return fmt.Errorf("simulator: decode master frame: %w", err)
	}
	s.sent = append(s.sent, f)

	var resp []byte
	switch {
	case f.Control.PRM && f.Control.FuncCode == frame.FuncResetLinkStations:
		resp = s.encodeLinkFrame(frame.FuncAck)
	case f.Control.PRM && f.Control.FuncCode == frame.FuncRequestLinkStatus:
		resp = s.encodeLinkFrame(frame.FuncLinkStatus)
	case f.Control.PRM && f.Control.FuncCode == frame.FuncConfirmedUserData:
		resp, err = s.handleApplication(f)
		if err != nil {
			return err
		}
	default:
		// Unknown primary frame: respond with a link NACK so the master's
		// handshake/validation surfaces the failure deterministically rather
		// than hanging.
		resp = s.encodeLinkFrame(frame.FuncNack)
	}

	s.pending = append(s.pending, resp)
	return nil
}

// Receive pops the next queued response frame. It blocks up to a short bounded
// poll (well under any realistic master timeout) so a misused Receive without
// a prior Send surfaces as a timeout rather than an indefinite hang.
func (s *MVPOutstationSimulator) Receive() ([]byte, error) {
	deadline := time.Now().Add(5 * time.Second)
	for {
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			return nil, ErrTransportClosed
		}
		if len(s.pending) > 0 {
			out := s.pending[0]
			s.pending = s.pending[1:]
			s.mu.Unlock()
			return out, nil
		}
		s.mu.Unlock()
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("simulator: receive timeout (no queued response)")
		}
		time.Sleep(time.Millisecond)
	}
}

// --- response builders (called with s.mu held) ---

// encodeLinkFrame builds a secondary (PRM=0) link frame addressed from the
// simulated outstation to the master with the given function code.
func (s *MVPOutstationSimulator) encodeLinkFrame(fc byte) []byte {
	f := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: fc},
		DestAddr: s.masterAddr,
		SrcAddr:  s.outstationAddr,
	}
	raw, _ := frame.Encode(f)
	return raw
}

// handleApplication decodes the application request inside a Confirmed User
// Data frame and returns the deterministic response frame (DNP3-036).
func (s *MVPOutstationSimulator) handleApplication(f *frame.Frame) ([]byte, error) {
	frag, err := tl.DecodeFragment(f.Data)
	if err != nil {
		return nil, fmt.Errorf("simulator: decode TL fragment: %w", err)
	}
	apdu, err := al.Decode(frag.Data)
	if err != nil {
		return nil, fmt.Errorf("simulator: decode APDU: %w", err)
	}
	seq := apdu.Control.Seq

	var appData []byte
	switch apdu.FuncCode {
	case al.FuncRead:
		appData = s.buildReadResponseData(apdu.Data)
	case al.FuncDirectOperate, al.FuncOperate, al.FuncSelect:
		appData = s.buildOperateResponseData(apdu.Data)
	default:
		// Unknown application function: return an empty response with IIN
		// indicating func-not-supported, so the master surfaces a clear error.
		appData = nil
	}

	resp := al.NewAppResponse(seq, al.IIN{}, appData)
	// Apply IIN: a one-shot IIN (DNP3-053 test hook) takes precedence over the
	// sticky IIN for this single response, then is consumed.
	iinBytes := s.iin
	if s.hasNext {
		iinBytes = s.nextIIN
		s.hasNext = false
	}
	resp.IIN.SetIIN(iinBytes)
	// Record application Read groups for DNP3-053 follow-on-integrity assertions.
	if apdu.FuncCode == al.FuncRead {
		s.readGroups = append(s.readGroups, requestedGroups(apdu.Data)...)
	}

	tlData := tl.EncodeFragment(tl.Fragment{FIR: true, FIN: true, Data: resp.Encode()})
	dllFrame := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncConfirmedUserDataR},
		DestAddr: s.masterAddr,
		SrcAddr:  s.outstationAddr,
		Data:     tlData,
	}
	raw, _ := frame.Encode(dllFrame)
	return raw, nil
}

// buildReadResponseData builds the object data for a Class-0 Read response,
// emitting only the MVP-supported groups requested (G1/G20/G30) using count8
// qualifiers so the public parsers populate one point each.
func (s *MVPOutstationSimulator) buildReadResponseData(reqData []byte) []byte {
	requested := requestedGroups(reqData)
	var obj []byte
	for _, g := range requested {
		switch g {
		case 1:
			obj = append(obj, encodeBinaryInputs(s.binaryInputs)...)
		case 20:
			obj = append(obj, encodeCounters(s.counters)...)
		case 30:
			obj = append(obj, encodeAnalogInputs(s.analogInputs)...)
		}
	}
	return obj
}

// buildOperateResponseData echoes the request's G12V1 object header but with the
// per-point command status replaced by the configured status (DNP3-020/021).
func (s *MVPOutstationSimulator) buildOperateResponseData(reqData []byte) []byte {
	if len(reqData) < 4 {
		return nil
	}
	// The request layout for G12V1 DirectOperate is:
	//   header(4): group, variation, qualifier(0x00), count(1)
	//   index(2), CROB(11): code,count,onTime(4),offTime(4),status(1)
	// The response echoes header + index + CROB, replacing the status byte.
	group := reqData[0]
	variation := reqData[1]
	if group != 12 || variation != 1 {
		return nil
	}
	hdr := reqData[:4]
	if len(reqData) < 4+2+11 {
		return nil
	}
	rest := append([]byte(nil), reqData[4:]...) // index + CROB
	// status byte is the last octet of the CROB (offset 2+10 within rest).
	statusIdx := 2 + 10
	if statusIdx >= len(rest) {
		return nil
	}
	rest[statusIdx] = s.commandStatus
	return append(hdr, rest...)
}

// requestedGroups extracts the distinct group numbers from a Read request's
// object headers (group,variation,qualifier[,count]) sequence.
func requestedGroups(reqData []byte) []uint8 {
	var groups []uint8
	seen := make(map[uint8]bool)
	off := 0
	for off+4 <= len(reqData) {
		g := reqData[off]
		q := reqData[off+2]
		switch q {
		case 0x07: // count8: 4-byte header
			off += 4
		case 0x06: // all-objects: 4-byte header
			off += 4
		case 0x00: // index8: 4-byte header + per-point index (1 byte each)
			count := int(reqData[off+3])
			off += 4 + count
		default:
			off += 4
		}
		if !seen[g] {
			seen[g] = true
			groups = append(groups, g)
		}
	}
	return groups
}

// encodeBinaryInputs encodes G1V1 packed binary input points with a count8
// qualifier (DNP3-036 golden data).
func encodeBinaryInputs(points []*types.BinaryInput) []byte {
	n := uint8(len(points))
	out := []byte{0x01, 0x01, 0x07, n} // G1V1, count8
	packed := make([]byte, (n+7)/8)
	for i, p := range points {
		if p.Value {
			packed[i/8] |= 1 << (uint(i) % 8)
		}
	}
	return append(out, packed...)
}

// encodeCounters encodes G20V1 counter points (5 octets/point: uint32 + flags)
// with a count8 qualifier.
func encodeCounters(points []*types.Counter) []byte {
	n := uint8(len(points))
	out := []byte{0x14, 0x01, 0x07, n} // G20V1, count8
	for _, p := range points {
		var buf [5]byte
		binary.LittleEndian.PutUint32(buf[:4], p.Value)
		buf[4] = byte(p.Quality)
		out = append(out, buf[:]...)
	}
	return out
}

// encodeAnalogInputs encodes G30V1 analog input points (5 octets/point: int32 +
// flags) with a count8 qualifier.
func encodeAnalogInputs(points []*types.AnalogInput) []byte {
	n := uint8(len(points))
	out := []byte{0x1E, 0x01, 0x07, n} // G30V1, count8
	for _, p := range points {
		var buf [5]byte
		binary.LittleEndian.PutUint32(buf[:4], uint32(int32(p.Value)))
		buf[4] = byte(p.Quality)
		out = append(out, buf[:]...)
	}
	return out
}

// Avoid unused-import warnings if math/rand etc. aren't otherwise referenced;
// math is kept for potential float32 encoding of analog outputs in a later
// profile. Suppress by referencing once at package init.
var _ = math.Float32bits
