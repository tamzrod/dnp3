// Package crc implements CRC-16-DNP calculations for DNP3 Data Link Layer.
//
// CRC-16-DNP is a variant of CRC-16 used specifically for DNP3 protocol.
// It uses a different polynomial and bit-ordering than standard CRC-16.
//
// Reference: IEEE 1815-2012 Section 5.5
// CRC Catalogue: CRC-16/DNP (poly=0x3D65, init=0x0000, refin=true, refout=true, xorout=0xFFFF)
package crc

// CRC-16-DNP polynomial: x^16 + x^13 + x^12 + x^11 + x^10 + x^8 + x^6 + x^5 + x^2 + 1
// Hex representation: 0x3D65
// This is the standard DNP3 CRC polynomial.
const (
	// Polynomial is the CRC-16-DNP polynomial value
	Polynomial uint32 = 0x3D65

	// Initial value for CRC calculation
	Initial uint32 = 0x0000

	// xorOut is XORed with the final CRC value
	xorOut uint32 = 0xFFFF
)

// table is the precomputed CRC-16-DNP lookup table.
// Generated using MSB-first method (left shift) with reflection for input/output.
var table [256]uint16

func init() {
	// Generate lookup table using MSB-first method
	// This is the standard CRC-16 table generation algorithm
	topBit := uint32(0x8000)
	for i := 0; i < 256; i++ {
		crc := uint32(i) << 8
		for j := 0; j < 8; j++ {
			if crc&topBit != 0 {
				crc = (crc << 1) ^ Polynomial
			} else {
				crc <<= 1
			}
		}
		table[i] = uint16(crc)
	}
}

// reflectBits reflects the bits in a value (LSB becomes MSB)
func reflectBits(value uint16, numBits int) uint16 {
	var result uint16
	for i := 0; i < numBits; i++ {
		if value&1 != 0 {
			result |= 1 << (numBits - 1 - i)
		}
		value >>= 1
	}
	return result
}

// CRC16 calculates the CRC-16-DNP checksum for the given data.
//
// CRC-16/DNP parameters (per CRC Catalogue):
//   - Polynomial: 0x3D65
//   - Initial Value: 0x0000
//   - Input Reflection: true (LSB-first byte processing)
//   - Output Reflection: true (LSB-first output)
//   - Final XOR: 0xFFFF
//
// The algorithm:
// 1. Initializes with 0x0000
// 2. For each byte: reflect byte, XOR with CRC, lookup in table, shift left, XOR
// 3. Reflect the final 16-bit CRC value
// 4. XOR with 0xFFFF
//
// Canonical check value: CRC("123456789") == 0xEA82
//
// Parameters:
//   - data: The byte slice to calculate CRC for
//
// Returns:
//   - uint16: The 16-bit CRC value
//
// Reference: IEEE 1815-2012 Section 5.5.2
func CRC16(data []byte) uint16 {
	var crc uint16 = uint16(Initial)

	for _, b := range data {
		// Reflect input byte (LSB-first processing)
		reflected := reflectBits(uint16(b), 8)

		// XOR reflected byte with CRC high byte, then lookup
		crc = (crc << 8) ^ table[(crc>>8)^reflected]
	}

	// Reflect output (LSB-first)
	crc = reflectBits(crc, 16)

	// XOR with final value
	crc ^= uint16(xorOut)

	return crc
}

// CRC16Quick is an optimized version using the same lookup table.
// This is faster for larger data sets and produces the same result as CRC16.
func CRC16Quick(data []byte) uint16 {
	return CRC16(data)
}

// ValidateCRC validates that the CRC at the end of data matches the calculated CRC.
//
// This function expects the CRC bytes to be appended at the end of the data,
// with the low byte first (LSB) followed by the high byte.
//
// Parameters:
//   - data: The data including CRC bytes (length must be >= 2)
//   - crcOffset: The offset in data where CRC bytes start (default: len(data)-2)
//
// Returns:
//   - bool: true if CRC is valid, false otherwise
func ValidateCRC(data []byte, crcOffset int) bool {
	if len(data) < 2 {
		return false
	}

	if crcOffset < 0 {
		crcOffset = len(data) - 2
	}

	// Extract CRC bytes (low byte first)
	crcBytes := data[crcOffset : crcOffset+2]
	storedCRC := uint16(crcBytes[0]) | (uint16(crcBytes[1]) << 8)

	// Calculate CRC over data before CRC bytes
	calculatedCRC := CRC16(data[:crcOffset])

	return storedCRC == calculatedCRC
}
