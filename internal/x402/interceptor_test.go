package x402

import (
	"context"
	"testing"
	"time"

	"github.com/coinbase/x402/go/mechanisms/evm"
	"go.uber.org/zap"
)

func TestInterceptorHTTPClientHasBoundedTimeout(t *testing.T) {
	t.Parallel()

	interceptor := NewInterceptor(fakeSignerProvider{}, nil, Config{ChainID: 84532}, zap.NewNop().Sugar())

	client, err := interceptor.HTTPClient(context.Background())
	if err != nil {
		t.Fatalf("HTTPClient() error = %v", err)
	}
	if client.Timeout == 0 {
		t.Fatal("HTTPClient() returned an unbounded client timeout")
	}
	if client.Timeout < 15*time.Second {
		t.Fatalf("HTTPClient() timeout = %s, want at least 15s", client.Timeout)
	}
}

func TestInterceptorHTTPClientReusesBoundedClient(t *testing.T) {
	t.Parallel()

	interceptor := NewInterceptor(fakeSignerProvider{}, nil, Config{ChainID: 84532}, zap.NewNop().Sugar())

	first, err := interceptor.HTTPClient(context.Background())
	if err != nil {
		t.Fatalf("first HTTPClient() error = %v", err)
	}
	second, err := interceptor.HTTPClient(context.Background())
	if err != nil {
		t.Fatalf("second HTTPClient() error = %v", err)
	}
	if first != second {
		t.Fatal("HTTPClient() did not reuse the cached client")
	}
	if second.Timeout == 0 {
		t.Fatal("cached HTTPClient() returned an unbounded client timeout")
	}
}

type fakeSignerProvider struct{}

func (fakeSignerProvider) EvmSigner(context.Context) (evm.ClientEvmSigner, error) {
	return fakeEvmSigner{}, nil
}

type fakeEvmSigner struct{}

func (fakeEvmSigner) Address() string { return "0x0000000000000000000000000000000000000001" }

func (fakeEvmSigner) SignTypedData(context.Context, evm.TypedDataDomain, map[string][]evm.TypedDataField, string, map[string]interface{}) ([]byte, error) {
	return []byte("signature"), nil
}
