# DNP3 Workbench UX Implementation Plan

**Plan ID**: DNP3-WORKBENCH-UX-IMPLEMENTATION-001  
**Reference**: UX-DESKTOP-ENGINEERING-001, DNP3-WORKBENCH-UX-ASSESSMENT-001, DNP3-WORKBENCH-UX-ROADMAP-001  
**Status**: DRAFT  
**Date**: 2026-07-27  

---

## 1. Introduction

This implementation plan provides detailed tasks for modernizing the DNP3 Workbench UI. Tasks are prioritized by value and risk, following the phased approach defined in the Modernization Roadmap.

### 1.1 Task Format

Each task includes:
- **Title**: Clear, actionable name
- **Purpose**: Why this task exists (traced to Expert Knowledge Entry)
- **Affected Files**: Specific files to modify
- **Architectural Impact**: Low/Medium/High
- **Estimated Complexity**: Hours and difficulty rating
- **Acceptance Criteria**: Testable conditions for completion

### 1.2 Prioritization Strategy

1. **HIGH priority**: Foundation features (menus, shortcuts) with low risk
2. **MEDIUM priority**: Efficiency features with moderate complexity
3. **LOW priority**: Advanced features requiring significant work

---

## 2. Phase 1 Tasks: Desktop Foundation

### Task 1.1: Implement Complete Menu Bar

**Title**: Implement Complete Menu Bar

**Purpose**: Establish standard desktop application menu structure for discoverability.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 4.1-4.4  
> File, Edit, View, Tools, Help menus are standard in all professional tools

**Affected Files**:
- `cmd/workbench/main.go`

**Architectural Impact**: LOW

**Estimated Complexity**: 6 hours | MEDIUM

**Implementation Details**:
```go
// File Menu
├── New Session           Ctrl+N
├── Open Configuration... Ctrl+O
├── Open Recent      →    (submenu)
├── ─────────────
├── Save Configuration   Ctrl+S
├── Save As...
├── ─────────────
├── Export Log...
├── ─────────────
└── Exit                  Alt+F4

// Edit Menu
├── Undo                  Ctrl+Z
├── Redo                  Ctrl+Y
├── ─────────────
├── Cut                   Ctrl+X
├── Copy                  Ctrl+C
├── Paste                 Ctrl+V
├── ─────────────
├── Find in Log           Ctrl+F
└── Select All            Ctrl+A

// View Menu
├── Zoom In               Ctrl++
├── Zoom Out              Ctrl+-
├── ─────────────
├── Sidebar               (toggle)
├── Log Panel             (toggle)
├── ─────────────
└── Fullscreen            F11

// Session Menu
├── Connect               F5
├── Disconnect            F6
├── ─────────────
├── Read Class 0          (all enabled when connected)
├── Read Class 1
├── Read Class 2
├── Read Class 3
├── ─────────────
├── Clear Log
└── ─────────────

// Help Menu
├── Documentation         F1
├── Keyboard Shortcuts
├── ─────────────
└── About DNP3 Workbench
```

**Acceptance Criteria**:
- [ ] All menu items visible and labeled correctly
- [ ] Menu separators at appropriate locations
- [ ] Keyboard shortcuts displayed next to menu items
- [ ] Menu items trigger correct actions
- [ ] Disabled menu items shown grayed out when appropriate

---

### Task 1.2: Register Keyboard Shortcuts

**Title**: Register Keyboard Shortcuts

**Purpose**: Enable efficient keyboard-driven operation for power users.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 4.2-4.4  
> Standard shortcuts (Ctrl+S, Ctrl+C) reduce learning curve

**Affected Files**:
- `cmd/workbench/main.go`
- `cmd/workbench/internal/ui/window.go`

**Architectural Impact**: LOW

**Estimated Complexity**: 4 hours | LOW

