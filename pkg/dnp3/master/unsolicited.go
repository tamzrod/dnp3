package master

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dnp3/internal/al"
	"dnp3/internal/master"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
	"dnp3/pkg/transport"
)

// classEnableHeader is Group 60 Variations 2, 3, and 4, qualifier 0x06.
func classEnableHeader() []byte {
	return []byte{
		60, 2, al.QualAllObjects,
		60, 3, al.QualAllObjects,
		60, 4, al.QualAllObjects,
	}
}

func (c *client) sendClassFunction(ctx context.Context, funcCode uint8) error {
	if ctx == nil {
		ctx = context.Background()
	}
	c.mu.RLock()
	state := c.state
	c.mu.RUnlock()
	if state < dnp3.StateConnected || state == dnp3.StateError {
		return dnp3.ErrNotConnected
	}
	apdu := al.NewRequest(0, funcCode)
	apdu.Data = classEnableHeader()
	return c.roundTrip(ctx, apdu)
}

// ReadEvents implements Client.ReadEvents.
// Supported-profile: Reject by default — event objects are outside the v0
// supported profile. It returns ErrUnsupportedOption with no wire traffic
// unless the experimental event delivery switch is on.
func (c *client) ReadEvents(ctx context.Context) (*UnsolicitedResponse, error) {
	if !c.config.ExperimentalEventDelivery {
		return nil, fmt.Errorf("%w: event read (ReadEvents); the v0 supported profile does not read event objects", dnp3.ErrUnsupportedOption)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	c.mu.RLock()
	state := c.state
	c.mu.RUnlock()
	if state < dnp3.StateConnected || state == dnp3.StateError {
		return nil, dnp3.ErrNotConnected
	}
	apdu := al.NewRequest(0, al.FuncRead)
	apdu.Data = classEnableHeader()
	data, err := c.roundTripData(ctx, apdu)
	if err != nil {
		return nil, err
	}
	resp, err := al.DecodeResponse(data)
	if err != nil {
		return nil, fmt.Errorf("decode event read: %w", err)
	}
	evs, err := al.DecodeEventObjects(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("decode event objects: %w", err)
	}
	return eventsToResponse(resp.IIN, evs), nil
}

func (c *client) roundTrip(ctx context.Context, apdu *al.APDU) error {
	_, err := c.roundTripData(ctx, apdu)
	return err
}

func (c *client) roundTripData(ctx context.Context, apdu *al.APDU) ([]byte, error) {
	deadline := time.Now().Add(c.config.Timeout)
	if dl, ok := ctx.Deadline(); ok && dl.Before(deadline) {
		deadline = dl
	}
	type result struct {
		data []byte
		err  error
	}
	done := make(chan result, 1)
	go func() {
		data, err := c.internalMaster.SendRequestWithRetryAndGetResponseDeadline(apdu, c.config.OutstationAddress, deadline)
		done <- result{data: data, err: err}
	}()
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("%w: %v", dnp3.ErrContextCanceled, ctx.Err())
	case res := <-done:
		if res.err != nil {
			if master.IsDisconnectError(res.err) {
				c.markTransportDisconnected()
			}
			return nil, wrapInternalError("request failed", res.err)
		}
		return res.data, nil
	}
}

// WatchUnsolicited implements Client.WatchUnsolicited.
// Supported-profile: Reject by default — the v0 supported profile does not
// deliver unsolicited responses. It returns ErrUnsupportedOption with no wire
// traffic unless the experimental event delivery switch is on.
func (c *client) WatchUnsolicited(ctx context.Context) error {
	if !c.config.ExperimentalEventDelivery {
		return fmt.Errorf("%w: unsolicited watch (WatchUnsolicited); the v0 supported profile does not deliver unsolicited responses", dnp3.ErrUnsupportedOption)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	c.mu.RLock()
	state := c.state
	c.mu.RUnlock()
	if state < dnp3.StateConnected || state == dnp3.StateError {
		return dnp3.ErrNotConnected
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := c.internalMaster.PollUnsolicited(c.config.OutstationAddress, 200*time.Millisecond)
		if err == nil {
			continue
		}
		if errors.Is(err, transport.ErrTimeout) || errors.Is(err, master.ErrTimeout) || errors.Is(err, context.DeadlineExceeded) {
			continue
		}
		if master.IsDisconnectError(err) {
			c.markTransportDisconnected()
			return err
		}
		return err
	}
}

func (c *client) deliverUnsolicitedApp(app []byte) {
	resp, err := al.DecodeResponse(app)
	if err != nil {
		return
	}
	evs, err := al.DecodeEventObjects(resp.Data)
	if err != nil {
		return
	}
	ur := eventsToResponse(resp.IIN, evs)
	c.mu.RLock()
	handlers := append([]UnsolicitedHandler(nil), c.handlers...)
	c.mu.RUnlock()
	for _, h := range handlers {
		if h != nil {
			h(ur)
		}
	}
}

func eventsToResponse(iin al.IIN, evs []al.EventObject) *UnsolicitedResponse {
	ur := &UnsolicitedResponse{
		IIN:       iin.Bytes(),
		Timestamp: time.Now(),
	}
	for _, ev := range evs {
		var ts *types.Timestamp
		if ev.HasTime {
			ts = types.NewTimestamp(ev.Time)
		}
		switch ev.Group {
		case 2:
			ur.BinaryInputs = append(ur.BinaryInputs, &types.BinaryInput{
				Index:   ev.Index,
				Value:   ev.Flags&al.BinEventState != 0,
				Quality: types.QualityFlags(ev.Flags &^ al.BinEventState),
				Time:    ts,
			})
		case 32:
			ur.AnalogInputs = append(ur.AnalogInputs, &types.AnalogInput{
				Index:   ev.Index,
				Value:   float64(ev.Analog),
				Quality: types.QualityFlags(ev.Flags),
				Time:    ts,
			})
		case 22:
			ur.Counters = append(ur.Counters, &types.Counter{
				Index:   ev.Index,
				Value:   ev.Counter,
				Quality: types.QualityFlags(ev.Flags),
				Time:    ts,
			})
		}
	}
	return ur
}
