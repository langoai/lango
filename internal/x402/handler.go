package x402

import (
	x402sdk "github.com/coinbase/x402/go"
	evmclient "github.com/coinbase/x402/go/mechanisms/evm/exact/client"
)

// NewX402Client creates an X402 SDK client configured for the given chain and signer.
// The client is registered with the exact EVM scheme for the specified CAIP-2 network.
func NewX402Client(signerProvider SignerProvider, chainID int64) (*x402sdk.X402Client, error) {
	signer, err := signerProvider.EvmSigner(nil)
	if err != nil {
		return nil, err
	}

	network := x402sdk.Network(CAIP2Network(chainID))
	scheme := evmclient.NewExactEvmScheme(signer)

	client := x402sdk.Newx402Client()
	client.Register(network, scheme)

	return client, nil
}
