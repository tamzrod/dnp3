# DNP3 Engineering Workbench - Implementation Plan (Fyne)

**Plan ID**: DNP3-WORKBENCH-IMPLEMENTATION-001  
**Title**: Windows DNP3 Engineering Workbench Implementation Plan  
**Status**: PLANNED  
**Date**: 2026-07-25  
**Authority**: KDE Runtime (DNP3 Library)  
**Framework**: Fyne v2.4.0  
**Bootstrap**: SUCCESS

---

## 1. Executive Summary

### 1.1 Purpose

Produce a comprehensive implementation plan for a Windows desktop engineering tool built using the **Fyne GUI Framework** for validating and debugging the native Go DNP3 library.

### 1.2 Framework Decision

| Decision | Rationale |
|----------|-----------|
| **Fyne v2.4.0** | Pure Go, no C dependencies, MIT license, cross-platform |
| Linux Development | Native Go toolchain on Linux |
| Windows Deployment | Go cross-compilation for Windows executables |

### 1.3 Project Scope

| Phase | Scope |
|-------|-------|
| **MVP** | Master Mode only, engineering tool |
| **Future** | Outstation Mode support |

### 1.4 Key Constraints

- Fyne exclusively - no framework evaluation
- MVP intentionally small
- Engineering productivity over appearance
- No production SCADA features

---

## 2. Overall Architecture

### 2.1 Layer Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Presentation Layer                          │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐ │
│  │   Windows   │ │   Layouts   │ │     Widgets            │ │
│  │  Container  │ │  Split/Box  │ │ Button/Entry/Table     │ │
│  └─────────────┘ └─────────────┘ └─────────────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                      Controller Layer                           │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                 MainController                            │ │
│  │  • Event routing    • State management    • Navigation │ │
│  └─────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                      Session Layer                             │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              SessionManager + MasterSession               │ │
│  │  • Connection lifecycle  • Command dispatch  • Events   │ │
│  └─────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                      Protocol Layer                            │
│  ┌──────────────────────┐ ┌──────────────────────────────┐  │
│  │   DNP3 Library       │ │   Protocol Decoder          │  │
│  │   (pkg/dnp3)         │ │   (DLL/TL/AL parsing)      │  │
│  └──────────────────────┘ └──────────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                      Transport Layer                           │
│  ┌──────────────────────┐ ┌──────────────────────────────┐  │
│  │   TCP Handler        │ │   TLS Handler (Future)      │  │
│  │   (pkg/transport)    │ │                            │  │
│  └──────────────────────┘ └──────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Data Flow Architecture

```
User Action (Button Click)
        │
        ▼
Fyne Event Handler (goroutine)
        │
        ▼
Controller.ReceiveEvent()
        │
        ▼
SessionManager.RouteCommand()
        │
        ▼
MasterSession.SendCommand()
        │
        ├─────────────────────────┐
        │                         │
        ▼                         ▼
DNP3 Library              Log Entry
(send request)            (append to log)
        │
        ▼
Response received
        │
        ▼
Parse & Decode Response
        │
        ▼
SessionEvent emitted
        │
        ▼
UI Update (fyne.NewRunnable)
        │
        ▼
Widget Refresh
```

---

## 3. Recommended Fyne Application Structure

### 3.1 Project Layout

```
cmd/workbench/
├── main.go                          # Application entry point
│
├── internal/
│   │
│   ├── app/
│   │   ├── app.go                  # Application initialization
│   │   └── config.go               # App configuration
│   │
│   ├── controller/
│   │   └── controller.go           # Main application controller
│   │
│   ├── session/
│   │   ├── session.go              # Session interface
│   │   ├── master.go               # Master session implementation
│   │   ├── manager.go              # Session manager
│   │   └── events.go               # Event types
│   │
│   ├── protocol/
│   │   ├── decoder.go              # Protocol decoder
│   │   └── formatter.go            # Hex/text formatting
│   │
│   ├── logger/
│   │   ├── logger.go               # Logger interface
│   │   └── buffer.go               # Circular log buffer
│   │
│   └── ui/
│       ├── window.go               # Main window
│       ├── container.go            # Container layout
│       │
│       ├── panels/
│       │   ├── panel.go            # Base panel interface
│       │   ├── mode.go             # Mode selection panel
│       │   ├── connection.go       # Connection config panel
│       │   ├── commands.go         # Command buttons panel
│       │   ├── response.go         # Response viewer panel
│       │   ├── points.go           # Point value panel
│       │   ├── protocol.go         # Protocol decoder panel
│       │   ├── log.go              # Communication log panel
│       │   └── status.go           # Status bar
│       │
│       ├── dialogs/
│       │   ├── dialog.go           # Base dialog
│       │   ├── operate.go          # Operate command dialog
│       │   └── settings.go         # Settings dialog
│       │
│       └── widgets/
│           ├── table.go            # Custom table widget
│           └── hex.go              # Hex display widget
│
├── go.mod
└── README.md
```

