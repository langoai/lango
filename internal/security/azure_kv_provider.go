//go:build kms_azure || kms_all

package security

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azkeys"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
)

var azureLogger = logging.SubsystemSugar("azure-kv")

// AzureKVProvider implements CryptoProvider using Azure Key Vault.
type AzureKVProvider struct {
	client         *azkeys.Client
	defaultKeyName string
	keyVersion     string
	maxRetries     int
	timeout        time.Duration
}

var _ CryptoProvider = (*AzureKVProvider)(nil)

func newAzureKVProvider(kmsConfig config.KMSConfig) (CryptoProvider, error) {
	if kmsConfig.Azure.VaultURL == "" {
		return nil, fmt.Errorf("new Azure KV provider: vault URL is required")
	}
	if kmsConfig.KeyID == "" {
		return nil, fmt.Errorf("new Azure KV provider: %w", ErrKMSInvalidKey)
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("new Azure credential: %w", err)
	}

	client, err := azkeys.NewClient(kmsConfig.Azure.VaultURL, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("new Azure KV client: %w", err)
	}

	maxRetries := kmsConfig.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	timeout := kmsConfig.TimeoutPerOperation
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	azureLogger.Infow("Azure Key Vault provider initialized",
		"vaultUrl", kmsConfig.Azure.VaultURL,
		"keyId", kmsConfig.KeyID,
		"maxRetries", maxRetries,
	)

	return &AzureKVProvider{
		client:         client,
		defaultKeyName: kmsConfig.KeyID,
		keyVersion:     kmsConfig.Azure.KeyVersion,
		maxRetries:     maxRetries,
		timeout:        timeout,
	}, nil
}

// Sign generates a signature using Azure Key Vault ES256.
func (p *AzureKVProvider) Sign(ctx context.Context, keyID string, payload []byte) ([]byte, error) {
	resolved := p.resolveKey(keyID)

	// Compute SHA-256 digest for ES256 signing.
	digest := sha256.Sum256(payload)

	var result []byte
	err := withRetry(ctx, p.maxRetries, func() error {
		opCtx, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()

		alg := azkeys.SignatureAlgorithmES256
		resp, err := p.client.Sign(opCtx, resolved, p.keyVersion, azkeys.SignParameters{
			Algorithm: &alg,
			Value:     digest[:],
		}, nil)
		if err != nil {
			return p.classifyError("sign", resolved, err)
		}
		result = resp.Result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Encrypt encrypts plaintext using Azure Key Vault RSA-OAEP.
func (p *AzureKVProvider) Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error) {
	resolved := p.resolveKey(keyID)

	var result []byte
	err := withRetry(ctx, p.maxRetries, func() error {
		opCtx, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()

		alg := azkeys.EncryptionAlgorithmRSAOAEP
		resp, err := p.client.Encrypt(opCtx, resolved, p.keyVersion, azkeys.KeyOperationParameters{
			Algorithm: &alg,
			Value:     plaintext,
		}, nil)
		if err != nil {
			return p.classifyError("encrypt", resolved, err)
		}
		result = resp.Result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Decrypt decrypts ciphertext using Azure Key Vault RSA-OAEP.
func (p *AzureKVProvider) Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error) {
	resolved := p.resolveKey(keyID)

	var result []byte
	err := withRetry(ctx, p.maxRetries, func() error {
		opCtx, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()

		alg := azkeys.EncryptionAlgorithmRSAOAEP
		resp, err := p.client.Decrypt(opCtx, resolved, p.keyVersion, azkeys.KeyOperationParameters{
			Algorithm: &alg,
			Value:     ciphertext,
		}, nil)
		if err != nil {
			return p.classifyError("decrypt", resolved, err)
		}
		result = resp.Result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// resolveKey maps "local" and "default" aliases to the configured default key.
func (p *AzureKVProvider) resolveKey(keyID string) string {
	if keyID == "local" || keyID == "default" || keyID == "" {
		return p.defaultKeyName
	}
	return keyID
}

// classifyError maps Azure Key Vault errors to sentinel errors wrapped in KMSError.
func (p *AzureKVProvider) classifyError(op, keyID string, err error) error {
	kmsErr := &KMSError{
		Provider: "azure",
		Op:       op,
		KeyID:    keyID,
	}

	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) {
		switch respErr.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			kmsErr.Err = fmt.Errorf("%w: %s", ErrKMSAccessDenied, err)
		case http.StatusTooManyRequests:
			kmsErr.Err = fmt.Errorf("%w: %s", ErrKMSThrottled, err)
		case http.StatusServiceUnavailable:
			kmsErr.Err = fmt.Errorf("%w: %s", ErrKMSUnavailable, err)
		default:
			kmsErr.Err = err
		}
	} else {
		kmsErr.Err = err
	}

	return kmsErr
}
