// AO-005 negative-interop raw-frame injector (LAB-ONLY test harness).
//
// PURPOSE: send a single, hand-constructed DNP3 link frame carrying a Direct
// Operate (function 5) request to the isolated Go outstation fixture, so that
// malformed Group 41 request objects can be exercised deterministically. A
// normal master API can only emit valid objects, so this harness builds the
// wire bytes itself from IEEE 1815 first principles and never touches project
// encoding/decoding code.
//
// SAFETY / SCOPE:
//   - Refuses any target other than exactly 10.77.0.2:20000 (fail closed before
//     dialing).
//   - Sends at most: one Reset Link Stations frame, then one crafted user-data
//     frame. It reads (bounded) for the response and reports the exact bytes or
//     their absence. No field/NGCP/SCADA endpoint is ever contacted.
//   - Test harness ONLY: no production code, no hardware I/O.
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	labOutstationAddr = "10.77.0.2"
	labOutstationPort = 20000
	linkMasterAddr    = 3 // this injector
	linkOutstation    = 4 // the Go outstation
)

// crcDNP is CRC-16/DNP (poly 0x3D65 reflected 0xA6BC, init 0, xorout 0xFFFF),
// re-implemented here from the standard so the injector is independent of the
// project's CRC package.
func crcDNP(data []byte) uint16 {
	var crc uint16
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xA6BC
			} else {
				crc >>= 1
			}
		}
	}
	return crc ^ 0xFFFF
}

// encodeFrame builds one DNP3 DLL frame: sync, length, control, addresses,
// header CRC, then each 16-octet data block with its CRC (LSB first).
//
// The header CRC here is computed over the full 8-byte prefix INCLUDING the two
// sync octets, matching this project's link encoder/decoder. (The Go outstation
// validates the header CRC over bytes [0:8], so the injector must match it.)
func encodeFrame(control byte, dst, src uint16, data []byte) []byte {
	length := byte(5 + len(data)) // control(1)+dst(2)+src(2)+data
	out := []byte{0x05, 0x64, length, control, byte(dst), byte(dst >> 8), byte(src), byte(src >> 8)}
	hc := crcDNP(out[0:8])
	out = append(out, byte(hc), byte(hc>>8))
	for i := 0; i < len(data); i += 16 {
		end := i + 16
		if end > len(data) {
			end = len(data)
		}
		blk := data[i:end]
		out = append(out, blk...)
		bc := crcDNP(blk)
		out = append(out, byte(bc), byte(bc>>8))
	}
	return out
}

// apduUserData wraps an application APDU in one FIR|FIN transport fragment.
func apduUserData(funcCode byte, data []byte) []byte {
	apdu := append([]byte{0xC0, funcCode}, data...) // app control FIR|FIN, seq 0
	return append([]byte{0xC0}, apdu...)            // transport FIR|FIN, seq 0
}

// g41Object builds a Group 41 (0x29) request object: header(5) + index(2) +
// value + status(1).
func g41Object(variation byte, count uint16, index uint16, value []byte, status byte) []byte {
	obj := []byte{0x29, variation, 0x28, byte(count), byte(count >> 8), byte(index), byte(index >> 8)}
	obj = append(obj, value...)
	return append(obj, status)
}

func le32(v uint32) []byte { return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)} }
func le16(v uint16) []byte { return []byte{byte(v), byte(v >> 8)} }
func le64(v uint64) []byte {
	return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24), byte(v >> 32), byte(v >> 40), byte(v >> 48), byte(v >> 56)}
}

