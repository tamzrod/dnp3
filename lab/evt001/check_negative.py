#!/usr/bin/env python3
"""Check an EVT-001 negative-test PCAP.

duplicate: the OpenDNP3 confirm link frame was written twice toward TCP
port 20000, the first event was not sent again, and the later unsolicited
fragment carried only the second event at the next application sequence.

withhold: the outstation sent one unsolicited fragment and two identical
retries, no application CONFIRM reached TCP port 20000, at least one
CONFIRM is present toward the proxy, and a later response still carries
the event.
"""

import struct
import sys

LINK_HEADER = {0: 4, 1: 14, 113: 16, 276: 20}

# Milliseconds since 2000-01-01, matching the harness stamps.
TIME_LIVE = 699685567008
TIME_STAMP = 676368550000


def load(path):
    data = open(path, "rb").read()
    if len(data) < 24:
        raise SystemExit("pcap too small")
    magic = data[:4]
    if magic in (b"\xd4\xc3\xb2\xa1", b"\x4d\x3c\xb2\xa1"):
        endian = "<"
    elif magic in (b"\xa1\xb2\xc3\xd4", b"\xa1\xb2\x3c\x4d"):
        endian = ">"
    else:
        raise SystemExit("not a classic pcap")
    nano = magic in (b"\x4d\x3c\xb2\xa1", b"\xa1\xb2\x3c\x4d")
    _, _, _, _, _, _, dlt = struct.unpack(endian + "IHHIIII", data[:24])
    return data[24:], endian, dlt, nano


def iter_packets(data, endian, nano):
    off = 0
    hdr = struct.Struct(endian + "IIII")
    scale = 1e-9 if nano else 1e-6
    while off + 16 <= len(data):
        ts_sec, ts_sub, incl, _ = hdr.unpack_from(data, off)
        off += 16
        if off + incl > len(data):
            break
        yield ts_sec + ts_sub * scale, data[off : off + incl]
        off += incl


def tcp_payload(frame, dlt):
    skip = LINK_HEADER.get(dlt)
    raw = frame[skip:] if skip is not None else frame
    if len(raw) < 40 or raw[0] >> 4 != 4:
        return None
    ihl = (raw[0] & 0x0F) * 4
    if ihl < 20 or len(raw) < ihl + 20 or raw[9] != 6:
        return None
    tcp = raw[ihl:]
    data_off = ((tcp[12] >> 4) & 0x0F) * 4
    if data_off < 20 or len(tcp) < data_off:
        return None
    sport = tcp[0] << 8 | tcp[1]
    dport = tcp[2] << 8 | tcp[3]
    return sport, dport, tcp[data_off:]


def pop_frames(payload):
    frames = []
    buf = payload
    while len(buf) >= 10:
        start = buf.find(b"\x05\x64")
        if start < 0:
            break
        buf = buf[start:]
        if len(buf) < 3:
            break
        length = buf[2]
        if length < 5 or length > 255:
            buf = buf[2:]
            continue
        user = length - 5
        blocks = 0 if user == 0 else (user + 15) // 16
        total = 10 + user + 2 * blocks
        if len(buf) < total:
            break
        frames.append(buf[:total])
        buf = buf[total:]
    return frames


def user_data(frame):
    user = frame[2] - 5
    raw = frame[10:]
    out = bytearray()
    left = user
    off = 0
    while left:
        n = 16 if left > 16 else left
        if off + n + 2 > len(raw):
            return None
        out += raw[off : off + n]
        off += n + 2
        left -= n
    return bytes(out)


def describe(frame):
    user = user_data(frame)
    if user is None or len(user) < 3:
        return None
    ctrl = user[1]
    func = user[2]
    events = []
    # user layout: transport, application control, function, IIN, IIN, objects.
    if func in (0x81, 0x82) and len(user) >= 5:
        events = parse_events(user[5:])
    return {
        "ctrl": ctrl,
        "func": func,
        "seq": ctrl & 0x0F,
        "events": events,
        "raw": frame,
        "user": user,
    }


def parse_events(obj):
    events = []
    i = 0
    while i + 3 <= len(obj):
        group, var, qual = obj[i], obj[i + 1], obj[i + 2]
        if qual == 0x06:
            i += 3
            continue
        if qual == 0x17:
            if i + 4 > len(obj):
                break
            count = obj[i + 3]
            i += 4
            index_width = 1
        elif qual == 0x28:
            if i + 5 > len(obj):
                break
            count = obj[i + 3] | (obj[i + 4] << 8)
            i += 5
            index_width = 2
        else:
            break
        for _ in range(count):
            if group == 2 and var == 2:
                need = index_width + 7
            elif group == 2 and var == 1:
                need = index_width + 1
            else:
                return events
            if i + need > len(obj):
                return events
            if index_width == 1:
                index = obj[i]
            else:
                index = obj[i] | (obj[i + 1] << 8)
            flags = obj[i + index_width]
            when = None
            if group == 2 and var == 2:
                when = int.from_bytes(obj[i + index_width + 1 : i + index_width + 7], "little")
            events.append({"index": index, "flags": flags, "time": when, "group": group, "var": var})
            i += need
    return events


