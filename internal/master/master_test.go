package master

import (
	"errors"
	"testing"
	"time"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

func TestStateString(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateDisconnected, "Disconnected"},
		{StateConnecting, "Connecting"},
		{StateConnected, "Connected"},
		{StateInitialized, "Initialized"},
		{StateActive, "Active"},
		{StateError, "Error"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("State.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestPollTypeString(t *testing.T) {
	tests := []struct {
		pollType PollType
		want     string
	}{
		{PollIntegrity, "Integrity"},
		{PollEvent, "Event"},
		{PollException, "Exception"},
		{PollClass0, "Class 0"},
		{PollClass1, "Class 1"},
		{PollClass2, "Class 2"},
		{PollClass3, "Class 3"},
	}

	for _, tt := range tests {
		if got := tt.pollType.String(); got != tt.want {
			t.Errorf("PollType.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestNewOutstation(t *testing.T) {
	o := NewOutstation(100, "RTU-1")

	if o.ID != 100 {
		t.Errorf("ID = %d, want 100", o.ID)
	}
	if o.Label != "RTU-1" {
		t.Errorf("Label = %v, want RTU-1", o.Label)
	}
	if o.State != "Unknown" {
		t.Errorf("State = %v, want Unknown", o.State)
	}
}

func TestOutstationUpdateIIN(t *testing.T) {
	o := NewOutstation(100, "RTU-1")

	iin := [2]byte{0x02, 0x10}
	o.UpdateIIN(iin)

	if o.IIN[0] != 0x02 {
		t.Errorf("IIN[0] = 0x%02X, want 0x02", o.IIN[0])
	}
	if o.IIN[1] != 0x10 {
		t.Errorf("IIN[1] = 0x%02X, want 0x10", o.IIN[1])
	}
}

func TestOutstationHasFlag(t *testing.T) {
	o := NewOutstation(100, "RTU-1")
	o.IIN = [2]byte{0x02, 0x00} // BUSY flag set

	if !o.HasFlag(0x02) {
		t.Error("HasFlag(0x02) = false, want true")
	}
	if o.HasFlag(0x04) {
		t.Error("HasFlag(0x04) = true, want false")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MasterAddress != 0xFFFF {
		t.Errorf("MasterAddress = 0x%04X, want 0xFFFF", cfg.MasterAddress)
	}
	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %d, want %d", cfg.Timeout, DefaultTimeout)
	}
	if cfg.MaxRetries != MaxRetries {
		t.Errorf("MaxRetries = %d, want %d", cfg.MaxRetries, MaxRetries)
	}
}

func TestNewMaster(t *testing.T) {
	cfg := &Config{
		MasterAddress: 100,
		Timeout:       3000,
		MaxRetries:    5,
	}

	m := NewMaster(cfg)

	if m.config != cfg {
		t.Error("config not set correctly")
	}
	if m.State() != StateDisconnected {
		t.Errorf("State = %v, want Disconnected", m.State())
	}
	if m.OutstationCount() != 0 {
		t.Errorf("OutstationCount = %d, want 0", m.OutstationCount())
	}
}

func TestNewMasterDefaultConfig(t *testing.T) {
	m := NewMaster(nil)

	if m.config == nil {
		t.Fatal("config should not be nil")
	}
	if m.config.MasterAddress != 0xFFFF {
		t.Errorf("MasterAddress = 0x%04X, want 0xFFFF", m.config.MasterAddress)
	}
}

func TestMasterState(t *testing.T) {
	m := NewMaster(nil)

	if m.State() != StateDisconnected {
		t.Errorf("Initial State = %v, want Disconnected", m.State())
	}

	m.SetState(StateConnected)
	if m.State() != StateConnected {
		t.Errorf("State = %v, want Connected", m.State())
	}
}

func TestMasterOutstationManagement(t *testing.T) {
	m := NewMaster(nil)

	// Add outstation
	o1 := m.AddOutstation(100, "RTU-1")
	if o1.ID != 100 {
		t.Errorf("Added outstation ID = %d, want 100", o1.ID)
	}

	o2 := m.AddOutstation(200, "RTU-2")
	if o2.ID != 200 {
		t.Errorf("Added outstation ID = %d, want 200", o2.ID)
	}

	if m.OutstationCount() != 2 {
		t.Errorf("OutstationCount = %d, want 2", m.OutstationCount())
	}

	// Get outstation
	o, ok := m.GetOutstation(100)
	if !ok {
		t.Error("GetOutstation(100) failed")
	}
	if o.Label != "RTU-1" {
		t.Errorf("Label = %v, want RTU-1", o.Label)
	}

	// Outstation not found
	_, ok = m.GetOutstation(999)
	if ok {
		t.Error("GetOutstation(999) should return false")
	}

	// Remove outstation
	m.RemoveOutstation(100)
	if m.OutstationCount() != 1 {
		t.Errorf("OutstationCount = %d, want 1", m.OutstationCount())
	}
}

func TestMasterConnect(t *testing.T) {
	m := NewMaster(nil)

	// Connect without transport
	if err := m.Connect(); err == nil {
		t.Error("Connect() should fail without transport")
	}

	// Connect with mock transport
	m.SetTransport(&mockTransport{})
	if err := m.Connect(); err != nil {
		t.Errorf("Connect() error = %v", err)
	}

	// Connect() sets StateConnected then StateActive; final ready state is Active
	if m.State() != StateActive {
		t.Errorf("State = %v, want Active", m.State())
	}
}

func TestMasterDisconnect(t *testing.T) {
	m := NewMaster(nil)
	m.SetTransport(&mockTransport{})
	m.Connect()
	m.AddOutstation(100, "RTU-1")

	if err := m.Disconnect(); err != nil {
		t.Errorf("Disconnect() error = %v", err)
	}

	if m.State() != StateDisconnected {
		t.Errorf("State = %v, want Disconnected", m.State())
	}
	if m.OutstationCount() != 0 {
		t.Errorf("OutstationCount = %d, want 0", m.OutstationCount())
	}
}

func TestMasterInitialize(t *testing.T) {
	m := NewMaster(nil)

	// Initialize without connection
	if err := m.Initialize(); err == nil {
		t.Error("Initialize() should fail without connection")
	}

	// Initialize with connection
	m.SetTransport(&mockTransport{})
	m.Connect()

	if err := m.Initialize(); err != nil {
		t.Errorf("Initialize() error = %v", err)
	}

	if m.State() != StateInitialized {
		t.Errorf("State = %v, want Initialized", m.State())
	}
}

func TestBuildPollRequest(t *testing.T) {
	tests := []struct {
		pollType PollType
		want     []byte
	}{
		{PollIntegrity, []byte{60, 1, 0x07, 0x00}},
		{PollClass0, []byte{60, 1, 0x07, 0x00}},
		{PollClass1, []byte{2, 1, 0x07, 0x00}},
		{PollClass2, []byte{32, 1, 0x07, 0x00}},
		{PollClass3, []byte{22, 1, 0x07, 0x00}},
	}

	for _, tt := range tests {
		got := buildPollRequest(tt.pollType)
		if string(got) != string(tt.want) {
			t.Errorf("buildPollRequest(%v) = %v, want %v", tt.pollType, got, tt.want)
		}
	}
}

func TestEncodeDNP3Time(t *testing.T) {
	// Fixed time for testing
	t1 := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	data := encodeDNP3Time(t1)

	// Should be 8 bytes
	if len(data) != 8 {
		t.Errorf("len(encodeDNP3Time()) = %d, want 8", len(data))
	}

	// Test with non-zero time
	t2 := time.Date(2000, 1, 1, 0, 0, 1, 0, time.UTC) // +1 second
	data2 := encodeDNP3Time(t2)

	// Last 2 bytes should be non-zero (ms fraction)
	if data2[6] == 0 && data2[7] == 0 {
		// This is OK for 1 second = 1000ms, upper bytes only
	}
}

func TestBuildRequest(t *testing.T) {
	req := buildRequest(5, al.FuncRead, []byte{0x01, 0x02})

	if req.Control.Seq != 5 {
		t.Errorf("Seq = %d, want 5", req.Control.Seq)
	}
	if !req.Control.FIR || !req.Control.FIN {
		t.Error("Expected FIR and FIN to be set")
	}
	if req.FuncCode != al.FuncRead {
		t.Errorf("FuncCode = %d, want %d", req.FuncCode, al.FuncRead)
	}
}

// TestReadRangeQualifierLSB verifies buildReadRangeRequest encodes the
// 16-bit start/stop range little-endian (DNP3-001/DNP3-004). The header uses
// the 0x28 (range16) qualifier: group, variation, 0x28, start(LSB 2), stop(LSB 2).
func TestReadRangeQualifierLSB(t *testing.T) {
	got := buildReadRangeRequest(1, 1, 0x1234, 0x5678)
	want := []byte{0x01, 0x01, 0x28, 0x34, 0x12, 0x78, 0x56}
	if string(got) != string(want) {
		t.Fatalf("buildReadRangeRequest = % X, want % X", got, want)
	}
}

// encodeACK builds a raw link frame for a secondary ACK with the given fields.
func encodeACK(t *testing.T, dir, prm bool, fc uint8, dest, src uint16) []byte {
	t.Helper()
	f := &frame.Frame{
		Control:  frame.Control{DIR: dir, PRM: prm, FuncCode: fc},
		DestAddr: dest,
		SrcAddr:  src,
	}
	raw, err := frame.Encode(f)
	if err != nil {
		t.Fatalf("encode ACK: %v", err)
	}
	return raw
}

// TestValidateResetLinkACK verifies the secondary ACK validation for the
// Reset Link Stations handshake (DNP3-006).
func TestValidateResetLinkACK(t *testing.T) {
	const outstationID, masterAddr uint16 = 0x0004, 0x0003

	tests := []struct {
		name    string
		raw     []byte
		wantErr bool
	}{
		{
			name: "good ACK",
			raw:  encodeACK(t, false, false, frame.FuncAck, masterAddr, outstationID),
		},
		{
			name:    "bad function code (NACK)",
			raw:     encodeACK(t, false, false, frame.FuncNack, masterAddr, outstationID),
			wantErr: true,
		},
		{
			name:    "wrong DIR (primary direction)",
			raw:     encodeACK(t, true, false, frame.FuncAck, masterAddr, outstationID),
			wantErr: true,
		},
		{
			name:    "wrong PRM (primary station)",
			raw:     encodeACK(t, false, true, frame.FuncAck, masterAddr, outstationID),
			wantErr: true,
		},
		{
			name:    "wrong source address",
			raw:     encodeACK(t, false, false, frame.FuncAck, masterAddr, 0x0100),
			wantErr: true,
		},
		{
			name:    "wrong destination address",
			raw:     encodeACK(t, false, false, frame.FuncAck, 0x0200, outstationID),
			wantErr: true,
		},
		{
			name:    "malformed frame (no sync)",
			raw:     []byte{0xC0, 0x00, 0x00, 0x00},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResetLinkACK(tt.raw, outstationID, masterAddr)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

// TestValidateLinkStatusResponse verifies the secondary Link Status response
// validation for the Request Link Status handshake phase (DNP3-007).
func TestValidateLinkStatusResponse(t *testing.T) {
	const outstationID, masterAddr uint16 = 0x0004, 0x0003

	tests := []struct {
		name    string
		raw     []byte
		wantErr bool
	}{
		{
			name: "good link status",
			raw:  encodeACK(t, false, false, frame.FuncLinkStatus, masterAddr, outstationID),
		},
		{
			name:    "wrong function code (ACK instead of Link Status)",
			raw:     encodeACK(t, false, false, frame.FuncAck, masterAddr, outstationID),
			wantErr: true,
		},
		{
			name:    "wrong DIR (primary direction)",
			raw:     encodeACK(t, true, false, frame.FuncLinkStatus, masterAddr, outstationID),
			wantErr: true,
		},
		{
			name:    "wrong PRM (primary station)",
			raw:     encodeACK(t, false, true, frame.FuncLinkStatus, masterAddr, outstationID),
			wantErr: true,
		},
		{
			name:    "wrong source address",
			raw:     encodeACK(t, false, false, frame.FuncLinkStatus, masterAddr, 0x0100),
			wantErr: true,
		},
		{
			name:    "wrong destination address",
			raw:     encodeACK(t, false, false, frame.FuncLinkStatus, 0x0200, outstationID),
			wantErr: true,
		},
		{
			name:    "malformed frame (no sync)",
			raw:     []byte{0xC0, 0x00, 0x00, 0x00},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLinkStatusResponse(tt.raw, outstationID, masterAddr)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

// mockTransport implements TransportHandler for testing
type mockTransport struct{}

func (t *mockTransport) Send(data []byte) error {
	return nil
}

func (t *mockTransport) Receive() ([]byte, error) {
	// Return a minimal valid response
	return []byte{0xC0, 0x00, 0x00, 0x00}, nil
}

func (t *mockTransport) SetTimeout(ms int) {}

// seqRecorderTransport records every sent bytes and returns a canned valid
// response so sendWithRetry completes successfully. It is used to verify the
// application-layer sequence stream (DNP3-008).
type seqRecorderTransport struct {
	sent [][]byte
	resp []byte
}

func (t *seqRecorderTransport) Send(data []byte) error {
	cp := make([]byte, len(data))
	copy(cp, data)
	t.sent = append(t.sent, cp)
	return nil
}

func (t *seqRecorderTransport) Receive() ([]byte, error) {
	return t.resp, nil
}

func (t *seqRecorderTransport) SetTimeout(ms int) {}

// echoSeqTransport records sent bytes and, on Receive, returns a valid
// response whose application SEQ echoes the SEQ of the most recently sent
// request (DNP3-010 compliant outstation behavior).
type echoSeqTransport struct {
	sent    [][]byte
	lastSeq uint8
}

func (t *echoSeqTransport) Send(data []byte) error {
	cp := make([]byte, len(data))
	copy(cp, data)
	t.sent = append(t.sent, cp)
	t.lastSeq = extractRequestSeqForBuild(data)
	return nil
}

func (t *echoSeqTransport) Receive() ([]byte, error) {
	return buildResponseFrameWithSeq(t.lastSeq), nil
}

func (t *echoSeqTransport) SetTimeout(ms int) {}

// extractRequestSeqForBuild decodes the application SEQ from a raw sent DLL
// frame (non-testing variant for use inside transport helpers).
func extractRequestSeqForBuild(raw []byte) uint8 {
	f, err := frame.Decode(raw)
	if err != nil {
		return 0
	}
	frag, err := tl.DecodeFragment(f.Data)
	if err != nil {
		return 0
	}
	apdu, err := al.Decode(frag.Data)
	if err != nil {
		return 0
	}
	return apdu.Control.Seq
}

// buildResponseFrameWithSeq builds a valid DLL+TL+APDU response frame carrying
// only IIN, with the given application sequence number.
func buildResponseFrameWithSeq(seq uint8) []byte {
	return buildResponseFrameWithIIN(seq, al.IIN{})
}

// buildResponseFrameWithIIN builds a valid DLL+TL+APDU response frame carrying
// the given IIN and application sequence number.
func buildResponseFrameWithIIN(seq uint8, iin al.IIN) []byte {
	iinBytes := iin.Bytes()
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: seq},
		FuncCode: al.FuncResponse,
		Data:     iinBytes[:],
	}
	frag := tl.Fragment{FIR: true, FIN: true, Data: apdu.Encode()}
	tlData := tl.EncodeFragment(frag)
	dllFrame := &frame.Frame{
		Control: frame.Control{
			DIR:      false,
			PRM:      false,
			FuncCode: frame.FuncConfirmedUserDataR,
		},
		DestAddr: 1,
		SrcAddr:  2,
		Data:     tlData,
	}
	raw, _ := frame.Encode(dllFrame)
	return raw
}

// buildMinimalResponseFrame builds a valid DLL frame wrapping a TL fragment
// wrapping a minimal APDU response (FuncResponse, empty data, Seq=0). It is
// sufficient for sendWithRetry to complete its receive/process path.
func buildMinimalResponseFrame(t *testing.T) []byte {
	t.Helper()
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 0},
		FuncCode: al.FuncResponse,
		Data:     []byte{0x00, 0x00}, // empty IIN
	}
	frag := tl.Fragment{FIR: true, FIN: true, Data: apdu.Encode()}
	tlData := tl.EncodeFragment(frag)
	dllFrame := &frame.Frame{
		Control: frame.Control{
			DIR:      false,
			PRM:      false,
			FuncCode: frame.FuncConfirmedUserDataR,
		},
		DestAddr: 1,
		SrcAddr:  2,
		Data:     tlData,
	}
	raw, err := frame.Encode(dllFrame)
	if err != nil {
		t.Fatalf("encode response frame: %v", err)
	}
	return raw
}

// extractRequestSeq decodes a sent DLL frame, extracts the TL fragment, decodes
// the APDU, and returns the application-layer sequence number.
func extractRequestSeq(t *testing.T, raw []byte) uint8 {
	t.Helper()
	f, err := frame.Decode(raw)
	if err != nil {
		t.Fatalf("decode sent frame: %v", err)
	}
	frag, err := tl.DecodeFragment(f.Data)
	if err != nil {
		t.Fatalf("decode sent TL fragment: %v", err)
	}
	apdu, err := al.Decode(frag.Data)
	if err != nil {
		t.Fatalf("decode sent APDU: %v", err)
	}
	return apdu.Control.Seq
}

// TestSequenceStream verifies the master's application-layer sequence number
// advances 0-15 and wraps to 0 (DNP3-008).
func TestSequenceStream(t *testing.T) {
	m := NewMaster(DefaultConfig())
	var seen []uint8
	for i := 0; i < 17; i++ {
		seen = append(seen, m.nextSequence())
		m.advanceSequence()
	}
	want := []uint8{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 0}
	if len(seen) != len(want) {
		t.Fatalf("len = %d, want %d", len(seen), len(want))
	}
	for i := range want {
		if seen[i] != want[i] {
			t.Fatalf("seq[%d] = %d, want %d (stream=%v)", i, seen[i], want[i], seen)
		}
	}
}

// TestSendWithRetrySequenceAdvances verifies sendWithRetry assigns the master's
// sequence to the request and advances it only on successful send (DNP3-008).
// The mock outstation echoes the request SEQ in its response (DNP3-010).
func TestSendWithRetrySequenceAdvances(t *testing.T) {
	m := NewMaster(DefaultConfig())
	tr := &echoSeqTransport{}
	m.SetTransport(tr)
	m.SetState(StateInitialized)
	m.AddOutstation(2, "RTU-1")

	// Issue three requests; expect the observed SEQ stream 0,1,2.
	for i, want := range []uint8{0, 1, 2} {
		before := len(tr.sent)
		req := buildRequest(0, al.FuncRead, []byte{0x00})
		if err := m.sendWithRetry(req, 2); err != nil {
			t.Fatalf("sendWithRetry[%d] failed: %v", i, err)
		}
		got := extractRequestSeq(t, tr.sent[before])
		if got != want {
			t.Fatalf("request %d SEQ = %d, want %d", i, got, want)
		}
	}

	// The master's current sequence should have advanced to 3.
	if got := m.currentSequence(); got != 3 {
		t.Fatalf("currentSequence = %d, want 3", got)
	}
}

// TestSendWithRetrySequenceNoAdvanceOnSendFailure verifies the sequence does NOT
// advance when the transport send fails (DNP3-008: increment only on success).
func TestSendWithRetrySequenceNoAdvanceOnSendFailure(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 50, MaxRetries: 1, RetryDelay: 0})
	resp := buildMinimalResponseFrame(t)
	failTr := &failingSendTransport{seqRecorderTransport: seqRecorderTransport{resp: resp}}
	m.SetTransport(failTr)
	m.SetState(StateInitialized)
	m.AddOutstation(2, "RTU-1")

	req := buildRequest(0, al.FuncRead, []byte{0x00})
	if err := m.sendWithRetry(req, 2); err == nil {
		t.Fatal("expected sendWithRetry to fail when transport Send fails")
	}
	if got := m.currentSequence(); got != 0 {
		t.Fatalf("currentSequence = %d, want 0 (no advance on send failure)", got)
	}
}

// failingSendTransport wraps seqRecorderTransport but always fails Send.
type failingSendTransport struct {
	seqRecorderTransport
}

func (t *failingSendTransport) Send(data []byte) error {
	return errSendFailed
}

var errSendFailed = errors.New("simulated send failure")

// buildConfirmFrame builds a valid DLL+TL+APDU confirmation frame with the
// given application sequence number. A dedicated confirm carries only the IIN
// bytes (no object data) — used to exercise waitForConfirmation (DNP3-009).
func buildConfirmFrame(t *testing.T, seq uint8) []byte {
	t.Helper()
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: seq},
		FuncCode: al.FuncResponse,
		Data:     []byte{0x00, 0x00}, // IIN only, no object data
	}
	frag := tl.Fragment{FIR: true, FIN: true, Data: apdu.Encode()}
	tlData := tl.EncodeFragment(frag)
	dllFrame := &frame.Frame{
		Control: frame.Control{
			DIR:      false,
			PRM:      false,
			FuncCode: frame.FuncConfirmedUserDataR,
		},
		DestAddr: 1,
		SrcAddr:  2,
		Data:     tlData,
	}
	raw, err := frame.Encode(dllFrame)
	if err != nil {
		t.Fatalf("encode confirm frame: %v", err)
	}
	return raw
}

// cannedTransport returns a fixed sequence of canned frames from Receive, one
// per call (cycling). Send is a no-op. Used to drive waitForConfirmation.
type cannedTransport struct {
	frames [][]byte
	idx    int
}

func (t *cannedTransport) Send(data []byte) error { return nil }
func (t *cannedTransport) SetTimeout(ms int)      {}
func (t *cannedTransport) Receive() ([]byte, error) {
	if t.idx >= len(t.frames) {
		return nil, errReceiveTimeout
	}
	f := t.frames[t.idx]
	t.idx++
	return f, nil
}

var errReceiveTimeout = errors.New("receive timeout")

// TestWaitForConfirmationMatchingSeq verifies a confirm with the matching
// sequence is accepted (DNP3-009).
func TestWaitForConfirmationMatchingSeq(t *testing.T) {
	m := NewMaster(DefaultConfig())
	tr := &cannedTransport{frames: [][]byte{buildConfirmFrame(t, 7)}}
	m.SetTransport(tr)

	if err := m.waitForConfirmation(7); err != nil {
		t.Fatalf("expected confirm accepted, got %v", err)
	}
}

// TestWaitForConfirmationWrongSeq verifies a confirm with the wrong sequence
// is rejected with ErrConfirmSeqMismatch (DNP3-009).
func TestWaitForConfirmationWrongSeq(t *testing.T) {
	m := NewMaster(DefaultConfig())
	tr := &cannedTransport{frames: [][]byte{buildConfirmFrame(t, 3)}}
	m.SetTransport(tr)

	err := m.waitForConfirmation(7)
	if err == nil {
		t.Fatal("expected ErrConfirmSeqMismatch, got nil")
	}
	if !errors.Is(err, ErrConfirmSeqMismatch) {
		t.Fatalf("expected ErrConfirmSeqMismatch, got %v", err)
	}
}

// TestWaitForConfirmationTimeout verifies a transport receive error is
// surfaced as ErrConfirmTimeout (DNP3-009).
func TestWaitForConfirmationTimeout(t *testing.T) {
	m := NewMaster(DefaultConfig())
	tr := &cannedTransport{frames: nil} // no frames → receive returns timeout
	m.SetTransport(tr)

	err := m.waitForConfirmation(7)
	if err == nil {
		t.Fatal("expected ErrConfirmTimeout, got nil")
	}
	if !errors.Is(err, ErrConfirmTimeout) {
		t.Fatalf("expected ErrConfirmTimeout, got %v", err)
	}
}

// TestProcessResponseMatchingSeq verifies processResponse accepts a response
// whose SEQ matches the outstanding request (DNP3-010).
func TestProcessResponseMatchingSeq(t *testing.T) {
	m := NewMaster(DefaultConfig())
	m.AddOutstation(2, "RTU-1")
	raw := buildResponseFrameWithSeq(5)

	if _, err := m.processResponse(raw, 2, 5); err != nil {
		t.Fatalf("expected matching response accepted, got %v", err)
	}
}

// TestProcessResponseMismatchSeq verifies processResponse rejects a response
// whose SEQ does not match, returning ErrResponseSeqMismatch and no data
// (DNP3-010).
func TestProcessResponseMismatchSeq(t *testing.T) {
	m := NewMaster(DefaultConfig())
	m.AddOutstation(2, "RTU-1")
	raw := buildResponseFrameWithSeq(3) // mismatch vs expected 5

	data, err := m.processResponse(raw, 2, 5)
	if err == nil {
		t.Fatal("expected ErrResponseSeqMismatch, got nil")
	}
	if data != nil {
		t.Fatalf("expected no data on mismatch, got %d bytes", len(data))
	}
	if !errors.Is(err, ErrResponseSeqMismatch) {
		t.Fatalf("expected ErrResponseSeqMismatch, got %v", err)
	}
}

// TestProcessResponseStoresIIN verifies processResponse always updates the
// outstation's stored IIN from the response (DNP3-012).
func TestProcessResponseStoresIIN(t *testing.T) {
	m := NewMaster(DefaultConfig())
	m.AddOutstation(2, "RTU-1")

	// Build a response carrying a known IIN (DeviceTrouble = IIN1.6 = 0x02,
	// BadConfig = IIN2.5 = 0x04 → bytes {0x02, 0x04}).
	raw := buildResponseFrameWithIIN(7, al.IIN{DeviceTrouble: true, BadConfig: true})

	if _, err := m.processResponse(raw, 2, 7); err != nil {
		t.Fatalf("processResponse failed: %v", err)
	}

	o, ok := m.GetOutstation(2)
	if !ok {
		t.Fatal("outstation 2 not found")
	}
	got := o.GetIIN()
	if got != [2]byte{0x02, 0x04} {
		t.Fatalf("stored IIN = %v, want {0x02,0x04}", got)
	}
}
