#!/usr/bin/env python3
"""Lab-only fault injector in front of the independent OpenDNP3 master.

Stock OpenDNP3 3.1.2 confirms an unsolicited fragment once. It does not emit
a second application CONFIRM, and it does not withhold one. This proxy sits
on 127.0.0.1 between that master and the Go outstation.

The application CONFIRM octets (control, function 0x00, sequence) are the
bytes OpenDNP3 wrote. The proxy does not build a confirm.

Modes:
  duplicate  Hold the first application CONFIRM that matches the unsolicited
             sequence just forwarded. When /tmp/evt-proxy-release appears,
             write that same link frame to the outstation twice. Later
             confirms are forwarded once. The second write is the injection.
  withhold   Do not write application CONFIRM frames toward the outstation.
             They remain on the master-to-proxy TCP leg. Creating
             /tmp/evt-proxy-pass restores forwarding for a later class read.

Every other frame is forwarded unchanged. Bind and upstream are loopback.
"""

import hashlib
import os
import socket
import sys
import threading
import time

RELEASE = "/tmp/evt-proxy-release"
PASSTHROUGH = "/tmp/evt-proxy-pass"


def log(fh, lock, msg):
    line = msg + "\n"
    with lock:
        sys.stdout.write(line)
        sys.stdout.flush()
        fh.write(line)
        fh.flush()


def frame_total(buf):
    if len(buf) < 3 or buf[0] != 0x05 or buf[1] != 0x64:
        return None
    length = buf[2]
    if length < 5 or length > 255:
        return None
    user = length - 5
    blocks = 0 if user == 0 else (user + 15) // 16
    return 10 + user + 2 * blocks


def pop_frame(buf):
    while True:
        if len(buf) < 3:
            return None, buf
        start = buf.find(b"\x05\x64")
        if start < 0:
            return None, buf[-1:] if buf[-1:] == b"\x05" else b""
        if start:
            buf = buf[start:]
        total = frame_total(buf)
        if total is None:
            buf = buf[2:]
            continue
        if len(buf) < total:
            return None, buf
        return buf[:total], buf[total:]


def user_data(frame):
    length = frame[2]
    user = length - 5
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


def app_header(frame):
    user = user_data(frame)
    if user is None or len(user) < 3:
        return None
    # user[0] is the transport header. Application control and function follow.
    return user[1], user[2]


def read_frames(sock, buf):
    try:
        chunk = sock.recv(4096)
    except socket.timeout:
        return [], buf, False
    except OSError:
        return [], buf, True
    if not chunk:
        return [], buf, True
    buf += chunk
    frames = []
    while True:
        frame, buf = pop_frame(buf)
        if frame is None:
            break
        frames.append(frame)
    return frames, buf, False


