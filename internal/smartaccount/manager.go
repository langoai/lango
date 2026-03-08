package smartaccount

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/langoai/lango/internal/contract"
	"github.com/langoai/lango/internal/smartaccount/bundler"
	"github.com/langoai/lango/internal/wallet"
)

// Compile-time check.
var _ AccountManager = (*Manager)(nil)

// Manager implements AccountManager for Safe-based smart accounts
// with ERC-7579 module support and ERC-4337 UserOp submission.
type Manager struct {
	factory     *Factory
	bundler     *bundler.Client
	caller      contract.ContractCaller
	wallet      wallet.WalletProvider
	chainID     int64
	entryPoint  common.Address
	accountAddr common.Address
	modules     []ModuleInfo
	mu          sync.Mutex
}

// NewManager creates a smart account manager.
func NewManager(
	factory *Factory,
	bundlerClient *bundler.Client,
	caller contract.ContractCaller,
	wp wallet.WalletProvider,
	chainID int64,
	entryPoint common.Address,
) *Manager {
	return &Manager{
		factory:    factory,
		bundler:    bundlerClient,
		caller:     caller,
		wallet:     wp,
		chainID:    chainID,
		entryPoint: entryPoint,
		modules:    make([]ModuleInfo, 0),
	}
}

// GetOrDeploy returns the account info, deploying if needed.
func (m *Manager) GetOrDeploy(
	ctx context.Context,
) (*AccountInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ownerAddr, err := m.ownerAddress(ctx)
	if err != nil {
		return nil, err
	}

	// If we already have a cached account address, check deployment.
	if m.accountAddr != (common.Address{}) {
		deployed, err := m.factory.IsDeployed(
			ctx, m.accountAddr,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"check deployment: %w", err,
			)
		}
		if deployed {
			return m.buildInfo(ownerAddr, true), nil
		}
	}

	// Compute the deterministic address.
	salt := big.NewInt(0)
	computed := m.factory.ComputeAddress(ownerAddr, salt)

	// Check if already deployed at computed address.
	deployed, err := m.factory.IsDeployed(ctx, computed)
	if err != nil {
		return nil, fmt.Errorf(
			"check deployment: %w", err,
		)
	}
	if deployed {
		m.accountAddr = computed
		return m.buildInfo(ownerAddr, true), nil
	}

	// Deploy new account.
	addr, _, err := m.factory.Deploy(ctx, ownerAddr, salt)
	if err != nil {
		return nil, fmt.Errorf("deploy account: %w", err)
	}
	m.accountAddr = addr
	return m.buildInfo(ownerAddr, true), nil
}

// Info returns current account metadata without deploying.
func (m *Manager) Info(
	ctx context.Context,
) (*AccountInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ownerAddr, err := m.ownerAddress(ctx)
	if err != nil {
		return nil, err
	}

	if m.accountAddr == (common.Address{}) {
		// Compute deterministic address.
		salt := big.NewInt(0)
		m.accountAddr = m.factory.ComputeAddress(
			ownerAddr, salt,
		)
	}

	deployed, err := m.factory.IsDeployed(
		ctx, m.accountAddr,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"check deployment: %w", err,
		)
	}

	return m.buildInfo(ownerAddr, deployed), nil
}

// InstallModule installs an ERC-7579 module on the smart account.
func (m *Manager) InstallModule(
	ctx context.Context,
	moduleType ModuleType,
	addr common.Address,
	initData []byte,
) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.accountAddr == (common.Address{}) {
		return "", ErrAccountNotDeployed
	}

	// Check if module is already installed.
	for _, mod := range m.modules {
		if mod.Address == addr && mod.Type == moduleType {
			return "", ErrModuleAlreadyInstalled
		}
	}

	// Build installModule calldata via the Safe7579 adapter ABI.
	calldata, err := m.packSafe7579Call(
		"installModule",
		new(big.Int).SetUint64(uint64(moduleType)),
		addr,
		initData,
	)
	if err != nil {
		return "", fmt.Errorf(
			"encode install module: %w", err,
		)
	}

	txHash, err := m.submitUserOp(ctx, calldata)
	if err != nil {
		return "", fmt.Errorf(
			"install module %s: %w",
			moduleType.String(), err,
		)
	}

	// Track the module locally.
	m.modules = append(m.modules, ModuleInfo{
		Address:     addr,
		Type:        moduleType,
		Name:        moduleType.String(),
		InstalledAt: time.Now(),
	})

	return txHash, nil
}

