# DNP3 Engineering Workbench - Engineering Plan

**Plan ID**: DNP3-ENG-WORKBENCH-001  
**Title**: DNP3 Engineering Workbench  
**Status**: PLANNED  
**Date**: 2026-07-25  
**Authority**: KDE Runtime (DNP3 Library)  
**Bootstrap**: SUCCESS

---

## 1. Executive Summary

### 1.1 Purpose

Design a Windows desktop application that serves as an engineering workbench for validating and debugging the native Go DNP3 library.

### 1.2 Scope

| Phase | Scope |
|-------|-------|
| MVP (Phase 1) | Master Mode only |
| Future (Phase 2) | Outstation Mode |

### 1.3 Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Windows Desktop GUI | Engineering workstation target |
| Go + GUI Framework | Matches library language |
| Master Mode MVP | Validate core library functionality first |
| Minimal UI | Engineering tool, not production SCADA |

---

## 2. Architecture Overview

### 2.1 Overall Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        DNP3 Engineering Workbench                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐│
│  │   UI Layer  │  │  ViewModel  │  │   Shared Components     ││
│  │  (windows)  │──│  (logic)    │──│   (widgets, panels)    ││
│  └─────────────┘  └─────────────┘  └─────────────────────────┘│
│         │                │                      │               │
│         └────────────────┼──────────────────────┘               │
│                          │                                      │
│                    ┌─────▼─────┐                                │
│                    │  Session  │                                │
│                    │  Manager  │                                │
│                    └─────┬─────┘                                │
│                          │                                      │
│  ┌──────────────────────┼──────────────────────────────────┐  │
│  │                      │               DNP3 Library        │  │
│  │              ┌───────▼───────┐                          │  │
│  │              │   Session     │                          │  │
│  │              │   (Master or  │                          │  │
│  │              │   Outstation) │                          │  │
│  │              └───────┬───────┘                          │  │
│  │                      │                                  │  │
│  │              ┌───────▼───────┐                          │  │
│  │              │   Transport   │                          │  │
│  │              │   (TCP/TLS)   │                          │  │
│  │              └───────────────┘                          │  │
│  │                                                         │  │
│  │              ┌───────────────┐                          │  │
│  │              │   Protocol    │                          │  │
│  │              │   Decoder     │                          │  │
│  │              └───────────────┘                          │  │
│  └─────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Logging Subsystem                     │    │
│  │              (Communication Log Panel)                   │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Architecture Principles

1. **Mode Abstraction**: Master and Outstation modes share common infrastructure
2. **Single Session**: Only one active session at a time (MVP constraint)
3. **Protocol Visibility**: All protocol layers visible to engineers
4. **Immediate Usability**: No configuration files required for basic operation

---

## 3. Screen Layout

### 3.1 Main Window Layout

