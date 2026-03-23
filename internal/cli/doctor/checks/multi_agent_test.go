package checks

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/testutil"
	"github.com/langoai/lango/internal/turntrace"
)

func TestMultiAgentCheck_RunWithBootstrap_ShowsFailuresAndLeaks(t *testing.T) {
	t.Parallel()

	client := testutil.TestEntClient(t)
	traceStore := turntrace.NewEntStore(client)
	now := time.Now()
	require.NoError(t, traceStore.CreateTrace(context.Background(), turntrace.Trace{
		TraceID:    "trace-1",
		SessionKey: "telegram:test",
		Entrypoint: "channel",
		Outcome:    turntrace.OutcomeRunning,
		StartedAt:  now,
	}))
	require.NoError(t, traceStore.FinishTrace(
		context.Background(),
		"trace-1",
		turntrace.OutcomeLoopDetected,
		"loop_detected on payment_balance",
		"E007",
		"repeated_call_signature",
		"payment_balance {} repeated",
		now.Add(time.Second),
	))

	sess, err := client.Session.Create().
		SetKey("telegram:test").
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Save(context.Background())
	require.NoError(t, err)
	_, err = client.Message.Create().
		SetSession(sess).
		SetRole("assistant").
		SetAuthor("vault").
		SetContent("raw isolated leak").
		SetTimestamp(now).
		Save(context.Background())
	require.NoError(t, err)

	cfg := &config.Config{}
	cfg.Agent.MultiAgent = true
	cfg.Agent.Provider = "openai"

	check := &MultiAgentCheck{}
	result := check.RunWithBootstrap(context.Background(), cfg, &bootstrap.Result{
		Config:   cfg,
		DBClient: client,
	})

	assert.Equal(t, StatusWarn, result.Status)
	assert.Contains(t, result.Message, "recent failed trace")
	assert.Contains(t, result.Details, "trace-1")
	assert.Contains(t, result.Details, "E007")
	assert.Contains(t, result.Details, "repeated_call_signature")
	assert.Contains(t, result.Details, "Persisted raw isolated specialist turns detected: 1")
}
