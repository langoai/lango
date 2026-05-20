package bindings

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/contract"
)

type fakeContractCaller struct {
	readResults  map[string]*contract.ContractCallResult
	readErrors   map[string]error
	writeResults map[string]*contract.ContractCallResult
	writeErrors  map[string]error
	reads        []contract.ContractCallRequest
	writes       []contract.ContractCallRequest
}

func newFakeContractCaller() *fakeContractCaller {
	return &fakeContractCaller{
		readResults:  make(map[string]*contract.ContractCallResult),
		readErrors:   make(map[string]error),
		writeResults: make(map[string]*contract.ContractCallResult),
		writeErrors:  make(map[string]error),
	}
}

func (f *fakeContractCaller) Read(
	_ context.Context,
	req contract.ContractCallRequest,
) (*contract.ContractCallResult, error) {
	f.reads = append(f.reads, req)
	if err := f.readErrors[req.Method]; err != nil {
		return nil, err
	}
	if result := f.readResults[req.Method]; result != nil {
		return result, nil
	}
	return &contract.ContractCallResult{}, nil
}

func (f *fakeContractCaller) Write(
	_ context.Context,
	req contract.ContractCallRequest,
) (*contract.ContractCallResult, error) {
	f.writes = append(f.writes, req)
	if err := f.writeErrors[req.Method]; err != nil {
		return nil, err
	}
	if result := f.writeResults[req.Method]; result != nil {
		return result, nil
	}
	return &contract.ContractCallResult{TxHash: "0xdefault"}, nil
}

func TestParseABI(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		parsed, err := ParseABI(Safe7579ABI)
		if err != nil {
			t.Fatalf("ParseABI returned error: %v", err)
		}
		if _, ok := parsed.Methods["installModule"]; !ok {
			t.Fatal("ParseABI result missing installModule")
		}
	})

	t.Run("error wraps parser error", func(t *testing.T) {
		_, err := ParseABI(`[{`)
		if err == nil {
			t.Fatal("ParseABI returned nil error")
		}
		if !strings.Contains(err.Error(), "parse ABI") {
			t.Fatalf("ParseABI error = %q, want parse ABI wrapper", err.Error())
		}
	})
}

