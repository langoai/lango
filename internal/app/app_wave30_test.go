package app

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/langoai/lango/internal/toolchain"
)

func TestWave30AppModeOptionsSetExpectedModes(t *testing.T) {
	t.Parallel()

	var local appOptions
	WithLocalChat()(&local)
	assert.Equal(t, AppModeLocalChat, local.mode)

	var cockpit appOptions
	WithCockpit()(&cockpit)
	assert.Equal(t, AppModeCockpit, cockpit.mode)
}

func TestWave30NewLocalChatInitializesCoreWithoutExternalListeners(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "google"
	cfg.Agent.Model = "gemini-2.0-flash"
	cfg.Providers = map[string]config.ProviderConfig{
		"google": {
			Type:   "gemini",
			APIKey: "test-key",
		},
	}
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "wave30.db")

	app, err := New(testBoot(t, cfg), WithLocalChat())
	require.NoError(t, err)
	require.NotNil(t, app)
	t.Cleanup(func() { require.NoError(t, app.Stop(context.Background())) })

	assert.NotNil(t, app.Agent)
	assert.NotNil(t, app.Store)
	assert.NotNil(t, app.Gateway)
}

func TestWave30ModeAllowlistResolverHandlesMissingInputsAndExpandsCategories(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Modes = map[string]config.SessionMode{
		"review": {
			Name:  "review",
			Tools: []string{"@code", "direct_tool", ""},
		},
		"empty": {
			Name:  "empty",
			Tools: nil,
		},
	}

	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{Name: "code", Enabled: true})
	catalog.Register("code", []*agent.Tool{
		{Name: "read_file"},
		{Name: "write_file"},
	})

	resolver := buildModeAllowlistResolver(cfg, catalog)

	allow, ok := resolver(context.Background())
	assert.False(t, ok)
	assert.Nil(t, allow)

	allow, ok = resolver(session.WithModeName(context.Background(), "missing"))
	assert.False(t, ok)
	assert.Nil(t, allow)

	allow, ok = resolver(session.WithModeName(context.Background(), "empty"))
	assert.False(t, ok)
	assert.Nil(t, allow)

	allow, ok = resolver(session.WithModeName(context.Background(), "review"))
	require.True(t, ok)
	assert.True(t, allow["read_file"])
	assert.True(t, allow["write_file"])
	assert.True(t, allow["direct_tool"])
	assert.False(t, allow[""])

	allow, ok = buildModeAllowlistResolver(nil, catalog)(session.WithModeName(context.Background(), "review"))
	assert.False(t, ok)
	assert.Nil(t, allow)

	allow, ok = buildModeAllowlistResolver(cfg, nil)(session.WithModeName(context.Background(), "review"))
	assert.False(t, ok)
	assert.Nil(t, allow)
}

func TestWave30AppHotspotHelpersCoverMetaModulesWiringAndP2P(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Security.Exportability.Enabled = true
	assert.True(t, exportabilityPolicyEnabled(cfg))
	cfg.Security.Exportability.Enabled = false
	assert.False(t, exportabilityPolicyEnabled(cfg))
	assert.Equal(t, config.DefaultConfig().Security.Exportability.Enabled, exportabilityPolicyEnabled(nil))

	root := t.TempDir()
	assert.True(t, pathInsideDir(filepath.Join(root, "child", "artifact.txt"), root))
	assert.True(t, pathInsideDir(root, root))
	assert.False(t, pathInsideDir(filepath.Join(root, "..", "outside.txt"), root))

	module := &networkModule{cfg: config.DefaultConfig()}
	assert.Equal(t, "network", module.Name())
	assert.Equal(t, []appinit.Provides{
		appinit.ProvidesPayment,
		appinit.ProvidesP2P,
		appinit.ProvidesEconomy,
		appinit.ProvidesContract,
		appinit.ProvidesSmartAccount,
		appinit.ProvidesWorkspace,
	}, module.Provides())
	assert.Equal(t, []appinit.Provides{appinit.ProvidesSecurity, appinit.ProvidesSessionStore}, module.DependsOn())
	assert.False(t, module.Enabled())

	enabledCfg := config.DefaultConfig()
	enabledCfg.P2P.Enabled = true
	assert.True(t, (&networkModule{cfg: enabledCfg}).Enabled())
	assert.NotEmpty(t, buildAgentOptions(enabledCfg, nil))
	assert.Nil(t, buildProvenanceAgentOptions(nil, nil))
	assert.Nil(t, buildProvenanceAgentOptions(&provenanceValues{}, nil))

	wallet := &wiringP2PWallet{signature: []byte("signed"), publicKey: []byte("public")}
	signer := &walletHandshakeSigner{wp: wallet}
	signature, err := signer.SignMessage(context.Background(), []byte("challenge"))
	require.NoError(t, err)
	assert.Equal(t, []byte("signed"), signature)
	publicKey, err := signer.PublicKey(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("public"), publicKey)
	assert.Equal(t, security.AlgorithmSecp256k1Keccak256, signer.Algorithm())

	tools := buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, cfg, nil, nil, nil, nil, nil, nil, nil)
	assert.NotNil(t, findTool(tools, "save_knowledge"))
}

