package smartaccount

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/langoai/lango/internal/contract"
)

// safeFactoryABI is the ABI for the Safe proxy factory.
const safeFactoryABI = `[
	{
		"inputs": [
			{"name": "_singleton", "type": "address"},
			{"name": "initializer", "type": "bytes"},
			{"name": "saltNonce", "type": "uint256"}
		],
		"name": "createProxyWithNonce",
		"outputs": [{"name": "proxy", "type": "address"}],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "proxyCreationCode",
		"outputs": [{"name": "", "type": "bytes"}],
		"stateMutability": "pure",
		"type": "function"
	}
]`

// Factory handles Safe smart account deployment.
type Factory struct {
	caller       contract.ContractCaller
	rpc          *ethclient.Client
	factoryAddr  common.Address
	safe7579Addr common.Address
	fallbackAddr common.Address
	chainID      int64

	// proxyCode caches the result of proxyCreationCode() view call.
	proxyCodeMu sync.Mutex
	proxyCode   []byte
}

// NewFactory creates a smart account factory.
func NewFactory(
	caller contract.ContractCaller,
	rpc *ethclient.Client,
	factoryAddr common.Address,
	safe7579Addr common.Address,
	fallbackAddr common.Address,
	chainID int64,
) *Factory {
	return &Factory{
		caller:       caller,
		rpc:          rpc,
		factoryAddr:  factoryAddr,
		safe7579Addr: safe7579Addr,
		fallbackAddr: fallbackAddr,
		chainID:      chainID,
	}
}

// ComputeAddress computes the counterfactual Safe address via CREATE2.
// Uses the SafeProxyFactory's salt derivation:
//
//	deploymentSalt = keccak256(keccak256(initializer) ++ saltNonce)
//
// and the proxy initCode hash for the CREATE2 formula.
func (f *Factory) ComputeAddress(
	ctx context.Context,
	owner common.Address,
	salt *big.Int,
) (common.Address, error) {
	// Build initializer calldata (same as in Deploy).
	initData := buildSafeInitializer(
		owner, f.safe7579Addr, f.fallbackAddr,
	)

	// CREATE2 salt: keccak256(keccak256(initializer) ++ saltNonce)
	initHash := crypto.Keccak256(initData)
	saltBytes := make([]byte, 32)
	if salt != nil {
		b := salt.Bytes()
		copy(saltBytes[32-len(b):], b)
	}
	deploymentSalt := crypto.Keccak256(
		append(initHash, saltBytes...),
	)

	// Proxy initCode = proxyCreationCode ++ abi.encode(singleton)
	proxyCode, err := f.getProxyCreationCode(ctx)
	if err != nil {
		return common.Address{}, fmt.Errorf("get proxy creation code: %w", err)
	}

	singletonPadded := make([]byte, 32)
	copy(singletonPadded[12:], f.safe7579Addr.Bytes())
	initCode := append(proxyCode, singletonPadded...)
	initCodeHash := crypto.Keccak256(initCode)

	// CREATE2: keccak256(0xff ++ factory ++ salt ++ keccak256(initCode))
	data := make([]byte, 0, 85)
	data = append(data, 0xFF)
	data = append(data, f.factoryAddr.Bytes()...)
	data = append(data, deploymentSalt...)
	data = append(data, initCodeHash...)

	hash := crypto.Keccak256(data)
	return common.BytesToAddress(hash[12:]), nil
}

