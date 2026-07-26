// Package integration provides integration tests for DNP3 Master-Outstation communication.
package integration

import (
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/outstation"
)

// TestOutstationProcessReadRequest tests READ request processing.
func TestOutstationProcessReadRequest(t *testing.T) {
	ost := outstation.NewOutstation(nil)
	ost.Initialize()
	ost.Start()

	// Create a READ request for all data
	request := &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: false,
			Seq: 1,
		},
		FuncCode: al.FuncRead,
		Data:     []byte{60, 1, 0x07, 0x00}, // Group 60, all data
	}

	// Process request
	response, err := ost.ProcessRequest(request)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.FuncCode != al.FuncResponse {
		t.Errorf("Expected FuncCode RESPONSE (0), got %d", response.FuncCode)
	}

	if len(response.Data) == 0 {
		t.Error("Expected data in response")
	}

	t.Logf("READ request processed: FuncCode=%d, DataLen=%d", response.FuncCode, len(response.Data))
}

// TestOutstationProcessWriteRequest tests WRITE request processing.
func TestOutstationProcessWriteRequest(t *testing.T) {
	ost := outstation.NewOutstation(nil)
	ost.Initialize()
	ost.Start()

	// Create a WRITE request
	// Format: Group(1) + Variation(1) + Qualifier(1) + Count(1) + [Index(2) + Value(n)]...
	request := &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			Seq: 1,
		},
		FuncCode: al.FuncWrite,
		// Group=12 (CROB), Variation=1, Qualifier=0, Count=1, Index=0, CROB value=11 bytes
		Data: []byte{0x0C, 0x01, 0x00, 0x01, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}

	// Process request
	response, err := ost.ProcessRequest(request)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.FuncCode != al.FuncResponse {
		t.Errorf("Expected FuncCode RESPONSE (0), got %d", response.FuncCode)
	}

	t.Logf("WRITE request processed successfully")
}

// TestOutstationProcessEnableUnsolicited tests ENABLE UNSOLICITED processing.
func TestOutstationProcessEnableUnsolicited(t *testing.T) {
	ost := outstation.NewOutstation(nil)
	ost.Initialize()
	ost.Start()

	// Create ENABLE UNSOLICITED request
	request := &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			Seq: 1,
		},
		FuncCode: al.FuncEnableUnsolicited,
		Data:     nil,
	}

	// Process request
	response, err := ost.ProcessRequest(request)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.FuncCode != al.FuncResponse {
		t.Errorf("Expected FuncCode RESPONSE (0), got %d", response.FuncCode)
	}

	t.Logf("ENABLE UNSOLICITED processed successfully")
}

// TestOutstationProcessUnsupportedRequest tests unsupported function code handling.
func TestOutstationProcessUnsupportedRequest(t *testing.T) {
	ost := outstation.NewOutstation(nil)
	ost.Initialize()
	ost.Start()

	// Create an unsupported request (FILE_OPEN = 13)
	request := &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			Seq: 1,
		},
		FuncCode: al.FuncFileOpen,
		Data:     nil,
	}

	// Process request - should return error
	_, err := ost.ProcessRequest(request)
	if err == nil {
		t.Error("Expected error for unsupported function code")
	}

	if err != nil && err.Error() != "unsupported function code: 13" {
		t.Errorf("Unexpected error: %v", err)
	}

	t.Logf("Unsupported function code handled correctly")
}

// TestOutstationStateTransitions tests state machine transitions.
func TestOutstationStateTransitions(t *testing.T) {
	ost := outstation.NewOutstation(nil)

	// Initial state
	if ost.State() != outstation.StateDown {
		t.Errorf("Expected initial state Down, got %s", ost.State())
	}

	// Initialize
	if err := ost.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	if ost.State() != outstation.StateInitialized {
		t.Errorf("Expected state Initialized, got %s", ost.State())
	}

	// Start
	if err := ost.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if ost.State() != outstation.StateOperational {
		t.Errorf("Expected state Operational, got %s", ost.State())
	}

	// Stop
	if err := ost.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if ost.State() != outstation.StateDown {
		t.Errorf("Expected state Down, got %s", ost.State())
	}
}

// TestOutstationIIN tests Internal Indication handling.
func TestOutstationIIN(t *testing.T) {
	ost := outstation.NewOutstation(nil)
	ost.Initialize()
	ost.Start()

	iin := ost.IIN()

	// Should have no errors initially
	if iin.AllStop || iin.Busy || iin.ParamUnavail {
		t.Error("Expected clean IIN on startup")
	}
}

// TestOutstationDefaultDataHandler tests the default data handler.
func TestOutstationDefaultDataHandler(t *testing.T) {
	ost := outstation.NewOutstation(nil)
	ost.Initialize()
	ost.Start()

	// Request all data
	request := &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			Seq: 1,
		},
		FuncCode: al.FuncRead,
		Data:     []byte{60, 1, 0x07, 0x00}, // Group 60, all data
	}

	response, err := ost.ProcessRequest(request)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	// Should have substantial data (Binary + Analog + Counter)
	if len(response.Data) < 20 {
		t.Errorf("Expected substantial data, got %d bytes", len(response.Data))
	}

	t.Logf("Default data handler working: %d bytes returned", len(response.Data))
}

// TestOutstationCustomDataHandler tests custom data handler.
func TestOutstationCustomDataHandler(t *testing.T) {
	// Create custom data handler
	customData := &CustomDataHandler{
		binaryInputs: []outstation.BinaryInput{
			{Value: true, Quality: outstation.BinaryQualityOnline},
		},
		analogInputs: []outstation.AnalogInput{
			{Value: 999.5, Quality: outstation.AnalogQualityOnline},
		},
		counters: []outstation.Counter{
			{Value: 99999, Quality: outstation.CounterQualityOnline},
		},
	}

	ost := outstation.NewOutstation(nil)
	ost.SetDataHandler(customData)
	ost.Initialize()
	ost.Start()

	// Request all data
	request := &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			Seq: 1,
		},
		FuncCode: al.FuncRead,
		Data:     []byte{60, 1, 0x07, 0x00},
	}

	response, err := ost.ProcessRequest(request)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	t.Logf("Custom data handler working: %d bytes returned", len(response.Data))
}

// CustomDataHandler implements outstation.DataHandler with custom data.
type CustomDataHandler struct {
	binaryInputs   []outstation.BinaryInput
	analogInputs   []outstation.AnalogInput
	counters       []outstation.Counter
	frozenCounters []outstation.Counter
}

func (c *CustomDataHandler) GetBinaryInputs() []outstation.BinaryInput {
	return c.binaryInputs
}

func (c *CustomDataHandler) GetAnalogInputs() []outstation.AnalogInput {
	return c.analogInputs
}

func (c *CustomDataHandler) GetCounters() []outstation.Counter {
	return c.counters
}

func (c *CustomDataHandler) GetFrozenCounters() []outstation.Counter {
	return c.frozenCounters
}

func (c *CustomDataHandler) FreezeCounters(clear bool) error {
	return nil
}
