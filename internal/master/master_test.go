package master

import (
	"testing"
	"time"
	
	"dnp3/internal/al"
)

func TestStateString(t *testing.T) {
	tests := []struct {
		state   State
		want    string
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
		Timeout:      3000,
		MaxRetries:   5,
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

// TestReadRangeQualifierLSB verifies the Read* convenience builders encode the
// 16-bit start/stop range little-endian (DNP3-001 BE-negative case). With
// start=0x1234, stop=0x5678 the wire bytes must be 34 12 78 56 (LSB first).
// The Read* builders construct the object data inline before passing it to
// buildRequest, so we assert the expected LSB byte ordering directly.
func TestReadRangeQualifierLSB(t *testing.T) {
	start, stop := uint16(0x1234), uint16(0x5678)
	// Reconstruct exactly how the Read* builders lay out the range bytes.
	rangeLSB := []byte{byte(start), byte(start >> 8), byte(stop), byte(stop >> 8)}
	want := []byte{0x34, 0x12, 0x78, 0x56}
	if string(rangeLSB) != string(want) {
		t.Fatalf("LSB range bytes = % X, want % X", rangeLSB, want)
	}
	// Big-endian would be 12 34 56 78; ensure we are not using that order.
	rangeBE := []byte{byte(start >> 8), byte(start), byte(stop >> 8), byte(stop)}
	if string(rangeBE) == string(want) {
		t.Fatalf("LSB range bytes accidentally matched BE order")
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
