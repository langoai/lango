package wallet

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/langowarny/lango/internal/security"
)

// LocalWallet implements WalletProvider using a locally-stored encrypted private key.
// The key is loaded from SecretsStore, used for signing, then immediately zeroed.
type LocalWallet struct {
	secrets  *security.SecretsStore
	rpcURL   string
	chainID  int64
	keyName  string

	mu     sync.Mutex
	client *ethclient.Client
}

// NewLocalWallet creates a wallet that loads its private key from the secrets store.
func NewLocalWallet(secrets *security.SecretsStore, rpcURL string, chainID int64) *LocalWallet {
	return &LocalWallet{
		secrets: secrets,
		rpcURL:  rpcURL,
		chainID: chainID,
		keyName: "wallet.privatekey",
	}
}

// Address derives the wallet address from the stored private key.
func (w *LocalWallet) Address(ctx context.Context) (string, error) {
	keyBytes, err := w.secrets.Get(ctx, w.keyName)
	if err != nil {
		return "", fmt.Errorf("load wallet key: %w", err)
	}
	defer zeroBytes(keyBytes)

	privateKey, err := crypto.ToECDSA(keyBytes)
	if err != nil {
		return "", fmt.Errorf("parse wallet key: %w", err)
	}

	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	return addr.Hex(), nil
}

// Balance returns the native token (ETH) balance of the wallet.
func (w *LocalWallet) Balance(ctx context.Context) (*big.Int, error) {
	client, err := w.getClient()
	if err != nil {
		return nil, err
	}

	addr, err := w.Address(ctx)
	if err != nil {
		return nil, err
	}

	balance, err := client.BalanceAt(ctx, common.HexToAddress(addr), nil)
	if err != nil {
		return nil, fmt.Errorf("query balance: %w", err)
	}

	return balance, nil
}

// SignTransaction signs a raw transaction hash with the wallet's private key.
func (w *LocalWallet) SignTransaction(ctx context.Context, rawTx []byte) ([]byte, error) {
	keyBytes, err := w.secrets.Get(ctx, w.keyName)
	if err != nil {
		return nil, fmt.Errorf("load wallet key: %w", err)
	}
	defer zeroBytes(keyBytes)

	privateKey, err := crypto.ToECDSA(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse wallet key: %w", err)
	}

	sig, err := crypto.Sign(rawTx, privateKey)
	if err != nil {
		return nil, fmt.Errorf("sign transaction: %w", err)
	}

	return sig, nil
}

// SignMessage signs an arbitrary message with the wallet's private key.
func (w *LocalWallet) SignMessage(ctx context.Context, message []byte) ([]byte, error) {
	keyBytes, err := w.secrets.Get(ctx, w.keyName)
	if err != nil {
		return nil, fmt.Errorf("load wallet key: %w", err)
	}
	defer zeroBytes(keyBytes)

	privateKey, err := crypto.ToECDSA(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse wallet key: %w", err)
	}

	hash := crypto.Keccak256(message)
	sig, err := crypto.Sign(hash, privateKey)
	if err != nil {
		return nil, fmt.Errorf("sign message: %w", err)
	}

	return sig, nil
}

// getClient returns a cached ethclient, creating one on first call.
func (w *LocalWallet) getClient() (*ethclient.Client, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.client != nil {
		return w.client, nil
	}

	client, err := ethclient.Dial(w.rpcURL)
	if err != nil {
		return nil, fmt.Errorf("connect to RPC %q: %w", w.rpcURL, err)
	}

	w.client = client
	return client, nil
}

// zeroBytes overwrites a byte slice with zeros.
func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
