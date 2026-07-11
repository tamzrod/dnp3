// Package benchmarks provides performance benchmarks for the DNP3 implementation.
package benchmarks

import (
	"testing"

	"dnp3/internal/tl"
)

// BenchmarkFragmentizeSmall measures fragmentation for small data
func BenchmarkFragmentizeSmall(b *testing.B) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	fragmenter := tl.NewFragmenter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fragmenter.Fragmentize(data)
	}
}

// BenchmarkFragmentizeLarge measures fragmentation for large data
func BenchmarkFragmentizeLarge(b *testing.B) {
	data := make([]byte, 600)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}
	fragmenter := tl.NewFragmenter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fragmenter.Fragmentize(data)
	}
}

// BenchmarkFragmentizeMaxSize measures fragmentation for max-sized data
func BenchmarkFragmentizeMaxSize(b *testing.B) {
	data := make([]byte, tl.MaxFragmentData)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}
	fragmenter := tl.NewFragmenter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fragmenter.Fragmentize(data)
	}
}

// BenchmarkReassembleSmall measures reassembly for small data
func BenchmarkReassembleSmall(b *testing.B) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	fragmenter := tl.NewFragmenter()
	fragments := fragmenter.Fragmentize(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := tl.NewReassembler()
		for _, f := range fragments {
			r.Push(f)
		}
	}
}

// BenchmarkReassembleMultiFragment measures reassembly for multi-fragment data
func BenchmarkReassembleMultiFragment(b *testing.B) {
	data := make([]byte, 600)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}
	fragmenter := tl.NewFragmenter()
	fragments := fragmenter.Fragmentize(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := tl.NewReassembler()
		for _, f := range fragments {
			r.Push(f)
		}
	}
}

// BenchmarkEncodeFragment measures fragment encoding
func BenchmarkEncodeFragment(b *testing.B) {
	fragment := tl.Fragment{
		FIR: true,
		FIN: true,
		Seq: 0,
		Data: make([]byte, 100),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tl.EncodeFragment(fragment)
	}
}

// BenchmarkDecodeFragment measures fragment decoding
func BenchmarkDecodeFragment(b *testing.B) {
	fragment := tl.Fragment{
		FIR: true,
		FIN: true,
		Seq: 0,
		Data: make([]byte, 100),
	}
	encoded := tl.EncodeFragment(fragment)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tl.DecodeFragment(encoded)
	}
}

// BenchmarkFragmentHeader measures header byte operations
func BenchmarkFragmentHeader(b *testing.B) {
	f := tl.Fragment{
		FIR: true,
		FIN: true,
		Seq: 15,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Header()
	}
}

// BenchmarkFragmentSetHeader measures header byte parsing
func BenchmarkFragmentSetHeader(b *testing.B) {
	var f tl.Fragment

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.SetHeader(0xCF)
	}
}

// BenchmarkFragmentCount measures fragment count calculation
func BenchmarkFragmentCount(b *testing.B) {
	sizes := []int{0, 100, 291, 292, 500, 1000, 5000}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, size := range sizes {
			tl.FragmentCount(size)
		}
	}
}

// BenchmarkIsMultiFragment measures multi-fragment detection
func BenchmarkIsMultiFragment(b *testing.B) {
	sizes := []int{0, 100, 291, 292, 500, 1000}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, size := range sizes {
			tl.IsMultiFragment(size)
		}
	}
}

// BenchmarkEncodedFragments measures full fragmentation + encoding
func BenchmarkEncodedFragments(b *testing.B) {
	data := make([]byte, 600)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}
	fragmenter := tl.NewFragmenter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fragments := fragmenter.Fragmentize(data)
		_ = fragments
	}
}
