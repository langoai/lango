package security

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/langoai/lango/internal/logging"
)

// tokenPattern matches {{secret:name}} and {{decrypt:id}} reference tokens.
var tokenPattern = regexp.MustCompile(`\{\{(secret|decrypt):([^}]+)\}\}`)

var refLogger = logging.SubsystemSugar("secret-ref")

// RefStore manages mapping between opaque reference tokens and secret
// plaintext values. It prevents AI agents from seeing actual secret values
// by substituting them with safe reference tokens.
type RefStore struct {
	mu      sync.RWMutex
	secrets map[string][]byte // token -> plaintext
}

// NewRefStore creates a new RefStore.
func NewRefStore() *RefStore {
	return &RefStore{
		secrets: make(map[string][]byte),
	}
}

// Store stores a secret value and returns its reference token in
// the format {{secret:name}}.
func (r *RefStore) Store(name string, value []byte) string {
	token := fmt.Sprintf("{{secret:%s}}", name)

	r.mu.Lock()
	defer r.mu.Unlock()

	stored := make([]byte, len(value))
	copy(stored, value)
	r.secrets[token] = stored

	refLogger.Debugw("stored secret reference", "name", name)
	return token
}

// StoreDecrypted stores a decrypted value and returns its reference token
// in the format {{decrypt:id}}.
func (r *RefStore) StoreDecrypted(id string, value []byte) string {
	token := fmt.Sprintf("{{decrypt:%s}}", id)

	r.mu.Lock()
	defer r.mu.Unlock()

	stored := make([]byte, len(value))
	copy(stored, value)
	r.secrets[token] = stored

	refLogger.Debugw("stored decrypted reference", "id", id)
	return token
}

// Resolve resolves a single reference token to its plaintext value.
// Returns the plaintext and true if found, or nil and false otherwise.
func (r *RefStore) Resolve(token string) ([]byte, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	val, ok := r.secrets[token]
	if !ok {
		return nil, false
	}

	// Return a copy to prevent external mutation
	result := make([]byte, len(val))
	copy(result, val)
	return result, true
}

// ResolveAll replaces all {{secret:...}} and {{decrypt:...}} tokens in
// the input string with their actual plaintext values.
func (r *RefStore) ResolveAll(input string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return tokenPattern.ReplaceAllStringFunc(input, func(token string) string {
		if val, ok := r.secrets[token]; ok {
			return string(val)
		}
		return token
	})
}

// Values returns all stored plaintext values. This is useful for output
// scanning to detect accidental secret leakage.
func (r *RefStore) Values() [][]byte {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([][]byte, 0, len(r.secrets))
	for _, val := range r.secrets {
		cp := make([]byte, len(val))
		copy(cp, val)
		result = append(result, cp)
	}
	return result
}

// Names returns a mapping of plaintext value (as string) to its reference
// name. This is used by the scanner to mask secrets in output, replacing
// them with tokens like [SECRET:name].
func (r *RefStore) Names() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]string, len(r.secrets))
	for token, val := range r.secrets {
		matches := tokenPattern.FindStringSubmatch(token)
		if len(matches) == 3 {
			result[string(val)] = matches[2]
		}
	}
	return result
}

// Clear removes all stored references.
func (r *RefStore) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.secrets = make(map[string][]byte)
	refLogger.Debugw("cleared all secret references")
}
