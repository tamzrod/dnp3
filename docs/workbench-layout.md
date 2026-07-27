# DNP3 Engineering Workbench - UI Layout

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│ File  Edit  View  Session  Settings  Help                              [─] [□] [✕]    │
├─────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                           │
│  ┌─────────────────┐  ┌───────────────────────────────────────────────────────────────┐  │
│  │ WORKBENCH       │  │ DATA MONITORING                                    [Clear][↻]   │  │
│  │                 │  ├───────────────────────────────────────────────────────────────┤  │
│  │ MODE            │  │ Index │ Type │ Value     │ Quality │ Time                      │  │
│  │ ○ Poll Out      │  │───────┼──────┼───────────┼─────────┼──────────────────────────│  │
│  │ ○ Simulate      │  │   0   │  DI  │ ON        │ ONLINE  │ 2026-07-27 11:29:51    │  │
│  │                 │  │   1   │  AI  │ 42.50     │ ONLINE  │ 2026-07-27 11:29:51    │  │
│  ├─────────────────┤  │   2   │  DI  │ OFF       │ ONLINE  │ 2026-07-27 11:29:51    │  │
│  │ CONNECTION      │  │   3   │  CTR │ 1234      │ ONLINE  │ 2026-07-27 11:29:51    │  │
│  │ Address:        │  │   ... │ ...  │ ...       │ ...     │ ...                      │  │
│  │ [localhost    ] │  │       │      │           │         │                          │  │
│  │ Port:           │  │       │      │           │         │                          │  │
│  │ [20000       ] │  └───────────────────────────────────────────────────────────────┘  │
│  │ [Connect]       │  ┌───────────────────────────────────────────────────────────────┐  │
│  ├─────────────────┤  │ CONTROL PANEL                                    □ SBO         │  │
│  │ COMMAND         │  │ Selected Points: DO-0, DO-1                                   │  │
│  │                 │  │ ┌─────────────────────┐  ┌───────────────────────────────────┐ │  │
│  │ [Class 0]       │  │ │ DO-0               │  │ Binary Output:                    │ │  │
│  │ [Class 1]       │  │ │ AO-5               │  │   [ON]  [OFF]                    │ │  │
│  │ [Class 2]       │  │ └─────────────────────┘  │                                   │ │  │
│  │ [Class 3]       │  │                           │ Analog Output:                    │ │  │
│  │                 │  │                           │ [Enter value...____] [Set]       │ │  │
│  └─────────────────┘  └───────────────────────────────────────────────────────────────┘  │
│  │ 250px fixed     │  │  Flexible (HSplit offset 0.75)                              │  │
│  └─────────────────┘  └───────────────────────────────────────────────────────────────┘  │
│                         │                                                           │  │
│                         ├───────────────────────────────────────────────────────────┤  │
│                         │ PROTOCOL LOG                                              │  │
│                         │ 11:29:51.234 [TX] 05 64 0C 01 00 ...                    │  │
│                         │ 11:29:51.256 [RX] 05 64 0C 01 00 ...                    │  │
│                         │ 11:29:51.890 [TX] 05 64 0C 02 00 ...                    │  │
│                         └───────────────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────────────────────────────────────────┤
│ Disconnected  │  localhost:20000  │  0x0000                      │  │ ⚙ Polling: --  │
└─────────────────────────────────────────────────────────────────────────────────────────────┘

Legend:
├── Native title bar with window controls (minimize, maximize, close)
├── Native menu bar (File, Edit, View, Session, Settings, Help)
├── Left sidebar (fixed 250px) with collapsible panels
├── Right content area with HSplit (25%/75% default split)
├── DATA MONITORING table with scrollable rows
├── CONTROL PANEL for operating outputs
├── PROTOCOL LOG with hex dump
└── Status bar with connection state and IIN
```

## Panel Details

### Left Sidebar (250px fixed)
| Panel | Contents |
|-------|----------|
| **WORKBENCH** | App title |
| **MODE** | Radio: Poll Outstation / Simulate Outstation |
| **CONNECTION** | Address input, Port input, Connect/Disconnect button |
| **COMMAND** | Read Class 0-3 buttons |

### Right Content Area
| Panel | Contents |
|-------|----------|
| **DATA MONITORING** | Table with columns: Index, Type, Value, Quality, Time |
| **CONTROL PANEL** | Selected points list, Binary ON/OFF, Analog input + Set |
| **PROTOCOL LOG** | Timestamped hex dump of TX/RX messages |

### Status Bar
- Connection state (Disconnected/Connected/Error)
- Address:Port
- IIN hex value (Internal Indications)
- Polling status

## Responsive Behavior
- Minimum window size: 1024x720
- Sidebar: 250px fixed
- Split pane: Draggable divider, default 25%/75%
- Table: Scrollable with sticky header row
