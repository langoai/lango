package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/librarian"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/p2p/handshake"
	p2pproto "github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/skill"
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/langoai/lango/internal/turnrunner"
)

func TestWave51WirePostAgentP2PFallbackApprovalGrantsBeforeSandboxDenial(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.ToolIsolation.Enabled = false
	cfg.P2P.MaxSafetyLevel = "dangerous"
	sessions, err := handshake.NewSessionStore(time.Minute)
	require.NoError(t, err)
	peerSession, err := sessions.Create("did:lango:wave51-peer", false)
	require.NoError(t, err)

	handler := p2pproto.NewHandler(p2pproto.HandlerConfig{Sessions: sessions})
	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{Name: "wave51", Description: "Wave 51", Enabled: true})
	catalog.Register("wave51", []*agent.Tool{{
		Name:        "wave51_dangerous",
		Description: "requires explicit owner approval",
		SafetyLevel: agent.SafetyLevelDangerous,
	}})

	composite := approval.NewCompositeProvider()
	approver := &wave51ApprovalProvider{approved: true}
	composite.SetP2PFallback(approver)
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
			appinit.ProvidesPayment: &paymentComponents{limiter: &wave37Limiter{}},
		},
		[]*agent.Tool{{
			Name:        "wave51_dangerous",
			Description: "requires explicit owner approval",
			SafetyLevel: agent.SafetyLevelDangerous,
			Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
				return map[string]interface{}{"executed": true}, nil
			},
		}},
		eventbus.New(),
		composite,
		grants,
		nil,
		nil,
	)

	request := p2pproto.Request{
		Type:         p2pproto.RequestToolInvoke,
		SessionToken: peerSession.Token,
		RequestID:    "wave51-owner-approved",
		Payload: map[string]interface{}{
			"toolName": "wave51_dangerous",
			"params":   map[string]interface{}{"path": "/tmp/wave51"},
		},
	}
	var input bytes.Buffer
	require.NoError(t, json.NewEncoder(&input).Encode(request))
	var output bytes.Buffer
	stream := &wave37ProtocolStream{reader: &input, writer: &output}

	handler.StreamHandler()(stream)

	var response p2pproto.Response
	require.NoError(t, json.NewDecoder(&output).Decode(&response))
	assert.Equal(t, p2pproto.ResponseStatusDenied, response.Status)
	assert.Equal(t, p2pproto.ErrNoSandboxExecutor.Error(), response.Error)
	assert.True(t, grants.IsGranted("p2p:did:lango:wave51-peer", "wave51_dangerous"))
	require.Len(t, approver.requests, 1)
	assert.Equal(t, "wave51_dangerous", approver.requests[0].ToolName)
	assert.Equal(t, "p2p:did:lango:wave51-peer", approver.requests[0].SessionKey)
	assert.True(t, stream.closed)
}

func TestWave51WireMemoryAndTurnCallbacksQueuesMemoryAndLibrarianBuffers(t *testing.T) {
	t.Parallel()

	memBuffer := memory.NewBuffer(
		nil,
		nil,
		nil,
		1,
		1,
		func(string) ([]session.Message, error) { return nil, nil },
		zap.NewNop().Sugar(),
	)
	libBuffer := librarian.NewProactiveBuffer(
		nil,
		nil,
		nil,
		nil,
		func(string) ([]session.Message, error) { return nil, nil },
		nil,
		librarian.ProactiveBufferConfig{},
		zap.NewNop().Sugar(),
	)
	runner := turnrunner.New(
		turnrunner.Config{HardCeiling: time.Second, StaleTimeout: -1},
		&wave37TurnExecutor{response: "wave51 complete"},
		newWave37SessionStore(),
		nil,
	)
	application := &App{
		MemoryBuffer:             memBuffer,
		LibrarianProactiveBuffer: libBuffer,
		TurnRunner:               runner,
	}

	wireMemoryAndTurnCallbacks(application, nil, &foundationValues{Store: &stubSessionStore{}})
	result, err := runner.Run(context.Background(), turnrunner.Request{
		SessionKey: "wave51-session",
		Input:      "finish turn",
	})

	require.NoError(t, err)
	assert.Equal(t, "wave51 complete", result.ResponseText)
	assert.Equal(t, 1, wave51TriggerQueueLen(t, memBuffer))
	assert.Equal(t, 1, wave51TriggerQueueLen(t, libBuffer))
}