func TestWave30ProvenanceModuleMetadataAndMemoryInit(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Provenance.Enabled = true
	cfg.Hooks.Enabled = true
	boot := &bootstrap.Result{
		Config:       cfg,
		ExplicitKeys: map[string]bool{"agent.provider": true},
		AutoEnabled:  config.AutoEnabledSet{Retrieval: true},
	}

	module := &provenanceModule{cfg: cfg, boot: boot}
	assert.Equal(t, "provenance", module.Name())
	assert.Equal(t, []appinit.Provides{appinit.ProvidesProvenance}, module.Provides())
	assert.Equal(t, []appinit.Provides{appinit.ProvidesRunLedger}, module.DependsOn())
	assert.True(t, module.Enabled())

	result, err := module.Init(context.Background(), staticResolver{})
	require.NoError(t, err)
	require.NotNil(t, result)

	values, ok := result.Values[appinit.ProvidesProvenance].(*provenanceValues)
	require.True(t, ok)
	require.NotNil(t, values.checkpointService)
	require.NotNil(t, values.sessionTree)
	require.NotNil(t, values.attribution)
	require.NotNil(t, values.bundle)
	assert.NotEmpty(t, values.configMetadata["config_fingerprint"])
	assert.JSONEq(t, "[]", values.configMetadata["hook_registry"])

	entry := requireCatalogEntry(t, result.CatalogEntries, "provenance")
	assert.True(t, entry.Enabled)
	assert.Equal(t, "provenance.enabled", entry.ConfigKey)

	metadata := computeConfigMetadata(boot, toolchain.NewHookRegistry())
	assert.NotEmpty(t, metadata["config_fingerprint"])
	assert.JSONEq(t, "[]", metadata["hook_registry"])
	assert.Nil(t, computeConfigMetadata(nil, nil))
}

func TestWave30HookRegistrySnapshotIncludesPreAndPostHooks(t *testing.T) {
	t.Parallel()

	registry := toolchain.NewHookRegistry()
	registry.RegisterPre(wave30Hook{name: "pre", priority: 20})
	registry.RegisterPost(wave30Hook{name: "post", priority: 10})

	var entries []hookEntry
	require.NoError(t, json.Unmarshal([]byte(buildHookRegistrySnapshot(registry)), &entries))

	assert.ElementsMatch(t, []hookEntry{
		{Name: "pre", Priority: 20},
		{Name: "post", Priority: 10},
	}, entries)
	assert.JSONEq(t, "[]", buildHookRegistrySnapshot(nil))
}

type wave30Hook struct {
	name     string
	priority int
}

func (h wave30Hook) Name() string  { return h.name }
func (h wave30Hook) Priority() int { return h.priority }
func (h wave30Hook) Pre(toolchain.HookContext) (toolchain.PreHookResult, error) {
	return toolchain.PreHookResult{Action: toolchain.Continue}, nil
}
func (h wave30Hook) Post(toolchain.HookContext, interface{}, error) error {
	return nil
}