```
┌──────────────────────────────────────────────────────────────────────────┐
│ DNP3 Engineering Workbench                              [_][□][X]        │
├──────────────────────────────────────────────────────────────────────────┤
│ File   Edit   View   Session   Help                                      │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────────┐  ┌──────────────────────────────────────────┐  │
│  │ MODE SELECTION      │  │                                          │  │
│  │                     │  │           RESPONSE / DATA VIEWER          │  │
│  │ ○ Master Mode       │  │                                          │  │
│  │ ● Outstation Mode   │  │  ┌────────────────────────────────────┐ │  │
│  │                     │  │  │ Binary Inputs (Group 1)             │ │  │
│  ├─────────────────────┤  │  │ Index  Value  Quality  Time        │ │  │
│  │ CONNECTION          │  │  │   0     ON     ONLINE   10:23:01   │ │  │
│  │                     │  │  │   1     OFF    ONLINE   10:23:01   │ │  │
│  │ IP Address:         │  │  └────────────────────────────────────┘ │  │
│  │ [localhost      ]   │  │                                          │  │
│  │                     │  │  ┌────────────────────────────────────┐ │  │
│  │ TCP Port:           │  │  │ Analog Inputs (Group 30)          │ │  │
│  │ [20000         ]   │  │  │ Index  Value    Quality  Time      │ │  │
│  │                     │  │  │   0     123.45   ONLINE   10:23:01 │ │  │
│  │ [ Connect ]         │  │  └────────────────────────────────────┘ │  │
│  │                     │  │                                          │  │
│  ├─────────────────────┤  │  ┌────────────────────────────────────┐ │  │
│  │ COMMAND PANEL       │  │  │ Counters (Group 20)                │ │  │
│  │                     │  │  │ Index  Value    Quality  Time       │ │  │
│  │ [Read Class 0    ]  │  │  │   0     99999    ONLINE  10:23:01 │ │  │
│  │ [Read Class 1    ]  │  │  └────────────────────────────────────┘ │  │
│  │ [Read Class 2    ]  │  │                                          │  │
│  │ [Read Class 3    ]  │  └──────────────────────────────────────────┘  │
│  │                     │                                                 │
│  │ [Operate DO      ]  │  ┌──────────────────────────────────────────┐  │
│  │                     │  │           PROTOCOL DECODER                │  │
│  │ [Enable Unsolicited]│  │                                          │  │
│  │                     │  │ DLL: DIR=0 PRM=1 FCB=1 FCV=1 FUNC=3     │  │
│  ├─────────────────────┤  │     DEST=1024 SRC=1 LEN=12             │  │
│  │ STATUS              │  │                                          │  │
│  │                     │  │ TL:  FIR=1 FIN=1 CON=0 UNS=0 SEQ=5     │  │
│  │ Connection: ●       │  │                                          │  │
│  │ State: Connected    │  │ AL:  FUNC=READ (0x01)                   │  │
│  │ IIN: 0x0000        │  │     Objects: Group 1 Var 1 (All)        │  │
│  │                     │  │                                          │  │
│  └─────────────────────┘  └──────────────────────────────────────────┘  │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  COMMUNICATION LOG                                                       │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │ [10:23:01.123] TX → 1024: 05 64 01 C0 01 01 07 00             │   │
│  │ [10:23:01.145] RX ← 1024: 81 00 00 00 01 01 01 00 81 00       │   │
│  │ [10:23:01.156] IIN: 0x0000 (No issues)                         │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│  [Clear Log]                              Bytes TX: 6  Bytes RX: 10       │
└──────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Layout Regions

| Region | Purpose | MVP Essential |
|--------|---------|---------------|
| Mode Selection | Switch Master/Outstation | Yes |
| Connection Config | IP/Port settings | Yes |
| Connect Button | Session control | Yes |
| Command Panel | Send DNP3 requests | Yes |
| Response Viewer | Display decoded data | Yes |
| Protocol Decoder | Show raw bytes decoded | Yes |
| Communication Log | Full message history | Yes |
| Status Bar | Connection state, IIN | Yes |

---

## 4. Window/Component Hierarchy

### 4.1 Component Structure

```
MainWindow
├── MenuBar
│   ├── FileMenu
│   ├── EditMenu
│   ├── ViewMenu
│   ├── SessionMenu
│   └── HelpMenu
├── ModePanel
│   ├── RadioButton: MasterMode
│   └── RadioButton: OutstationMode
├── ConnectionPanel
│   ├── TextField: IPAddress
│   ├── TextField: TCPPort
│   ├── Button: Connect
│   └── Button: Disconnect
├── CommandPanel (Master)
│   ├── Button: ReadClass0
│   ├── Button: ReadClass1
│   ├── Button: ReadClass2
│   ├── Button: ReadClass3
│   ├── ButtonGroup: ReadObjects
│   ├── Button: Operate
│   └── Button: EnableUnsolicited
├── CommandPanel (Outstation) [Future]
│   ├── ...
├── DataPanel
│   ├── Table: BinaryInputs
│   ├── Table: AnalogInputs
│   └── Table: Counters
├── ProtocolPanel
│   ├── TextArea: DllDecode
│   ├── TextArea: TlDecode
│   └── TextArea: AlDecode
├── LogPanel
│   ├── ListView: LogEntries
│   ├── Button: ClearLog
│   └── Label: BytesTX/RX
└── StatusBar
    ├── Indicator: ConnectionStatus
    ├── Label: State
    └── Label: IIN
