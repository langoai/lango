package paymaster

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProvider struct {
	calls   int
	results []*SponsorResult
	errors  []error
}

func (m *mockProvider) SponsorUserOp(_ context.Context, _ *SponsorRequest) (*SponsorResult, error) {
	idx := m.calls
	m.calls++
	if idx < len(m.errors) && m.errors[idx] != nil {
		return nil, m.errors[idx]
	}
	if idx < len(m.results) {
		return m.results[idx], nil
	}
	return &SponsorResult{PaymasterAndData: []byte{0x01}}, nil
}

func (m *mockProvider) Type() string { return "mock" }

func TestRecoverableProvider_SuccessOnFirstAttempt(t *testing.T) {
	mock := &mockProvider{
		results: []*SponsorResult{{PaymasterAndData: []byte{0x42}}},
	}
	rp := NewRecoverableProvider(mock, DefaultRecoveryConfig())
	result, err := rp.SponsorUserOp(context.Background(), &SponsorRequest{})
	require.NoError(t, err)
	assert.Equal(t, []byte{0x42}, result.PaymasterAndData)
	assert.Equal(t, 1, mock.calls)
}

func TestRecoverableProvider_RetryOnTransient(t *testing.T) {
	mock := &mockProvider{
		errors:  []error{ErrPaymasterTimeout, ErrPaymasterTimeout, nil},
		results: []*SponsorResult{nil, nil, {PaymasterAndData: []byte{0x99}}},
	}
	cfg := RecoveryConfig{MaxRetries: 2, BaseDelay: time.Millisecond, FallbackMode: FallbackAbort}
	rp := NewRecoverableProvider(mock, cfg)
	result, err := rp.SponsorUserOp(context.Background(), &SponsorRequest{})
	require.NoError(t, err)
	assert.Equal(t, []byte{0x99}, result.PaymasterAndData)
	assert.Equal(t, 3, mock.calls)
}

func TestRecoverableProvider_PermanentError(t *testing.T) {
	mock := &mockProvider{
		errors: []error{ErrPaymasterRejected},
	}
	rp := NewRecoverableProvider(mock, DefaultRecoveryConfig())
	_, err := rp.SponsorUserOp(context.Background(), &SponsorRequest{})
	assert.ErrorIs(t, err, ErrPaymasterRejected)
	assert.Equal(t, 1, mock.calls) // no retry
}

func TestRecoverableProvider_FallbackDirectGas(t *testing.T) {
	mock := &mockProvider{
		errors: []error{ErrPaymasterTimeout, ErrPaymasterTimeout, ErrPaymasterTimeout},
	}
	cfg := RecoveryConfig{MaxRetries: 2, BaseDelay: time.Millisecond, FallbackMode: FallbackDirectGas}
	rp := NewRecoverableProvider(mock, cfg)
	result, err := rp.SponsorUserOp(context.Background(), &SponsorRequest{})
	require.NoError(t, err)
	assert.Empty(t, result.PaymasterAndData) // direct gas fallback
	assert.Equal(t, 3, mock.calls)
}

func TestRecoverableProvider_FallbackAbort(t *testing.T) {
	mock := &mockProvider{
		errors: []error{ErrPaymasterTimeout, ErrPaymasterTimeout, ErrPaymasterTimeout},
	}
	cfg := RecoveryConfig{MaxRetries: 2, BaseDelay: time.Millisecond, FallbackMode: FallbackAbort}
	rp := NewRecoverableProvider(mock, cfg)
	_, err := rp.SponsorUserOp(context.Background(), &SponsorRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "paymaster retries exhausted")
}

func TestRecoverableProvider_ContextCancellation(t *testing.T) {
	mock := &mockProvider{
		errors: []error{ErrPaymasterTimeout, ErrPaymasterTimeout},
	}
	cfg := RecoveryConfig{MaxRetries: 3, BaseDelay: time.Second, FallbackMode: FallbackAbort}
	rp := NewRecoverableProvider(mock, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := rp.SponsorUserOp(ctx, &SponsorRequest{})
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRecoverableProvider_Type(t *testing.T) {
	mock := &mockProvider{}
	rp := NewRecoverableProvider(mock, DefaultRecoveryConfig())
	assert.Equal(t, "mock+recovery", rp.Type())
}
