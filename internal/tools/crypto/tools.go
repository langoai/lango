package crypto

import (
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/security"
)

// BuildTools creates crypto agent tools.
func BuildTools(cp security.CryptoProvider, keys *security.KeyRegistry, refs *security.RefStore, scanner *agent.SecretScanner) []*agent.Tool {
	ct := New(cp, keys, refs, scanner)
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
