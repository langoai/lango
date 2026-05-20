package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/agentregistry"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/skill"
)

func TestViewSkillRejectsSymlinkEscape(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	skillsDir := filepath.Join(t.TempDir(), "skills")
	store := skill.NewFileSkillStore(skillsDir, zap.NewNop().Sugar())
	registry := skill.NewRegistry(store, []*agent.Tool{{Name: "viewSkillRejectsSymlinkEscape_base_tool"}}, zap.NewNop().Sugar())
	require.NoError(t, registry.CreateSkill(ctx, skill.SkillEntry{
		Name:        "viewSkillRejectsSymlinkEscape8-skill",
		Description: "Symlink escape fixture skill",
		Type:        skill.SkillTypeInstruction,
		Context:     "safe skill context",
	}))
	require.NoError(t, registry.ActivateSkill(ctx, "viewSkillRejectsSymlinkEscape8-skill"))

	outside := filepath.Join(t.TempDir(), "outside.md")
	require.NoError(t, os.WriteFile(outside, []byte("outside"), 0o600))
	require.NoError(t, os.Symlink(outside, filepath.Join(skillsDir, "viewSkillRejectsSymlinkEscape8-skill", "leak.md")))

	tool := findTool(
		buildMetaTools(nil, nil, registry, config.SkillConfig{SkillsDir: skillsDir}, nil, nil),
		"view_skill",
	)
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"name": "viewSkillRejectsSymlinkEscape8-skill",
		"path": "leak.md",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, `path "leak.md" escapes the skill directory`)
}

func TestExtensionModuleInitObservabilityEnabledWithoutExternalServers(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.MCP.Enabled = false
	cfg.Observability.Enabled = true
	cfg.Observability.Health.Enabled = true
	cfg.Observability.Tokens.Enabled = true
	cfg.Observability.Tokens.PersistHistory = false
	cfg.Observability.Tokens.RetentionDays = 30

	result, err := (&extensionModule{cfg: cfg, bus: eventbus.New()}).Init(
		context.Background(),
		staticResolver{},
	)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotNil(t, result.Values[appinit.ProvidesObservability])
	assert.Nil(t, result.Values[appinit.ProvidesMCP])
	assert.Empty(t, result.Components)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "mcp").Enabled)
	assert.Nil(t, networkModuleDisabledFlagSkipsRuntimeServicesCatalogEntry(result.CatalogEntries, "observability"))
}

func TestInitAgentRegistryIgnoresBrokenUserStoreAndKeepsEmbeddedAgents(t *testing.T) {
	t.Parallel()

	agentsPath := filepath.Join(t.TempDir(), "agents-file")
	require.NoError(t, os.WriteFile(agentsPath, []byte("not a directory"), 0o600))

	registry, err := initAgentRegistry(&config.AgentConfig{AgentsDir: agentsPath})

	require.NoError(t, err)
	require.NotNil(t, registry)
	embedded, ok := registry.Get("planner")
	require.True(t, ok, "broken user agent stores should not suppress embedded agents")
	assert.Equal(t, agentregistry.SourceEmbedded, embedded.Source)
	assert.Equal(t, agentregistry.StatusActive, embedded.Status)
}

func TestWalletHandshakeSignerDelegatesDirectMethodsWithoutNetwork(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("wallet unavailable")
	signer := &walletHandshakeSigner{wp: &wiringP2PWallet{
		signErr: wantErr,
		pubErr:  wantErr,
	}}

	signature, err := signer.SignMessage(context.Background(), []byte("challenge"))
	require.ErrorIs(t, err, wantErr)
	assert.Nil(t, signature)

	publicKey, err := signer.PublicKey(context.Background())
	require.ErrorIs(t, err, wantErr)
	assert.Nil(t, publicKey)
	assert.Equal(t, security.AlgorithmSecp256k1Keccak256, signer.Algorithm())
}
