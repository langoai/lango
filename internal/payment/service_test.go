package payment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/paymenttx"
	"github.com/langoai/lango/internal/sqlitedriver"
	"github.com/langoai/lango/internal/testutil/schemautil"
)

// --- mock implementations ---

type mockWallet struct {
	address    string
	addressErr error
	signBytes  []byte
	signErr    error
}

func (m *mockWallet) Address(_ context.Context) (string, error) {
	return m.address, m.addressErr
}

func (m *mockWallet) Balance(_ context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (m *mockWallet) SignTransaction(_ context.Context, _ []byte) ([]byte, error) {
	return m.signBytes, m.signErr
}

func (m *mockWallet) SignMessage(_ context.Context, _ []byte) ([]byte, error) {
	return nil, nil
}

func (m *mockWallet) PublicKey(_ context.Context) ([]byte, error) {
	return nil, nil
}

type mockLimiter struct {
	checkErr  error
	recordErr error
}

func (m *mockLimiter) Check(_ context.Context, _ *big.Int) error {
	return m.checkErr
}

func (m *mockLimiter) Record(_ context.Context, _ *big.Int) error {
	return m.recordErr
}

func (m *mockLimiter) DailySpent(_ context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (m *mockLimiter) DailyRemaining(_ context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (m *mockLimiter) IsAutoApprovable(_ context.Context, _ *big.Int) (bool, error) {
	return false, nil
}

// validAddr is a well-formed Ethereum address for tests.
const validAddr = "0x1234567890abcdef1234567890abcdef12345678"

var paymentTestDBSeq uint64
var paymentTestDBMu sync.Mutex

func testEntClient(t testing.TB) *ent.Client {
	t.Helper()
	paymentTestDBMu.Lock()
	defer paymentTestDBMu.Unlock()

	name := strings.NewReplacer("/", "-", " ", "-", ":", "-").Replace(t.Name())
	seq := atomic.AddUint64(&paymentTestDBSeq, 1)
	db, err := sql.Open(sqlitedriver.DriverName(), sqlitedriver.MemoryDSN(fmt.Sprintf("%s-%d", name, seq)))
	require.NoError(t, err)
	require.NoError(t, sqlitedriver.ConfigureConnection(db, false))
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))
	require.NoError(t, schemautil.CreateSchema(context.Background(), client))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// --- Send tests ---

func TestService_Send_InvalidAddress(t *testing.T) {
	tests := []struct {
		give         string
		wantMsg      string
		wantCauseMsg string
	}{
		{give: "not-an-address", wantMsg: "invalid recipient", wantCauseMsg: `invalid address format: "not-an-address"`},
		{give: "0x123", wantMsg: "invalid recipient", wantCauseMsg: `invalid address format: "0x123"`},
		{give: "", wantMsg: "invalid recipient", wantCauseMsg: `invalid address format: ""`},
		{give: "0xzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", wantMsg: "invalid recipient", wantCauseMsg: `invalid hex address: "0xzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"`},
	}

	svc := &Service{
		wallet:  &mockWallet{address: validAddr},
		limiter: &mockLimiter{},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			receipt, err := svc.Send(context.Background(), PaymentRequest{
				To:     tt.give,
				Amount: "1.00",
			})
			require.Error(t, err)
			assert.Nil(t, receipt)
			assert.Contains(t, err.Error(), tt.wantMsg)
			assert.Contains(t, err.Error(), tt.wantCauseMsg)
		})
	}
}

func TestService_Send_InvalidAmount(t *testing.T) {
	tests := []struct {
		give         string
		wantMsg      string
		wantCauseMsg string
	}{
		{give: "not-a-number", wantMsg: "invalid amount", wantCauseMsg: `invalid USDC amount: "not-a-number"`},
		{give: "", wantMsg: "invalid amount", wantCauseMsg: `invalid USDC amount: ""`},
		{give: "abc", wantMsg: "invalid amount", wantCauseMsg: `invalid USDC amount: "abc"`},
		{give: "0.0000001", wantMsg: "invalid amount", wantCauseMsg: `USDC amount "0.0000001" has too many decimal places`},
	}

	svc := &Service{
		wallet:  &mockWallet{address: validAddr},
		limiter: &mockLimiter{},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			receipt, err := svc.Send(context.Background(), PaymentRequest{
				To:     validAddr,
				Amount: tt.give,
			})
			require.Error(t, err)
			assert.Nil(t, receipt)
			assert.Contains(t, err.Error(), tt.wantMsg)
			assert.Contains(t, err.Error(), tt.wantCauseMsg)
		})
	}
}