func TestWriteWrappers(t *testing.T) {
	ctx := context.Background()
	address := common.HexToAddress("0x2000000000000000000000000000000000000002")
	module := common.HexToAddress("0x3000000000000000000000000000000000000003")
	sessionKey := common.HexToAddress("0x4000000000000000000000000000000000000004")
	chainID := int64(11155111)
	txHash := "0xtx"

	tests := []struct {
		name       string
		method     string
		abiJSON    string
		wantArgs   []interface{}
		call       func(*fakeContractCaller) (string, error)
		errWrapper string
	}{
		{
			name:     "Safe7579 InstallModule",
			method:   "installModule",
			abiJSON:  Safe7579ABI,
			wantArgs: []interface{}{big.NewInt(1), module, []byte{0xaa}},
			call: func(caller *fakeContractCaller) (string, error) {
				return NewSafe7579Client(caller, address, chainID).InstallModule(
					ctx, big.NewInt(1), module, []byte{0xaa},
				)
			},
			errWrapper: "install module",
		},
		{
			name:     "Safe7579 UninstallModule",
			method:   "uninstallModule",
			abiJSON:  Safe7579ABI,
			wantArgs: []interface{}{big.NewInt(2), module, []byte{0xbb}},
			call: func(caller *fakeContractCaller) (string, error) {
				return NewSafe7579Client(caller, address, chainID).UninstallModule(
					ctx, big.NewInt(2), module, []byte{0xbb},
				)
			},
			errWrapper: "uninstall module",
		},
		{
			name:     "Safe7579 Execute",
			method:   "execute",
			abiJSON:  Safe7579ABI,
			wantArgs: []interface{}{bytes32(0x01), []byte{0xcc}},
			call: func(caller *fakeContractCaller) (string, error) {
				return NewSafe7579Client(caller, address, chainID).Execute(
					ctx, bytes32(0x01), []byte{0xcc},
				)
			},
			errWrapper: "execute",
		},
		{
			name:     "SessionValidator RegisterSessionKey",
			method:   "registerSessionKey",
			abiJSON:  SessionValidatorABI,
			wantArgs: []interface{}{sessionKey, "policy"},
			call: func(caller *fakeContractCaller) (string, error) {
				return NewSessionValidatorClient(caller, address, chainID).RegisterSessionKey(
					ctx, sessionKey, "policy",
				)
			},
			errWrapper: "register session key",
		},
		{
			name:     "SessionValidator RevokeSessionKey",
			method:   "revokeSessionKey",
			abiJSON:  SessionValidatorABI,
			wantArgs: []interface{}{sessionKey},
			call: func(caller *fakeContractCaller) (string, error) {
				return NewSessionValidatorClient(caller, address, chainID).RevokeSessionKey(
					ctx, sessionKey,
				)
			},
			errWrapper: "revoke session key",
		},
		{
			name:     "SpendingHook SetLimits",
			method:   "setLimits",
			abiJSON:  SpendingHookABI,
			wantArgs: []interface{}{big.NewInt(10), big.NewInt(20), big.NewInt(30)},
			call: func(caller *fakeContractCaller) (string, error) {
				return NewSpendingHookClient(caller, address, chainID).SetLimits(
					ctx, big.NewInt(10), big.NewInt(20), big.NewInt(30),
				)
			},
			errWrapper: "set limits",
		},
		{
			name:    "EscrowExecutor ExecuteBatchedEscrow",
			method:  "executeBatchedEscrow",
			abiJSON: EscrowExecutorABI,
			wantArgs: []interface{}{[]interface{}{
				struct {
					Target   common.Address
					Value    *big.Int
					CallData []byte
				}{
					Target:   module,
					Value:    big.NewInt(0),
					CallData: []byte{0xdd},
				},
			}},
			call: func(caller *fakeContractCaller) (string, error) {
				return NewEscrowExecutorClient(caller, address, chainID).ExecuteBatchedEscrow(
					ctx,
					[]EscrowExecution{{Target: module, CallData: []byte{0xdd}}},
				)
			},
			errWrapper: "execute batched escrow",
		},
		{
			name:     "EscrowExecutor ReleaseEscrow",
			method:   "releaseEscrow",
			abiJSON:  EscrowExecutorABI,
			wantArgs: []interface{}{bytes32(0xee)},
			call: func(caller *fakeContractCaller) (string, error) {
				return NewEscrowExecutorClient(caller, address, chainID).ReleaseEscrow(
					ctx, bytes32(0xee),
				)
			},
			errWrapper: "release escrow",
		},
		{
			name:     "EscrowExecutor RefundEscrow",
			method:   "refundEscrow",
			abiJSON:  EscrowExecutorABI,
			wantArgs: []interface{}{bytes32(0xff)},
			call: func(caller *fakeContractCaller) (string, error) {
				return NewEscrowExecutorClient(caller, address, chainID).RefundEscrow(
					ctx, bytes32(0xff),
				)
			},
			errWrapper: "refund escrow",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			caller := newFakeContractCaller()
			caller.writeResults[tc.method] = &contract.ContractCallResult{TxHash: txHash}
			got, err := tc.call(caller)
			if err != nil {
				t.Fatalf("write wrapper returned error: %v", err)
			}
			if got != txHash {
				t.Fatalf("tx hash = %q, want %q", got, txHash)
			}
			assertSingleWrite(t, caller, contract.ContractCallRequest{
				ChainID: chainID,
				Address: address,
				ABI:     tc.abiJSON,
				Method:  tc.method,
				Args:    tc.wantArgs,
			})

			sentinel := errors.New("write failed")
			caller = newFakeContractCaller()
			caller.writeErrors[tc.method] = sentinel
			got, err = tc.call(caller)
			if err == nil {
				t.Fatal("write wrapper returned nil error")
			}
			if got != "" {
				t.Fatalf("tx hash on error = %q, want empty", got)
			}
			if !errors.Is(err, sentinel) || !strings.Contains(err.Error(), tc.errWrapper) {
				t.Fatalf("error = %v, want wrapper %q around sentinel", err, tc.errWrapper)
			}
		})
	}
}

