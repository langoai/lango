package app

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/lifecycle"
	"github.com/langoai/lango/internal/receipts"
)

func TestNetworkModuleInitNoPaymentEscrowRegistersEscrowAndSentinelTools(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = false
	cfg.P2P.Enabled = false
	cfg.P2P.Workspace.Enabled = false
	cfg.Economy.Enabled = true
	cfg.Economy.Escrow.Enabled = true

	result, err := (&networkModule{cfg: cfg, bus: eventbus.New()}).Init(
		context.Background(),
		staticResolver{
			appinit.ProvidesSupervisor: &foundationValues{
				ReceiptStore: receipts.NewStore(),
			},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, result)

	economyValues, ok := result.Values[appinit.ProvidesEconomy].(*economyComponents)
	require.True(t, ok)
	require.NotNil(t, economyValues)
	require.NotNil(t, economyValues.escrowEngine)
	require.NotNil(t, economyValues.escrowSettler)
	require.NotNil(t, economyValues.sentinelEngine)
	t.Cleanup(func() { _ = economyValues.sentinelEngine.Stop() })

	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "payment").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "p2p").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "economy").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "escrow").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "sentinel").Enabled)

	assert.NotNil(t, findTool(result.Tools, "economy_escrow_create"))
	assert.NotNil(t, findTool(result.Tools, "escrow_create"))
	assert.NotNil(t, findTool(result.Tools, "sentinel_status"))
}

func TestRegisterPostBuildLifecycleChannelStartErrorIsSwallowed(t *testing.T) {
	t.Parallel()

	startErr := errors.New("channel start failed")
	channel := &networkModuleAndLifecycleResidualsFailingChannel{
		started: make(chan struct{}),
		stopped: make(chan struct{}),
		err:     startErr,
	}
	application := &App{
		Channels: []Channel{channel},
		registry: lifecycle.NewRegistry(),
	}
	registerPostBuildLifecycle(application)
	assert.Equal(t, []string{"gateway", "channel-0"}, application.registry.Names())

	component := networkModuleAndLifecycleResidualsComponentByName(t, application.registry, "channel-0")
	require.NoError(t, component.Start(context.Background(), &application.wg))

	select {
	case <-channel.started:
	case <-time.After(time.Second):
		t.Fatal("channel start was not invoked")
	}

	waitDone := make(chan struct{})
	go func() {
		application.wg.Wait()
		close(waitDone)
	}()
	select {
	case <-waitDone:
	case <-time.After(time.Second):
		t.Fatal("channel start error goroutine did not drain")
	}

	require.NoError(t, component.Stop(context.Background()))
	select {
	case <-channel.stopped:
	case <-time.After(time.Second):
		t.Fatal("channel stop was not invoked")
	}
	assert.Equal(t, 1, channel.startCount)
	assert.Equal(t, 1, channel.stopCount)
}

func networkModuleAndLifecycleResidualsComponentByName(t *testing.T, registry *lifecycle.Registry, name string) lifecycle.Component {
	t.Helper()

	value := reflect.ValueOf(registry).Elem().FieldByName("entries")
	entries := reflect.NewAt(value.Type(), unsafe.Pointer(value.UnsafeAddr())).Elem()
	for i := 0; i < entries.Len(); i++ {
		entry := entries.Index(i)
		component := entry.FieldByName("Component").Interface().(lifecycle.Component)
		if component.Name() == name {
			return component
		}
	}
	t.Fatalf("component %q not registered", name)
	return nil
}

type networkModuleAndLifecycleResidualsFailingChannel struct {
	mu         sync.Mutex
	started    chan struct{}
	stopped    chan struct{}
	err        error
	startCount int
	stopCount  int
}

func (c *networkModuleAndLifecycleResidualsFailingChannel) Name() string {
	return "failing"
}

func (c *networkModuleAndLifecycleResidualsFailingChannel) Start(context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.startCount++
	if c.startCount == 1 {
		close(c.started)
	}
	return c.err
}

func (c *networkModuleAndLifecycleResidualsFailingChannel) Stop(context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stopCount++
	if c.stopCount == 1 {
		close(c.stopped)
	}
	return nil
}
