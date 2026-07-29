// Package master provides the DNP3 Master client API.
//
// This package implements the Master Station role of the DNP3 protocol.
// A Master station initiates all communication with outstations.
//
// # Usage
//
// Create a new client:
//
//	config := master.NewConfig().
//		WithAddress(1024).
//		WithTransport(dnp3.TCP, "192.168.1.100", 20000)
//
//	client, err := master.NewClient(config)
//
// Connect and read data:
//
//	ctx := context.Background()
//	if err := client.Connect(ctx); err != nil {
//		log.Fatal(err)
//	}
//	defer client.Disconnect(ctx)
//
//	resp, err := client.Read(ctx, &types.ReadRequest{
//		Groups: types.ReadAllStatic,
//	})
package master

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"math"
	"sync"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/internal/al"
	"dnp3/internal/master"
	"dnp3/pkg/dnp3/types"
	"dnp3/pkg/transport"
)

// Client represents a DNP3 Master client.
// The client connects to outstations and issues read/write/control commands.
type Client interface {
	// Connect establishes a connection to the outstation
	Connect(ctx context.Context) error

	// Disconnect closes the connection
	Disconnect(ctx context.Context) error

	// State returns the current connection state
	State() dnp3.ConnectionState

	// Read issues a read request to the outstation
	Read(ctx context.Context, request *types.ReadRequest) (*ReadResponse, error)

	// Operate issues a control command to the outstation
	Operate(ctx context.Context, command *types.ControlOutput) (*OperateResponse, error)

	// EnableUnsolicited enables unsolicited responses from the outstation
	EnableUnsolicited(ctx context.Context) error

	// DisableUnsolicited disables unsolicited responses
	DisableUnsolicited(ctx context.Context) error

	// SetUnsolicitedHandler sets the handler for unsolicited responses
	SetUnsolicitedHandler(handler UnsolicitedHandler)

	// Close gracefully shuts down the client
	Close() error
}

// Config holds Master client configuration
type Config struct {
	// MasterAddress is this master's DNP3 address
	MasterAddress uint16
	// OutstationAddress is the target outstation's DNP3 address
	OutstationAddress uint16
	// TransportType specifies TCP or TLS
	TransportType dnp3.TransportType
	// Address is the network address (IP or hostname)
	Address string
	// Port is the network port
	Port int
	// TLSConfig is the TLS configuration (required for TLS transport)
	TLSConfig *tls.Config
	// Timeout is the response timeout
	Timeout time.Duration
	// RetryCount is the maximum number of retries
	RetryCount int
	// RetryDelay is the delay between retries
	RetryDelay time.Duration
	// KeepAliveInterval is the keep-alive interval
	KeepAliveInterval time.Duration
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		MasterAddress:     0xFFFF,
		OutstationAddress: 1024,
		TransportType:     dnp3.TCP,
		Address:           "localhost",
		Port:              20000,
		Timeout:            5 * time.Second,
		RetryCount:         3,
		RetryDelay:         100 * time.Millisecond,
		KeepAliveInterval:  30 * time.Second,
	}
}

// ConfigOption is a functional option for configuring the client
type ConfigOption func(*Config)

// WithMasterAddress sets the master address
func WithMasterAddress(addr uint16) ConfigOption {
	return func(c *Config) {
		c.MasterAddress = addr
	}
}

// WithOutstationAddress sets the outstation address
func WithOutstationAddress(addr uint16) ConfigOption {
	return func(c *Config) {
		c.OutstationAddress = addr
	}
}

// WithTransport sets the transport type and network address
func WithTransport(t dnp3.TransportType, address string, port int) ConfigOption {
	return func(c *Config) {
		c.TransportType = t
		c.Address = address
		c.Port = port
	}
}

// WithTLS sets the TLS configuration
func WithTLS(config *tls.Config) ConfigOption {
	return func(c *Config) {
		c.TLSConfig = config
		c.TransportType = dnp3.TLS
	}
}

// WithTimeout sets the response timeout
func WithTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithRetry sets the retry count and delay
func WithRetry(count int, delay time.Duration) ConfigOption {
	return func(c *Config) {
		c.RetryCount = count
		c.RetryDelay = delay
	}
}

