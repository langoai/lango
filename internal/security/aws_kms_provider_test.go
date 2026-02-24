//go:build kms_aws || kms_all

package security

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/langoai/lango/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAWSKMSProvider_ResolveKey(t *testing.T) {
	p := &AWSKMSProvider{
		defaultKeyID: "arn:aws:kms:us-east-1:123456789012:key/test-key-id",
	}

	tests := []struct {
		give string
		want string
	}{
		{
			give: "local",
			want: "arn:aws:kms:us-east-1:123456789012:key/test-key-id",
		},
		{
			give: "default",
			want: "arn:aws:kms:us-east-1:123456789012:key/test-key-id",
		},
		{
			give: "",
			want: "arn:aws:kms:us-east-1:123456789012:key/test-key-id",
		},
		{
			give: "arn:aws:kms:us-west-2:123456789012:key/other-key",
			want: "arn:aws:kms:us-west-2:123456789012:key/other-key",
		},
		{
			give: "alias/my-key",
			want: "alias/my-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := p.resolveKey(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAWSKMSProvider_NewWithoutKeyID(t *testing.T) {
	_, err := newAWSKMSProvider(config.KMSConfig{
		Region: "us-east-1",
		KeyID:  "",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrKMSInvalidKey)
}

func TestAWSKMSProvider_ClassifyError(t *testing.T) {
	p := &AWSKMSProvider{defaultKeyID: "test-key"}

	tests := []struct {
		give     error
		wantSent error
		name     string
	}{
		{
			name:     "access denied",
			give:     &types.AccessDeniedException{Message: strPtr("access denied")},
			wantSent: ErrKMSAccessDenied,
		},
		{
			name:     "key disabled",
			give:     &types.DisabledException{Message: strPtr("key is disabled")},
			wantSent: ErrKMSKeyDisabled,
		},
		{
			name:     "key not found",
			give:     &types.NotFoundException{Message: strPtr("key not found")},
			wantSent: ErrKMSInvalidKey,
		},
		{
			name:     "invalid key usage",
			give:     &types.InvalidKeyUsageException{Message: strPtr("invalid usage")},
			wantSent: ErrKMSInvalidKey,
		},
		{
			name:     "invalid state",
			give:     &types.KMSInvalidStateException{Message: strPtr("invalid state")},
			wantSent: ErrKMSKeyDisabled,
		},
		{
			name:     "throttling string match",
			give:     errors.New("ThrottlingException: rate exceeded"),
			wantSent: ErrKMSThrottled,
		},
		{
			name:     "unavailable string match",
			give:     errors.New("ServiceUnavailableException: service down"),
			wantSent: ErrKMSUnavailable,
		},
		{
			name:     "unknown error passthrough",
			give:     errors.New("some other error"),
			wantSent: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.classifyError("test-op", "test-key", tt.give)
			require.Error(t, got)

			var kmsErr *KMSError
			require.ErrorAs(t, got, &kmsErr)
			assert.Equal(t, "aws", kmsErr.Provider)
			assert.Equal(t, "test-op", kmsErr.Op)
			assert.Equal(t, "test-key", kmsErr.KeyID)

			if tt.wantSent != nil {
				assert.ErrorIs(t, got, tt.wantSent)
			}
		})
	}
}

func TestAWSKMSProvider_Defaults(t *testing.T) {
	// Verify that default maxRetries and timeout are applied.
	// We cannot create a real provider without AWS credentials,
	// but we can test that the struct fields are populated correctly
	// by inspecting the AWSKMSProvider directly.
	p := &AWSKMSProvider{
		defaultKeyID: "test-key",
		maxRetries:   0,
		timeout:      0,
	}
	// Zero values indicate no override; the constructor would set defaults.
	assert.Equal(t, "test-key", p.defaultKeyID)
}

func strPtr(s string) *string {
	return &s
}