**Implementation Details**:
```go
// Register shortcuts via Fyne shortcuts API
shortcuts := []fyne.Shortcut{
    &fyne.ShortcutCut{},
    &fyne.ShortcutCopy{},
    &fyne.ShortcutPaste{},
    &fyne.ShortcutSelectAll{},
    &fyne.ShortcutFind{},
}

// Custom shortcuts
window.Canvas().AddShortcut(&fyne.KeyShortcut{
    Key: fyne.KeyF5,
    Modifiers: 0,
}, func(e *fyne.KeyEvent) {
    ctrl.Connect(lastAddress, lastPort)
})

window.Canvas().AddShortcut(&fyne.KeyShortcut{
    Key: fyne.KeyF6,
    Modifiers: 0,
}, func(e *fyne.KeyEvent) {
    ctrl.Disconnect()
})

window.Canvas().AddShortcut(&fyne.KeyShortcut{
    Key: fyne.KeyF11,
    Modifiers: 0,
}, func(e *fyne.KeyEvent) {
    window.ToggleFullscreen()
})
```

**Acceptance Criteria**:
- [ ] Ctrl+S saves configuration (stub implementation OK)
- [ ] Ctrl+O opens file dialog (stub implementation OK)
- [ ] F5 connects with current settings
- [ ] F6 disconnects
- [ ] F11 toggles fullscreen
- [ ] Ctrl+F opens Find in Log
- [ ] Escape closes dialogs

---

### Task 1.3: Configure Window Constraints

**Title**: Configure Window Constraints

**Purpose**: Prevent unusable UI states and enable proper multi-monitor support.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 3.2  
> Minimum Size: 800×600 (prevents unusable UI)

**Affected Files**:
- `cmd/workbench/main.go`
- `cmd/workbench/internal/ui/window.go`

**Architectural Impact**: LOW

**Estimated Complexity**: 2 hours | LOW

**Implementation Details**:
```go
// In main.go or window.go
const (
    MinWindowWidth  = 800
    MinWindowHeight = 600
)

func setupWindow() {
    window := ui.NewMainWindow(app, controller)
    
    // Set minimum size
    window.SetMinSize(fyne.NewSize(MinWindowWidth, MinWindowHeight))
    
    // Set starting size
    window.Resize(fyne.NewSize(1200, 800))
    
    // Center on screen
    window.CenterOnScreen()
}
```

**Acceptance Criteria**:
- [ ] Window cannot be resized below 800×600
- [ ] Window starts at 1200×800 on first launch
- [ ] Window centers on primary monitor

---

### Task 1.4: Create About Dialog

**Title**: Create About Dialog

**Purpose**: Provide standard application information to users.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 4.1  
> Help menu should include About option

**Affected Files**:
- `cmd/workbench/main.go` (menu item)
- `cmd/workbench/internal/ui/dialogs/about.go` (new file)

**Architectural Impact**: LOW

**Estimated Complexity**: 2 hours | LOW

**Implementation Details**:
```go
// dialogs/about.go
package dialogs

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/theme"
)

func ShowAbout(parent fyne.Window) {
    content := container.NewVBox(
        widget.NewLabelWithStyle("DNP3 Engineering Workbench", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
        widget.NewLabel("Version 0.1.0"),
        widget.NewLabel(""),
        widget.NewLabel("A desktop application for validating and debugging"),
        widget.NewLabel("the native Go DNP3 library."),
        widget.NewLabel(""),
        widget.NewLabel("Built with Fyne"),
    )
    
    dialog.ShowCustom("About", "Close", content, parent)
}
```

**Acceptance Criteria**:
- [ ] About dialog opens from Help menu
- [ ] Dialog shows application name, version
- [ ] Dialog has Close button
- [ ] Dialog is modal (blocks main window)

---

## 3. Phase 2 Tasks: Navigation and Feedback

### Task 2.1: Enhance Status Bar

**Title**: Enhance Status Bar with Visual Indicators

**Purpose**: Provide immediate visual feedback on connection state.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 7.1  
> Connection/state indicator with visual feedback

**Affected Files**:
- `cmd/workbench/internal/ui/panels/statusbar.go`

**Architectural Impact**: LOW

**Estimated Complexity**: 6 hours | MEDIUM

**Implementation Details**:
```go
// Add visual indicator to status bar
type StatusBar struct {
    container *fyne.Container
    state     binding.String
    connection binding.String
    iin       binding.String
    
    // New visual elements
    connectionIndicator *widget.RichText // Colored dot/text
    progressBar         *widget.ProgressBar
    errorLabel          *widget.Label
}

func (sb *StatusBar) SetConnectionState(state session.ConnectionState) {
    // Green = Connected
    // Yellow = Connecting
    // Red = Error
    // Gray = Disconnected
}
```

