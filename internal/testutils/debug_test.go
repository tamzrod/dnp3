package testutils

import (
    "context"
    "testing"
    "time"
    "fmt"
    
    "dnp3/internal/master"
    "dnp3/internal/outstation"
)

func TestDebugMasterOutstation(t *testing.T) {
    _, masterTransport, outstationTransport := NewBidirectionalTransport()

    masterConfig := &master.Config{
        MasterAddress:     1,
        Timeout:           1000,
        MaxRetries:        1,
    }
    outstationConfig := &outstation.Config{
        MasterAddress:     1,
        SBOTimeout:        5000,
        MaxEventBuffers:   1000,
    }

    m := master.NewMaster(masterConfig)
    o := outstation.NewOutstation(outstationConfig)

    m.SetTransport(masterTransport)
    o.SetTransport(outstationTransport)

    o.Initialize()
    o.Start()

    done := make(chan error, 1)
    ctx, cancel := context.WithCancel(context.Background())
    go func() {
        fmt.Println("Outstation goroutine starting...")
        err := o.RunWithContext(ctx)
        fmt.Printf("Outstation goroutine finished with error: %v\n", err)
        done <- err
    }()
    defer o.Stop()

    err := m.Connect()
    if err != nil {
        t.Fatalf("Connect failed: %v", err)
    }

    err = m.Initialize()
    if err != nil {
        t.Fatalf("Initialize failed: %v", err)
    }

    // Wait a bit for the outstation to start
    time.Sleep(500 * time.Millisecond)

    // Try to read - this should trigger communication
    fmt.Println("Calling ReadBinaryInputs...")
    err = m.ReadBinaryInputs(1024, 0, 5)
    fmt.Printf("ReadBinaryInputs returned: %v\n", err)
    if err != nil {
        t.Errorf("ReadBinaryInputs failed: %v", err)
    }

    cancel()
    time.Sleep(100 * time.Millisecond)
}
