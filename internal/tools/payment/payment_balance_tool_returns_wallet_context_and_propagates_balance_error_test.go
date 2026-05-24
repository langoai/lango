package payment

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	corepayment "github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/paymentgate"
	"github.com/langoai/lango/internal/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentBalanceTool_ReturnsWalletContextAndPropagatesBalanceError(t *testing.T) {
	t.Parallel()

	svc := &paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorPaymentService{
		balance: "12.34",
		address: "0x3333333333333333333333333333333333333333",
		chainID: 8453,
	}
	tools := BuildTools(svc, nil, nil, 8453, nil, nil, &fakeExecutionAuditor{})
	balanceTool := findTool(tools, "payment_balance")
	require.NotNil(t, balanceTool)

	result, err := balanceTool.Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)

	payload, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "12.34", payload["balance"])
	assert.Equal(t, wallet.CurrencyUSDC, payload["currency"])
	assert.Equal(t, "0x3333333333333333333333333333333333333333", payload["address"])
	assert.Equal(t, int64(8453), payload["chainId"])
	assert.Equal(t, wallet.NetworkName(8453), payload["network"])

	svc.balanceErr = errors.New("balance unavailable")
	got, err := balanceTool.Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "balance unavailable")
}

func TestPaymentHistoryTool_UsesDefaultAndExplicitLimits(t *testing.T) {
	t.Parallel()

	svc := &paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorPaymentService{
		chainID: 84532,
		history: []corepayment.TransactionInfo{
			{
				TxHash:    "0xabc",
				Status:    "confirmed",
				Amount:    "1.25",
				From:      "0x1111111111111111111111111111111111111111",
				To:        "0x2222222222222222222222222222222222222222",
				ChainID:   84532,
				CreatedAt: time.Unix(1700000000, 0).UTC(),
			},
		},
	}
	tools := BuildTools(svc, nil, nil, 84532, nil, nil, &fakeExecutionAuditor{})
	historyTool := findTool(tools, "payment_history")
	require.NotNil(t, historyTool)

	result, err := historyTool.Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	payload, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 20, svc.lastHistoryLimit)
	assert.Equal(t, 1, payload["count"])
	assert.Equal(t, svc.history, payload["transactions"])

	_, err = historyTool.Handler(context.Background(), map[string]interface{}{"limit": float64(7)})
	require.NoError(t, err)
	assert.Equal(t, 7, svc.lastHistoryLimit)
}

func TestPaymentLimitsTool_ReturnsGenericAndEntLimiterFields(t *testing.T) {
	t.Parallel()

	genericLimiter := &paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorLimiter{
		spent:     paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorMustParseUSDC(t, "2.50"),
		remaining: paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorMustParseUSDC(t, "17.50"),
	}
	tools := BuildTools(&paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorPaymentService{chainID: 84532}, genericLimiter, nil, 84532, nil, nil, &fakeExecutionAuditor{})
	limitsTool := findTool(tools, "payment_limits")
	require.NotNil(t, limitsTool)

	result, err := limitsTool.Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	payload, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "2.50", payload["dailySpent"])
	assert.Equal(t, "17.50", payload["dailyRemaining"])
	assert.Equal(t, wallet.CurrencyUSDC, payload["currency"])
	assert.NotContains(t, payload, "maxPerTx")

	entLimiter, err := wallet.NewStoreSpendingLimiter(
		paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorUsageStore{amounts: []string{"1.00", "bad-amount", "0.75"}},
		"5.00",
		"20.00",
		"1.00",
	)
	require.NoError(t, err)
	tools = BuildTools(&paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorPaymentService{chainID: 84532}, entLimiter, nil, 84532, nil, nil, &fakeExecutionAuditor{})
	limitsTool = findTool(tools, "payment_limits")
	require.NotNil(t, limitsTool)

	result, err = limitsTool.Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	payload, ok = result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "5.00", payload["maxPerTx"])
	assert.Equal(t, "20.00", payload["maxDaily"])
	assert.Equal(t, "1.75", payload["dailySpent"])
	assert.Equal(t, "18.25", payload["dailyRemaining"])
}

func TestPaymentLimitsTool_PropagatesLimiterErrors(t *testing.T) {
	t.Parallel()

	limitsTool := buildLimitsTool(&paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorLimiter{spentErr: errors.New("spent failed")})
	got, err := limitsTool.Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "get daily spent: spent failed")

	limitsTool = buildLimitsTool(&paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorLimiter{spent: big.NewInt(0), remainingErr: errors.New("remaining failed")})
	got, err = limitsTool.Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "get daily remaining: remaining failed")
}