### 3.2 Package Responsibilities

| Package | Responsibility | Public API |
|---------|----------------|-----------|
| `app` | Application initialization | `Run()`, `Configure()` |
| `controller` | Event routing, state management | `HandleEvent()`, `GetState()` |
| `session` | DNP3 connection management | `Connect()`, `SendCommand()`, `Events()` |
| `protocol` | Protocol decoding/formatting | `Decode()`, `FormatHex()` |
| `logger` | Log management | `Log()`, `GetEntries()`, `Clear()` |
| `ui/panels` | UI panel components | `Refresh()`, `Update()` |
| `ui/dialogs` | Dialog windows | `Show()`, `Hide()` |

---

## 4. Window Layout

### 4.1 Main Window Structure

```
┌──────────────────────────────────────────────────────────────────────────┐
│ DNP3 Engineering Workbench                                    [_][□][X]  │
├──────────────────────────────────────────────────────────────────────────┤
│ File   Edit   View   Session   Help                                        │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────────┐  ┌────────────────────────────────────────┐  │
│  │ MODE SELECTION       │  │  RESPONSE VIEWER                        │  │
│  │                     │  │  ┌──────────────────────────────────────┐│  │
│  │ (•) Master Mode     │  │  │ Request: READ Class 0                ││  │
│  │ ( ) Outstation     │  │  │ Response: 0x81 SUCCESS               ││  │
│  │                     │  │  │ IIN: 0x0000                         ││  │
│  ├─────────────────────┤  │  └──────────────────────────────────────┘│  │
│  │ CONNECTION          │  ├────────────────────────────────────────┤  │
│  │                     │  │  POINT VALUES                           │  │
│  │ IP: [localhost  ]   │  │  ┌──────────────────────────────────────┐│  │
│  │ Port: [20000   ]   │  │  │ Binary Inputs (Group 1)             ││  │
│  │                     │  │  │  Index │ Value │ Quality │ Time    ││  │
│  │ [Connect]           │  │  │  0     │ ON    │ ONLINE  │ 10:23:01│  │
│  │ [Disconnect]        │  │  │  1     │ OFF   │ ONLINE  │ 10:23:01│  │
│  │                     │  │  └──────────────────────────────────────┘│  │
│  │ Status: Connected   │  │  ┌──────────────────────────────────────┐│  │
│  ├─────────────────────┤  │  │ Analog Inputs (Group 30)            ││  │
│  │ OPERATIONS          │  │  │  Index │ Value    │ Quality │ Time  ││  │
│  │                     │  │  │  0     │ 123.45  │ ONLINE  │10:23:01│  │
│  │ [Read Class 0]      │  │  └──────────────────────────────────────┘│  │
│  │ [Read Class 1]      │  └────────────────────────────────────────┘  │
│  │ [Read Class 2]      │                                                │
│  │ [Read Class 3]      │  ┌────────────────────────────────────────┐  │
│  │                     │  │  PROTOCOL DECODER                      │  │
│  │ [Operate...]       │  │  ┌──────────────────────────────────┐ │  │
│  │                     │  │  │ DLL: DIR=0 PRM=1 FCB=1           │ │  │
│  │ [Enable Unsolicited]│  │  │       DEST=1024 SRC=1 LEN=15     │ │  │
│  │                     │  │  │ TL:  FIR=1 FIN=1 SEQ=5           │ │  │
│  └─────────────────────┘  │  │ AL:  FUNC=READ Group=60 Var=1   │ │  │
│                            │  └──────────────────────────────────┘ │  │
│                            └────────────────────────────────────────┘  │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  COMMUNICATION LOG                    Bytes: TX=24  RX=18               │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ [10:23:01.123] TX → 05 64 01 C0 01 3C 01 00 ...              │  │
│  │ [10:23:01.145] RX ← 81 00 00 00 01 01 01 00 81 00 ...         │  │
│  │ [10:23:01.156] DLL: Frame decoded successfully                  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│  [Clear Log]                                     Status: Ready           │
└──────────────────────────────────────────────────────────────────────────┘
```

