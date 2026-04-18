package storagebroker

import (
	"context"

	"github.com/langoai/lango/internal/security"
)

type payloadProtector struct {
	api API
}

// NewPayloadProtector adapts the broker API to security.PayloadProtector.
func NewPayloadProtector(api API) security.PayloadProtector {
	if api == nil {
		return nil
	}
	return &payloadProtector{api: api}
}

func (p *payloadProtector) EncryptPayload(plaintext []byte) (ciphertext []byte, nonce []byte, keyVersion int, err error) {
	result, err := p.api.EncryptPayload(contextBackground(), plaintext)
	if err != nil {
		return nil, nil, 0, err
	}
	return result.Ciphertext, result.Nonce, result.KeyVersion, nil
}

func (p *payloadProtector) DecryptPayload(ciphertext []byte, nonce []byte, keyVersion int) ([]byte, error) {
	result, err := p.api.DecryptPayload(contextBackground(), ciphertext, nonce, keyVersion)
	if err != nil {
		return nil, err
	}
	return result.Plaintext, nil
}

func contextBackground() context.Context {
	return context.Background()
}