func TestPaymentWalletInfoTool_ReturnsNetworkAndPropagatesAddressError(t *testing.T) {
	t.Parallel()

	svc := &paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorPaymentService{
		address: "0x4444444444444444444444444444444444444444",
		chainID: 8453,
	}
	walletInfoTool := buildWalletInfoTool(svc)

	result, err := walletInfoTool.Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	payload, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "0x4444444444444444444444444444444444444444", payload["address"])
	assert.Equal(t, int64(8453), payload["chainId"])
	assert.Equal(t, wallet.NetworkName(8453), payload["network"])

	svc.walletErr = errors.New("wallet missing")
	got, err := walletInfoTool.Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "wallet missing")
}

func TestPaymentExecutionDeniedMessage_Variants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		reason paymentgate.DenyReason
		txID   string
		subID  string
		want   string
	}{
		{
			name:   "missing transaction receipt id",
			reason: paymentgate.ReasonMissingReceipt,
			want:   "transaction_receipt_id is required",
		},
		{
			name:   "missing canonical submission",
			reason: paymentgate.ReasonMissingReceipt,
			txID:   "tx-1",
			want:   "current canonical submission receipt was not found",
		},
		{
			name:   "missing explicit submission",
			reason: paymentgate.ReasonMissingReceipt,
			txID:   "tx-1",
			subID:  "submission-1",
			want:   "submission_receipt_id was not found",
		},
		{
			name:   "approval not approved",
			reason: paymentgate.ReasonApprovalNotApproved,
			want:   "canonical payment approval is not approved",
		},
		{
			name:   "execution mode mismatch",
			reason: paymentgate.ReasonExecutionModeMismatch,
			want:   "canonical settlement hint must be prepay",
		},
		{
			name:   "stale state",
			reason: paymentgate.ReasonStaleState,
			want:   "canonical payment state is stale",
		},
		{
			name:   "unknown reason",
			reason: paymentgate.DenyReason("unknown"),
			want:   "direct payment execution denied",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := PaymentExecutionDeniedMessage(tt.reason, tt.txID, tt.subID)
			assert.Contains(t, got, tt.want)
		})
	}
}

type paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorPaymentService struct {
	balance          string
	balanceErr       error
	history          []corepayment.TransactionInfo
	historyErr       error
	lastHistoryLimit int
	address          string
	walletErr        error
	chainID          int64
}

func (s *paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorPaymentService) Send(context.Context, corepayment.PaymentRequest) (*corepayment.PaymentReceipt, error) {
	return &corepayment.PaymentReceipt{}, nil
}

func (s *paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorPaymentService) Balance(context.Context) (string, error) {
	if s.balanceErr != nil {
		return "", s.balanceErr
	}
	return s.balance, nil
}

func (s *paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorPaymentService) History(_ context.Context, limit int) ([]corepayment.TransactionInfo, error) {
	s.lastHistoryLimit = limit
	if s.historyErr != nil {
		return nil, s.historyErr
	}
	return s.history, nil
}

func (s *paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorPaymentService) WalletAddress(context.Context) (string, error) {
	if s.walletErr != nil {
		return "", s.walletErr
	}
	return s.address, nil
}

func (s *paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorPaymentService) ChainID() int64 {
	return s.chainID
}

func (s *paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorPaymentService) RecordX402Payment(context.Context, corepayment.X402PaymentRecord) error {
	return nil
}

type paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorLimiter struct {
	spent        *big.Int
	spentErr     error
	remaining    *big.Int
	remainingErr error
}

func (l *paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorLimiter) Check(context.Context, *big.Int) error {
	return nil
}

func (l *paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorLimiter) Record(context.Context, *big.Int) error {
	return nil
}

func (l *paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorLimiter) DailySpent(context.Context) (*big.Int, error) {
	if l.spentErr != nil {
		return nil, l.spentErr
	}
	return new(big.Int).Set(l.spent), nil
}

func (l *paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorLimiter) DailyRemaining(context.Context) (*big.Int, error) {
	if l.remainingErr != nil {
		return nil, l.remainingErr
	}
	return new(big.Int).Set(l.remaining), nil
}

func (l *paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorLimiter) IsAutoApprovable(context.Context, *big.Int) (bool, error) {
	return false, nil
}

type paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorUsageStore struct {
	amounts []string
}

func (s paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorUsageStore) DailySpendSince(context.Context, time.Time) ([]string, error) {
	return s.amounts, nil
}

func paymentBalanceToolReturnsWalletContextAndPropagatesBalanceErrorMustParseUSDC(t *testing.T, amount string) *big.Int {
	t.Helper()

	parsed, err := wallet.ParseUSDC(amount)
	require.NoError(t, err)
	return parsed
}
