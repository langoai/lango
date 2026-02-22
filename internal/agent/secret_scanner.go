package agent

import (
	"strings"
	"sync"

	"github.com/langoai/lango/internal/logging"
)

var secretLogger = logging.SubsystemSugar("secret-scanner")

// SecretScanner scans text output for known secret values and replaces them
// with masked placeholders. This prevents AI agents from leaking secret values
// in their responses.
type SecretScanner struct {
	mu      sync.RWMutex
	secrets map[string]string // plaintext value -> secret name
}

// NewSecretScanner creates a new SecretScanner with an empty secret registry.
func NewSecretScanner() *SecretScanner {
	return &SecretScanner{
		secrets: make(map[string]string),
	}
}

// Register adds a known secret value with its name. Values shorter than 4
// characters are ignored to avoid false positives during scanning.
func (s *SecretScanner) Register(name string, value []byte) {
	plaintext := string(value)
	if len(plaintext) < 4 {
		secretLogger.Debugw("ignoring short secret value",
			"name", name, "length", len(plaintext))
		return
	}

	s.mu.Lock()
	s.secrets[plaintext] = name
	s.mu.Unlock()

	secretLogger.Infow("registered secret", "name", name)
}

// Scan replaces any known secret values found in text with [SECRET:name]
// placeholders.
func (s *SecretScanner) Scan(text string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := text
	for plaintext, name := range s.secrets {
		result = strings.ReplaceAll(result, plaintext, "[SECRET:"+name+"]")
	}
	return result
}

// HasSecrets returns true if any secrets are registered.
func (s *SecretScanner) HasSecrets() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.secrets) > 0
}

// Clear removes all registered secrets.
func (s *SecretScanner) Clear() {
	s.mu.Lock()
	s.secrets = make(map[string]string)
	s.mu.Unlock()

	secretLogger.Infow("cleared all secrets")
}