func TestReadWrappersTypedData(t *testing.T) {
	ctx := context.Background()
	address := common.HexToAddress("0x5000000000000000000000000000000000000005")
	module := common.HexToAddress("0x6000000000000000000000000000000000000006")
	account := common.HexToAddress("0x7000000000000000000000000000000000000007")
	sessionKey := common.HexToAddress("0x8000000000000000000000000000000000000008")
	chainID := int64(1)

	t.Run("Safe7579 IsModuleInstalled", func(t *testing.T) {
		caller := newFakeContractCaller()
		caller.readResults["isModuleInstalled"] = &contract.ContractCallResult{Data: []interface{}{true}}
		got, err := NewSafe7579Client(caller, address, chainID).IsModuleInstalled(
			ctx, big.NewInt(1), module, []byte{0x01},
		)
		if err != nil {
			t.Fatalf("IsModuleInstalled returned error: %v", err)
		}
		if !got {
			t.Fatal("IsModuleInstalled = false, want true")
		}
		assertSingleRead(t, caller, contract.ContractCallRequest{
			ChainID: chainID,
			Address: address,
			ABI:     Safe7579ABI,
			Method:  "isModuleInstalled",
			Args:    []interface{}{big.NewInt(1), module, []byte{0x01}},
		})
	})

	t.Run("Safe7579 AccountID", func(t *testing.T) {
		caller := newFakeContractCaller()
		caller.readResults["accountId"] = &contract.ContractCallResult{Data: []interface{}{"lango.safe.7579"}}
		got, err := NewSafe7579Client(caller, address, chainID).AccountID(ctx)
		if err != nil {
			t.Fatalf("AccountID returned error: %v", err)
		}
		if got != "lango.safe.7579" {
			t.Fatalf("AccountID = %q, want lango.safe.7579", got)
		}
		assertSingleRead(t, caller, contract.ContractCallRequest{
			ChainID: chainID,
			Address: address,
			ABI:     Safe7579ABI,
			Method:  "accountId",
			Args:    []interface{}{},
		})
	})

	t.Run("Safe7579 SupportsModule", func(t *testing.T) {
		caller := newFakeContractCaller()
		caller.readResults["supportsModule"] = &contract.ContractCallResult{Data: []interface{}{true}}
		got, err := NewSafe7579Client(caller, address, chainID).SupportsModule(ctx, big.NewInt(2))
		if err != nil {
			t.Fatalf("SupportsModule returned error: %v", err)
		}
		if !got {
			t.Fatal("SupportsModule = false, want true")
		}
		assertSingleRead(t, caller, contract.ContractCallRequest{
			ChainID: chainID,
			Address: address,
			ABI:     Safe7579ABI,
			Method:  "supportsModule",
			Args:    []interface{}{big.NewInt(2)},
		})
	})

	t.Run("SessionValidator GetSessionKeyPolicy", func(t *testing.T) {
		caller := newFakeContractCaller()
		policy := struct{ Active bool }{Active: true}
		caller.readResults["getSessionKeyPolicy"] = &contract.ContractCallResult{Data: []interface{}{policy}}
		got, err := NewSessionValidatorClient(caller, address, chainID).GetSessionKeyPolicy(ctx, sessionKey)
		if err != nil {
			t.Fatalf("GetSessionKeyPolicy returned error: %v", err)
		}
		if !reflect.DeepEqual(got, policy) {
			t.Fatalf("policy = %#v, want %#v", got, policy)
		}
		assertSingleRead(t, caller, contract.ContractCallRequest{
			ChainID: chainID,
			Address: address,
			ABI:     SessionValidatorABI,
			Method:  "getSessionKeyPolicy",
			Args:    []interface{}{sessionKey},
		})
	})

	t.Run("SessionValidator IsSessionKeyActive", func(t *testing.T) {
		caller := newFakeContractCaller()
		caller.readResults["isSessionKeyActive"] = &contract.ContractCallResult{Data: []interface{}{true}}
		got, err := NewSessionValidatorClient(caller, address, chainID).IsSessionKeyActive(ctx, sessionKey)
		if err != nil {
			t.Fatalf("IsSessionKeyActive returned error: %v", err)
		}
		if !got {
			t.Fatal("IsSessionKeyActive = false, want true")
		}
		assertSingleRead(t, caller, contract.ContractCallRequest{
			ChainID: chainID,
			Address: address,
			ABI:     SessionValidatorABI,
			Method:  "isSessionKeyActive",
			Args:    []interface{}{sessionKey},
		})
	})

	t.Run("SpendingHook GetConfig", func(t *testing.T) {
		caller := newFakeContractCaller()
		caller.readResults["getConfig"] = &contract.ContractCallResult{
			Data: []interface{}{big.NewInt(10), big.NewInt(20), big.NewInt(30)},
		}
		got, err := NewSpendingHookClient(caller, address, chainID).GetConfig(ctx, account)
		if err != nil {
			t.Fatalf("GetConfig returned error: %v", err)
		}
		assertBigInt(t, "PerTxLimit", got.PerTxLimit, 10)
		assertBigInt(t, "DailyLimit", got.DailyLimit, 20)
		assertBigInt(t, "CumulativeLimit", got.CumulativeLimit, 30)
		assertSingleRead(t, caller, contract.ContractCallRequest{
			ChainID: chainID,
			Address: address,
			ABI:     SpendingHookABI,
			Method:  "getConfig",
			Args:    []interface{}{account},
		})
	})

	t.Run("SpendingHook GetSpendState", func(t *testing.T) {
		caller := newFakeContractCaller()
		caller.readResults["getSpendState"] = &contract.ContractCallResult{
			Data: []interface{}{big.NewInt(40), big.NewInt(50), big.NewInt(60)},
		}
		got, err := NewSpendingHookClient(caller, address, chainID).GetSpendState(ctx, account, sessionKey)
		if err != nil {
			t.Fatalf("GetSpendState returned error: %v", err)
		}
		assertBigInt(t, "DailySpent", got.DailySpent, 40)
		assertBigInt(t, "CumulativeSpent", got.CumulativeSpent, 50)
		assertBigInt(t, "LastResetDay", got.LastResetDay, 60)
		assertSingleRead(t, caller, contract.ContractCallRequest{
			ChainID: chainID,
			Address: address,
			ABI:     SpendingHookABI,
			Method:  "getSpendState",
			Args:    []interface{}{account, sessionKey},
		})
	})

	t.Run("EscrowExecutor GetEscrowStatus", func(t *testing.T) {
		caller := newFakeContractCaller()
		caller.readResults["getEscrowStatus"] = &contract.ContractCallResult{
			Data: []interface{}{uint8(2), big.NewInt(70)},
		}
		gotStatus, gotAmount, err := NewEscrowExecutorClient(caller, address, chainID).GetEscrowStatus(
			ctx, bytes32(0xab),
		)
		if err != nil {
			t.Fatalf("GetEscrowStatus returned error: %v", err)
		}
		if gotStatus != 2 {
			t.Fatalf("status = %d, want 2", gotStatus)
		}
		assertBigInt(t, "amount", gotAmount, 70)
		assertSingleRead(t, caller, contract.ContractCallRequest{
			ChainID: chainID,
			Address: address,
			ABI:     EscrowExecutorABI,
			Method:  "getEscrowStatus",
			Args:    []interface{}{bytes32(0xab)},
		})
	})
}

