package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/types"
)

func TestGraphFormWiresFieldsAndValidators(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.Backend = "bolt"
	cfg.Graph.DatabasePath = "~/.lango/custom-graph.db"
	cfg.Graph.MaxTraversalDepth = 4
	cfg.Graph.MaxExpansionResults = 25

	form := NewGraphForm(cfg)

	assert.Equal(t, "Graph Store Configuration", form.Title)
	assert.Equal(t, []string{
		"graph_enabled",
		"graph_backend",
		"graph_db_path",
		"graph_max_depth",
		"graph_max_expand",
	}, formKeys(form))

	assert.True(t, fieldByKey(form, "graph_enabled").Checked)
	assert.Equal(t, tuicore.InputBool, fieldByKey(form, "graph_enabled").Type)

	backend := fieldByKey(form, "graph_backend")
	require.NotNil(t, backend)
	assert.Equal(t, tuicore.InputSelect, backend.Type)
	assert.Equal(t, "bolt", backend.Value)
	assert.Equal(t, []string{"bolt"}, backend.Options)

	assert.Equal(t, "~/.lango/custom-graph.db", fieldByKey(form, "graph_db_path").Value)
	assert.Equal(t, "~/.lango/graph.db", fieldByKey(form, "graph_db_path").Placeholder)

	maxDepth := fieldByKey(form, "graph_max_depth")
	require.NotNil(t, maxDepth)
	assert.Equal(t, "4", maxDepth.Value)
	require.NoError(t, maxDepth.Validate("1"))
	assert.EqualError(t, maxDepth.Validate("0"), "must be a positive integer")
	assert.EqualError(t, maxDepth.Validate("not-a-number"), "must be a positive integer")

	maxExpand := fieldByKey(form, "graph_max_expand")
	require.NotNil(t, maxExpand)
	assert.Equal(t, "25", maxExpand.Value)
	require.NoError(t, maxExpand.Validate("100"))
	assert.EqualError(t, maxExpand.Validate("-1"), "must be a positive integer")
}

func TestLibrarianFormWiresProviderModelAndValidators(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "primary"
	cfg.Agent.Model = "agent-model"
	cfg.Providers = map[string]config.ProviderConfig{
		"primary": {Type: types.ProviderOpenAI},
		"worker":  {Type: types.ProviderAnthropic},
	}
	cfg.Librarian.Enabled = true
	cfg.Librarian.ObservationThreshold = 6
	cfg.Librarian.InquiryCooldownTurns = 4
	cfg.Librarian.MaxPendingInquiries = 3
	cfg.Librarian.AutoSaveConfidence = types.ConfidenceMedium
	cfg.Librarian.Provider = ""
	cfg.Librarian.Model = "librarian-model"

	form := NewLibrarianForm(cfg)

	assert.Equal(t, "Librarian Configuration", form.Title)
	assert.Equal(t, []string{
		"lib_enabled",
		"lib_obs_threshold",
		"lib_cooldown",
		"lib_max_inquiries",
		"lib_auto_save",
		"lib_provider",
		"lib_model",
	}, formKeys(form))

	assert.True(t, fieldByKey(form, "lib_enabled").Checked)

	obsThreshold := fieldByKey(form, "lib_obs_threshold")
	require.NotNil(t, obsThreshold)
	assert.Equal(t, "6", obsThreshold.Value)
	require.NoError(t, obsThreshold.Validate("1"))
	assert.EqualError(t, obsThreshold.Validate("0"), "must be a positive integer")

	cooldown := fieldByKey(form, "lib_cooldown")
	require.NotNil(t, cooldown)
	assert.Equal(t, "4", cooldown.Value)
	require.NoError(t, cooldown.Validate("0"))
	assert.EqualError(t, cooldown.Validate("-1"), "must be a non-negative integer")

	maxInquiries := fieldByKey(form, "lib_max_inquiries")
	require.NotNil(t, maxInquiries)
	assert.Equal(t, "3", maxInquiries.Value)
	require.NoError(t, maxInquiries.Validate("0"))
	assert.EqualError(t, maxInquiries.Validate("abc"), "must be a non-negative integer")

	autoSave := fieldByKey(form, "lib_auto_save")
	require.NotNil(t, autoSave)
	assert.Equal(t, tuicore.InputSelect, autoSave.Type)
	assert.Equal(t, "medium", autoSave.Value)
	assert.Equal(t, []string{"high", "medium", "low"}, autoSave.Options)

	provider := fieldByKey(form, "lib_provider")
	require.NotNil(t, provider)
	assert.Equal(t, tuicore.InputSelect, provider.Type)
	assert.Equal(t, "", provider.Value)
	assert.Equal(t, []string{"", "primary", "worker"}, provider.Options)
	assert.Equal(t, "(inherits from Agent)", provider.Placeholder)
	assert.Contains(t, provider.Description, "primary")

	model := fieldByKey(form, "lib_model")
	require.NotNil(t, model)
	assert.Equal(t, tuicore.InputText, model.Type)
	assert.Equal(t, "librarian-model", model.Value)
	assert.Contains(t, model.Description, "Could not fetch models")

	cmd := provider.OnChange("")
	require.NotNil(t, cmd)
	assert.True(t, model.Loading)
	msg, ok := cmd().(tuicore.FieldOptionsLoadedMsg)
	require.True(t, ok)
	assert.Equal(t, "lib_model", msg.FieldKey)
	assert.Equal(t, "primary", msg.ProviderID)
	require.Error(t, msg.Err)
	assert.Contains(t, msg.Err.Error(), "missing API key")
}

