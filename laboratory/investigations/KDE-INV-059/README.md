# KDE-INV-059: Workbench GUI Usability Issues

**Investigation**: KDE-INV-059  
**Domain**: UI/UX, Workbench Application  
**Status**: open  
**Created**: 2026-07-27  
**Severity**: HIGH  
**Engine**: kde-engine-ux-001

---

## Executive Summary

The DNP3 Engineering Workbench Windows GUI has multiple usability issues that make it unusable:
1. Missing native window decorations (menu bar, minimize/maximize/close buttons)
2. Confusing Master/Outstation mode selection with incorrect UI behavior
3. Data display uses simple labels instead of table view with checkboxes

---

## Issue 1: Missing Native Window Decorations

### Problem
The Windows GUI does not show a native menu bar or window controls (minimize, maximize, restore, close). This makes the application unusable on Windows.

### Evidence

**File**: `cmd/workbench/main.go`

```go
// Line 39-44
// Apply saved theme or default to light (UX Standard: Platform consistency)
// Fyne uses native decorations by default on all platforms
if cfg.Appearance.Theme == "Dark" {
    a.Settings().SetTheme(theme.DarkTheme())
} else {
    a.Settings().SetTheme(theme.LightTheme())
}
```

The comment claims "Fyne uses native decorations by default" but this is not happening.

**File**: `cmd/workbench/internal/ui/window.go`

```go
// Line 62 - creates window without explicit decoration settings
window:  app.NewWindow("DNP3 Engineering Workbench"),
```

### Root Cause

Fyne uses custom decorations by default since v2.0. To use native decorations, the window must be created with `fyne.NewWindow()` with proper settings OR the app must explicitly disable custom titlebar.

### Expected Behavior

On Windows, the application should show:
- Native window frame with minimize, maximize, restore, close buttons
- Native menu bar with File, Edit, View, Session, Help menus
- Native title bar with application name

### Resolution

Enable native window decorations in Fyne by either:
1. Using `app.NewWindow()` with proper settings
2. Adding `NativeDecorations: true` to window properties

---

## Issue 2: Master/Outstation Mode Confusion

### Problem
The Mode Selection panel offers both "Master Mode" and "Outstation Mode" but:
1. MVP is "Master Mode Only" (per code comment)
2. Outstation nodes don't initiate connections - they LISTEN for connections
3. The Connect button doesn't change behavior based on mode

### Evidence

**File**: `cmd/workbench/internal/ui/panels/mode.go`

```go
// Line 36-37
note := widget.NewLabel("MVP: Master Mode Only")
note.TextStyle.Italic = true

p.masterRB = widget.NewRadioGroup([]string{"Master Mode", "Outstation Mode"}, func(selected string) {
    p.isMaster = (selected == "Master Mode")
    if p.onModeChange != nil {
        p.onModeChange(p.isMaster)
    }
})
```

**File**: `cmd/workbench/internal/ui/panels/connection.go`

The connection panel always shows "Connect" button regardless of mode selection.

### DNP3 Protocol Context

| Role | Behavior | UI Implication |
|------|----------|----------------|
| **Master** | Initiates TCP connection to Outstation | Show Connect button |
| **Outstation** | Listens for TCP connections | Show "Start Server" button, NOT Connect |

### Root Cause

Confusing terminology and incorrect UI mapping. In DNP3:
- A **Master** is a client that polls an Outstation
- An **Outstation** is a server that responds to the Master

The current UI calls them "Master Mode" and "Outstation Mode" which is backwards semantically.

### Resolution Options

**Option A (Recommended)**: Simplify to Master-only for MVP
- Remove Outstation Mode option entirely
- Update UI to clarify this is a Master polling tool

**Option B**: Implement both roles properly
- Master Mode → Connect button (TCP client)
- Outstation Mode → Start Server button (TCP server)
- Separate the concepts into distinct roles

---

## Issue 3: Data Panel Lacks Table View

### Problem
The Point Values panel shows simple text labels instead of a proper table with checkboxes for data monitoring.

### Evidence

**File**: `cmd/workbench/internal/ui/panels/data.go`

```go
// Lines 30-32 - Simple labels instead of table
p.binaryLabel = widget.NewLabel("Binary Inputs: (No data)")
p.analogLabel = widget.NewLabel("Analog Inputs: (No data)")
p.counterLabel = widget.NewLabel("Counters: (No data)")
```

### Expected Behavior

User expects a table view with:

| Index | AI | AO | DI | DO | Counters | Value | Quality | Timestamp |
|-------|----|----|----|----|----------|-------|---------|-----------|
| 0 | ☐ | ☐ | ☑ | ☐ | ☐ | 42.5 | Good | 10:23:45 |
| 1 | ☑ | ☐ | ☐ | ☐ | ☑ | 1234 | Good | 10:23:45 |

Checkboxes allow the user to:
- Select which points to monitor
- Enable/disable polling for specific points
- Batch operations on selected points

### DNP3 Data Types

| Abbr | Name | Description |
|------|------|-------------|
| AI | Analog Input | 16/32-bit integer or float values |
| AO | Analog Output | Control values written to outstation |
| DI | Binary Input | Single-bit status inputs |
| DO | Binary Output | Single-bit control outputs |
| Counters | Counter | 16/32-bit accumulating values |

### Resolution

Replace simple labels with Fyne's `widget.Table` or `widget.Tree` with:
- Column headers for each data type
- Checkboxes for point selection
- Sortable columns
- Scrollable view

---

## Related Files

| File | Issue | Relevance |
|------|-------|-----------|
| `cmd/workbench/main.go` | Window decorations | HIGH |
| `cmd/workbench/internal/ui/window.go` | Window setup | HIGH |
| `cmd/workbench/internal/ui/panels/mode.go` | Master/Outstation confusion | HIGH |
| `cmd/workbench/internal/ui/panels/connection.go` | Connect button behavior | HIGH |
| `cmd/workbench/internal/ui/panels/data.go` | Table view needed | HIGH |

---

## Investigation Notes

### UX Standards Reference
The code references "UX Standard Section X" comments but the actual UX standard document may not exist or be enforced.

### Fyne Framework Considerations
- Native decorations require explicit configuration in Fyne v2.x
- Table widget support exists but requires custom implementation for checkboxes

---

## Status History

| Date | Status | Notes |
|------|--------|-------|
| 2026-07-27 | open | Initial investigation |
