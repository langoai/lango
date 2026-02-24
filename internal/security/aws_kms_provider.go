//go:build kms_aws || kms_all

package security

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
)

var awsLogger = logging.SubsystemSugar("aws-kms")

// AWSKMSProvider implements CryptoProvider using AWS KMS.
type AWSKMSProvider struct {
	client       *kms.Client
	defaultKeyID string
	maxRetries   int
	timeout      time.Duration
}

var _ CryptoProvider = (*AWSKMSProvider)(nil)

func newAWSKMSProvider(kmsConfig config.KMSConfig) (CryptoProvider, error) {
	if kmsConfig.KeyID == "" {
		return nil, fmt.Errorf("new AWS KMS provider: %w", ErrKMSInvalidKey)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var optFns []func(*awsconfig.LoadOptions) error
	if kmsConfig.Region != "" {
		optFns = append(optFns, awsconfig.WithRegion(kmsConfig.Region))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	var kmsOptFns []func(*kms.Options)
	if kmsConfig.Endpoint != "" {
		kmsOptFns = append(kmsOptFns, func(o *kms.Options) {
			o.BaseEndpoint = aws.String(kmsConfig.Endpoint)
		})
	}

	client := kms.NewFromConfig(cfg, kmsOptFns...)

	maxRetries := kmsConfig.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	timeout := kmsConfig.TimeoutPerOperation
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	awsLogger.Infow("AWS KMS provider initialized",
		"region", kmsConfig.Region,
		"keyId", kmsConfig.KeyID,
		"maxRetries", maxRetries,
	)

	return &AWSKMSProvider{
		client:       client,
		defaultKeyID: kmsConfig.KeyID,
		maxRetries:   maxRetries,
		timeout:      timeout,
	}, nil
}

// Sign generates a signature using AWS KMS ECDSA_SHA_256.
func (p *AWSKMSProvider) Sign(ctx context.Context, keyID string, payload []byte) ([]byte, error) {
	resolved := p.resolveKey(keyID)

	var result []byte
	err := withRetry(ctx, p.maxRetries, func() error {
		opCtx, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()

		out, err := p.client.Sign(opCtx, &kms.SignInput{
			KeyId:            aws.String(resolved),
			Message:          payload,
			SigningAlgorithm: types.SigningAlgorithmSpecEcdsaSha256,
			MessageType:      types.MessageTypeRaw,
		})
		if err != nil {
			return p.classifyError("sign", resolved, err)
		}
		result = out.Signature
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Encrypt encrypts plaintext using AWS KMS symmetric encryption.
func (p *AWSKMSProvider) Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error) {
	resolved := p.resolveKey(keyID)

	var result []byte
	err := withRetry(ctx, p.maxRetries, func() error {
		opCtx, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()

		out, err := p.client.Encrypt(opCtx, &kms.EncryptInput{
			KeyId:               aws.String(resolved),
			Plaintext:           plaintext,
			EncryptionAlgorithm: types.EncryptionAlgorithmSpecSymmetricDefault,
		})
		if err != nil {
			return p.classifyError("encrypt", resolved, err)
		}
		result = out.CiphertextBlob
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Decrypt decrypts ciphertext using AWS KMS symmetric encryption.
func (p *AWSKMSProvider) Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error) {
	resolved := p.resolveKey(keyID)

	var result []byte
	err := withRetry(ctx, p.maxRetries, func() error {
		opCtx, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()

		out, err := p.client.Decrypt(opCtx, &kms.DecryptInput{
			KeyId:               aws.String(resolved),
			CiphertextBlob:      ciphertext,
			EncryptionAlgorithm: types.EncryptionAlgorithmSpecSymmetricDefault,
		})
		if err != nil {
			return p.classifyError("decrypt", resolved, err)
		}
		result = out.Plaintext
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// resolveKey maps "local" and "default" aliases to the configured default key.
func (p *AWSKMSProvider) resolveKey(keyID string) string {
	if keyID == "local" || keyID == "default" || keyID == "" {
		return p.defaultKeyID
	}
	return keyID
}

// classifyError maps AWS KMS errors to sentinel errors wrapped in KMSError.
func (p *AWSKMSProvider) classifyError(op, keyID string, err error) error {
	kmsErr := &KMSError{
		Provider: "aws",
		Op:       op,
		KeyID:    keyID,
	}

	var accessDenied *types.AccessDeniedException
	var disabled *types.DisabledException
	var notFound *types.NotFoundException
	var invalidKeyUsage *types.InvalidKeyUsageException
	var kmsInvalidState *types.KMSInvalidStateException

	switch {
	case errors.As(err, &accessDenied):
		kmsErr.Err = fmt.Errorf("%w: %s", ErrKMSAccessDenied, err)
	case errors.As(err, &disabled):
		kmsErr.Err = fmt.Errorf("%w: %s", ErrKMSKeyDisabled, err)
	case errors.As(err, &notFound):
		kmsErr.Err = fmt.Errorf("%w: %s", ErrKMSInvalidKey, err)
	case errors.As(err, &invalidKeyUsage):
		kmsErr.Err = fmt.Errorf("%w: %s", ErrKMSInvalidKey, err)
	case errors.As(err, &kmsInvalidState):
		kmsErr.Err = fmt.Errorf("%w: %s", ErrKMSKeyDisabled, err)
	case isAWSThrottling(err):
		kmsErr.Err = fmt.Errorf("%w: %s", ErrKMSThrottled, err)
	case isAWSUnavailable(err):
		kmsErr.Err = fmt.Errorf("%w: %s", ErrKMSUnavailable, err)
	default:
		kmsErr.Err = err
	}

	return kmsErr
}

// isAWSThrottling checks if the error message indicates throttling.
func isAWSThrottling(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "ThrottlingException") ||
		strings.Contains(msg, "Throttling") ||
		strings.Contains(msg, "Rate exceeded")
}

// isAWSUnavailable checks if the error message indicates service unavailability.
func isAWSUnavailable(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "ServiceUnavailableException") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "no such host")
}
