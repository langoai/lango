//go:build integration

package hub

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/contract"
	"github.com/langoai/lango/internal/eventbus"
)

const anvilRPC = "http://127.0.0.1:8545"

// Anvil pre-funded accounts (default mnemonic).
var (
	// Account 0: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
	anvilKey0, _ = crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	// Account 1: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
	anvilKey1, _ = crypto.HexToECDSA("59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d")
	// Account 2: 0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC
	anvilKey2, _ = crypto.HexToECDSA("5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a")
)

type testWallet struct {
	key  *ecdsa.PrivateKey
	addr common.Address
}

func newTestWallet(key *ecdsa.PrivateKey) *testWallet {
	return &testWallet{
		key:  key,
		addr: crypto.PubkeyToAddress(key.PublicKey),
	}
}

func (w *testWallet) Address(_ context.Context) (string, error) {
	return w.addr.Hex(), nil
}

func (w *testWallet) SignTransaction(_ context.Context, hash []byte) ([]byte, error) {
	return crypto.Sign(hash, w.key)
}

func (w *testWallet) Balance(_ context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (w *testWallet) SignMessage(_ context.Context, message []byte) ([]byte, error) {
	hash := crypto.Keccak256(message)
	return crypto.Sign(hash, w.key)
}

func (w *testWallet) PublicKey(_ context.Context) ([]byte, error) {
	return crypto.CompressPubkey(&w.key.PublicKey), nil
}

// deployContract deploys raw bytecode and returns the contract address.
func deployContract(t *testing.T, client *ethclient.Client, key *ecdsa.PrivateKey, bytecodeData []byte, chainID *big.Int) common.Address {
	t.Helper()
	ctx := context.Background()

	from := crypto.PubkeyToAddress(key.PublicKey)
	nonce, err := client.PendingNonceAt(ctx, from)
	require.NoError(t, err)

	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From: from,
		Data: bytecodeData,
	})
	require.NoError(t, err)

	header, err := client.HeaderByNumber(ctx, nil)
	require.NoError(t, err)
	baseFee := header.BaseFee
	if baseFee == nil {
		baseFee = big.NewInt(1000000000)
	}
	maxPriorityFee := big.NewInt(1500000000)
	maxFee := new(big.Int).Add(new(big.Int).Mul(baseFee, big.NewInt(2)), maxPriorityFee)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasFeeCap: maxFee,
		GasTipCap: maxPriorityFee,
		Gas:       gasLimit * 2,
		Data:      bytecodeData,
	})

	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, key)
	require.NoError(t, err)

	err = client.SendTransaction(ctx, signedTx)
	require.NoError(t, err)

	receipt := waitReceipt(t, client, signedTx.Hash())
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	return receipt.ContractAddress
}