// NewConfig creates a new configuration with options applied
func NewConfig(opts ...ConfigOption) *Config {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Timeout <= 0 {
		return &dnp3.ConfigurationError{
			Field:   "Timeout",
			Message: "must be positive",
		}
	}
	if c.RetryCount < 0 {
		return &dnp3.ConfigurationError{
			Field:   "RetryCount",
			Message: "must be non-negative",
		}
	}
	if c.RetryDelay < 0 {
		return &dnp3.ConfigurationError{
			Field:   "RetryDelay",
			Message: "must be non-negative",
		}
	}
	if c.TransportType == dnp3.TLS && c.TLSConfig == nil {
		return &dnp3.ConfigurationError{
			Field:   "TLSConfig",
			Message: "required for TLS transport",
		}
	}
	return nil
}

// UnsolicitedHandler is called when an unsolicited response is received
type UnsolicitedHandler func(response *UnsolicitedResponse)

// ReadResponse contains the response data from a read request
type ReadResponse struct {
	// IIN is the Internal Indication from the outstation
	IIN [2]byte
	// BinaryInputs contains binary input data
	BinaryInputs []*types.BinaryInput
	// AnalogInputs contains analog input data
	AnalogInputs []*types.AnalogInput
	// Counters contains counter data
	Counters []*types.Counter
	// FrozenCounters contains frozen counter data
	FrozenCounters []*types.FrozenCounter
	// Timestamp is when the response was received
	Timestamp time.Time
}

// OperateResponse contains the response from a control operation
type OperateResponse struct {
	// IIN is the Internal Indication from the outstation
	IIN [2]byte
	// Status is the command status
	Status types.ControlStatus
	// Timestamp is when the response was received
	Timestamp time.Time
}

// UnsolicitedResponse contains data from an unsolicited response
type UnsolicitedResponse struct {
	// IIN is the Internal Indication from the outstation
	IIN [2]byte
	// BinaryInputs contains binary input event data
	BinaryInputs []*types.BinaryInput
	// AnalogInputs contains analog input event data
	AnalogInputs []*types.AnalogInput
	// Counters contains counter event data
	Counters []*types.Counter
	// Timestamp is when the response was received
	Timestamp time.Time
}

// transportAdapter wraps the public transport.Handler to satisfy internal/master.TransportHandler
type transportAdapter struct {
	transport.Handler
}

// client is the internal implementation of Client
type client struct {
	config   *Config
	state    dnp3.ConnectionState
	mu       sync.RWMutex
	handlers []UnsolicitedHandler

	// Internal implementation
	internalMaster *master.Master
	transport      transport.Handler

	// For response parsing
	sequence uint8
}

// sendAndReceive sends a request and receives a response, returning parsed data
func (c *client) sendAndReceive(data []byte) ([]byte, error) {
	// Send request
	if err := c.transport.Send(data); err != nil {
		return nil, fmt.Errorf("send failed: %w", err)
	}

	// Receive response with timeout
	c.transport.SetTimeout(int(c.config.Timeout.Milliseconds()))
	resp, err := c.transport.Receive()
	if err != nil {
		return nil, fmt.Errorf("receive failed: %w", err)
	}

	return resp, nil
}

// NewClient creates a new Master client with the given configuration
func NewClient(config *Config) (Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Create internal master
	internalConfig := &master.Config{
		MasterAddress: config.MasterAddress,
		Timeout:       int(config.Timeout.Milliseconds()),
		MaxRetries:    config.RetryCount,
		RetryDelay:    int(config.RetryDelay.Milliseconds()),
	}
	internal := master.NewMaster(internalConfig)

	// Create transport based on type
	var t transport.Handler
	switch config.TransportType {
	case dnp3.TCP:
		tcpConfig := &transport.TCPConfig{
			Address:         config.Address,
			Port:            config.Port,
			ConnectTimeout:  int(config.Timeout.Milliseconds()),
			ReceiveTimeout: int(config.Timeout.Milliseconds()),
			KeepAlive:      int(config.KeepAliveInterval.Milliseconds()),
		}
		t = transport.NewTCPTransport(tcpConfig)
	case dnp3.TLS:
		// For TLS, we would need to create a TLSTransport
		// For now, fall back to TCP
		tcpConfig := &transport.TCPConfig{
			Address:         config.Address,
			Port:            config.Port,
			ConnectTimeout:  int(config.Timeout.Milliseconds()),
			ReceiveTimeout: int(config.Timeout.Milliseconds()),
			KeepAlive:      int(config.KeepAliveInterval.Milliseconds()),
		}
		t = transport.NewTCPTransport(tcpConfig)
	}

	// Set up internal master with transport
	internal.SetTransport(&transportAdapter{Handler: t})

	return &client{
		config:         config,
		state:          dnp3.StateDisconnected,
		handlers:       make([]UnsolicitedHandler, 0),
		internalMaster:  internal,
		transport:      t,
	}, nil
}

