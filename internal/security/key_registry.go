package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/langowarny/lango/internal/ent"
	"github.com/langowarny/lango/internal/ent/key"
)

// KeyType represents the purpose of a key.
type KeyType string

const (
	KeyTypeEncryption KeyType = "encryption"
	KeyTypeSigning    KeyType = "signing"
)

// Valid reports whether t is a known key type.
func (t KeyType) Valid() bool {
	switch t {
	case KeyTypeEncryption, KeyTypeSigning:
		return true
	}
	return false
}

// Values returns all known key types.
func (t KeyType) Values() []KeyType {
	return []KeyType{KeyTypeEncryption, KeyTypeSigning}
}

// KeyInfo represents key metadata.
type KeyInfo struct {
	ID          uuid.UUID
	Name        string
	RemoteKeyID string
	Type        KeyType
	CreatedAt   time.Time
	LastUsedAt  *time.Time
}

// KeyRegistry manages encryption/signing keys.
type KeyRegistry struct {
	mu     sync.RWMutex
	client *ent.Client
}

// NewKeyRegistry creates a new KeyRegistry.
func NewKeyRegistry(client *ent.Client) *KeyRegistry {
	return &KeyRegistry{
		client: client,
	}
}

// RegisterKey registers a new key.
func (r *KeyRegistry) RegisterKey(ctx context.Context, name, remoteKeyID string, keyType KeyType) (*KeyInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if key with this name already exists
	existing, err := r.client.Key.Query().Where(key.NameEQ(name)).Only(ctx)
	if err == nil {
		// Update existing key
		updated, err := existing.Update().
			SetRemoteKeyID(remoteKeyID).
			SetType(key.Type(keyType)).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("update key: %w", err)
		}
		return toKeyInfo(updated), nil
	}

	// Create new key
	created, err := r.client.Key.Create().
		SetName(name).
		SetRemoteKeyID(remoteKeyID).
		SetType(key.Type(keyType)).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create key: %w", err)
	}

	return toKeyInfo(created), nil
}

// GetKey retrieves a key by name.
func (r *KeyRegistry) GetKey(ctx context.Context, name string) (*KeyInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	k, err := r.client.Key.Query().Where(key.NameEQ(name)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("get key %q: %w", name, ErrKeyNotFound)
		}
		return nil, fmt.Errorf("get key %q: %w", name, err)
	}

	return toKeyInfo(k), nil
}

// GetDefaultKey retrieves the default encryption key (most recently created).
func (r *KeyRegistry) GetDefaultKey(ctx context.Context) (*KeyInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	k, err := r.client.Key.Query().
		Where(key.TypeEQ(key.TypeEncryption)).
		Order(ent.Desc(key.FieldCreatedAt)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNoEncryptionKeys
		}
		return nil, fmt.Errorf("get default key: %w", err)
	}

	return toKeyInfo(k), nil
}

// ListKeys returns all registered keys.
func (r *KeyRegistry) ListKeys(ctx context.Context) ([]*KeyInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys, err := r.client.Key.Query().Order(ent.Desc(key.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list keys: %w", err)
	}

	result := make([]*KeyInfo, len(keys))
	for i, k := range keys {
		result[i] = toKeyInfo(k)
	}

	return result, nil
}

// UpdateLastUsed updates the last used timestamp for a key.
func (r *KeyRegistry) UpdateLastUsed(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.client.Key.Update().
		Where(key.NameEQ(name)).
		SetLastUsedAt(time.Now()).
		Save(ctx)
	return err
}

// DeleteKey removes a key by name.
func (r *KeyRegistry) DeleteKey(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.client.Key.Delete().Where(key.NameEQ(name)).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete key: %w", err)
	}
	return nil
}

func toKeyInfo(k *ent.Key) *KeyInfo {
	return &KeyInfo{
		ID:          k.ID,
		Name:        k.Name,
		RemoteKeyID: k.RemoteKeyID,
		Type:        KeyType(k.Type),
		CreatedAt:   k.CreatedAt,
		LastUsedAt:  k.LastUsedAt,
	}
}
