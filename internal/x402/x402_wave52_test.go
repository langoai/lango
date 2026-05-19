package x402

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/coinbase/x402/go/mechanisms/evm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWave52InterceptorAccessorsExposeChainIDAndEnabledState(t *testing.T) {
	t.Parallel()

	interceptor := NewInterceptor(
		&wave52X402SignerProvider{},
		nil,
		Config{Enabled: true, ChainID: 84532},
		zap.NewNop().Sugar(),
	)

	assert.True(t, interceptor.IsEnabled())
	assert.Equal(t, int64(84532), interceptor.ChainID())
}

func TestWave52InterceptorHTTPClientPropagatesSignerAndConfigErrorsWithoutCaching(t *testing.T) {
	t.Parallel()

	t.Run("signer provider error", func(t *testing.T) {
		t.Parallel()

		provider := &wave52X402SignerProvider{err: errors.New("signer unavailable")}
		interceptor := NewInterceptor(
			provider,
			nil,
			Config{ChainID: 84532},
			zap.NewNop().Sugar(),
		)

		client, err := interceptor.HTTPClient(context.Background())
		require.Error(t, err)
		assert.Nil(t, client)
		assert.ErrorContains(t, err, "create EVM signer")
		assert.ErrorContains(t, err, "signer unavailable")

		client, err = interceptor.HTTPClient(context.Background())
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Equal(t, 2, provider.calls)
	})

	t.Run("invalid max auto pay amount", func(t *testing.T) {
		t.Parallel()

		provider := &wave52X402SignerProvider{
			signer: wave52X402Signer{address: "0x0000000000000000000000000000000000000052"},
		}
		interceptor := NewInterceptor(
			provider,
			nil,
			Config{ChainID: 84532, MaxAutoPayAmount: "not-usdc"},
			zap.NewNop().Sugar(),
		)

		client, err := interceptor.HTTPClient(context.Background())
		require.Error(t, err)
		assert.Nil(t, client)
		assert.ErrorContains(t, err, "parse maxAutoPayAmount")

		client, err = interceptor.HTTPClient(context.Background())
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Equal(t, 2, provider.calls)
	})
}

func TestWave52InterceptorHTTPClientCreatesAndCachesBoundedClient(t *testing.T) {
	t.Parallel()

	provider := &wave52X402SignerProvider{
		signer: wave52X402Signer{address: "0x0000000000000000000000000000000000000052"},
	}
	limiter := &wave52SpendingLimiter{}
	interceptor := NewInterceptor(
		provider,
		limiter,
		Config{Enabled: true, ChainID: 84532, MaxAutoPayAmount: "1.25"},
		zap.NewNop().Sugar(),
	)

	first, err := interceptor.HTTPClient(context.Background())
	require.NoError(t, err)
	require.NotNil(t, first)
	assert.Equal(t, defaultX402HTTPClientTimeout, first.Timeout)

	second, err := interceptor.HTTPClient(context.Background())
	require.NoError(t, err)
	assert.Same(t, first, second)
	assert.Equal(t, 1, provider.calls)
	// Client creation wires payment hooks but does not charge or check limits
	// until a payment challenge is handled by the SDK.
	assert.Zero(t, limiter.checkCalls)
}

type wave52X402SignerProvider struct {
	signer evm.ClientEvmSigner
	err    error
	calls  int
}

func (p *wave52X402SignerProvider) EvmSigner(context.Context) (evm.ClientEvmSigner, error) {
	p.calls++
	if p.err != nil {
		return nil, p.err
	}
	if p.signer != nil {
		return p.signer, nil
	}
	return wave52X402Signer{address: "0x0000000000000000000000000000000000000001"}, nil
}

type wave52X402Signer struct {
	address string
}

func (s wave52X402Signer) Address() string { return s.address }

func (s wave52X402Signer) SignTypedData(
	context.Context,
	evm.TypedDataDomain,
	map[string][]evm.TypedDataField,
	string,
	map[string]interface{},
) ([]byte, error) {
	return []byte("signature"), nil
}

type wave52SpendingLimiter struct {
	checkCalls int
}

func (l *wave52SpendingLimiter) Check(context.Context, *big.Int) error {
	l.checkCalls++
	return nil
}

func (*wave52SpendingLimiter) Record(context.Context, *big.Int) error { return nil }

func (*wave52SpendingLimiter) DailySpent(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (*wave52SpendingLimiter) DailyRemaining(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (*wave52SpendingLimiter) IsAutoApprovable(context.Context, *big.Int) (bool, error) {
	return true, nil
}