```

### 4.2 Dialog Windows

| Dialog | Purpose | Phase |
|--------|---------|-------|
| OperateDialog | Configure control command | MVP |
| ReadObjectsDialog | Select specific object groups | MVP |
| SettingsDialog | Configure timeouts, retries | MVP |
| AboutDialog | Version information | MVP |
| OutstationConfigDialog | Configure outstation params | Future |

---

## 5. Internal Package Structure

### 5.1 Package Hierarchy

```
cmd/workbench/
├── main.go                    # Application entry point
├── internal/
│   ├── ui/                   # UI components
│   │   ├── window.go         # Main window
│   │   ├── panels/          # UI panels
│   │   │   ├── mode.go      # Mode selection panel
│   │   │   ├── connection.go # Connection panel
│   │   │   ├── commands.go   # Command panel
│   │   │   ├── data.go      # Data display panel
│   │   │   ├── protocol.go   # Protocol decoder panel
│   │   │   └── log.go       # Log panel
│   │   └── dialogs/         # Dialog windows
│   │       ├── operate.go   # Operate dialog
│   │       └── settings.go  # Settings dialog
│   │
│   ├── session/              # Session management
│   │   ├── session.go       # Session interface
│   │   ├── master.go        # Master session
│   │   ├── outstation.go    # Outstation session (Future)
│   │   └── manager.go       # Session manager
│   │
│   ├── protocol/            # Protocol utilities
│   │   ├── decoder.go       # Protocol decoder
│   │   └── formatter.go     # Hex/text formatting
│   │
│   └── logging/             # Logging subsystem
│       ├── logger.go        # Logger interface
│       ├── entry.go         # Log entry
│       └── buffer.go        # Log buffer (circular)
│
├── go.mod
└── go.sum
```

### 5.2 Session Interface

```go
// Session represents an active DNP3 session
type Session interface {
    // Mode returns the session mode (Master or Outstation)
    Mode() SessionMode
    
    // Connect establishes the connection
    Connect(ctx context.Context, address string, port int) error
    
    // Disconnect closes the connection
    Disconnect(ctx context.Context) error
    
    // State returns current connection state
    State() ConnectionState
    
    // SendCommand sends a DNP3 command
    SendCommand(ctx context.Context, cmd Command) (*Response, error)
    
    // Events returns a channel for session events
    Events() <-chan SessionEvent
    
    // Close terminates the session
    Close() error
}
```

### 5.3 Shared Components

| Component | Location | Reusable |
|-----------|----------|----------|
| Hex formatting | protocol/formatter.go | Yes |
| Protocol decoder | protocol/decoder.go | Yes |
| Log buffer | logging/buffer.go | Yes |
| Connection config | session/config.go | Yes |

---

## 6. Development Phases

### 6.1 Phase 1: MVP (Master Mode Only)

**Goal**: Minimal usable engineering tool for validating DNP3 Master functionality

#### Week 1: Foundation
| Task | Deliverable | Effort |
|------|-------------|--------|
| Project setup | Go project with GUI framework | 1 day |
| Main window shell | Empty window with layout | 1 day |
| Connection panel | IP/Port input, Connect button | 1 day |
| Basic TCP connection | Connect to outstation | 1 day |

#### Week 2: Core Master Function
| Task | Deliverable | Effort |
|------|-------------|--------|
| Read Class 0 | Integrity poll command | 1 day |
| Read Class 1/2/3 | Event poll commands | 1 day |
| Data display | Table widgets for responses | 1 day |
| Protocol decoder | Show decoded bytes | 1 day |

#### Week 3: Polish
| Task | Deliverable | Effort |
|------|-------------|--------|
| Communication log | Message history panel | 1 day |
| Operate command | Send control commands | 1 day |
| Error handling | Connection errors, timeouts | 1 day |
| Testing | Manual testing, bug fixes | 1 day |

**MVP Deliverables**:
- [ ] Connect/Disconnect to outstation
- [ ] Read Class 0/1/2/3
- [ ] Display decoded response data
- [ ] Show protocol layer breakdown
- [ ] Communication log with TX/RX
- [ ] Operate command (binary output)

### 6.2 Phase 2: Outstation Mode (Future)

**Goal**: Add Outstation support with same UI infrastructure

#### Week 1-2: Outstation Foundation
| Task | Deliverable |
|------|-------------|
| Outstation session | Listen for connections |
| Address configuration | Master/Link address settings |
| Point database | Digital, analog, counter inputs |

#### Week 3-4: Outstation Features
| Task | Deliverable |
|------|-------------|
| Event generation | Simulate value changes |
| Control handling | Respond to operate commands |
| Unsolicited responses | Push events to master |

### 6.3 Phase 3: Advanced Features (Future)

| Feature | Priority | Notes |
|---------|----------|-------|
| TLS support | Medium | Secure transport |
| Command logging | Medium | Save/load command sequences |
| Scripting | Low | Automate test sequences |
| Protocol conformance | Low | Validate against spec |

---

## 7. MVP Scope Definition

### 7.1 Included Features (MVP)

| Feature | Description | Status |
|---------|-------------|--------|
| Master Mode | Act as DNP3 Master | Required |
| TCP Connection | Connect to outstation | Required |
| Read Class 0 | Integrity poll | Required |
| Read Class 1/2/3 | Event polls | Required |
| Data Display | Show decoded values | Required |
| Protocol Decoder | Layer breakdown | Required |
| Communication Log | TX/RX history | Required |
| Clear Log | Clear message history | Required |
| Operate Command | Send binary control | Required |
| Connection Status | Show state/IIN | Required |

### 7.2 Excluded Features (MVP)

| Feature | Reason for Exclusion | Future Phase |
|---------|---------------------|--------------|
| Outstation Mode | Not in MVP scope | Phase 2 |
| TLS Transport | Not needed for validation | Phase 3 |
| Save/Load Config | No config files in MVP | Phase 3 |
| Scripting | Not required for debugging | Phase 3 |
| Multiple Sessions | Single session only | Phase 3 |
| Unsolicited Responses | Master receives only | Phase 2 |
| File Transfer | Not needed for library validation | Future |
| Secure Authentication | Not needed for basic testing | Phase 3 |

### 7.3 MVP Constraints

| Constraint | Value |
|------------|-------|
| Max concurrent sessions | 1 |
| Transport protocols | TCP only |
| Operating system | Windows |
| Configuration | No files (in-memory) |

---

## 8. Future Outstation Integration Strategy

### 8.1 Architecture Compatibility

The architecture is designed so Outstation Mode can be added without refactoring:

```
┌─────────────────────────────────────────────────────────┐
│                    Shared Infrastructure                 │
├─────────────────────────────────────────────────────────┤
│                                                         │
│   ┌───────────────┐         ┌───────────────┐         │
│   │   UI Panels  │         │   Protocol    │         │
│   │   (Shared)   │         │   Decoder     │         │
│   └───────────────┘         └───────────────┘         │
│                                                         │
│   ┌───────────────┐         ┌───────────────┐         │
│   │   Session    │         │    Logger     │         │
│   │   Manager    │         │               │         │
│   └───────────────┘         └───────────────┘         │
│                                                         │
├─────────────────────────────────────────────────────────┤
│                    Mode-Specific                        │
├─────────────────────────────────────────────────────────┤
│                                                         │
│   ┌───────────────┐         ┌───────────────┐         │
│   │    Master    │         │   Outstation  │         │
│   │   Session    │         │   Session     │         │
│   └───────────────┘         └───────────────┘         │
│         │                         │                    │
│         ▼                         ▼                    │
│   ┌───────────┐           ┌───────────┐              │
│   │  Master   │           │  Outstn   │              │
│   │  Commands │           │  Points   │              │
│   └───────────┘           └───────────┘              │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### 8.2 Adding Outstation Mode

