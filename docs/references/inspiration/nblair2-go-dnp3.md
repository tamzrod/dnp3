# Reference / Inspiration: `nblair2/go-dnp3`

**Source:** https://github.com/nblair2/go-dnp3  
**Compared against:** https://github.com/tamzrod/dnp3  
**Purpose:** Preserve useful implementation ideas for later investigation. This is a reference/inspiration record, not an endorsement to copy architecture or code.

## Scope of the reference

`nblair2/go-dnp3` describes itself as a Go library for parsing and encoding DNP3 frames, with `gopacket` integration, TCP/UDP port 20000 registration, stream parsing via `ParseFrames`, round-trip serialization, JSON/string inspection, and PCAP-based tests sourced from opendnp3 conformance reports. citehttps://github.com/nblair2/go-dnp3/blob/main/README.md

Its architecture is oriented around a packet/codec representation (`Frame` containing DataLink, Transport, and Application) rather than a full Master/Outstation runtime. citehttps://github.com/nblair2/go-dnp3/blob/main/dnp3/dnp3.go

## High-value ideas to revisit later

### 1. Object header generalization

`ObjectHeader` separates Group/Variation from qualifier details using `PointPrefixCode`, `RangeSpecCode`, and a `RangeField` abstraction. Range-field construction is selected by qualifier and the object header tracks encoded size. citehttps://github.com/nblair2/go-dnp3/blob/main/dnp3/objectHeader.go

**Why useful later:** Tamzrod/dnp3 currently uses a deliberately narrow, explicit v0 qualifier allow-list. The nblair2 representation is a useful design reference if the supported profile is expanded beyond the current qualifiers.

### 2. Group/Variation -> constructor/packer registry

`objectType` maps `(Group, Variation)` to a description plus a point constructor and packer. The repository uses this registry to dispatch supported object encodings instead of spreading group/variation special cases throughout the parser. citehttps://github.com/nblair2/go-dnp3/blob/main/dnp3/objectType.go

**Why useful later:** This is a concrete pattern for growing the object model while keeping point decoding/encoding dispatch explicit.

### 3. Point interface + capability-aware fields

The `Point` interface exposes common operations such as index, flags, timestamps, value access, and serialization. `PointFields` describes which optional fields are present for a point type. `PointBit` demonstrates how packed bits and flag-bearing representations can share a point abstraction while keeping representation-specific behavior explicit. citehttps://github.com/nblair2/go-dnp3/blob/main/dnp3/point.go citehttps://github.com/nblair2/go-dnp3/blob/main/dnp3/pointBit.go

**Why useful later:** Could inform a future generalized point model when Tamzrod/dnp3 expands beyond its verified MVP profile.

### 4. Stream frame extraction helper

`ParseFrames` walks a TCP byte buffer, emits every complete frame it can parse, and preserves a trailing partial frame for the caller. citehttps://github.com/nblair2/go-dnp3/blob/main/dnp3/dnp3.go

**Why useful later:** This is a small, concrete utility pattern for stream consumers, even though Tamzrod/dnp3's transport/session architecture is more stateful.

### 5. PCAP replay / conformance-oriented fixtures

The test suite contains raw DNP3 vectors and supports feeding custom PCAPs into packet decoding and round-trip tests. The README states that test data is sourced from opendnp3 conformance reports. citehttps://github.com/nblair2/go-dnp3/blob/main/dnp3/dnp3_test.go citehttps://github.com/nblair2/go-dnp3/blob/main/README.md

**Why useful later:** A useful complement to Tamzrod/dnp3's existing golden vectors, loopback tests, and real-TCP tests.

### 6. Optional gopacket adapter

The repository makes its frame implement gopacket interfaces and registers DNP3 on TCP/UDP port 20000, making packet capture tooling integration straightforward. citehttps://github.com/nblair2/go-dnp3/blob/main/dnp3/dnp3.go

