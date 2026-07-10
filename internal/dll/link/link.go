// Package link implements the DNP3 Data Link Layer state machine.
//
// The Data Link Layer operates in either:
// - Unbalanced mode: Master is always Primary, Outstation is always Secondary
// - Balanced mode: Either device can be Primary
//
// This implementation supports both modes with proper state tracking.
//
// Reference: IEEE 1815-2012 Section 5.4
package link

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"dnp3/internal/dll/frame"
)

// State represents the current state of a link endpoint.
type State int

const (
	StateLinkDown State = iota
	StateLinkReset
	StateOperational
	StateWaitingConfirm
	StateWaitingResponse
	StateError
)

// String returns a human-readable state name.
func (s State) String() string {
	switch s {
	case StateLinkDown:
		return "LinkDown"
	case StateLinkReset:
		return "LinkReset"
	case StateOperational:
		return "Operational"
	case StateWaitingConfirm:
		return "WaitingConfirm"
	case StateWaitingResponse:
		return "WaitingResponse"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// Role represents whether this endpoint is Primary or Secondary.
type Role int

const (
	RoleUnknown Role = iota
	RolePrimary
	RoleSecondary
)

// String returns a human-readable role name.
func (r Role) String() string {
	switch r {
	case RoleUnknown:
		return "Unknown"
	case RolePrimary:
		return "Primary"
	case RoleSecondary:
		return "Secondary"
	default:
		return "Unknown"
	}
}

// Config holds configuration for the link state machine.
type Config struct {
	// Local address of this endpoint
	LocalAddr uint16

	// Remote address of the peer
	RemoteAddr uint16

	// Role: Primary or Secondary
	Role Role

	// Timeout for confirmations
	ConfirmTimeout time.Duration

	// Timeout for responses
	ResponseTimeout time.Duration

	// Maximum number of retries
	MaxRetries int

	// Use unbalanced mode (Master-Outstation pattern)
	Unbalanced bool
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig(localAddr, remoteAddr uint16, role Role) Config {
	return Config{
		LocalAddr:        localAddr,
		RemoteAddr:       remoteAddr,
		Role:             role,
		ConfirmTimeout:   5 * time.Second,
		ResponseTimeout:   10 * time.Second,
		MaxRetries:       3,
		Unbalanced:       role == RolePrimary, // Primary is always unbalanced
	}
}

// Event represents events that can occur in the state machine.
type Event int

const (
	EventSendData Event = iota
	EventRecvFrame
	EventTimeout
	EventConfirm
	EventResponse
	EventNack
	EventError
	EventReset
	EventClose
)

// StateMachine implements the DNP3 Data Link Layer state machine.
type StateMachine struct {
	config Config

	mu sync.Mutex

	// Current state
	state State

	// FCB state for balanced mode
	fcb        bool
	expectFCB  bool

	// Retry count for current operation
	retries int

	// Sequence number for tracking
	sequence uint8

	// Error state
	lastError error

	// Channels for communication
	frameCh chan *frame.Frame
	errCh   chan error
	doneCh  chan struct{}

	// Statistics
	framesSent     atomic.Uint64
	framesReceived atomic.Uint64
	errors         atomic.Uint64
}

// NewStateMachine creates a new link state machine.
func NewStateMachine(config Config) *StateMachine {
	return &StateMachine{
		config:   config,
		state:    StateLinkDown,
		frameCh:  make(chan *frame.Frame, 100),
		errCh:    make(chan error, 10),
		doneCh:   make(chan struct{}),
	}
}

// Start begins the state machine operation.
func (sm *StateMachine) Start(ctx context.Context) error {
	sm.mu.Lock()

	switch sm.state {
	case StateLinkDown:
		sm.state = StateLinkReset
		sm.mu.Unlock()
		return nil
	default:
		sm.mu.Unlock()
		return fmt.Errorf("state machine already started")
	}
}

// ResetLinkStations sends a Reset Link Stations frame.
func (sm *StateMachine) ResetLinkStations(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.state != StateLinkReset && sm.state != StateOperational {
		return fmt.Errorf("cannot reset from state %s", sm.state)
	}

	f := &frame.Frame{
		Control: frame.Control{
			DIR:      true,
			PRM:      true,
			FCB:      false,
			FCV:      false,
			FuncCode: frame.FuncResetLinkStations,
		},
		DestAddr: sm.config.RemoteAddr,
		SrcAddr:  sm.config.LocalAddr,
		Data:     nil,
	}

	return sm.sendFrame(ctx, f)
}

// SendData sends user data with optional confirmation.
func (sm *StateMachine) SendData(ctx context.Context, data []byte, confirm bool) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.state != StateOperational {
		return fmt.Errorf("cannot send data from state %s", sm.state)
	}

	ctrl := frame.Control{
		DIR:      true,
		PRM:      true,
		FuncCode: frame.FuncConfirmedUserData,
		FCV:      confirm,
		FCB:      sm.fcb,
	}

	f := &frame.Frame{
		Control: ctrl,
		DestAddr: sm.config.RemoteAddr,
		SrcAddr:  sm.config.LocalAddr,
		Data:     data,
	}

	if confirm {
		sm.fcb = !sm.fcb // Toggle FCB for next confirmed frame
		sm.state = StateWaitingConfirm
	}

	return sm.sendFrame(ctx, f)
}

// sendFrame encodes and sends a frame.
func (sm *StateMachine) sendFrame(ctx context.Context, f *frame.Frame) error {
	data, err := frame.Encode(f)
	if err != nil {
		sm.errors.Add(1)
		sm.lastError = err
		return err
	}

	// In a real implementation, this would write to a connection
	// For now, we just update statistics
	sm.framesSent.Add(1)

	// Simulate sending (in real implementation, this would go to connection)
	select {
	case sm.frameCh <- f:
	case <-ctx.Done():
		return ctx.Err()
	case <-sm.doneCh:
		return fmt.Errorf("state machine closed")
	}

	return nil
}

// HandleFrame processes an incoming frame.
func (sm *StateMachine) HandleFrame(f *frame.Frame) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.framesReceived.Add(1)

	switch sm.config.Role {
	case RolePrimary:
		return sm.handlePrimaryFrame(f)
	case RoleSecondary:
		return sm.handleSecondaryFrame(f)
	default:
		return fmt.Errorf("unknown role")
	}
}

// handlePrimaryFrame handles frames received by a Primary station.
func (sm *StateMachine) handlePrimaryFrame(f *frame.Frame) error {
	if !f.Control.PRM {
		// This is a response from Secondary
		switch f.Control.FuncCode {
		case frame.FuncAck:
			if sm.state == StateWaitingConfirm {
				sm.state = StateOperational
				return nil
			}
			return fmt.Errorf("unexpected ACK in state %s", sm.state)

		case frame.FuncNack:
			if sm.state == StateWaitingConfirm {
				sm.retries++
				if sm.retries >= sm.config.MaxRetries {
					sm.state = StateError
					sm.lastError = fmt.Errorf("max retries exceeded")
					return sm.lastError
				}
				// Will retry in Operational state
				sm.state = StateOperational
				return nil
			}

		case frame.FuncLinkStatus:
			// Response received
			if sm.state == StateWaitingResponse {
				sm.state = StateOperational
				return nil
			}
		}
	}

	return fmt.Errorf("unexpected frame in primary mode: %s", f)
}

// handleSecondaryFrame handles frames received by a Secondary station.
func (sm *StateMachine) handleSecondaryFrame(f *frame.Frame) error {
	if f.Control.PRM {
		// This is a request from Primary
		switch f.Control.FuncCode {
		case frame.FuncResetLinkStations:
			sm.state = StateOperational
			// Send ACK
			return sm.sendAck()

		case frame.FuncConfirmedUserData:
			if !f.Control.FCV {
				return fmt.Errorf("confirmed data with FCV=0")
			}

			// Check FCB matches expected
			if f.Control.FCB != sm.expectFCB {
				// Duplicate or out-of-sequence
				// Echo the FCB back in NACK to request retransmission
				return sm.sendNack()
			}

			sm.expectFCB = !sm.expectFCB
			// Send response
			return sm.sendResponse(f.Data)

		case frame.FuncReturnLinkStatus:
			// Send link status response
			return sm.sendLinkStatus()
		}
	}

	return fmt.Errorf("unexpected frame in secondary mode: %s", f)
}

// sendAck sends an ACK response.
func (sm *StateMachine) sendAck() error {
	f := &frame.Frame{
		Control: frame.Control{
			DIR:      false,
			PRM:      false,
			FuncCode: frame.FuncAck,
		},
		DestAddr: sm.config.RemoteAddr,
		SrcAddr:  sm.config.LocalAddr,
	}

	data, err := frame.Encode(f)
	if err != nil {
		return err
	}

	sm.framesSent.Add(1)
	_ = data // In real implementation, would write to connection
	return nil
}

// sendNack sends a NACK response.
func (sm *StateMachine) sendNack() error {
	f := &frame.Frame{
		Control: frame.Control{
			DIR:      false,
			PRM:      false,
			FuncCode: frame.FuncNack,
			DFC:      true, // Data link busy
		},
		DestAddr: sm.config.RemoteAddr,
		SrcAddr:  sm.config.LocalAddr,
	}

	data, err := frame.Encode(f)
	if err != nil {
		return err
	}

	sm.framesSent.Add(1)
	_ = data
	return nil
}

// sendResponse sends a confirmed user data response.
func (sm *StateMachine) sendResponse(data []byte) error {
	f := &frame.Frame{
		Control: frame.Control{
			DIR:      false,
			PRM:      false,
			FCB:      sm.expectFCB, // Echo FCB
			FCV:      true,
			FuncCode: frame.FuncConfirmedUserDataR,
		},
		DestAddr: sm.config.RemoteAddr,
		SrcAddr:  sm.config.LocalAddr,
		Data:     data,
	}

	encoded, err := frame.Encode(f)
	if err != nil {
		return err
	}

	sm.framesSent.Add(1)
	_ = encoded
	return nil
}

// sendLinkStatus sends a link status response.
func (sm *StateMachine) sendLinkStatus() error {
	f := &frame.Frame{
		Control: frame.Control{
			DIR:      false,
			PRM:      false,
			FuncCode: frame.FuncLinkStatus,
		},
		DestAddr: sm.config.RemoteAddr,
		SrcAddr:  sm.config.LocalAddr,
	}

	data, err := frame.Encode(f)
	if err != nil {
		return err
	}

	sm.framesSent.Add(1)
	_ = data
	return nil
}

// Close shuts down the state machine.
func (sm *StateMachine) Close() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	select {
	case <-sm.doneCh:
		return nil // Already closed
	default:
	}

	sm.state = StateLinkDown
	close(sm.doneCh)
	close(sm.frameCh)
	close(sm.errCh)

	return nil
}

// State returns the current state.
func (sm *StateMachine) State() State {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.state
}

// Stats returns current statistics.
func (sm *StateMachine) Stats() Stats {
	return Stats{
		FramesSent:     sm.framesSent.Load(),
		FramesReceived: sm.framesReceived.Load(),
		Errors:         sm.errors.Load(),
		State:          sm.state,
	}
}

// Stats holds state machine statistics.
type Stats struct {
	FramesSent     uint64
	FramesReceived uint64
	Errors        uint64
	State         State
}

// FrameChan returns the channel for received frames.
func (sm *StateMachine) FrameChan() <-chan *frame.Frame {
	return sm.frameCh
}

// Done returns a channel that's closed when the state machine is closed.
func (sm *StateMachine) Done() <-chan struct{} {
	return sm.doneCh
}