// UninstallModule removes a module from the smart account.
func (m *Manager) UninstallModule(
	ctx context.Context,
	moduleType ModuleType,
	addr common.Address,
	deInitData []byte,
) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.accountAddr == (common.Address{}) {
		return "", ErrAccountNotDeployed
	}

	// Check that the module is installed.
	found := false
	for _, mod := range m.modules {
		if mod.Address == addr && mod.Type == moduleType {
			found = true
			break
		}
	}
	if !found {
		return "", ErrModuleNotInstalled
	}

	calldata, err := m.packSafe7579Call(
		"uninstallModule",
		new(big.Int).SetUint64(uint64(moduleType)),
		addr,
		deInitData,
	)
	if err != nil {
		return "", fmt.Errorf(
			"encode uninstall module: %w", err,
		)
	}

	txHash, err := m.submitUserOp(ctx, calldata)
	if err != nil {
		return "", fmt.Errorf(
			"uninstall module %s: %w",
			moduleType.String(), err,
		)
	}

	// Remove from local tracking.
	filtered := make([]ModuleInfo, 0, len(m.modules))
	for _, mod := range m.modules {
		if mod.Address == addr && mod.Type == moduleType {
			continue
		}
		filtered = append(filtered, mod)
	}
	m.modules = filtered

	return txHash, nil
}

// Execute builds and submits a UserOp for contract calls.
func (m *Manager) Execute(
	ctx context.Context,
	calls []ContractCall,
) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.accountAddr == (common.Address{}) {
		return "", ErrAccountNotDeployed
	}
	if len(calls) == 0 {
		return "", fmt.Errorf(
			"execute: %w", ErrInvalidUserOp,
		)
	}

	calldata, err := m.encodeCalls(calls)
	if err != nil {
		return "", fmt.Errorf(
			"encode calls: %w", err,
		)
	}

	txHash, err := m.submitUserOp(ctx, calldata)
	if err != nil {
		return "", fmt.Errorf("execute calls: %w", err)
	}

	return txHash, nil
}

// submitUserOp constructs a UserOp, estimates gas, signs it,
// and submits it via the bundler.
func (m *Manager) submitUserOp(
	ctx context.Context,
	calldata []byte,
) (string, error) {
	op := &UserOperation{
		Sender:               m.accountAddr,
		Nonce:                big.NewInt(0),
		InitCode:             []byte{},
		CallData:             calldata,
		CallGasLimit:         big.NewInt(0),
		VerificationGasLimit: big.NewInt(0),
		PreVerificationGas:   big.NewInt(0),
		MaxFeePerGas:         big.NewInt(0),
		MaxPriorityFeePerGas: big.NewInt(0),
		PaymasterAndData:     []byte{},
		Signature:            []byte{},
	}

	// Estimate gas via bundler.
	bOp := toBundlerOp(op)
	gasEstimate, err := m.bundler.EstimateGas(ctx, bOp)
	if err != nil {
		return "", fmt.Errorf("estimate gas: %w", err)
	}
	op.CallGasLimit = gasEstimate.CallGasLimit
	op.VerificationGasLimit = gasEstimate.VerificationGasLimit
	op.PreVerificationGas = gasEstimate.PreVerificationGas

	// Compute the UserOp hash for signing.
	opHash := m.computeUserOpHash(op)

	// Sign with wallet.
	sig, err := m.wallet.SignMessage(ctx, opHash)
	if err != nil {
		return "", fmt.Errorf("sign user op: %w", err)
	}
	op.Signature = sig

	// Submit via bundler.
	bOp = toBundlerOp(op)
	result, err := m.bundler.SendUserOperation(ctx, bOp)
	if err != nil {
		return "", fmt.Errorf("submit user op: %w", err)
	}

	return result.UserOpHash.Hex(), nil
}