// caseObject returns the DirectOperate object bytes for a named case. funcCode
// is the application function code to use (5 = DirectOperate, 2 = WRITE).
func caseObject(name string) (obj []byte, funcCode byte) {
	funcCode = 0x05
	switch name {
	// Positive controls (valid idx0, one point) so the injector cannot masquerade
	// as an application regression.
	case "pos-v1":
		return g41Object(1, 1, 0, le32(2300), 0), funcCode
	case "pos-v2":
		return g41Object(2, 1, 0, le16(12345), 0), funcCode
	case "pos-v3":
		return g41Object(3, 1, 0, le32(0x3FC00000), 0), funcCode // 1.5f
	case "pos-v4":
		return g41Object(4, 1, 0, le64(0x3FF8000000000000), 0), funcCode // 1.5
	// Valid but out-of-range index (explicit non-success echo, zero writes).
	case "idx1-v1":
		return g41Object(1, 1, 1, le32(2300), 0), funcCode
	case "idx1-v4":
		return g41Object(4, 1, 1, le64(0x3FF8000000000000), 0), funcCode
	// Malformed: truncated value payloads.
	case "trunc-v1": // needs 4 value octets, send 3
		return g41Object(1, 1, 0, le32(2300)[:3], 0), funcCode
	case "trunc-v2": // needs 2, send 1
		return g41Object(2, 1, 0, le16(12345)[:1], 0), funcCode
	case "trunc-v3": // needs 4, send 3
		return g41Object(3, 1, 0, le32(0x3FC00000)[:3], 0), funcCode
	case "trunc-v4": // needs 8, send 7
		return g41Object(4, 1, 0, le64(0x3FF8000000000000)[:7], 0), funcCode
	// Malformed: wrong (too long) value payload.
	case "long-v1": // needs 4, send 5
		return g41Object(1, 1, 0, append(le32(2300), 0x00), 0), funcCode
	// Malformed: unsupported qualifier (0x17 = 16-bit count + 8-bit start/stop).
	case "badqual-17":
		return []byte{0x29, 0x01, 0x17, 0x01, 0x00, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x00}, funcCode
	// Malformed: unsupported variation.
	case "var0":
		return g41Object(0, 1, 0, le32(2300), 0), funcCode
	case "var5":
		return g41Object(5, 1, 0, le32(2300), 0), funcCode
	// Malformed: wrong count (multi-point contract violation).
	case "count2":
		return g41Object(1, 2, 0, le32(2300), 0), funcCode
	case "count0":
		return g41Object(1, 0, 0, le32(2300), 0), funcCode
	// Out-of-range index beyond the fixture's 0..1 window.
	case "idx2":
		return g41Object(1, 1, 2, le32(2300), 0), funcCode
	// G41 via WRITE must be rejected regardless of object shape.
	case "fc-write":
		return g41Object(1, 1, 0, le32(2300), 0), 0x02
	}
	return nil, funcCode
}

func main() {
	name := flag.String("case", "", "named injector case")
	host := flag.String("host", labOutstationAddr, "target address (must be the lab outstation)")
	port := flag.Int("port", labOutstationPort, "target port (must be the lab port)")
	readWait := flag.Duration("read-wait", 3*time.Second, "bounded window to collect the response")
	corruptCRC := flag.Bool("corrupt-header-crc", false, "flip the header CRC of the crafted frame (link-layer negative)")
	flag.Parse()

	if *host != labOutstationAddr || *port != labOutstationPort {
		fmt.Fprintf(os.Stderr, "injector: refusing non-lab target %s:%d (only %s:%d is permitted)\n", *host, *port, labOutstationAddr, labOutstationPort)
		os.Exit(2)
	}
	obj, funcCode := caseObject(*name)
	if obj == nil {
		fmt.Fprintf(os.Stderr, "injector: unknown case %q\n", *name)
		os.Exit(2)
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", *host, *port), 5*time.Second)
	if err != nil {
		fmt.Printf("injector: case=%s connect_error=%v\n", *name, err)
		os.Exit(1)
	}
	defer conn.Close()

	// Link reset first (wire func 0, DIR=1 PRM=1), then drain the ACK.
	reset := encodeFrame(0xC0, linkOutstation, linkMasterAddr, nil)
	_, _ = conn.Write(reset)
	time.Sleep(150 * time.Millisecond)

	// Crafted user-data frame (wire func 4, DIR=1 PRM=1).
	frame := encodeFrame(0xC4, linkOutstation, linkMasterAddr, apduUserData(funcCode, obj))
	if *corruptCRC {
		frame[8] ^= 0xFF // corrupt the transmitted header CRC octet
	}
	reqHex := fmt.Sprintf("%X", frame)
	if _, err := conn.Write(frame); err != nil {
		fmt.Printf("injector: case=%s write_error=%v\n", *name, err)
		os.Exit(1)
	}
	fmt.Printf("injector: case=%s func=0x%02X request_hex=%s corrupt_header_crc=%v\n", *name, funcCode, reqHex, *corruptCRC)

	// Bounded read window.
	deadline := time.Now().Add(*readWait)
	var buf []byte
	tmp := make([]byte, 4096)
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(deadline)
		n, err := conn.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
	}
	if len(buf) == 0 {
		fmt.Printf("injector: case=%s response=NONE\n", *name)
		return
	}
	fmt.Printf("injector: case=%s response_hex=%X\n", *name, buf)
	for _, f := range splitFrames(buf) {
		fmt.Printf("injector: case=%s frame=%X\n", *name, f)
	}
}

// splitFrames returns the well-formed DLL frames found in a byte stream.
func splitFrames(stream []byte) [][]byte {
	var frames [][]byte
	i := 0
	for i+3 <= len(stream) {
		if stream[i] != 0x05 || stream[i+1] != 0x64 {
			i++
			continue
		}
		userLen := int(stream[i+2]) - 5
		if userLen < 0 {
			i++
			continue
		}
		nblk := (userLen + 15) / 16
		total := 2 + 1 + 5 + userLen + 2*nblk
		if i+total > len(stream) {
			break
		}
		frames = append(frames, stream[i:i+total])
		i += total
	}
	return frames
}
