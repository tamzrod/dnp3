package master

import (
	"context"
	"strings"
	"testing"
)

func TestUnsolicitedOperationsAreExplicitlyUnsupported(t *testing.T) {
	c, err := NewClient(NewConfig())
	if err != nil { t.Fatal(err) }
	for name, call := range map[string]func(context.Context) error{
		"enable": c.EnableUnsolicited,
		"disable": c.DisableUnsolicited,
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(context.Background()); err == nil || !strings.Contains(err.Error(), "unsupported") { t.Fatalf("error = %v", err) }
		})
	}
}
