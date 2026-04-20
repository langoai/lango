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
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, nil)
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
