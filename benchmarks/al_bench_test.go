// Package benchmarks provides performance benchmarks for the DNP3 implementation.
package benchmarks

import (
	"testing"

	"dnp3/internal/al"
)

// BenchmarkAPDUEncodeRequest measures APDU request encoding
func BenchmarkAPDUEncodeRequest(b *testing.B) {
	req := al.NewRequest(5, al.FuncRead)
	req.Data = []byte{0x01, 0x02, 0x03, 0x04}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.Encode()
	}
}

// BenchmarkAPDUEncodeResponse measures APDU response encoding
func BenchmarkAPDUEncodeResponse(b *testing.B) {
	resp := al.NewAppResponse(5, al.IIN{}, []byte{0x01, 0x02, 0x03, 0x04})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp.Encode()
	}
}

// BenchmarkAPDUDecodeRequest measures APDU request decoding
func BenchmarkAPDUDecodeRequest(b *testing.B) {
	req := al.NewRequest(5, al.FuncRead)
	req.Data = []byte{0x01, 0x02, 0x03, 0x04}
	encoded := req.Encode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		al.Decode(encoded)
	}
}

// BenchmarkAPDUDecodeResponse measures APDU response decoding
func BenchmarkAPDUDecodeResponse(b *testing.B) {
	resp := al.NewAppResponse(5, al.IIN{}, []byte{0x01, 0x02, 0x03, 0x04})
	encoded := resp.Encode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		al.DecodeResponse(encoded)
	}
}

// BenchmarkAPDUEncodeLargeData measures APDU encoding with large data
func BenchmarkAPDUEncodeLargeData(b *testing.B) {
	data := make([]byte, 200)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}
	req := al.NewRequest(5, al.FuncRead)
	req.Data = data

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.Encode()
	}
}

// BenchmarkAPDUDecodeLargeData measures APDU decoding with large data
func BenchmarkAPDUDecodeLargeData(b *testing.B) {
	data := make([]byte, 200)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}
	req := al.NewRequest(5, al.FuncRead)
	req.Data = data
	encoded := req.Encode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		al.Decode(encoded)
	}
}

// BenchmarkAppControlHeader measures application control field encoding
func BenchmarkAppControlHeader(b *testing.B) {
	ctrl := al.AppControl{
		FIR: true,
		FIN: true,
		CON: false,
		UNS: false,
		Seq: 5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctrl.Header()
	}
}

// BenchmarkIINEncode measures IIN encoding
func BenchmarkIINEncode(b *testing.B) {
	iin := al.IIN{
		AllStop:       true,
		ByteOver:      true,
		CheckFail:     true,
		ConfigError:   true,
		NeedsTimeSync: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iin.Bytes()
	}
}

// BenchmarkIINDecode measures IIN decoding
func BenchmarkIINDecode(b *testing.B) {
	encoded := []byte{0xFF, 0xFF}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		al.DecodeIIN(encoded)
	}
}

// BenchmarkNewUnsolicited measures unsolicited response creation
func BenchmarkNewUnsolicited(b *testing.B) {
	data := []byte{0x01, 0x02, 0x03, 0x04}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		al.NewUnsolicited(5, data)
	}
}

// BenchmarkIsValidFunctionCode measures function code validation
func BenchmarkIsValidFunctionCode(b *testing.B) {
	codes := []uint8{0, 1, 2, 3, 4, 5, 6, 7, 10, 21, 27, 28, 32, 41, 42, 48, 127}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, code := range codes {
			al.IsValidFunctionCode(code)
		}
	}
}

// BenchmarkIsReadFunction measures read function detection
func BenchmarkIsReadFunction(b *testing.B) {
	codes := []uint8{0, 1, 2, 3, 4, 5, 6, 7, 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, code := range codes {
			al.IsReadFunction(code)
		}
	}
}

// BenchmarkIsControlFunction measures control function detection
func BenchmarkIsControlFunction(b *testing.B) {
	codes := []uint8{0, 1, 2, 3, 4, 5, 6, 7, 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, code := range codes {
			al.IsControlFunction(code)
		}
	}
}

// BenchmarkFunctionCodeName measures function code name lookup
func BenchmarkFunctionCodeName(b *testing.B) {
	codes := []uint8{0, 2, 3, 4, 5, 6, 7, 21, 27, 28, 32, 41, 42}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, code := range codes {
			al.FunctionCodeName(code)
		}
	}
}
