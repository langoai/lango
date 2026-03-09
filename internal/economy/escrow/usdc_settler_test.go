package escrow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestUSDCSettler_Interface(t *testing.T) {
	t.Parallel()

	// Compile-time interface check — verifies USDCSettler satisfies SettlementExecutor.
	var _ SettlementExecutor = (*USDCSettler)(nil)
}

func TestNewUSDCSettler_Defaults(t *testing.T) {
	t.Parallel()

	s := NewUSDCSettler(nil, nil, nil, 8453)
	require.NotNil(t, s)
	assert.Equal(t, 2*time.Minute, s.receiptTimeout)
	assert.Equal(t, 3, s.maxRetries)
	assert.NotNil(t, s.logger)
	assert.Equal(t, int64(8453), s.chainID.Int64())
}

func TestNewUSDCSettler_WithOptions(t *testing.T) {
	t.Parallel()

	customLogger := zap.NewExample().Sugar()

	tests := []struct {
		give        string
		giveOpts    []USDCSettlerOption
		wantTimeout time.Duration
		wantRetries int
	}{
		{
			give:        "custom timeout",
			giveOpts:    []USDCSettlerOption{WithReceiptTimeout(5 * time.Minute)},
			wantTimeout: 5 * time.Minute,
			wantRetries: 3,
		},
		{
			give:        "custom retries",
			giveOpts:    []USDCSettlerOption{WithMaxRetries(5)},
			wantTimeout: 2 * time.Minute,
			wantRetries: 5,
		},
		{
			give:        "all options",
			giveOpts:    []USDCSettlerOption{WithReceiptTimeout(30 * time.Second), WithMaxRetries(1), WithLogger(customLogger)},
			wantTimeout: 30 * time.Second,
			wantRetries: 1,
		},
		{
			give:        "zero values ignored",
			giveOpts:    []USDCSettlerOption{WithReceiptTimeout(0), WithMaxRetries(0)},
			wantTimeout: 2 * time.Minute,
			wantRetries: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			s := NewUSDCSettler(nil, nil, nil, 84532, tt.giveOpts...)
			require.NotNil(t, s)
			assert.Equal(t, tt.wantTimeout, s.receiptTimeout)
			assert.Equal(t, tt.wantRetries, s.maxRetries)
		})
	}
}
