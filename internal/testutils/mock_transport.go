package testutils

import (
	"container/list"
	"sync"
	"time"
)

// MockTransport implements an in-memory transport for testing.
// It allows a Master and Outstation to communicate without network I/O.
type MockTransport struct {
	mu         sync.Mutex
	sendQueue  *list.List
	timeoutMs  int
	closed     bool
	latency    time.Duration
	packetLoss float64
}

// NewMockTransport creates a new mock transport.
func NewMockTransport() *MockTransport {
	return &MockTransport{
		sendQueue: list.New(),
		timeoutMs: 1000,
		latency:   0,
		packetLoss: 0,
	}
}

// SetLatency sets simulated network latency.
func (t *MockTransport) SetLatency(d time.Duration) {
	t.latency = d
}

// SetPacketLoss sets simulated packet loss probability (0.0-1.0).
func (t *MockTransport) SetPacketLoss(p float64) {
	t.packetLoss = p
}

// Send queues data for the other side to receive.
func (t *MockTransport) Send(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return ErrTransportClosed
	}

	// Simulate packet loss
	if t.packetLoss > 0 && (float64(time.Now().UnixNano())/1e9) < t.packetLoss {
		return nil // "Packet lost"
	}

	// Queue the data
	t.sendQueue.PushBack(data)
	return nil
}

// Receive retrieves queued data, blocking until available or timeout.
func (t *MockTransport) Receive() ([]byte, error) {
	t.mu.Lock()

	if t.closed {
		t.mu.Unlock()
		return nil, ErrTransportClosed
	}

	// Wait for data
	if t.sendQueue.Len() == 0 {
		t.mu.Unlock()
		
		// Wait with timeout
		timeout := time.Duration(t.timeoutMs) * time.Millisecond
		deadline := time.Now().Add(timeout)
		
		for time.Now().Before(deadline) {
			t.mu.Lock()
			if t.sendQueue.Len() > 0 {
				break
			}
			t.mu.Unlock()
			time.Sleep(10 * time.Millisecond)
		}
		
		t.mu.Lock()
		if t.sendQueue.Len() == 0 {
			t.mu.Unlock()
			return nil, ErrTimeout
		}
	}

	// Get the data
	elem := t.sendQueue.Front()
	data := elem.Value.([]byte)
	t.sendQueue.Remove(elem)
	t.mu.Unlock()

	// Apply simulated latency
	if t.latency > 0 {
		time.Sleep(t.latency)
	}

	return data, nil
}

// SetTimeout sets the receive timeout.
func (t *MockTransport) SetTimeout(ms int) {
	t.timeoutMs = ms
}

// Close closes the transport.
func (t *MockTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	t.sendQueue.Init()
	return nil
}

// IsEmpty returns true if no data is queued.
func (t *MockTransport) IsEmpty() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sendQueue.Len() == 0
}

// QueueSize returns the number of queued messages.
func (t *MockTransport) QueueSize() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sendQueue.Len()
}

// Errors for transport operations
var (
	ErrTransportClosed = &TransportError{"transport closed"}
	ErrTimeout         = &TransportError{"receive timeout"}
)

// TransportError represents a transport error.
type TransportError struct {
	msg string
}

func (e *TransportError) Error() string {
	return e.msg
}

// MockTransportPair creates a connected pair of transports for bidirectional communication.
type MockTransportPair struct {
	MasterToOutstation *MockTransport
	OutstationToMaster *MockTransport
}

// NewMockTransportPair creates a bidirectional transport pair.
func NewMockTransportPair() *MockTransportPair {
	return &MockTransportPair{
		MasterToOutstation: NewMockTransport(),
		OutstationToMaster: NewMockTransport(),
	}
}

// BidirectionalTransport wraps a MockTransportPair and provides
// separate send/receive interfaces for each side.
type BidirectionalTransport struct {
	pair *MockTransportPair
}

// MasterTransport is the transport used by the Master side.
type MasterTransport struct {
	send *MockTransport // Data going to Outstation
	recv *MockTransport // Data coming from Outstation
}

// OutstationTransport is the transport used by the Outstation side.
type OutstationTransport struct {
	send *MockTransport // Data going to Master
	recv *MockTransport // Data coming from Master
}

// NewBidirectionalTransport creates a new bidirectional transport.
func NewBidirectionalTransport() (*BidirectionalTransport, *MasterTransport, *OutstationTransport) {
	pair := NewMockTransportPair()
	
	bt := &BidirectionalTransport{pair: pair}
	
	masterTransport := &MasterTransport{
		send: pair.MasterToOutstation,
		recv: pair.OutstationToMaster,
	}
	
	outstationTransport := &OutstationTransport{
		send: pair.OutstationToMaster,
		recv: pair.MasterToOutstation,
	}
	
	return bt, masterTransport, outstationTransport
}

// Send sends data (from Master to Outstation direction).
func (t *MasterTransport) Send(data []byte) error {
	return t.send.Send(data)
}

// Receive receives data (from Outstation direction).
func (t *MasterTransport) Receive() ([]byte, error) {
	return t.recv.Receive()
}

// SetTimeout sets the receive timeout.
func (t *MasterTransport) SetTimeout(ms int) {
	t.recv.SetTimeout(ms)
}

// Send sends data (from Outstation to Master direction).
func (t *OutstationTransport) Send(data []byte) error {
	return t.send.Send(data)
}

// Receive receives data (from Master direction).
func (t *OutstationTransport) Receive() ([]byte, error) {
	return t.recv.Receive()
}

// SetTimeout sets the receive timeout.
func (t *OutstationTransport) SetTimeout(ms int) {
	t.recv.SetTimeout(ms)
}
