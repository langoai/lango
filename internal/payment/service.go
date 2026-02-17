package payment

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"

	"github.com/langowarny/lango/internal/ent"
	"github.com/langowarny/lango/internal/ent/paymenttx"
	"github.com/langowarny/lango/internal/wallet"
)

// Service orchestrates blockchain payment operations.
type Service struct {
	wallet    wallet.WalletProvider
	limiter   wallet.SpendingLimiter
	builder   *TxBuilder
	client    *ent.Client
	rpcClient *ethclient.Client
	chainID   int64
}

// NewService creates a payment service.
func NewService(
	wp wallet.WalletProvider,
	limiter wallet.SpendingLimiter,
	builder *TxBuilder,
	client *ent.Client,
	rpcClient *ethclient.Client,
	chainID int64,
) *Service {
	return &Service{
		wallet:    wp,
		limiter:   limiter,
		builder:   builder,
		client:    client,
		rpcClient: rpcClient,
		chainID:   chainID,
	}
}

// Send executes a payment: limit check → build tx → sign → submit → record.
func (s *Service) Send(ctx context.Context, req PaymentRequest) (*PaymentReceipt, error) {
	// Validate recipient address
	if err := ValidateAddress(req.To); err != nil {
		return nil, fmt.Errorf("invalid recipient: %w", err)
	}

	// Parse amount
	amount, err := wallet.ParseUSDC(req.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}
	if amount.Sign() <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	// Check spending limits
	if err := s.limiter.Check(ctx, amount); err != nil {
		return nil, fmt.Errorf("spending limit: %w", err)
	}

	// Get sender address
	fromAddr, err := s.wallet.Address(ctx)
	if err != nil {
		return nil, fmt.Errorf("get wallet address: %w", err)
	}

	// Create pending transaction record
	ptx, err := s.client.PaymentTx.Create().
		SetFromAddress(fromAddr).
		SetToAddress(req.To).
		SetAmount(req.Amount).
		SetChainID(s.chainID).
		SetStatus(paymenttx.StatusPending).
		SetNillableSessionKey(nilIfEmpty(req.SessionKey)).
		SetNillablePurpose(nilIfEmpty(req.Purpose)).
		SetNillableX402URL(nilIfEmpty(req.X402URL)).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create tx record: %w", err)
	}

	// Build transaction
	from := common.HexToAddress(fromAddr)
	to := common.HexToAddress(req.To)
	tx, err := s.builder.BuildTransferTx(ctx, from, to, amount)
	if err != nil {
		s.failTx(ctx, ptx.ID, err)
		return nil, fmt.Errorf("build transaction: %w", err)
	}

	// Sign transaction
	signer := types.LatestSignerForChainID(big.NewInt(s.chainID))
	txHash := signer.Hash(tx)
	sig, err := s.wallet.SignTransaction(ctx, txHash.Bytes())
	if err != nil {
		s.failTx(ctx, ptx.ID, err)
		return nil, fmt.Errorf("sign transaction: %w", err)
	}

	signedTx, err := tx.WithSignature(signer, sig)
	if err != nil {
		s.failTx(ctx, ptx.ID, err)
		return nil, fmt.Errorf("apply signature: %w", err)
	}

	// Submit transaction
	if err := s.rpcClient.SendTransaction(ctx, signedTx); err != nil {
		s.failTx(ctx, ptx.ID, err)
		return nil, fmt.Errorf("submit transaction: %w", err)
	}

	// Update record with tx hash
	txHashHex := signedTx.Hash().Hex()
	s.client.PaymentTx.UpdateOneID(ptx.ID).
		SetTxHash(txHashHex).
		SetStatus(paymenttx.StatusSubmitted).
		SaveX(ctx)

	// Record spending
	if err := s.limiter.Record(ctx, amount); err != nil {
		// Non-fatal: tx already submitted
	}

	return &PaymentReceipt{
		TxHash:    txHashHex,
		Status:    string(paymenttx.StatusSubmitted),
		Amount:    req.Amount,
		From:      fromAddr,
		To:        req.To,
		ChainID:   s.chainID,
		Timestamp: time.Now(),
	}, nil
}

// Balance returns the wallet's USDC balance as a formatted string.
func (s *Service) Balance(ctx context.Context) (string, error) {
	// Query USDC ERC-20 balance via eth_call
	addr, err := s.wallet.Address(ctx)
	if err != nil {
		return "", fmt.Errorf("get address: %w", err)
	}

	contract := s.builder.USDCContract()
	balanceOfSelector := []byte{0x70, 0xa0, 0x82, 0x31} // balanceOf(address)
	data := make([]byte, 4+32)
	copy(data[:4], balanceOfSelector)
	addrBytes := common.HexToAddress(addr)
	copy(data[4+12:4+32], addrBytes.Bytes())

	result, err := s.rpcClient.CallContract(ctx, ethereum.CallMsg{
		To:   &contract,
		Data: data,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("query USDC balance: %w", err)
	}

	balance := new(big.Int).SetBytes(result)
	return wallet.FormatUSDC(balance), nil
}

// History returns recent payment transactions.
func (s *Service) History(ctx context.Context, limit int) ([]TransactionInfo, error) {
	if limit <= 0 {
		limit = 20
	}

	txs, err := s.client.PaymentTx.Query().
		Order(ent.Desc(paymenttx.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}

	result := make([]TransactionInfo, len(txs))
	for i, tx := range txs {
		result[i] = TransactionInfo{
			TxHash:       tx.TxHash,
			Status:       string(tx.Status),
			Amount:       tx.Amount,
			From:         tx.FromAddress,
			To:           tx.ToAddress,
			ChainID:      tx.ChainID,
			Purpose:      tx.Purpose,
			X402URL:      tx.X402URL,
			ErrorMessage: tx.ErrorMessage,
			CreatedAt:    tx.CreatedAt,
		}
	}

	return result, nil
}

// HandleX402 processes an X402 payment challenge.
func (s *Service) HandleX402(ctx context.Context, challenge X402Challenge) (string, error) {
	receipt, err := s.Send(ctx, PaymentRequest{
		To:      challenge.RecipientAddress,
		Amount:  challenge.Amount,
		Purpose: "X402 payment for " + challenge.PaymentURL,
		X402URL: challenge.PaymentURL,
	})
	if err != nil {
		return "", fmt.Errorf("x402 payment: %w", err)
	}
	return receipt.TxHash, nil
}

// WalletAddress returns the wallet's public address.
func (s *Service) WalletAddress(ctx context.Context) (string, error) {
	return s.wallet.Address(ctx)
}

// ChainID returns the configured chain ID.
func (s *Service) ChainID() int64 {
	return s.chainID
}

// failTx marks a transaction as failed with an error message.
func (s *Service) failTx(ctx context.Context, id uuid.UUID, txErr error) {
	s.client.PaymentTx.UpdateOneID(id).
		SetStatus(paymenttx.StatusFailed).
		SetErrorMessage(txErr.Error()).
		SaveX(ctx)
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
