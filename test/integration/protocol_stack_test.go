// Package integration provides integration tests for the complete DNP3 protocol stack.
package integration

import (
	"bytes"
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/outstation"
	"dnp3/internal/tl"
)

// TestProtocolStackMasterToOutstation tests the complete send path:
// AL encode → TL fragment → DLL encode
func TestProtocolStackMasterToOutstation(t *testing.T) {
	apdu := &al.APDU{
		Control: al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncRead,
		Data:     []byte{0x0C, 0x01, 0x07, 0x00},
	}

	// AL encode
	apduData := apdu.Encode()
	if len(apduData) == 0 {
		t.Fatal("APDU encode returned empty data")
	}
	t.Logf("AL: Encoded APDU (%d bytes)", len(apduData))

	// TL fragment
	fragmenter := tl.NewFragmenter()
	fragments := fragmenter.Fragmentize(apduData)
	t.Logf("TL: Created %d fragment(s)", len(fragments))

	// DLL encode each fragment
	for i, frag := range fragments {
		tlEncoded := tl.EncodeFragment(frag)
		dllFr := &frame.Frame{
			Control: frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncConfirmedUserData},
			DestAddr: 1024, SrcAddr: 1,
			Data: tlEncoded,
		}
		dllEncoded, err := frame.Encode(dllFr)
		if err != nil {
			t.Fatalf("DLL encode failed for fragment %d: %v", i, err)
		}

		// Decode back to verify
		decoded, err := frame.Decode(dllEncoded)
		if err != nil {
			t.Fatalf("DLL decode failed for fragment %d: %v", i, err)
		}

		if decoded.DestAddr != 1024 {
			t.Errorf("Fragment %d: DestAddr mismatch", i)
		}

		t.Logf("Fragment %d: DLL frame (%d bytes) verified", i, len(dllEncoded))
	}

	t.Log("Master to Outstation protocol stack verified")
}

// TestProtocolStackOutstationToMaster tests the complete receive path:
// DLL decode → TL reassemble → AL decode
func TestProtocolStackOutstationToMaster(t *testing.T) {
	origApdu := &al.APDU{
		Control: al.AppControl{FIR: true, FIN: true, Seq: 5},
		FuncCode: al.FuncResponse,
		Data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05},
	}

	// Build DLL frames
	apduData := origApdu.Encode()
	fragmenter := tl.NewFragmenter()
	fragments := fragmenter.Fragmentize(apduData)

	var tcpData []byte
	for _, frag := range fragments {
		tlEncoded := tl.EncodeFragment(frag)
		dllFr := &frame.Frame{
			Control: frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncConfirmedUserData},
			DestAddr: 1, SrcAddr: 1024,
			Data: tlEncoded,
		}
		enc, _ := frame.Encode(dllFr)
		tcpData = append(tcpData, enc...)
	}

	t.Logf("Built %d DLL frame(s) totaling %d bytes", len(fragments), len(tcpData))

	// Receive path
	reassembler := tl.NewReassembler()
	offset := 0

	for offset < len(tcpData) {
		dllDecoded, err := frame.Decode(tcpData[offset:])
		if err != nil {
			break
		}
		offset += frame.EncodedSize(len(dllDecoded.Data))

		tlFrag, err := tl.DecodeFragment(dllDecoded.Data)
		if err != nil {
			t.Fatalf("TL decode failed: %v", err)
		}

		msg, _ := reassembler.Push(tlFrag)
		if msg != nil {
			decodedApdu, err := al.Decode(msg)
			if err != nil {
				t.Fatalf("AL decode failed: %v", err)
			}

			if decodedApdu.FuncCode != origApdu.FuncCode {
				t.Errorf("FuncCode mismatch")
			}
			if decodedApdu.Control.Seq != origApdu.Control.Seq {
				t.Errorf("Seq mismatch")
			}
			if !bytes.Equal(decodedApdu.Data, origApdu.Data) {
				t.Error("Data mismatch")
			}
			break
		}
	}

	t.Log("Outstation to Master protocol stack verified")
}

