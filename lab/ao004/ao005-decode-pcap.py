#!/usr/bin/env python3
"""AO-005 independent raw-PCAP decoder (lab-only evidence tool).

Reads a classic pcap captured on the master-side veth, reassembles the TCP
stream, extracts and verifies DNP3 link frames (header/block CRC), decodes the
transport segment(s) and the application-layer DirectOperate request / RESPONSE,
and prints a deterministic decoded record with literal on-wire hex.

It deliberately re-implements the wire parsing from IEEE 1815 first principles
(no project code) so it is an independent check of the Go/C++ endpoints.
"""
import struct
import sys


def crc_dnp(data: bytes) -> int:
    """CRC-16/DNP: poly 0x3D65, init 0, reflected, xorout 0xFFFF."""
    crc = 0
    for b in data:
        crc ^= b
        for _ in range(8):
            if crc & 1:
                crc = (crc >> 1) ^ 0xA6BC  # reflected 0x3D65
            else:
                crc >>= 1
    return crc ^ 0xFFFF


def read_pcap(path):
    with open(path, "rb") as f:
        data = f.read()
    magic = struct.unpack("<I", data[:4])[0]
    if magic == 0xA1B2C3D4:
        endian, nano = "<", False
    elif magic == 0xD4C3B2A1:
        endian, nano = ">", False
    elif magic == 0xA1B23C4D:
        endian, nano = "<", True
    elif magic == 0x4D3CB2A1:
        endian, nano = ">", True
    else:
        raise SystemExit(f"not a pcap (magic {magic:#x})")
    off = 24
    pkts = []
    while off + 16 <= len(data):
        ts_s, ts_u, incl, orig = struct.unpack(endian + "IIII", data[off:off + 16])
        off += 16
        pkts.append(data[off:off + incl])
        off += incl
    return pkts, nano


def parse_eth_ip_tcp(frame):
    if len(frame) < 14:
        return None
    eth_type = struct.unpack(">H", frame[12:14])[0]
    if eth_type != 0x0800:  # IPv4 only (lab veth)
        return None
    ip = frame[14:]
    if len(ip) < 20:
        return None
    ihl = (ip[0] & 0x0F) * 4
    proto = ip[9]
    if proto != 6:
        return None
    src = ".".join(str(b) for b in ip[12:16])
    dst = ".".join(str(b) for b in ip[16:20])
    tcp = ip[ihl:]
    sport, dport = struct.unpack(">HH", tcp[:4])
    doff = (tcp[12] >> 4) * 4
    payload = tcp[doff:]
    return src, dst, sport, dport, payload


def reassemble(pkts):
    """Return ordered list of (src,sport,dst,dport, bytes) streams, plus counts."""
    streams = {}
    order = []
    for p in pkts:
        r = parse_eth_ip_tcp(p)
        if not r:
            continue
        src, dst, sport, dport, payload = r
        key = (src, sport, dst, dport)
        if key not in streams:
            streams[key] = bytearray()
            order.append(key)
        streams[key] += payload
    return [(k, bytes(streams[k])) for k in order]


def extract_link_frames(stream):
    frames = []
    i = 0
    n = len(stream)
    while i + 3 <= n:
        if stream[i] != 0x05 or stream[i + 1] != 0x64:
            i += 1
            continue
        length = stream[i + 2]
        user_len = length - 5
        if user_len < 0:
            i += 1
            continue
        nblk = (user_len + 15) // 16
        total = 2 + 1 + 5 + 2 + user_len + 2 * nblk
        if i + total > n:
            break
        frames.append(stream[i:i + total])
        i += total
    return frames


def crc_ok(chunk: bytes) -> bool:
    crc = struct.unpack("<H", chunk[-2:])[0]
    return crc == crc_dnp(chunk[:-2])


def parse_link(frame):
    length = frame[2]
    ctrl = frame[3]
    dest = struct.unpack("<H", frame[4:6])[0]
    src = struct.unpack("<H", frame[6:8])[0]
    header_crc_ok = crc_ok(frame[0:10])
    user_len = length - 5
    # user data blocks
    data = bytearray()
    crcs_ok = True
    p = 10
    remaining = user_len
    while remaining > 0:
        blk = min(16, remaining)
        data += frame[p:p + blk]
        if not crc_ok(frame[p:p + blk + 2]):
            crcs_ok = False
        p += blk + 2
        remaining -= blk
    return {
        "len": length, "ctrl": ctrl, "dest": dest, "src": src,
        "dir": "master->outstation" if ctrl & 0x80 else "outstation->master",
        "prm": bool(ctrl & 0x80), "func": ctrl & 0x0F,
        "header_crc_ok": header_crc_ok, "block_crc_ok": crcs_ok,
        "user": bytes(data), "raw": frame,
    }


