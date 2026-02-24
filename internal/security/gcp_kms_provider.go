//go:build kms_gcp || kms_all

package security

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	kmsapi "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var gcpLogger = logging.SubsystemSugar("gcp-kms")

// GCPKMSProvider implements CryptoProvider using Google Cloud KMS.
type GCPKMSProvider struct {
	client       *kmsapi.KeyManagementClient
	defaultKeyID string
	maxRetries   int
	timeout      time.Duration
}

var _ CryptoProvider = (*GCPKMSProvider)(nil)

func newGCPKMSProvider(kmsConfig config.KMSConfig) (CryptoProvider, error) {
	if kmsConfig.KeyID == "" {
		return nil, fmt.Errorf("new GCP KMS provider: %w", ErrKMSInvalidKey)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var opts []option.ClientOption
	if kmsConfig.Endpoint != "" {
		opts = append(opts, option.WithEndpoint(kmsConfig.Endpoint))
	}

	client, err := kmsapi.NewKeyManagementClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create GCP KMS client: %w", err)
	}

	maxRetries := kmsConfig.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	timeout := kmsConfig.TimeoutPerOperation
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	gcpLogger.Infow("GCP KMS provider initialized",
		"keyId", kmsConfig.KeyID,
		"maxRetries", maxRetries,
	)

	return &GCPKMSProvider{
		client:       client,
		defaultKeyID: kmsConfig.KeyID,
		maxRetries:   maxRetries,
		timeout:      timeout,
	}, nil
}

// Sign generates a signature using GCP KMS asymmetric signing.
// The payload is SHA-256 hashed before signing.
func (p *GCPKMSProvider) Sign(ctx context.Context, keyID string, payload []byte) ([]byte, error) {
	resolved := p.resolveKey(keyID)

	digest := sha256.Sum256(payload)

	var result []byte
	err := withRetry(ctx, p.maxRetries, func() error {
		opCtx, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()

		resp, err := p.client.AsymmetricSign(opCtx, &kmspb.AsymmetricSignRequest{
			Name: resolved,
			Digest: &kmspb.Digest{
				Digest: &kmspb.Digest_Sha256{
					Sha256: digest[:],
				},
			},
		})
		if err != nil {
			return p.classifyError("sign", resolved, err)
		}
		result = resp.Signature
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Encrypt encrypts plaintext using GCP KMS symmetric encryption.
func (p *GCPKMSProvider) Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error) {
	resolved := p.resolveKey(keyID)

	var result []byte
	err := withRetry(ctx, p.maxRetries, func() error {
		opCtx, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()

		resp, err := p.client.Encrypt(opCtx, &kmspb.EncryptRequest{
			Name:      resolved,
			Plaintext: plaintext,
		})
		if err != nil {
			return p.classifyError("encrypt", resolved, err)
		}
		result = resp.Ciphertext
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Decrypt decrypts ciphertext using GCP KMS symmetric encryption.
func (p *GCPKMSProvider) Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error) {
	resolved := p.resolveKey(keyID)

	var result []byte
	err := withRetry(ctx, p.maxRetries, func() error {
		opCtx, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()

		resp, err := p.client.Decrypt(opCtx, &kmspb.DecryptRequest{
			Name:       resolved,
			Ciphertext: ciphertext,
		})
		if err != nil {
			return p.classifyError("decrypt", resolved, err)
		}
		result = resp.Plaintext
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// resolveKey maps "local" and "default" aliases to the configured default key.
func (p *GCPKMSProvider) resolveKey(keyID string) string {
	if keyID == "local" || keyID == "default" || keyID == "" {
		return p.defaultKeyID
	}
	return keyID
}

// classifyError maps gRPC status codes to sentinel errors wrapped in KMSError.
func (p *GCPKMSProvider) classifyError(op, keyID string, err error) error {
	kmsErr := &KMSError{
		Provider: "gcp",
		Op:       op,
		KeyID:    keyID,
	}

	st, ok := status.FromError(err)
	if !ok {
		kmsErr.Err = err
		return kmsErr
	}

	switch st.Code() {
	case codes.PermissionDenied, codes.Unauthenticated:
		kmsErr.Err = fmt.Errorf("%w: %s", ErrKMSAccessDenied, st.Message())
	case codes.NotFound:
		kmsErr.Err = fmt.Errorf("%w: %s", ErrKMSInvalidKey, st.Message())
	case codes.FailedPrecondition:
		kmsErr.Err = fmt.Errorf("%w: %s", ErrKMSKeyDisabled, st.Message())
	case codes.ResourceExhausted:
		kmsErr.Err = fmt.Errorf("%w: %s", ErrKMSThrottled, st.Message())
	case codes.Unavailable, codes.Internal:
		kmsErr.Err = fmt.Errorf("%w: %s", ErrKMSUnavailable, st.Message())
	case codes.InvalidArgument:
		kmsErr.Err = fmt.Errorf("%w: %s", ErrKMSInvalidKey, st.Message())
	default:
		kmsErr.Err = err
	}

	return kmsErr
}
