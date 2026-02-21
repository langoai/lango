package secrets

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/langowarny/lango/internal/agent"
	"github.com/langowarny/lango/internal/logging"
	"github.com/langowarny/lango/internal/security"
)

var logger = logging.SubsystemSugar("tool.secrets")

// Tool provides secure secrets management for AI agents.
// Secret values are never returned as plaintext; instead, opaque reference
// tokens are returned that are resolved at execution time by the exec tool.
type Tool struct {
	store   *security.SecretsStore
	refs    *security.RefStore
	scanner *agent.SecretScanner
}

// New creates a new secrets tool.
// If scanner is non-nil, retrieved secrets are registered for output scanning.
func New(store *security.SecretsStore, refs *security.RefStore, scanner *agent.SecretScanner) *Tool {
	return &Tool{store: store, refs: refs, scanner: scanner}
}

// StoreParams are the parameters for the store operation.
type StoreParams struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// GetParams are the parameters for the get operation.
type GetParams struct {
	Name string `json:"name"`
}

// DeleteParams are the parameters for the delete operation.
type DeleteParams struct {
	Name string `json:"name"`
}

// SecretEntry represents a secret in the list response.
type SecretEntry struct {
	Name        string `json:"name"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	AccessCount int    `json:"accessCount"`
}

// ListResult is the result of the list operation.
type ListResult struct {
	Secrets []SecretEntry `json:"secrets"`
	Count   int           `json:"count"`
}

// Store encrypts and stores a secret value.
func (t *Tool) Store(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var p StoreParams
	if err := mapToStruct(params, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if p.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if p.Value == "" {
		return nil, fmt.Errorf("value is required")
	}

	logger.Infow("storing secret", "name", p.Name)

	if err := t.store.Store(ctx, p.Name, []byte(p.Value)); err != nil {
		return nil, fmt.Errorf("store secret: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Secret '%s' stored successfully", p.Name),
	}, nil
}

// Get retrieves a secret and returns an opaque reference token.
// The plaintext value is stored in the RefStore and resolved at execution time.
// The agent never sees the actual secret value.
func (t *Tool) Get(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var p GetParams
	if err := mapToStruct(params, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if p.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	logger.Infow("retrieving secret reference", "name", p.Name)

	value, err := t.store.Get(ctx, p.Name)
	if err != nil {
		return nil, err
	}

	// Store plaintext in RefStore and return opaque reference token.
	ref := t.refs.Store(p.Name, value)

	// Register with scanner for output-side secret detection.
	if t.scanner != nil {
		t.scanner.Register(p.Name, value)
	}

	return map[string]interface{}{
		"success": true,
		"name":    p.Name,
		"value":   ref,
		"note":    "This is a reference token. Use it directly in exec commands — it will be resolved at execution time.",
	}, nil
}

// List returns metadata for all stored secrets.
func (t *Tool) List(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	secrets, err := t.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}

	entries := make([]SecretEntry, len(secrets))
	for i, s := range secrets {
		entries[i] = SecretEntry{
			Name:        s.Name,
			CreatedAt:   s.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			AccessCount: s.AccessCount,
		}
	}

	return ListResult{
		Secrets: entries,
		Count:   len(entries),
	}, nil
}

// Delete removes a secret by name.
func (t *Tool) Delete(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var p DeleteParams
	if err := mapToStruct(params, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if p.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	logger.Infow("deleting secret", "name", p.Name)

	if err := t.store.Delete(ctx, p.Name); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Secret '%s' deleted successfully", p.Name),
	}, nil
}

func mapToStruct(m map[string]interface{}, v interface{}) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