// TestProtocolStackRoundTrip tests complete Master↔Outstation round trip
func TestProtocolStackRoundTrip(t *testing.T) {
	// Master request
	masterApdu := &al.APDU{
		Control: al.AppControl{FIR: true, FIN: true, Seq: 3},
		FuncCode: al.FuncRead,
		Data:     []byte{0x0C, 0x01, 0x07, 0x00},
	}

	// MASTER → OUTSTATION
	apduData := masterApdu.Encode()
	fragmenter := tl.NewFragmenter()
	frags := fragmenter.Fragmentize(apduData)

	var masterTxData []byte
	for _, frag := range frags {
		tlEnc := tl.EncodeFragment(frag)
		dllFr := &frame.Frame{
			Control: frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncConfirmedUserData},
			DestAddr: 1024, SrcAddr: 1, Data: tlEnc,
		}
		enc, _ := frame.Encode(dllFr)
		masterTxData = append(masterTxData, enc...)
	}

	// OUTSTATION receives - decode each frame properly
	outReasm := tl.NewReassembler()
	var outRecvMsg []byte
	offset := 0
	for offset < len(masterTxData) {
		remaining := masterTxData[offset:]
		dllDec, err := frame.Decode(remaining)
		if err != nil {
			t.Logf("Frame decode error at offset %d: %v", offset, err)
			offset++
			continue
		}

		// Calculate frame size for next iteration using the DNP3 wire model.
		offset += frame.EncodedSize(len(dllDec.Data))

		// Parse TL fragment
		tlFrag, err := tl.DecodeFragment(dllDec.Data)
		if err != nil {
			continue
		}
		msg, err := outReasm.Push(tlFrag)
		if err != nil {
			t.Logf("Reassembler error: %v", err)
			continue
		}
		if msg != nil {
			outRecvMsg = msg
		}
	}
	if outRecvMsg == nil {
		t.Fatal("Outstation did not receive complete message")
	}
	recvApdu, _ := al.Decode(outRecvMsg)

	// OUTSTATION responds
	respApdu := &al.APDU{
		Control: al.AppControl{FIR: true, FIN: true, Seq: recvApdu.Control.Seq},
		FuncCode: al.FuncResponse,
		Data:     []byte{0x01, 0x02, 0x03},
	}

	respData := respApdu.Encode()
	respFrags := fragmenter.Fragmentize(respData)

	var outTxData []byte
	for _, frag := range respFrags {
		tlEnc := tl.EncodeFragment(frag)
		dllFr := &frame.Frame{
			Control: frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncConfirmedUserDataR},
			DestAddr: 1, SrcAddr: 1024, Data: tlEnc,
		}
		enc, _ := frame.Encode(dllFr)
		outTxData = append(outTxData, enc...)
	}

	// MASTER receives - decode each frame properly
	masterReasm := tl.NewReassembler()
	var masterRecvMsg []byte
	offset = 0
	for offset < len(outTxData) {
		remaining := outTxData[offset:]
		dllDec, err := frame.Decode(remaining)
		if err != nil {
			t.Logf("Frame decode error at offset %d: %v", offset, err)
			offset++
			continue
		}

		// Calculate frame size for next iteration
		offset += frame.EncodedSize(len(dllDec.Data))

		// Parse TL fragment
		tlFrag, err := tl.DecodeFragment(dllDec.Data)
		if err != nil {
			continue
		}
		msg, err := masterReasm.Push(tlFrag)
		if err != nil {
			t.Logf("Reassembler error: %v", err)
			continue
		}
		if msg != nil {
			masterRecvMsg = msg
		}
	}
	if masterRecvMsg == nil {
		t.Fatal("Master did not receive complete message")
	}
	recvResp, _ := al.Decode(masterRecvMsg)

	// Verify
	if recvResp.FuncCode != respApdu.FuncCode {
		t.Errorf("Response FuncCode mismatch")
	}
	if recvResp.Control.Seq != respApdu.Control.Seq {
		t.Errorf("Response Seq mismatch")
	}

	t.Log("Complete Master to Outstation round trip verified")
}