func TestService_Send_ZeroOrNegativeAmount(t *testing.T) {
	tests := []struct {
		give            string
		wantMsg         string
		notWantCauseMsg string
	}{
		{give: "0", wantMsg: "amount must be positive", notWantCauseMsg: "invalid amount"},
		{give: "0.00", wantMsg: "amount must be positive", notWantCauseMsg: "invalid amount"},
		{give: "-1.00", wantMsg: "amount must be positive", notWantCauseMsg: "invalid amount"},
	}

	svc := &Service{
		wallet:  &mockWallet{address: validAddr},
		limiter: &mockLimiter{},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			receipt, err := svc.Send(context.Background(), PaymentRequest{
				To:     validAddr,
				Amount: tt.give,
			})
			require.Error(t, err)
			assert.Nil(t, receipt)
			assert.Contains(t, err.Error(), tt.wantMsg)
			assert.NotContains(t, err.Error(), tt.notWantCauseMsg)
		})
	}
}

func TestService_Send_LimiterCheckFails(t *testing.T) {
	limiterErr := errors.New("daily limit exceeded")

	svc := &Service{
		wallet:  &mockWallet{address: validAddr},
		limiter: &mockLimiter{checkErr: limiterErr},
	}

	receipt, err := svc.Send(context.Background(), PaymentRequest{
		To:     validAddr,
		Amount: "1.00",
	})

	require.Error(t, err)
	assert.Nil(t, receipt)
	assert.Contains(t, err.Error(), "spending limit")
	assert.Contains(t, err.Error(), "daily limit exceeded")
	assert.NotContains(t, err.Error(), "invalid amount")
	assert.ErrorIs(t, err, limiterErr)
}

func TestService_Send_WalletAddressFails(t *testing.T) {
	walletErr := errors.New("wallet locked")

	svc := &Service{
		wallet:  &mockWallet{addressErr: walletErr},
		limiter: &mockLimiter{},
	}

	receipt, err := svc.Send(context.Background(), PaymentRequest{
		To:     validAddr,
		Amount: "1.00",
	})

	require.Error(t, err)
	assert.Nil(t, receipt)
	assert.Contains(t, err.Error(), "get wallet address")
	assert.Contains(t, err.Error(), "wallet locked")
	assert.NotContains(t, err.Error(), "invalid amount")
	assert.NotContains(t, err.Error(), "spending limit")
	assert.ErrorIs(t, err, walletErr)
}

func TestService_Send_MissingLimiterFailsClosed(t *testing.T) {
	svc := &Service{
		wallet: &mockWallet{address: validAddr},
	}

	receipt, err := svc.Send(context.Background(), PaymentRequest{
		To:     validAddr,
		Amount: "1.00",
	})

	require.Error(t, err)
	assert.Nil(t, receipt)
	assert.Contains(t, err.Error(), "spending limit")
	assert.Contains(t, err.Error(), "spending limiter unavailable")
}

func TestService_Send_MissingWalletFailsClosed(t *testing.T) {
	svc := &Service{
		limiter: &mockLimiter{},
	}

	receipt, err := svc.Send(context.Background(), PaymentRequest{
		To:     validAddr,
		Amount: "1.00",
	})

	require.Error(t, err)
	assert.Nil(t, receipt)
	assert.Contains(t, err.Error(), "get wallet address")
	assert.Contains(t, err.Error(), "wallet provider unavailable")
}

func TestService_Send_MissingStoreFailsClosed(t *testing.T) {
	svc := &Service{
		wallet:  &mockWallet{address: validAddr},
		limiter: &mockLimiter{},
	}

	receipt, err := svc.Send(context.Background(), PaymentRequest{
		To:     validAddr,
		Amount: "1.00",
	})

	require.Error(t, err)
	assert.Nil(t, receipt)
	assert.Contains(t, err.Error(), "create tx record")
	assert.Contains(t, err.Error(), "payment store unavailable")
}

// --- History tests ---

