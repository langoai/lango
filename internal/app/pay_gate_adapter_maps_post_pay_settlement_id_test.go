package app

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/handshake"
	"github.com/langoai/lango/internal/p2p/paygate"
	p2pproto "github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/toolcatalog"
)

func TestPayGateAdapterMapsPostPaySettlementID(t *testing.T) {
	t.Parallel()

	gate := paygate.New(paygate.Config{
		PricingFn: func(string) (string, bool) {
			return "0.75", false
		},
		ReputationFn: func(context.Context, string) (float64, error) {
			return 0.95, nil
		},
		TrustCfg: paygate.TrustConfig{PostPayMinScore: 0.8},
		Logger:   testLog(),
	})
	adapter := &payGateAdapter{gate: gate}

	got, err := adapter.Check("did:lango:trusted", "payGateAdapterMapsPostPaySettlementId_paid_tool", map[string]interface{}{})

	require.NoError(t, err)
	assert.Equal(t, string(paygate.StatusPostPayApproved), got.Status)
	assert.NotEmpty(t, got.SettlementID)
	assert.Nil(t, got.Auth)
	assert.Nil(t, got.PriceQuote)
}

func TestWirePostAgentRegistersP2PRoutesWhenHandlerIsNil(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	p2pc := &p2pComponents{
		pricingCfg: config.P2PPricingConfig{
			Enabled:    true,
			PerQuery:   "0.31",
			ToolPrices: map[string]string{"payGateAdapterMapsPostPaySettlementId_special": "1.23"},
		},
	}
	application := &App{
		Config:  cfg,
		Gateway: initGateway(cfg, nil, &stubSessionStore{}, nil),
	}

	wirePostAgent(
		application,
		staticResolver{appinit.ProvidesP2P: p2pc},
		nil,
		eventbus.New(),
		approval.NewCompositeProvider(),
		approval.NewGrantStore(),
		nil,
		nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/p2p/pricing?tool=missing_tool", nil)
	rec := httptest.NewRecorder()
	application.Gateway.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "missing_tool", body["tool"])
	assert.Equal(t, "0.31", body["price"])
}

func TestWirePostAgentInvalidMaxSafetyLevelDefaultsToModerate(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.ToolIsolation.Enabled = false
	cfg.P2P.MaxSafetyLevel = "not-a-safety-level"
	cfg.P2P.AllowedTools = []string{"payGateAdapterMapsPostPaySettlementId_allowlisted"}
	handler := p2pproto.NewHandler(p2pproto.HandlerConfig{})
	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{Name: "runAndCollectUsesLearnedFixRetryResponse4", Description: "Pay gate fixture", Enabled: true})
	catalog.Register("runAndCollectUsesLearnedFixRetryResponse4", []*agent.Tool{
		{Name: "payGateAdapterMapsPostPaySettlementId_moderate", Description: "moderate", SafetyLevel: agent.SafetyLevelModerate},
		{Name: "payGateAdapterMapsPostPaySettlementId_allowlisted", Description: "allowlisted", SafetyLevel: agent.SafetyLevelDangerous},
	})
	application := &App{
		Config:      cfg,
		Gateway:     initGateway(cfg, nil, &stubSessionStore{}, nil),
		ToolCatalog: catalog,
	}

	wirePostAgent(
		application,
		staticResolver{appinit.ProvidesP2P: &p2pComponents{handler: handler}},
		[]*agent.Tool{{Name: "payGateAdapterMapsPostPaySettlementId_moderate", Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"ok": true}, nil
		}}},
		eventbus.New(),
		approval.NewCompositeProvider(),
		approval.NewGrantStore(),
		nil,
		nil,
	)

	handlerValue := reflect.ValueOf(handler).Elem()
	assert.Equal(t, int64(agent.SafetyLevelModerate), handlerValue.FieldByName("maxSafetyLevel").Int())
	allowedTools := handlerValue.FieldByName("allowedTools")
	require.Len(t, cfg.P2P.AllowedTools, allowedTools.Len())
	assert.Equal(t, "payGateAdapterMapsPostPaySettlementId_allowlisted", allowedTools.Index(0).String())
	assert.False(t, handlerValue.FieldByName("safetyChecker").IsNil())
}