def collect(path):
    data, endian, dlt, nano = load(path)
    rows = []
    packets = 0
    for ts, frame in iter_packets(data, endian, nano):
        packets += 1
        parsed = tcp_payload(frame, dlt)
        if not parsed or not parsed[2]:
            continue
        sport, dport, payload = parsed
        for link in pop_frames(payload):
            info = describe(link)
            if info is None:
                continue
            info["ts"] = ts
            info["sport"] = sport
            info["dport"] = dport
            rows.append(info)
    return packets, rows


def fail(msg):
    print("PCAP_CHECK FAIL " + msg)
    return 1


def one_event(info, index, flags, when):
    if len(info["events"]) != 1:
        return False
    ev = info["events"][0]
    return ev["index"] == index and ev["flags"] == flags and ev["time"] == when


def check_duplicate(rows):
    unsol = [r for r in rows if r["func"] == 0x82 and r["sport"] == 20000]
    first = [r for r in unsol if one_event(r, 9, 0x01, TIME_LIVE)]
    second = [r for r in unsol if one_event(r, 13, 0x81, TIME_STAMP)]
    if len(first) != 1:
        return fail("index 9 unsolicited count=%d want 1" % len(first))
    if len(second) != 1:
        return fail("index 13 unsolicited count=%d want 1" % len(second))
    if second[0]["ts"] <= first[0]["ts"]:
        return fail("index 13 was sent before index 9")
    if ((first[0]["seq"] + 1) & 0x0F) != second[0]["seq"]:
        return fail("sequence advanced %d -> %d" % (first[0]["seq"], second[0]["seq"]))
    if any(ev["index"] == 9 for ev in second[0]["events"]):
        return fail("second fragment still contained index 9")
    confirms = [
        r
        for r in rows
        if r["func"] == 0x00 and r["dport"] == 20000 and r["seq"] == first[0]["seq"]
    ]
    if len(confirms) != 2:
        return fail("duplicate confirm count=%d want 2" % len(confirms))
    if confirms[0]["raw"] != confirms[1]["raw"]:
        return fail("the two confirm frames differ")
    if (confirms[0]["ctrl"] & 0xF0) != 0xD0:
        return fail("confirm control is 0x%02x, want FIR+FIN+UNS" % confirms[0]["ctrl"])
    later = [
        r
        for r in rows
        if r["func"] == 0x00 and r["dport"] == 20000 and r["seq"] == second[0]["seq"]
    ]
    if len(later) != 1:
        return fail("follow-up confirm count=%d want 1" % len(later))
    recovered = [
        r
        for r in rows
        if r["func"] == 0x81 and r["sport"] == 20000 and any(ev["index"] == 14 and ev["flags"] == 0x81 for ev in r["events"])
    ]
    if len(recovered) != 1:
        return fail("operational index 14 response count=%d" % len(recovered))
    print(
        "PCAP_CHECK duplicate PASS confirms=2 seq=%d next=%d operational=1"
        % (first[0]["seq"], second[0]["seq"])
    )
    return 0


def check_withhold(rows):
    unsol = [r for r in rows if r["func"] == 0x82 and r["sport"] == 20000]
    if len(unsol) != 3:
        return fail("unsolicited count=%d want 3" % len(unsol))
    if not all(one_event(r, 15, 0x01, TIME_LIVE) for r in unsol):
        return fail("retry payload changed")
    if unsol[0]["user"] != unsol[1]["user"] or unsol[1]["user"] != unsol[2]["user"]:
        return fail("retry application bytes differ")
    gaps = (unsol[1]["ts"] - unsol[0]["ts"], unsol[2]["ts"] - unsol[1]["ts"])
    for gap in gaps:
        if gap < 1.5 or gap > 3.0:
            return fail("retry gap %.3f outside 1.5-3.0s" % gap)
    delivered = [r for r in rows if r["func"] == 0x00 and r["dport"] == 20000]
    if delivered:
        return fail("confirm reached outstation count=%d" % len(delivered))
    generated = [r for r in rows if r["func"] == 0x00 and r["dport"] == 20010]
    if not generated:
        return fail("OpenDNP3 did not generate a confirm toward the proxy")
    if any((r["ctrl"] & 0xF0) != 0xD0 for r in generated):
        return fail("master-leg confirm was not FIR+FIN+UNS")
    recovered = [
        r
        for r in rows
        if r["func"] == 0x81
        and r["sport"] == 20000
        and any(ev["index"] == 15 and ev["flags"] == 0x01 and ev["time"] == TIME_LIVE for ev in r["events"])
    ]
    if len(recovered) != 1:
        return fail("recovery response count=%d want 1" % len(recovered))
    if recovered[0]["ts"] <= unsol[2]["ts"]:
        return fail("recovery response preceded the retry limit")
    print(
        "PCAP_CHECK withhold PASS retries=3 gaps=%.3f,%.3f master_confirms=%d delivered_confirms=0"
        % (gaps[0], gaps[1], len(generated))
    )
    return 0


def main():
    if len(sys.argv) != 3 or sys.argv[1] not in ("duplicate", "withhold"):
        raise SystemExit("usage: check_negative.py <duplicate|withhold> <file.pcap>")
    packets, rows = collect(sys.argv[2])
    print("pcap packets=%d dnp_app=%d" % (packets, len(rows)))
    if sys.argv[1] == "duplicate":
        return check_duplicate(rows)
    return check_withhold(rows)


if __name__ == "__main__":
    sys.exit(main())