func TestService_History(t *testing.T) {
	client := testEntClient(t)
	ctx := context.Background()

	// Seed two payment records.
	client.PaymentTx.Create().
		SetFromAddress(validAddr).
		SetToAddress("0xabcdef1234567890abcdef1234567890abcdef12").
		SetAmount("1.00").
		SetChainID(84532).
		SetStatus(paymenttx.StatusSubmitted).
		SetTxHash("0xaaa").
		SetPurpose("test payment 1").
		SaveX(ctx)

	client.PaymentTx.Create().
		SetFromAddress(validAddr).
		SetToAddress("0xabcdef1234567890abcdef1234567890abcdef12").
		SetAmount("2.50").
		SetChainID(84532).
		SetStatus(paymenttx.StatusFailed).
		SetErrorMessage("gas estimation failed").
		SaveX(ctx)

	svc := &Service{store: NewEntTxStore(client)}

	t.Run("returns all records", func(t *testing.T) {
		result, err := svc.History(ctx, 10)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("returns records in descending created_at order", func(t *testing.T) {
		client.PaymentTx.Create().
			SetFromAddress(validAddr).
			SetToAddress("0xabcdef1234567890abcdef1234567890abcdef12").
			SetAmount("9.99").
			SetChainID(84532).
			SetStatus(paymenttx.StatusConfirmed).
			SetTxHash("0xnewest").
			SetCreatedAt(time.Now().Add(2 * time.Hour)).
			SaveX(ctx)

		result, err := svc.History(ctx, 10)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(result), 3)
		assert.Equal(t, "0xnewest", result[0].TxHash)
		assert.True(t, !result[0].CreatedAt.Before(result[1].CreatedAt))
		assert.True(t, !result[1].CreatedAt.Before(result[2].CreatedAt))
	})

	t.Run("respects limit", func(t *testing.T) {
		result, err := svc.History(ctx, 1)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "0xnewest", result[0].TxHash)
	})

	t.Run("default limit when zero", func(t *testing.T) {
		result, err := svc.History(ctx, 0)
		require.NoError(t, err)
		assert.Len(t, result, 3) // all records, DefaultHistoryLimit > 3
	})

	t.Run("default limit when negative", func(t *testing.T) {
		result, err := svc.History(ctx, -1)
		require.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("fields populated correctly", func(t *testing.T) {
		result, err := svc.History(ctx, 10)
		require.NoError(t, err)
		require.NotEmpty(t, result)

		// Find the failed transaction.
		var failedTx TransactionInfo
		for _, tx := range result {
			if tx.Status == string(paymenttx.StatusFailed) {
				failedTx = tx
				break
			}
		}
		assert.Equal(t, "2.50", failedTx.Amount)
		assert.Equal(t, validAddr, failedTx.From)
		assert.Equal(t, "0xabcdef1234567890abcdef1234567890abcdef12", failedTx.To)
		assert.Equal(t, int64(84532), failedTx.ChainID)
		assert.Equal(t, "gas estimation failed", failedTx.ErrorMessage)
	})
}

// --- RecordX402Payment tests ---

func TestService_RecordX402Payment(t *testing.T) {
	client := testEntClient(t)
	ctx := context.Background()

	svc := &Service{store: NewEntTxStore(client)}

	t.Run("happy path", func(t *testing.T) {
		err := svc.RecordX402Payment(ctx, X402PaymentRecord{
			URL:     "https://api.example.com/resource",
			Amount:  "0.05",
			From:    validAddr,
			To:      "0xabcdef1234567890abcdef1234567890abcdef12",
			ChainID: 84532,
		})
		require.NoError(t, err)

		// Verify the record was created.
		txs, err := client.PaymentTx.Query().All(ctx)
		require.NoError(t, err)
		require.Len(t, txs, 1)

		tx := txs[0]
		assert.Equal(t, validAddr, tx.FromAddress)
		assert.Equal(t, "0xabcdef1234567890abcdef1234567890abcdef12", tx.ToAddress)
		assert.Equal(t, "0.05", tx.Amount)
		assert.Equal(t, int64(84532), tx.ChainID)
		assert.Equal(t, paymenttx.StatusSubmitted, tx.Status)
		assert.Equal(t, purposeX402AutoPayment, tx.Purpose)
		assert.Equal(t, "https://api.example.com/resource", tx.X402URL)
		assert.Equal(t, paymenttx.PaymentMethodX402V2, tx.PaymentMethod)
	})
}

func TestService_RecordX402Payment_MissingStoreFailsClosed(t *testing.T) {
	svc := &Service{}

	err := svc.RecordX402Payment(context.Background(), X402PaymentRecord{
		URL:     "https://api.example.com/resource",
		Amount:  "0.05",
		From:    validAddr,
		To:      "0xabcdef1234567890abcdef1234567890abcdef12",
		ChainID: 84532,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "record X402 payment")
	assert.Contains(t, err.Error(), "payment store unavailable")
}

// --- failTx tests ---

func TestService_failTx(t *testing.T) {
	client := testEntClient(t)
	ctx := context.Background()

	svc := &Service{store: NewEntTxStore(client)}

	// Create a pending transaction.
	ptx := client.PaymentTx.Create().
		SetFromAddress(validAddr).
		SetToAddress("0xabcdef1234567890abcdef1234567890abcdef12").
		SetAmount("1.00").
		SetChainID(84532).
		SetStatus(paymenttx.StatusPending).
		SaveX(ctx)

	require.Equal(t, paymenttx.StatusPending, ptx.Status)

	// Mark it as failed.
	txErr := errors.New("gas estimation failed")
	svc.failTx(ctx, ptx.ID, txErr)

	// Verify the record was updated.
	updated := client.PaymentTx.GetX(ctx, ptx.ID)
	assert.Equal(t, paymenttx.StatusFailed, updated.Status)
	assert.Equal(t, "gas estimation failed", updated.ErrorMessage)
}

// --- WalletAddress tests ---

func TestService_WalletAddress(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		svc := &Service{
			wallet: &mockWallet{address: validAddr},
		}

		addr, err := svc.WalletAddress(context.Background())
		require.NoError(t, err)
		assert.Equal(t, validAddr, addr)
	})

	t.Run("wallet error", func(t *testing.T) {
		walletErr := errors.New("wallet locked")
		svc := &Service{
			wallet: &mockWallet{addressErr: walletErr},
		}

		addr, err := svc.WalletAddress(context.Background())
		require.Error(t, err)
		assert.Empty(t, addr)
		assert.ErrorIs(t, err, walletErr)
	})

	t.Run("missing wallet provider", func(t *testing.T) {
		svc := &Service{}

		addr, err := svc.WalletAddress(context.Background())
		require.Error(t, err)
		assert.Empty(t, addr)
		assert.Contains(t, err.Error(), "wallet provider unavailable")
	})
}

// --- Balance tests ---

func TestService_Balance_FailsClosedWhenUnwired(t *testing.T) {
	t.Run("missing wallet provider", func(t *testing.T) {
		svc := &Service{}

		balance, err := svc.Balance(context.Background())
		require.Error(t, err)
		assert.Empty(t, balance)
		assert.Contains(t, err.Error(), "wallet provider unavailable")
	})

	t.Run("missing builder", func(t *testing.T) {
		svc := &Service{
			wallet: &mockWallet{address: validAddr},
		}

		balance, err := svc.Balance(context.Background())
		require.Error(t, err)
		assert.Empty(t, balance)
		assert.Contains(t, err.Error(), "balance builder unavailable")
	})

	t.Run("missing rpc client", func(t *testing.T) {
		svc := &Service{
			wallet:  &mockWallet{address: validAddr},
			builder: &TxBuilder{},
		}

		balance, err := svc.Balance(context.Background())
		require.Error(t, err)
		assert.Empty(t, balance)
		assert.Contains(t, err.Error(), "balance RPC unavailable")
	})
}

func TestService_History_MissingStoreFailsClosed(t *testing.T) {
	svc := &Service{}

	result, err := svc.History(context.Background(), 10)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "query history")
	assert.Contains(t, err.Error(), "payment store unavailable")
}