// Connect implements Client.Connect
//
// Proper order:
// 1. Add outstation (needed for handshake destinations)
// 2. Transport connect (TCP)
// 3. Internal master connect (handshake with registered outstations)
func (c *client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != dnp3.StateDisconnected {
		return fmt.Errorf("already connected: %s", c.state)
	}

	c.state = dnp3.StateConnecting

	// Step 1: Add outstation FIRST (needed for link handshake)
	c.internalMaster.AddOutstation(c.config.OutstationAddress, fmt.Sprintf("outstation-%d", c.config.OutstationAddress))

	// Step 2: Connect transport
	if err := c.transport.Connect(); err != nil {
		c.state = dnp3.StateError
		return fmt.Errorf("transport connect failed: %w", err)
	}

	// Step 3: Connect internal master (performs link-layer handshake)
	if err := c.internalMaster.Connect(); err != nil {
		c.state = dnp3.StateError
		return fmt.Errorf("master connect failed: %w", err)
	}

	// Step 4: Initialize master
	if err := c.internalMaster.Initialize(); err != nil {
		c.state = dnp3.StateError
		return fmt.Errorf("master initialize failed: %w", err)
	}

	c.state = dnp3.StateActive
	return nil
}

// Disconnect implements Client.Disconnect
func (c *client) Disconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == dnp3.StateDisconnected {
		return nil
	}

	// Disconnect internal master
	_ = c.internalMaster.Disconnect()

	// Close transport
	_ = c.transport.Close()

	c.state = dnp3.StateDisconnected
	return nil
}

// State implements Client.State
func (c *client) State() dnp3.ConnectionState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// Read implements Client.Read
func (c *client) Read(ctx context.Context, request *types.ReadRequest) (*ReadResponse, error) {
	if request == nil || len(request.Groups) == 0 {
		return nil, dnp3.ErrInvalidRequest
	}

	c.mu.RLock()
	state := c.state
	c.mu.RUnlock()

	if state < dnp3.StateConnected {
		return nil, dnp3.ErrNotConnected
	}

	// Wire to internal master for proper protocol handling
	outstationID := c.config.OutstationAddress

	// Build APDU using internal master for proper retry/timeout handling
	apdu := buildReadRequest(c.sequence, request)
	c.sequence = (c.sequence + 1) % 16

	// Use internal master's sendWithRetryAndGetResponse which:
	// 1. Sends the request with retry logic
	// 2. Waits for and processes the response
	// 3. Returns the processed application layer data
	respData, err := c.internalMaster.SendRequestWithRetryAndGetResponse(apdu, outstationID)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	// Decode response (respData is already transport/data-link decoded)
	resp, err := al.DecodeResponse(respData)
	if err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	// Parse data into public types
	binaryInputs := parseBinaryInputs(resp.Data)
	analogInputs := parseAnalogInputs(resp.Data)
	counters := parseCounters(resp.Data)

	return &ReadResponse{
		IIN:           resp.IIN.Bytes(),
		Timestamp:     time.Now(),
		BinaryInputs:  binaryInputs,
		AnalogInputs:  analogInputs,
		Counters:      counters,
	}, nil
}

// buildReadRequest builds an APDU for a read request
func buildReadRequest(seq uint8, request *types.ReadRequest) *al.APDU {
	// Build object headers from request groups
	var data []byte
	for _, group := range request.Groups {
		data = append(data, group.Group)
		data = append(data, group.Variation)
		data = append(data, 0x00) // Qualifier: all
		data = append(data, 0x07) // Range: 0x06=prefix, 0x01=count
	}

	return &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: false,
			Seq: seq,
		},
		FuncCode: al.FuncRead, // 0x02
		Data:     data,
	}
}

