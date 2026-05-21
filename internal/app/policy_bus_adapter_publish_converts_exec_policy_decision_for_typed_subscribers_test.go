package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/lifecycle"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/testutil"
	execpkg "github.com/langoai/lango/internal/tools/exec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyBusAdapter_PublishConvertsExecPolicyDecisionForTypedSubscribers(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	var got eventbus.PolicyDecisionEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.PolicyDecisionEvent) {
		got = evt
	})

	adapter := &policyBusAdapter{bus: bus}
	adapter.Publish(execpkg.PolicyDecisionData{
		Command:    "rm -rf /tmp/example",
		Unwrapped:  "rm -rf /tmp/example",
		Verdict:    "block",
		Reason:     string(execpkg.ReasonKillVerb),
		Message:    "blocked destructive command",
		SessionKey: "session-1",
		AgentName:  "agent-1",
	})

	assert.Equal(t, eventbus.PolicyDecisionEvent{
		Command:    "rm -rf /tmp/example",
		Unwrapped:  "rm -rf /tmp/example",
		Verdict:    "block",
		Reason:     string(execpkg.ReasonKillVerb),
		Message:    "blocked destructive command",
		SessionKey: "session-1",
		AgentName:  "agent-1",
	}, got)
}

func TestPolicyBusAdapter_PublishForwardsNonPolicyDecisionEvents(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	var got appPolicyBusAdapterPublishConvertsExecPolicyDecisionForTypedSubscribersPolicyEvent
	bus.Subscribe(eventbus.EventPolicyDecision, func(evt eventbus.Event) {
		got = evt.(appPolicyBusAdapterPublishConvertsExecPolicyDecisionForTypedSubscribersPolicyEvent)
	})

	adapter := &policyBusAdapter{bus: bus}
	adapter.Publish(appPolicyBusAdapterPublishConvertsExecPolicyDecisionForTypedSubscribersPolicyEvent{value: "raw-event"})

	assert.Equal(t, appPolicyBusAdapterPublishConvertsExecPolicyDecisionForTypedSubscribersPolicyEvent{value: "raw-event"}, got)
}

func TestWirePostAgent_RegistersObservabilityMetricsRoutesWhenCollectorAvailable(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	application := &App{
		Config:  cfg,
		Gateway: initGateway(cfg, nil, &stubSessionStore{}, nil),
	}

	wirePostAgent(
		application,
		staticResolver{
			appinit.ProvidesObservability: &observabilityComponents{
				collector: observability.NewCollector(),
			},
		},
		nil,
		eventbus.New(),
		approval.NewCompositeProvider(),
		approval.NewGrantStore(),
		nil,
		nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	application.Gateway.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"toolExecutions"`)
	assert.Contains(t, rec.Body.String(), `"tokenUsage"`)
}

func TestWirePostAgent_RegistersObservabilityAlertsAndAuditRecorder(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Observability.Audit.Enabled = true
	client := testutil.TestEntClient(t)
	boot := &bootstrap.Result{
		Storage: storage.NewFacade(nil, nil, storage.WithEntClient(client)),
	}
	bus := eventbus.New()
	application := &App{
		Config:  cfg,
		Gateway: initGateway(cfg, nil, &stubSessionStore{}, nil),
	}

	wirePostAgent(
		application,
		staticResolver{
			appinit.ProvidesObservability: &observabilityComponents{
				collector: observability.NewCollector(),
			},
		},
		nil,
		bus,
		approval.NewCompositeProvider(),
		approval.NewGrantStore(),
		boot,
		nil,
	)

	bus.Publish(eventbus.AlertEvent{
		Type:       "wire_post_agent_alerts",
		Severity:   "warning",
		Message:    "storage-backed alert",
		SessionKey: "wire-post-agent-alerts-session",
		Timestamp:  time.Now().UTC(),
	})

	req := httptest.NewRequest(http.MethodGet, "/alerts?days=1", nil)
	rec := httptest.NewRecorder()
	application.Gateway.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"total":1`)
	assert.Contains(t, rec.Body.String(), `"type":"wire_post_agent_alerts"`)
	assert.Contains(t, rec.Body.String(), `"message":"storage-backed alert"`)
}