// TestProtocolStackMultiFragment tests multi-fragment handling
func TestProtocolStackMultiFragment(t *testing.T) {
	// Use 500 bytes which will create multiple fragments
	// Each TL fragment is max 249 bytes, so we need at least 3 fragments
	largeData := make([]byte, 500)
	for i := range largeData {
		largeData[i] = byte(i)
	}

	apdu := &al.APDU{
		Control: al.AppControl{FIR: true, FIN: true, Seq: 7},
		FuncCode: al.FuncRead,
		Data:     largeData,
	}

	apduData := apdu.Encode()
	fragmenter := tl.NewFragmenter()
	frags := fragmenter.Fragmentize(apduData)

	t.Logf("Large APDU (%d bytes) fragmented into %d pieces", len(apduData), len(frags))

	if len(frags) < 2 {
		t.Fatal("Expected multi-fragment APDU")
	}

	// Verify all fragments can be encoded/decoded
	for i, frag := range frags {
		t.Logf("Fragment %d: FIR=%v, FIN=%v, Seq=%d, DataLen=%d",
			i, frag.FIR, frag.FIN, frag.Seq, len(frag.Data))

		// Verify fragment data fits within DLL limits
		if len(frag.Data) > int(frame.MaxDataSize) {
			t.Fatalf("Fragment %d data (%d bytes) exceeds MaxDataSize (%d bytes)",
				i, len(frag.Data), frame.MaxDataSize)
		}

		tlEnc := tl.EncodeFragment(frag)
		t.Logf("Fragment %d: TL encoded len=%d", i, len(tlEnc))

		dllFr := &frame.Frame{
			Control: frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncConfirmedUserData},
			DestAddr: 1024, SrcAddr: 1, Data: tlEnc,
		}
		dllEnc, err := frame.Encode(dllFr)
		if err != nil {
			t.Fatalf("Fragment %d: DLL encode failed: %v", i, err)
		}
		t.Logf("Fragment %d: DLL encoded len=%d", i, len(dllEnc))

		if _, err := frame.Decode(dllEnc); err != nil {
			t.Fatalf("Fragment %d: DLL decode failed: %v", i, err)
		}
		t.Logf("Fragment %d: DLL decoded successfully", i)
	}

	// Test reassembly with a single reassembler
	t.Log("Testing reassembly with single reassembler...")
	reassembler := tl.NewReassembler()
	var completeMsg []byte

	for i, frag := range frags {
		tlEnc := tl.EncodeFragment(frag)
		dllFr := &frame.Frame{
			Control: frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncConfirmedUserData},
			DestAddr: 1024, SrcAddr: 1, Data: tlEnc,
		}
		dllEnc, _ := frame.Encode(dllFr)
		decoded, _ := frame.Decode(dllEnc)
		tlFrag, _ := tl.DecodeFragment(decoded.Data)
		msg, err := reassembler.Push(tlFrag)
		if err != nil {
			t.Fatalf("Fragment %d: Reassembler error: %v", i, err)
		}
		if msg != nil {
			completeMsg = msg
			t.Logf("Reassembly complete after fragment %d: %d bytes", i, len(msg))
		}
	}

	if completeMsg == nil {
		t.Error("Reassembly did not complete")
	} else if len(completeMsg) != len(apduData) {
		t.Errorf("Reassembled message length mismatch: got %d, want %d", len(completeMsg), len(apduData))
	}

	t.Log("Multi-fragment handling verified")
}

// TestProtocolStackOutstationDirectProcessing tests outstation.ProcessRequest
func TestProtocolStackOutstationDirectProcessing(t *testing.T) {
	ost := outstation.NewOutstation(nil)
	ost.Initialize()
	ost.Start()

	request := &al.APDU{
		Control: al.AppControl{FIR: true, FIN: true, Seq: 2},
		FuncCode: al.FuncRead,
		Data:     []byte{0x0C, 0x01, 0x07, 0x00},
	}

	response, err := ost.ProcessRequest(request)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response")
	}

	if response.FuncCode != al.FuncResponse {
		t.Errorf("FuncCode = %d, want %d", response.FuncCode, al.FuncResponse)
	}

	if len(response.Data) < 10 {
		t.Errorf("Response data too short: %d bytes", len(response.Data))
	}

	t.Logf("Direct processing verified: response (%d bytes)", len(response.Data))
}

// TestProtocolStackFrameValidation tests DLL frame validation
func TestProtocolStackFrameValidation(t *testing.T) {
	tests := []struct {
		name     string
		destAddr uint16
		wantSkip bool
	}{
		{"Correct address", 1024, false},
		{"Wrong address", 9999, true},
		{"Broadcast", 0xFFFF, true},
	}

	const outstationAddr = 1024
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &frame.Frame{
				Control: frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncConfirmedUserData},
				DestAddr: tt.destAddr, SrcAddr: 1,
				Data: []byte{0x01},
			}
			enc, _ := frame.Encode(f)
			dec, _ := frame.Decode(enc)

			shouldProcess := dec.DestAddr == outstationAddr
			if shouldProcess == tt.wantSkip {
				t.Errorf("Address validation incorrect for %s", tt.name)
			}
		})
	}
}
