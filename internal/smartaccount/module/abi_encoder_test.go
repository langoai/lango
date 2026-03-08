package module

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeInstallModule_SelectorAndLayout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give         string
		giveModType  uint8
		giveAddr     common.Address
		giveInitData []byte
		wantSelector []byte
		wantLen      int
	}{
		{
			give:         "validator with empty init data",
			giveModType:  1,
			giveAddr:     common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
			giveInitData: []byte{},
			wantSelector: crypto.Keccak256([]byte("installModule(uint256,address,bytes)"))[:4],
			// 4 (selector) + 32 (uint256) + 32 (address) + 32 (offset) + 32 (length) = 132
			wantLen: 132,
		},
		{
			give:         "executor with 5-byte init data",
			giveModType:  2,
			giveAddr:     common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
			giveInitData: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			wantSelector: crypto.Keccak256([]byte("installModule(uint256,address,bytes)"))[:4],
			// 4 + 32 + 32 + 32 (offset) + 32 (length) + 32 (data padded) = 164
			wantLen: 164,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			got, err := EncodeInstallModule(tt.giveModType, tt.giveAddr, tt.giveInitData)
			require.NoError(t, err)

			// Verify selector (first 4 bytes).
			assert.Equal(t, tt.wantSelector, got[:4], "selector mismatch")

			// Verify total length.
			assert.Equal(t, tt.wantLen, len(got), "encoded length mismatch")
		})
	}
}

func TestEncodeInstallModule_ModuleTypeByte(t *testing.T) {
	t.Parallel()

	addr := common.HexToAddress("0x1111111111111111111111111111111111111111")

	got, err := EncodeInstallModule(4, addr, nil)
	require.NoError(t, err)

	// Module type is the first ABI-encoded word (bytes 4..36), big-endian uint256.
	moduleTypeWord := got[4:36]
	moduleType := new(big.Int).SetBytes(moduleTypeWord)
	assert.Equal(t, uint64(4), moduleType.Uint64(), "module type should be 4 (hook)")
}

func TestEncodeInstallModule_AddressEncoding(t *testing.T) {
	t.Parallel()

	addr := common.HexToAddress("0xDeadBeefDeadBeefDeadBeefDeadBeefDeadBeef")

	got, err := EncodeInstallModule(1, addr, []byte{})
	require.NoError(t, err)

	// Address is the second ABI-encoded word (bytes 36..68).
	// Left-padded: 12 zero bytes + 20 address bytes.
	addrWord := got[36:68]

	// First 12 bytes must be zero.
	assert.Equal(t, make([]byte, 12), addrWord[:12], "address left-padding should be zero")

	// Last 20 bytes must match the address.
	assert.Equal(t, addr.Bytes(), addrWord[12:], "address bytes mismatch")
}

func TestEncodeInstallModule_InitDataRoundtrip(t *testing.T) {
	t.Parallel()

	initData := []byte{0xCA, 0xFE, 0xBA, 0xBE}
	addr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	got, err := EncodeInstallModule(1, addr, initData)
	require.NoError(t, err)

	// The third ABI word (bytes 68..100) is the offset to the dynamic bytes data.
	// The fourth word (bytes 100..132) is the length of the bytes data.
	lengthWord := got[100:132]
	dataLen := new(big.Int).SetBytes(lengthWord)
	assert.Equal(t, uint64(len(initData)), dataLen.Uint64(), "init data length mismatch")

	// The actual data starts at byte 132, padded to 32 bytes.
	actualData := got[132 : 132+len(initData)]
	assert.Equal(t, initData, actualData, "init data content mismatch")
}

func TestEncodeUninstallModule_SelectorAndLayout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give           string
		giveModType    uint8
		giveAddr       common.Address
		giveDeInitData []byte
		wantSelector   []byte
		wantLen        int
	}{
		{
			give:           "validator with empty deinit data",
			giveModType:    1,
			giveAddr:       common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
			giveDeInitData: []byte{},
			wantSelector:   crypto.Keccak256([]byte("uninstallModule(uint256,address,bytes)"))[:4],
			wantLen:        132,
		},
		{
			give:           "executor with 10-byte deinit data",
			giveModType:    2,
			giveAddr:       common.HexToAddress("0xCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC"),
			giveDeInitData: make([]byte, 10),
			wantSelector:   crypto.Keccak256([]byte("uninstallModule(uint256,address,bytes)"))[:4],
			wantLen:        164,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			got, err := EncodeUninstallModule(tt.giveModType, tt.giveAddr, tt.giveDeInitData)
			require.NoError(t, err)

			assert.Equal(t, tt.wantSelector, got[:4], "selector mismatch")
			assert.Equal(t, tt.wantLen, len(got), "encoded length mismatch")
		})
	}
}

func TestEncodeInstallModule_DifferentSelectorFromUninstall(t *testing.T) {
	t.Parallel()

	addr := common.HexToAddress("0x1111111111111111111111111111111111111111")

	installData, err := EncodeInstallModule(1, addr, nil)
	require.NoError(t, err)

	uninstallData, err := EncodeUninstallModule(1, addr, nil)
	require.NoError(t, err)

	installSel := hex.EncodeToString(installData[:4])
	uninstallSel := hex.EncodeToString(uninstallData[:4])

	assert.NotEqual(t, installSel, uninstallSel, "install and uninstall selectors must differ")
}

func TestEncodeInstallModule_NilInitData(t *testing.T) {
	t.Parallel()

	addr := common.HexToAddress("0x1111111111111111111111111111111111111111")

	got, err := EncodeInstallModule(1, addr, nil)
	require.NoError(t, err)

	// nil bytes should encode the same as empty bytes.
	gotEmpty, err := EncodeInstallModule(1, addr, []byte{})
	require.NoError(t, err)

	assert.Equal(t, got, gotEmpty, "nil and empty init data should produce identical encoding")
}

func TestEncodeInstallModule_LargeInitData(t *testing.T) {
	t.Parallel()

	// 64-byte init data spans two 32-byte words.
	initData := make([]byte, 64)
	for i := range initData {
		initData[i] = byte(i)
	}
	addr := common.HexToAddress("0x1111111111111111111111111111111111111111")

	got, err := EncodeInstallModule(1, addr, initData)
	require.NoError(t, err)

	// 4 + 32 + 32 + 32 (offset) + 32 (length) + 64 (data, exactly 2 words) = 196
	assert.Equal(t, 196, len(got), "encoded length for 64-byte init data")

	// Verify data bytes match.
	actualData := got[132 : 132+64]
	assert.Equal(t, initData, actualData)
}

func TestEncodeUninstallModule_ModuleTypeAndAddress(t *testing.T) {
	t.Parallel()

	addr := common.HexToAddress("0xDeadBeefDeadBeefDeadBeefDeadBeefDeadBeef")

	got, err := EncodeUninstallModule(3, addr, []byte{0xFF})
	require.NoError(t, err)

	// Module type is bytes 4..36.
	moduleType := new(big.Int).SetBytes(got[4:36])
	assert.Equal(t, uint64(3), moduleType.Uint64(), "module type should be 3 (fallback)")

	// Address is bytes 36..68.
	assert.Equal(t, addr.Bytes(), got[48:68], "address bytes mismatch")
}
