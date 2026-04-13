package escrow

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
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/wallet"
)

// Compile-time check.
var _ SettlementExecutor = (*USDCSettler)(nil)

// USDCSettlerOption configures a USDCSettler.
type USDCSettlerOption func(*USDCSettler)

// WithReceiptTimeout sets the maximum wait for on-chain confirmation.
func WithReceiptTimeout(d time.Duration) USDCSettlerOption {
	return func(s *USDCSettler) {
		if d > 0 {
			s.receiptTimeout = d
		}
	}
}

// WithMaxRetries sets the maximum transaction submission attempts.
func WithMaxRetries(n int) USDCSettlerOption {
	return func(s *USDCSettler) {
		if n > 0 {
			s.maxRetries = n
		}
	}
}

// WithLogger sets a structured logger.
func WithLogger(l *zap.SugaredLogger) USDCSettlerOption {
	return func(s *USDCSettler) {
		if l != nil {
			s.logger = l
		}
	}
}

// USDCSettler implements SettlementExecutor using on-chain USDC transfers.
// Lock verifies balance sufficiency (custodian model — funds held in agent wallet).
// Release transfers USDC from agent wallet to seller.
// Refund transfers USDC from agent wallet to buyer.
type USDCSettler struct {
	wallet    wallet.WalletProvider
	txBuilder *payment.TxBuilder
	rpc       *ethclient.Client
	chainID   *big.Int
	resolver  AddressResolver // v1+v2 DID → Ethereum address (nil = v1 only fallback)

	receiptTimeout time.Duration
	maxRetries     int
	logger         *zap.SugaredLogger

	// nonceMu serializes transaction building to avoid nonce collisions.
	nonceMu sync.Mutex
}

// WithAddressResolver sets the address resolver for v1+v2 DID support.
func WithAddressResolver(r AddressResolver) USDCSettlerOption {
	return func(s *USDCSettler) {
		if r != nil {
			s.resolver = r
		}
	}
}

