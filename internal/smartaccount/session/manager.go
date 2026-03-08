package session

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"

	sa "github.com/langoai/lango/internal/smartaccount"
)

// CryptoEncryptFunc encrypts private key material.
type CryptoEncryptFunc func(
	ctx context.Context, keyID string, plaintext []byte,
) ([]byte, error)

// CryptoDecryptFunc decrypts private key material.
type CryptoDecryptFunc func(
	ctx context.Context, keyID string, ciphertext []byte,
) ([]byte, error)

// RegisterOnChainFunc registers a session key on-chain.
type RegisterOnChainFunc func(
	ctx context.Context, sessionAddr common.Address, policy sa.SessionPolicy,
) (string, error)

// RevokeOnChainFunc revokes a session key on-chain.
type RevokeOnChainFunc func(
	ctx context.Context, sessionAddr common.Address,
) (string, error)

// Manager handles session key lifecycle.
type Manager struct {
	store       Store
	encrypt     CryptoEncryptFunc
	decrypt     CryptoDecryptFunc
	registerFn  RegisterOnChainFunc
	revokeFn    RevokeOnChainFunc
	maxDuration time.Duration
	maxKeys     int
	mu          sync.Mutex
}

// NewManager creates a session key manager.
func NewManager(store Store, opts ...ManagerOption) *Manager {
	m := &Manager{
		store:       store,
		maxDuration: 24 * time.Hour,
		maxKeys:     10,
	}
	for _, o := range opts {
		o.apply(m)
	}
	return m
}

// ManagerOption configures the Manager.
type ManagerOption interface {
	apply(*Manager)
}

type encryptionOption struct {
	encrypt CryptoEncryptFunc
	decrypt CryptoDecryptFunc
}

func (o encryptionOption) apply(m *Manager) {
	m.encrypt = o.encrypt
	m.decrypt = o.decrypt
}

// WithEncryption sets the encryption/decryption functions for key material.
func WithEncryption(
	encrypt CryptoEncryptFunc, decrypt CryptoDecryptFunc,
) ManagerOption {
	return encryptionOption{encrypt: encrypt, decrypt: decrypt}
}

type onChainRegistrationOption struct{ fn RegisterOnChainFunc }

func (o onChainRegistrationOption) apply(m *Manager) { m.registerFn = o.fn }

// WithOnChainRegistration sets the on-chain registration callback.
func WithOnChainRegistration(fn RegisterOnChainFunc) ManagerOption {
	return onChainRegistrationOption{fn: fn}
}

type onChainRevocationOption struct{ fn RevokeOnChainFunc }

func (o onChainRevocationOption) apply(m *Manager) { m.revokeFn = o.fn }

// WithOnChainRevocation sets the on-chain revocation callback.
func WithOnChainRevocation(fn RevokeOnChainFunc) ManagerOption {
	return onChainRevocationOption{fn: fn}
}

type maxDurationOption struct{ d time.Duration }

func (o maxDurationOption) apply(m *Manager) { m.maxDuration = o.d }

// WithMaxDuration sets the maximum allowed session duration.
func WithMaxDuration(d time.Duration) ManagerOption {
	return maxDurationOption{d: d}
}

type maxKeysOption struct{ n int }

func (o maxKeysOption) apply(m *Manager) { m.maxKeys = o.n }

// WithMaxKeys sets the maximum number of active session keys.
func WithMaxKeys(n int) ManagerOption {
	return maxKeysOption{n: n}
}

