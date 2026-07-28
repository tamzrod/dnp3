# DNP3-INV-063 Specification: Terminal-based Workbench (TUI)

**Document ID**: DNP3-INV-063
**Version**: 1.0.0
**Date**: 2026-07-28
**Status**: DRAFT
**Authority**: KDE Runtime (DNP3 Library)
**Investigation**: DNP3-INV-063

---

## 1. Executive Summary

This specification defines the implementation of a terminal-based workbench (TUI) for DNP3 testing, replacing the Fyne-based GUI approach.

### 1.1 Design Decision

**Decision**: Build a custom terminal UI using ANSI escape codes.

**Rationale**:
1. Simplicity: Only need table + status display
2. Performance: Direct terminal control is fast
3. Portability: Works everywhere with terminal
4. Maintainability: No framework dependencies

---

## 2. Terminal Layout

### 2.1 Screen Regions

```
┌────────────────────────────────────────────────────────────────────────┐
│ [MASTER] DNP3 Workbench │ Connected: 127.0.0.1:20000 │ 10:23:45    │
├────────────────────────────────────────────────────────────────────────┤
│                                                                        │
│  ┌─ DATA POINTS ──────────────────────────────────────────────────┐  │
│  │ Type  │ Index │ Value        │ Quality  │ Timestamp             │  │
│  ├───────┼───────┼──────────────┼──────────┼──────────────────────┤  │
│  │ BI    │ 0     │ true         │ ONLINE  │ 10:23:45.123         │  │
│  │ BI    │ 1     │ false        │ ONLINE  │ 10:23:45.124         │  │
│  │ AI    │ 0     │ 100.5        │ ONLINE  │ 10:23:45.125         │  │
│  │ AI    │ 1     │ 50.25        │ ONLINE  │ 10:23:45.126         │  │
│  │ CTR   │ 0     │ 1000         │ ONLINE  │ 10:23:45.127         │  │
│  │ BO    │ 0     │ false        │ -       │ -                    │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                                                                        │
├────────────────────────────────────────────────────────────────────────┤
│ LOG │ [10:23:45.100] → SEND: Read Class 0                          │
│      │ [10:23:45.200] ← RECV: 10 binary, 5 analog, 3 counters     │
├────────────────────────────────────────────────────────────────────────┤
│ [q]uit [c]onnect [r]ead [1-4]class [↑↓]nav [Enter]select [l]clear  │
└────────────────────────────────────────────────────────────────────────┘
```

### 2.2 Region Dimensions

| Region | Height | Position |
|--------|--------|----------|
| Header | 1 line | Top |
| Main (Table) | N-5 lines | Middle |
| Log | 2-3 lines | Bottom-up |
| Controls | 1 line | Bottom |

---

## 3. ANSI Escape Codes

### 3.1 Color Codes

```go
const (
    // Foreground
    Black   = "\033[30m"
    Red     = "\033[31m"
    Green   = "\033[32m"
    Yellow  = "\033[33m"
    Blue    = "\033[34m"
    Magenta = "\033[35m"
    Cyan    = "\033[36m"
    White   = "\033[37m"

    // Bright foreground
    BrightBlack   = "\033[90m"
    BrightRed     = "\033[91m"
    BrightGreen   = "\033[92m"
    BrightYellow  = "\033[93m"
    BrightBlue    = "\033[94m"
    BrightMagenta = "\033[95m"
    BrightCyan    = "\033[96m"
    BrightWhite   = "\033[97m"

    // Styles
    Bold      = "\033[1m"
    Dim       = "\033[2m"
    Underline = "\033[4m"
    Reverse   = "\033[7m"

    // Reset
    Reset = "\033[0m"
)
```

### 3.2 Cursor Control

```go
// Move cursor
func MoveTo(row, col int) string {
    return fmt.Sprintf("\033[%d;%dH", row, col)
}

// Clear screen
ClearScreen = "\033[2J"

// Clear line
ClearLine = "\033[K"

// Hide/show cursor
HideCursor = "\033[?25l"
ShowCursor = "\033[?25h"

// Save/restore cursor
SaveCursor    = "\033[s"
RestoreCursor = "\033[u"
```

### 3.3 Style Application