func TestObservationalMemoryFormWiresLimitsAndProviderReload(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "primary"
	cfg.Providers = map[string]config.ProviderConfig{
		"primary": {Type: types.ProviderOpenAI},
	}
	cfg.ObservationalMemory.Enabled = true
	cfg.ObservationalMemory.MessageTokenThreshold = 1200
	cfg.ObservationalMemory.ObservationTokenThreshold = 2400
	cfg.ObservationalMemory.MaxMessageTokenBudget = 9600
	cfg.ObservationalMemory.MaxReflectionsInContext = 0
	cfg.ObservationalMemory.MaxObservationsInContext = 12

	form := NewObservationalMemoryForm(cfg)

	assert.Equal(t, "Observational Memory", form.Title)
	assert.True(t, fieldByKey(form, "om_enabled").Checked)
	assert.Equal(t, "1200", fieldByKey(form, "om_msg_threshold").Value)
	assert.Equal(t, "2400", fieldByKey(form, "om_obs_threshold").Value)
	assert.Equal(t, "9600", fieldByKey(form, "om_max_budget").Value)

	reflections := fieldByKey(form, "om_max_reflections")
	require.NotNil(t, reflections)
	assert.Equal(t, "0", reflections.Value)
	require.NoError(t, reflections.Validate("0"))
	assert.EqualError(t, reflections.Validate("-1"), "must be a non-negative integer (0 = unlimited)")

	observations := fieldByKey(form, "om_max_observations")
	require.NotNil(t, observations)
	assert.Equal(t, "12", observations.Value)
	require.NoError(t, observations.Validate("8"))
	assert.EqualError(t, observations.Validate("bad"), "must be a non-negative integer (0 = unlimited)")

	provider := fieldByKey(form, "om_provider")
	model := fieldByKey(form, "om_model")
	require.NotNil(t, provider)
	require.NotNil(t, model)

	cmd := provider.OnChange("")
	require.NotNil(t, cmd)
	assert.True(t, model.Loading)
	msg, ok := cmd().(tuicore.FieldOptionsLoadedMsg)
	require.True(t, ok)
	assert.Equal(t, "om_model", msg.FieldKey)
	assert.Equal(t, "primary", msg.ProviderID)
	require.Error(t, msg.Err)
}

func TestEmbeddingFormWiresProviderOptionsValidationAndReload(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"alpha": {Type: types.ProviderOpenAI},
		"zeta":  {Type: types.ProviderAnthropic},
	}
	cfg.Embedding.Provider = "alpha"
	cfg.Embedding.Model = "text-embedding-3-large"
	cfg.Embedding.Dimensions = 3072
	cfg.Embedding.Local.BaseURL = "http://localhost:9999/v1"
	cfg.Embedding.RAG.Enabled = true
	cfg.Embedding.RAG.MaxResults = 11
	cfg.Embedding.RAG.Collections = []string{"docs", "tickets"}

	form := NewEmbeddingForm(cfg)

	assert.Equal(t, "Embedding & RAG Configuration", form.Title)

	provider := fieldByKey(form, "emb_provider_id")
	require.NotNil(t, provider)
	assert.Equal(t, tuicore.InputSelect, provider.Type)
	assert.Equal(t, "alpha", provider.Value)
	assert.Equal(t, []string{"alpha", "local", "zeta"}, provider.Options)

	model := fieldByKey(form, "emb_model")
	require.NotNil(t, model)
	assert.Equal(t, tuicore.InputText, model.Type)
	assert.Equal(t, "text-embedding-3-large", model.Value)
	assert.Contains(t, model.Description, "Could not fetch models")

	dimensions := fieldByKey(form, "emb_dimensions")
	require.NotNil(t, dimensions)
	assert.Equal(t, "3072", dimensions.Value)
	require.NoError(t, dimensions.Validate("0"))
	assert.EqualError(t, dimensions.Validate("-1"), "must be a non-negative integer")

	assert.Equal(t, "http://localhost:9999/v1", fieldByKey(form, "emb_local_baseurl").Value)
	assert.True(t, fieldByKey(form, "emb_rag_enabled").Checked)

	maxResults := fieldByKey(form, "emb_rag_max_results")
	require.NotNil(t, maxResults)
	assert.Equal(t, "11", maxResults.Value)
	require.NoError(t, maxResults.Validate("1"))
	assert.EqualError(t, maxResults.Validate("bad"), "must be a non-negative integer")

	assert.Equal(t, "docs,tickets", fieldByKey(form, "emb_rag_collections").Value)
	assert.Nil(t, provider.OnChange(""))
	assert.Nil(t, provider.OnChange("local"))

	cmd := provider.OnChange("alpha")
	require.NotNil(t, cmd)
	assert.True(t, model.Loading)
	msg, ok := cmd().(tuicore.FieldOptionsLoadedMsg)
	require.True(t, ok)
	assert.Equal(t, "emb_model", msg.FieldKey)
	assert.Equal(t, "alpha", msg.ProviderID)
	require.Error(t, msg.Err)
	assert.Contains(t, msg.Err.Error(), "missing API key")
}

