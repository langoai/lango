package smartaccount

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// ModuleType represents ERC-7579 module types.
type ModuleType uint8

const (
	ModuleTypeValidator ModuleType = 1
	ModuleTypeExecutor  ModuleType = 2
	ModuleTypeFallback  ModuleType = 3
	ModuleTypeHook      ModuleType = 4
)

// String returns the module type name.
func (t ModuleType) String() string {
	switch t {
	case ModuleTypeValidator:
		return "validator"
	case ModuleTypeExecutor:
		return "executor"
	case ModuleTypeFallback:
		return "fallback"
	case ModuleTypeHook:
		return "hook"
	default:
		return "unknown"
	}
}

// SessionKey represents a session key with its associated policy.
type SessionKey struct {
	ID            string         `json:"id"`
	PublicKey     []byte         `json:"publicKey"`
	Address       common.Address `json:"address"`
	PrivateKeyRef string         `json:"privateKeyRef"` // CryptoProvider key ID
	Policy        SessionPolicy  `json:"policy"`
	ParentID      string         `json:"parentId,omitempty"` // empty = master session
	CreatedAt     time.Time      `json:"createdAt"`
	ExpiresAt     time.Time      `json:"expiresAt"`
	Revoked       bool           `json:"revoked"`
}

// IsMaster returns true if this is a master (root) session key.
func (sk *SessionKey) IsMaster() bool { return sk.ParentID == "" }

// IsExpired returns true if the session key has expired.
func (sk *SessionKey) IsExpired() bool { return time.Now().After(sk.ExpiresAt) }

// IsActive returns true if the session key is usable.
func (sk *SessionKey) IsActive() bool { return !sk.Revoked && !sk.IsExpired() }

// SessionPolicy defines the constraints for a session key.
type SessionPolicy struct {
	AllowedTargets   []common.Address `json:"allowedTargets"`
	AllowedFunctions []string         `json:"allowedFunctions"` // 4-byte hex selectors
	SpendLimit       *big.Int         `json:"spendLimit"`
	ValidAfter       time.Time        `json:"validAfter"`
	ValidUntil       time.Time        `json:"validUntil"`
}

// ModuleInfo describes an installed ERC-7579 module.
type ModuleInfo struct {
	Address     common.Address `json:"address"`
	Type        ModuleType     `json:"type"`
	Name        string         `json:"name"`
	InstalledAt time.Time      `json:"installedAt"`
}

// UserOperation represents an ERC-4337 UserOperation.
type UserOperation struct {
	Sender               common.Address `json:"sender"`
	Nonce                *big.Int       `json:"nonce"`
	InitCode             []byte         `json:"initCode"`
	CallData             []byte         `json:"callData"`
	CallGasLimit         *big.Int       `json:"callGasLimit"`
	VerificationGasLimit *big.Int       `json:"verificationGasLimit"`
	PreVerificationGas   *big.Int       `json:"preVerificationGas"`
	MaxFeePerGas         *big.Int       `json:"maxFeePerGas"`
	MaxPriorityFeePerGas *big.Int       `json:"maxPriorityFeePerGas"`
	PaymasterAndData     []byte         `json:"paymasterAndData"`
	Signature            []byte         `json:"signature"`
}

// ContractCall represents a call to be executed via the smart account.
type ContractCall struct {
	Target      common.Address `json:"target"`
	Value       *big.Int       `json:"value"`
	Data        []byte         `json:"data"`
	FunctionSig string         `json:"functionSig,omitempty"`
}

// AccountInfo holds smart account metadata.
type AccountInfo struct {
	Address      common.Address `json:"address"`
	IsDeployed   bool           `json:"isDeployed"`
	Modules      []ModuleInfo   `json:"modules"`
	OwnerAddress common.Address `json:"ownerAddress"`
	ChainID      int64          `json:"chainId"`
	EntryPoint   common.Address `json:"entryPoint"`
}

// AccountManager defines the smart account management interface.
type AccountManager interface {
	// GetOrDeploy returns the account address, deploying if needed.
	GetOrDeploy(ctx context.Context) (*AccountInfo, error)
	// Info returns account metadata without deploying.
	Info(ctx context.Context) (*AccountInfo, error)
	// InstallModule installs an ERC-7579 module.
	InstallModule(
		ctx context.Context,
		moduleType ModuleType,
		addr common.Address,
		initData []byte,
	) (string, error)
	// UninstallModule removes an ERC-7579 module.
	UninstallModule(
		ctx context.Context,
		moduleType ModuleType,
		addr common.Address,
		deInitData []byte,
	) (string, error)
	// Execute submits a UserOperation via bundler.
	Execute(ctx context.Context, calls []ContractCall) (string, error)
}