### 4.2 Layout Grid

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Menu Bar (App-level)                        │
├────────────┬──────────────────────────────────────────────────────┤
│            │                                                      │
│  Sidebar  │                   Main Content                         │
│  (250px)  │                  (Split Container)                   │
│            │                                                      │
│  ┌──────┐ │  ┌─────────────────────┬────────────────────────┐   │
│  │ Mode │ │  │                     │                        │   │
│  ├──────┤ │  │   Response/Points   │   Protocol Decoder     │   │
│  │ Conn │ │  │   (Table)          │   (Text)               │   │
│  ├──────┤ │  │                     │                        │   │
│  │ Ops  │ │  │                     │                        │   │
│  ├──────┤ │  │                     │                        │   │
│  │Status│ │  └─────────────────────┴────────────────────────┘   │
│  └──────┘ │                                                      │
├────────────┴──────────────────────────────────────────────────────┤
│                         Log Panel (200px height)                    │
├─────────────────────────────────────────────────────────────────────┤
│                         Status Bar (25px height)                    │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 5. Component Hierarchy

### 5.1 Fyne Widget Usage

| Component | Fyne Widget | Rationale |
|-----------|-------------|-----------|
| Main Window | `fyne.MainWindow` | Application shell |
| Sidebar | `fyne.Container` (VBox) | Fixed-width left panel |
| Buttons | `widget.Button` | Standard action triggers |
| Text Input | `widget.Entry` | IP/Port configuration |
| Data Tables | `widget.Table` | Point value display |
| Protocol View | `widget.RichText` | Multi-line decoded output |
| Log List | `widget.List` | Scrollable log entries |
| Status Bar | `fyne.Container` (HBox) | Fixed-height bottom |
| Dialogs | `dialog.Dialog` | Modal dialogs |

### 5.2 Component Tree

```
MainWindow
├── Canvas (full window)
│   ├── MenuBar
│   │   ├── FileMenu
│   │   ├── EditMenu
│   │   ├── ViewMenu
│   │   ├── SessionMenu
│   │   └── HelpMenu
│   │
│   ├── Container (Border layout)
│   │   ├── Top: nil
│   │   ├── Bottom: StatusBar
│   │   ├── Left: Sidebar (fixed 250px)
│   │   ├── Right: nil
│   │   └── Center: ContentSplit
│   │
│   └── ContentSplit (HSplit, resizable)
│       ├── Left: ResponsePoints (VSplit)
│       │   ├── Top: ResponsePanel
│       │   └── Bottom: PointsPanel
│       └── Right: ProtocolPanel
│
├── LogPanel (fixed bottom, 200px)
│
└── StatusBar (fixed bottom, 25px)
```

### 5.3 Sidebar Components

```
Sidebar (VBox, 250px width)
├── ModePanel
│   └── RadioGroup: [Master] [Outstation*disabled*]
│
├── ConnectionPanel
│   ├── Label: "IP Address:"
│   ├── Entry: ipAddress
│   ├── Label: "TCP Port:"
│   ├── Entry: tcpPort
│   ├── Button: connectBtn
│   ├── Button: disconnectBtn
│   └── Label: connectionStatus
│
├── OperationsPanel
│   ├── Label: "Read Commands"
│   ├── Button: readClass0
│   ├── Button: readClass1
│   ├── Button: readClass2
│   ├── Button: readClass3
│   ├── Separator
│   ├── Label: "Control Commands"
│   ├── Button: operateBtn
│   ├── Separator
│   ├── Button: enableUnsolicited
│
└── SessionStatusPanel
    ├── Label: "Connection:"
    ├── Label: statusValue
    ├── Label: "IIN:"
    └── Label: iinValue
```

---

## 6. Package Structure

### 6.1 Package Diagram

