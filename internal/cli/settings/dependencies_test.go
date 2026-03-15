package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

func TestDependencyIndex_SmartAccountUnmet(t *testing.T) {
	idx := NewDependencyIndex()
	cfg := &config.Config{}

	tests := []struct {
		give  string
		setup func(*config.Config)
		want  int
	}{
		{
			give:  "payment disabled, signer empty",
			setup: func(c *config.Config) {},
			want:  2, // payment + security signer
		},
		{
			give: "payment enabled but no RPC URL",
			setup: func(c *config.Config) {
				c.Payment.Enabled = true
			},
			want: 2, // payment misconfigured + security signer
		},
		{
			give: "payment enabled with RPC, signer empty",
			setup: func(c *config.Config) {
				c.Payment.Enabled = true
				c.Payment.Network.RPCURL = "https://rpc.example.com"
			},
			want: 1, // security signer only
		},
		{
			give: "all configured",
			setup: func(c *config.Config) {
				c.Payment.Enabled = true
				c.Payment.Network.RPCURL = "https://rpc.example.com"
				c.Security.Signer.Provider = "local"
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			testCfg := *cfg // copy
			tt.setup(&testCfg)
			got := idx.UnmetRequired("smartaccount", &testCfg)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDependencyIndex_Evaluate(t *testing.T) {
	idx := NewDependencyIndex()
	cfg := &config.Config{} // all disabled

	results := idx.Evaluate("smartaccount", cfg)
	require.Len(t, results, 3) // payment, security, economy(optional)

	assert.Equal(t, "payment", results[0].CategoryID)
	assert.Equal(t, DepNotEnabled, results[0].Status)
	assert.True(t, results[0].Required)

	assert.Equal(t, "security", results[1].CategoryID)
	assert.Equal(t, DepNotEnabled, results[1].Status)
	assert.True(t, results[1].Required)

	assert.Equal(t, "economy", results[2].CategoryID)
	assert.Equal(t, DepNotEnabled, results[2].Status)
	assert.False(t, results[2].Required)
}

func TestDependencyIndex_TransitiveResolution(t *testing.T) {
	idx := NewDependencyIndex()
	cfg := &config.Config{} // all disabled

	// smartaccount_session → smartaccount → payment + security
	unmet := idx.AllTransitiveUnmet("smartaccount_session", cfg)

	require.True(t, len(unmet) >= 3, "expected at least 3 transitive deps, got %d", len(unmet))

	ids := make([]string, len(unmet))
	for i, d := range unmet {
		ids[i] = d.CategoryID
	}

	paymentIdx := indexOf(ids, "payment")
	securityIdx := indexOf(ids, "security")
	saIdx := indexOf(ids, "smartaccount")

	assert.True(t, paymentIdx >= 0, "payment should be in transitive deps")
	assert.True(t, securityIdx >= 0, "security should be in transitive deps")
	assert.True(t, saIdx >= 0, "smartaccount should be in transitive deps")
	assert.True(t, paymentIdx < saIdx, "payment should come before smartaccount")
	assert.True(t, securityIdx < saIdx, "security should come before smartaccount")
}

func TestDependencyIndex_CycleGuard(t *testing.T) {
	idx := NewDependencyIndex()
	cfg := &config.Config{}

	unmet := idx.AllTransitiveUnmet("p2p_pricing", cfg)
	assert.True(t, len(unmet) >= 2, "p2p_pricing should have at least 2 unmet deps")
}

func TestDependencyIndex_Dependents(t *testing.T) {
	idx := NewDependencyIndex()

	dependents := idx.Dependents("payment")
	assert.True(t, len(dependents) > 0, "payment should have dependents")

	found := false
	for _, id := range dependents {
		if id == "smartaccount" {
			found = true
			break
		}
	}
	assert.True(t, found, "smartaccount should depend on payment")
}

func TestDependencyIndex_NoDepsCategory(t *testing.T) {
	idx := NewDependencyIndex()
	cfg := &config.Config{}

	assert.False(t, idx.HasDependencies("agent"))
	assert.Equal(t, 0, idx.UnmetRequired("agent", cfg))
	assert.Nil(t, idx.Evaluate("agent", cfg))
}

func TestDependencyPanel_NilWhenAllMet(t *testing.T) {
	results := []DepResult{
		{Dependency: Dependency{CategoryID: "payment", Label: "Payment", Required: true}, Status: DepMet},
		{Dependency: Dependency{CategoryID: "security", Label: "Security", Required: true}, Status: DepMet},
	}
	panel := NewDependencyPanel("smartaccount", results)
	assert.Nil(t, panel, "panel should be nil when all deps are met")
}

func TestDependencyPanel_CreatedWhenUnmet(t *testing.T) {
	results := []DepResult{
		{Dependency: Dependency{CategoryID: "payment", Label: "Payment", Required: true}, Status: DepNotEnabled},
		{Dependency: Dependency{CategoryID: "security", Label: "Security", Required: true}, Status: DepMet},
	}
	panel := NewDependencyPanel("smartaccount", results)
	require.NotNil(t, panel)
	assert.Equal(t, "smartaccount", panel.CategoryID)
	assert.Equal(t, 1, panel.UnmetCount())
	assert.Equal(t, "payment", panel.SelectedCategoryID())
}

func TestDependencyPanel_Navigation(t *testing.T) {
	results := []DepResult{
		{Dependency: Dependency{CategoryID: "payment", Label: "Payment"}, Status: DepNotEnabled},
		{Dependency: Dependency{CategoryID: "security", Label: "Security"}, Status: DepNotEnabled},
	}
	panel := NewDependencyPanel("smartaccount", results)
	require.NotNil(t, panel)

	assert.Equal(t, "payment", panel.SelectedCategoryID())

	panel.MoveDown()
	assert.Equal(t, "security", panel.SelectedCategoryID())

	panel.MoveDown()
	assert.Equal(t, "security", panel.SelectedCategoryID())

	panel.MoveUp()
	assert.Equal(t, "payment", panel.SelectedCategoryID())

	panel.MoveUp()
	assert.Equal(t, "payment", panel.SelectedCategoryID())
}

func TestDependencyPanel_View(t *testing.T) {
	results := []DepResult{
		{
			Dependency: Dependency{CategoryID: "payment", Label: "Payment", Required: true, FixHint: "Enable payment"},
			Status:     DepNotEnabled,
		},
	}
	panel := NewDependencyPanel("smartaccount", results)
	require.NotNil(t, panel)

	view := panel.View()
	assert.Contains(t, view, "Prerequisites")
	assert.Contains(t, view, "Payment")
	assert.Contains(t, view, "Enable payment")
}

func TestSetupFlow_Creation(t *testing.T) {
	tests := []struct {
		give      string
		unmetDeps []DepResult
		wantNil   bool
	}{
		{
			give:      "no unmet deps",
			unmetDeps: nil,
			wantNil:   true,
		},
		{
			give: "one unmet dep",
			unmetDeps: []DepResult{
				{Dependency: Dependency{CategoryID: "payment", Label: "Payment"}, Status: DepNotEnabled},
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			state := tuicore.NewConfigStateWith(&config.Config{})
			sf := NewSetupFlow("smartaccount", tt.unmetDeps, state)
			if tt.wantNil {
				assert.Nil(t, sf)
			} else {
				assert.NotNil(t, sf)
			}
		})
	}
}

func TestSetupFlow_StepProgression(t *testing.T) {
	unmet := []DepResult{
		{Dependency: Dependency{CategoryID: "payment", Label: "Payment"}, Status: DepNotEnabled},
		{Dependency: Dependency{CategoryID: "security", Label: "Security"}, Status: DepNotEnabled},
	}

	state := tuicore.NewConfigStateWith(&config.Config{})
	sf := NewSetupFlow("smartaccount", unmet, state)
	require.NotNil(t, sf)
	assert.Equal(t, SetupInProgress, sf.State())
	assert.Equal(t, "smartaccount", sf.TargetID())

	sf.NextStep()
	assert.Equal(t, SetupInProgress, sf.State())

	sf.SkipStep()
	assert.Equal(t, SetupCompleted, sf.State())
}

func TestSetupFlow_Cancel(t *testing.T) {
	unmet := []DepResult{
		{Dependency: Dependency{CategoryID: "payment", Label: "Payment"}, Status: DepNotEnabled},
	}

	state := tuicore.NewConfigStateWith(&config.Config{})
	sf := NewSetupFlow("smartaccount", unmet, state)
	require.NotNil(t, sf)

	sf.Cancel()
	assert.Equal(t, SetupCancelled, sf.State())
}

func TestSetupFlow_View(t *testing.T) {
	unmet := []DepResult{
		{Dependency: Dependency{CategoryID: "payment", Label: "Payment"}, Status: DepNotEnabled},
	}

	state := tuicore.NewConfigStateWith(&config.Config{})
	sf := NewSetupFlow("smartaccount", unmet, state)
	require.NotNil(t, sf)

	view := sf.View()
	assert.Contains(t, view, "Guided Setup")
	assert.Contains(t, view, "Payment")
}

func TestSetupFlow_DeduplicatesDeps(t *testing.T) {
	unmet := []DepResult{
		{Dependency: Dependency{CategoryID: "payment", Label: "Payment"}, Status: DepNotEnabled},
		{Dependency: Dependency{CategoryID: "payment", Label: "Payment"}, Status: DepNotEnabled},
		{Dependency: Dependency{CategoryID: "security", Label: "Security"}, Status: DepNotEnabled},
	}

	state := tuicore.NewConfigStateWith(&config.Config{})
	sf := NewSetupFlow("smartaccount", unmet, state)
	require.NotNil(t, sf)

	sf.NextStep() // payment
	assert.Equal(t, SetupInProgress, sf.State())
	sf.NextStep() // security
	assert.Equal(t, SetupCompleted, sf.State())
}

func TestMenuModel_ReadyFilter(t *testing.T) {
	m := NewMenuModel()
	m.DependencyChecker = func(id string) int {
		if id == "smartaccount" || id == "smartaccount_session" {
			return 2
		}
		return 0
	}

	m.searching = true
	m.searchInput.SetValue("@ready")
	m.applyFilter()

	for _, cat := range m.filtered {
		assert.NotEqual(t, "smartaccount", cat.ID, "blocked category should not appear in @ready filter")
		assert.NotEqual(t, "smartaccount_session", cat.ID, "blocked category should not appear in @ready filter")
	}
	assert.True(t, len(m.filtered) > 0, "should have some ready categories")
}

// Helper

func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}