// parseBinaryInputs parses binary input data from response
func parseBinaryInputs(data []byte) []*types.BinaryInput {
	var result []*types.BinaryInput
	offset := 0

	for offset < len(data) {
		if offset+4 > len(data) {
			break
		}

		group := data[offset]
		variation := data[offset+1]
		_ = data[offset+2] // qualifier
		count := data[offset+3]
		offset += 4

		if group != 1 { // Group 1 = Binary Input
			// Skip past this group's data using correct group/variation sizing
			offset = skipGroupData(offset, data, group, variation, count)
			continue
		}

		for i := 0; i < int(count) && offset+3 <= len(data); i++ {
			index := binary.BigEndian.Uint16(data[offset:offset+2])
			offset += 2

			var value bool
			var quality types.QualityFlags

			switch variation {
			case 1: // With flags
				val := data[offset]
				value = (val & 0x01) != 0
				quality = types.QualityFlags(val & 0xE0)
				offset++
			case 2: // Without flags
				value = data[offset] != 0
				quality = types.QualityOnline
				offset++
			default:
				offset++
			}

			result = append(result, &types.BinaryInput{
				Index:   index,
				Value:   value,
				Quality: quality,
			})
		}
	}

	return result
}

// parseAnalogInputs parses analog input data from response
func parseAnalogInputs(data []byte) []*types.AnalogInput {
	var result []*types.AnalogInput
	offset := 0

	for offset < len(data) {
		if offset+4 > len(data) {
			break
		}

		group := data[offset]
		variation := data[offset+1]
		_ = data[offset+2] // qualifier
		count := data[offset+3]
		offset += 4

		if group != 30 { // Group 30 = Analog Input
			// Skip past this group's data
			offset = skipGroupData(offset, data, group, variation, count)
			continue
		}

		for i := 0; i < int(count) && offset+5 <= len(data); i++ {
			index := binary.BigEndian.Uint16(data[offset:offset+2])
			offset += 2

			var value float64
			var quality types.QualityFlags

			switch variation {
			case 1: // 32-bit float with flags
				bits := binary.BigEndian.Uint32(data[offset : offset+4])
				value = float64(math.Float32frombits(bits))
				offset += 4
				quality = types.QualityFlags(data[offset])
				offset++
			case 2: // 16-bit integer with flags
				val := int16(binary.BigEndian.Uint16(data[offset:offset+2]))
				value = float64(val)
				offset += 2
				quality = types.QualityFlags(data[offset])
				offset++
			case 3: // 32-bit integer with flags
				val := int32(binary.BigEndian.Uint32(data[offset : offset+4]))
				value = float64(val)
				offset += 4
				quality = types.QualityFlags(data[offset])
				offset++
			case 5: // 32-bit float without flags
				bits := binary.BigEndian.Uint32(data[offset : offset+4])
				value = float64(math.Float32frombits(bits))
				offset += 4
				quality = types.QualityOnline
			default:
				offset += 5
			}

			result = append(result, &types.AnalogInput{
				Index:   index,
				Value:   value,
				Quality: quality,
			})
		}
	}

	return result
}

// parseCounters parses counter data from response
func parseCounters(data []byte) []*types.Counter {
	var result []*types.Counter
	offset := 0

	for offset < len(data) {
		if offset+4 > len(data) {
			break
		}

		group := data[offset]
		variation := data[offset+1]
		_ = data[offset+2] // qualifier
		count := data[offset+3]
		offset += 4

		if group != 20 { // Group 20 = Counter
			// Skip past this group's data
			offset = skipGroupData(offset, data, group, variation, count)
			continue
		}

		for i := 0; i < int(count) && offset+5 <= len(data); i++ {
			index := binary.BigEndian.Uint16(data[offset:offset+2])
			offset += 2

			var value uint32
			var quality types.QualityFlags

			switch variation {
			case 1: // 32-bit with flags
				value = binary.BigEndian.Uint32(data[offset : offset+4])
				offset += 4
				quality = types.QualityFlags(data[offset])
				offset++
			case 5: // 16-bit with flags
				value = uint32(binary.BigEndian.Uint16(data[offset:offset+2]))
				offset += 2
				quality = types.QualityFlags(data[offset])
				offset++
			case 6: // 32-bit without flags
				value = binary.BigEndian.Uint32(data[offset : offset+4])
				offset += 4
				quality = types.QualityOnline
			default:
				offset += 5
			}

			result = append(result, &types.Counter{
				Index:   index,
				Value:   value,
				Quality: quality,
			})
		}
	}

	return result
}

