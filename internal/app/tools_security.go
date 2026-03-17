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
			Parameters: agent.Schema().
				Str("data", "The data to encrypt").
				Str("keyId", "Key ID to use (default: default key)").
				Required("data").
				Build(),
			Handler: ct.Encrypt,
		},
		{
			Name:        "crypto_decrypt",
			Description: "Decrypt data using a registered key. Returns an opaque {{decrypt:id}} reference token. The decrypted value never enters the agent context.",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: agent.Schema().
				Str("ciphertext", "Base64-encoded ciphertext to decrypt").
				Str("keyId", "Key ID to use (default: default key)").
				Required("ciphertext").
				Build(),
			Handler: ct.Decrypt,
		},
		{
			Name:        "crypto_sign",
			Description: "Generate a digital signature for data",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: agent.Schema().
				Str("data", "The data to sign").
				Str("keyId", "Key ID to use").
				Required("data").
				Build(),
			Handler: ct.Sign,
		},
		{
			Name:        "crypto_hash",
			Description: "Compute a cryptographic hash of data",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: agent.Schema().
				Str("data", "The data to hash").
				Enum("algorithm", "Hash algorithm: sha256 or sha512", "sha256", "sha512").
				Required("data").
				Build(),
			Handler: ct.Hash,
		},
		{
			Name:        "crypto_keys",
			Description: "List all registered cryptographic keys",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: agent.Schema().Build(),
			Handler:     ct.Keys,
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
			Parameters: agent.Schema().
				Str("name", "Unique name for the secret").
				Str("value", "The secret value to store").
				Required("name", "value").
				Build(),
			Handler: st.Store,
		},
		{
			Name:        "secrets_get",
			Description: "Retrieve a stored secret as a reference token. Returns an opaque {{secret:name}} token that is resolved at execution time by exec tools. The actual secret value never enters the agent context.",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: agent.Schema().
				Str("name", "Name of the secret to retrieve").
				Required("name").
				Build(),
			Handler: st.Get,
		},
		{
			Name:        "secrets_list",
			Description: "List all stored secrets (metadata only, no values)",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: agent.Schema().Build(),
			Handler:     st.List,
		},
		{
			Name:        "secrets_delete",
			Description: "Delete a stored secret",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: agent.Schema().
				Str("name", "Name of the secret to delete").
				Required("name").
				Build(),
			Handler: st.Delete,
		},
	}
}
