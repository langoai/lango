package bundler

import (
	"encoding/hex"
	"strings"
	"unicode/utf8"
)

// Standard Solidity Error(string) selector: bytes4(keccak256("Error(string)"))
const errorSelector = "08c379a0"

// Standard Solidity Panic(uint256) selector: bytes4(keccak256("Panic(uint256)"))
const panicSelector = "4e487b71"

// DecodeRevertReason attempts to decode ABI-encoded revert data into a
// human-readable string. Handles:
//   - Error(string) — standard require/revert messages
//   - Panic(uint256) — overflow, division by zero, etc.
//   - Raw hex data — returned as-is when not decodable
func DecodeRevertReason(hexData string) string {
	hexData = strings.TrimPrefix(hexData, "0x")
	if hexData == "" {
		return ""
	}

	data, err := hex.DecodeString(hexData)
	if err != nil {
		return "0x" + hexData
	}

	// Minimum: 4-byte selector + 32-byte offset + 32-byte length = 68 bytes
	if len(data) >= 68 && hex.EncodeToString(data[:4]) == errorSelector {
		return decodeErrorString(data[4:])
	}

	// Panic(uint256): 4-byte selector + 32-byte code = 36 bytes
	if len(data) >= 36 && hex.EncodeToString(data[:4]) == panicSelector {
		return decodePanicCode(data[4:])
	}

	// Return truncated hex for unknown selectors.
	if len(hexData) > 128 {
		return "0x" + hexData[:128] + "..."
	}
	return "0x" + hexData
}

// decodeErrorString decodes the ABI-encoded string payload after the selector.
func decodeErrorString(data []byte) string {
	if len(data) < 64 {
		return ""
	}

	// First 32 bytes: offset (should be 0x20 = 32)
	// Next 32 bytes: string length (big-endian uint256)
	lengthWord := data[32:64]
	length := 0
	for _, b := range lengthWord {
		length = length*256 + int(b)
	}

	if length <= 0 || 64+length > len(data) {
		return ""
	}

	msg := string(data[64 : 64+length])
	if utf8.ValidString(msg) {
		return msg
	}
	return ""
}

// decodePanicCode maps Solidity panic codes to descriptions.
func decodePanicCode(data []byte) string {
	if len(data) < 32 {
		return "panic(unknown)"
	}
	code := int(data[31]) // Low byte of 32-byte uint256

	panicCodes := map[int]string{
		0x00: "generic panic",
		0x01: "assertion failure",
		0x11: "arithmetic overflow/underflow",
		0x12: "division by zero",
		0x21: "invalid enum value",
		0x22: "storage encoding error",
		0x31: "pop on empty array",
		0x32: "array index out of bounds",
		0x41: "too much memory allocated",
		0x51: "zero-initialized function pointer",
	}

	if desc, ok := panicCodes[code]; ok {
		return "panic: " + desc
	}
	return "panic(unknown code)"
}
