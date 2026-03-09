package smartaccount

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/contract"
)

// stubContractCaller implements contract.ContractCaller for testing.
type stubContractCaller struct {
	readResult  *contract.ContractCallResult
	writeResult *contract.ContractCallResult
	readErr     error
	writeErr    error
	writeCalls  int
	readCalls   int
	lastWrite   contract.ContractCallRequest
	lastRead    contract.ContractCallRequest
}

func (s *stubContractCaller) Read(
	_ context.Context, req contract.ContractCallRequest,
) (*contract.ContractCallResult, error) {
	s.readCalls++
	s.lastRead = req
	if s.readErr != nil {
		return nil, s.readErr
	}
	return s.readResult, nil
}

func (s *stubContractCaller) Write(
	_ context.Context, req contract.ContractCallRequest,
) (*contract.ContractCallResult, error) {
	s.writeCalls++
	s.lastWrite = req
	if s.writeErr != nil {
		return nil, s.writeErr
	}
	return s.writeResult, nil
}

func newTestFactory(caller contract.ContractCaller) *Factory {
	return NewFactory(
		caller,
		common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
		common.HexToAddress("0xCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC"),
		84532,
	)
}

func TestComputeAddress_Deterministic(t *testing.T) {
	t.Parallel()

	f := newTestFactory(nil)
	owner := common.HexToAddress(
		"0x1234567890abcdef1234567890abcdef12345678",
	)

	tests := []struct {
		give     string
		giveSalt *big.Int
	}{
		{give: "salt=0", giveSalt: big.NewInt(0)},
		{give: "salt=1", giveSalt: big.NewInt(1)},
		{give: "salt=large", giveSalt: big.NewInt(999999)},
		{give: "salt=nil", giveSalt: nil},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			addr1 := f.ComputeAddress(owner, tt.giveSalt)
			addr2 := f.ComputeAddress(owner, tt.giveSalt)

			assert.Equal(t, addr1, addr2,
				"same inputs must produce same address")
			assert.NotEqual(t, common.Address{}, addr1,
				"address must not be zero")
		})
	}
}

func TestComputeAddress_DifferentSaltsDifferentAddresses(t *testing.T) {
	t.Parallel()

	f := newTestFactory(nil)
	owner := common.HexToAddress(
		"0x1234567890abcdef1234567890abcdef12345678",
	)

	addr0 := f.ComputeAddress(owner, big.NewInt(0))
	addr1 := f.ComputeAddress(owner, big.NewInt(1))
	addr2 := f.ComputeAddress(owner, big.NewInt(2))

	assert.NotEqual(t, addr0, addr1, "salt 0 vs 1")
	assert.NotEqual(t, addr1, addr2, "salt 1 vs 2")
	assert.NotEqual(t, addr0, addr2, "salt 0 vs 2")
}

func TestComputeAddress_DifferentOwnersDifferentAddresses(t *testing.T) {
	t.Parallel()

	f := newTestFactory(nil)
	salt := big.NewInt(0)

	ownerA := common.HexToAddress(
		"0x1111111111111111111111111111111111111111",
	)
	ownerB := common.HexToAddress(
		"0x2222222222222222222222222222222222222222",
	)

	addrA := f.ComputeAddress(ownerA, salt)
	addrB := f.ComputeAddress(ownerB, salt)

	assert.NotEqual(t, addrA, addrB,
		"different owners must produce different addresses")
}

func TestComputeAddress_DifferentFactoryAddresses(t *testing.T) {
	t.Parallel()

	owner := common.HexToAddress(
		"0x1234567890abcdef1234567890abcdef12345678",
	)
	salt := big.NewInt(0)

	f1 := NewFactory(
		nil,
		common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
		common.HexToAddress("0xCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC"),
		84532,
	)
	f2 := NewFactory(
		nil,
		common.HexToAddress("0xDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD"),
		common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
		common.HexToAddress("0xCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC"),
		84532,
	)

	addr1 := f1.ComputeAddress(owner, salt)
	addr2 := f2.ComputeAddress(owner, salt)

	assert.NotEqual(t, addr1, addr2,
		"different factory addresses must produce different addresses")
}

func TestComputeAddress_NilSaltEqualsZeroSalt(t *testing.T) {
	t.Parallel()

	f := newTestFactory(nil)
	owner := common.HexToAddress(
		"0x1234567890abcdef1234567890abcdef12345678",
	)

	addrNil := f.ComputeAddress(owner, nil)
	addrZero := f.ComputeAddress(owner, big.NewInt(0))

	assert.Equal(t, addrNil, addrZero,
		"nil salt and zero salt must produce the same address")
}

func TestBuildSafeInitializer_NotNil(t *testing.T) {
	t.Parallel()

	owner := common.HexToAddress(
		"0x1234567890abcdef1234567890abcdef12345678",
	)
	safe7579 := common.HexToAddress(
		"0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
	)
	fallback := common.HexToAddress(
		"0xCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC",
	)

	data := buildSafeInitializer(owner, safe7579, fallback)
	require.NotNil(t, data, "initializer must not be nil")
	assert.True(t, len(data) > 4,
		"initializer must contain function selector + params")
}

func TestBuildSafeInitializer_Deterministic(t *testing.T) {
	t.Parallel()

	owner := common.HexToAddress(
		"0x1234567890abcdef1234567890abcdef12345678",
	)
	safe7579 := common.HexToAddress(
		"0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
	)
	fallback := common.HexToAddress(
		"0xCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC",
	)

	data1 := buildSafeInitializer(owner, safe7579, fallback)
	data2 := buildSafeInitializer(owner, safe7579, fallback)

	assert.Equal(t, data1, data2,
		"same inputs must produce identical initializer data")
}