**Why useful later:** Potential packet inspection/dissection integration without making gopacket the core protocol architecture.

## Ideas not adopted by reference

The following are intentionally recorded as **not architectural targets** for Tamzrod/dnp3 based on the audit:

- A single gopacket-centric `Frame` as the core abstraction.
- The thin transport representation that exposes transport-header information but does not own full message reassembly/session behavior.
- Treating application data primarily as a decoded object tree without the Master/Outstation/session semantics being the center of the stack.

These observations are scope/context notes, not claims that the other implementation is incorrect.

## Audit snapshot: nblair2 vs Tamzrod

### Data Link

Both implementations cover DNP3 framing, CRC, addressing and control fields. Tamzrod/dnp3 has stronger explicit framing boundaries and size discipline in `internal/dll/frame`, including explicit `MaxDataSize=250`, `MaxFrameSize=292`, CRC validation boundaries, and explicit oversized-frame rejection. citehttps://github.com/tamzrod/dnp3/blob/main/internal/dll/frame/frame.go

The nblair2 implementation is compact and useful as a codec reference. Its `DataLink` parses/serializes the 10-byte header and validates primary function-code/FCV combinations. citehttps://github.com/nblair2/go-dnp3/blob/main/dnp3/dataLink.go

### Transport

Tamzrod/dnp3 has a real `Fragmenter`/`Reassembler` with FIR/FIN, modulo-64 sequence tracking, missing-first-fragment detection, mismatch/duplicate handling, and buffer overflow protection. citehttps://github.com/tamzrod/dnp3/blob/main/internal/tl/transport.go

nblair2's `Transport` is deliberately thinner: it represents FIR/FIN/sequence and CRCs and removes/interleaves CRCs, but does not provide the same message-level reassembly state machine. citehttps://github.com/nblair2/go-dnp3/blob/main/dnp3/transport.go

### Application model

Tamzrod/dnp3 currently prefers an explicit supported-profile approach: the object-header implementation accepts only verified qualifiers (`0x06`, `0x07`, `0x00`, `0x28`, `0x27`) and rejects unsupported qualifiers instead of silently generalizing. citehttps://github.com/tamzrod/dnp3/blob/main/internal/al/object_header.go

nblair2 has a broader generic object-header model and a much wider Group/Variation registry with point constructors/packers. citehttps://github.com/nblair2/go-dnp3/blob/main/dnp3/objectHeader.go citehttps://github.com/nblair2/go-dnp3/blob/main/dnp3/objectType.go

### Testing / verification philosophy

nblair2 strongly emphasizes codec round trips and PCAP/replay inspection: decode bytes, require no unsupported `Extra` payload, serialize back, and compare wire bytes. citehttps://github.com/nblair2/go-dnp3/blob/main/dnp3/dnp3_test.go

Tamzrod/dnp3 goes beyond codec round trips into Master/Outstation behavior: a real-TCP integration test exercises Connect, IntegrityPoll, Operate, command dispatch, and Disconnect against the in-repo outstation. The repository explicitly documents that third-party interoperability is still open and that the project is not claiming full IEEE 1815 conformance. citehttps://github.com/tamzrod/dnp3/blob/main/test/integration/real_tcp_full_mvp_test.go citehttps://github.com/tamzrod/dnp3/blob/main/README.md

## Bottom line

Treat `nblair2/go-dnp3` as a **reference implementation for packet/object modeling and analysis tooling**, not as the architectural model for Tamzrod/dnp3.

The most useful future references are:

1. ObjectHeader + RangeField abstraction.
2. Group/Variation -> constructor/packer registry.
3. Generalized Point interface and optional-field modeling.
4. TCP `ParseFrames` stream extraction.
5. PCAP replay/conformance fixture strategy.
6. Optional gopacket adapter.

When expanding Tamzrod/dnp3, these should be revisited as **inspiration/reference**, with independent implementation and verification against the existing Tamzrod architecture and supported-profile contracts.