```go
// Example: Green text on black background
GreenOnBlack = "\033[32;40m"

// Bold green
BoldGreen = "\033[1;32m"

// Underline yellow
UnderlineYellow = "\033[4;33m"
```

---

## 4. Implementation Specification

### 4.1 Package Structure

```
cmd/workbench/tui/
├── app.go          # Main application loop
├── layout.go       # Screen layout management
├── render.go       # Rendering utilities
├── table.go        # Data table widget
├── statusbar.go    # Status bar widget
├── log.go          # Log display widget
├── input.go        # Keyboard input handling
└── colors.go       # Color/style constants
```

### 4.2 Core Types

```go
// App is the main TUI application
type App struct {
    Mode    Mode          // master or outstation
    Layout  *Layout
    Table   *Table
    Status  *StatusBar
    Log     *Log
    Events  chan Event
    done    chan struct{}
}

// Mode represents the operating mode
type Mode string

const (
    ModeMaster    Mode = "master"
    ModeOutstation Mode = "outstation"
)

// Event represents user input events
type Event struct {
    Type EventType
    Data interface{}
}

type EventType int

const (
    EventKey EventType = iota
    EventResize
    EventUpdate
)
```

### 4.3 Layout Manager

```go
// Layout manages screen regions
type Layout struct {
    Width  int
    Height int
    Header int // Header height (lines)
    Footer int // Footer height (lines)
}

// GetMainBounds returns the main content area
func (l *Layout) GetMainBounds() (top, bottom int) {
    return l.Header, l.Height - l.Footer
}

// Resize updates layout for new terminal size
func (l *Layout) Resize(width, height int) {
    l.Width = width
    l.Height = height
}
```

### 4.4 Table Widget

```go
// Table displays scrollable data
type Table struct {
    Headers []string
    Rows    [][]string
    ColWidths []int
    Cursor   int
    Offset   int
    Selected int
}

// Draw renders the table
func (t *Table) Draw(screen *Screen, bounds Rect) {
    // Draw headers
    // Draw rows with highlighting
    // Draw scroll indicators
}

// HandleInput processes keyboard events
func (t *Table) HandleInput(key Key) {
    switch key {
    case KeyUp:
        t.MoveUp()
    case KeyDown:
        t.MoveDown()
    case KeyEnter:
        t.Select()
    }
}
```

### 4.5 Rendering

```go
// Screen provides low-level rendering
type Screen struct {
    buf bytes.Buffer
}

// Print writes to screen buffer
func (s *Screen) Print(x, y int, text string) {
    s.buf.WriteString(MoveTo(y, x))
    s.buf.WriteString(text)
}

// PrintStyled writes with style
func (s *Screen) PrintStyled(x, y int, text string, style string) {
    s.buf.WriteString(MoveTo(y, x))
    s.buf.WriteString(style)
    s.buf.WriteString(text)
    s.buf.WriteString(Reset)
}

// Flush outputs to terminal
func (s *Screen) Flush() error {
    os.Stdout.Write(s.buf.Bytes())
    s.buf.Reset()
}
```

---

## 5. Input Handling

### 5.1 Key Definitions

```go
type Key int

const (
    KeyUp Key = iota + 256
    KeyDown
    KeyLeft
    KeyRight
    KeyEnter
    KeyEscape
    KeyBackspace
    KeyTab
    KeySpace
    KeyCtrlA
    KeyCtrlC
    KeyCtrlD
    KeyCtrlL
    KeyCtrlQ
    KeyCtrlR
)

// Key names for display
var KeyNames = map[Key]string{
    KeyUp:         "↑",
    KeyDown:       "↓",
    KeyLeft:       "←",
    KeyRight:      "→",
    KeyEnter:      "Enter",
    KeyEscape:     "Esc",
    // ... etc
}
```

### 5.2 Input Loop

```go
// RunInputLoop handles keyboard input
func (a *App) RunInputLoop() {
    reader := bufio.NewReader(os.Stdin)

    for {
        select {
        case <-a.done:
            return
        default:
            // Set terminal to raw mode
            // Read single key
            // Convert to Key type
            // Send to event channel
        }
    }
}
```

---

## 6. File Changes

### 6.1 New Files