func TestBuildSafeInitializer_DifferentOwners(t *testing.T) {
	t.Parallel()

	safe7579 := common.HexToAddress(
		"0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
	)
	fallback := common.HexToAddress(
		"0xCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC",
	)

	dataA := buildSafeInitializer(
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
		safe7579, fallback,
	)
	dataB := buildSafeInitializer(
		common.HexToAddress("0x2222222222222222222222222222222222222222"),
		safe7579, fallback,
	)

	assert.NotEqual(t, dataA, dataB,
		"different owners must produce different initializer data")
}

func TestDeploy_Success(t *testing.T) {
	t.Parallel()

	deployedAddr := common.HexToAddress(
		"0xDeployedDeployedDeployedDeployedDeployed",
	)
	caller := &stubContractCaller{
		writeResult: &contract.ContractCallResult{
			Data:   []interface{}{deployedAddr},
			TxHash: "0xabc123",
		},
	}

	f := newTestFactory(caller)
	owner := common.HexToAddress(
		"0x1234567890abcdef1234567890abcdef12345678",
	)

	addr, txHash, err := f.Deploy(context.Background(), owner, big.NewInt(0))
	require.NoError(t, err)
	assert.Equal(t, deployedAddr, addr)
	assert.Equal(t, "0xabc123", txHash)
	assert.Equal(t, 1, caller.writeCalls)
	assert.Equal(t, "createProxyWithNonce", caller.lastWrite.Method)
}

func TestDeploy_FallsBackToComputedAddress(t *testing.T) {
	t.Parallel()

	caller := &stubContractCaller{
		writeResult: &contract.ContractCallResult{
			Data:   []interface{}{}, // empty data — no address returned
			TxHash: "0xdef456",
		},
	}

	f := newTestFactory(caller)
	owner := common.HexToAddress(
		"0x1234567890abcdef1234567890abcdef12345678",
	)
	salt := big.NewInt(42)

	addr, txHash, err := f.Deploy(context.Background(), owner, salt)
	require.NoError(t, err)
	assert.Equal(t, "0xdef456", txHash)

	// Should fall back to computed address.
	expected := f.ComputeAddress(owner, salt)
	assert.Equal(t, expected, addr)
}

func TestDeploy_NilSaltDefaultsToZero(t *testing.T) {
	t.Parallel()

	caller := &stubContractCaller{
		writeResult: &contract.ContractCallResult{
			Data:   []interface{}{},
			TxHash: "0x111",
		},
	}

	f := newTestFactory(caller)
	owner := common.HexToAddress(
		"0x1234567890abcdef1234567890abcdef12345678",
	)

	_, _, err := f.Deploy(context.Background(), owner, nil)
	require.NoError(t, err)

	// Verify the salt argument passed was big.NewInt(0), not nil.
	args := caller.lastWrite.Args
	require.Len(t, args, 3)
	saltArg, ok := args[2].(*big.Int)
	require.True(t, ok, "third arg must be *big.Int")
	assert.Equal(t, big.NewInt(0), saltArg)
}

func TestDeploy_WriteError(t *testing.T) {
	t.Parallel()

	caller := &stubContractCaller{
		writeErr: errors.New("rpc unavailable"),
	}

	f := newTestFactory(caller)
	owner := common.HexToAddress(
		"0x1234567890abcdef1234567890abcdef12345678",
	)

	_, _, err := f.Deploy(context.Background(), owner, big.NewInt(0))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deploy safe account")
	assert.ErrorIs(t, err, caller.writeErr)
}

func TestIsDeployed_True(t *testing.T) {
	t.Parallel()

	caller := &stubContractCaller{
		readResult: &contract.ContractCallResult{
			Data: []interface{}{false},
		},
	}

	f := newTestFactory(caller)
	addr := common.HexToAddress("0xABCDABCDABCDABCDABCDABCDABCDABCDABCDABCD")

	deployed, err := f.IsDeployed(context.Background(), addr)
	require.NoError(t, err)
	assert.True(t, deployed)
	assert.Equal(t, 1, caller.readCalls)
}

func TestIsDeployed_False(t *testing.T) {
	t.Parallel()

	caller := &stubContractCaller{
		readErr: errors.New("execution reverted"),
	}

	f := newTestFactory(caller)
	addr := common.HexToAddress("0xABCDABCDABCDABCDABCDABCDABCDABCDABCDABCD")

	deployed, err := f.IsDeployed(context.Background(), addr)
	require.NoError(t, err, "read error should not propagate")
	assert.False(t, deployed)
}

func TestNewFactory(t *testing.T) {
	t.Parallel()

	caller := &stubContractCaller{}
	factoryAddr := common.HexToAddress("0xFACE")
	safe7579 := common.HexToAddress("0x7579")
	fallback := common.HexToAddress("0xFB00")

	f := NewFactory(caller, factoryAddr, safe7579, fallback, 1)
	require.NotNil(t, f)
	assert.Equal(t, factoryAddr, f.factoryAddr)
	assert.Equal(t, safe7579, f.safe7579Addr)
	assert.Equal(t, fallback, f.fallbackAddr)
	assert.Equal(t, int64(1), f.chainID)
}
