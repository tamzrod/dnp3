#!/usr/bin/env python3
"""AO-005 negative-interop independent raw-PCAP decoder (lab-only evidence tool).

Groups the capture by TCP connection (each negative case is a fresh connection),
then, for each connection, reports the literal on-wire request frame(s) sent by
the master/injector, the response frame(s) returned by the outstation/peer (or
their absence), and a decoded summary. The wire parsing is re-used from the
existing AO-005 decoder (itself an IEEE 1815 first-principles re-implementation
with no project code), so this is an independent check of both endpoints.
"""
import importlib.util
import os
import sys

_HERE = os.path.dirname(os.path.abspath(__file__))
_spec = importlib.util.spec_from_file_location("ao005dec", os.path.join(_HERE, "ao005-decode-pcap.py"))
dec = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(dec)


def summarize_stream(stream):
    """Return (requests, responses, reset_frames) for one TCP connection. A
    malformed object that the independent parser cannot fully walk is reported
    with objdesc=None (the literal request hex is always preserved)."""
    reqs, resps, resets = [], [], []
    for fr in dec.extract_link_frames(stream):
        lk = dec.parse_link(fr)
        if lk["func"] == 0:  # link reset / ack (no user data)
            resets.append((lk["dir"], fr.hex(), lk["header_crc_ok"]))
            continue
        for _tc, app in dec.decode_transport(lk["user"]):
            try:
                a = dec.decode_app(app)
            except Exception:
                a = {"func_name": None, "_parse_error": True}
            fn = a.get("func_name")
            if lk["prm"] and fn in ("DIRECT_OPERATE", "WRITE"):
                reqs.append((fn, a, fr.hex()))
            elif fn == "RESPONSE":
                resps.append((a, fr.hex()))
            else:
                reqs.append((fn or "?", a, fr.hex()))
    return reqs, resps, resets


def group_connections(pkts):
    """Group packets into TCP connections (unordered endpoint pair), preserving
    per-connection packet order across both directions. Returns a list of
    ((a,b), concatenated_bytes, first_client_hint)."""
    order = []
    streams = {}
    meta = {}
    for p in pkts:
        r = dec.parse_eth_ip_tcp(p)
        if not r:
            continue
        src, dst, sport, dport, payload = r
        key = tuple(sorted([(src, sport), (dst, dport)]))
        if key not in streams:
            streams[key] = bytearray()
            meta[key] = (src, sport, dst, dport)
            order.append(key)
        streams[key] += payload
    return [(k, bytes(streams[k]), meta[k]) for k in order]


def main():
    path = sys.argv[1]
    label = sys.argv[2] if len(sys.argv) > 2 else "?"
    pkts, _nano = dec.read_pcap(path)
    conn = 0
    for key, stream, (src, sport, dst, dport) in group_connections(pkts):
        conn += 1
        try:
            reqs, resps, resets = summarize_stream(stream)
        except Exception as e:  # never abort the whole evidence dump on one frame
            print(f"CONN {conn} {label} DECODE_ERROR {e}")
            continue
        if not reqs and not resps:
            print(f"CONN {conn} {label} no-application-frames bytes={len(stream)}")
            continue
        for fn, a, hexr in reqs:
            if a.get("_parse_error"):
                objdesc = "OBJECT_PARSE_LIMIT (independent parser could not fully walk the object)"
            else:
                parts = []
                for o in a.get("objects", []):
                    if o.get("group") == 41:
                        pts = o.get("points", [])
                        pt = pts[0] if pts else {}
                        parts.append(f"g{o['group']}v{o['variation']} qual=0x{o['qualifier']:02x} "
                                     f"count={o.get('count')} idx={pt.get('index')} val={pt.get('value')}")
                    else:
                        parts.append(f"g{o.get('group')}v{o.get('variation')} qual=0x{o.get('qualifier', 0):02x}"
                                     + (f" tail={o['raw_tail']}" if "raw_tail" in o else ""))
                objdesc = " ; ".join(parts)
            print(f"CONN {conn} {label} REQUEST func={fn} {objdesc} hex={hexr}")
        if not resps:
            print(f"CONN {conn} {label} RESPONSE none")
        for a, hexp in resps:
            st = ""
            for o in a.get("objects", []):
                for pt in o.get("points", []):
                    st = f"status={pt.get('status_hex')} idx={pt.get('index')}"
            print(f"CONN {conn} {label} RESPONSE iin={a.get('iin_hex')} {st} hex={hexp}")


if __name__ == "__main__":
    main()