```
cmd/workbench/
│
├── main.go
│   └── Creates app.App, runs controller
│
├── internal/
│   │
│   ├── app/
│   │   └── app.go
│   │       └── type App struct { Controller, Window }
│   │       └── func Run()
│   │
│   ├── controller/
│   │   └── controller.go
│   │       └── type Controller struct { SessionManager, Logger }
│   │       └── func (c *Controller) HandleConnect()
│   │       └── func (c *Controller) HandleRead()
│   │       └── func (c *Controller) HandleOperate()
│   │
│   ├── session/
│   │   ├── interface.go        # Session interface
│   │   ├── master.go         # MasterSession implementation
│   │   ├── manager.go        # SessionManager
│   │   └── events.go         # Event types
│   │
│   ├── protocol/
│   │   ├── decoder.go
│   │   │   └── DecodeDLL(), DecodeTL(), DecodeAL()
│   │   └── formatter.go
│   │       └── FormatHex(), FormatFrame()
│   │
│   ├── logger/
│   │   ├── interface.go
│   │   ├── buffer.go          # Circular buffer
│   │   └── file.go           # Optional file logging
│   │
│   └── ui/
│       ├── window.go          # Main window setup
│       ├── container.go       # Layout helpers
│       │
│       ├── panels/
│       │   ├── panel.go       # BasePanel interface
│       │   ├── mode.go        # ModePanel
│       │   ├── connection.go   # ConnectionPanel
│       │   ├── commands.go    # CommandsPanel
│       │   ├── response.go    # ResponsePanel
│       │   ├── points.go      # PointsPanel
│       │   ├── protocol.go    # ProtocolPanel
│       │   ├── log.go        # LogPanel
│       │   └── status.go      # StatusBar
│       │
│       ├── dialogs/
│       │   ├── dialog.go      # Base dialog
│       │   ├── operate.go     # OperateDialog
│       │   └── settings.go    # SettingsDialog
│       │
│       └── widgets/
│           ├── table.go       # PointTable widget
│           └── hex.go         # HexView widget
```

### 6.2 Event Flow

```
┌──────────────────────────────────────────────────────────────────┐
│                         EVENT FLOW                                │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. UI Event (Fyne thread)                                      │
│     Button.OnTapped → OnConnect handler                           │
│                                                                  │
│  2. Controller (goroutine)                                       │
│     controller.HandleConnect(address, port)                        │
│         │                                                       │
│         ▼                                                       │
│  3. Session Layer                                               │
│     sessionManager.CreateMasterSession()                          │
│     masterSession.Connect(ctx, addr, port)                        │
│         │                                                       │
│         ▼                                                       │
│  4. DNP3 Library                                               │
│     transport.Send() → TCP connection                             │
│     transport.Receive() → response bytes                        │
│         │                                                       │
│         ▼                                                       │
│  5. Protocol Decoder                                           │
│     protocol.Decode(responseBytes) → DecodedFrame                 │
│         │                                                       │
│         ▼                                                       │
│  6. Session Event                                              │
│     events <- SessionEvent{Type: "response", Data: frame}       │
│         │                                                       │
│         ▼                                                       │
│  7. UI Update (main thread)                                    │
│     fyne.NewRunnable(func() {                                   │
│         responsePanel.Update(frame)                              │
│         pointsPanel.Update(points)                              │
│         protocolPanel.Update(decoded)                            │
│         logPanel.Append(entry)                                  │
│     })                                                         │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

---

## 7. Controller Design

### 7.1 Controller Interface

```go
// Controller handles application events and state.
type Controller interface {
    // Lifecycle
    Start() error
    Stop() error
    
    // Connection management
    Connect(address string, port int) error
    Disconnect() error
    
    // Command execution
    ReadClass(class int) error
    ReadObjects(groups []uint8) error
    Operate(point uint16, value interface{}) error
    
    // Event handlers (called from UI)
    OnConnectClicked()
    OnDisconnectClicked()
    OnReadClass0Clicked()
    OnReadClass1Clicked()
    OnOperateClicked()
    OnClearLogClicked()
    
    // State access
    State() *AppState
    Logger() *Logger
}

// AppState represents the current application state.
type AppState struct {
    Mode           SessionMode
    Connection    ConnectionState
    LastResponse   *Response
    DecodedFrame   *DecodedFrame
    LogEntries    []LogEntry
}
```

### 7.2 Controller Implementation

```go
type controller struct {
    app       *fyne.App
    session   Session
    logger    *Logger
    decoder   *protocol.Decoder
    state     *AppState
    stateLock sync.RWMutex
    events    chan SessionEvent
}

// OnConnectClicked handles connect button click.
func (c *controller) OnConnectClicked() {
    go func() {
        addr := c.getAddress()
        port := c.getPort()
        
        c.logger.Info("Connecting to %s:%d", addr, port)
        
        if err := c.session.Connect(context.Background(), addr, port); err != nil {
            c.logger.Error("Connection failed: %v", err)
            c.updateState(func(s *AppState) {
                s.Connection = StateError
            })
            return
        }
        
        c.logger.Info("Connected successfully")
        c.updateState(func(s *AppState) {
            s.Connection = StateConnected
        })
        
        // Start event pump
        c.pumpEvents()
    }()
}

