package security

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/secret"
	"github.com/langoai/lango/internal/logging"
)

var secretsLogger = logging.SubsystemSugar("secrets-store")

// SecretInfo represents secret metadata (without the actual value).
type SecretInfo struct {
	ID          uuid.UUID
	Name        string
	KeyID       uuid.UUID
	KeyName     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	AccessCount int
}

// SecretsStore manages encrypted secrets.
type SecretsStore struct {
	client   *ent.Client
	registry *KeyRegistry
	crypto   CryptoProvider
}

// NewSecretsStore creates a new SecretsStore.
func NewSecretsStore(client *ent.Client, registry *KeyRegistry, crypto CryptoProvider) *SecretsStore {
	return &SecretsStore{
		client:   client,
		registry: registry,
		crypto:   crypto,
	}
}

// Store encrypts and stores a secret value.
func (s *SecretsStore) Store(ctx context.Context, name string, value []byte) error {
	// Get default key for encryption
	keyInfo, err := s.registry.GetDefaultKey(ctx)
	if err != nil {
		return fmt.Errorf("no encryption key available: %w", err)
	}

	// Encrypt the value
	encrypted, err := s.crypto.Encrypt(ctx, keyInfo.RemoteKeyID, value)
	if err != nil {
		return fmt.Errorf("encrypt secret: %w", err)
	}

	// Get the key entity
	keyEntity, err := s.client.Key.Get(ctx, keyInfo.ID)
	if err != nil {
		return fmt.Errorf("get key entity: %w", err)
	}

	// Check if secret exists
	existing, err := s.client.Secret.Query().Where(secret.NameEQ(name)).Only(ctx)
	if err == nil {
		// Update existing secret
		_, err = existing.Update().
			SetEncryptedValue(encrypted).
			SetKey(keyEntity).
			SetAccessCount(0). // Reset access count on update
			Save(ctx)
		if err != nil {
			return fmt.Errorf("update secret: %w", err)
		}
		return nil
	}

	// Create new secret
	_, err = s.client.Secret.Create().
		SetName(name).
		SetEncryptedValue(encrypted).
		SetKey(keyEntity).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create secret: %w", err)
	}

	return nil
}

// Get retrieves and decrypts a secret value.
func (s *SecretsStore) Get(ctx context.Context, name string) ([]byte, error) {
	// Query secret with key edge
	sec, err := s.client.Secret.Query().
		Where(secret.NameEQ(name)).
		WithKey().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("secret not found: %s", name)
		}
		return nil, fmt.Errorf("get secret: %w", err)
	}

	// Get key info for decryption
	keyEntity := sec.Edges.Key
	if keyEntity == nil {
		return nil, fmt.Errorf("secret has no associated key")
	}

	// Decrypt the value
	decrypted, err := s.crypto.Decrypt(ctx, keyEntity.RemoteKeyID, sec.EncryptedValue)
	if err != nil {
		return nil, fmt.Errorf("decrypt secret: %w", err)
	}

	// Increment access count
	if _, err := sec.Update().SetAccessCount(sec.AccessCount + 1).Save(ctx); err != nil {
		secretsLogger.Warnw("update access count", "name", sec.Name, "error", err)
	}

	// Update key last used
	if err := s.registry.UpdateLastUsed(ctx, keyEntity.Name); err != nil {
		secretsLogger.Warnw("update key last used", "key", keyEntity.Name, "error", err)
	}

	return decrypted, nil
}

// List returns metadata for all secrets.
func (s *SecretsStore) List(ctx context.Context) ([]*SecretInfo, error) {
	secrets, err := s.client.Secret.Query().WithKey().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}

	result := make([]*SecretInfo, len(secrets))
	for i, sec := range secrets {
		info := &SecretInfo{
			ID:          sec.ID,
			Name:        sec.Name,
			CreatedAt:   sec.CreatedAt,
			UpdatedAt:   sec.UpdatedAt,
			AccessCount: sec.AccessCount,
		}
		if sec.Edges.Key != nil {
			info.KeyID = sec.Edges.Key.ID
			info.KeyName = sec.Edges.Key.Name
		}
		result[i] = info
	}

	return result, nil
}

// Delete removes a secret by name.
func (s *SecretsStore) Delete(ctx context.Context, name string) error {
	deleted, err := s.client.Secret.Delete().Where(secret.NameEQ(name)).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete secret: %w", err)
	}
	if deleted == 0 {
		return fmt.Errorf("secret not found: %s", name)
	}
	return nil
}