func waitReceipt(t *testing.T, client *ethclient.Client, txHash common.Hash) *types.Receipt {
	t.Helper()
	ctx := context.Background()
	for i := 0; i < 30; i++ {
		receipt, err := client.TransactionReceipt(ctx, txHash)
		if err == nil {
			return receipt
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("receipt timeout for %s", txHash.Hex())
	return nil
}

func mustABIType(typStr string) abi.Type {
	typ, _ := abi.NewType(typStr, "", nil)
	return typ
}

// getBytecodeFromForgeOutput reads the creation bytecode from forge's compiled output.
func getBytecodeFromForgeOutput(t *testing.T, contractName string) string {
	t.Helper()
	path := "../../../../contracts/out/" + contractName + ".sol/" + contractName + ".json"
	data, err := os.ReadFile(path)
	require.NoError(t, err, "forge build output not found; run: cd contracts && forge build")

	var artifact struct {
		Bytecode struct {
			Object string `json:"object"`
		} `json:"bytecode"`
	}
	err = json.Unmarshal(data, &artifact)
	require.NoError(t, err)

	return strings.TrimPrefix(artifact.Bytecode.Object, "0x")
}

func deployMockUSDC(t *testing.T, client *ethclient.Client, key *ecdsa.PrivateKey, chainID *big.Int) common.Address {
	t.Helper()
	bc, err := hex.DecodeString(getBytecodeFromForgeOutput(t, "MockUSDC"))
	require.NoError(t, err)
	return deployContract(t, client, key, bc, chainID)
}

func deployHub(t *testing.T, client *ethclient.Client, key *ecdsa.PrivateKey, chainID *big.Int, arbitrator common.Address) common.Address {
	t.Helper()
	bc, err := hex.DecodeString(getBytecodeFromForgeOutput(t, "LangoEscrowHub"))
	require.NoError(t, err)

	arg, err := abi.Arguments{{Type: mustABIType("address")}}.Pack(arbitrator)
	require.NoError(t, err)
	bc = append(bc, arg...)

	return deployContract(t, client, key, bc, chainID)
}

func deployVaultImpl(t *testing.T, client *ethclient.Client, key *ecdsa.PrivateKey, chainID *big.Int) common.Address {
	t.Helper()
	bc, err := hex.DecodeString(getBytecodeFromForgeOutput(t, "LangoVault"))
	require.NoError(t, err)
	return deployContract(t, client, key, bc, chainID)
}

func deployFactory(t *testing.T, client *ethclient.Client, key *ecdsa.PrivateKey, chainID *big.Int, impl common.Address) common.Address {
	t.Helper()
	bc, err := hex.DecodeString(getBytecodeFromForgeOutput(t, "LangoVaultFactory"))
	require.NoError(t, err)

	arg, err := abi.Arguments{{Type: mustABIType("address")}}.Pack(impl)
	require.NoError(t, err)
	bc = append(bc, arg...)

	return deployContract(t, client, key, bc, chainID)
}

func mintUSDC(t *testing.T, caller *contract.Caller, usdcAddr, to common.Address, amount *big.Int, chainID int64) {
	t.Helper()
	mintABI := `[{"inputs":[{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],"name":"mint","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
	_, err := caller.Write(context.Background(), contract.ContractCallRequest{
		ChainID: chainID,
		Address: usdcAddr,
		ABI:     mintABI,
		Method:  "mint",
		Args:    []interface{}{to, amount},
	})
	require.NoError(t, err)
}

func approveUSDC(t *testing.T, caller *contract.Caller, usdcAddr, spender common.Address, amount *big.Int, chainID int64) {
	t.Helper()
	approveABI := `[{"inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"}]`
	_, err := caller.Write(context.Background(), contract.ContractCallRequest{
		ChainID: chainID,
		Address: usdcAddr,
		ABI:     approveABI,
		Method:  "approve",
		Args:    []interface{}{spender, amount},
	})
	require.NoError(t, err)
}

// increaseTime advances the Anvil EVM time via JSON-RPC.
func increaseTime(t *testing.T, _ *ethclient.Client, seconds int) {
	t.Helper()

	payload := fmt.Sprintf(`{"jsonrpc":"2.0","method":"evm_increaseTime","params":[%d],"id":1}`, seconds)
	resp, err := http.Post(anvilRPC, "application/json", bytes.NewBufferString(payload))
	require.NoError(t, err)
	resp.Body.Close()

	minePayload := `{"jsonrpc":"2.0","method":"evm_mine","params":[],"id":2}`
	resp2, err := http.Post(anvilRPC, "application/json", bytes.NewBufferString(minePayload))
	require.NoError(t, err)
	resp2.Body.Close()
}

// ---- Integration Tests ----

func TestHubIntegration_FullLifecycle(t *testing.T) {
	ctx := context.Background()
	client, err := ethclient.Dial(anvilRPC)
	require.NoError(t, err)
	defer client.Close()

	chainID := big.NewInt(31337)
	cache := contract.NewABICache()

	buyerWallet := newTestWallet(anvilKey0)
	sellerWallet := newTestWallet(anvilKey1)
	arbitratorAddr := crypto.PubkeyToAddress(anvilKey2.PublicKey)

	buyerCaller := contract.NewCaller(client, buyerWallet, 31337, cache)

	usdcAddr := deployMockUSDC(t, client, anvilKey0, chainID)
	hubAddr := deployHub(t, client, anvilKey0, chainID, arbitratorAddr)

	mintUSDC(t, buyerCaller, usdcAddr, buyerWallet.addr, big.NewInt(10_000_000_000), 31337)
	approveUSDC(t, buyerCaller, usdcAddr, hubAddr, big.NewInt(10_000_000_000), 31337)

	hubClient := NewHubClient(buyerCaller, hubAddr, 31337)
	dl := big.NewInt(time.Now().Unix() + 86400)
	dealID, txHash, err := hubClient.CreateDeal(ctx, sellerWallet.addr, usdcAddr, big.NewInt(1000_000_000), dl)
	require.NoError(t, err)
	assert.NotEmpty(t, txHash)
	assert.Equal(t, big.NewInt(0), dealID)

	_, err = hubClient.Deposit(ctx, dealID)
	require.NoError(t, err)

	sellerCaller := contract.NewCaller(client, sellerWallet, 31337, cache)
	sellerHub := NewHubClient(sellerCaller, hubAddr, 31337)
	var workHash [32]byte
	copy(workHash[:], crypto.Keccak256([]byte("integration test result")))
	_, err = sellerHub.SubmitWork(ctx, dealID, workHash)
	require.NoError(t, err)

	_, err = hubClient.Release(ctx, dealID)
	require.NoError(t, err)

	deal, err := hubClient.GetDeal(ctx, dealID)
	require.NoError(t, err)
	assert.Equal(t, DealStatusReleased, deal.Status)
}

func TestHubIntegration_DisputeAndResolve(t *testing.T) {
	ctx := context.Background()
	client, err := ethclient.Dial(anvilRPC)
	require.NoError(t, err)
	defer client.Close()

	chainID := big.NewInt(31337)
	cache := contract.NewABICache()

	buyerWallet := newTestWallet(anvilKey0)
	sellerWallet := newTestWallet(anvilKey1)
	arbitratorWallet := newTestWallet(anvilKey2)

	buyerCaller := contract.NewCaller(client, buyerWallet, 31337, cache)

	usdcAddr := deployMockUSDC(t, client, anvilKey0, chainID)
	hubAddr := deployHub(t, client, anvilKey0, chainID, arbitratorWallet.addr)

	amount := big.NewInt(2000_000_000)
	mintUSDC(t, buyerCaller, usdcAddr, buyerWallet.addr, big.NewInt(10_000_000_000), 31337)
	approveUSDC(t, buyerCaller, usdcAddr, hubAddr, big.NewInt(10_000_000_000), 31337)

	hubClient := NewHubClient(buyerCaller, hubAddr, 31337)
	dl := big.NewInt(time.Now().Unix() + 86400)
	dealID, _, err := hubClient.CreateDeal(ctx, sellerWallet.addr, usdcAddr, amount, dl)
	require.NoError(t, err)

	_, err = hubClient.Deposit(ctx, dealID)
	require.NoError(t, err)

	_, err = hubClient.Dispute(ctx, dealID)
	require.NoError(t, err)

	arbCaller := contract.NewCaller(client, arbitratorWallet, 31337, cache)
	arbHub := NewHubClient(arbCaller, hubAddr, 31337)
	sellerAmt := big.NewInt(1200_000_000)
	buyerAmt := big.NewInt(800_000_000)
	_, err = arbHub.ResolveDispute(ctx, dealID, true, sellerAmt, buyerAmt)
	require.NoError(t, err)

	deal, err := hubClient.GetDeal(ctx, dealID)
	require.NoError(t, err)
	assert.Equal(t, DealStatusResolved, deal.Status)
}

func TestHubIntegration_RefundAfterDeadline(t *testing.T) {
	ctx := context.Background()
	client, err := ethclient.Dial(anvilRPC)
	require.NoError(t, err)
	defer client.Close()

	chainID := big.NewInt(31337)
	cache := contract.NewABICache()

	buyerWallet := newTestWallet(anvilKey0)
	sellerWallet := newTestWallet(anvilKey1)
	arbitratorAddr := crypto.PubkeyToAddress(anvilKey2.PublicKey)

	buyerCaller := contract.NewCaller(client, buyerWallet, 31337, cache)

	usdcAddr := deployMockUSDC(t, client, anvilKey0, chainID)
	hubAddr := deployHub(t, client, anvilKey0, chainID, arbitratorAddr)

	mintUSDC(t, buyerCaller, usdcAddr, buyerWallet.addr, big.NewInt(10_000_000_000), 31337)
	approveUSDC(t, buyerCaller, usdcAddr, hubAddr, big.NewInt(10_000_000_000), 31337)

	hubClient := NewHubClient(buyerCaller, hubAddr, 31337)
	dl := big.NewInt(time.Now().Unix() + 60)
	dealID, _, err := hubClient.CreateDeal(ctx, sellerWallet.addr, usdcAddr, big.NewInt(500_000_000), dl)
	require.NoError(t, err)

	_, err = hubClient.Deposit(ctx, dealID)
	require.NoError(t, err)

	increaseTime(t, client, 120)

	_, err = hubClient.Refund(ctx, dealID)
	require.NoError(t, err)

	deal, err := hubClient.GetDeal(ctx, dealID)
	require.NoError(t, err)
	assert.Equal(t, DealStatusRefunded, deal.Status)
}

func TestVaultIntegration_FullLifecycle(t *testing.T) {
	ctx := context.Background()
	client, err := ethclient.Dial(anvilRPC)
	require.NoError(t, err)
	defer client.Close()

	chainID := big.NewInt(31337)
	cache := contract.NewABICache()

	buyerWallet := newTestWallet(anvilKey0)
	sellerWallet := newTestWallet(anvilKey1)
	arbitratorAddr := crypto.PubkeyToAddress(anvilKey2.PublicKey)

	buyerCaller := contract.NewCaller(client, buyerWallet, 31337, cache)

	usdcAddr := deployMockUSDC(t, client, anvilKey0, chainID)
	implAddr := deployVaultImpl(t, client, anvilKey0, chainID)
	factoryAddr := deployFactory(t, client, anvilKey0, chainID, implAddr)

	amount := big.NewInt(1000_000_000)
	mintUSDC(t, buyerCaller, usdcAddr, buyerWallet.addr, big.NewInt(10_000_000_000), 31337)

	factoryClient := NewFactoryClient(buyerCaller, factoryAddr, 31337)
	dl := big.NewInt(time.Now().Unix() + 86400)
	info, _, err := factoryClient.CreateVault(ctx, sellerWallet.addr, usdcAddr, amount, dl, arbitratorAddr)
	require.NoError(t, err)
	require.NotEqual(t, common.Address{}, info.VaultAddress)

	approveUSDC(t, buyerCaller, usdcAddr, info.VaultAddress, amount, 31337)

	vaultClient := NewVaultClient(buyerCaller, info.VaultAddress, 31337)
	_, err = vaultClient.Deposit(ctx)
	require.NoError(t, err)

	sellerCaller := contract.NewCaller(client, sellerWallet, 31337, cache)
	sellerVault := NewVaultClient(sellerCaller, info.VaultAddress, 31337)
	var wh [32]byte
	copy(wh[:], crypto.Keccak256([]byte("vault work")))
	_, err = sellerVault.SubmitWork(ctx, wh)
	require.NoError(t, err)

	_, err = vaultClient.Release(ctx)
	require.NoError(t, err)

	status, err := vaultClient.Status(ctx)
	require.NoError(t, err)
	// Vault status enum: Released = 4
	assert.Equal(t, OnChainDealStatus(4), status)
}

func TestVaultIntegration_DisputeAndResolve(t *testing.T) {
	ctx := context.Background()
	client, err := ethclient.Dial(anvilRPC)
	require.NoError(t, err)
	defer client.Close()

	chainID := big.NewInt(31337)
	cache := contract.NewABICache()

	buyerWallet := newTestWallet(anvilKey0)
	sellerWallet := newTestWallet(anvilKey1)
	arbitratorWallet := newTestWallet(anvilKey2)

	buyerCaller := contract.NewCaller(client, buyerWallet, 31337, cache)

	usdcAddr := deployMockUSDC(t, client, anvilKey0, chainID)
	implAddr := deployVaultImpl(t, client, anvilKey0, chainID)
	factoryAddr := deployFactory(t, client, anvilKey0, chainID, implAddr)

	amount := big.NewInt(1000_000_000)
	mintUSDC(t, buyerCaller, usdcAddr, buyerWallet.addr, big.NewInt(10_000_000_000), 31337)

	factoryClient := NewFactoryClient(buyerCaller, factoryAddr, 31337)
	dl := big.NewInt(time.Now().Unix() + 86400)
	info, _, err := factoryClient.CreateVault(ctx, sellerWallet.addr, usdcAddr, amount, dl, arbitratorWallet.addr)
	require.NoError(t, err)

	approveUSDC(t, buyerCaller, usdcAddr, info.VaultAddress, amount, 31337)

	vaultClient := NewVaultClient(buyerCaller, info.VaultAddress, 31337)
	_, err = vaultClient.Deposit(ctx)
	require.NoError(t, err)

	_, err = vaultClient.Dispute(ctx)
	require.NoError(t, err)

	arbCaller := contract.NewCaller(client, arbitratorWallet, 31337, cache)
	arbVault := NewVaultClient(arbCaller, info.VaultAddress, 31337)
	_, err = arbVault.Resolve(ctx, true, big.NewInt(700_000_000), big.NewInt(300_000_000))
	require.NoError(t, err)

	status, err := vaultClient.Status(ctx)
	require.NoError(t, err)
	// Vault status enum: Resolved = 7
	assert.Equal(t, OnChainDealStatus(7), status)
}

func TestFactoryIntegration_MultipleVaults(t *testing.T) {
	ctx := context.Background()
	client, err := ethclient.Dial(anvilRPC)
	require.NoError(t, err)
	defer client.Close()

	chainID := big.NewInt(31337)
	cache := contract.NewABICache()

	buyerWallet := newTestWallet(anvilKey0)
	buyerCaller := contract.NewCaller(client, buyerWallet, 31337, cache)

	usdcAddr := deployMockUSDC(t, client, anvilKey0, chainID)
	implAddr := deployVaultImpl(t, client, anvilKey0, chainID)
	factoryAddr := deployFactory(t, client, anvilKey0, chainID, implAddr)

	factoryClient := NewFactoryClient(buyerCaller, factoryAddr, 31337)
	dl := big.NewInt(time.Now().Unix() + 86400)
	arbitratorAddr := crypto.PubkeyToAddress(anvilKey2.PublicKey)
	sellerAddr := crypto.PubkeyToAddress(anvilKey1.PublicKey)

	for i := 0; i < 3; i++ {
		_, _, err := factoryClient.CreateVault(ctx, sellerAddr, usdcAddr, big.NewInt(100_000_000), dl, arbitratorAddr)
		require.NoError(t, err)
	}

	count, err := factoryClient.VaultCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(3), count)

	addrs := make(map[common.Address]bool, 3)
	for i := 0; i < 3; i++ {
		addr, err := factoryClient.GetVault(ctx, big.NewInt(int64(i)))
		require.NoError(t, err)
		assert.NotEqual(t, common.Address{}, addr)
		addrs[addr] = true
	}
	assert.Len(t, addrs, 3)
}

func TestMonitorIntegration_EventDetection(t *testing.T) {
	ctx := context.Background()
	client, err := ethclient.Dial(anvilRPC)
	require.NoError(t, err)
	defer client.Close()

	chainID := big.NewInt(31337)
	cache := contract.NewABICache()
	bus := eventbus.New()

	buyerWallet := newTestWallet(anvilKey0)
	sellerWallet := newTestWallet(anvilKey1)
	arbitratorAddr := crypto.PubkeyToAddress(anvilKey2.PublicKey)

	buyerCaller := contract.NewCaller(client, buyerWallet, 31337, cache)

	usdcAddr := deployMockUSDC(t, client, anvilKey0, chainID)
	hubAddr := deployHub(t, client, anvilKey0, chainID, arbitratorAddr)

	mintUSDC(t, buyerCaller, usdcAddr, buyerWallet.addr, big.NewInt(10_000_000_000), 31337)
	approveUSDC(t, buyerCaller, usdcAddr, hubAddr, big.NewInt(10_000_000_000), 31337)

	monitor, err := NewEventMonitor(client, bus, nil, hubAddr,
		WithPollInterval(1*time.Second),
	)
	require.NoError(t, err)

	var startWg sync.WaitGroup
	startWg.Add(1)
	err = monitor.Start(ctx, &startWg)
	require.NoError(t, err)
	defer func() { _ = monitor.Stop(ctx) }()

	var depositReceived eventbus.EscrowOnChainDepositEvent
	var depositMu sync.Mutex
	bus.Subscribe(func(e eventbus.EscrowOnChainDepositEvent) {
		depositMu.Lock()
		depositReceived = e
		depositMu.Unlock()
	})

	hubClient := NewHubClient(buyerCaller, hubAddr, 31337)
	dl := big.NewInt(time.Now().Unix() + 86400)
	dealID, _, err := hubClient.CreateDeal(ctx, sellerWallet.addr, usdcAddr, big.NewInt(500_000_000), dl)
	require.NoError(t, err)

	_, err = hubClient.Deposit(ctx, dealID)
	require.NoError(t, err)

	time.Sleep(3 * time.Second)

	depositMu.Lock()
	assert.NotEmpty(t, depositReceived.TxHash)
	assert.Equal(t, big.NewInt(500_000_000), depositReceived.Amount)
	depositMu.Unlock()
}
