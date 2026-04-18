package payment

import (
	"context"
	"fmt"
	"github.com/langoai/lango/internal/logging"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"

	"github.com/langoai/lango/internal/ent/paymenttx"
	"github.com/langoai/lango/internal/wallet"
)

// DefaultHistoryLimit is the default number of transactions returned by History.
const DefaultHistoryLimit = 20

const purposeX402AutoPayment = "X402 auto-payment"

// DefaultReceiptTimeout is the maximum time to wait for on-chain confirmation.
const DefaultReceiptTimeout = 2 * time.Minute

// DefaultMaxRetries is the default number of transaction submission attempts.
const DefaultMaxRetries = 3

// Service orchestrates blockchain payment operations.
type Service struct {
	wallet    wallet.WalletProvider
	limiter   wallet.SpendingLimiter
	builder   *TxBuilder
	store     TxStore
	rpcClient *ethclient.Client
	chainID   int64

	// nonceMu serializes transaction building to prevent nonce collisions.
	nonceMu        sync.Mutex
	receiptTimeout time.Duration
	maxRetries     int
}

// NewService creates a payment service.
func NewService(
	wp wallet.WalletProvider,
	limiter wallet.SpendingLimiter,
	builder *TxBuilder,
	store TxStore,
	rpcClient *ethclient.Client,
	chainID int64,
) *Service {
	return &Service{
		wallet:         wp,
		limiter:        limiter,
		builder:        builder,
		store:          store,
		rpcClient:      rpcClient,
		chainID:        chainID,
		receiptTimeout: DefaultReceiptTimeout,
		maxRetries:     DefaultMaxRetries,
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
	ptx, err := s.store.Create(ctx, TxRecord{
		ID:            uuid.New(),
		FromAddress:   fromAddr,
		ToAddress:     req.To,
		Amount:        req.Amount,
		ChainID:       s.chainID,
		Status:        paymenttx.StatusPending,
		SessionKey:    req.SessionKey,
		Purpose:       req.Purpose,
		X402URL:       req.X402URL,
		PaymentMethod: paymenttx.PaymentMethodDirectTransfer,
	})
	if err != nil {
		return nil, fmt.Errorf("create tx record: %w", err)
	}

	// Build, sign, and submit under nonce lock to prevent collisions.
	s.nonceMu.Lock()

	from := common.HexToAddress(fromAddr)
	to := common.HexToAddress(req.To)
	tx, err := s.builder.BuildTransferTx(ctx, from, to, amount)
	if err != nil {
		s.nonceMu.Unlock()
		s.failTx(ctx, ptx.ID, err)
		return nil, fmt.Errorf("build transaction: %w", err)
	}

	signer := types.LatestSignerForChainID(big.NewInt(s.chainID))
	txSigHash := signer.Hash(tx)
	sig, err := s.wallet.SignTransaction(ctx, txSigHash.Bytes())
	if err != nil {
		s.nonceMu.Unlock()
		s.failTx(ctx, ptx.ID, err)
		return nil, fmt.Errorf("sign transaction: %w", err)
	}

	signedTx, err := tx.WithSignature(signer, sig)
	if err != nil {
		s.nonceMu.Unlock()
		s.failTx(ctx, ptx.ID, err)
		return nil, fmt.Errorf("apply signature: %w", err)
	}

	// Submit with retry (exponential backoff).
	txHashHex, err := s.submitWithRetry(ctx, signedTx)
	s.nonceMu.Unlock()
	if err != nil {
		s.failTx(ctx, ptx.ID, err)
		return nil, fmt.Errorf("submit transaction: %w", err)
	}

	// Update record to submitted.
	if err := s.store.UpdateStatus(ctx, ptx.ID, paymenttx.StatusSubmitted, txHashHex, ""); err != nil {
		return nil, fmt.Errorf("mark submitted: %w", err)
	}

	// Wait for on-chain confirmation.
	receipt, err := s.waitForConfirmation(ctx, signedTx.Hash())
	if err != nil {
		s.failTx(ctx, ptx.ID, err)
		return nil, fmt.Errorf("confirm transaction: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		txErr := fmt.Errorf("tx %s reverted (status=%d)", txHashHex, receipt.Status)
		s.failTx(ctx, ptx.ID, txErr)
		return nil, txErr
	}

	// Update record to confirmed.
	if err := s.store.UpdateStatus(ctx, ptx.ID, paymenttx.StatusConfirmed, "", ""); err != nil {
		return nil, fmt.Errorf("mark confirmed: %w", err)
	}

	// Record spending — non-fatal since tx is already confirmed.
	_ = s.limiter.Record(ctx, amount)

	return &PaymentReceipt{
		TxHash:      txHashHex,
		Status:      string(paymenttx.StatusConfirmed),
		Amount:      req.Amount,
		From:        fromAddr,
		To:          req.To,
		ChainID:     s.chainID,
		GasUsed:     receipt.GasUsed,
		BlockNumber: receipt.BlockNumber.Uint64(),
		Timestamp:   time.Now(),
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
	data := make([]byte, 4+32)
	copy(data[:4], BalanceOfSelector)
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
		limit = DefaultHistoryLimit
	}

	txs, err := s.store.List(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}

	result := make([]TransactionInfo, len(txs))
	for i, tx := range txs {
		result[i] = TransactionInfo{
			TxHash:        tx.TxHash,
			Status:        string(tx.Status),
			Amount:        tx.Amount,
			From:          tx.FromAddress,
			To:            tx.ToAddress,
			ChainID:       tx.ChainID,
			Purpose:       tx.Purpose,
			X402URL:       tx.X402URL,
			PaymentMethod: string(tx.PaymentMethod),
			ErrorMessage:  tx.ErrorMessage,
			CreatedAt:     tx.CreatedAt,
		}
	}

	return result, nil
}

// RecordX402Payment records an X402 automatic payment for audit trail.
// Unlike Send(), this does not build or submit a transaction — the SDK handles
// payment signing. This only creates the database record for tracking.
func (s *Service) RecordX402Payment(ctx context.Context, record X402PaymentRecord) error {
	_, err := s.store.Create(ctx, TxRecord{
		ID:            uuid.New(),
		FromAddress:   record.From,
		ToAddress:     record.To,
		Amount:        record.Amount,
		ChainID:       record.ChainID,
		Status:        paymenttx.StatusSubmitted,
		Purpose:       purposeX402AutoPayment,
		X402URL:       record.URL,
		PaymentMethod: paymenttx.PaymentMethodX402V2,
	})
	if err != nil {
		return fmt.Errorf("record X402 payment: %w", err)
	}
	return nil
}

// WalletAddress returns the wallet's public address.
func (s *Service) WalletAddress(ctx context.Context) (string, error) {
	return s.wallet.Address(ctx)
}

// ChainID returns the configured chain ID.
func (s *Service) ChainID() int64 {
	return s.chainID
}

// submitWithRetry sends the signed transaction with exponential backoff.
func (s *Service) submitWithRetry(ctx context.Context, tx *types.Transaction) (string, error) {
	var lastErr error
	for attempt := 0; attempt < s.maxRetries; attempt++ {
		if err := s.rpcClient.SendTransaction(ctx, tx); err == nil {
			return tx.Hash().Hex(), nil
		} else {
			lastErr = err
		}

		logging.SubsystemSugar("payment").Warnw("tx submission retry", "attempt", attempt+1, "error", lastErr)

		backoff := time.Duration(1<<uint(attempt)) * time.Second
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(backoff):
		}
	}
	return "", fmt.Errorf("submit after %d retries: %w", s.maxRetries, lastErr)
}

// waitForConfirmation polls for a transaction receipt with exponential backoff.
func (s *Service) waitForConfirmation(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	deadline := time.After(s.receiptTimeout)
	backoff := 1 * time.Second
	maxBackoff := 16 * time.Second

	for {
		receipt, err := s.rpcClient.TransactionReceipt(ctx, txHash)
		if err == nil {
			return receipt, nil
		}

		select {
		case <-deadline:
			return nil, fmt.Errorf("receipt timeout for %s after %v", txHash.Hex(), s.receiptTimeout)
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}

		if backoff < maxBackoff {
			backoff *= 2
		}
	}
}

// failTx marks a transaction as failed with an error message.
func (s *Service) failTx(ctx context.Context, id uuid.UUID, txErr error) {
	_ = s.store.UpdateStatus(ctx, id, paymenttx.StatusFailed, "", txErr.Error())
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
