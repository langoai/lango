//go:build kms_gcp || kms_all

package security

import (
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGCPKMSProvider_ResolveKey(t *testing.T) {
	p := &GCPKMSProvider{
		defaultKeyID: "projects/my-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key",
	}

	tests := []struct {
		give string
		want string
	}{
		{
			give: "local",
			want: "projects/my-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key",
		},
		{
			give: "default",
			want: "projects/my-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key",
		},
		{
			give: "",
			want: "projects/my-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key",
		},
		{
			give: "projects/other/locations/global/keyRings/ring2/cryptoKeys/key2",
			want: "projects/other/locations/global/keyRings/ring2/cryptoKeys/key2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := p.resolveKey(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGCPKMSProvider_NewWithoutKeyID(t *testing.T) {
	_, err := newGCPKMSProvider(config.KMSConfig{
		KeyID: "",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrKMSInvalidKey)
}

func TestGCPKMSProvider_ClassifyError(t *testing.T) {
	p := &GCPKMSProvider{defaultKeyID: "test-key"}

	tests := []struct {
		name     string
		give     error
		wantSent error
	}{
		{
			name:     "permission denied",
			give:     status.Error(codes.PermissionDenied, "caller lacks permission"),
			wantSent: ErrKMSAccessDenied,
		},
		{
			name:     "unauthenticated",
			give:     status.Error(codes.Unauthenticated, "invalid credentials"),
			wantSent: ErrKMSAccessDenied,
		},
		{
			name:     "not found",
			give:     status.Error(codes.NotFound, "key not found"),
			wantSent: ErrKMSInvalidKey,
		},
		{
			name:     "failed precondition (disabled)",
			give:     status.Error(codes.FailedPrecondition, "key is disabled"),
			wantSent: ErrKMSKeyDisabled,
		},
		{
			name:     "resource exhausted (throttled)",
			give:     status.Error(codes.ResourceExhausted, "quota exceeded"),
			wantSent: ErrKMSThrottled,
		},
		{
			name:     "unavailable",
			give:     status.Error(codes.Unavailable, "service unavailable"),
			wantSent: ErrKMSUnavailable,
		},
		{
			name:     "internal error",
			give:     status.Error(codes.Internal, "internal error"),
			wantSent: ErrKMSUnavailable,
		},
		{
			name:     "invalid argument",
			give:     status.Error(codes.InvalidArgument, "bad request"),
			wantSent: ErrKMSInvalidKey,
		},
		{
			name:     "unknown code passthrough",
			give:     status.Error(codes.Canceled, "operation canceled"),
			wantSent: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.classifyError("test-op", "test-key", tt.give)
			require.Error(t, got)

			var kmsErr *KMSError
			require.ErrorAs(t, got, &kmsErr)
			assert.Equal(t, "gcp", kmsErr.Provider)
			assert.Equal(t, "test-op", kmsErr.Op)
			assert.Equal(t, "test-key", kmsErr.KeyID)

			if tt.wantSent != nil {
				assert.ErrorIs(t, got, tt.wantSent)
			}
		})
	}
}

func TestGCPKMSProvider_ClassifyNonGRPCError(t *testing.T) {
	p := &GCPKMSProvider{defaultKeyID: "test-key"}

	// Non-gRPC errors should be wrapped as-is.
	plainErr := assert.AnError
	got := p.classifyError("encrypt", "test-key", plainErr)

	var kmsErr *KMSError
	require.ErrorAs(t, got, &kmsErr)
	assert.Equal(t, "gcp", kmsErr.Provider)
	assert.Equal(t, plainErr, kmsErr.Err)
}