**Step 1**: Implement `OutstationSession` interface
```go
type OutstationSession interface {
    Session
    StartListener(ctx context.Context, port int) error
    SetPointDatabase(db PointDatabase)
    SetEventHandler(handler EventHandler)
}
```

**Step 2**: Add Outstation-specific UI panels
- Point Configuration panel
- Event Simulation panel
- Listen status indicator

**Step 3**: Add mode-specific command handlers
- Master: Read commands
- Outstation: Event generation

### 8.3 Shared Components

| Component | Master Usage | Outstation Usage |
|-----------|--------------|------------------|
| Connection Panel | Connect to outstation | Start listener |
| Protocol Decoder | Decode responses | Encode responses |
| Communication Log | Log TX/RX | Log RX/TX |
| Data Panel | Display read values | Display point values |

---

## 9. GUI Framework Selection

### 9.1 Options Considered

| Framework | Pros | Cons | Recommendation |
|-----------|------|------|----------------|
| **Fyne** | Pure Go, cross-platform, modern | Less mature on Windows | **Recommended** |
| Wails | Web frontend, mature | Adds complexity | Alternative |
| Gio | Pure Go, immediate mode | Steeper learning curve | Alternative |
| Electron+Go | Familiar web tech | Heavy, not native | Not recommended |

