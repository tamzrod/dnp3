package master

import (
	"context"
	"testing"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg.MasterAddress != 0xFFFF {
		t.Errorf("Default MasterAddress = %v, want 0xFFFF", cfg.MasterAddress)
	}
	if cfg.OutstationAddress != 1024 {
		t.Errorf("Default OutstationAddress = %v, want 1024", cfg.OutstationAddress)
	}
	if cfg.Timeout != 5*time.Second {
		t.Errorf("Default Timeout = %v, want 5s", cfg.Timeout)
	}
	if cfg.RetryCount != 3 {
		t.Errorf("Default RetryCount = %v, want 3", cfg.RetryCount)
	}
}

func TestConfigOptions(t *testing.T) {
	cfg := NewConfig(
		WithMasterAddress(100),
		WithOutstationAddress(200),
		WithTransport(dnp3.TCP, "192.168.1.1", 12345),
		WithTimeout(10*time.Second),
		WithRetry(5, 200*time.Millisecond),
	)

	if cfg.MasterAddress != 100 {
		t.Errorf("MasterAddress = %v, want 100", cfg.MasterAddress)
	}
	if cfg.OutstationAddress != 200 {
		t.Errorf("OutstationAddress = %v, want 200", cfg.OutstationAddress)
	}
	if cfg.Address != "192.168.1.1" {
		t.Errorf("Address = %v, want 192.168.1.1", cfg.Address)
	}
	if cfg.Port != 12345 {
		t.Errorf("Port = %v, want 12345", cfg.Port)
	}
	if cfg.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want 10s", cfg.Timeout)
	}
	if cfg.RetryCount != 5 {
		t.Errorf("RetryCount = %v, want 5", cfg.RetryCount)
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
			name: "negative timeout",
			config: &Config{
				Timeout: -1,
			},
			wantErr: true,
		},
		{
			name: "negative retry count",
			config: &Config{
				RetryCount: -1,
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

func TestNewClient(t *testing.T) {
	cfg := NewConfig()
	client, err := NewClient(cfg)

	if err != nil {
		t.Errorf("NewClient() error = %v, want nil", err)
	}
	if client == nil {
		t.Error("NewClient() returned nil client")
	}
}

func TestClientState(t *testing.T) {
	cfg := NewConfig()
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if client.State() != dnp3.StateDisconnected {
		t.Errorf("Initial state = %v, want StateDisconnected", client.State())
	}
}

func TestClientConnect(t *testing.T) {
	t.Skip("Skipping integration test - requires mock server")
	cfg := NewConfig()
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		t.Errorf("Connect() error = %v", err)
	}
}

func TestClientDisconnect(t *testing.T) {
	cfg := NewConfig()
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()
	_ = client.Connect(ctx)
	err = client.Disconnect(ctx)
	if err != nil {
		t.Errorf("Disconnect() error = %v", err)
	}

	if client.State() != dnp3.StateDisconnected {
		t.Errorf("After disconnect state = %v, want StateDisconnected", client.State())
	}
}

func TestClientRead(t *testing.T) {
	t.Skip("Skipping integration test - requires mock server")
	cfg := NewConfig()
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()

	// Should fail when not connected
	_, err = client.Read(ctx, nil)
	if err == nil {
		t.Error("Read() should fail when not connected")
	}

	// Should succeed when connected (TODO: real implementation)
	_ = client.Connect(ctx)
	resp, err := client.Read(ctx, &types.ReadRequest{})
	if err != nil {
		t.Errorf("Read() error = %v", err)
	}
	if resp == nil {
		t.Error("Read() returned nil response")
	}
}

func TestClientOperate(t *testing.T) {
	cfg := NewConfig()
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()

	// Should fail when not connected
	_, err = client.Operate(ctx, nil)
	if err == nil {
		t.Error("Operate() should fail when not connected")
	}
}

func TestClientClose(t *testing.T) {
	cfg := NewConfig()
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	err = client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
