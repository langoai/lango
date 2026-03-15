package bundler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeRevertReason(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give     string
		giveHex  string
		wantPart string
	}{
		{
			give:     "empty data returns empty",
			giveHex:  "",
			wantPart: "",
		},
		{
			give:     "0x only returns empty",
			giveHex:  "0x",
			wantPart: "",
		},
		{
			give: "Error(string) with simple message",
			// Error("Caller is not the owner")
			giveHex:  "0x08c379a00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001843616c6c6572206973206e6f7420746865206f776e65720000000000000000",
			wantPart: "Caller is not the owner",
		},
		{
			give: "Panic(uint256) arithmetic overflow",
			// Panic(0x11)
			giveHex:  "0x4e487b710000000000000000000000000000000000000000000000000000000000000011",
			wantPart: "arithmetic overflow",
		},
		{
			give: "Panic(uint256) division by zero",
			// Panic(0x12)
			giveHex:  "0x4e487b710000000000000000000000000000000000000000000000000000000000000012",
			wantPart: "division by zero",
		},
		{
			give:     "unknown selector returns hex",
			giveHex:  "0xdeadbeef0102030405060708",
			wantPart: "0x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			result := DecodeRevertReason(tt.giveHex)
			if tt.wantPart == "" {
				assert.Empty(t, result)
			} else {
				assert.Contains(t, result, tt.wantPart)
			}
		})
	}
}
