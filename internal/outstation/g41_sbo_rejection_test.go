package outstation

import (
	"testing"

	"dnp3/internal/al"
)

// AO-008 / G41 SBO fail-closed boundary.
//
// The approved G41 command profile admits exactly one mode: Direct Operate
// WITH response. A Group 41 SELECT or OPERATE (select-before-operate pair) must
// never arm a later OPERATE, invoke a host analog handler, or be acknowledged
// as success. These tests drive the reachable ProcessRequest path.

// g41SBOHandler records analog and CROB handler calls so the tests can assert
// that a rejected SBO exchange has no side effects.
type g41SBOHandler struct {
	DefaultDataHandler
	analogCalls int
	crobCalls   int
}

func (h *g41SBOHandler) WriteAnalogOutput(index uint16, value interface{}, variation uint8) error {
	h.analogCalls++
	return nil
}

func (h *g41SBOHandler) WriteBinaryOutput(index uint16, crob *CROB) error {
	h.crobCalls++
	return nil
}

func newG41SBOOutstation(h *g41SBOHandler) *Outstation {
	o := NewOutstation(nil)
	o.Initialize()
	o.Start()
	o.SetDataHandler(h)
	return o
}

// g41ControlAPDU builds a one-point 0x28-qualified Group 41 command APDU with
// the requested function code. V1 int32 value 42 at index 0.
func g41ControlAPDU(fc uint8) *al.APDU {
	data := []byte{0x29, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00, 0x00, 0x00}
	return &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: fc,
		Data:     data,
	}
}

// TestG41SelectAndOperateFailClosed verifies that neither SELECT nor OPERATE of
// a G41 SBO sequence is acknowledged, that both set ParameterError, and that
// neither invokes the host analog handler.
func TestG41SelectAndOperateFailClosed(t *testing.T) {
	for _, tc := range []struct {
		name string
		fc   uint8
	}{
		{"SELECT", al.FuncSelect},
		{"OPERATE", al.FuncOperate},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := &g41SBOHandler{}
			o := newG41SBOOutstation(h)

			resp, err := o.ProcessRequest(g41ControlAPDU(tc.fc))
			if err == nil {
				t.Fatalf("G41 %s must be rejected, got resp=%v", tc.name, resp)
			}
			if resp != nil {
				t.Fatalf("G41 %s rejection must not acknowledge success, got %v", tc.name, resp)
			}
			if !o.iin.ParameterError {
				t.Fatalf("G41 %s rejection must set IIN ParameterError", tc.name)
			}
			if h.analogCalls != 0 || h.crobCalls != 0 {
				t.Fatalf("G41 %s invoked a handler: analog=%d crob=%d", tc.name, h.analogCalls, h.crobCalls)
			}
		})
	}
}

// TestG41SelectCannotArmOperate verifies that a rejected G41 SELECT leaves no
// pending select, so a following G41 OPERATE is independently rejected rather
// than executed against armed state.
func TestG41SelectCannotArmOperate(t *testing.T) {
	h := &g41SBOHandler{}
	o := newG41SBOOutstation(h)

	if _, err := o.ProcessRequest(g41ControlAPDU(al.FuncSelect)); err == nil {
		t.Fatal("expected G41 SELECT rejection")
	}
	if got := len(o.pendingSelects); got != 0 {
		t.Fatalf("rejected G41 SELECT armed a pending select: count=%d", got)
	}
	if _, err := o.ProcessRequest(g41ControlAPDU(al.FuncOperate)); err == nil {
		t.Fatal("expected G41 OPERATE rejection after SELECT")
	}
	if h.analogCalls != 0 {
		t.Fatalf("G41 SBO invoked the analog handler: calls=%d", h.analogCalls)
	}
}

// TestG41SBORejectionDoesNotBreakDirectOperate verifies the guard is specific
// to SELECT/OPERATE and that the supported Direct Operate path still executes.
func TestG41SBORejectionDoesNotBreakDirectOperate(t *testing.T) {
	h := &g41SBOHandler{}
	o := newG41SBOOutstation(h)

	resp, err := o.ProcessRequest(g41ControlAPDU(al.FuncDirectOperate))
	if err != nil {
		t.Fatalf("G41 Direct Operate should still be accepted, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected a Direct Operate response")
	}
	if h.analogCalls != 1 {
		t.Fatalf("analog handler calls=%d, want 1", h.analogCalls)
	}
}
