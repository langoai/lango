package paymaster

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildApproveCalldata_Selector(t *testing.T) {
	t.Parallel()

	spender := common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	amount := big.NewInt(1000)

	got := BuildApproveCalldata(spender, amount)

	wantSelector := crypto.Keccak256([]byte("approve(address,uint256)"))[:4]
	assert.Equal(t, wantSelector, got[:4], "first 4 bytes must be ERC-20 approve selector")
}

func TestBuildApproveCalldata_Layout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give       string
		giveAddr   common.Address
		giveAmount *big.Int
		wantLen    int
	}{
		{
			give:       "normal amount",
			giveAddr:   common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
			giveAmount: big.NewInt(500_000_000),
			wantLen:    68,
		},
		{
			give:       "zero amount",
			giveAddr:   common.HexToAddress("0x1111111111111111111111111111111111111111"),
			giveAmount: big.NewInt(0),
			wantLen:    68,
		},
		{
			give:       "nil amount",
			giveAddr:   common.HexToAddress("0x2222222222222222222222222222222222222222"),
			giveAmount: nil,
			wantLen:    68,
		},
		{
			give:       "max uint256",
			giveAddr:   common.HexToAddress("0x3333333333333333333333333333333333333333"),
			giveAmount: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)),
			wantLen:    68,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			got := BuildApproveCalldata(tt.giveAddr, tt.giveAmount)
			assert.Equal(t, tt.wantLen, len(got), "calldata must always be 68 bytes (4+32+32)")
		})
	}
}

func TestBuildApproveCalldata_SpenderEncoding(t *testing.T) {
	t.Parallel()

	spender := common.HexToAddress("0xDeadBeefDeadBeefDeadBeefDeadBeefDeadBeef")
	amount := big.NewInt(42)

	got := BuildApproveCalldata(spender, amount)

	// Spender occupies bytes 4..36, left-padded with 12 zero bytes.
	spenderWord := got[4:36]
	assert.Equal(t, make([]byte, 12), spenderWord[:12], "spender left-padding should be zero")
	assert.Equal(t, spender.Bytes(), spenderWord[12:], "spender address bytes mismatch")
}

func TestBuildApproveCalldata_AmountEncoding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give       string
		giveAmount *big.Int
		wantBytes  []byte
	}{
		{
			give:       "amount 1 USDC (6 decimals)",
			giveAmount: big.NewInt(1_000_000),
			wantBytes:  big.NewInt(1_000_000).Bytes(),
		},
		{
			give:       "amount zero",
			giveAmount: big.NewInt(0),
			wantBytes:  nil, // zero produces empty Bytes()
		},
		{
			give:       "nil amount",
			giveAmount: nil,
			wantBytes:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			spender := common.HexToAddress("0x1111111111111111111111111111111111111111")
			got := BuildApproveCalldata(spender, tt.giveAmount)

			// Amount occupies bytes 36..68, big-endian left-padded.
			amountWord := got[36:68]

			if len(tt.wantBytes) == 0 {
				// All zeros expected.
				assert.Equal(t, make([]byte, 32), amountWord, "amount should be all zeros")
			} else {
				// Verify the last N bytes match, leading bytes are zero.
				n := len(tt.wantBytes)
				assert.Equal(t, make([]byte, 32-n), amountWord[:32-n], "amount left-padding should be zero")
				assert.Equal(t, tt.wantBytes, amountWord[32-n:], "amount bytes mismatch")
			}
		})
	}
}

func TestNewApprovalCall_Fields(t *testing.T) {
	t.Parallel()

	tokenAddr := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	paymasterAddr := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")
	amount := big.NewInt(10_000_000)

	call := NewApprovalCall(tokenAddr, paymasterAddr, amount)

	require.NotNil(t, call)
	assert.Equal(t, tokenAddr, call.TokenAddress)
	assert.Equal(t, paymasterAddr, call.PaymasterAddr)
	assert.Equal(t, amount, call.Amount)

	// ApproveCalldata should equal BuildApproveCalldata with paymaster as spender.
	wantCalldata := BuildApproveCalldata(paymasterAddr, amount)
	assert.Equal(t, wantCalldata, call.ApproveCalldata)
}

func TestNewApprovalCall_CalldataUsesPaymasterAsSpender(t *testing.T) {
	t.Parallel()

	tokenAddr := common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	paymasterAddr := common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	amount := big.NewInt(999)

	call := NewApprovalCall(tokenAddr, paymasterAddr, amount)

	// Verify the spender in calldata is the paymaster address, not the token address.
	spenderWord := call.ApproveCalldata[4:36]
	assert.Equal(t, paymasterAddr.Bytes(), spenderWord[12:],
		"spender in calldata should be paymaster address")
}

func TestBuildApproveCalldata_Deterministic(t *testing.T) {
	t.Parallel()

	spender := common.HexToAddress("0xCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC")
	amount := big.NewInt(12345)

	got1 := BuildApproveCalldata(spender, amount)
	got2 := BuildApproveCalldata(spender, amount)

	assert.Equal(t, got1, got2, "identical inputs must produce identical calldata")
}