def decode_transport(user):
    """Split user data into transport segments; return list of (tc, app)."""
    segs = []
    i = 0
    while i < len(user):
        tc = user[i]
        segs.append((tc, user[i + 1:]))
        break  # lab frames are single-segment (bounded, < 250 bytes)
    return segs


def decode_app(app):
    out = {}
    if not app:
        return out
    ac = app[0]
    out["app_ctrl"] = ac
    out["app_fir"] = bool(ac & 0x80)
    out["app_fin"] = bool(ac & 0x40)
    out["app_con"] = bool(ac & 0x20)
    out["app_uns"] = bool(ac & 0x10)
    out["app_seq"] = ac & 0x0F
    fn = app[1]
    out["func"] = fn
    names = {0x01: "READ", 0x02: "WRITE", 0x05: "DIRECT_OPERATE",
             0x06: "DIRECT_OPERATE_NR", 0x81: "RESPONSE"}
    out["func_name"] = names.get(fn, f"0x{fn:02x}")
    body = app[2:]
    if fn in (0x05, 0x06):  # request: object header + objects
        objs = []
        p = 0
        while p + 3 <= len(body):
            grp, var, qual = body[p], body[p + 1], body[p + 2]
            p += 3
            o = {"group": grp, "variation": var, "qualifier": qual}
            objs.append(o)
            # qualifier 0x28 = 16-bit count + 16-bit index prefix (the contract
            # this implementation emits for control objects).
            if qual == 0x28:
                cnt = struct.unpack("<H", body[p:p + 2])[0]; p += 2
                o["count"] = cnt
                pts = []
                for _ in range(cnt):
                    idx = struct.unpack("<H", body[p:p + 2])[0]; p += 2
                    val, p = _g41_value(var, body, p)
                    pts.append({"index": idx, "value": val})
                o["points"] = pts
            elif qual == 0x17:  # 8-bit count + 8-bit index prefix
                cnt = body[p]; p += 1
                o["count"] = cnt
                pts = []
                for _ in range(cnt):
                    idx = body[p]; p += 1
                    val, p = _g41_value(var, body, p)
                    pts.append({"index": idx, "value": val})
                o["points"] = pts
            else:
                o["raw_tail"] = body[p:].hex()
                p = len(body)
        out["objects"] = objs
    elif fn == 0x81:  # response: IIN(2) + optional object header
        out["iin"] = struct.unpack("<H", body[0:2])[0]
        out["iin_hex"] = f"0x{out['iin']:04x}"
        rest = body[2:]
        if len(rest) >= 3:
            grp, var, qual = rest[0], rest[1], rest[2]
            o = {"group": grp, "variation": var, "qualifier": qual}
            p = 3
            if qual == 0x28:
                cnt = struct.unpack("<H", rest[p:p + 2])[0]; p += 2
                o["count"] = cnt
                pts = []
                for _ in range(cnt):
                    idx = struct.unpack("<H", rest[p:p + 2])[0]; p += 2
                    # Control-object response echoes the request value followed
                    # by the command-status octet (last byte of the point).
                    if var in (1, 2, 3, 4):
                        val, p = _g41_value(var, rest, p)
                    else:
                        val = None
                    status = rest[p]; p += 1
                    pts.append({"index": idx, "value": val, "status": status,
                                "status_hex": f"0x{status:02x}"})
                o["points"] = pts
            elif qual == 0x17:
                cnt = rest[p]; p += 1
                o["count"] = cnt
                pts = []
                for _ in range(cnt):
                    idx = rest[p]; p += 1
                    status = rest[p]; p += 1
                    pts.append({"index": idx, "status": status,
                                "status_hex": f"0x{status:02x}"})
                o["points"] = pts
            elif qual == 0x07:  # 8-bit count, no index prefix
                cnt = rest[p]; p += 1
                o["count"] = cnt
                pts = []
                for _ in range(cnt):
                    status = rest[p]; p += 1
                    pts.append({"status": status, "status_hex": f"0x{status:02x}"})
                o["points"] = pts
            out["objects"] = [o]
    return out


def _g41_value(var, body, p):
    if var == 1:  # int32
        v = struct.unpack("<i", body[p:p + 4])[0]
        return f"int32:{v}", p + 4
    if var == 2:  # int16
        v = struct.unpack("<h", body[p:p + 2])[0]
        return f"int16:{v}", p + 2
    if var == 3:  # float32
        bits = struct.unpack("<I", body[p:p + 4])[0]
        v = struct.unpack("<f", body[p:p + 4])[0]
        return f"float32:{v!r} bits=0x{bits:08x}", p + 4
    if var == 4:  # float64
        bits = struct.unpack("<Q", body[p:p + 8])[0]
        v = struct.unpack("<d", body[p:p + 8])[0]
        return f"float64:{v!r} bits=0x{bits:016x}", p + 8
    return f"<var{var} unparsed>", p


