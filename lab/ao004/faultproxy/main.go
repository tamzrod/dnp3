// AO-005 negative-interop fault proxy (LAB-ONLY test harness).
//
// PURPOSE: sit on the TCP path between an opted-in Go master and the isolated
// OpenDNP3 3.1.2 C++ outstation peer and deterministically inject an
// after-send fault: withhold the response to a DirectOperate, or disconnect
// immediately after the request is on the wire. This proves the master treats a
// command that was definitely transmitted (and applied by the peer) as an
// UNKNOWN outcome rather than a silent success, and that it does not blindly
// retry.
//
// SAFETY / SCOPE:
//   - Binds only 10.77.0.1:20000 (the master-side lab address); connects only to
//     10.77.0.2:20000 (the peer). Refuses any other address (fail closed).
//   - Test harness ONLY: no production code, no hardware I/O, no field endpoint.
package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

const (
	labMasterAddr   = "10.77.0.1"
	labPeerAddr     = "10.77.0.2"
	labPort         = 20000
	funcDirectOp    = 0x05
	funcWrite       = 0x02
	wireUserData    = 0x04 // link function UNCONFIRMED_USER_DATA
	wireConfUserDat = 0x03 // link function CONFIRMED_USER_DATA
)

// apduFuncScan returns the application function codes seen in a byte stream,
// scanning well-formed DLL frames. It is a lightweight independent scan (not a
// full decode) used only to decide when the first control request has crossed.
func apduFuncScan(stream []byte) (funcs []byte, directSeen bool) {
	i := 0
	for i+3 <= len(stream) {
		if stream[i] != 0x05 || stream[i+1] != 0x64 {
			i++
			continue
		}
		userLen := int(stream[i+2]) - 5
		if userLen < 1 {
			i++
			continue
		}
		nblk := (userLen + 15) / 16
		total := 2 + 1 + 5 + userLen + 2*nblk
		if i+total > len(stream) {
			break
		}
		// transport header at [10]; app control at [11]; func at [12]
		if stream[i+10] == 0xC0 && i+13 <= len(stream) {
			f := stream[i+12]
			funcs = append(funcs, f)
			if f == funcDirectOp || f == funcWrite {
				directSeen = true
			}
		}
		i += total
	}
	return funcs, directSeen
}

func main() {
	mode := flag.String("mode", "drop-response", "pass|drop-response|close-after-request")
	listenAddr := flag.String("listen", labMasterAddr, "listen address (must be the lab master address)")
	peerAddr := flag.String("peer", labPeerAddr, "peer address (must be the lab outstation address)")
	port := flag.Int("port", labPort, "TCP port")
	flag.Parse()

	if *listenAddr != labMasterAddr || *peerAddr != labPeerAddr || *port != labPort {
		fmt.Fprintf(os.Stderr, "faultproxy: refusing non-lab endpoint listen=%s peer=%s:%d (only %s -> %s:%d)\n",
			*listenAddr, *peerAddr, *port, labMasterAddr, labPeerAddr, labPort)
		os.Exit(2)
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *listenAddr, *port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "faultproxy: listen: %v\n", err)
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Printf("faultproxy: mode=%s listening on %s:%d -> %s:%d\n", *mode, *listenAddr, *port, *peerAddr, *port)

	conn, err := ln.Accept()
	if err != nil {
		fmt.Fprintf(os.Stderr, "faultproxy: accept: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	peer, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", *peerAddr, *port), 5*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "faultproxy: dial peer: %v\n", err)
		os.Exit(1)
	}
	defer peer.Close()
	fmt.Println("faultproxy: client connected, peer dialed")

	// client->peer direction, scanning for the first control request.
	var mu sync.Mutex
	faultArmed := false
	dropResponses := false
	done := make(chan struct{})

	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		var seen []byte
		for {
			n, err := conn.Read(buf)
			if n > 0 {
				chunk := buf[:n]
				mu.Lock()
				armed := faultArmed
				mu.Unlock()
				// Forward unless we are dropping everything after the fault.
				if !armed {
					if _, werr := peer.Write(chunk); werr != nil {
						return
					}
					seen = append(seen, chunk...)
					if _, direct := apduFuncScan(seen); direct && *mode != "pass" {
						mu.Lock()
						faultArmed = true
						if *mode == "close-after-request" {
							dropResponses = true
						}
						mu.Unlock()
						fmt.Println("faultproxy: control request on the wire; arming fault")
						switch *mode {
						case "close-after-request":
							// Tear the path down immediately, before the peer's
							// response can be forwarded, so the master observes a
							// post-send disconnect with no response.
							fmt.Println("faultproxy: closing client+peer connections after request")
							_ = conn.Close()
							_ = peer.Close()
							return
						case "drop-response":
							// keep reading (discard) but never forward the response
							fmt.Println("faultproxy: withholding peer response")
						}
					}
				}
			}
			if err != nil {
				return
			}
		}
	}()

	// peer->client direction: forward unless the fault is armed (drop-response)
	// or the path has been torn down (close-after-request).
	buf := make([]byte, 4096)
	for {
		n, err := peer.Read(buf)
		if n > 0 {
			mu.Lock()
			drop := dropResponses || (*mode == "drop-response" && faultArmed)
			mu.Unlock()
			if !drop {
				if _, werr := conn.Write(buf[:n]); werr != nil {
					return
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				return
			}
			return
		}
		select {
		case <-done:
			return
		default:
		}
	}
}