// computeUserOpHash computes the hash of a UserOp for signing.
// The hash is keccak256(abi.encode(userOpHash, entryPoint, chainId)).
func (m *Manager) computeUserOpHash(
	op *UserOperation,
) []byte {
	// Pack UserOp fields (simplified hash for signature).
	packed := make([]byte, 0, 256)
	packed = append(packed, op.Sender.Bytes()...)
	if op.Nonce != nil {
		nonceBytes := op.Nonce.Bytes()
		padded := make([]byte, 32)
		copy(padded[32-len(nonceBytes):], nonceBytes)
		packed = append(packed, padded...)
	}
	packed = append(
		packed, crypto.Keccak256(op.InitCode)...,
	)
	packed = append(
		packed, crypto.Keccak256(op.CallData)...,
	)

	// Add gas parameters.
	for _, v := range []*big.Int{
		op.CallGasLimit,
		op.VerificationGasLimit,
		op.PreVerificationGas,
		op.MaxFeePerGas,
		op.MaxPriorityFeePerGas,
	} {
		padded := make([]byte, 32)
		if v != nil {
			b := v.Bytes()
			copy(padded[32-len(b):], b)
		}
		packed = append(packed, padded...)
	}
	packed = append(
		packed, crypto.Keccak256(op.PaymasterAndData)...,
	)

	innerHash := crypto.Keccak256(packed)

	// Final hash: keccak256(innerHash ++ entryPoint ++ chainId)
	final := make([]byte, 0, 84)
	final = append(final, innerHash...)
	// Left-pad entryPoint to 32 bytes.
	epPadded := make([]byte, 32)
	copy(epPadded[12:], m.entryPoint.Bytes())
	final = append(final, epPadded...)
	// Left-pad chainID to 32 bytes.
	chainIDBytes := big.NewInt(m.chainID).Bytes()
	chainPadded := make([]byte, 32)
	copy(
		chainPadded[32-len(chainIDBytes):],
		chainIDBytes,
	)
	final = append(final, chainPadded...)

	return crypto.Keccak256(final)
}

// packSafe7579Call encodes a call to the Safe7579 adapter contract.
func (m *Manager) packSafe7579Call(
	method string,
	args ...interface{},
) ([]byte, error) {
	parsed, err := contract.ParseABI(Safe7579ABI)
	if err != nil {
		return nil, fmt.Errorf("parse Safe7579 ABI: %w", err)
	}
	data, err := parsed.Pack(method, args...)
	if err != nil {
		return nil, fmt.Errorf(
			"pack %s call: %w", method, err,
		)
	}
	return data, nil
}

// encodeCalls encodes contract calls into Safe7579 execute calldata.
// Single calls use the single execution mode; multiple calls use
// batch execution mode.
func (m *Manager) encodeCalls(
	calls []ContractCall,
) ([]byte, error) {
	if len(calls) == 1 {
		return m.encodeSingleCall(calls[0])
	}
	return m.encodeBatchCalls(calls)
}

