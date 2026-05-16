package payment

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		give    string
		wantErr bool
	}{
		{give: "0x1234567890abcdef1234567890abcdef12345678", wantErr: false},
		{give: "0xABCDEF1234567890ABCDEF1234567890ABCDEF12", wantErr: false},
		{give: "not-an-address", wantErr: true},
		{give: "0x123", wantErr: true},
		{give: "", wantErr: true},
		{give: "1234567890abcdef1234567890abcdef12345678", wantErr: true}, // missing 0x prefix
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			err := ValidateAddress(tt.give)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateAddress(%q) expected error, got nil", tt.give)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateAddress(%q) unexpected error: %v", tt.give, err)
			}
		})
	}
}

func TestERC20TransferMethodID(t *testing.T) {
	// transfer(address,uint256) keccak256 should start with 0xa9059cbb
	if len(ERC20TransferMethodID) != 4 {
		t.Fatalf("expected 4-byte selector, got %d bytes", len(ERC20TransferMethodID))
	}
	if ERC20TransferMethodID[0] != 0xa9 || ERC20TransferMethodID[1] != 0x05 ||
		ERC20TransferMethodID[2] != 0x9c || ERC20TransferMethodID[3] != 0xbb {
		t.Errorf("unexpected method ID: %x", ERC20TransferMethodID)
	}
}

func TestTxBuilderBuildTransferTx_FailsClosedWhenUnwired(t *testing.T) {
	t.Run("nil builder", func(t *testing.T) {
		var builder *TxBuilder

		tx, err := builder.BuildTransferTx(
			context.Background(),
			common.HexToAddress(validAddr),
			common.HexToAddress("0xabcdef1234567890abcdef1234567890abcdef12"),
			big.NewInt(1_000_000),
		)

		require.Error(t, err)
		assert.Nil(t, tx)
		assert.Contains(t, err.Error(), "transaction builder unavailable")
	})

	t.Run("missing rpc client", func(t *testing.T) {
		builder := &TxBuilder{
			chainID:      big.NewInt(84532),
			usdcContract: common.HexToAddress(validAddr),
		}

		tx, err := builder.BuildTransferTx(
			context.Background(),
			common.HexToAddress(validAddr),
			common.HexToAddress("0xabcdef1234567890abcdef1234567890abcdef12"),
			big.NewInt(1_000_000),
		)

		require.Error(t, err)
		assert.Nil(t, tx)
		assert.Contains(t, err.Error(), "transaction RPC unavailable")
	})
}
