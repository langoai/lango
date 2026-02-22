package crypto

import (
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/security"
)

var logger = logging.SubsystemSugar("tool.crypto")

// Tool provides cryptographic operations for AI agents.
// Decrypted values are returned as opaque reference tokens; the plaintext
// never enters the agent context.
type Tool struct {
	crypto   security.CryptoProvider
	registry *security.KeyRegistry
	refs     *security.RefStore
	scanner  *agent.SecretScanner
}

// New creates a new crypto tool.
// If scanner is non-nil, decrypted values are registered for output scanning.
func New(crypto security.CryptoProvider, registry *security.KeyRegistry, refs *security.RefStore, scanner *agent.SecretScanner) *Tool {
	return &Tool{
		crypto:   crypto,
		registry: registry,
		refs:     refs,
		scanner:  scanner,
	}
}

// EncryptParams are the parameters for the encrypt operation.
type EncryptParams struct {
	Data  string `json:"data"`
	KeyID string `json:"keyId,omitempty"`
}

// DecryptParams are the parameters for the decrypt operation.
type DecryptParams struct {
	Ciphertext string `json:"ciphertext"`
	KeyID      string `json:"keyId,omitempty"`
}

// SignParams are the parameters for the sign operation.
type SignParams struct {
	Data  string `json:"data"`
	KeyID string `json:"keyId,omitempty"`
}

// HashParams are the parameters for the hash operation.
type HashParams struct {
	Data      string `json:"data"`
	Algorithm string `json:"algorithm,omitempty"` // sha256, sha512
}

// KeyEntry represents a key in the list response.
type KeyEntry struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	CreatedAt  string  `json:"createdAt"`
	LastUsedAt *string `json:"lastUsedAt,omitempty"`
}

// Encrypt encrypts data using the specified key.
func (t *Tool) Encrypt(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var p EncryptParams
	if err := mapToStruct(params, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if p.Data == "" {
		return nil, fmt.Errorf("data is required")
	}

	keyID := p.KeyID
	if keyID == "" {
		// Use default key
		keyInfo, err := t.registry.GetDefaultKey(ctx)
		if err != nil {
			return nil, fmt.Errorf("no default key available: %w", err)
		}
		keyID = keyInfo.RemoteKeyID
	}

	logger.Infow("encrypting data", "keyId", keyID)

	ciphertext, err := t.crypto.Encrypt(ctx, keyID, []byte(p.Data))
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	return map[string]interface{}{
		"success":    true,
		"ciphertext": base64.StdEncoding.EncodeToString(ciphertext),
	}, nil
}

// Decrypt decrypts data and returns an opaque reference token.
// The plaintext is stored in the RefStore and resolved at execution time.
func (t *Tool) Decrypt(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var p DecryptParams
	if err := mapToStruct(params, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if p.Ciphertext == "" {
		return nil, fmt.Errorf("ciphertext is required")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(p.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("invalid ciphertext encoding: %w", err)
	}

	keyID := p.KeyID
	if keyID == "" {
		keyInfo, err := t.registry.GetDefaultKey(ctx)
		if err != nil {
			return nil, fmt.Errorf("no default key available: %w", err)
		}
		keyID = keyInfo.RemoteKeyID
	}

	logger.Infow("decrypting data", "keyId", keyID)

	plaintext, err := t.crypto.Decrypt(ctx, keyID, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	// Store plaintext in RefStore and return opaque reference token.
	refID := uuid.New().String()
	ref := t.refs.StoreDecrypted(refID, plaintext)

	// Register with scanner for output-side secret detection.
	if t.scanner != nil {
		t.scanner.Register("decrypt:"+refID, plaintext)
	}

	return map[string]interface{}{
		"success": true,
		"data":    ref,
		"note":    "This is a reference token. Use it directly in exec commands — it will be resolved at execution time.",
	}, nil
}

// Sign generates a signature for the data.
func (t *Tool) Sign(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var p SignParams
	if err := mapToStruct(params, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if p.Data == "" {
		return nil, fmt.Errorf("data is required")
	}

	keyID := p.KeyID
	if keyID == "" {
		keyID = "local"
	}

	logger.Infow("signing data", "keyId", keyID)

	signature, err := t.crypto.Sign(ctx, keyID, []byte(p.Data))
	if err != nil {
		return nil, fmt.Errorf("signing failed: %w", err)
	}

	return map[string]interface{}{
		"success":   true,
		"signature": base64.StdEncoding.EncodeToString(signature),
	}, nil
}

// Hash computes a hash of the data.
func (t *Tool) Hash(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var p HashParams
	if err := mapToStruct(params, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if p.Data == "" {
		return nil, fmt.Errorf("data is required")
	}

	algorithm := p.Algorithm
	if algorithm == "" {
		algorithm = "sha256"
	}

	var h hash.Hash
	switch algorithm {
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s (supported: sha256, sha512)", algorithm)
	}

	h.Write([]byte(p.Data))
	hashBytes := h.Sum(nil)

	return map[string]interface{}{
		"success":   true,
		"hash":      hex.EncodeToString(hashBytes),
		"algorithm": algorithm,
	}, nil
}

// Keys lists available keys.
func (t *Tool) Keys(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	keys, err := t.registry.ListKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("list keys: %w", err)
	}

	entries := make([]KeyEntry, len(keys))
	for i, k := range keys {
		entry := KeyEntry{
			ID:        k.ID.String(),
			Name:      k.Name,
			Type:      string(k.Type),
			CreatedAt: k.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if k.LastUsedAt != nil {
			formatted := k.LastUsedAt.Format("2006-01-02T15:04:05Z")
			entry.LastUsedAt = &formatted
		}
		entries[i] = entry
	}

	return map[string]interface{}{
		"keys":  entries,
		"count": len(entries),
	}, nil
}

func mapToStruct(m map[string]interface{}, v interface{}) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