func TestWave51ExtensionModuleObservabilityWiresSynchronousMetrics(t *testing.T) {
	t.Parallel()

	cfg := wave40ModuleConfig(t)
	cfg.Observability.Enabled = true
	cfg.Observability.Health.Enabled = true
	cfg.Observability.Tokens.Enabled = true
	cfg.Observability.Metrics.Format = "prometheus"
	cfg.Alerting.Enabled = true
	bus := eventbus.New()

	result, err := (&extensionModule{cfg: cfg, bus: bus}).Init(context.Background(), staticResolver{})
	require.NoError(t, err)
	require.NotNil(t, result)

	obsc, ok := result.Values[appinit.ProvidesObservability].(*observabilityComponents)
	require.True(t, ok)
	require.NotNil(t, obsc)
	require.NotNil(t, obsc.collector)
	require.NotNil(t, obsc.healthRegistry)
	require.NotNil(t, obsc.tracker)
	require.NotNil(t, obsc.promExporter)
	assert.Nil(t, wave51CatalogEntry(result.CatalogEntries, "observability"))
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "mcp").Enabled)

	bus.Publish(eventbus.TokenUsageEvent{
		Provider:     "wave51-provider",
		Model:        "wave51-model",
		SessionKey:   "wave51-session",
		AgentName:    "operator",
		InputTokens:  11,
		OutputTokens: 7,
		TotalTokens:  18,
	})
	bus.Publish(eventbus.PolicyDecisionEvent{Verdict: "block", Reason: "protected_path"})
	bus.Publish(toolchain.ToolExecutedEvent{
		ToolName:  "wave51_tool",
		AgentName: "operator",
		Duration:  time.Millisecond,
		Success:   false,
		Error:     "denied",
	})

	snapshot := obsc.collector.Snapshot()
	assert.Equal(t, int64(18), snapshot.TokenUsageTotal.TotalTokens)
	assert.Equal(t, int64(1), snapshot.Policy.Blocks)
	assert.Equal(t, int64(1), snapshot.Policy.ByReason["protected_path"])
	assert.Equal(t, int64(1), snapshot.ToolExecutions)
	assert.Equal(t, int64(1), snapshot.ToolBreakdown["wave51_tool"].Count)
}

func TestWave51ViewSkillReadsExtensionPackSkillRoot(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	skillsDir := t.TempDir()
	packSkillDir := filepath.Join(skillsDir, "ext-wave51-pack", "packed-skill")
	require.NoError(t, os.MkdirAll(packSkillDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(packSkillDir, "SKILL.md"), []byte(`---
name: packed-skill
description: Packed Wave 51 skill
---

# Packed Skill

Read from an extension pack.
`), 0o600))

	logger := zap.NewNop().Sugar()
	store := skill.NewFileSkillStore(skillsDir, logger)
	store.AllowedExtPacks = map[string]bool{"wave51-pack": true}
	registry := skill.NewRegistry(store, []*agent.Tool{{Name: "wave51_base"}}, logger)
	tool := findTool(
		buildMetaTools(nil, nil, registry, config.SkillConfig{SkillsDir: skillsDir}, nil, nil),
		"view_skill",
	)
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{"name": "packed-skill"})

	require.NoError(t, err)
	payload, ok := got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "packed-skill", payload["name"])
	assert.Contains(t, payload["content"], "Read from an extension pack.")
	wantPath, err := filepath.EvalSymlinks(filepath.Join(packSkillDir, "SKILL.md"))
	require.NoError(t, err)
	assert.Equal(t, wantPath, payload["path"])
}

func TestWave51WalletHandshakeSignerPropagatesWalletErrors(t *testing.T) {
	t.Parallel()

	signErr := errors.New("wave51 sign unavailable")
	pubErr := errors.New("wave51 public key unavailable")
	signer := &walletHandshakeSigner{wp: &wiringP2PWallet{
		signErr: signErr,
		pubErr:  pubErr,
	}}

	signature, err := signer.SignMessage(context.Background(), []byte("challenge"))
	require.ErrorIs(t, err, signErr)
	assert.Nil(t, signature)

	pub, err := signer.PublicKey(context.Background())
	require.ErrorIs(t, err, pubErr)
	assert.Nil(t, pub)
	assert.Equal(t, "secp256k1-keccak256", signer.Algorithm())
}

func TestWave51BuildSubAgentPromptFuncFallsBackToDefaultIdentity(t *testing.T) {
	t.Parallel()

	promptsDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(promptsDir, "SAFETY.md"), []byte("Wave51 shared safety."), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(promptsDir, "agents", "other"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(promptsDir, "agents", "other", "IDENTITY.md"),
		[]byte("Other agent override must not leak."),
		0o600,
	))

	buildPrompt := buildSubAgentPromptFunc(&config.AgentConfig{PromptsDir: promptsDir})
	got := buildPrompt("wave51-reviewer", "Default Wave51 reviewer identity.")

	assert.Contains(t, got, "Default Wave51 reviewer identity.")
	assert.Contains(t, got, "Wave51 shared safety.")
	assert.NotContains(t, got, "Other agent override must not leak.")
}

type wave51ApprovalProvider struct {
	approved bool
	mu       sync.Mutex
	requests []approval.ApprovalRequest
}

func (p *wave51ApprovalProvider) RequestApproval(_ context.Context, req approval.ApprovalRequest) (approval.ApprovalResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.requests = append(p.requests, req)
	return approval.ApprovalResponse{Approved: p.approved, Provider: "wave51"}, nil
}

func (p *wave51ApprovalProvider) CanHandle(sessionKey string) bool {
	return sessionKey == "p2p:did:lango:wave51-peer"
}

func wave51TriggerQueueLen(t *testing.T, buffer interface{}) int {
	t.Helper()

	inner := reflect.ValueOf(buffer).Elem().FieldByName("inner")
	require.False(t, inner.IsNil())
	queue := inner.Elem().FieldByName("queue")
	return queue.Len()
}

func wave51CatalogEntry(entries []appinit.CatalogEntry, category string) *appinit.CatalogEntry {
	for i := range entries {
		if entries[i].Category == category {
			return &entries[i]
		}
	}
	return nil
}