// skipGroupData advances the offset past the data for a group.
// Returns the new offset position.
func skipGroupData(offset int, data []byte, group uint8, variation uint8, count uint8) int {
	// Calculate bytes per point based on group/variation
	var bytesPerPoint int
	switch group {
	case 1: // Binary Input Group
		switch variation {
		case 1: // With flags: index(2) + value+flags(1) = 3
			bytesPerPoint = 3
		case 2: // Without flags: index(2) + value(1) = 3
			bytesPerPoint = 3
		default:
			bytesPerPoint = 3
		}
	case 30: // Analog Input Group
		switch variation {
		case 1: // 32-bit float with flags: index(2) + float(4) + flags(1) = 7
			bytesPerPoint = 7
		case 2: // 16-bit int with flags: index(2) + int16(2) + flags(1) = 5
			bytesPerPoint = 5
		case 3: // 32-bit int with flags: index(2) + int32(4) + flags(1) = 7
			bytesPerPoint = 7
		case 5: // 32-bit float without flags: index(2) + float(4) = 6
			bytesPerPoint = 6
		default:
			bytesPerPoint = 7
		}
	case 20: // Counter Group
		switch variation {
		case 1: // 32-bit with flags: index(2) + value(4) + flags(1) = 7
			bytesPerPoint = 7
		case 5: // 16-bit with flags: index(2) + value(2) + flags(1) = 5
			bytesPerPoint = 5
		case 6: // 32-bit without flags: index(2) + value(4) = 6
			bytesPerPoint = 6
		default:
			bytesPerPoint = 7
		}
	default:
		bytesPerPoint = 7 // Conservative default
	}

	newOffset := offset + int(count)*bytesPerPoint
	if newOffset > len(data) {
		return len(data)
	}
	return newOffset
}

// Operate implements Client.Operate
func (c *client) Operate(ctx context.Context, command *types.ControlOutput) (*OperateResponse, error) {
	if command == nil {
		return nil, dnp3.ErrInvalidRequest
	}

	c.mu.RLock()
	state := c.state
	internal := c.internalMaster
	outstationID := c.config.OutstationAddress
	c.mu.RUnlock()

	if state < dnp3.StateConnected {
		return nil, dnp3.ErrNotConnected
	}

	// Translate command
	selectThenOperate := command.CommandType == types.SelectThenOperate

	// Perform operate through internal master
	if err := internal.Operate(outstationID, selectThenOperate, command.Group, command.Variation, command.Index, command.Value); err != nil {
		return nil, fmt.Errorf("operate failed: %w", err)
	}

	// Get outstation state
	outstation, _ := internal.GetOutstation(outstationID)
	var iin [2]byte
	if outstation != nil {
		iin = outstation.IIN
	}

	return &OperateResponse{
		IIN:       iin,
		Status:    types.ControlSuccess,
		Timestamp: time.Now(),
	}, nil
}

// EnableUnsolicited implements Client.EnableUnsolicited
func (c *client) EnableUnsolicited(ctx context.Context) error {
	c.mu.RLock()
	state := c.state
	internal := c.internalMaster
	outstationID := c.config.OutstationAddress
	c.mu.RUnlock()

	if state < dnp3.StateConnected {
		return dnp3.ErrNotConnected
	}

	return internal.EnableUnsolicited(outstationID)
}

// DisableUnsolicited implements Client.DisableUnsolicited
func (c *client) DisableUnsolicited(ctx context.Context) error {
	c.mu.RLock()
	state := c.state
	internal := c.internalMaster
	outstationID := c.config.OutstationAddress
	c.mu.RUnlock()

	if state < dnp3.StateConnected {
		return dnp3.ErrNotConnected
	}

	return internal.DisableUnsolicited(outstationID)
}

// SetUnsolicitedHandler implements Client.SetUnsolicitedHandler
func (c *client) SetUnsolicitedHandler(handler UnsolicitedHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers = append(c.handlers, handler)
}

// Close implements Client.Close
func (c *client) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.Disconnect(ctx)
}