func TestReadWrappersFallbackData(t *testing.T) {
	ctx := context.Background()
	address := common.HexToAddress("0x9000000000000000000000000000000000000009")
	chainID := int64(10)

	t.Run("Safe7579 fallbacks", func(t *testing.T) {
		caller := newFakeContractCaller()
		client := NewSafe7579Client(caller, address, chainID)

		installed, err := client.IsModuleInstalled(ctx, big.NewInt(1), address, nil)
		if err != nil {
			t.Fatalf("IsModuleInstalled returned error: %v", err)
		}
		if installed {
			t.Fatal("IsModuleInstalled fallback = true, want false")
		}

		accountID, err := client.AccountID(ctx)
		if err != nil {
			t.Fatalf("AccountID returned error: %v", err)
		}
		if accountID != "" {
			t.Fatalf("AccountID fallback = %q, want empty", accountID)
		}

		supported, err := client.SupportsModule(ctx, big.NewInt(2))
		if err != nil {
			t.Fatalf("SupportsModule returned error: %v", err)
		}
		if supported {
			t.Fatal("SupportsModule fallback = true, want false")
		}
	})

	t.Run("SessionValidator fallbacks", func(t *testing.T) {
		caller := newFakeContractCaller()
		client := NewSessionValidatorClient(caller, address, chainID)

		policy, err := client.GetSessionKeyPolicy(ctx, address)
		if err != nil {
			t.Fatalf("GetSessionKeyPolicy returned error: %v", err)
		}
		if policy != nil {
			t.Fatalf("policy fallback = %#v, want nil", policy)
		}

		active, err := client.IsSessionKeyActive(ctx, address)
		if err != nil {
			t.Fatalf("IsSessionKeyActive returned error: %v", err)
		}
		if active {
			t.Fatal("IsSessionKeyActive fallback = true, want false")
		}
	})

	t.Run("SpendingHook fallbacks", func(t *testing.T) {
		caller := newFakeContractCaller()
		client := NewSpendingHookClient(caller, address, chainID)

		config, err := client.GetConfig(ctx, address)
		if err != nil {
			t.Fatalf("GetConfig returned error: %v", err)
		}
		assertBigInt(t, "PerTxLimit", config.PerTxLimit, 0)
		assertBigInt(t, "DailyLimit", config.DailyLimit, 0)
		assertBigInt(t, "CumulativeLimit", config.CumulativeLimit, 0)

		state, err := client.GetSpendState(ctx, address, address)
		if err != nil {
			t.Fatalf("GetSpendState returned error: %v", err)
		}
		assertBigInt(t, "DailySpent", state.DailySpent, 0)
		assertBigInt(t, "CumulativeSpent", state.CumulativeSpent, 0)
		assertBigInt(t, "LastResetDay", state.LastResetDay, 0)
	})

	t.Run("EscrowExecutor fallback", func(t *testing.T) {
		caller := newFakeContractCaller()
		status, amount, err := NewEscrowExecutorClient(caller, address, chainID).GetEscrowStatus(
			ctx, bytes32(0x11),
		)
		if err != nil {
			t.Fatalf("GetEscrowStatus returned error: %v", err)
		}
		if status != 0 {
			t.Fatalf("status fallback = %d, want 0", status)
		}
		assertBigInt(t, "amount", amount, 0)
	})
}