**Acceptance Criteria**:
- [ ] Connection indicator shows colored icon
- [ ] Icon color reflects connection state
- [ ] Progress bar visible during operations
- [ ] Error messages displayed prominently

---

### Task 2.2: Implement Command State Management

**Title**: Implement Command State Management

**Purpose**: Prevent invalid operations by disabling unavailable commands.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 6.5  
> Disable controls when action unavailable

**Affected Files**:
- `cmd/workbench/internal/ui/panels/commands.go`
- `cmd/workbench/internal/controller/controller.go`

**Architectural Impact**: LOW

**Estimated Complexity**: 4 hours | LOW

**Implementation Details**:
```go
// commands.go
type CommandPanel struct {
    container *fyne.Container
    // ... existing fields
    
    // Add button references
    readClass0Btn  *widget.Button
    readClass1Btn  *widget.Button
    operateOnBtn   *widget.Button
    operateOffBtn  *widget.Button
}

func (cp *CommandPanel) SetConnected(connected bool) {
    if connected {
        cp.readClass0Btn.Enable()
        cp.readClass1Btn.Enable()
        cp.operateOnBtn.Enable()
        cp.operateOffBtn.Enable()
    } else {
        cp.readClass0Btn.Disable()
        cp.readClass1Btn.Disable()
        cp.operateOnBtn.Disable()
        cp.operateOffBtn.Disable()
    }
}
```

**Acceptance Criteria**:
- [ ] Read commands disabled when disconnected
- [ ] Operate commands disabled when disconnected
- [ ] Toolbar buttons reflect connection state
- [ ] Menu items disabled when appropriate

---

### Task 2.3: Add Error Notification System

**Title**: Add Error Notification System

**Purpose**: Make errors visible without requiring log inspection.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 7.4  
> User-visible error display with clear messages

**Affected Files**:
- `cmd/workbench/internal/controller/controller.go`
- `cmd/workbench/internal/ui/window.go`
- `cmd/workbench/internal/ui/panels/statusbar.go`

**Architectural Impact**: LOW

**Estimated Complexity**: 4 hours | MEDIUM

**Implementation Details**:
```go
// controller.go
type Controller struct {
    // ... existing fields
    onError func(error) // Callback for UI notification
}

// Handle errors with both logging and notification
func (c *Controller) handleError(format string, args ...interface{}) {
    errMsg := fmt.Sprintf(format, args...)
    c.logger.Error(errMsg)
    
    if c.onError != nil {
        c.onError(fmt.Errorf(errMsg))
    }
}

// window.go
func (w *MainWindow) Show() {
    w.ctrl.OnError = func(err error) {
        w.statusBar.ShowError(err.Error())
    }
}
```

**Acceptance Criteria**:
- [ ] Connection errors appear in status bar
- [ ] Error text is user-friendly (no technical jargon)
- [ ] Errors auto-clear after 30 seconds or on next action
- [ ] Errors still logged for debugging

---

### Task 2.4: Enable Keyboard Navigation

**Title**: Enable Keyboard Navigation

**Purpose**: Support tab-based navigation for accessibility.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 6.6  
> Tab navigation between controls

**Affected Files**:
- `cmd/workbench/internal/ui/panels/*.go`

**Architectural Impact**: LOW

**Estimated Complexity**: 3 hours | MEDIUM

**Implementation Details**:
Fyne widgets support keyboard navigation by default. This task focuses on ensuring proper focus order and visibility.

```go
// Ensure all interactive widgets are focusable
addressEntry := widget.NewEntry()
addressEntry.FocusGained()
addressEntry.FocusLost()

// Set focus chain in container
container := fyne.NewContainer(...)
container.FocusChain()
```

**Acceptance Criteria**:
- [ ] Tab cycles through all input fields
- [ ] Focus indicators visible
- [ ] Enter activates focused button
- [ ] Escape closes focused dialog

---

## 4. Phase 3 Tasks: Workflow Efficiency

### Task 3.1: Implement Toolbar

**Title**: Implement Toolbar for Frequent Actions

**Purpose**: Provide one-click access to common operations.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 5.1-5.4  
> Frequently-used actions in toolbar with tooltips

**Affected Files**:
- `cmd/workbench/internal/ui/toolbar.go` (new file)
- `cmd/workbench/internal/ui/window.go`

**Architectural Impact**: LOW

