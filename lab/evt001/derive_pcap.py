#!/usr/bin/env python3
"""Derive DNP3 function counts from an EVT-001 loopback PCAP.

The parser walks pcap packets, finds DNP3 link frames (start octets 0x0564),
and counts application function codes in the first user-data block. It does
not claim conformance. It records what this capture contains.
"""

import struct
import sys

DLT_NULL = 0
DLT_EN10MB = 1
DLT_LINUX_SLL = 113
DLT_LINUX_SLL2 = 276

LINK_HEADER = {
    DLT_NULL: 4,
    DLT_EN10MB: 14,
    DLT_LINUX_SLL: 16,
    DLT_LINUX_SLL2: 20,
}


def load(path):
    data = open(path, "rb").read()
    if len(data) < 24:
        raise SystemExit("pcap too small")
    magic = data[:4]
    endian = "<" if magic in (b"\xd4\xc3\xb2\xa1", b"\x4d\x3c\xb2\xa1") else ">"
    if magic not in (b"\xd4\xc3\xb2\xa1", b"\xa1\xb2\xc3\xd4", b"\x4d\x3c\xb2\xa1", b"\xa1\xb2\x3c\x4d"):
        raise SystemExit("not a classic pcap: %r" % magic)
    _, _, _, _, _, _, dlt = struct.unpack(endian + "IHHIIII", data[:24])
    return data[24:], endian, dlt


def iter_packets(data, endian):
    off = 0
    hdr = struct.Struct(endian + "IIII")
    while off + 16 <= len(data):
        ts_sec, ts_sub, incl, orig = hdr.unpack_from(data, off)
        off += 16
        if off + incl > len(data):
            break
        yield data[off : off + incl]
        off += incl


def tcp_payload(frame, dlt):
    skip = LINK_HEADER.get(dlt)
    blob = frame[skip:] if skip is not None else frame
    # Also search the raw frame so an unexpected DLT still yields DNP3.
    hits = []
    for raw in (blob, frame):
        if len(raw) < 20:
            continue
        if raw[0] >> 4 == 4:
            ihl = (raw[0] & 0x0F) * 4
            if ihl < 20 or len(raw) < ihl + 20 or raw[9] != 6:
                continue
            tcp = raw[ihl:]
            data_off = ((tcp[12] >> 4) & 0x0F) * 4
            if data_off < 20 or len(tcp) < data_off:
                continue
            hits.append(tcp[data_off:])
    return hits


def app_functions(payload):
    found = []
    i = 0
    while i + 10 <= len(payload):
        if payload[i] != 0x05 or payload[i + 1] != 0x64:
            i += 1
            continue
        length = payload[i + 2]
        if length < 5 or i + 10 > len(payload):
            i += 1
            continue
        # User data begins after the 10-octet link header (includes CRC).
        # Transport header is the first user octet. Application control and
        # function follow it when the link length includes them.
        user = length - 5
        if user >= 3 and i + 10 + 3 <= len(payload):
            app_ctrl = payload[i + 11]
            func = payload[i + 12]
            found.append((app_ctrl, func))
        # Advance by the first block only. CRC spacing makes a full walk
        # unnecessary for counting function codes that sit in the first block.
        i += 10 + min(user, 16) + 2
    return found


def main():
    if len(sys.argv) != 2:
        raise SystemExit("usage: derive_pcap.py <file.pcap>")
    data, endian, dlt = load(sys.argv[1])
    packets = 0
    frames = 0
    counts = {}
    for frame in iter_packets(data, endian):
        packets += 1
        for payload in tcp_payload(frame, dlt):
            for ctrl, func in app_functions(payload):
                frames += 1
                counts[func] = counts.get(func, 0) + 1
    names = {
        0x00: "CONFIRM",
        0x01: "READ",
        0x02: "WRITE",
        0x14: "ENABLE_UNSOLICITED",
        0x15: "DISABLE_UNSOLICITED",
        0x81: "RESPONSE",
        0x82: "UNSOLICITED",
    }
    print("pcap_dlt=%d packets=%d dnp_app=%d" % (dlt, packets, frames))
    for func in sorted(counts):
        print("func 0x%02x %s count=%d" % (func, names.get(func, "OTHER"), counts[func]))
    if frames == 0:
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