// pumpEvents handles session events on the controller goroutine.
func (c *controller) pumpEvents() {
    for event := range c.session.Events() {
        c.handleSessionEvent(event)
    }
}
```

---

## 8. Session Design

### 8.1 Session Interface

```go
// Session represents an active DNP3 session.
type Session interface {
    Mode() SessionMode
    State() ConnectionState
    
    Connect(ctx context.Context, address string, port int) error
    Disconnect(ctx context.Context) error
    
    SendCommand(ctx context.Context, cmd Command) (*Response, error)
    
    Events() <-chan SessionEvent
    Close() error
}

// Command represents a DNP3 command.
type Command interface {
    Type() CommandType
}

// MasterSession implements Session for Master mode.
type MasterSession struct {
    mu         sync.RWMutex
    state      ConnectionState
    transport  transport.Handler
    client     master.Client
    decoder    *protocol.Decoder
    events     chan SessionEvent
    log        *Logger
}
```

### 8.2 Session Manager

```go
// Manager creates and manages sessions.
type Manager struct {
    mu      sync.RWMutex
    current Session
    logger  *Logger
}

// CreateMasterSession creates a new Master session.
func (m *Manager) CreateMasterSession() *MasterSession {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if m.current != nil {
        m.current.Close()
    }
    
    session := newMasterSession(m.logger)
    m.current = session
    return session
}
```

---

## 9. Build Workflow

### 9.1 Linux Development Workflow

```bash
# 1. Install Fyne dependencies (one-time)
sudo apt-get install libgl1-mesa-dev xorg-dev

# 2. Create go.mod
cd cmd/workbench
go mod init dnp3/cmd/workbench
go get fyne.io/fyne/v2@v2.4.0

# 3. Development build (fast iteration)
go build -o workbench-dev .
./workbench-dev

# 4. Run tests
go test ./...

# 5. Cross-compile for Windows
GOOS=windows GOARCH=amd64 go build -o workbench.exe .
```

### 9.2 Windows Cross-Compilation

```bash
# Set environment
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1

# Build with Fyne dependencies
go build -ldflags="-s -w" -o dnp3-workbench.exe .

# Or use make
make build-windows
```

### 9.3 Makefile Targets

```makefile
# Makefile for DNP3 Engineering Workbench

.PHONY: all build test clean run windows linux

# Build for current OS
all: build

# Build for current OS
build:
	go build -o workbench .

# Build for Windows (cross-compile)
windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o dnp3-workbench.exe .

# Build for Linux
linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o workbench-linux .

# Run locally
run:
	go run .

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f workbench workbench.exe workbench-linux

# Download dependencies
deps:
	go mod tidy
	go mod download

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate Fyne mobile resources (if needed)
resources:
	# fyne bundle -o bundled.go resources/
```

### 9.4 GitHub Actions Workflow

```yaml
# .github/workflows/build.yml
name: Build Workbench