func TestWirePostAgentDangerousToolRequiresOwnerApprovalAndDeniesUnavailableProvider(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.ToolIsolation.Enabled = false
	cfg.P2P.MaxSafetyLevel = "dangerous"
	sessions, err := handshake.NewSessionStore(time.Minute)
	require.NoError(t, err)
	session, err := sessions.Create("did:lango:runAndCollectUsesLearnedFixRetryResponse4-peer", false)
	require.NoError(t, err)
	handler := p2pproto.NewHandler(p2pproto.HandlerConfig{Sessions: sessions})
	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{Name: "runAndCollectUsesLearnedFixRetryResponse4", Description: "Pay gate fixture", Enabled: true})
	catalog.Register("runAndCollectUsesLearnedFixRetryResponse4", []*agent.Tool{{
		Name:        "payGateAdapterMapsPostPaySettlementId_dangerous",
		Description: "dangerous",
		SafetyLevel: agent.SafetyLevelDangerous,
	}})
	grants := approval.NewGrantStore()
	application := &App{
		Config:      cfg,
		Gateway:     initGateway(cfg, nil, &stubSessionStore{}, nil),
		ToolCatalog: catalog,
	}

	wirePostAgent(
		application,
		staticResolver{
			appinit.ProvidesP2P:     &p2pComponents{handler: handler},
			appinit.ProvidesPayment: &paymentComponents{limiter: &payGateAdapterMapsPostPaySettlementIdLimiter{}},
		},
		[]*agent.Tool{{
			Name:        "payGateAdapterMapsPostPaySettlementId_dangerous",
			Description: "dangerous",
			SafetyLevel: agent.SafetyLevelDangerous,
			Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
				return map[string]interface{}{"executed": true}, nil
			},
		}},
		eventbus.New(),
		approval.NewCompositeProvider(),
		grants,
		nil,
		nil,
	)

	request := p2pproto.Request{
		Type:         p2pproto.RequestToolInvoke,
		SessionToken: session.Token,
		RequestID:    "runAndCollectUsesLearnedFixRetryResponse4-dangerous-tool",
		Payload: map[string]interface{}{
			"toolName": "payGateAdapterMapsPostPaySettlementId_dangerous",
			"params":   map[string]interface{}{"path": "/tmp/example"},
		},
	}
	var input bytes.Buffer
	require.NoError(t, json.NewEncoder(&input).Encode(request))
	var output bytes.Buffer
	stream := &initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream{reader: &input, writer: &output}

	handler.StreamHandler()(stream)

	var response p2pproto.Response
	require.NoError(t, json.NewDecoder(&output).Decode(&response))
	assert.Equal(t, p2pproto.ResponseStatusDenied, response.Status)
	assert.Equal(t, p2pproto.ErrDeniedByOwner.Error(), response.Error)
	assert.False(t, grants.IsGranted("p2p:did:lango:runAndCollectUsesLearnedFixRetryResponse4-peer", "payGateAdapterMapsPostPaySettlementId_dangerous"))
	assert.True(t, stream.closed)
}

type payGateAdapterMapsPostPaySettlementIdLimiter struct{}

func (l *payGateAdapterMapsPostPaySettlementIdLimiter) Check(context.Context, *big.Int) error {
	return nil
}
func (l *payGateAdapterMapsPostPaySettlementIdLimiter) Record(context.Context, *big.Int) error {
	return nil
}
func (l *payGateAdapterMapsPostPaySettlementIdLimiter) DailySpent(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}
func (l *payGateAdapterMapsPostPaySettlementIdLimiter) DailyRemaining(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}
func (l *payGateAdapterMapsPostPaySettlementIdLimiter) IsAutoApprovable(context.Context, *big.Int) (bool, error) {
	return false, nil
}
