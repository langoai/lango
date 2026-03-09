package sentinel

import (
	"math/big"
	"testing"
	"time"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_StartStop(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	eng := New(bus, DefaultSentinelConfig())

	require.NoError(t, eng.Start())
	status := eng.Status()
	assert.True(t, status["running"].(bool))

	// Idempotent start.
	require.NoError(t, eng.Start())

	require.NoError(t, eng.Stop())
	status = eng.Status()
	assert.False(t, status["running"].(bool))
}

func TestEngine_RapidCreation(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	cfg := DefaultSentinelConfig()
	cfg.RapidCreationMax = 3
	cfg.RapidCreationWindow = 1 * time.Minute
	eng := New(bus, cfg)
	require.NoError(t, eng.Start())

	// Publish 4 creation events from the same peer — should trigger alert.
	for i := 0; i < 4; i++ {
		bus.Publish(eventbus.EscrowCreatedEvent{
			EscrowID: "escrow-" + string(rune('a'+i)),
			PayerDID: "did:peer:spammer",
			PayeeDID: "did:peer:victim",
			Amount:   big.NewInt(100),
		})
	}

	alerts := eng.AlertsByLevel(SeverityHigh)
	require.NotEmpty(t, alerts)

	found := false
	for _, a := range alerts {
		if a.Type == "rapid_creation" {
			found = true
			assert.Equal(t, "did:peer:spammer", a.PeerDID)
		}
	}
	assert.True(t, found, "expected rapid_creation alert")
}

func TestEngine_LargeWithdrawal(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	cfg := DefaultSentinelConfig()
	cfg.LargeWithdrawalAmount = "5000"
	eng := New(bus, cfg)
	require.NoError(t, eng.Start())

	bus.Publish(eventbus.EscrowReleasedEvent{
		EscrowID: "escrow-big",
		Amount:   big.NewInt(10000),
	})

	alerts := eng.AlertsByLevel(SeverityHigh)
	require.NotEmpty(t, alerts)

	found := false
	for _, a := range alerts {
		if a.Type == "large_withdrawal" {
			found = true
			assert.Equal(t, "escrow-big", a.DealID)
		}
	}
	assert.True(t, found, "expected large_withdrawal alert")
}

func TestEngine_Acknowledge(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	cfg := DefaultSentinelConfig()
	cfg.LargeWithdrawalAmount = "100"
	eng := New(bus, cfg)
	require.NoError(t, eng.Start())

	bus.Publish(eventbus.EscrowReleasedEvent{
		EscrowID: "escrow-1",
		Amount:   big.NewInt(500),
	})

	alerts := eng.ActiveAlerts()
	require.NotEmpty(t, alerts)

	alertID := alerts[0].ID
	require.NoError(t, eng.Acknowledge(alertID))

	// After acknowledgment, active alerts should be empty.
	assert.Empty(t, eng.ActiveAlerts())

	// All alerts still contains it.
	assert.NotEmpty(t, eng.Alerts())

	// Acknowledging non-existent alert returns error.
	err := eng.Acknowledge("non-existent")
	assert.Error(t, err)
}

func TestEngine_Status(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	eng := New(bus, DefaultSentinelConfig())
	require.NoError(t, eng.Start())

	status := eng.Status()
	assert.True(t, status["running"].(bool))
	assert.Equal(t, 0, status["totalAlerts"].(int))
	assert.Equal(t, 0, status["activeAlerts"].(int))

	detectors := status["detectors"].([]string)
	assert.Len(t, detectors, 5)
	assert.Contains(t, detectors, "rapid_creation")
	assert.Contains(t, detectors, "large_withdrawal")
	assert.Contains(t, detectors, "repeated_dispute")
	assert.Contains(t, detectors, "unusual_timing")
	assert.Contains(t, detectors, "balance_drop")
}

func TestEngine_UnusualTiming(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	cfg := DefaultSentinelConfig()
	cfg.WashTradeWindow = 5 * time.Second
	eng := New(bus, cfg)
	require.NoError(t, eng.Start())

	// Create then immediately release — should detect wash trade.
	bus.Publish(eventbus.EscrowCreatedEvent{
		EscrowID: "escrow-wash",
		PayerDID: "did:peer:washer",
		PayeeDID: "did:peer:target",
		Amount:   big.NewInt(100),
	})
	bus.Publish(eventbus.EscrowReleasedEvent{
		EscrowID: "escrow-wash",
		Amount:   big.NewInt(100),
	})

	alerts := eng.AlertsByLevel(SeverityMedium)
	found := false
	for _, a := range alerts {
		if a.Type == "unusual_timing" {
			found = true
			assert.Equal(t, "escrow-wash", a.DealID)
		}
	}
	assert.True(t, found, "expected unusual_timing alert")
}

func TestEngine_Config(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	cfg := DefaultSentinelConfig()
	cfg.RapidCreationMax = 10
	eng := New(bus, cfg)

	assert.Equal(t, 10, eng.Config().RapidCreationMax)
}
