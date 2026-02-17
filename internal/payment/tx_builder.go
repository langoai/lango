package payment

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ERC20TransferMethodID is the function selector for transfer(address,uint256).
var ERC20TransferMethodID = crypto.Keccak256([]byte("transfer(address,uint256)"))[:4]

// TxBuilder constructs ERC-20 transfer transactions.
type TxBuilder struct {
	client       *ethclient.Client
	chainID      *big.Int
	usdcContract common.Address
}

// NewTxBuilder creates a transaction builder for the given chain.
func NewTxBuilder(client *ethclient.Client, chainID int64, usdcContract string) *TxBuilder {
	return &TxBuilder{
		client:       client,
		chainID:      big.NewInt(chainID),
		usdcContract: common.HexToAddress(usdcContract),
	}
}

// BuildTransferTx constructs an EIP-1559 ERC-20 transfer transaction.
// Returns the sighash (transaction hash to sign) and the unsigned transaction.
func (b *TxBuilder) BuildTransferTx(ctx context.Context, from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	// Encode ERC-20 transfer(address,uint256) calldata
	data := b.encodeTransferData(to, amount)

	// Get nonce
	nonce, err := b.client.PendingNonceAt(ctx, from)
	if err != nil {
		return nil, fmt.Errorf("get nonce: %w", err)
	}

	// Estimate gas
	gasLimit, err := b.client.EstimateGas(ctx, ethereum.CallMsg{
		From: from,
		To:   &b.usdcContract,
		Data: data,
	})
	if err != nil {
		return nil, fmt.Errorf("estimate gas: %w", err)
	}

	// Get gas price parameters (EIP-1559)
	header, err := b.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("get block header: %w", err)
	}

	baseFee := header.BaseFee
	if baseFee == nil {
		baseFee = big.NewInt(1_000_000_000) // 1 gwei fallback
	}

	// maxPriorityFee = 1.5 gwei, maxFee = 2 * baseFee + maxPriorityFee
	maxPriorityFee := big.NewInt(1_500_000_000)
	maxFee := new(big.Int).Add(
		new(big.Int).Mul(baseFee, big.NewInt(2)),
		maxPriorityFee,
	)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   b.chainID,
		Nonce:     nonce,
		GasFeeCap: maxFee,
		GasTipCap: maxPriorityFee,
		Gas:       gasLimit,
		To:        &b.usdcContract,
		Value:     big.NewInt(0),
		Data:      data,
	})

	return tx, nil
}

// encodeTransferData encodes the transfer(address,uint256) call data.
func (b *TxBuilder) encodeTransferData(to common.Address, amount *big.Int) []byte {
	// Method selector (4 bytes) + address (32 bytes, left-padded) + amount (32 bytes, left-padded)
	data := make([]byte, 4+32+32)
	copy(data[:4], ERC20TransferMethodID)
	copy(data[4+12:4+32], to.Bytes())
	amount.FillBytes(data[4+32 : 4+64])
	return data
}

// USDCContract returns the configured USDC contract address.
func (b *TxBuilder) USDCContract() common.Address {
	return b.usdcContract
}

// ValidateAddress checks if a string is a valid Ethereum address.
func ValidateAddress(addr string) error {
	if !strings.HasPrefix(addr, "0x") || len(addr) != 42 {
		return fmt.Errorf("invalid address format: %q", addr)
	}
	if !common.IsHexAddress(addr) {
		return fmt.Errorf("invalid hex address: %q", addr)
	}
	return nil
}