// NewUSDCSettler creates a USDC settler with the given dependencies and options.
func NewUSDCSettler(w wallet.WalletProvider, txb *payment.TxBuilder, rpc *ethclient.Client, chainID int64, opts ...USDCSettlerOption) *USDCSettler {
	s := &USDCSettler{
		wallet:         w,
		txBuilder:      txb,
		rpc:            rpc,
		chainID:        big.NewInt(chainID),
		receiptTimeout: 2 * time.Minute,
		maxRetries:     3,
		logger:         zap.NewNop().Sugar(),
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Lock verifies that the agent wallet holds sufficient USDC for the escrow.
// In the custodian model, actual fund transfer is external (e.g. EIP-3009);
// this method only validates balance sufficiency.
func (s *USDCSettler) Lock(ctx context.Context, buyerDID string, amount *big.Int) error {
	addr, err := s.agentAddress(ctx)
	if err != nil {
		return err
	}

	balance, err := s.queryUSDCBalance(ctx, addr)
	if err != nil {
		return fmt.Errorf("query USDC balance: %w", err)
	}

	if balance.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient USDC balance: have %s, need %s", balance.String(), amount.String())
	}

	s.logger.Infow("escrow lock verified",
		"buyerDID", buyerDID, "amount", amount.String(), "balance", balance.String())
	return nil
}

// Release transfers USDC from the agent wallet to the seller.
func (s *USDCSettler) Release(ctx context.Context, sellerDID string, amount *big.Int) error {
	to, err := s.resolveAddress(sellerDID)
	if err != nil {
		return fmt.Errorf("resolve seller address: %w", err)
	}
	return s.transferFromAgent(ctx, to, amount, "release", sellerDID)
}

// Refund transfers USDC from the agent wallet back to the buyer.
func (s *USDCSettler) Refund(ctx context.Context, buyerDID string, amount *big.Int) error {
	to, err := s.resolveAddress(buyerDID)
	if err != nil {
		return fmt.Errorf("resolve buyer address: %w", err)
	}
	return s.transferFromAgent(ctx, to, amount, "refund", buyerDID)
}

// resolveAddress uses the injected resolver if available, otherwise falls back
// to the package-level v1-only function.
func (s *USDCSettler) resolveAddress(did string) (common.Address, error) {
	if s.resolver != nil {
		return s.resolver.ResolveAddress(did)
	}
	return ResolveAddress(did)
}

// transferFromAgent builds, signs, submits, and confirms a USDC transfer
// from the agent wallet to the given address.
func (s *USDCSettler) transferFromAgent(ctx context.Context, to common.Address, amount *big.Int, op, peerDID string) error {
	s.nonceMu.Lock()
	defer s.nonceMu.Unlock()

	from, err := s.agentAddress(ctx)
	if err != nil {
		return err
	}

	tx, err := s.txBuilder.BuildTransferTx(ctx, from, to, amount)
	if err != nil {
		return fmt.Errorf("build %s tx: %w", op, err)
	}

	signedTx, err := s.signTx(ctx, tx)
	if err != nil {
		return fmt.Errorf("sign %s tx: %w", op, err)
	}

	txHash, err := s.submitWithRetry(ctx, signedTx)
	if err != nil {
		return fmt.Errorf("submit %s tx: %w", op, err)
	}

	s.logger.Infow("escrow "+op+" tx submitted",
		"txHash", txHash, "to", to.Hex(), "peerDID", peerDID, "amount", amount.String())

	if err := s.waitForConfirmation(ctx, common.HexToHash(txHash)); err != nil {
		return fmt.Errorf("confirm %s tx: %w", op, err)
	}

	s.logger.Infow("escrow "+op+" confirmed", "txHash", txHash)
	return nil
}

// agentAddress returns the agent wallet's Ethereum address.
func (s *USDCSettler) agentAddress(ctx context.Context) (common.Address, error) {
	addrStr, err := s.wallet.Address(ctx)
	if err != nil {
		return common.Address{}, fmt.Errorf("get agent wallet address: %w", err)
	}
	return common.HexToAddress(addrStr), nil
}

// queryUSDCBalance calls balanceOf on the USDC contract for the given address.
func (s *USDCSettler) queryUSDCBalance(ctx context.Context, addr common.Address) (*big.Int, error) {
	contract := s.txBuilder.USDCContract()
	data := make([]byte, 4+32)
	copy(data[:4], payment.BalanceOfSelector)
	copy(data[4+12:4+32], addr.Bytes())

	result, err := s.rpc.CallContract(ctx, ethereum.CallMsg{
		To:   &contract,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetBytes(result), nil
}

// signTx signs an unsigned transaction using the wallet provider.
func (s *USDCSettler) signTx(ctx context.Context, tx *types.Transaction) (*types.Transaction, error) {
	signer := types.LatestSignerForChainID(s.chainID)
	txHash := signer.Hash(tx)

	sig, err := s.wallet.SignTransaction(ctx, txHash.Bytes())
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}

	return tx.WithSignature(signer, sig)
}

// submitWithRetry sends the signed transaction with exponential backoff.
func (s *USDCSettler) submitWithRetry(ctx context.Context, tx *types.Transaction) (string, error) {
	var lastErr error
	for attempt := 0; attempt < s.maxRetries; attempt++ {
		if err := s.rpc.SendTransaction(ctx, tx); err == nil {
			return tx.Hash().Hex(), nil
		} else {
			lastErr = err
		}

		s.logger.Warnw("escrow tx submission retry",
			"attempt", attempt+1, "error", lastErr)

		backoff := time.Duration(1<<uint(attempt)) * time.Second
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(backoff):
		}
	}
	return "", fmt.Errorf("submit after %d retries: %w", s.maxRetries, lastErr)
}

// waitForConfirmation polls for transaction receipt with exponential backoff.
func (s *USDCSettler) waitForConfirmation(ctx context.Context, txHash common.Hash) error {
	deadline := time.After(s.receiptTimeout)
	backoff := 1 * time.Second
	maxBackoff := 16 * time.Second

	for {
		receipt, err := s.rpc.TransactionReceipt(ctx, txHash)
		if err == nil {
			if receipt.Status == types.ReceiptStatusSuccessful {
				return nil
			}
			return fmt.Errorf("tx reverted: status=%d", receipt.Status)
		}

		select {
		case <-deadline:
			return fmt.Errorf("receipt timeout after %v", s.receiptTimeout)
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}

		if backoff < maxBackoff {
			backoff *= 2
		}
	}
}
