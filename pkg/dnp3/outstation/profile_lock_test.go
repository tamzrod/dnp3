package outstation

import (
	"errors"
	"testing"

	"dnp3/pkg/dnp3"
)

// TestSupportedProfileRejectionMatrix verifies DNP3-087: non-v0 options are
// rejected by NewServer (via Config.Validate) with clear ConfigurationError
// messages naming the offending field — no silent fallback to a v0 behavior.
func TestSupportedProfileRejectionMatrix(t *testing.T) {
	cases := []struct {
		name    string
		cfg     *Config
		wantErr bool
		field   string // expected ConfigurationError.Field, if wantErr
	}{
		{
			name:    "valid TCP single-master",
			cfg:     NewConfig(WithTransport(dnp3.TCP, "127.0.0.1", 0)),
			wantErr: false,
		},
		{
			name:    "TLS via WithTLS rejected",
			cfg:     NewConfig(WithTLS(nil)),
			wantErr: true,
			field:   "TransportType",
		},
		{
			name: "TLS transport type rejected (even with config)",
			cfg: func() *Config {
				c := NewConfig()
				c.TransportType = dnp3.TLS
				c.TLSConfig = nil // non-nil would need a real *tls.Config; nil is fine
				return c
			}(),
			wantErr: true,
			field:   "TransportType",
		},
		{
			name:    "unsolicited mode rejected",
			cfg:     NewConfig(WithUnsolicitedMode(true)),
			wantErr: true,
			field:   "UnsolicitedMode",
		},
		{
			name:    "MaxConnections zero rejected",
			cfg:     NewConfig(WithMaxConnections(0)),
			wantErr: true,
			field:   "MaxConnections",
		},
		{
			name:    "valid TCP with unsolicited disabled",
			cfg:     NewConfig(WithTransport(dnp3.TCP, "127.0.0.1", 0), WithUnsolicitedMode(false)),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewServer(tc.cfg)
			if (err != nil) != tc.wantErr {
				t.Fatalf("NewServer err = %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr {
				return
			}
			var ce *dnp3.ConfigurationError
			if !errors.As(err, &ce) {
				t.Fatalf("expected *ConfigurationError, got %T: %v", err, err)
			}
			if ce.Field != tc.field {
				t.Fatalf("ConfigurationError.Field = %q, want %q (clear error)", ce.Field, tc.field)
			}
		})
	}
}

// TestTLSDoesNotSilentlyFallBackToTCP verifies DNP3-087: a TLS configuration
// must not be silently accepted and then run as TCP — NewServer must reject it.
func TestTLSDoesNotSilentlyFallBackToTCP(t *testing.T) {
	cfg := NewConfig(WithTransport(dnp3.TCP, "127.0.0.1", 0))
	// Flip to TLS after valid construction; Validate must catch it.
	cfg.TransportType = dnp3.TLS
	if _, err := NewServer(cfg); err == nil {
		t.Fatal("TLS must be rejected, not silently accepted as TCP")
	}
}
