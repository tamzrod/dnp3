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
		headerSize := 10
		crcSize := ((len(dllDecoded.Data) + 1) / 2) * 2
		offset += headerSize + len(dllDecoded.Data) + crcSize

		tlFrag, err := tl.DecodeFragment(dllDecoded.Data)
		if err != nil {
			t.Fatalf("TL decode failed: %v", err)
		}

		msg, _ := reassembler.Push(tlFrag)
		if msg != nil {
			break
		}
	}

	// Decode AL
	decodedApdu, err := al.Decode(reassembler.GetData())
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

	// OUTSTATION receives
	outReasm := tl.NewReassembler()
	offset := 0
	for offset < len(masterTxData) {
		dllDec, _ := frame.Decode(masterTxData[offset:])
		headerSize := 10
		crcSize := ((len(dllDec.Data)+1)/2)*2 - ((len(dllDec.Data)+1)%2)
		offset += headerSize + len(dllDec.Data) + crcSize
		tlFrag, _ := tl.DecodeFragment(dllDec.Data)
		outReasm.Push(tlFrag)
	}
	recvApdu, _ := al.Decode(outReasm.GetData())

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
			Control: frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncConfirmedUserData},
			DestAddr: 1, SrcAddr: 1024, Data: tlEnc,
		}
		enc, _ := frame.Encode(dllFr)
		outTxData = append(outTxData, enc...)
	}

	// MASTER receives
	masterReasm := tl.NewReassembler()
	offset = 0
	for offset < len(outTxData) {
		dllDec, _ := frame.Decode(outTxData[offset:])
		headerSize := 10
		crcSize := ((len(dllDec.Data)+1)/2)*2 - ((len(dllDec.Data)+1)%2)
		offset += headerSize + len(dllDec.Data) + crcSize
		tlFrag, _ := tl.DecodeFragment(dllDec.Data)
		masterReasm.Push(tlFrag)
	}
	recvResp, _ := al.Decode(masterReasm.GetData())

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
	largeData := make([]byte, 600)
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

	for i, frag := range frags {
		tlEnc := tl.EncodeFragment(frag)
		dllFr := &frame.Frame{
			Control: frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncConfirmedUserData},
			DestAddr: 1024, SrcAddr: 1, Data: tlEnc,
		}
		dllEnc, err := frame.Encode(dllFr)
		if err != nil {
			t.Fatalf("Fragment %d: DLL encode failed: %v", i, err)
		}

		decoded, err := frame.Decode(dllEnc)
		if err != nil {
			t.Fatalf("Fragment %d: DLL decode failed: %v", i, err)
		}

		reassembler := tl.NewReassembler()
		tlFrag, _ := tl.DecodeFragment(decoded.Data)
		msg, _ := reassembler.Push(tlFrag)

		if i == len(frags)-1 && msg == nil {
			t.Error("Last fragment should complete reassembly")
		}
	}

	t.Log("Multi-fragment protocol stack verified")
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
