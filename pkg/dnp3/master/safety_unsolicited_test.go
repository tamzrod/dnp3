package master

import (
	"context"
	"errors"
	"testing"

	"dnp3/pkg/dnp3"
)

func TestUnsolicitedOperationsRequireConnection(t *testing.T) {
	c, err := NewClient(NewConfig())
	if err != nil {
		t.Fatal(err)
	}
	for name, call := range map[string]func(context.Context) error{
		"enable":  c.EnableUnsolicited,
		"disable": c.DisableUnsolicited,
		"watch":   c.WatchUnsolicited,
	} {
		t.Run(name, func(t *testing.T) {
			err := call(context.Background())
			if !errors.Is(err, dnp3.ErrNotConnected) {
				t.Fatalf("error = %v, want ErrNotConnected", err)
			}
		})
	}
	if _, err := c.ReadEvents(context.Background()); !errors.Is(err, dnp3.ErrNotConnected) {
		t.Fatalf("ReadEvents error = %v, want ErrNotConnected", err)
	}
}