**Estimated Complexity**: 8 hours | MEDIUM

**Implementation Details**:
```go
// toolbar.go
package ui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/widget"
    "fyne.io/fyne/v2/container"
)

type Toolbar struct {
    container *fyne.Container
    connectBtn    *widget.Button
    disconnectBtn *widget.Button
    clearBtn      *widget.Button
}

func NewToolbar(ctrl *controller.Controller) *Toolbar {
    t := &Toolbar{}
    
    // Create icon buttons with tooltips
    t.connectBtn = widget.NewButtonWithIcon("Connect", theme.ConfirmIcon(), func() {
        ctrl.Connect(...)
    })
    t.connectBtn.SetTooltip("Connect to outstation (F5)")
    
    t.disconnectBtn = widget.NewButtonWithIcon("Disconnect", theme.CancelIcon(), func() {
        ctrl.Disconnect()
    })
    t.disconnectBtn.Disable()
    
    t.clearBtn = widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), func() {
        ctrl.ClearLog()
    })
    
    // Group with separators
    t.container = container.NewHBox(
        t.connectBtn,
        t.disconnectBtn,
        widget.NewSeparator(),
        t.clearBtn,
    )
    
    return t
}
```

**Acceptance Criteria**:
- [ ] Toolbar appears below menu bar
- [ ] Connect button enabled when disconnected
- [ ] Disconnect button enabled when connected
- [ ] Buttons show tooltips with keyboard shortcuts
- [ ] Icons are recognizable (standard Fyne icons OK)

---

### Task 3.2: Implement Panel Visibility Toggles

**Title**: Implement Panel Visibility Toggles

**Purpose**: Allow users to customize workspace layout.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 6.3  
> Collapsible panels for maximizing workspace

**Affected Files**:
- `cmd/workbench/internal/ui/window.go`
- `cmd/workbench/internal/ui/panels/*.go`

**Architectural Impact**: MEDIUM

**Estimated Complexity**: 6 hours | MEDIUM

**Implementation Details**:
```go
// window.go
type MainWindow struct {
    // ... existing fields
    
    // Panel visibility state
    sidebarVisible  bool
    logPanelVisible bool
}

func (w *MainWindow) ToggleSidebar() {
    w.sidebarVisible = !w.sidebarVisible
    w.updateLayout()
}

func (w *MainWindow) updateLayout() {
    // Rebuild content with current visibility
    var leftPanel fyne.CanvasObject
    if w.sidebarVisible {
        leftPanel = w.leftSidebar
    } else {
        leftPanel = nil
    }
    
    w.window.SetContent(...)
}
```

**Acceptance Criteria**:
- [ ] View menu has Sidebar toggle
- [ ] View menu has Log Panel toggle
- [ ] Toggle reflects current state (checked/unchecked)
- [ ] Layout updates immediately

---

### Task 3.3: Improve Log Panel

**Title**: Improve Log Panel Size and Behavior

**Purpose**: Provide more log space for troubleshooting.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 7.3  
> Log panel should be scrollable with sufficient height

**Affected Files**:
- `cmd/workbench/internal/ui/panels/log.go`

**Architectural Impact**: LOW

**Estimated Complexity**: 4 hours | LOW

**Implementation Details**:
```go
// Increase default log height
const (
    DefaultLogHeight = 200  // Was 150
    MinLogHeight      = 100
    MaxLogHeight      = 400
)

// Make log panel resizable via splitter
mainContent := container.NewVSplit(
    dataAndProtocol,
    w.logPanel.Container(),
)
mainContent.Offset = 0.7 // 70% to main content, 30% to log
```

**Acceptance Criteria**:
- [ ] Log panel taller by default (200px)
- [ ] Log panel resizable via drag handle
- [ ] Auto-scroll to bottom on new entries
- [ ] Clear button still functional

---

## 5. Phase 4 Tasks: Persistence

### Task 4.1: Create Config Package

**Title**: Create Config Package for Persistence

**Purpose**: Establish foundation for all user preference storage.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 8.1  
> Store in user's profile directory with platform-appropriate locations

**Affected Files**:
- `cmd/workbench/internal/config/config.go` (new file)
- `cmd/workbench/internal/config/persistence.go` (new file)

**Architectural Impact**: MODERATE

**Estimated Complexity**: 10 hours | MEDIUM