// Create creates a new session key with the given policy.
// If parentID is non-empty, creates a task session (child) scoped
// within parent bounds.
func (m *Manager) Create(
	ctx context.Context, policy sa.SessionPolicy, parentID string,
) (*sa.SessionKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate parent for task sessions.
	if parentID != "" {
		parent, err := m.store.Get(ctx, parentID)
		if err != nil {
			return nil, fmt.Errorf("get parent session: %w", err)
		}
		if !parent.IsActive() {
			if parent.Revoked {
				return nil, sa.ErrSessionRevoked
			}
			return nil, sa.ErrSessionExpired
		}
		policy = intersectPolicies(parent.Policy, policy)
	}

	// Validate duration.
	duration := policy.ValidUntil.Sub(policy.ValidAfter)
	if duration > m.maxDuration {
		return nil, fmt.Errorf(
			"session duration %v exceeds max %v: %w",
			duration, m.maxDuration, sa.ErrPolicyViolation,
		)
	}

	// Check max active keys.
	active, err := m.store.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active sessions: %w", err)
	}
	if len(active) >= m.maxKeys {
		return nil, fmt.Errorf(
			"active session limit %d reached: %w",
			m.maxKeys, sa.ErrPolicyViolation,
		)
	}

	// Generate ECDSA key pair.
	privKey, err := GenerateSessionKey()
	if err != nil {
		return nil, err
	}

	pubKeyBytes := SerializePublicKey(&privKey.PublicKey)
	addr := AddressFromPublicKey(&privKey.PublicKey)

	// Encrypt and store private key material.
	keyID := uuid.New().String()
	keyRef := keyID
	if m.encrypt != nil {
		privBytes := SerializePrivateKey(privKey)
		encrypted, encErr := m.encrypt(ctx, keyID, privBytes)
		if encErr != nil {
			return nil, fmt.Errorf("encrypt session key: %w", encErr)
		}
		keyRef = hex.EncodeToString(encrypted)
	}

	now := time.Now()
	sk := &sa.SessionKey{
		ID:            uuid.New().String(),
		PublicKey:     pubKeyBytes,
		Address:       addr,
		PrivateKeyRef: keyRef,
		Policy:        policy,
		ParentID:      parentID,
		CreatedAt:     now,
		ExpiresAt:     policy.ValidUntil,
		Revoked:       false,
	}

	if err := m.store.Save(ctx, sk); err != nil {
		return nil, fmt.Errorf("save session key: %w", err)
	}

	// Register on-chain if callback is set.
	if m.registerFn != nil {
		if _, regErr := m.registerFn(ctx, addr, policy); regErr != nil {
			return nil, fmt.Errorf("register on-chain: %w", regErr)
		}
	}

	return sk, nil
}

// Get retrieves a session key by ID.
func (m *Manager) Get(
	ctx context.Context, id string,
) (*sa.SessionKey, error) {
	return m.store.Get(ctx, id)
}

// List returns all session keys.
func (m *Manager) List(ctx context.Context) ([]*sa.SessionKey, error) {
	return m.store.List(ctx)
}

// Revoke revokes a session key and all its children.
func (m *Manager) Revoke(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key, err := m.store.Get(ctx, id)
	if err != nil {
		return err
	}

	key.Revoked = true
	if err := m.store.Save(ctx, key); err != nil {
		return fmt.Errorf("save revoked key: %w", err)
	}

	// Revoke all children recursively.
	if err := m.revokeChildren(ctx, id); err != nil {
		return fmt.Errorf("revoke children: %w", err)
	}

	// Revoke on-chain if callback is set.
	if m.revokeFn != nil {
		if _, revErr := m.revokeFn(ctx, key.Address); revErr != nil {
			return fmt.Errorf("revoke on-chain: %w", revErr)
		}
	}

	return nil
}

// revokeChildren recursively revokes all child sessions.
func (m *Manager) revokeChildren(ctx context.Context, parentID string) error {
	children, err := m.store.ListByParent(ctx, parentID)
	if err != nil {
		return err
	}
	for _, child := range children {
		if child.Revoked {
			continue
		}
		child.Revoked = true
		if err := m.store.Save(ctx, child); err != nil {
			return fmt.Errorf("save revoked child %s: %w", child.ID, err)
		}
		if err := m.revokeChildren(ctx, child.ID); err != nil {
			return err
		}
	}
	return nil
}

// RevokeAll revokes all active session keys.
func (m *Manager) RevokeAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	active, err := m.store.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("list active sessions: %w", err)
	}

	for _, key := range active {
		key.Revoked = true
		if err := m.store.Save(ctx, key); err != nil {
			return fmt.Errorf("save revoked key %s: %w", key.ID, err)
		}
		if m.revokeFn != nil {
			if _, revErr := m.revokeFn(ctx, key.Address); revErr != nil {
				return fmt.Errorf("revoke on-chain %s: %w", key.ID, revErr)
			}
		}
	}
	return nil
}

// SignUserOp signs a UserOperation with a session key.
func (m *Manager) SignUserOp(
	ctx context.Context, sessionID string, userOp *sa.UserOperation,
) ([]byte, error) {
	key, err := m.store.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if !key.IsActive() {
		if key.Revoked {
			return nil, sa.ErrSessionRevoked
		}
		return nil, sa.ErrSessionExpired
	}

	// Decrypt private key material.
	privKeyBytes := []byte(key.PrivateKeyRef)
	if m.decrypt != nil {
		ciphertext, hexErr := hex.DecodeString(key.PrivateKeyRef)
		if hexErr != nil {
			return nil, fmt.Errorf("decode encrypted key: %w", hexErr)
		}
		decrypted, decErr := m.decrypt(ctx, key.ID, ciphertext)
		if decErr != nil {
			return nil, fmt.Errorf("decrypt session key: %w", decErr)
		}
		privKeyBytes = decrypted
	}

	privKey, err := DeserializePrivateKey(privKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("restore session key: %w", err)
	}

	// Hash the UserOp fields to produce a signing digest.
	hash := hashUserOp(userOp)

	sig, err := crypto.Sign(hash, privKey)
	if err != nil {
		return nil, fmt.Errorf("sign user op: %w", err)
	}
	return sig, nil
}

