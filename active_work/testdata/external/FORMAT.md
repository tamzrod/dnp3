# VEC-01 capture fixture format (`.vec`)

A capture fixture is a line-oriented UTF-8 text file with the `.vec` extension.
It pairs a metadata header with an ordered list of frames observed on the wire
during one logical exchange. An optional PCAP sidecar (`.pcap`/`.pcapng`, same
basename) may accompany a `.vec` for raw network capture.

The format is deliberately simple and reviewable in a text editor / diff.

## 1. Metadata header

The file begins with metadata as `key: value` lines. Required keys:

| Key           | Meaning                                                              |
|---------------|----------------------------------------------------------------------|
| `id`          | Stable fixture id (e.g. `VEC-01-class0-g1-g20-g30`).                 |
| `title`       | Human-readable summary.                                              |
| `source`      | Where the capture came from (product + version, or "third-party").   |
| `capture`     | Capture conditions (TCP loopback / serial / link; date or "synthetic-from-spec"). |
| `role`        | Which side this repo's master played: `master` or `outstation`.      |
| `master-addr` | DNP3 source address used by the master.                              |
| `outstation-addr` | DNP3 destination address of the peer.                            |
| `pcap`        | Optional sidecar filename (relative to this file). Omit if none.     |
| `notes`       | Free-form notes (optional).                                          |

A blank line ends the metadata header.

## 2. Frame list

After the blank line, frames are listed in wire order. Each frame is a record:

```
@ <direction> <layer>
<hex bytes — whitespace-separated, one logical frame per line or per block>
```

- `direction` is one of:
  - `M->O` — master to outstation
  - `O->M` — outstation to master
- `layer` is one of:
  - `link`   — a full DNP3 data-link frame (starts with `05 64`, includes CRCs)
  - `app`    — reassembled application-layer APDU bytes (no DLL/TL framing)
  - `raw`    — raw TCP byte stream segment (use only with a PCAP sidecar)

The hex block for a frame may span multiple lines; a frame ends at the next
`@` record or at a line beginning with `#`. `#`-prefixed lines are comments
and are ignored by any loader.

## 3. Example (excerpt)

```
id: VEC-01-class0-g1-g20-g30
title: Class-0 integrity poll, G1/G20/G30, master-initiated
source: third-party outstation (placeholder)
capture: synthetic-from-spec
role: master
master-addr: 3
outstation-addr: 4
notes: placeholder fixture; replace with a real capture in MEXT-033.

@ M->O link
05 64 05 C0 04 00 03 00 F9 56

@ O->M link
05 64 05 00 03 00 04 00 3C 56

@ M->O link
05 64 05 C9 04 00 03 00 B6 20

@ O->M link
05 64 05 02 03 00 04 00 30 10
```

## 4. Loader (optional)

No loader is required by MEXT-020. When a loader is added (MEXT-022+), it MUST:
- skip blank lines and `#` comments,
- parse `key: value` metadata until the first blank line,
- parse `@ <direction> <layer>` records and collect hex bytes until the next
  record/comment/EOF,
- reject unknown directions/layers and non-hex bytes (fail closed).

## 5. Naming

`<id>.vec` where `<id>` is the `id` metadata value. A PCAP sidecar, if any, is
`<id>.pcap` (or `.pcapng`).