func TestRegisterPostBuildLifecycle_ChannelStartAndStopAreLifecycleManaged(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Server.Host = "127.0.0.1"
	cfg.Server.Port = 0

	ch := &appPolicyBusAdapterPublishConvertsExecPolicyDecisionForTypedSubscribersRecordingChannel{
		started: make(chan struct{}),
		stopped: make(chan struct{}),
	}
	application := &App{
		Gateway:  initGateway(cfg, nil, &stubSessionStore{}, nil),
		Channels: []Channel{ch},
		registry: lifecycle.NewRegistry(),
	}
	registerPostBuildLifecycle(application)

	require.NoError(t, application.Start(context.Background()))
	t.Cleanup(func() {
		_ = application.Stop(context.Background())
	})

	select {
	case <-ch.started:
	case <-time.After(time.Second):
		t.Fatal("channel was not started by lifecycle registration")
	}

	require.NoError(t, application.Stop(context.Background()))

	select {
	case <-ch.stopped:
	case <-time.After(time.Second):
		t.Fatal("channel was not stopped by lifecycle registration")
	}
	assert.Equal(t, 1, ch.startCount)
	assert.Equal(t, 1, ch.stopCount)
}

func TestAppStop_JoinsLifecycleStopAndWaitErrors(t *testing.T) {
	t.Parallel()

	stopErr := errors.New("stop failed")
	ctx, cancelApp := context.WithCancel(context.Background())
	defer cancelApp()

	stopCalled := make(chan struct{})
	application := &App{
		ctx:      ctx,
		cancel:   cancelApp,
		registry: lifecycle.NewRegistry(),
	}
	application.registry.Register(lifecycle.NewFuncComponent(
		"failing-stop",
		func(context.Context, *sync.WaitGroup) error { return nil },
		func(context.Context) error {
			close(stopCalled)
			return stopErr
		},
	), lifecycle.PriorityCore)
	require.NoError(t, application.Start(context.Background()))

	application.wg.Add(1)
	stopCtx, cancelStop := context.WithCancel(context.Background())
	defer cancelStop()
	go func() {
		<-stopCalled
		cancelStop()
	}()

	err := application.Stop(stopCtx)
	application.wg.Done()

	require.Error(t, err)
	assert.ErrorIs(t, err, stopErr)
	assert.ErrorIs(t, err, context.Canceled)
	assert.ErrorIs(t, application.ctx.Err(), context.Canceled)
}

type appPolicyBusAdapterPublishConvertsExecPolicyDecisionForTypedSubscribersPolicyEvent struct {
	value string
}

func (e appPolicyBusAdapterPublishConvertsExecPolicyDecisionForTypedSubscribersPolicyEvent) EventName() string {
	return eventbus.EventPolicyDecision
}

type appPolicyBusAdapterPublishConvertsExecPolicyDecisionForTypedSubscribersRecordingChannel struct {
	mu         sync.Mutex
	started    chan struct{}
	stopped    chan struct{}
	startCount int
	stopCount  int
}

func (c *appPolicyBusAdapterPublishConvertsExecPolicyDecisionForTypedSubscribersRecordingChannel) Name() string {
	return "recording"
}

func (c *appPolicyBusAdapterPublishConvertsExecPolicyDecisionForTypedSubscribersRecordingChannel) Start(context.Context) error {
	c.mu.Lock()
	c.startCount++
	if c.startCount == 1 {
		close(c.started)
	}
	c.mu.Unlock()
	return nil
}

func (c *appPolicyBusAdapterPublishConvertsExecPolicyDecisionForTypedSubscribersRecordingChannel) Stop(context.Context) error {
	c.mu.Lock()
	c.stopCount++
	if c.stopCount == 1 {
		close(c.stopped)
	}
	c.mu.Unlock()
	return nil
}
