# KDE-DEC-060: Workbench UI Redesign Decision Record

**Decision**: KDE-DEC-060  
**Date**: 2026-07-27  
**Status**: approved  
**Engine**: kde-engine-ux-001  
**Investigation**: KDE-INV-059, KDE-INV-060  

---

## Context

The DNP3 Engineering Workbench has usability issues:
1. Missing native window decorations on Windows
2. Confusing Master/Outstation mode terminology
3. Data displayed as simple labels instead of table view
4. No write/control capabilities in UI

## Decision

Implement the following changes to the Workbench UI:

### 1. Window Decorations

**Decision**: Use Fyne's native window decorations.

**Rationale**:
- Fyne v2.x uses custom decorations by default
- Native decorations provide proper minimize/maximize/close
- Native menus work correctly on all platforms

**Implementation**: 
```go
// In main.go
w := app.NewWindow("DNP3 Engineering Workbench")
// Default is native on Windows, explicit on Linux/macOS
```

### 2. Mode Separation

**Decision**: Clearly separate "Poll Outstation" (Master) and "Simulate Outstation" modes.

**Rationale**:
- Current "Master Mode" / "Outstation Mode" is backwards
- Master connects TO outstation, doesn't host one
- User confusion resolved by clear action names

**Implementation**:
| Mode | Label | Action | Button |
|------|-------|--------|--------|
| Master | "Poll Outstation" | Connect to remote outstation | "Connect" |
| Outstation | "Simulate Outstation" | Listen for master connections | "Start Server" |

### 3. Data Table View

**Decision**: Replace labels with a virtualized table with checkboxes.

**Rationale**:
- Real DNP3 systems have hundreds of points
- Checkboxes allow selection for operations
- Table provides sortable, scannable view

**Implementation**:
- Use Fyne `widget.NewTable` with custom cell factory
- Columns: Select | Index | DI | DO | AI | AO | CTR | Value | Quality | Time
- Checkboxes in DI/DO/AI/AO/CTR columns indicate point type at that index

### 4. Control Panel

**Decision**: Add a control panel for write operations.

**Rationale**:
- Users need to operate binary outputs (DO)
- Analog outputs (AO) need value setting
- Selection via table checkboxes enables targeted control

**Implementation**:
- Panel shows selected output points
- Operate buttons for DO (ON/OFF)
- Value input for AO
- Select-Then-Operate safety toggle

---

## Consequences

### Positive
- Usable Windows application with native decorations
- Clear terminology matching DNP3 roles
- Scalable data view for real-world use
- Functional control capabilities

### Negative
- Significant UI refactoring required
- New OutstationSession needs implementation
- Testing complexity increases

---

## Alternatives Considered

### Alternative 1: Keep Labels, Add Dialogs
- Pros: Minimal code change
- Cons: Doesn't scale, poor UX

### Alternative 2: Use Web UI (HTML/JS)
- Pros: Better cross-platform, modern UI
- Cons: New frontend stack, IPC complexity

### Alternative 3: Qt Instead of Fyne
- Pros: Mature, native look
- Cons: Go bindings less mature, larger binary

---

## Implementation Plan

| Phase | Task | Files |
|-------|------|-------|
| 1 | Create OutstationSession | session/outstation.go |
| 2 | Update ModePanel | panels/mode.go |
| 3 | Create DataTablePanel | panels/datatable.go |
| 4 | Create ControlPanel | panels/control.go |
| 5 | Create OutstationPanel | panels/outstation.go |
| 6 | Update window layout | ui/window.go |
| 7 | Test and verify | - |

---

## Status History

| Date | Status | Notes |
|------|--------|-------|
| 2026-07-27 | approved | Initial decision |

---

## Sign-off

| Role | Name | Date |
|------|------|------|
| Approver | OpenHands | 2026-07-27 |
