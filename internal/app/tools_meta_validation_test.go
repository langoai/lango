package app

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/knowledge"
)

func newAppKnowledgeStore(t *testing.T) *knowledge.Store {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	return knowledge.NewStore(client, zap.NewNop().Sugar())
}

func TestSaveKnowledge_SourceClassValidationAndDefaults(t *testing.T) {
	store := newAppKnowledgeStore(t)
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, nil, nil)
	tool := findTool(tools, "save_knowledge")
	require.NotNil(t, tool)

	_, err := tool.Handler(context.Background(), map[string]interface{}{
		"key":          "app-validation",
		"category":     "fact",
		"content":      "validation should reject this",
		"source_class": "top-secret",
	})
	require.Error(t, err)

	res, err := tool.Handler(context.Background(), map[string]interface{}{
		"key":      "app-defaults",
		"category": "fact",
		"content":  "default source class should be applied",
	})
	require.NoError(t, err)
	assert.Equal(t, "saved", res.(map[string]interface{})["status"])

	got, err := store.GetKnowledge(context.Background(), "app-defaults")
	require.NoError(t, err)
	assert.Equal(t, "private-confidential", got.SourceClass)
	assert.Equal(t, "app-defaults", got.AssetLabel)
}

func TestSaveKnowledge_RequiresCanonicalInputs(t *testing.T) {
	store := newAppKnowledgeStore(t)
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, nil, nil)
	tool := findTool(tools, "save_knowledge")
	require.NotNil(t, tool)

	cases := []struct {
		name      string
		params    map[string]interface{}
		wantError string
	}{
		{
			name: "missing key",
			params: map[string]interface{}{
				"category": "fact",
				"content":  "knowledge content",
			},
			wantError: "missing key parameter",
		},
		{
			name: "missing category",
			params: map[string]interface{}{
				"key":     "app-missing-category",
				"content": "knowledge content",
			},
			wantError: "missing category parameter",
		},
		{
			name: "missing content",
			params: map[string]interface{}{
				"key":      "app-missing-content",
				"category": "fact",
			},
			wantError: "missing content parameter",
		},
		{
			name: "invalid category",
			params: map[string]interface{}{
				"key":      "app-invalid-category",
				"category": "bogus",
				"content":  "knowledge content",
			},
			wantError: `invalid category "bogus"`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tool.Handler(context.Background(), tt.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantError)
		})
	}
}

func TestKnowledgeHistoryAndSearch_RequireCanonicalInputs(t *testing.T) {
	store := newAppKnowledgeStore(t)
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, nil, nil)
	historyTool := findTool(tools, "get_knowledge_history")
	searchTool := findTool(tools, "search_knowledge")
	require.NotNil(t, historyTool)
	require.NotNil(t, searchTool)

	historyResult, err := historyTool.Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, historyResult)
	assert.ErrorContains(t, err, "missing key parameter")

	searchResult, err := searchTool.Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, searchResult)
	assert.ErrorContains(t, err, "missing query parameter")
}