// hashUserOp produces a keccak256 hash of the UserOperation fields.
func hashUserOp(op *sa.UserOperation) []byte {
	var data []byte
	data = append(data, op.Sender.Bytes()...)
	if op.Nonce != nil {
		data = append(data, op.Nonce.Bytes()...)
	}
	data = append(data, op.InitCode...)
	data = append(data, op.CallData...)
	if op.CallGasLimit != nil {
		data = append(data, op.CallGasLimit.Bytes()...)
	}
	if op.VerificationGasLimit != nil {
		data = append(data, op.VerificationGasLimit.Bytes()...)
	}
	if op.PreVerificationGas != nil {
		data = append(data, op.PreVerificationGas.Bytes()...)
	}
	if op.MaxFeePerGas != nil {
		data = append(data, op.MaxFeePerGas.Bytes()...)
	}
	if op.MaxPriorityFeePerGas != nil {
		data = append(data, op.MaxPriorityFeePerGas.Bytes()...)
	}
	data = append(data, op.PaymasterAndData...)
	return crypto.Keccak256(data)
}

// CleanupExpired removes expired session keys and returns the count removed.
func (m *Manager) CleanupExpired(ctx context.Context) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	all, err := m.store.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("list sessions: %w", err)
	}

	removed := 0
	for _, key := range all {
		if key.IsExpired() {
			if delErr := m.store.Delete(ctx, key.ID); delErr != nil {
				return removed, fmt.Errorf(
					"delete expired key %s: %w", key.ID, delErr,
				)
			}
			removed++
		}
	}
	return removed, nil
}

// intersectPolicies produces a policy that is the intersection
// (tighter bounds) of parent and child policies.
func intersectPolicies(
	parent, child sa.SessionPolicy,
) sa.SessionPolicy {
	result := child

	// ValidAfter: use the later of the two.
	if parent.ValidAfter.After(child.ValidAfter) {
		result.ValidAfter = parent.ValidAfter
	}

	// ValidUntil: use the earlier of the two.
	if parent.ValidUntil.Before(child.ValidUntil) {
		result.ValidUntil = parent.ValidUntil
	}

	// SpendLimit: use the smaller of the two.
	if parent.SpendLimit != nil && child.SpendLimit != nil {
		if parent.SpendLimit.Cmp(child.SpendLimit) < 0 {
			result.SpendLimit = new(big.Int).Set(parent.SpendLimit)
		}
	} else if parent.SpendLimit != nil {
		result.SpendLimit = new(big.Int).Set(parent.SpendLimit)
	}

	// AllowedTargets: intersection of address lists.
	if len(parent.AllowedTargets) > 0 {
		if len(child.AllowedTargets) > 0 {
			result.AllowedTargets = intersectAddresses(
				parent.AllowedTargets, child.AllowedTargets,
			)
		} else {
			targets := make([]common.Address, len(parent.AllowedTargets))
			copy(targets, parent.AllowedTargets)
			result.AllowedTargets = targets
		}
	}

	// AllowedFunctions: intersection of function selectors.
	if len(parent.AllowedFunctions) > 0 {
		if len(child.AllowedFunctions) > 0 {
			result.AllowedFunctions = intersectStrings(
				parent.AllowedFunctions, child.AllowedFunctions,
			)
		} else {
			fns := make([]string, len(parent.AllowedFunctions))
			copy(fns, parent.AllowedFunctions)
			result.AllowedFunctions = fns
		}
	}

	return result
}

// intersectAddresses returns addresses present in both slices.
func intersectAddresses(
	a, b []common.Address,
) []common.Address {
	set := make(map[common.Address]struct{}, len(a))
	for _, addr := range a {
		set[addr] = struct{}{}
	}
	var result []common.Address
	for _, addr := range b {
		if _, ok := set[addr]; ok {
			result = append(result, addr)
		}
	}
	return result
}

// intersectStrings returns strings present in both slices.
func intersectStrings(a, b []string) []string {
	set := make(map[string]struct{}, len(a))
	for _, s := range a {
		set[s] = struct{}{}
	}
	var result []string
	for _, s := range b {
		if _, ok := set[s]; ok {
			result = append(result, s)
		}
	}
	return result
}
