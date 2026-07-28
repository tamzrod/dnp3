# DNP3-INV-063 Conclusion

**Investigation ID**: DNP3-INV-063
**Status**: IN PROGRESS
**Date**: 2026-07-28

---

## 1. Summary

### 1.1 Decision

Replace Fyne-based GUI (DNP3-INV-062) with terminal-based TUI approach.

### 1.2 Rationale

| Factor | Fyne (INV-062) | TUI (INV-063) |
|--------|----------------|----------------|
| Dependencies | Heavy (X11, OpenGL) | None (ANSI) |
| Testing | Requires display | Terminal only |
| Cross-platform | Complex | Simple |
| Complexity | High API surface | Simple, custom |
| Performance | Good | Excellent |
| User skill | GUI expertise | Terminal knowledge |

### 1.3 Proposed Solution

Build a custom TUI using ANSI escape codes:
- Similar to `top`, `htop`, `btop`
- Full terminal control
- No external dependencies beyond Go stdlib

---

## 2. Architecture

### 2.1 Layout

```
┌─────────────────────────────────────────────────┐
│ [MASTER] DNP3 Workbench │ Connected │ 10:23:45 │
├─────────────────────────────────────────────────┤
│ ┌─ DATA POINTS ────────────────────────────┐  │
│ │ Type │ Index │ Value │ Quality │ Time    │  │
│ │──────────────────────────────────────────│  │
│ │ BI   │ 0     │ true  │ ONLINE │ 10:23  │  │
│ │ AI   │ 0     │ 100.5 │ ONLINE │ 10:23  │  │
│ └──────────────────────────────────────────┘  │
├─────────────────────────────────────────────────┤
│ LOG: [10:23] → READ Class 0                 │
├─────────────────────────────────────────────────┤
│ [q]uit [c]onnect [r]ead [↑↓]nav              │
└─────────────────────────────────────────────────┘
```

### 2.2 Components

| Component | Description |
|-----------|-------------|
| `tui/app.go` | Main application loop |
| `tui/render.go` | ANSI rendering |
| `tui/table.go` | Data table widget |
| `tui/statusbar.go` | Status display |
| `tui/input.go` | Keyboard handling |

---

## 3. Implementation Status

| Phase | Status |
|-------|--------|
| Investigation | IN PROGRESS |
| Specification | ✅ COMPLETE |
| Implementation | PENDING |
| Testing | PENDING |
| Deployment | PENDING |

---

## 4. Next Actions

1. [ ] Await approval of specification
2. [ ] Create tui package structure
3. [ ] Implement ANSI rendering
4. [ ] Create layout manager
5. [ ] Build table widget
6. [ ] Add keyboard input
7. [ ] Test Master mode
8. [ ] Test Outstation mode

---

*Investigation conclusion: PENDING*
