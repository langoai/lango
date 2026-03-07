package sentinel

import (
	"math/big"
	"testing"
	"time"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRapidCreationDetector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		events    int
		wantAlert bool
	}{
		{
			give:      "under threshold produces no alert",
			events:    3,
			wantAlert: false,
		},
		{
			give:      "at threshold produces no alert",
			events:    5,
			wantAlert: false,
		},
		{
			give:      "over threshold produces alert",
			events:    6,
			wantAlert: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			d := NewRapidCreationDetector(1*time.Minute, 5)
			var lastAlert *Alert

			for i := 0; i < tt.events; i++ {
				lastAlert = d.Analyze(eventbus.EscrowCreatedEvent{
					EscrowID: "escrow-" + string(rune('a'+i)),
					PayerDID: "did:peer:alice",
					Amount:   big.NewInt(1000),
				})
			}

			if tt.wantAlert {
				require.NotNil(t, lastAlert)
				assert.Equal(t, SeverityHigh, lastAlert.Severity)
				assert.Equal(t, "rapid_creation", lastAlert.Type)
				assert.Equal(t, "did:peer:alice", lastAlert.PeerDID)
			} else {
				assert.Nil(t, lastAlert)
			}
		})
	}
}

func TestRapidCreationDetector_IgnoresWrongEvent(t *testing.T) {
	t.Parallel()

	d := NewRapidCreationDetector(1*time.Minute, 5)
	alert := d.Analyze(eventbus.EscrowReleasedEvent{EscrowID: "x"})
	assert.Nil(t, alert)
}

func TestLargeWithdrawalDetector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		amount    int64
		threshold string
		wantAlert bool
	}{
		{
			give:      "under threshold",
			amount:    5000,
			threshold: "10000",
			wantAlert: false,
		},
		{
			give:      "at threshold",
			amount:    10000,
			threshold: "10000",
			wantAlert: false,
		},
		{
			give:      "over threshold",
			amount:    10001,
			threshold: "10000",
			wantAlert: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			d := NewLargeWithdrawalDetector(tt.threshold)
			alert := d.Analyze(eventbus.EscrowReleasedEvent{
				EscrowID: "escrow-1",
				Amount:   big.NewInt(tt.amount),
			})

			if tt.wantAlert {
				require.NotNil(t, alert)
				assert.Equal(t, SeverityHigh, alert.Severity)
				assert.Equal(t, "large_withdrawal", alert.Type)
			} else {
				assert.Nil(t, alert)
			}
		})
	}
}

func TestLargeWithdrawalDetector_NilAmount(t *testing.T) {
	t.Parallel()

	d := NewLargeWithdrawalDetector("10000")
	alert := d.Analyze(eventbus.EscrowReleasedEvent{
		EscrowID: "escrow-1",
		Amount:   nil,
	})
	assert.Nil(t, alert)
}

func TestRepeatedDisputeDetector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		events    int
		wantAlert bool
	}{
		{
			give:      "under threshold",
			events:    2,
			wantAlert: false,
		},
		{
			give:      "over threshold",
			events:    4,
			wantAlert: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			d := NewRepeatedDisputeDetector(1*time.Hour, 3)
			var lastAlert *Alert

			for i := 0; i < tt.events; i++ {
				lastAlert = d.Analyze(eventbus.EscrowMilestoneEvent{
					EscrowID:    "escrow-1",
					MilestoneID: "ms-" + string(rune('a'+i)),
					Index:       i,
				})
			}

			if tt.wantAlert {
				require.NotNil(t, lastAlert)
				assert.Equal(t, SeverityHigh, lastAlert.Severity)
				assert.Equal(t, "repeated_dispute", lastAlert.Type)
			} else {
				assert.Nil(t, lastAlert)
			}
		})
	}
}

func TestUnusualTimingDetector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		wantAlert bool
	}{
		{
			give:      "create then immediate release triggers alert",
			wantAlert: true,
		},
		{
			give:      "release without create is ignored",
			wantAlert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			d := NewUnusualTimingDetector(1 * time.Minute)

			if tt.wantAlert {
				// Create then immediately release.
				d.Analyze(eventbus.EscrowCreatedEvent{
					EscrowID: "escrow-1",
					PayerDID: "did:peer:alice",
					Amount:   big.NewInt(1000),
				})
				alert := d.Analyze(eventbus.EscrowReleasedEvent{
					EscrowID: "escrow-1",
					Amount:   big.NewInt(1000),
				})
				require.NotNil(t, alert)
				assert.Equal(t, SeverityMedium, alert.Severity)
				assert.Equal(t, "unusual_timing", alert.Type)
			} else {
				alert := d.Analyze(eventbus.EscrowReleasedEvent{
					EscrowID: "escrow-unknown",
					Amount:   big.NewInt(1000),
				})
				assert.Nil(t, alert)
			}
		})
	}
}

func TestBalanceDropDetector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give        string
		balances    []int64
		wantAlertAt int // -1 means no alert expected
	}{
		{
			give:        "first event sets baseline, no alert",
			balances:    []int64{1000},
			wantAlertAt: -1,
		},
		{
			give:        "small drop no alert",
			balances:    []int64{1000, 600},
			wantAlertAt: -1,
		},
		{
			give:        "drop over 50% triggers critical",
			balances:    []int64{1000, 400},
			wantAlertAt: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			d := NewBalanceDropDetector()
			var gotAlert *Alert
			alertIdx := -1

			for i, bal := range tt.balances {
				alert := d.Analyze(BalanceChangeEvent{NewBalance: big.NewInt(bal)})
				if alert != nil {
					gotAlert = alert
					alertIdx = i
				}
			}

			if tt.wantAlertAt >= 0 {
				require.NotNil(t, gotAlert)
				assert.Equal(t, SeverityCritical, gotAlert.Severity)
				assert.Equal(t, "balance_drop", gotAlert.Type)
				assert.Equal(t, tt.wantAlertAt, alertIdx)
			} else {
				assert.Nil(t, gotAlert)
			}
		})
	}
}

func TestBalanceDropDetector_IgnoresWrongEvent(t *testing.T) {
	t.Parallel()

	d := NewBalanceDropDetector()
	alert := d.Analyze(eventbus.EscrowCreatedEvent{EscrowID: "x"})
	assert.Nil(t, alert)
}
