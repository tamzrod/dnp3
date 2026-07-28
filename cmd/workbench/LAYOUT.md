# DNP3 Engineering Workbench - Layout Preview

## Main Window Layout

```
┌──────────────────────────────────────────────────────────────────────────────┐
│  DNP3 Engineering Workbench                                        [_][□][X] │
├──────────────────────────────────────────────────────────────────────────────┤
│  File  Edit  View  Settings  Session  Help                                    │
├────────────────────┬─────────────────────────────────────────────────────────┤
│                    │                                                         │
│  ┌──────────────┐ │  ┌─────────────────────────────────────────────────┐   │
│  │WORKBENCH MODE│ │  │              DATA TABLE PANEL                    │   │
│  │              │ │  │  ┌───────┬───────┬────────┬────────┬─────────┐│   │
│  │(•) Poll      │ │  │  │ Index │ Type  │ Value  │ Quality│ Time    ││   │
│  │( ) Simulate  │ │  │  ├───────┼───────┼────────┼────────┼─────────┤│   │
│  │              │ │  │  │ 0     │ BI    │ true   │ ONLINE │ 10:23:01││   │
│  │ Poll: Connect │ │  │  │ 1     │ BI    │ false  │ ONLINE │ 10:23:01││   │
│  │ to remote    │ │  │  │ 0     │ AI    │ 100.5  │ ONLINE │ 10:23:01││   │
│  │ outstation   │ │  │  │ 1     │ AI    │ 50.25  │ ONLINE │ 10:23:01││   │
│  │              │ │  │  │ 0     │ CTR   │ 1000   │ ONLINE │ 10:23:01││   │
│  │ Simulate: Act│ │  │  │ ...   │ ...   │ ...    │ ...    │ ...     ││   │
│  │ as DNP3     │ │  │  └───────┴───────┴────────┴────────┴─────────┘│   │
│  │ server       │ │  │  [Read All] [Export CSV] [Refresh]          │   │
│  └──────────────┘ │  └─────────────────────────────────────────────────┘   │
│                    │                                                         │
│  ┌──────────────┐ │  ┌─────────────────────────────────────────────────┐   │
│  │CONNECTION    │ │  │              CONTROL PANEL                      │   │
│  │              │ │  │  Selected: Binary Input #0                     │   │
│  │ Address:     │ │  │  Current Value: true                          │   │
│  │ [127.0.0.1] │ │  │  Quality: ONLINE                             │   │
│  │ Port:        │ │  │                                                 │   │
│  │ [20000     ] │ │  │  ┌─────────────────┐  ┌─────────────────┐   │   │
│  │              │ │  │  │  [  ON  ] [OFF] │  │  [Set Value]    │   │   │
│  │ [ Connect ]  │ │  │  └─────────────────┘  └─────────────────┘   │   │
│  │ [Disconnect] │ │  │                                                 │   │
│  └──────────────┘ │  └─────────────────────────────────────────────────┘   │
│                    │                                                         │
│  ┌──────────────┐ │                                                         │
│  │COMMANDS      │ │                                                         │
│  │              │ │                                                         │
│  │ [Read Class 0]│                                                         │
│  │ [Read Class 1]│                                                         │
│  │ [Read Class 2]│                                                         │
│  │ [Read Class 3]│                                                         │
│  │ [Enable Unsol]│                                                         │
│  └──────────────┘ │                                                         │
│                    │                                                         │
│  25%              │  75%                                                  │
├────────────────────┴─────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │ LOG PANEL                                                           │    │
│  │ [10:23:01.123] ← RECV: 05 64 0C 01 00 00 00 ...                  │    │
│  │ [10:23:01.456] → SEND: 00 03 00 00 00 06 C0 ...                   │    │
│  │ [10:23:02.001] ← RECV: 05 64 0C 01 00 00 00 ...                  │    │
│  │ [10:23:02.234] → SEND: 00 03 00 00 00 06 C0 ...                   │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
├──────────────────────────────────────────────────────────────────────────────┤
│ ● Connected │ State: Connected │ Connection: 127.0.0.1:20000 │ IIN: 0x0000 │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Panel Descriptions

### Left Sidebar (25% width)

| Panel | Description |
|-------|-------------|
| **Mode Panel** | Radio buttons to switch between Poll Outstation (Master) and Simulate Outstation modes |
| **Connection Panel** | IP address and port input fields with Connect/Disconnect buttons |
| **Commands Panel** | Buttons for Read Class 0/1/2/3 and Enable Unsolicited |

### Main Content Area (75% width)

| Panel | Description |
|-------|-------------|
| **Data Table Panel** | Sortable table showing all data points with Index, Type, Value, Quality, Timestamp |
| **Control Panel** | Shows selected point details and allows value changes (for binary/analog outputs) |

### Bottom Panel

| Panel | Description |
|-------|-------------|
| **Log Panel** | Scrollable log showing timestamped DNP3 protocol traffic (← RECV, → SEND) |

### Status Bar

| Element | Description |
|---------|-------------|
| Connection Indicator | ● Green = Connected, ○ Gray = Disconnected |
| State | Current connection state |
| Connection | Remote address:port |
| IIN | DNP3 Internal Indications (hex) |

## Layout Features

- **Resizable Split Pane**: Left sidebar can be resized (25% default)
- **Collapsible Panels**: Sidebar and Log Panel can be toggled via View menu or status bar buttons
- **Fullscreen Mode**: Toggle via View menu or keyboard shortcut

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Ctrl+N | New Session |
| Ctrl+O | Open Configuration |
| Ctrl+S | Save Configuration |
| Ctrl+F | Find in Log |
| Ctrl+L | Clear Log |
| F11 | Toggle Fullscreen |

## Menu Structure

```
File
├── New Session
├── Open Configuration...
├── Save Configuration
├── Save Configuration As...
├── Export Log...
├── Print...
└── Exit

Edit
├── Undo
├── Redo
├── Cut
├── Copy
├── Paste
├── Delete
├── Find in Log
└── Select All

View
├── Zoom In
├── Zoom Out
├── Reset Zoom
├── Sidebar
├── Log Panel
└── Fullscreen

Settings
└── Preferences...

Session
├── Connect
├── Disconnect
├── Read Class 0
├── Read Class 1
├── Read Class 2
├── Read Class 3
└── Clear Log

Help
├── Documentation
├── Keyboard Shortcuts
└── About DNP3 Workbench
```