func TestService_submitWithRetry_MissingRPCFailsClosed(t *testing.T) {
	svc := &Service{}

	tx := types.NewTx(&types.LegacyTx{
		Nonce: 0,
		To:    ptrAddress(common.HexToAddress(validAddr)),
		Value: big.NewInt(0),
		Gas:   21000,
	})

	txHash, err := svc.submitWithRetry(context.Background(), tx)
	require.Error(t, err)
	assert.Empty(t, txHash)
	assert.Contains(t, err.Error(), "transaction RPC unavailable")
}

func TestService_waitForConfirmation_MissingRPCFailsClosed(t *testing.T) {
	svc := &Service{}

	receipt, err := svc.waitForConfirmation(context.Background(), common.HexToHash("0x1"))
	require.Error(t, err)
	assert.Nil(t, receipt)
	assert.Contains(t, err.Error(), "receipt RPC unavailable")
}

// --- ChainID tests ---

func TestService_ChainID(t *testing.T) {
	tests := []struct {
		give int64
		want int64
	}{
		{give: 1, want: 1},
		{give: 84532, want: 84532},
		{give: 8453, want: 8453},
	}

	for _, tt := range tests {
		svc := &Service{chainID: tt.give}
		assert.Equal(t, tt.want, svc.ChainID())
	}
}

