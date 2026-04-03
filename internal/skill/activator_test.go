package skill

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/agent"
)

func newTestActivator(t *testing.T) (*Activator, *Registry) {
	t.Helper()

	dir := filepath.Join(t.TempDir(), "skills")
	logger := zap.NewNop().Sugar()
	store := NewFileSkillStore(dir, logger)
	baseTool := &agent.Tool{Name: "base_tool", Description: "base"}
	registry := NewRegistry(store, []*agent.Tool{baseTool}, logger)
	return NewActivator(registry), registry
}

func activateSkill(t *testing.T, registry *Registry, entry SkillEntry) {
	t.Helper()

	ctx := context.Background()
	require.NoError(t, registry.CreateSkill(ctx, entry))
	require.NoError(t, registry.ActivateSkill(ctx, entry.Name))
}

func TestActivator_CheckPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give        string
		skills      []SkillEntry
		editedPaths []string
		wantNames   []string
	}{
		{
			give:        "no edited paths returns nil",
			skills:      nil,
			editedPaths: nil,
			wantNames:   nil,
		},
		{
			give:        "empty edited paths returns nil",
			skills:      nil,
			editedPaths: []string{},
			wantNames:   nil,
		},
		{
			give: "no skills with paths returns nil",
			skills: []SkillEntry{
				{
					Name:       "no-paths",
					Type:       "template",
					Definition: map[string]interface{}{"template": "hi"},
				},
			},
			editedPaths: []string{"foo.go"},
			wantNames:   nil,
		},
		{
			give: "exact filename match",
			skills: []SkillEntry{
				{
					Name:       "go-skill",
					Type:       "template",
					Definition: map[string]interface{}{"template": "go"},
					Paths:      []string{"main.go"},
				},
			},
			editedPaths: []string{"main.go"},
			wantNames:   []string{"go-skill"},
		},
		{
			give: "glob star match",
			skills: []SkillEntry{
				{
					Name:       "go-files",
					Type:       "template",
					Definition: map[string]interface{}{"template": "go"},
					Paths:      []string{"*.go"},
				},
			},
			editedPaths: []string{"handler.go"},
			wantNames:   []string{"go-files"},
		},
		{
			give: "directory glob match",
			skills: []SkillEntry{
				{
					Name:       "cmd-skill",
					Type:       "template",
					Definition: map[string]interface{}{"template": "cmd"},
					Paths:      []string{"cmd/*.go"},
				},
			},
			editedPaths: []string{"cmd/main.go"},
			wantNames:   []string{"cmd-skill"},
		},
		{
			give: "no match",
			skills: []SkillEntry{
				{
					Name:       "py-skill",
					Type:       "template",
					Definition: map[string]interface{}{"template": "py"},
					Paths:      []string{"*.py"},
				},
			},
			editedPaths: []string{"handler.go"},
			wantNames:   nil,
		},
		{
			give: "multiple skills some match",
			skills: []SkillEntry{
				{
					Name:       "go-skill",
					Type:       "template",
					Definition: map[string]interface{}{"template": "go"},
					Paths:      []string{"*.go"},
				},
				{
					Name:       "py-skill",
					Type:       "template",
					Definition: map[string]interface{}{"template": "py"},
					Paths:      []string{"*.py"},
				},
				{
					Name:       "all-skill",
					Type:       "template",
					Definition: map[string]interface{}{"template": "all"},
					Paths:      []string{"*"},
				},
			},
			editedPaths: []string{"handler.go"},
			wantNames:   []string{"go-skill", "all-skill"},
		},
		{
			give: "multiple edited paths any match triggers",
			skills: []SkillEntry{
				{
					Name:       "test-skill",
					Type:       "template",
					Definition: map[string]interface{}{"template": "test"},
					Paths:      []string{"*_test.go"},
				},
			},
			editedPaths: []string{"main.go", "handler_test.go"},
			wantNames:   []string{"test-skill"},
		},
		{
			give: "multiple globs in one skill",
			skills: []SkillEntry{
				{
					Name:       "multi-glob",
					Type:       "template",
					Definition: map[string]interface{}{"template": "multi"},
					Paths:      []string{"*.go", "*.py"},
				},
			},
			editedPaths: []string{"script.py"},
			wantNames:   []string{"multi-glob"},
		},
		{
			give: "skill with empty paths is skipped",
			skills: []SkillEntry{
				{
					Name:       "empty-paths",
					Type:       "template",
					Definition: map[string]interface{}{"template": "x"},
					Paths:      []string{},
				},
				{
					Name:       "has-paths",
					Type:       "template",
					Definition: map[string]interface{}{"template": "y"},
					Paths:      []string{"*.go"},
				},
			},
			editedPaths: []string{"foo.go"},
			wantNames:   []string{"has-paths"},
		},
		{
			give: "malformed glob is skipped gracefully",
			skills: []SkillEntry{
				{
					Name:       "bad-glob",
					Type:       "template",
					Definition: map[string]interface{}{"template": "bad"},
					Paths:      []string{"[invalid"},
				},
			},
			editedPaths: []string{"foo.go"},
			wantNames:   nil,
		},
		{
			give: "character class glob",
			skills: []SkillEntry{
				{
					Name:       "char-class",
					Type:       "template",
					Definition: map[string]interface{}{"template": "cc"},
					Paths:      []string{"*.[tj]s"},
				},
			},
			editedPaths: []string{"index.ts"},
			wantNames:   []string{"char-class"},
		},
		{
			give: "question mark glob",
			skills: []SkillEntry{
				{
					Name:       "qmark",
					Type:       "template",
					Definition: map[string]interface{}{"template": "q"},
					Paths:      []string{"?.go"},
				},
			},
			editedPaths: []string{"a.go"},
			wantNames:   []string{"qmark"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			activator, registry := newTestActivator(t)
			ctx := context.Background()

			for _, sk := range tt.skills {
				activateSkill(t, registry, sk)
			}

			got, err := activator.CheckPaths(ctx, tt.editedPaths)
			require.NoError(t, err)

			if tt.wantNames == nil {
				assert.Empty(t, got)
				return
			}

			gotNames := make([]string, 0, len(got))
			for _, s := range got {
				gotNames = append(gotNames, s.Name)
			}
			assert.ElementsMatch(t, tt.wantNames, gotNames)
		})
	}
}

func TestActivator_CheckPaths_NoDuplicates(t *testing.T) {
	t.Parallel()

	activator, registry := newTestActivator(t)
	ctx := context.Background()

	// A skill with multiple globs that could match the same edited path.
	activateSkill(t, registry, SkillEntry{
		Name:       "multi-match",
		Type:       "template",
		Definition: map[string]interface{}{"template": "mm"},
		Paths:      []string{"*.go", "main.*"},
	})

	got, err := activator.CheckPaths(ctx, []string{"main.go"})
	require.NoError(t, err)

	// The skill should appear exactly once despite matching two globs.
	require.Len(t, got, 1)
	assert.Equal(t, "multi-match", got[0].Name)
}