func TestReadWrappersWrapErrors(t *testing.T) {
	ctx := context.Background()
	address := common.HexToAddress("0xa00000000000000000000000000000000000000a")
	chainID := int64(8453)
	sentinel := errors.New("read failed")

	tests := []struct {
		name       string
		method     string
		call       func(*fakeContractCaller) error
		errWrapper string
	}{
		{
			name:   "Safe7579 IsModuleInstalled",
			method: "isModuleInstalled",
			call: func(caller *fakeContractCaller) error {
				_, err := NewSafe7579Client(caller, address, chainID).IsModuleInstalled(
					ctx, big.NewInt(1), address, nil,
				)
				return err
			},
			errWrapper: "check module installed",
		},
		{
			name:   "Safe7579 AccountID",
			method: "accountId",
			call: func(caller *fakeContractCaller) error {
				_, err := NewSafe7579Client(caller, address, chainID).AccountID(ctx)
				return err
			},
			errWrapper: "get account id",
		},
		{
			name:   "Safe7579 SupportsModule",
			method: "supportsModule",
			call: func(caller *fakeContractCaller) error {
				_, err := NewSafe7579Client(caller, address, chainID).SupportsModule(ctx, big.NewInt(1))
				return err
			},
			errWrapper: "check module support",
		},
		{
			name:   "SessionValidator GetSessionKeyPolicy",
			method: "getSessionKeyPolicy",
			call: func(caller *fakeContractCaller) error {
				_, err := NewSessionValidatorClient(caller, address, chainID).GetSessionKeyPolicy(ctx, address)
				return err
			},
			errWrapper: "get session key policy",
		},
		{
			name:   "SessionValidator IsSessionKeyActive",
			method: "isSessionKeyActive",
			call: func(caller *fakeContractCaller) error {
				_, err := NewSessionValidatorClient(caller, address, chainID).IsSessionKeyActive(ctx, address)
				return err
			},
			errWrapper: "check session key",
		},
		{
			name:   "SpendingHook GetConfig",
			method: "getConfig",
			call: func(caller *fakeContractCaller) error {
				_, err := NewSpendingHookClient(caller, address, chainID).GetConfig(ctx, address)
				return err
			},
			errWrapper: "get config",
		},
		{
			name:   "SpendingHook GetSpendState",
			method: "getSpendState",
			call: func(caller *fakeContractCaller) error {
				_, err := NewSpendingHookClient(caller, address, chainID).GetSpendState(ctx, address, address)
				return err
			},
			errWrapper: "get spend state",
		},
		{
			name:   "EscrowExecutor GetEscrowStatus",
			method: "getEscrowStatus",
			call: func(caller *fakeContractCaller) error {
				_, _, err := NewEscrowExecutorClient(caller, address, chainID).GetEscrowStatus(
					ctx, bytes32(0x22),
				)
				return err
			},
			errWrapper: "get escrow status",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			caller := newFakeContractCaller()
			caller.readErrors[tc.method] = sentinel

			err := tc.call(caller)
			if err == nil {
				t.Fatal("read wrapper returned nil error")
			}
			if !errors.Is(err, sentinel) || !strings.Contains(err.Error(), tc.errWrapper) {
				t.Fatalf("error = %v, want wrapper %q around sentinel", err, tc.errWrapper)
			}
		})
	}
}