**Implementation Details**:
```go
// config/config.go
package config

import (
    "os"
    "path/filepath"
)

type Config struct {
    Version string
    
    Window struct {
        Width  int
        Height int
        X      int
        Y      int
        Max    bool
    }
    
    Layout struct {
        SidebarVisible  bool
        LogPanelVisible bool
        LogPanelHeight  int
    }
    
    RecentConnections []Connection
    
    LastConnection struct {
        Address string
        Port    int
    }
}

type Connection struct {
    Name    string
    Address string
    Port    int
    LastUsed time.Time
}

// persistence.go
func DefaultDir() string {
    switch runtime.GOOS {
    case "windows":
        return filepath.Join(os.Getenv("APPDATA"), "DNP3Workbench")
    case "darwin":
        return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "DNP3Workbench")
    default:
        return filepath.Join(os.Getenv("HOME"), ".config", "dnp3workbench")
    }
}

func Load() (*Config, error) {
    dir := DefaultDir()
    file := filepath.Join(dir, "config.json")
    
    data, err := os.ReadFile(file)
    if err != nil {
        if os.IsNotExist(err) {
            return Default(), nil
        }
        return nil, err
    }
    
    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}

func (c *Config) Save() error {
    dir := DefaultDir()
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }
    
    data, err := json.MarshalIndent(c, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(filepath.Join(dir, "config.json"), data, 0644)
}
```

**Acceptance Criteria**:
- [ ] Config stored in platform-appropriate directory
- [ ] Config survives application restart
- [ ] Invalid config falls back to defaults
- [ ] Config file is human-readable JSON

---

### Task 4.2: Implement Window Geometry Persistence

**Title**: Implement Window Geometry Persistence

**Purpose**: Restore user's preferred window size and position.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 8.2  
> Remember window size/position

**Affected Files**:
- `cmd/workbench/main.go`
- `cmd/workbench/internal/config/config.go`

**Architectural Impact**: LOW

**Estimated Complexity**: 4 hours | LOW

**Implementation Details**:
```go
// main.go
func main() {
    cfg, err := config.Load()
    if err != nil {
        cfg = config.Default()
    }
    
    // Create window
    window := ui.NewMainWindow(app, ctrl)
    
    // Restore geometry
    if cfg.Window.Width > 0 && cfg.Window.Height > 0 {
        window.Resize(fyne.NewSize(float32(cfg.Window.Width), float32(cfg.Window.Height)))
    } else {
        window.Resize(fyne.NewSize(1200, 800))
    }
    
    if cfg.Window.X >= 0 && cfg.Window.Y >= 0 {
        window.Move(fyne.NewPos(float32(cfg.Window.X), float32(cfg.Window.Y)))
    } else {
        window.CenterOnScreen()
    }
    
    if cfg.Window.Max {
        window.SetFullScreen(true)
    }
    
    // Save on close
    window.SetOnClosed(func() {
        geom := window. Geometry()
        cfg.Window.Width = int(geom.Width)
        cfg.Window.Height = int(geom.Height)
        pos := window.Position()
        cfg.Window.X = int(pos.X)
        cfg.Window.Y = int(pos.Y)
        cfg.Save()
    })
}
```

**Acceptance Criteria**:
- [ ] Window size saved on close
- [ ] Window position saved on close
- [ ] Geometry restored on next launch
- [ ] Invalid geometry falls back to defaults

---

### Task 4.3: Implement Recent Connections

**Title**: Implement Recent Connections List

**Purpose**: Quick access to frequently-used connection targets.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 8.2  
> Recent files/connections (up to 10)

**Affected Files**:
- `cmd/workbench/internal/config/config.go`
- `cmd/workbench/main.go` (menu item)
- `cmd/workbench/internal/ui/panels/connection.go`

**Architectural Impact**: LOW

**Estimated Complexity**: 4 hours | LOW