### 9.2 Recommended: Fyne

**Rationale**:
1. **Pure Go**: No C dependencies, matches library language
2. **Cross-platform**: Can build for Windows, Linux, macOS
3. **Simple API**: Faster development than Gio
4. **Active development**: Regular releases
5. **MIT License**: Permissive, matches project

### 9.3 Alternative: Wails

**Use if**:
- Web frontend team exists
- React/Vue preferred for UI
- Need complex UI components

---

## 10. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| GUI framework complexity | Medium | Medium | Use Fyne (simple API) |
| Protocol decoder accuracy | Low | High | Test against known frames |
| Thread safety in UI | Medium | Medium | Use Fyne's thread-safe API |
| Windows-specific issues | Low | Low | Test on target OS |
| Library API changes | Medium | Medium | Pin to version, add wrapper |

---

## 11. Implementation Notes

### 11.1 Thread Safety

All DNP3 library calls must be made from goroutines, with UI updates on the main thread:

```go
go func() {
    resp, err := client.Read(ctx, request)
    fyne.NewRunnable(func() {
        updateUI(resp, err)
    })
}()
```

### 11.2 Logging Architecture

Use a channel-based logger to decouple I/O from protocol handling:

```go
type LogEntry struct {
    Timestamp time.Time
    Direction string // "TX" or "RX"
    Address   uint16
    RawBytes  []byte
    Decoded   string
}

logChan := make(chan LogEntry, 100)
go logWriter(logChan)
```

### 11.3 Protocol Decoder Design

Decode protocol layers independently for visibility:

```go
type DecodedFrame struct {
    DLL *dll.Frame
    TL  *tl.Fragment
    AL  *al.APDU
    Raw []byte
}
```

---

## 12. Conclusion

### 12.1 Summary

| Item | Value |
|------|-------|
| Application Type | Windows Desktop GUI |
| Target Framework | Go + Fyne |
| MVP Scope | Master Mode only |
| Development Time | ~3 weeks |
| Phase 2 Scope | Outstation Mode |

### 12.2 Next Steps

1. **Approve plan** - Authorize implementation
2. **Select framework** - Confirm Fyne or choose alternative
3. **Initialize project** - Create cmd/workbench structure
4. **Implement MVP** - Build Phase 1 features
5. **Test** - Validate with real outstation

---

## Appendix A: Command Reference

### MVP Commands

| Command | DNP3 Function | Group/Variation |
|---------|----------------|-----------------|
| Read Class 0 | READ | Group 60 Var 1 |
| Read Class 1 | READ | Group 60 Var 2 |
| Read Class 2 | READ | Group 60 Var 3 |
| Read Class 3 | READ | Group 60 Var 4 |
| Read Binary Inputs | READ | Group 1 Var 1 |
| Read Analog Inputs | READ | Group 30 Var 1 |
| Read Counters | READ | Group 20 Var 1 |
| Operate Binary | OPERATE | Group 12 Var 1 |
| Enable Unsolicited | DIRECT OPERATE | Group 60 Var 3 |

---

*Plan created: 2026-07-25*  
*Authority: KDE Runtime (DNP3 Library)*  
*Status: AWAITING APPROVAL*