| File | Description |
|------|-------------|
| `cmd/workbench/tui/app.go` | Main TUI application |
| `cmd/workbench/tui/layout.go` | Screen layout |
| `cmd/workbench/tui/render.go` | Rendering utilities |
| `cmd/workbench/tui/table.go` | Data table widget |
| `cmd/workbench/tui/statusbar.go` | Status bar widget |
| `cmd/workbench/tui/log.go` | Log display widget |
| `cmd/workbench/tui/input.go` | Input handling |
| `cmd/workbench/tui/colors.go` | Color constants |

### 6.2 Modified Files

| File | Change |
|------|--------|
| `cmd/workbench/main.go` | Add --mode flag, use TUI instead of Fyne |

### 6.3 Deleted Files

| File | Reason |
|------|--------|
| `cmd/workbench/internal/ui/*.go` | Replaced by TUI |
| `cmd/workbench/internal/ui/dialogs/*.go` | Replaced by TUI |

---

## 7. Implementation Tasks

### 7.1 Task List

| # | Task | Priority | Status |
|---|------|----------|--------|
| 1 | Create tui package structure | HIGH | TODO |
| 2 | Implement ANSI rendering | HIGH | TODO |
| 3 | Create layout manager | HIGH | TODO |
| 4 | Implement table widget | HIGH | TODO |
| 5 | Implement status bar | MEDIUM | TODO |
| 6 | Implement log display | MEDIUM | TODO |
| 7 | Add keyboard input handling | HIGH | TODO |
| 8 | Create main application loop | HIGH | TODO |
| 9 | Update main.go to use TUI | HIGH | TODO |
| 10 | Test Master mode | HIGH | TODO |
| 11 | Test Outstation mode | HIGH | TODO |

### 7.2 Implementation Sequence

```
1. tui/colors.go - Define color constants
2. tui/render.go - Basic screen rendering
3. tui/layout.go - Layout manager
4. tui/input.go - Keyboard input
5. tui/table.go - Data table
6. tui/statusbar.go - Status bar
7. tui/log.go - Log display
8. tui/app.go - Main application
9. main.go - Update entry point
10. Integration testing
```

---

## 8. Testing Specification

### 8.1 Unit Tests

| Test | Description |
|------|-------------|
| `TestColorCodes` | Verify ANSI codes are correct |
| `TestLayoutResize` | Test layout calculations |
| `TestTableNavigation` | Test cursor movement |
| `TestKeyParsing` | Test key input parsing |

### 8.2 Integration Tests

| Test | Description |
|------|-------------|
| `TestMasterMode` | Run in master mode |
| `TestOutstationMode` | Run in outstation mode |
| `TestConcurrentUpdate` | Update while rendering |

### 8.3 Manual Test Plan

```bash
# Test basic rendering
./workbench --mode master

# Test keyboard input
./workbench --mode outstation

# Test resize handling
resize terminal, run workbench

# Test parallel execution
./workbench --mode master &
./workbench --mode outstation &
```

---

## 9. Performance Targets

| Metric | Target |
|--------|--------|
| Refresh Rate | 30-60 fps |
| Latency | < 16ms per frame |
| Memory | < 10MB |
| Startup | < 100ms |

---

## 10. Portability

### 10.1 Supported Platforms

| Platform | Status | Notes |
|----------|--------|-------|
| Linux (xterm) | Tested | Full support |
| Linux (tmux) | Tested | Full support |
| macOS (Terminal) | Supported | Should work |
| Windows (cmd) | Partial | Limited colors |
| Windows (PowerShell) | Supported | Full support |
| Windows Terminal | Supported | Full support |

### 10.2 Terminal Requirements

- ANSI escape code support
- Minimum 80x24 characters
- UTF-8 support (optional for ASCII)

---

## 11. Specification Summary

### 11.1 Core Features

1. **Layout**: Grid-based with header, table, log, controls
2. **Rendering**: Custom ANSI escape codes
3. **Input**: Keyboard-only navigation
4. **Data Display**: Scrollable table with multiple data types
5. **Status**: Real-time connection and mode display

### 11.2 Mode Support

| Feature | Master | Outstation |
|---------|--------|------------|
| Connect/Disconnect | ✓ | - |
| Start/Stop Server | - | ✓ |
| Read Class | ✓ | - |
| Data Points | ✓ | ✓ |
| Log | ✓ | ✓ |

---

*Specification created by DNP3-INV-063 Investigation*
*Status: PENDING APPROVAL*