on:
  push:
    branches: [main, feature/*]
  pull_request:
    branches: [main]

jobs:
  build:
    strategy:
      matrix:
        go-version: ['1.22']
        platform: [ubuntu-latest, windows-latest]
    
    runs-on: ${{ matrix.platform }}
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      
      - name: Install Linux dependencies
        if: matrix.platform == 'ubuntu-latest'
        run: |
          sudo apt-get update
          sudo apt-get install libgl1-mesa-dev xorg-dev
      
      - name: Get dependencies
        working-directory: cmd/workbench
        run: go mod tidy
      
      - name: Build
        working-directory: cmd/workbench
        run: go build -o workbench .
      
      - name: Test
        working-directory: cmd/workbench
        run: go test -v ./...
      
      - name: Build Windows
        if: matrix.platform == 'ubuntu-latest'
        run: |
          GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o workbench.exe .
      
      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: workbench-${{ matrix.platform }}
          path: cmd/workbench/workbench*
```

---

## 10. Cross-Compilation Workflow

### 10.1 Fyne Build Considerations

| Consideration | Recommendation |
|--------------|----------------|
| CGO | Required for Fyne graphics |
| Graphics libs | Install on build machine |
| Static linking | Use `-ldflags="-s -w"` |
| Bundle resources | Use `fyne bundle` |

### 10.2 Windows Build Environment (Linux)

```bash
# Install mingw for Windows cross-compilation
sudo apt-get install mingw-w64

# Build command
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
  CC=x86_64-w64-mingw32-gcc \
  go build -ldflags="-s -w" \
  -o dnp3-workbench.exe .
```

### 10.3 Build Script

```bash
#!/bin/bash
# scripts/build-windows.sh

set -e

echo "Building DNP3 Engineering Workbench for Windows..."

cd "$(dirname "$0")/.."

# Clean previous builds
rm -f dnp3-workbench.exe

# Cross-compile
GOOS=windows \
GOARCH=amd64 \
CGO_ENABLED=1 \
CC=x86_64-w64-mingw32-gcc \
go build \
  -ldflags="-s -w" \
  -o dnp3-workbench.exe \
  ./cmd/workbench

echo "Build complete: dnp3-workbench.exe"
ls -lh dnp3-workbench.exe
```

---

## 11. Implementation Roadmap

### Phase 1: Project Skeleton
**Duration**: 1 day

| Task | Deliverable | Acceptance Criteria |
|------|-------------|-------------------|
| Create project structure | Directory layout | All packages created |
| Initialize go.mod | go.mod with dependencies | `go mod tidy` succeeds |
| Create main.go | App entry point | App launches without error |
| Verify Fyne setup | Window displays | Window opens on Linux |

### Phase 2: Application Window
**Duration**: 1 day

| Task | Deliverable | Acceptance Criteria |
|------|-------------|-------------------|
| Create main window | MainWindow with layout | Layout matches spec |
| Create sidebar | Sidebar container | Contains mode/conn/ops panels |
| Create content split | HSplit layout | Resizable panels |
| Create status bar | Status bar | Shows state information |

### Phase 3: Connection Management
**Duration**: 2 days

| Task | Deliverable | Acceptance Criteria |
|------|-------------|-------------------|
| Create ConnectionPanel | IP/Port entry fields | Input accepted |
| Implement Connect button | Connect handler | Calls controller |
| Implement Session interface | Session abstraction | Can be mocked |
| Create SessionManager | Session factory | Creates master session |
| Integrate TCP transport | Connect to outstation | Connection established |

### Phase 4: Master Session Integration
**Duration**: 2 days

| Task | Deliverable | Acceptance Criteria |
|------|-------------|-------------------|
| Integrate DNP3 client | Real library calls | Library functions called |
| Implement ReadClass | Read command execution | Response received |
| Handle connection errors | Error display | Errors shown in UI |
| Implement Disconnect | Clean disconnect | Resources released |

### Phase 5: Polling Operations
**Duration**: 2 days

| Task | Deliverable | Acceptance Criteria |
|------|-------------|-------------------|
| Create CommandsPanel | Button panel | All buttons functional |
| Implement Read Class 0 | Integrity poll | All static data received |
| Implement Read Class 1/2/3 | Event polls | Event data received |
| Create OperateDialog | Control dialog | Values configurable |
| Implement Operate | Control command | Command sent |

### Phase 6: Protocol Viewer
**Duration**: 2 days

| Task | Deliverable | Acceptance Criteria |
|------|-------------|-------------------|
| Create ProtocolPanel | Protocol display | Shows decoded layers |
| Implement DLL decoder | DLL layer parsing | DIR, PRM, FCB, etc. |
| Implement TL decoder | TL layer parsing | FIR, FIN, SEQ |
| Implement AL decoder | AL layer parsing | FUNC, Objects |
| Format hex output | Hex display | Bytes formatted |

### Phase 7: Point Viewer
**Duration**: 2 days

| Task | Deliverable | Acceptance Criteria |
|------|-------------|-------------------|
| Create PointsPanel | Point table | Shows point data |
| Display Binary Inputs | Binary table | Index, Value, Quality |
| Display Analog Inputs | Analog table | Index, Value, Quality |
| Display Counters | Counter table | Index, Value, Quality |
| Update on response | Live updates | Points refresh |

### Phase 8: Logging
**Duration**: 1 day

| Task | Deliverable | Acceptance Criteria |
|------|-------------|-------------------|
| Create LogPanel | Log container | Shows message history |
| Implement LogBuffer | Circular buffer | Limited size |
| Log TX/RX messages | Message display | Timestamps shown |
| Implement Clear Log | Clear button | Log cleared |
| Format log entries | Log formatting | Human-readable |

### Phase 9: Windows Validation
**Duration**: 1 day

| Task | Deliverable | Acceptance Criteria |
|------|-------------|-------------------|
| Cross-compile | Windows exe | Builds successfully |
| Test on Windows | Functional app | All features work |
| Fix platform issues | Patched code | No Linux-specific code |
| Create release | Tagged release | Executable distributed |

### Phase Summary

| Phase | Duration | Cumulative |
|-------|----------|------------|
| 1: Skeleton | 1 day | 1 day |
| 2: Window | 1 day | 2 days |
| 3: Connection | 2 days | 4 days |
| 4: Session | 2 days | 6 days |
| 5: Polling | 2 days | 8 days |
| 6: Protocol | 2 days | 10 days |
| 7: Points | 2 days | 12 days |
| 8: Logging | 1 day | 13 days |
| 9: Windows | 1 day | 14 days |

**Total MVP Estimate**: ~14 days (2-3 weeks)

---

## 12. Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Fyne performance on Windows | Low | Medium | Use efficient widgets, avoid excessive redraws |
| Cross-compile issues | Medium | Low | Test on Linux first, CI validates builds |
| DNP3 library API changes | Medium | Medium | Wrap library calls, version pin |
| Thread safety in UI | Medium | High | Use `fyne.NewRunnable` for all UI updates |
| Mock session vs real | Low | Low | Session interface enables easy testing |
| Windows-specific bugs | Medium | Medium | Validate on Windows early |

### Risk Mitigation Strategies

1. **Thread Safety**
   - All DNP3 operations run in goroutines
   - UI updates use `fyne.NewRunnable`
   - State protected with mutexes

2. **Cross-Compilation**
   - CI validates Windows builds on every PR
   - Use `-ldflags="-s -w"` for smaller binaries
   - Test on Windows before release

3. **Library Versioning**
   - Pin DNP3 library version in go.mod
   - Test with mock transport first
   - Integration tests validate real transport

---

## 13. Future Outstation Integration

### 13.1 Architecture Compatibility

The MVP architecture supports future Outstation Mode without refactoring:

```
┌─────────────────────────────────────────────────────────────┐
│                   Shared Infrastructure                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   ┌───────────────┐         ┌───────────────┐             │
│   │    UI/View   │         │  Controller   │             │
│   │   (Shared)   │         │   (Shared)    │             │
│   └───────────────┘         └───────────────┘             │
│                                                             │
│   ┌───────────────┐         ┌───────────────┐             │
│   │    Logger    │         │   Protocol    │             │
│   │  (Shared)    │         │   (Shared)    │             │
│   └───────────────┘         └───────────────┘             │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                    Mode-Specific                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   ┌───────────────┐         ┌───────────────┐             │
│   │  MasterMode   │         │ OutstationMode │             │
│   │   (MVP)      │         │   (Future)    │             │
│   └───────────────┘         └───────────────┘             │
│         │                         │                          │
│         ▼                         ▼                          │
│   ┌───────────────┐         ┌───────────────┐             │
│   │ MasterSession│         │OutstationSession│             │
│   │ (connects)   │         │  (listens)     │             │
│   └───────────────┘         └───────────────┘             │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 13.2 Adding Outstation Mode

**Step 1**: Add OutstationSession to session package
```go
type OutstationSession struct {
    listener net.Listener
    points  *PointDatabase
    events  chan SessionEvent
}
```

**Step 2**: Add Outstation-specific UI panels
- OutstationConfigPanel (addresses, point database)
- EventSimulationPanel (simulate value changes)
- ListenStatusIndicator

**Step 3**: Update Controller with mode routing
```go
func (c *Controller) CreateSession(mode SessionMode) Session {
    switch mode {
    case ModeMaster:
        return c.sessionManager.CreateMasterSession()
    case ModeOutstation:
        return c.sessionManager.CreateOutstationSession()
    }
}
```

### 13.3 Shared Components (MVP + Future)

| Component | Master Usage | Outstation Usage |
|-----------|-------------|-----------------|
| ModePanel | Radio selection | Radio selection |
| ProtocolPanel | Decode RX | Encode TX |
| LogPanel | Log all | Log all |
| StatusBar | Connection state | Listen state |
| Controller | Command routing | Event routing |

---

## 14. Deferred Features

### 14.1 Not in MVP

| Feature | Reason | Future Phase |
|---------|--------|--------------|
| Outstation Mode | Not MVP scope | Phase 2 |
| TLS Transport | Not needed for validation | Phase 3 |
| Historian | Not engineering tool | Future |
| Database | Not engineering tool | Future |
| InfluxDB | Not engineering tool | Future |
| Trending | Not engineering tool | Future |
| Alarm Management | Not engineering tool | Future |
| Multi-device | Single device MVP | Phase 3 |
| User Accounts | Not needed | Future |
| Config Database | No config files in MVP | Phase 3 |
| Reporting | Not engineering tool | Future |
| Scripting | Not MVP scope | Phase 3 |

### 14.2 MVP Constraints

| Constraint | Value |
|------------|-------|
| Max sessions | 1 |
| Transport | TCP only |
| OS Target | Windows |
| Configuration | In-memory only |

---

## 15. Development Standards

### 15.1 Code Standards

| Standard | Tool |
|----------|------|
| Formatting | `go fmt` |
| Linting | `golangci-lint` |
| Vetting | `go vet` |
| Testing | `go test` |

### 15.2 Fyne Conventions

```go
// Thread safety
fyne.NewRunnable(func() {
    widget.Refresh()
})

// Widget naming
type ConnectionPanel struct {
    container *fyne.Container
    address   *widget.Entry
    port      *widget.Entry
    connect   *widget.Button
}

// Event handling
button.OnTapped = func() {
    go c.handleConnect() // Long-running in goroutine
}
```

### 15.3 Error Handling

```go
// Return errors to caller
func (s *MasterSession) Connect(ctx context.Context) error {
    if err := s.transport.Dial(address); err != nil {
        s.log.Error("Connect failed: %v", err)
        return fmt.Errorf("connect: %w", err)
    }
    return nil
}

// Log errors, don't expose internal details
func (c *Controller) OnConnectClicked() {
    if err := c.session.Connect(ctx); err != nil {
        c.logger.Error("Connection failed: %v", err)
        // Update UI with user-friendly message
        return
    }
}
```

---

## 16. Conclusion

### 16.1 Summary

| Item | Value |
|------|-------|
| Framework | Fyne v2.4.0 |
| Architecture | Layered (UI → Controller → Session → DNP3) |
| MVP Duration | ~14 days (2-3 weeks) |
| Phases | 9 |
| Package Count | 7 core packages |
| Platform | Windows (primary), Linux (development) |

### 16.2 Next Steps

1. **Approve plan** - Authorize implementation
2. **Phase 1: Skeleton** - Create project structure
3. **Phase 2: Window** - Build main window layout
4. **Continue phases** - Build through MVP completion
5. **Validate** - Test on Windows

### 16.3 Success Criteria

- [ ] Application launches on Windows
- [ ] Can connect to DNP3 outstation
- [ ] Read commands return decoded data
- [ ] Protocol layers are visible
- [ ] Communication log shows TX/RX
- [ ] UI is responsive
- [ ] No crashes during normal operation

---

## Appendix A: Fyne Widget Reference

### Core Widgets

| Widget | Use Case |
|--------|----------|
| `widget.Button` | Actions, commands |
| `widget.Entry` | Text input (IP, port) |
| `widget.Select` | Dropdown selection |
| `widget.RadioGroup` | Mode selection |
| `widget.Check` | Boolean options |
| `widget.Label` | Read-only text |
| `widget.List` | Scrollable list (log) |
| `widget.Table` | Point value display |
| `widget.Form` | Form layout |
| `widget.ProgressBar` | Progress indication |

### Container Widgets

| Widget | Use Case |
|--------|----------|
| `container.NewVBox` | Vertical stacking |
| `container.NewHBox` | Horizontal stacking |
| `container.NewHSplit` | Resizable horizontal split |
| `container.NewVSplit` | Resizable vertical split |
| `container.NewBorder` | Border-based layout |
| `container.NewGrid` | Grid layout |
| `container.NewMax/Min/Expand` | Size constraints |

### Dialogs

| Dialog | Use Case |
|--------|----------|
| `dialog.ShowInformation` | Simple messages |
| `dialog.ShowError` | Error display |
| `dialog.NewCustom` | Custom dialogs |
| `dialog.NewFileOpen` | File selection |
| `dialog.NewFileSave` | File save |

---

*Plan created: 2026-07-25*  
*Authority: KDE Runtime (DNP3 Library)*  
*Status: AWAITING APPROVAL*
