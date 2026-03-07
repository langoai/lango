package security

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPCProvider_Sign(t *testing.T) {
	t.Parallel()

	provider := NewRPCProvider()

	// Mock Sender that replies immediately
	provider.SetSender(func(event string, payload interface{}) error {
		assert.Equal(t, "sign.request", event)
		req := payload.(SignRequest)

		// Simulate response
		resp := SignResponse{
			ID:        req.ID,
			Signature: []byte("signature_bytes"),
		}

		// Handle response in a goroutine to avoid blocking if the channel buffer was 0 (it's 1, but good practice)
		go func() {
			require.NoError(t, provider.HandleSignResponse(resp))
		}()
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sig, err := provider.Sign(ctx, "key1", []byte("data"))
	require.NoError(t, err)
	assert.Equal(t, "signature_bytes", string(sig))
}

func TestRPCProvider_Encrypt(t *testing.T) {
	t.Parallel()

	provider := NewRPCProvider()

	provider.SetSender(func(event string, payload interface{}) error {
		assert.Equal(t, "encrypt.request", event)
		req := payload.(EncryptRequest)

		resp := EncryptResponse{
			ID:         req.ID,
			Ciphertext: []byte("encrypted_bytes"),
		}

		go func() {
			provider.HandleEncryptResponse(resp)
		}()
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cipher, err := provider.Encrypt(ctx, "key1", []byte("plaintext"))
	require.NoError(t, err)
	assert.Equal(t, "encrypted_bytes", string(cipher))
}

func TestRPCProvider_Decrypt(t *testing.T) {
	t.Parallel()

	provider := NewRPCProvider()

	provider.SetSender(func(event string, payload interface{}) error {
		assert.Equal(t, "decrypt.request", event)
		req := payload.(DecryptRequest)

		resp := DecryptResponse{
			ID:        req.ID,
			Plaintext: []byte("decrypted_bytes"),
		}

		go func() {
			provider.HandleDecryptResponse(resp)
		}()
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	plain, err := provider.Decrypt(ctx, "key1", []byte("ciphertext"))
	require.NoError(t, err)
	assert.Equal(t, "decrypted_bytes", string(plain))
}
