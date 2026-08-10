package master

import (
	"crypto/tls"
	"strings"
	"testing"
)

func TestNewClientRejectsUnsupportedTLS(t *testing.T) {
	_, err := NewClient(NewConfig(WithTLS(&tls.Config{})))
	if err == nil || !strings.Contains(err.Error(), "TLS transport is unsupported") {
		t.Fatalf("NewClient TLS error = %v", err)
	}
}
