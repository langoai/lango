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

func TestInterceptorAccessorsExposeChainIDAndEnabledState(t *testing.T) {
	t.Parallel()

	interceptor := NewInterceptor(
		&interceptorAccessorsExposeChainIdAndEnabledStateX402SignerProvider{},
		nil,
		Config{Enabled: true, ChainID: 84532},
		zap.NewNop().Sugar(),
	)

	assert.True(t, interceptor.IsEnabled())
	assert.Equal(t, int64(84532), interceptor.ChainID())
}

func TestInterceptorHTTPClientPropagatesSignerAndConfigErrorsWithoutCaching(t *testing.T) {
	t.Parallel()

	t.Run("signer provider error", func(t *testing.T) {
		t.Parallel()

		provider := &interceptorAccessorsExposeChainIdAndEnabledStateX402SignerProvider{err: errors.New("signer unavailable")}
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

		provider := &interceptorAccessorsExposeChainIdAndEnabledStateX402SignerProvider{
			signer: interceptorAccessorsExposeChainIdAndEnabledStateX402Signer{address: "0x0000000000000000000000000000000000000052"},
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

func TestInterceptorHTTPClientCreatesAndCachesBoundedClient(t *testing.T) {
	t.Parallel()

	provider := &interceptorAccessorsExposeChainIdAndEnabledStateX402SignerProvider{
		signer: interceptorAccessorsExposeChainIdAndEnabledStateX402Signer{address: "0x0000000000000000000000000000000000000052"},
	}
	limiter := &interceptorAccessorsExposeChainIdAndEnabledStateSpendingLimiter{}
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

type interceptorAccessorsExposeChainIdAndEnabledStateX402SignerProvider struct {
	signer evm.ClientEvmSigner
	err    error
	calls  int
}

func (p *interceptorAccessorsExposeChainIdAndEnabledStateX402SignerProvider) EvmSigner(context.Context) (evm.ClientEvmSigner, error) {
	p.calls++
	if p.err != nil {
		return nil, p.err
	}
	if p.signer != nil {
		return p.signer, nil
	}
	return interceptorAccessorsExposeChainIdAndEnabledStateX402Signer{address: "0x0000000000000000000000000000000000000001"}, nil
}

type interceptorAccessorsExposeChainIdAndEnabledStateX402Signer struct {
	address string
}

func (s interceptorAccessorsExposeChainIdAndEnabledStateX402Signer) Address() string {
	return s.address
}

func (s interceptorAccessorsExposeChainIdAndEnabledStateX402Signer) SignTypedData(
	context.Context,
	evm.TypedDataDomain,
	map[string][]evm.TypedDataField,
	string,
	map[string]interface{},
) ([]byte, error) {
	return []byte("signature"), nil
}

type interceptorAccessorsExposeChainIdAndEnabledStateSpendingLimiter struct {
	checkCalls int
}

func (l *interceptorAccessorsExposeChainIdAndEnabledStateSpendingLimiter) Check(context.Context, *big.Int) error {
	l.checkCalls++
	return nil
}

func (*interceptorAccessorsExposeChainIdAndEnabledStateSpendingLimiter) Record(context.Context, *big.Int) error {
	return nil
}

func (*interceptorAccessorsExposeChainIdAndEnabledStateSpendingLimiter) DailySpent(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (*interceptorAccessorsExposeChainIdAndEnabledStateSpendingLimiter) DailyRemaining(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (*interceptorAccessorsExposeChainIdAndEnabledStateSpendingLimiter) IsAutoApprovable(context.Context, *big.Int) (bool, error) {
	return true, nil
}