// --- nilIfEmpty tests ---

func TestNilIfEmpty(t *testing.T) {
	tests := []struct {
		give string
		want *string
	}{
		{give: "", want: nil},
		{give: "hello", want: strPtr("hello")},
		{give: " ", want: strPtr(" ")},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := nilIfEmpty(tt.give)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, *tt.want, *got)
			}
		})
	}
}

// --- Send with ent integration: record creation before builder ---

func TestService_Send_CreatesRecordBeforeBuild(t *testing.T) {
	// This test verifies that Send creates an ent record after passing
	// validation and limit checks, then fails closed when the builder is unavailable.
	client := testEntClient(t)
	ctx := context.Background()

	svc := &Service{
		wallet:  &mockWallet{address: validAddr},
		limiter: &mockLimiter{},
		store:   NewEntTxStore(client),
		chainID: 84532,
		// builder intentionally nil: Send should fail closed instead of panicking.
	}

	receipt, err := svc.Send(ctx, PaymentRequest{
		To:      "0xabcdef1234567890abcdef1234567890abcdef12",
		Amount:  "1.00",
		Purpose: "test purpose",
	})
	require.Error(t, err)
	assert.Nil(t, receipt)
	assert.Contains(t, err.Error(), "build transaction")
	assert.Contains(t, err.Error(), "transaction builder unavailable")

	// Verify a failed record was created with the actionable failure cause.
	txs, err := client.PaymentTx.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, txs, 1)

	tx := txs[0]
	assert.Equal(t, validAddr, tx.FromAddress)
	assert.Equal(t, "0xabcdef1234567890abcdef1234567890abcdef12", tx.ToAddress)
	assert.Equal(t, "1.00", tx.Amount)
	assert.Equal(t, int64(84532), tx.ChainID)
	assert.Equal(t, paymenttx.StatusFailed, tx.Status)
	assert.Equal(t, "test purpose", tx.Purpose)
	assert.Equal(t, "transaction builder unavailable", tx.ErrorMessage)
}

func TestService_Send_SetsOptionalFields(t *testing.T) {
	client := testEntClient(t)
	ctx := context.Background()

	svc := &Service{
		wallet:  &mockWallet{address: validAddr},
		limiter: &mockLimiter{},
		store:   NewEntTxStore(client),
		chainID: 84532,
	}

	receipt, err := svc.Send(ctx, PaymentRequest{
		To:         "0xabcdef1234567890abcdef1234567890abcdef12",
		Amount:     "5.00",
		Purpose:    "buy coffee",
		SessionKey: "session-abc",
		X402URL:    "https://api.example.com/paid",
	})
	require.Error(t, err)
	assert.Nil(t, receipt)
	assert.Contains(t, err.Error(), "build transaction")
	assert.Contains(t, err.Error(), "transaction builder unavailable")

	txs, err := client.PaymentTx.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, txs, 1)

	tx := txs[0]
	assert.Equal(t, "buy coffee", tx.Purpose)
	assert.Equal(t, "session-abc", tx.SessionKey)
	assert.Equal(t, "https://api.example.com/paid", tx.X402URL)
	assert.Equal(t, paymenttx.StatusFailed, tx.Status)
	assert.Equal(t, "transaction builder unavailable", tx.ErrorMessage)
}

