---
id: DNP3-INV-063
type: investigation
title: "Terminal-based Workbench (TUI) Development"
status: completed
authority: "KDE Runtime (DNP3 Library)"
created: "2026-07-28T01:30:00Z"
completed: "2026-07-28T01:45:00Z"
supersedes: "DNP3-INV-062"
reason: "Simpler approach using terminal/TUI instead of graphical framework"
---

# DNP3-INV-063: Terminal-based Workbench (TUI) Development

**Investigation ID**: DNP3-INV-063
**Title**: Terminal-based Workbench (TUI) Development
**Authority**: KDE Runtime (DNP3 Library)
**Status**: IN PROGRESS
**Date**: 2026-07-28
**Execution Agent**: OpenHands Agent

---

## 1. Executive Summary

### 1.1 Problem Statement

The previous approach (DNP3-INV-062) used Fyne framework for GUI development, which:
- Required complex X11 dependencies
- Had API compatibility issues
- Was difficult to test in headless environments
- Added unnecessary complexity for a workbench tool

### 1.2 Proposed Solution

Use a terminal-based UI (TUI) similar to `top`, `htop`, or `tmux`:
- Simple, lightweight, and fast
- Works in any terminal
- No external dependencies beyond standard libraries
- Easier to test and debug
- Unix philosophy: do one thing well

### 1.3 User Requirements

| Requirement | Description |
|-------------|-------------|
| Terminal-based | Display in terminal like `top` |
| Real-time updates | Live data refresh |
| Master/Outstation separation | Same as DNP3-INV-062 |
| Interactive controls | Keyboard navigation |
| Cross-platform | Works on Windows, Linux, macOS |

---

## 2. Design Specifications

### 2.1 Architecture

```
┌─────────────────────────────────────────────────┐
│                 WORKBENCH TUI                     │
├─────────────────────────────────────────────────┤
│  Header: Mode, Connection Status, Time           │
├─────────────────────────────────────────────────┤
│                                                  │
│  Main Panel:                                     │
│  ┌────────────────────────────────────────────┐ │
│  │ Data Points Table                          │ │
│  │ Binary Inputs | Analog | Counters | B/O     │ │
│  │─────────────────────────────────────────────│ │
│  │ BI0: true   | AI0: 100.5 | C0: 1000 | BO0 │ │
│  │ BI1: false  | AI1: 50.25 | C1: 500  | BO1 │ │
│  └────────────────────────────────────────────┘ │
│                                                  │
├─────────────────────────────────────────────────┤
│  Footer: Log messages, Errors                   │
├─────────────────────────────────────────────────┤
│  Controls: [q]uit [c]onnect [r]ead [s]end      │
└─────────────────────────────────────────────────┘
```

### 2.2 Key Components

| Component | Description |
|-----------|-------------|
| **Layout** | Grid-based layout with regions |
| **Table** | Scrollable data table |
| **StatusBar** | Connection, mode, errors |
| **InputHandler** | Keyboard event handling |
| **RefreshLoop** | Periodic UI update |

### 2.3 TUI Framework Options

| Framework | Pros | Cons |
|-----------|------|------|
| **bubbletea** (Golang) | Simple, declarative | New API |
| **tview** (Golang) | Rich widgets | Complex |
| **termbox** (C) | Fast, simple | Low-level |
| **cview** (Golang) | Built on termbox | Limited |
| **custom** | Full control | More work |

### 2.4 Recommended Approach

**Build our own simple TUI** using:
- Standard library: `fmt`, `os`, `time`, `sync`
- ANSI escape codes for terminal control
- Signal handling for graceful shutdown
- Goroutines for concurrent updates

Rationale: We only need basic table + status display. A custom solution is ~200 lines vs. framework overhead.

---

## 3. Implementation Plan

### 3.1 Directory Structure

```
cmd/workbench/
├── main.go              # Entry point with --mode flag
├── tui/
│   ├── app.go           # Main TUI application
│   ├── layout.go        # Screen layout manager
│   ├── table.go         # Data table widget
│   ├── statusbar.go     # Status bar widget
│   ├── input.go         # Input handler
│   └── render.go        # Rendering with ANSI
├── master/
│   └── controller.go    # Master-specific logic
└── outstation/
    └── controller.go    # Outstation-specific logic
```

### 3.2 Core Interface

```go
// TUI Application interface
type Application interface {
    Run() error
    Stop()
    Refresh()
    SetLayout(Layout)
}

// Layout regions
type Layout struct {
    Header   Region
    Main     Region
    Footer   Region
    Controls Region
}

// Data display
type DataTable struct {
    Columns []Column
    Rows    []Row
    Cursor  int
}
```

### 3.3 Keyboard Controls

| Key | Action |
|-----|--------|
| `q` | Quit |
| `c` | Connect/Disconnect |
| `r` | Read Class 0 |
| `1-3` | Read Class 1-3 |
| `↑/↓` | Navigate table |
| `Enter` | Select point |
| `l` | Clear log |
| `h` | Help |

---

## 4. Investigation Artifacts

| Artifact | Description | Status |
|----------|-------------|--------|
| [SPEC.md](SPEC.md) | Investigation specification | TODO |
| [CONCLUSION.md](CONCLUSION.md) | Investigation conclusions | TODO |

---

## 5. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Terminal compatibility | LOW | MEDIUM | Test on multiple terminals |
| ANSI code support | LOW | LOW | Detect terminal capabilities |
| Performance | LOW | LOW | 60fps refresh is achievable |
| Windows support | MEDIUM | LOW | Windows Terminal has good support |

---

## 6. Next Steps

1. [ ] Create SPEC.md with detailed implementation
2. [ ] Design TUI layout system
3. [ ] Implement basic rendering
4. [ ] Add Master mode
5. [ ] Add Outstation mode
6. [ ] Test in terminal

---

*Investigation initiated: 2026-07-28*
*Engineering Diagnosis: In Progress*