func assertSingleRead(
	t *testing.T,
	caller *fakeContractCaller,
	want contract.ContractCallRequest,
) {
	t.Helper()
	if len(caller.reads) != 1 {
		t.Fatalf("read count = %d, want 1", len(caller.reads))
	}
	assertRequest(t, "read", caller.reads[0], want)
}

func assertSingleWrite(
	t *testing.T,
	caller *fakeContractCaller,
	want contract.ContractCallRequest,
) {
	t.Helper()
	if len(caller.writes) != 1 {
		t.Fatalf("write count = %d, want 1", len(caller.writes))
	}
	assertRequest(t, "write", caller.writes[0], want)
}

func assertRequest(
	t *testing.T,
	kind string,
	got contract.ContractCallRequest,
	want contract.ContractCallRequest,
) {
	t.Helper()
	if got.ChainID != want.ChainID {
		t.Fatalf("%s ChainID = %d, want %d", kind, got.ChainID, want.ChainID)
	}
	if got.Address != want.Address {
		t.Fatalf("%s Address = %s, want %s", kind, got.Address, want.Address)
	}
	if got.ABI != want.ABI {
		t.Fatalf("%s ABI mismatch for method %s", kind, want.Method)
	}
	if got.Method != want.Method {
		t.Fatalf("%s Method = %q, want %q", kind, got.Method, want.Method)
	}
	if !reflect.DeepEqual(got.Args, want.Args) {
		t.Fatalf("%s Args = %s, want %s", kind, formatArgs(got.Args), formatArgs(want.Args))
	}
}

func assertBigInt(t *testing.T, name string, got *big.Int, want int64) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s = nil, want %d", name, want)
	}
	if got.Cmp(big.NewInt(want)) != 0 {
		t.Fatalf("%s = %s, want %d", name, got.String(), want)
	}
}

func bytes32(last byte) [32]byte {
	var out [32]byte
	out[31] = last
	return out
}

func formatArgs(args []interface{}) string {
	return fmt.Sprintf("%#v", args)
}
