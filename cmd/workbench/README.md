# DNP3 Engineering Workbench

A terminal-based (TUI) application for validating and debugging the native Go DNP3 library.

## Overview

The DNP3 Engineering Workbench is an engineering tool for protocol development and testing, NOT a production SCADA application. It provides:

- **Master Mode**: Connect to DNP3 outstations and issue commands
- **Outstation Mode**: Act as a DNP3 outstation responding to master requests
- **Protocol Visibility**: See decoded protocol layers
- **Communication Log**: Full TX/RX message history

## Features

- [x] Master Mode - Connect to outstations and issue commands
- [x] Outstation Mode - Act as a DNP3 outstation server
- [x] TCP transport
- [x] Read Class 0/1/2/3 commands
- [x] Operate commands (binary output)
- [x] Display decoded response data
- [x] Protocol layer decoder
- [x] Communication log with TX/RX
- [x] Connection status and IIN display

## Building

### Prerequisites

- Go 1.22+

### Build Commands

```bash
# Build for current OS
go build -o workbench ./cmd/workbench

# Build for Windows (cross-compile)
GOOS=windows GOARCH=amd64 go build -o workbench.exe ./cmd/workbench
```

## Running

### Command-Line Flags

```
-mode master|outstation   Operating mode (default: master)
-address <ip>            Remote address (master) or listen address (outstation)
-port <number>           Port number (default: 20000)
```

### Master Mode (Terminal 1 - Outstation)

```bash
# Terminal 1: Start an outstation listening on port 20000
./workbench -mode outstation -address 0.0.0.0 -port 20000
```

### Outstation Mode (Terminal 2 - Master)

```bash
# Terminal 2: Connect as master to the outstation
./workbench -mode master -address 127.0.0.1 -port 20000
```

### Two-Terminal Test

1. **Terminal 1 (Outstation)**:
   ```bash
   ./workbench -mode outstation -address 0.0.0.0 -port 20000
   ```

2. **Terminal 2 (Master)**:
   ```bash
   ./workbench -mode master -address 127.0.0.1 -port 20000
   ```

3. In Master terminal:
   - Use Connect button or type `connect`
   - Use Read Class 0 button or type `read 0`
   - View data and logs

## Usage

1. **Start Workbench**: Run with appropriate mode and flags
2. **Connect** (Master mode): The app connects to the outstation address
3. **Read Data**: Use Read Class buttons to poll the outstation
4. **View Responses**: Check Data Panel for decoded values
5. **View Protocol**: Check Protocol Decoder for layer breakdown
6. **Operate**: Send control commands to the outstation
7. **Monitor**: Watch the Communication Log for all activity

## Architecture

```
cmd/workbench/
├── main.go                    # Application entry point, CLI flags
├── tui/                       # Terminal UI components
│   ├── app.go                # TUI application
│   └── ...                   # TUI widgets
├── internal/
│   ├── master/              # Master controller
│   ├── outstation/          # Outstation controller
│   └── logger/              # Logging
```

## Development

This application is part of the DNP3 Library project and follows the KDE (Knowledge Discovery Engine) Runtime governance framework.

## License

Apache 2.0 - See project root for details.