// getProxyCreationCode retrieves and caches the proxy creation code from the factory.
func (f *Factory) getProxyCreationCode(ctx context.Context) ([]byte, error) {
	f.proxyCodeMu.Lock()
	defer f.proxyCodeMu.Unlock()

	if f.proxyCode != nil {
		return f.proxyCode, nil
	}

	result, err := f.caller.Read(ctx, contract.ContractCallRequest{
		ChainID: f.chainID,
		Address: f.factoryAddr,
		ABI:     safeFactoryABI,
		Method:  "proxyCreationCode",
		Args:    []interface{}{},
	})
	if err != nil {
		return nil, fmt.Errorf("call proxyCreationCode: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("proxyCreationCode returned empty data")
	}

	code, ok := result.Data[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("proxyCreationCode: unexpected return type %T", result.Data[0])
	}

	f.proxyCode = code
	return code, nil
}

// Deploy deploys a new Safe account with ERC-7579 adapter.
// Returns the deployed account address and transaction hash.
func (f *Factory) Deploy(
	ctx context.Context,
	owner common.Address,
	salt *big.Int,
) (common.Address, string, error) {
	// Build Safe setup initializer data that configures the 7579 adapter.
	// The setup call configures owners, threshold, and fallback handler.
	initData := buildSafeInitializer(
		owner, f.safe7579Addr, f.fallbackAddr,
	)

	saltNonce := big.NewInt(0)
	if salt != nil {
		saltNonce = salt
	}

	result, err := f.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: f.chainID,
		Address: f.factoryAddr,
		ABI:     safeFactoryABI,
		Method:  "createProxyWithNonce",
		Args: []interface{}{
			f.safe7579Addr,
			initData,
			saltNonce,
		},
	})
	if err != nil {
		return common.Address{}, "",
			fmt.Errorf("deploy safe account: %w", err)
	}

	// Try to parse the actual deployed address from the return data.
	if len(result.Data) > 0 {
		if addr, ok := result.Data[0].(common.Address); ok {
			return addr, result.TxHash, nil
		}
	}

	// Fallback: compute the deterministic address via CREATE2.
	computed, compErr := f.ComputeAddress(ctx, owner, salt)
	if compErr != nil {
		return common.Address{}, result.TxHash,
			fmt.Errorf("compute deployed address: %w", compErr)
	}
	return computed, result.TxHash, nil
}

// IsDeployed checks if the account has code deployed at its address
// by calling eth_getCode via the RPC client.
func (f *Factory) IsDeployed(
	ctx context.Context,
	addr common.Address,
) (bool, error) {
	if f.rpc == nil {
		return false, fmt.Errorf(
			"check deployment %s: rpc client not configured",
			addr.Hex(),
		)
	}
	code, err := f.rpc.CodeAt(ctx, addr, nil)
	if err != nil {
		return false, fmt.Errorf(
			"get code at %s: %w", addr.Hex(), err,
		)
	}
	return len(code) > 0, nil
}

// safeSetupABI is the ABI for the Safe.setup() function.
const safeSetupABI = `[{
	"inputs": [
		{"name": "_owners", "type": "address[]"},
		{"name": "_threshold", "type": "uint256"},
		{"name": "to", "type": "address"},
		{"name": "data", "type": "bytes"},
		{"name": "fallbackHandler", "type": "address"},
		{"name": "paymentToken", "type": "address"},
		{"name": "payment", "type": "uint256"},
		{"name": "paymentReceiver", "type": "address"}
	],
	"name": "setup",
	"outputs": [],
	"type": "function"
}]`

// buildSafeInitializer creates the Safe.setup() ABI-encoded calldata
// that configures the owner, threshold=1, and 7579 adapter.
func buildSafeInitializer(
	owner common.Address,
	safe7579Addr common.Address,
	fallbackAddr common.Address,
) []byte {
	// Safe.setup(address[] owners, uint256 threshold, address to,
	//   bytes data, address fallbackHandler, address paymentToken,
	//   uint256 payment, address paymentReceiver)
	//
	// For ERC-7579: to = safe7579Addr (delegate call for adapter setup),
	// data = empty (setup done post-deploy), fallbackHandler = fallbackAddr.
	parsed, err := contract.ParseABI(safeSetupABI)
	if err != nil {
		// ABI is a compile-time constant; this should never fail.
		return nil
	}

	owners := []common.Address{owner}
	data, err := parsed.Pack(
		"setup",
		owners,           // _owners
		big.NewInt(1),    // _threshold
		safe7579Addr,     // to (7579 adapter setup as delegate call)
		[]byte{},         // data (empty, setup done post-deploy)
		fallbackAddr,     // fallbackHandler
		common.Address{}, // paymentToken (zero, no payment)
		big.NewInt(0),    // payment
		common.Address{}, // paymentReceiver (zero)
	)
	if err != nil {
		return nil
	}
	return data
}
