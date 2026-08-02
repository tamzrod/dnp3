package outstation

import (
	"context"
	"testing"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg.OutstationAddress != 1024 {
		t.Errorf("Default OutstationAddress = %v, want 1024", cfg.OutstationAddress)
	}
	if cfg.MaxFragmentSize != 2048 {
		t.Errorf("Default MaxFragmentSize = %v, want 2048", cfg.MaxFragmentSize)
	}
	if cfg.MaxEventBuffers != 1000 {
		t.Errorf("Default MaxEventBuffers = %v, want 1000", cfg.MaxEventBuffers)
	}
}

func TestConfigOptions(t *testing.T) {
	cfg := NewConfig(
		WithAddress(100),
		WithMasterAddress(200),
		WithTransport(dnp3.TCP, "0.0.0.0", 54321),
		WithMaxFragmentSize(4096),
		WithUnsolicitedMode(true),
	)

	if cfg.OutstationAddress != 100 {
		t.Errorf("OutstationAddress = %v, want 100", cfg.OutstationAddress)
	}
	if cfg.MasterAddress != 200 {
		t.Errorf("MasterAddress = %v, want 200", cfg.MasterAddress)
	}
	if cfg.Port != 54321 {
		t.Errorf("Port = %v, want 54321", cfg.Port)
	}
	if cfg.MaxFragmentSize != 4096 {
		t.Errorf("MaxFragmentSize = %v, want 4096", cfg.MaxFragmentSize)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  NewConfig(),
			wantErr: false,
		},
		{
			name: "zero address",
			config: &Config{
				OutstationAddress: 0,
			},
			wantErr: true,
		},
		{
			name: "broadcast address",
			config: &Config{
				OutstationAddress: 0xFFFF,
			},
			wantErr: true,
		},
		{
			name: "invalid fragment size",
			config: &Config{
				MaxFragmentSize: 0,
			},
			wantErr: true,
		},
		{
			name: "TLS without config",
			config: &Config{
				TransportType: dnp3.TLS,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	cfg := NewConfig()
	server, err := NewServer(cfg)

	if err != nil {
		t.Errorf("NewServer() error = %v, want nil", err)
	}
	if server == nil {
		t.Error("NewServer() returned nil server")
	}
}

func TestServerState(t *testing.T) {
	cfg := NewConfig()
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server.State() != ServerStateDown {
		t.Errorf("Initial state = %v, want ServerStateDown", server.State())
	}
}

func TestServerStart(t *testing.T) {
	cfg := NewConfig()
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	ctx := context.Background()
	err = server.Start(ctx)
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}
}

func TestServerStop(t *testing.T) {
	cfg := NewConfig()
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	ctx := context.Background()
	_ = server.Start(ctx)
	err = server.Stop(ctx)
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	if server.State() != ServerStateDown {
		t.Errorf("After stop state = %v, want ServerStateDown", server.State())
	}
}

func TestServerSetDataHandler(t *testing.T) {
	cfg := NewConfig()
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Should not panic with nil handler
	server.SetDataHandler(nil)

	// Should accept valid handler
	handler := &DefaultDataHandler{}
	server.SetDataHandler(handler)
}

func TestDefaultDataHandler(t *testing.T) {
	handler := &DefaultDataHandler{}

	bi := handler.GetBinaryInputs()
	if bi == nil || len(bi) != 2 {
		t.Errorf("DefaultDataHandler.GetBinaryInputs() length = %v, want 2", len(bi))
	}

	ai := handler.GetAnalogInputs()
	if ai == nil || len(ai) != 2 {
		t.Errorf("DefaultDataHandler.GetAnalogInputs() length = %v, want 2", len(ai))
	}

	c := handler.GetCounters()
	if c == nil || len(c) != 2 {
		t.Errorf("DefaultDataHandler.GetCounters() length = %v, want 2", len(c))
	}
}

func TestServerClose(t *testing.T) {
	cfg := NewConfig()
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	err = server.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

// MockDataHandler for testing
type MockDataHandler struct {
	binaryInputs   []*types.BinaryInput
	analogInputs   []*types.AnalogInput
	counters       []*types.Counter
	binaryOutputs  []*types.BinaryOutput
	analogOutputs  []*types.AnalogOutput
	frozenCounters []*types.Counter
}

func (m *MockDataHandler) GetBinaryInputs() []*types.BinaryInput {
	return m.binaryInputs
}

func (m *MockDataHandler) GetAnalogInputs() []*types.AnalogInput {
	return m.analogInputs
}

func (m *MockDataHandler) GetCounters() []*types.Counter {
	return m.counters
}

func (m *MockDataHandler) GetBinaryOutputs() []*types.BinaryOutput {
	return m.binaryOutputs
}

func (m *MockDataHandler) GetAnalogOutputs() []*types.AnalogOutput {
	return m.analogOutputs
}

func (m *MockDataHandler) GetFrozenCounters() []*types.Counter {
	return m.frozenCounters
}

func (m *MockDataHandler) FreezeCounters(clear bool) error {
	return nil
}

func TestServerWithCustomDataHandler(t *testing.T) {
	cfg := NewConfig()
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	handler := &MockDataHandler{
		binaryInputs: []*types.BinaryInput{
			{Index: 0, Value: true, Quality: types.QualityOnline},
			{Index: 1, Value: false, Quality: types.QualityOnline},
		},
		analogInputs: []*types.AnalogInput{
			{Index: 0, Value: 123.45, Quality: types.QualityOnline},
		},
		counters: []*types.Counter{
			{Index: 0, Value: 1000, Quality: types.QualityOnline},
		},
	}

	server.SetDataHandler(handler)

	// Verify handler was set (this would be tested in full integration)
	_ = handler
}

func TestServerStateStrings(t *testing.T) {
	tests := []struct {
		state    ServerState
		expected string
	}{
		{ServerStateDown, "Down"},
		{ServerStateStarting, "Starting"},
		{ServerStateRunning, "Running"},
		{ServerStateStopping, "Stopping"},
		{ServerStateError, "Error"},
		{ServerState(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("ServerState.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestUnsolicitedMode(t *testing.T) {
	cfg := NewConfig(WithUnsolicitedMode(true))
	if !cfg.UnsolicitedMode {
		t.Error("UnsolicitedMode should be true")
	}

	cfg2 := NewConfig(WithUnsolicitedMode(false))
	if cfg2.UnsolicitedMode {
		t.Error("UnsolicitedMode should be false")
	}
}