def main():
    if len(sys.argv) != 5:
        sys.stderr.write("usage: confirm_proxy.py <listen> <upstream> <duplicate|withhold> <logfile>\n")
        return 2
    listen = int(sys.argv[1])
    upstream_port = int(sys.argv[2])
    mode = sys.argv[3]
    if mode not in ("duplicate", "withhold"):
        sys.stderr.write("mode must be duplicate or withhold\n")
        return 2
    fh = open(sys.argv[4], "w")
    lock = threading.Lock()
    log(fh, lock, "PROXY mode=%s listen=%d upstream=%d" % (mode, listen, upstream_port))
    if mode == "duplicate":
        log(fh, lock, "INJECTION stock=OpenDNP3 writes one application CONFIRM")
        log(fh, lock, "INJECTION action=hold that link frame, then write the same bytes to the outstation twice")
    else:
        log(fh, lock, "INJECTION stock=OpenDNP3 writes an application CONFIRM for each unsolicited fragment")
        log(fh, lock, "INJECTION action=do not write CONFIRM frames toward the outstation until passthrough")

    for path in (RELEASE, PASSTHROUGH):
        try:
            os.remove(path)
        except OSError:
            pass

    gate = threading.Event()
    state = {
        "last_unsol_seq": None,
        "duplicated": False,
        "unsol": 0,
        "dropped": 0,
        "dup_sent": False,
    }
    state_lock = threading.Lock()

    def watch_files():
        while not gate.is_set():
            if mode == "duplicate" and os.path.exists(RELEASE):
                gate.set()
                return
            time.sleep(0.02)

    threading.Thread(target=watch_files, daemon=True).start()

    server = socket.socket()
    server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server.bind(("127.0.0.1", listen))
    server.listen(1)
    server.settimeout(20)
    log(fh, lock, "PROXY LISTENING")
    try:
        master, _ = server.accept()
    except socket.timeout:
        log(fh, lock, "PROXY FAIL accept timeout")
        return 1
    master.setsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY, 1)
    master.settimeout(30)
    upstream = socket.create_connection(("127.0.0.1", upstream_port), 5)
    upstream.setsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY, 1)
    upstream.settimeout(30)
    log(fh, lock, "PROXY CONNECTED")

    def handle_master_frame(frame):
        header = app_header(frame)
        if header is None or header[1] != 0x00:
            upstream.sendall(frame)
            return
        ctrl, _ = header
        seq = ctrl & 0x0F
        if mode == "withhold" and not os.path.exists(PASSTHROUGH):
            with state_lock:
                state["dropped"] += 1
            log(fh, lock, "DROPPED_CONFIRM seq=%d ctrl=%02x func=00" % (seq, ctrl))
            return
        if mode == "duplicate":
            with state_lock:
                already = state["duplicated"]
                expect = state["last_unsol_seq"]
            if not already and expect is not None and seq == expect:
                log(fh, lock, "HELD_CONFIRM seq=%d ctrl=%02x func=00 bytes=%s" % (seq, ctrl, frame.hex()))
                if not gate.wait(1.5):
                    log(fh, lock, "PROXY FAIL release timeout")
                    upstream.sendall(frame)
                    return
                upstream.sendall(frame)
                time.sleep(0.15)
                upstream.sendall(frame)
                with state_lock:
                    state["duplicated"] = True
                    state["dup_sent"] = True
                log(fh, lock, "DUP_SENT n=2 seq=%d identical=1" % seq)
                return
        upstream.sendall(frame)

    def close_both():
        for sock in (master, upstream):
            try:
                sock.shutdown(socket.SHUT_RDWR)
            except OSError:
                pass

    def master_to_upstream():
        buf = b""
        try:
            while True:
                frames, buf, done = read_frames(master, buf)
                for frame in frames:
                    handle_master_frame(frame)
                if done:
                    return
        finally:
            close_both()

    def upstream_to_master():
        buf = b""
        try:
            while True:
                frames, buf, done = read_frames(upstream, buf)
                for frame in frames:
                    header = app_header(frame)
                    if header is not None and header[1] == 0x82:
                        seq = header[0] & 0x0F
                        digest = hashlib.sha256(frame).hexdigest()
                        with state_lock:
                            state["last_unsol_seq"] = seq
                            state["unsol"] += 1
                            n = state["unsol"]
                        log(fh, lock, "UNSOL n=%d seq=%d ctrl=%02x sha=%s" % (n, seq, header[0], digest))
                    master.sendall(frame)
                if done:
                    return
        finally:
            close_both()

    threads = [
        threading.Thread(target=master_to_upstream),
        threading.Thread(target=upstream_to_master),
    ]
    for t in threads:
        t.start()
    for t in threads:
        t.join(timeout=40)
    try:
        master.close()
    except OSError:
        pass
    try:
        upstream.close()
    except OSError:
        pass
    with state_lock:
        log(
            fh,
            lock,
            "PROXY_END mode=%s unsol=%d dropped_confirm=%d dup_sent=%d"
            % (mode, state["unsol"], state["dropped"], 1 if state["dup_sent"] else 0),
        )
    return 0


if __name__ == "__main__":
    sys.exit(main())
