package app

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/agentrt"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/skill"
)

func TestPathInsideDirRejectsSiblingAndParentEscapes(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "workspace")

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "same directory",
			path: root,
			want: true,
		},
		{
			name: "nested file",
			path: filepath.Join(root, "nested", "file.txt"),
			want: true,
		},
		{
			name: "sibling with shared prefix",
			path: root + "-other",
			want: false,
		},
		{
			name: "parent directory",
			path: filepath.Join(root, ".."),
			want: false,
		},
		{
			name: "relative escape",
			path: filepath.Join(root, "..", "outside.txt"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, pathInsideDir(tt.path, root))
		})
	}
}

func TestFoundationModuleInitWrapsSecurityProviderErrors(t *testing.T) {
	t.Parallel()

	cfg := foundationModuleInitBuildsBaseToolsAndDisabledOptionalCatalogModuleConfig(t)
	cfg.Security.Signer.Provider = "unsupported-signer"

	result, err := (&foundationModule{cfg: cfg}).Init(t.Context(), nil)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "security init")
	assert.ErrorContains(t, err, `unsupported security provider "unsupported-signer"`)
}

func TestModuleContractsAndEnabledFlags(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()

	tests := []struct {
		name   string
		module interface {
			Name() string
			Provides() []appinit.Provides
			DependsOn() []appinit.Provides
			Enabled() bool
		}
		wantName     string
		wantProvides []appinit.Provides
		wantDepends  []appinit.Provides
		wantEnabled  bool
	}{
		{
			name:        "foundation",
			module:      &foundationModule{cfg: cfg},
			wantName:    "foundation",
			wantEnabled: true,
			wantProvides: []appinit.Provides{
				appinit.ProvidesSupervisor,
				appinit.ProvidesSessionStore,
				appinit.ProvidesSecurity,
			},
		},
		{
			name:        "intelligence",
			module:      &intelligenceModule{cfg: cfg},
			wantName:    "intelligence",
			wantEnabled: true,
			wantProvides: []appinit.Provides{
				appinit.ProvidesKnowledge,
				appinit.ProvidesMemory,
				appinit.ProvidesGraph,
				appinit.ProvidesLibrarian,
				appinit.ProvidesSkills,
			},
			wantDepends: []appinit.Provides{
				appinit.ProvidesSessionStore,
				appinit.ProvidesSupervisor,
				appinit.ProvidesEconomy,
				appinit.ProvidesAutomation,
			},
		},
		{
			name:        "extension",
			module:      &extensionModule{cfg: cfg},
			wantName:    "extension",
			wantEnabled: true,
			wantProvides: []appinit.Provides{
				appinit.ProvidesMCP,
				appinit.ProvidesObservability,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantName, tt.module.Name())
			assert.Equal(t, tt.wantProvides, tt.module.Provides())
			assert.Equal(t, tt.wantDepends, tt.module.DependsOn())
			assert.Equal(t, tt.wantEnabled, tt.module.Enabled())
		})
	}

	automationOff := &automationModule{cfg: config.DefaultConfig()}
	assert.False(t, automationOff.Enabled())

	automationOnCfg := config.DefaultConfig()
	automationOnCfg.Background.Enabled = true
	assert.True(t, (&automationModule{cfg: automationOnCfg}).Enabled())

	networkOff := &networkModule{cfg: config.DefaultConfig()}
	assert.False(t, networkOff.Enabled())

	networkOnCfg := config.DefaultConfig()
	networkOnCfg.P2P.Enabled = true
	assert.True(t, (&networkModule{cfg: networkOnCfg}).Enabled())
}

func TestResolveHelpersReturnOnlyTypedComponentValues(t *testing.T) {
	t.Parallel()

	assert.Nil(t, resolveKC(nil))
	assert.Nil(t, resolveMC(nil))
	assert.Nil(t, resolveGC(nil))
	assert.Nil(t, resolveLC(nil))
	assert.Nil(t, resolveSR(nil))
	assert.Nil(t, resolveSR(&intelligenceValues{SkillRegistry: "not a registry"}))

	kc := &knowledgeComponents{}
	mc := &memoryComponents{}
	gc := &graphComponents{}
	lc := &librarianComponents{}
	registry := skill.NewRegistry(nil, []*agent.Tool{{Name: "pathInsideDirRejectsSiblingAndParentEscapes_tool"}}, zap.NewNop().Sugar())

	values := &intelligenceValues{
		KC:            kc,
		MC:            mc,
		GC:            gc,
		LC:            lc,
		SkillRegistry: registry,
	}

	assert.Same(t, kc, resolveKC(values))
	assert.Same(t, mc, resolveMC(values))
	assert.Same(t, gc, resolveGC(values))
	assert.Same(t, lc, resolveLC(values))
	assert.Same(t, registry, resolveSR(values))
}

func TestGraphAdmissionPolicyOnlyBuildsInObserveMode(t *testing.T) {
	t.Parallel()

	assert.Nil(t, newGraphAdmissionPolicy(nil, nil))

	offCfg := config.DefaultConfig()
	offCfg.Ontology.Governance.AdmissionMode = config.OntologyAdmissionModeOff
	assert.Nil(t, newGraphAdmissionPolicy(offCfg, nil))

	observeCfg := config.DefaultConfig()
	observeCfg.Ontology.Governance.AdmissionMode = config.OntologyAdmissionModeObserve
	assert.NotNil(t, newGraphAdmissionPolicy(observeCfg, nil))
}

func TestWireLoopAndCollaborationReadersPopulateAvailableAdapters(t *testing.T) {
	t.Parallel()

	application := &App{
		ReceiptStore:               receipts.NewStore(),
		CollaborationRuntimeReader: newCollaborationRuntimeBridge(eventbus.New()),
		AgentRunStore:              agentrt.NewInMemoryAgentRunStore(),
	}

	wireLoopReaders(application, nil)
	wireCollaborationReaders(application)

	assert.NotNil(t, application.LoopDeadLetterReader)
	assert.NotNil(t, application.CollaborationAgentRunReader)
	assert.Nil(t, application.CollaborationMissionLinkReader)
	assert.Nil(t, application.CollaborationDelegationReader)

	require.NotPanics(t, func() {
		wireLoopReaders(nil, nil)
		wireCollaborationReaders(nil)
	})
}
