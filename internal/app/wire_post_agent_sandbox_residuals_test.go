package app

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	p2pproto "github.com/langoai/lango/internal/p2p/protocol"
)

func TestWirePostAgentToolIsolationSubprocessWiresSandboxExecutor(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.ToolIsolation.Enabled = true
	cfg.P2P.ToolIsolation.Container.Enabled = false

	handler := p2pproto.NewHandler(p2pproto.HandlerConfig{})
	application := &App{
		Config:  cfg,
		Gateway: initGateway(cfg, nil, &stubSessionStore{}, nil),
	}

	wirePostAgent(
		application,
		staticResolver{appinit.ProvidesP2P: &p2pComponents{handler: handler}},
		nil,
		eventbus.New(),
		approval.NewCompositeProvider(),
		approval.NewGrantStore(),
		nil,
		nil,
	)

	value := reflect.ValueOf(handler).Elem()
	assert.False(t, value.FieldByName("sandboxExec").IsNil())
	assert.False(t, value.FieldByName("executor").IsNil())
}

func TestWirePostAgentToolIsolationContainerNativeWiresSandboxExecutor(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.ToolIsolation.Enabled = true
	cfg.P2P.ToolIsolation.Container.Enabled = true
	cfg.P2P.ToolIsolation.Container.Runtime = "native"
	cfg.P2P.ToolIsolation.Container.RequireContainer = false

	handler := p2pproto.NewHandler(p2pproto.HandlerConfig{})
	application := &App{
		Config:  cfg,
		Gateway: initGateway(cfg, nil, &stubSessionStore{}, nil),
	}

	wirePostAgent(
		application,
		staticResolver{appinit.ProvidesP2P: &p2pComponents{handler: handler}},
		nil,
		eventbus.New(),
		approval.NewCompositeProvider(),
		approval.NewGrantStore(),
		nil,
		nil,
	)

	value := reflect.ValueOf(handler).Elem()
	assert.False(t, value.FieldByName("sandboxExec").IsNil())
}

func TestWirePostAgentToolIsolationRequiredContainerLeavesSandboxUnset(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.ToolIsolation.Enabled = true
	cfg.P2P.ToolIsolation.Container.Enabled = true
	cfg.P2P.ToolIsolation.Container.Runtime = "native"
	cfg.P2P.ToolIsolation.Container.RequireContainer = true

	handler := p2pproto.NewHandler(p2pproto.HandlerConfig{})
	application := &App{
		Config:  cfg,
		Gateway: initGateway(cfg, nil, &stubSessionStore{}, nil),
	}

	wirePostAgent(
		application,
		staticResolver{appinit.ProvidesP2P: &p2pComponents{handler: handler}},
		nil,
		eventbus.New(),
		approval.NewCompositeProvider(),
		approval.NewGrantStore(),
		nil,
		nil,
	)

	value := reflect.ValueOf(handler).Elem()
	assert.True(t, value.FieldByName("sandboxExec").IsNil())
	assert.False(t, value.FieldByName("executor").IsNil())
}