func TestService_Send_OmitsEmptyOptionalFields(t *testing.T) {
	client := testEntClient(t)
	ctx := context.Background()

	svc := &Service{
		wallet:  &mockWallet{address: validAddr},
		limiter: &mockLimiter{},
		store:   NewEntTxStore(client),
		chainID: 84532,
	}

	receipt, err := svc.Send(ctx, PaymentRequest{
		To:     "0xabcdef1234567890abcdef1234567890abcdef12",
		Amount: "1.00",
		// Purpose, SessionKey, X402URL all empty
	})
	require.Error(t, err)
	assert.Nil(t, receipt)
	assert.Contains(t, err.Error(), "build transaction")
	assert.Contains(t, err.Error(), "transaction builder unavailable")

	txs, err := client.PaymentTx.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, txs, 1)

	tx := txs[0]
	assert.Empty(t, tx.Purpose)
	assert.Empty(t, tx.SessionKey)
	assert.Empty(t, tx.X402URL)
	assert.Equal(t, paymenttx.StatusFailed, tx.Status)
	assert.Equal(t, "transaction builder unavailable", tx.ErrorMessage)
}

// --- failTx with multiple errors ---

func TestService_failTx_PreservesErrorMessage(t *testing.T) {
	client := testEntClient(t)
	ctx := context.Background()

	svc := &Service{store: NewEntTxStore(client)}

	tests := []struct {
		give string
	}{
		{give: "nonce too low"},
		{give: "insufficient funds for gas"},
		{give: "execution reverted: ERC20: transfer amount exceeds balance"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			ptx := client.PaymentTx.Create().
				SetFromAddress(validAddr).
				SetToAddress("0xabcdef1234567890abcdef1234567890abcdef12").
				SetAmount("1.00").
				SetChainID(84532).
				SetStatus(paymenttx.StatusPending).
				SaveX(ctx)

			svc.failTx(ctx, ptx.ID, errors.New(tt.give))

			updated := client.PaymentTx.GetX(ctx, ptx.ID)
			assert.Equal(t, paymenttx.StatusFailed, updated.Status)
			assert.Equal(t, tt.give, updated.ErrorMessage)
		})
	}
}

// --- History ordering ---

func TestService_History_OrderByCreatedAtDesc(t *testing.T) {
	client := testEntClient(t)
	ctx := context.Background()

	// Create records — ent auto-sets created_at, so order depends on insertion.
	for i := 0; i < 3; i++ {
		client.PaymentTx.Create().
			SetFromAddress(validAddr).
			SetToAddress("0xabcdef1234567890abcdef1234567890abcdef12").
			SetAmount(fmt.Sprintf("%d.00", i+1)).
			SetChainID(84532).
			SetStatus(paymenttx.StatusSubmitted).
			SaveX(ctx)
	}

	svc := &Service{store: NewEntTxStore(client)}

	result, err := svc.History(ctx, 10)
	require.NoError(t, err)
	require.Len(t, result, 3)

	// Most recent first: 3.00, 2.00, 1.00
	assert.Equal(t, "3.00", result[0].Amount)
	assert.Equal(t, "2.00", result[1].Amount)
	assert.Equal(t, "1.00", result[2].Amount)
}

// --- History with empty database ---

func TestService_History_EmptyDatabase(t *testing.T) {
	client := testEntClient(t)
	svc := &Service{store: NewEntTxStore(client)}

	result, err := svc.History(context.Background(), 10)
	require.NoError(t, err)
	assert.Empty(t, result)
}

// --- RecordX402Payment duplicate records ---

func TestService_RecordX402Payment_MultipleSameURL(t *testing.T) {
	client := testEntClient(t)
	ctx := context.Background()

	svc := &Service{store: NewEntTxStore(client)}

	for i := 0; i < 3; i++ {
		err := svc.RecordX402Payment(ctx, X402PaymentRecord{
			URL:     "https://api.example.com/resource",
			Amount:  "0.01",
			From:    validAddr,
			To:      "0xabcdef1234567890abcdef1234567890abcdef12",
			ChainID: 84532,
		})
		require.NoError(t, err)
	}

	txs, err := client.PaymentTx.Query().All(ctx)
	require.NoError(t, err)
	assert.Len(t, txs, 3)
}

// --- DefaultHistoryLimit constant ---

func TestDefaultHistoryLimit(t *testing.T) {
	assert.Equal(t, 20, DefaultHistoryLimit)
}

// helpers

func strPtr(s string) *string {
	return &s
}

func ptrAddress(addr common.Address) *common.Address {
	return &addr
}

// Ensure mock types satisfy interfaces at compile time.
var _ = (*mockWallet)(nil)
var _ = (*mockLimiter)(nil)
