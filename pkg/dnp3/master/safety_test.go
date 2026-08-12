package master

import (
	"crypto/tls"
	"errors"
	"testing"

	"dnp3/pkg/dnp3"
)

func TestNewClientRejectsUnsupportedTLS(t *testing.T) {
	_, err := NewClient(NewConfig(WithTLS(&tls.Config{})))
	if err == nil {
		t.Fatalf("NewClient TLS: expected error, got nil")
	}
	if !errors.Is(err, dnp3.ErrUnsupportedOption) {
		t.Fatalf("NewClient TLS error = %v, want ErrUnsupportedOption", err)
	}
}
