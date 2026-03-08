package app

import (
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/security"
	toolcrypto "github.com/langoai/lango/internal/tools/crypto"
	toolsecrets "github.com/langoai/lango/internal/tools/secrets"
)

// buildCryptoTools wraps crypto.Tool methods as agent tools.
func buildCryptoTools(crypto security.CryptoProvider, keys *security.KeyRegistry, refs *security.RefStore, scanner *agent.SecretScanner) []*agent.Tool {
	ct := toolcrypto.New(crypto, keys, refs, scanner)
	return []*agent.Tool{
		{
			Name:        "crypto_encrypt",
			Description: "Encrypt data using a registered key",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"data":  map[string]interface{}{"type": "string", "description": "The data to encrypt"},
					"keyId": map[string]interface{}{"type": "string", "description": "Key ID to use (default: default key)"},
				},
				"required": []string{"data"},
			},
			Handler: ct.Encrypt,
		},
		{
			Name:        "crypto_decrypt",
			Description: "Decrypt data using a registered key. Returns an opaque {{decrypt:id}} reference token. The decrypted value never enters the agent context.",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"ciphertext": map[string]interface{}{"type": "string", "description": "Base64-encoded ciphertext to decrypt"},
					"keyId":      map[string]interface{}{"type": "string", "description": "Key ID to use (default: default key)"},
				},
				"required": []string{"ciphertext"},
			},
			Handler: ct.Decrypt,
		},
		{
			Name:        "crypto_sign",
			Description: "Generate a digital signature for data",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"data":  map[string]interface{}{"type": "string", "description": "The data to sign"},
					"keyId": map[string]interface{}{"type": "string", "description": "Key ID to use"},
				},
				"required": []string{"data"},
			},
			Handler: ct.Sign,
		},
		{
			Name:        "crypto_hash",
			Description: "Compute a cryptographic hash of data",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"data":      map[string]interface{}{"type": "string", "description": "The data to hash"},
					"algorithm": map[string]interface{}{"type": "string", "description": "Hash algorithm: sha256 or sha512", "enum": []string{"sha256", "sha512"}},
				},
				"required": []string{"data"},
			},
			Handler: ct.Hash,
		},
		{
			Name:        "crypto_keys",
			Description: "List all registered cryptographic keys",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: ct.Keys,
		},
	}
}

// buildSecretsTools wraps secrets.Tool methods as agent tools.
func buildSecretsTools(secretsStore *security.SecretsStore, refs *security.RefStore, scanner *agent.SecretScanner) []*agent.Tool {
	st := toolsecrets.New(secretsStore, refs, scanner)
	return []*agent.Tool{
		{
			Name:        "secrets_store",
			Description: "Encrypt and store a secret value",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":  map[string]interface{}{"type": "string", "description": "Unique name for the secret"},
					"value": map[string]interface{}{"type": "string", "description": "The secret value to store"},
				},
				"required": []string{"name", "value"},
			},
			Handler: st.Store,
		},
		{
			Name:        "secrets_get",
			Description: "Retrieve a stored secret as a reference token. Returns an opaque {{secret:name}} token that is resolved at execution time by exec tools. The actual secret value never enters the agent context.",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{"type": "string", "description": "Name of the secret to retrieve"},
				},
				"required": []string{"name"},
			},
			Handler: st.Get,
		},
		{
			Name:        "secrets_list",
			Description: "List all stored secrets (metadata only, no values)",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: st.List,
		},
		{
			Name:        "secrets_delete",
			Description: "Delete a stored secret",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{"type": "string", "description": "Name of the secret to delete"},
				},
				"required": []string{"name"},
			},
			Handler: st.Delete,
		},
	}
}
