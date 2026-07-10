// Package crc implements CRC-16-DNP calculations for DNP3 Data Link Layer.
//
// CRC-16-DNP is a variant of CRC-16 used specifically for DNP3 protocol.
// It uses a different polynomial and bit-ordering than standard CRC-16.
//
// Reference: IEEE 1815-2012 Section 5.5
package crc

import "hash/crc32"

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
// CRC-16-DNP uses LSB-first (reflected) representation.
var table = crc32.MakeTable(Polynomial)

// CRC16 calculates the CRC-16-DNP checksum for the given data.
//
// The CRC-16-DNP algorithm:
// 1. Initializes with 0x0000 (not 0xFFFF like standard CRC-16)
// 2. Processes data LSB-first (reflected)
// 3. XORs the final result with 0xFFFF
//
// Parameters:
//   - data: The byte slice to calculate CRC for
//
// Returns:
//   - uint16: The 16-bit CRC value
//
// Reference: IEEE 1815-2012 Section 5.5.2
func CRC16(data []byte) uint16 {
	crc := Initial

	for _, b := range data {
		// XOR data byte with current CRC
		crc ^= uint32(b)

		// Process each bit (LSB first - this is what makes it "reflected")
		for i := 0; i < 8; i++ {
			if crc&1 == 1 {
				// If LSB is 1, shift and XOR with polynomial
				crc = (crc >> 1) ^ Polynomial
			} else {
				// If LSB is 0, just shift
				crc >>= 1
			}
		}
	}

	// XOR with final value
	crc ^= xorOut

	return uint16(crc)
}

// CRC16Quick is an optimized version using a lookup table.
// This may be faster for larger data sets but produces the same result.
func CRC16Quick(data []byte) uint16 {
	crc := crc32.Update(Initial^xorOut, table, data)
	return uint16(crc ^ xorOut)
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