def emit_cases(pkts):
    """Emit one deterministic line per G41 DirectOperate case with literal
    on-wire request/response hex (a reproducible .vec-equivalent record).
    Requests and responses travel in opposite-direction streams, so they are
    collected in encounter order and paired."""
    streams = reassemble(pkts)
    reqs = []
    resps = []
    for (src, sport, dst, dport), stream in streams:
        frames = extract_link_frames(stream)
        for fr in frames:
            lk = parse_link(fr)
            for _tc, app in decode_transport(lk["user"]):
                a = decode_app(app)
                if a.get("func_name") == "DIRECT_OPERATE":
                    for o in a.get("objects", []):
                        if o.get("group") == 41:
                            pt = o["points"][0]
                            reqs.append((o["variation"], pt["index"], pt["value"], fr.hex()))
                elif a.get("func_name") == "RESPONSE":
                    for o in a.get("objects", []):
                        if o.get("group") == 41:
                            pt = o["points"][0]
                            resps.append((pt.get("status_hex"), a.get("iin_hex"),
                                          pt.get("value"), fr.hex()))
    for i in range(max(len(reqs), len(resps))):
        req = reqs[i] if i < len(reqs) else None
        resp = resps[i] if i < len(resps) else None
        if req:
            v, idx, val, hexr = req
        else:
            v = idx = val = hexr = "?"
        if resp:
            st, iin, rval, hexp = resp
        else:
            st = iin = rval = hexp = "?"
        print(f"CASE direction={DIR_LABEL} variation={v} index={idx} "
              f"req_value={val} req_hex={hexr} "
              f"resp_status={st} resp_value={rval} iin={iin} resp_hex={hexp}")


DIR_LABEL = "?"


def main():
    global DIR_LABEL
    path = sys.argv[1]
    if len(sys.argv) > 2 and sys.argv[2].startswith("dir="):
        DIR_LABEL = sys.argv[2].split("=", 1)[1]
    pkts, nano = read_pcap(path)
    if len(sys.argv) > 3 and sys.argv[3] == "cases":
        emit_cases(pkts)
        return
    print(f"PCAP {path}: {len(pkts)} packets, {'nano' if nano else 'micro'}second timestamps")
    streams = reassemble(pkts)
    total_frames = 0
    for (src, sport, dst, dport), stream in streams:
        print(f"\nTCP STREAM {src}:{sport} -> {dst}:{dport}  ({len(stream)} bytes)")
        frames = extract_link_frames(stream)
        for fi, fr in enumerate(frames):
            lk = parse_link(fr)
            total_frames += 1
            print(f"  LINK[{fi}] {lk['dir']} src={lk['src']} dest={lk['dest']} "
                  f"len={lk['len']} func=0x{lk['func']:02x} "
                  f"hdrCRC={'ok' if lk['header_crc_ok'] else 'BAD'} "
                  f"blkCRC={'ok' if lk['block_crc_ok'] else 'BAD'}")
            for tc, app in decode_transport(lk["user"]):
                print(f"    TRANSPORT tc=0x{tc:02x} FIR={bool(tc & 0x80)} "
                      f"FIN={bool(tc & 0x40)} seq={tc & 0x3F}")
                a = decode_app(app)
                print(f"    APP {a.get('func_name')} seq={a.get('app_seq')} "
                      f"FIR={a.get('app_fir')} FIN={a.get('app_fin')} "
                      f"con={a.get('app_con')}")
                if "iin" in a:
                    print(f"      IIN={a['iin_hex']}")
                for o in a.get("objects", []):
                    line = f"      OBJ g{o['group']}v{o['variation']} qual=0x{o['qualifier']:02x}"
                    if "count" in o:
                        line += f" count={o['count']}"
                    print(line)
                    for pt in o.get("points", []):
                        parts = [f"idx={pt.get('index', '?')}"]
                        if pt.get("value") is not None:
                            parts.append(str(pt["value"]))
                        if "status_hex" in pt:
                            # Group 41 command objects carry a command-status
                            # octet; other objects (e.g. Class-0 g1v1) carry a
                            # value+flags octet, labelled generically.
                            if o["group"] == 41:
                                parts.append(f"status={pt['status_hex']}")
                            else:
                                parts.append(f"octet={pt['status_hex']}")
                        print("        " + " ".join(parts))
            print(f"    HEX {fr.hex()}")
    print(f"\nTOTAL link frames: {total_frames}")


if __name__ == "__main__":
    main()
