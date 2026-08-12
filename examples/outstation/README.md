# Outstation example

Minimal DNP3 outstation server using the v0 MVP public API (DNP3-095).

- Listens on `0.0.0.0:20000` (TCP), outstation link address `1024`.
- Single-master profile (`WithMaxConnections(1)`, DNP3-084).
- Serves MVP static points: G1V1 binary input, G20V1 counter, G30V1 analog input.
- Accepts Group 12 Variation 1 direct binary control; rejects analog control
  (outside the v0 MVP profile).

## Build

```bash
go build ./examples/outstation
go vet ./examples/outstation
```

## Run

```bash
go run ./examples/outstation
```

Connect a DNP3 master to outstation address `1024` on TCP port `20000`.
Press Ctrl+C to stop.