**Implementation Details**:
```go
// Add connection to recent list
func AddRecentConnection(cfg *config.Config, address string, port int) {
    conn := config.Connection{
        Address:  address,
        Port:     port,
        LastUsed: time.Now(),
    }
    
    // Remove existing entry with same address:port
    cfg.RecentConnections = removeDuplicate(cfg.RecentConnections, conn)
    
    // Add to front
    cfg.RecentConnections = append([]config.Connection{conn}, cfg.RecentConnections...)
    
    // Keep only 10
    if len(cfg.RecentConnections) > 10 {
        cfg.RecentConnections = cfg.RecentConnections[:10]
    }
    
    cfg.Save()
}

// Build menu with recent connections
func buildRecentMenu(cfg *config.Config, connectFn func(string, int)) *fyne.Menu {
    items := make([]*fyne.MenuItem, len(cfg.RecentConnections))
    for i, conn := range cfg.RecentConnections {
        label := fmt.Sprintf("%s:%d", conn.Address, conn.Port)
        items[i] = fyne.NewMenuItem(label, func() {
            connectFn(conn.Address, conn.Port)
        })
    }
    return fyne.NewMenu("Open Recent", items...)
}
```

**Acceptance Criteria**:
- [ ] Successful connection adds to recent list
- [ ] Recent list limited to 10 entries
- [ ] Open Recent submenu shows all recent connections
- [ ] Selecting recent item connects to that target

---

## 6. Phase 5 Tasks: Advanced Productivity

### Task 5.1: Implement Log Filtering

**Title**: Implement Log Level and Content Filtering

**Purpose**: Enable efficient log analysis for troubleshooting.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 7.3  
> Log level filtering (debug, info, warn, error)

**Affected Files**:
- `cmd/workbench/internal/ui/panels/log.go`

**Architectural Impact**: MEDIUM

**Estimated Complexity**: 8 hours | MEDIUM

**Implementation Details**:
```go
type LogPanel struct {
    // ... existing fields
    
    // Filter controls
    levelFilter *widget.Select
    searchEntry *widget.Entry
    
    // Filtered entries
    filteredEntries []LogEntry
}

func (lp *LogPanel) ApplyFilters() {
    lp.mu.Lock()
    defer lp.mu.Unlock()
    
    lp.filteredEntries = nil
    level := lp.selectedLevel
    search := strings.ToLower(lp.searchEntry.Text)
    
    for _, e := range lp.entries {
        if !lp.matchesFilter(e, level, search) {
            continue
        }
        lp.filteredEntries = append(lp.filteredEntries, e)
    }
    
    lp.list.Refresh()
}
```

**Acceptance Criteria**:
- [ ] Filter dropdown shows: All, Debug, Info, Warn, Error
- [ ] Selecting filter updates displayed entries
- [ ] Search filters entries containing text
- [ ] Combined filter AND logic works

---

### Task 5.2: Implement Log Search

**Title**: Implement Log Search Functionality

**Purpose**: Find specific events in large logs.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 7.3  
> Search within log

**Affected Files**:
- `cmd/workbench/internal/ui/panels/log.go`

**Architectural Impact**: LOW

**Estimated Complexity**: 6 hours | MEDIUM

**Implementation Details**:
```go
func (lp *LogPanel) FindNext(searchText string) bool {
    current := lp.list.ScrolledRow() // Get current scroll position
    
    for i := current + 1; i < len(lp.filteredEntries); i++ {
        if strings.Contains(strings.ToLower(lp.filteredEntries[i].Message), searchText) {
            lp.list.ScrollTo(i)
            lp.list.Select(i)
            return true
        }
    }
    return false // Wrapped around
}

func (lp *LogPanel) FindPrevious(searchText string) bool {
    // Similar but iterating backwards
}
```

**Acceptance Criteria**:
- [ ] Ctrl+F opens search bar
- [ ] Enter finds next match
- [ ] Shift+Enter finds previous match
- [ ] Search wraps around
- [ ] Matches are highlighted

---

### Task 5.3: Implement Log Export

**Title**: Implement Log Export to File

**Purpose**: Enable offline log analysis.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 7.3  
> Export to file

**Affected Files**:
- `cmd/workbench/main.go` (menu item)
- `cmd/workbench/internal/ui/panels/log.go`

**Architectural Impact**: LOW

**Estimated Complexity**: 6 hours | MEDIUM

