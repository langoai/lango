package smartaccount

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/langoai/lango/internal/contract"
)

// safeFactoryABI is the ABI for the Safe proxy factory's createProxyWithNonce.
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
		"inputs": [
			{"name": "_singleton", "type": "address"},
			{"name": "initializer", "type": "bytes"},
			{"name": "saltNonce", "type": "uint256"}
		],
		"name": "proxyCreationCode",
		"outputs": [{"name": "", "type": "bytes"}],
		"stateMutability": "view",
		"type": "function"
	}
]`

// safeCodeAtABI is a minimal ABI for checking deployed code.
const safeCodeAtABI = `[
	{
		"inputs": [{"name": "addr", "type": "address"}],
		"name": "getCode",
		"outputs": [{"name": "", "type": "bytes"}],
		"stateMutability": "view",
		"type": "function"
	}
]`

// Factory handles Safe smart account deployment.
type Factory struct {
	caller       contract.ContractCaller
	factoryAddr  common.Address
	safe7579Addr common.Address
	fallbackAddr common.Address
	chainID      int64
}

// NewFactory creates a smart account factory.
func NewFactory(
	caller contract.ContractCaller,
	factoryAddr common.Address,
	safe7579Addr common.Address,
	fallbackAddr common.Address,
	chainID int64,
) *Factory {
	return &Factory{
		caller:       caller,
		factoryAddr:  factoryAddr,
		safe7579Addr: safe7579Addr,
		fallbackAddr: fallbackAddr,
		chainID:      chainID,
	}
}

// ComputeAddress computes the counterfactual Safe address via CREATE2.
// Uses the owner address and salt as deterministic deployment inputs.
func (f *Factory) ComputeAddress(
	owner common.Address,
	salt *big.Int,
) common.Address {
	// CREATE2: keccak256(0xff ++ factory ++ salt ++ keccak256(initCode))
	// The salt incorporates the owner for deterministic per-owner addresses.
	saltBytes := make([]byte, 32)
	if salt != nil {
		b := salt.Bytes()
		copy(saltBytes[32-len(b):], b)
	}

	// Combine owner and salt nonce into the CREATE2 salt.
	combinedSalt := crypto.Keccak256(
		owner.Bytes(),
		saltBytes,
	)

	// Simplified initCode hash using the singleton and owner.
	initCodeHash := crypto.Keccak256(
		f.safe7579Addr.Bytes(),
		owner.Bytes(),
	)

	// CREATE2 formula.
	data := make([]byte, 0, 85)
	data = append(data, 0xFF)
	data = append(data, f.factoryAddr.Bytes()...)
	data = append(data, combinedSalt...)
	data = append(data, initCodeHash...)

	hash := crypto.Keccak256(data)
	return common.BytesToAddress(hash[12:])
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

	// Extract the proxy address from the result.
	if len(result.Data) > 0 {
		if addr, ok := result.Data[0].(common.Address); ok {
			return addr, result.TxHash, nil
		}
	}

	// If the result data doesn't contain the address directly,
	// compute it deterministically.
	computed := f.ComputeAddress(owner, salt)
	return computed, result.TxHash, nil
}

// IsDeployed checks if the account has code deployed at its address.
func (f *Factory) IsDeployed(
	ctx context.Context,
	addr common.Address,
) (bool, error) {
	// Use a Read call to check if code exists at the address.
	// We attempt to call a view function; if the contract has code
	// the call proceeds, otherwise it fails.
	result, err := f.caller.Read(ctx, contract.ContractCallRequest{
		ChainID: f.chainID,
		Address: addr,
		ABI:     Safe7579ABI,
		Method:  "isModuleInstalled",
		Args: []interface{}{
			uint8(ModuleTypeValidator),
			common.Address{},
			[]byte{},
		},
	})
	if err != nil {
		// If the call fails, the contract is likely not deployed.
		return false, nil
	}
	// If the call succeeds, the contract exists.
	_ = result
	return true, nil
}

// Safe7579ABI is the ABI for the Safe7579 adapter contract.
// Exported for use by both Factory and Manager.
const Safe7579ABI = `[
	{
		"inputs": [
			{"name": "moduleTypeId", "type": "uint256"},
			{"name": "module", "type": "address"},
			{"name": "initData", "type": "bytes"}
		],
		"name": "installModule",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "moduleTypeId", "type": "uint256"},
			{"name": "module", "type": "address"},
			{"name": "deInitData", "type": "bytes"}
		],
		"name": "uninstallModule",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "mode", "type": "bytes32"},
			{"name": "executionCalldata", "type": "bytes"}
		],
		"name": "execute",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "moduleTypeId", "type": "uint256"},
			{"name": "module", "type": "address"},
			{"name": "additionalContext", "type": "bytes"}
		],
		"name": "isModuleInstalled",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	}
]`

// buildSafeInitializer creates the Safe setup calldata that
// configures the owner, threshold=1, and 7579 adapter.
func buildSafeInitializer(
	owner common.Address,
	safe7579Addr common.Address,
	fallbackAddr common.Address,
) []byte {
	// In a full implementation this would ABI-encode the Safe.setup()
	// call with owner list, threshold, to (7579 setup), data, fallback
	// handler, payment token, payment, and payment receiver.
	// For now, encode owner + adapter addresses as a placeholder.
	data := make([]byte, 0, 60)
	data = append(data, owner.Bytes()...)
	data = append(data, safe7579Addr.Bytes()...)
	data = append(data, fallbackAddr.Bytes()...)
	return data
}
