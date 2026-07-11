// Package benchmarks provides performance benchmarks for the DNP3 implementation.
package benchmarks

import (
	"testing"

	"dnp3/internal/dll/crc"
	"dnp3/internal/dll/frame"
)

// BenchmarkCRC16 measures CRC-16-DNP calculation performance
func BenchmarkCRC16(b *testing.B) {
	data := make([]byte, 292) // Max frame data size
	for i := range data {
		data[i] = byte(i & 0xFF)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crc.CRC16(data)
	}
}

// BenchmarkCRC16Small measures CRC-16 performance for small data
func BenchmarkCRC16Small(b *testing.B) {
	data := []byte{0x05, 0x64, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crc.CRC16(data)
	}
}

// BenchmarkFrameEncode measures frame encoding performance
func BenchmarkFrameEncode(b *testing.B) {
	f := &frame.Frame{
		Control: frame.Control{
			DIR:  true,
			PRM:  true,
			FCB:  false,
			FCV:  false,
			FuncCode: frame.FuncConfirmedUserData,
		},
		SrcAddr:  1,
		DestAddr: 1024,
		Data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		frame.Encode(f)
	}
}

// BenchmarkFrameEncodeLargeData measures frame encoding with large data
func BenchmarkFrameEncodeLargeData(b *testing.B) {
	data := make([]byte, 200)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}

	f := &frame.Frame{
		Control: frame.Control{
			DIR:  true,
			PRM:  true,
			FCB:  false,
			FCV:  false,
			FuncCode: frame.FuncConfirmedUserData,
		},
		SrcAddr:  1,
		DestAddr: 1024,
		Data:     data,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		frame.Encode(f)
	}
}

// BenchmarkFrameDecode measures frame decoding performance
func BenchmarkFrameDecode(b *testing.B) {
	f := &frame.Frame{
		Control: frame.Control{
			DIR:  true,
			PRM:  true,
			FCB:  false,
			FCV:  false,
			FuncCode: frame.FuncConfirmedUserData,
		},
		SrcAddr:  1,
		DestAddr: 1024,
		Data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05},
	}

	encoded, _ := frame.Encode(f)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		frame.Decode(encoded)
	}
}

// BenchmarkFrameDecodeLargeData measures frame decoding with large data
func BenchmarkFrameDecodeLargeData(b *testing.B) {
	data := make([]byte, 200)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}

	f := &frame.Frame{
		Control: frame.Control{
			DIR:  true,
			PRM:  true,
			FCB:  false,
			FCV:  false,
			FuncCode: frame.FuncConfirmedUserData,
		},
		SrcAddr:  1,
		DestAddr: 1024,
		Data:     data,
	}

	encoded, _ := frame.Encode(f)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		frame.Decode(encoded)
	}
}

// BenchmarkFrameResetLinkEncode measures Reset Link Stations frame encoding
func BenchmarkFrameResetLinkEncode(b *testing.B) {
	f := &frame.Frame{
		Control: frame.Control{
			DIR:  false,
			PRM:  true,
			FCB:  false,
			FCV:  false,
			FuncCode: frame.FuncResetLinkStations,
		},
		SrcAddr:  1,
		DestAddr: 1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		frame.Encode(f)
	}
}

// BenchmarkFrameACKEncode measures ACK frame encoding
func BenchmarkFrameACKEncode(b *testing.B) {
	f := &frame.Frame{
		Control: frame.Control{
			DIR:  false,
			PRM:  false,
			DFC:  false,
			FuncCode: frame.FuncAck,
		},
		SrcAddr:  1024,
		DestAddr: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		frame.Encode(f)
	}
}

// BenchmarkControlByteToByte measures control byte encoding
func BenchmarkControlByteToByte(b *testing.B) {
	c := frame.Control{
		DIR:       true,
		PRM:       true,
		FCB:       false,
		FCV:       true,
		FuncCode:  frame.FuncConfirmedUserData,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.ToByte()
	}
}

// BenchmarkControlByteFromByte measures control byte decoding
func BenchmarkControlByteFromByte(b *testing.B) {
	c := frame.Control{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.FromByte(0xC4) // DIR=1, PRM=1, FCV=1, Func=4
	}
}