**Implementation Details**:
```go
func (lp *LogPanel) ExportLog(filename string) error {
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer f.Close()
    
    buf := bufio.NewWriter(f)
    for _, e := range lp.entries {
        line := fmt.Sprintf("[%s] %s %s\n",
            e.Timestamp.Format(time.RFC3339),
            e.Direction,
            e.Message)
        if _, err := buf.WriteString(line); err != nil {
            return err
        }
    }
    return buf.Flush()
}

// Called from menu
menuItem := fyne.NewMenuItem("Export Log...", func() {
    dialog.ShowFileSave(func(uri fyne.URIWriteCloser, err error) {
        if err != nil || uri == nil {
            return
        }
        defer uri.Close()
        
        lp.ExportLog(uri.URI().Path())
    }, window)
})
```

**Acceptance Criteria**:
- [ ] Export Log menu item exists
- [ ] File dialog opens for save location
- [ ] Log written in readable format
- [ ] Success notification shown

---

### Task 5.4: Create Settings Dialog

**Title**: Create Settings/Preferences Dialog

**Purpose**: Centralize user preferences in one place.

**Expert Knowledge Reference**: UX-DESKTOP-ENGINEERING-001, Section 4.5  
> Settings/Preferences dialog

**Affected Files**:
- `cmd/workbench/internal/ui/dialogs/settings.go` (new file)
- `cmd/workbench/internal/config/config.go`

**Architectural Impact**: LOW

**Estimated Complexity**: 8 hours | MEDIUM

**Implementation Details**:
```go
// dialogs/settings.go
package dialogs

type SettingsDialog struct {
    dialog    *dialog.CustomDialog
    container *fyne.Container
    
    // Settings widgets
    autoScroll *widget.Check
    logLevel   *widget.Select
    timestamp  *widget.Check
}

func NewSettingsDialog(parent fyne.Window, cfg *config.Config) *SettingsDialog {
    s := &SettingsDialog{}
    
    s.autoScroll = widget.NewCheck("Auto-scroll log", func(checked bool) {
        cfg.Behavior.AutoScroll = checked
    })
    s.autoScroll.Checked = cfg.Behavior.AutoScroll
    
    s.logLevel = widget.NewSelect([]string{"Debug", "Info", "Warn", "Error"}, func(selected string) {
        cfg.Logging.Level = selected
    })
    s.logLevel.Selected = cfg.Logging.Level
    
    s.timestamp = widget.NewCheck("Show timestamps", func(checked bool) {
        cfg.Logging.ShowTimestamp = checked
    })
    s.timestamp.Checked = cfg.Logging.ShowTimestamp
    
    saveBtn := widget.NewButton("Save", func() {
        cfg.Save()
        s.dialog.Hide()
    })
    cancelBtn := widget.NewButton("Cancel", func() {
        s.dialog.Hide()
    })
    
    s.container = container.NewVBox(
        widget.NewLabel("Behavior"),
        s.autoScroll,
        widget.NewLabel(""),
        widget.NewLabel("Logging"),
        s.logLevel,
        s.timestamp,
        widget.NewLabel(""),
        container.NewHBox(saveBtn, cancelBtn),
    )
    
    s.dialog = dialog.NewCustom("Settings", "Close", s.container, parent)
    return s
}
```

**Acceptance Criteria**:
- [ ] Settings dialog opens from menu
- [ ] All preferences editable
- [ ] Save persists to config file
- [ ] Cancel discards changes

---

## 7. Summary

### 7.1 Task Count and Effort

| Phase | Tasks | Hours |
|-------|-------|-------|
| Phase 1 | 4 | 14 |
| Phase 2 | 4 | 17 |
| Phase 3 | 3 | 18 |
| Phase 4 | 3 | 18 |
| Phase 5 | 4 | 28 |
| **Total** | **18** | **95** |

### 7.2 Risk Summary

| Phase | Highest Risk | Mitigation |
|-------|--------------|-------------|
| Phase 1 | LOW | Well-understood patterns |
| Phase 2 | LOW | Extension of existing |
| Phase 3 | MEDIUM | New UI components |
| Phase 4 | MEDIUM | File I/O concerns |
| Phase 5 | MEDIUM | Multiple new features |

### 7.3 Next Steps

1. **Review and approve** this implementation plan
2. **Begin Phase 1, Task 1.1**: Implement Complete Menu Bar
3. **Establish testing criteria** for each task
4. **Track progress** via issue tracking

---

*Plan prepared: 2026-07-27*  
*Reference: UX-DESKTOP-ENGINEERING-001, DNP3-WORKBENCH-UX-ASSESSMENT-001, DNP3-WORKBENCH-UX-ROADMAP-001*