func TestAutoAdjustFormWiresVisibilityAndRatioValidators(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Retrieval.AutoAdjust.Enabled = false
	cfg.Retrieval.AutoAdjust.Mode = "active"
	cfg.Retrieval.AutoAdjust.BoostDelta = 0.15
	cfg.Retrieval.AutoAdjust.DecayDelta = 0.04
	cfg.Retrieval.AutoAdjust.DecayInterval = 17
	cfg.Retrieval.AutoAdjust.MinScore = 0.2
	cfg.Retrieval.AutoAdjust.MaxScore = 0.9
	cfg.Retrieval.AutoAdjust.WarmupTurns = 8

	form := NewAutoAdjustForm(cfg)

	enabled := fieldByKey(form, "aa_enabled")
	require.NotNil(t, enabled)
	assert.False(t, enabled.Checked)

	mode := fieldByKey(form, "aa_mode")
	require.NotNil(t, mode)
	assert.Equal(t, tuicore.InputSelect, mode.Type)
	assert.Equal(t, "active", mode.Value)
	assert.Equal(t, []string{"shadow", "active"}, mode.Options)
	require.NotNil(t, mode.VisibleWhen)
	assert.False(t, mode.VisibleWhen())

	enabled.Checked = true
	assert.True(t, mode.VisibleWhen())

	boost := fieldByKey(form, "aa_boost_delta")
	require.NotNil(t, boost)
	assert.Equal(t, "0.15", boost.Value)
	require.NoError(t, boost.Validate("1.0"))
	assert.EqualError(t, boost.Validate("1.1"), "must be a float between 0.0 and 1.0")
	assert.EqualError(t, boost.Validate("not-float"), "must be a float between 0.0 and 1.0")

	assert.Equal(t, "0.04", fieldByKey(form, "aa_decay_delta").Value)
	assert.Equal(t, "17", fieldByKey(form, "aa_decay_interval").Value)
	assert.Equal(t, "0.2", fieldByKey(form, "aa_min_score").Value)
	assert.Equal(t, "0.9", fieldByKey(form, "aa_max_score").Value)
	assert.Equal(t, "8", fieldByKey(form, "aa_warmup_turns").Value)
}

func TestContextBudgetFormFormatsAllocationsAndValidatesRatios(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Context.ModelWindow = 200000
	cfg.Context.ResponseReserve = 8192
	cfg.Context.Allocation.Knowledge = 0.10
	cfg.Context.Allocation.RAG = 0.20
	cfg.Context.Allocation.Memory = 0.30
	cfg.Context.Allocation.RunSummary = 0.15
	cfg.Context.Allocation.Headroom = 0.25

	form := NewContextBudgetForm(cfg)

	assert.Equal(t, "Context Budget Configuration", form.Title)
	assert.Equal(t, "200000", fieldByKey(form, "ctx_model_window").Value)
	assert.Equal(t, "8192", fieldByKey(form, "ctx_response_reserve").Value)

	tests := map[string]string{
		"ctx_alloc_knowledge":   "0.10",
		"ctx_alloc_rag":         "0.20",
		"ctx_alloc_memory":      "0.30",
		"ctx_alloc_run_summary": "0.15",
		"ctx_alloc_headroom":    "0.25",
	}
	for key, want := range tests {
		field := fieldByKey(form, key)
		require.NotNil(t, field, key)
		assert.Equal(t, want, field.Value)
		require.NoError(t, field.Validate("0.5"))
		assert.EqualError(t, field.Validate("-0.1"), "must be a float between 0.0 and 1.0")
		assert.EqualError(t, field.Validate("NaN-ish"), "must be a float between 0.0 and 1.0")
	}
}

func formKeys(form *tuicore.FormModel) []string {
	keys := make([]string, 0, len(form.Fields))
	for _, field := range form.Fields {
		if field == nil {
			keys = append(keys, "<nil>")
			continue
		}
		keys = append(keys, field.Key)
	}
	return keys
}
