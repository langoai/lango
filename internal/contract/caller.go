package contract

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/wallet"
)

// DefaultTimeout is the default context timeout for contract calls.
const DefaultTimeout = 30 * time.Second

// MaxRetries is the default number of retry attempts for transaction submission.
const MaxRetries = 3

// Caller provides read and write access to smart contracts.
type Caller struct {
	rpc        *ethclient.Client
	wallet     wallet.WalletProvider
	chainID    *big.Int
	cache      *ABICache
	nonceMu    sync.Mutex
	timeout    time.Duration
	maxRetries int
}

// NewCaller creates a contract caller.
func NewCaller(rpc *ethclient.Client, wp wallet.WalletProvider, chainID int64, cache *ABICache) *Caller {
	return &Caller{
		rpc:        rpc,
		wallet:     wp,
		chainID:    big.NewInt(chainID),
		cache:      cache,
		timeout:    DefaultTimeout,
		maxRetries: MaxRetries,
	}
}

// Read calls a view/pure function on a contract (no tx, no gas).
func (c *Caller) Read(ctx context.Context, req ContractCallRequest) (*ContractCallResult, error) {
	parsed, err := c.cache.GetOrParse(req.ChainID, req.Address, req.ABI)
	if err != nil {
		return nil, err
	}

	method, ok := parsed.Methods[req.Method]
	if !ok {
		return nil, fmt.Errorf("method %q not found in ABI", req.Method)
	}

	data, err := parsed.Pack(req.Method, req.Args...)
	if err != nil {
		return nil, fmt.Errorf("pack args for %q: %w", req.Method, err)
	}

	addr := req.Address
	result, err := c.rpc.CallContract(ctx, ethereum.CallMsg{
		To:   &addr,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("call contract %s.%s: %w", addr.Hex(), req.Method, err)
	}

	outputs, err := method.Outputs.Unpack(result)
	if err != nil {
		return nil, fmt.Errorf("unpack %q result: %w", req.Method, err)
	}

	return &ContractCallResult{Data: outputs}, nil
}

// Write sends a state-changing transaction to a contract.
func (c *Caller) Write(ctx context.Context, req ContractCallRequest) (*ContractCallResult, error) {
	parsed, err := c.cache.GetOrParse(req.ChainID, req.Address, req.ABI)
	if err != nil {
		return nil, err
	}

	if _, ok := parsed.Methods[req.Method]; !ok {
		return nil, fmt.Errorf("method %q not found in ABI", req.Method)
	}

	data, err := parsed.Pack(req.Method, req.Args...)
	if err != nil {
		return nil, fmt.Errorf("pack args for %q: %w", req.Method, err)
	}

	fromAddr, err := c.wallet.Address(ctx)
	if err != nil {
		return nil, fmt.Errorf("get wallet address: %w", err)
	}
	from := common.HexToAddress(fromAddr)
	to := req.Address

	// Get nonce under lock to prevent nonce collisions.
	c.nonceMu.Lock()
	nonce, err := c.rpc.PendingNonceAt(ctx, from)
	if err != nil {
		c.nonceMu.Unlock()
		return nil, fmt.Errorf("get nonce: %w", err)
	}
	c.nonceMu.Unlock()

	// Estimate gas.
	value := req.Value
	if value == nil {
		value = new(big.Int)
	}
	gasLimit, err := c.rpc.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    &to,
		Data:  data,
		Value: value,
	})
	if err != nil {
		return nil, fmt.Errorf("estimate gas: %w", err)
	}

	// EIP-1559 gas fee parameters (same pattern as payment/tx_builder.go).
	header, err := c.rpc.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("get block header: %w", err)
	}
	baseFee := header.BaseFee
	if baseFee == nil {
		baseFee = big.NewInt(payment.DefaultBaseFeeWei)
	}
	maxPriorityFee := big.NewInt(payment.DefaultMaxPriorityFeeWei)
	maxFee := new(big.Int).Add(
		new(big.Int).Mul(baseFee, big.NewInt(payment.BaseFeeMultiplier)),
		maxPriorityFee,
	)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   c.chainID,
		Nonce:     nonce,
		GasFeeCap: maxFee,
		GasTipCap: maxPriorityFee,
		Gas:       gasLimit,
		To:        &to,
		Value:     value,
		Data:      data,
	})

	// Sign via wallet.
	signer := types.LatestSignerForChainID(c.chainID)
	txHash := signer.Hash(tx)
	sig, err := c.wallet.SignTransaction(ctx, txHash.Bytes())
	if err != nil {
		return nil, fmt.Errorf("sign transaction: %w", err)
	}
	signedTx, err := tx.WithSignature(signer, sig)
	if err != nil {
		return nil, fmt.Errorf("apply signature: %w", err)
	}

	// Submit with retry.
	var submitErr error
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		submitErr = c.rpc.SendTransaction(ctx, signedTx)
		if submitErr == nil {
			break
		}
		if attempt < c.maxRetries-1 {
			time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
		}
	}
	if submitErr != nil {
		return nil, fmt.Errorf("submit transaction: %w", submitErr)
	}

	// Wait for receipt.
	receipt, err := c.waitForReceipt(ctx, signedTx.Hash())
	if err != nil {
		return &ContractCallResult{
			TxHash: signedTx.Hash().Hex(),
		}, nil // tx submitted but receipt unavailable
	}

	return &ContractCallResult{
		TxHash:  signedTx.Hash().Hex(),
		GasUsed: receipt.GasUsed,
	}, nil
}

// LoadABI parses and caches an ABI for later use.
func (c *Caller) LoadABI(chainID int64, address common.Address, abiJSON string) error {
	_, err := c.cache.GetOrParse(chainID, address, abiJSON)
	return err
}

// waitForReceipt polls for a transaction receipt with exponential backoff.
func (c *Caller) waitForReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	deadline := time.After(c.timeout)
	delay := 1 * time.Second

	for {
		receipt, err := c.rpc.TransactionReceipt(ctx, txHash)
		if err == nil {
			return receipt, nil
		}

		select {
		case <-deadline:
			return nil, fmt.Errorf("receipt timeout for %s", txHash.Hex())
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			delay = delay * 2
			if delay > 8*time.Second {
				delay = 8 * time.Second
			}
		}
	}
}