// encodeSingleCall encodes a single call for Safe7579 execute.
func (m *Manager) encodeSingleCall(
	call ContractCall,
) ([]byte, error) {
	// ERC-7579 single execution mode: 0x00 (left-padded to 32 bytes).
	mode := make([]byte, 32)

	// Execution calldata: abi.encodePacked(target, value, calldata)
	value := call.Value
	if value == nil {
		value = new(big.Int)
	}
	valuePadded := make([]byte, 32)
	vBytes := value.Bytes()
	copy(valuePadded[32-len(vBytes):], vBytes)

	execData := make([]byte, 0, 52+len(call.Data))
	execData = append(execData, call.Target.Bytes()...)
	execData = append(execData, valuePadded...)
	execData = append(execData, call.Data...)

	parsed, err := contract.ParseABI(Safe7579ABI)
	if err != nil {
		return nil, fmt.Errorf(
			"parse Safe7579 ABI: %w", err,
		)
	}
	return parsed.Pack("execute", [32]byte(mode), execData)
}

// encodeBatchCalls encodes multiple calls for Safe7579 executeBatch.
func (m *Manager) encodeBatchCalls(
	calls []ContractCall,
) ([]byte, error) {
	// ERC-7579 batch execution mode: 0x01 at byte 0
	// (left-padded to 32 bytes).
	mode := make([]byte, 32)
	mode[0] = 0x01

	// Batch calldata: abi.encode(Execution[])
	// Each Execution: (address target, uint256 value, bytes calldata)
	batchData := make([]byte, 0, len(calls)*84)
	for _, call := range calls {
		// Target address (20 bytes, left-padded to 32).
		targetPadded := make([]byte, 32)
		copy(targetPadded[12:], call.Target.Bytes())
		batchData = append(batchData, targetPadded...)

		// Value (32 bytes).
		value := call.Value
		if value == nil {
			value = new(big.Int)
		}
		valuePadded := make([]byte, 32)
		vBytes := value.Bytes()
		copy(valuePadded[32-len(vBytes):], vBytes)
		batchData = append(batchData, valuePadded...)

		// Calldata with length prefix.
		lenPadded := make([]byte, 32)
		lenBytes := big.NewInt(
			int64(len(call.Data)),
		).Bytes()
		copy(lenPadded[32-len(lenBytes):], lenBytes)
		batchData = append(batchData, lenPadded...)
		batchData = append(batchData, call.Data...)
	}

	parsed, err := contract.ParseABI(Safe7579ABI)
	if err != nil {
		return nil, fmt.Errorf(
			"parse Safe7579 ABI: %w", err,
		)
	}
	return parsed.Pack(
		"execute", [32]byte(mode), batchData,
	)
}

// ownerAddress gets the owner address from the wallet provider.
func (m *Manager) ownerAddress(
	ctx context.Context,
) (common.Address, error) {
	addrStr, err := m.wallet.Address(ctx)
	if err != nil {
		return common.Address{},
			fmt.Errorf("get owner address: %w", err)
	}
	return common.HexToAddress(addrStr), nil
}

// buildInfo constructs AccountInfo from current state.
func (m *Manager) buildInfo(
	ownerAddr common.Address,
	deployed bool,
) *AccountInfo {
	modules := make([]ModuleInfo, len(m.modules))
	copy(modules, m.modules)
	return &AccountInfo{
		Address:      m.accountAddr,
		IsDeployed:   deployed,
		Modules:      modules,
		OwnerAddress: ownerAddr,
		ChainID:      m.chainID,
		EntryPoint:   m.entryPoint,
	}
}

// toBundlerOp converts a smartaccount.UserOperation to
// bundler.UserOperation to avoid import cycles.
func toBundlerOp(op *UserOperation) *bundler.UserOperation {
	return &bundler.UserOperation{
		Sender:               op.Sender,
		Nonce:                op.Nonce,
		InitCode:             op.InitCode,
		CallData:             op.CallData,
		CallGasLimit:         op.CallGasLimit,
		VerificationGasLimit: op.VerificationGasLimit,
		PreVerificationGas:   op.PreVerificationGas,
		MaxFeePerGas:         op.MaxFeePerGas,
		MaxPriorityFeePerGas: op.MaxPriorityFeePerGas,
		PaymasterAndData:     op.PaymasterAndData,
		Signature:            op.Signature,
	}
}
