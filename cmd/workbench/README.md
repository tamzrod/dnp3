# DNP3 Engineering Workbench

A Windows desktop application for validating and debugging the native Go DNP3 library.

## Overview

The DNP3 Engineering Workbench is an engineering tool for protocol development and testing, NOT a production SCADA application. It provides:

- **Master Mode**: Connect to DNP3 outstations and exercise the library
- **Protocol Visibility**: See decoded protocol layers
- **Communication Log**: Full TX/RX message history
- **Control Operations**: Send operate commands

## MVP Features

### Included
- [x] Master Mode only
- [x] TCP connection to outstation
- [x] Read Class 0/1/2/3 commands
- [x] Display decoded response data
- [x] Protocol layer decoder
- [x] Communication log with TX/RX
- [x] Operate command (binary output)
- [x] Connection status and IIN display

### Excluded (Future)
- [ ] Outstation Mode
- [ ] TLS Transport
- [ ] Configuration files
- [ ] Multiple sessions
- [ ] Scripting

## Architecture

```
cmd/workbench/
├── main.go                    # Application entry point
├── internal/
│   ├── ui/
│   │   ├── window.go          # Main window
│   │   └── panels/           # UI panels
│   │       ├── mode.go       # Mode selection
│   │       ├── connection.go # Connection config
│   │       ├── commands.go   # Command buttons
│   │       ├── data.go       # Data display
│   │       ├── protocol.go   # Protocol decoder
│   │       ├── log.go        # Communication log
│   │       └── statusbar.go  # Status bar
│   └── session/
│       ├── session.go         # Session interface
│       └── manager.go        # Session manager
```

## Building

### Prerequisites

- Go 1.22+
- Fyne v2.4.0+

### Build Commands

```bash
# Download dependencies
go mod tidy

# Build for current OS
go build -o workbench .

# Build for Windows (on Linux/macOS with cross-compile)
GOOS=windows GOARCH=amd64 go build -o workbench.exe .
```

## Running

```bash
# Run directly
go run .

# Or run the built binary
./workbench
```

## Usage

1. **Connect**: Enter IP address and port, click Connect
2. **Read Data**: Use Read Class buttons to poll the outstation
3. **View Responses**: Check Data Panel for decoded values
4. **View Protocol**: Check Protocol Decoder for layer breakdown
5. **Operate**: Send control commands to the outstation
6. **Monitor**: Watch the Communication Log for all activity

## Development

This application is part of the DNP3 Library project and follows the KDE (Knowledge Discovery Engine) Runtime governance framework.

See [laboratory/planning/DNP3-ENG-WORKBENCH-001.md](../../laboratory/planning/DNP3-ENG-WORKBENCH-001.md) for the full engineering plan.

## License

Apache 2.0 - See project root for details.
