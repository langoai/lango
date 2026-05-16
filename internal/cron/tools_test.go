package cron

import (
	"context"
	"testing"
	"time"

	"github.com/langoai/lango/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBuildTools_RequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	runner := &mockAgentRunner{response: "ok"}
	executor := NewExecutor(runner, nil, store, zap.NewNop().Sugar())
	scheduler := New(store, executor, SchedulerConfig{
		Timezone:       "UTC",
		MaxJobs:        5,
		DefaultTimeout: 30 * time.Minute,
		Logger:         zap.NewNop().Sugar(),
	})
	tools := BuildTools(scheduler, nil)

	testCases := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name:    "add requires name",
			tool:    "cron_add",
			params:  map[string]interface{}{"schedule_type": "cron", "schedule": "0 9 * * *", "prompt": "Summarize news"},
			wantErr: "missing name parameter",
		},
		{
			name:    "add requires schedule type",
			tool:    "cron_add",
			params:  map[string]interface{}{"name": "news", "schedule": "0 9 * * *", "prompt": "Summarize news"},
			wantErr: "missing schedule_type parameter",
		},
		{
			name:    "add requires schedule",
			tool:    "cron_add",
			params:  map[string]interface{}{"name": "news", "schedule_type": "cron", "prompt": "Summarize news"},
			wantErr: "missing schedule parameter",
		},
		{
			name:    "add requires prompt",
			tool:    "cron_add",
			params:  map[string]interface{}{"name": "news", "schedule_type": "cron", "schedule": "0 9 * * *"},
			wantErr: "missing prompt parameter",
		},
		{
			name:    "pause requires id",
			tool:    "cron_pause",
			params:  map[string]interface{}{},
			wantErr: "missing id parameter",
		},
		{
			name:    "resume requires id",
			tool:    "cron_resume",
			params:  map[string]interface{}{},
			wantErr: "missing id parameter",
		},
		{
			name:    "remove requires id",
			tool:    "cron_remove",
			params:  map[string]interface{}{},
			wantErr: "missing id parameter",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := findCronTool(t, tools, tc.tool).Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func TestBuildTools_ReadOnlyInspectionCapabilityMetadata(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	runner := &mockAgentRunner{response: "ok"}
	executor := NewExecutor(runner, nil, store, zap.NewNop().Sugar())
	scheduler := New(store, executor, SchedulerConfig{
		Timezone:       "UTC",
		MaxJobs:        5,
		DefaultTimeout: 30 * time.Minute,
		Logger:         zap.NewNop().Sugar(),
	})
	tools := BuildTools(scheduler, nil)

	listTool := findCronTool(t, tools, "cron_list")
	assert.Equal(t, agent.ActivityQuery, listTool.Capability.Activity)
	assert.True(t, listTool.Capability.ReadOnly)

	historyTool := findCronTool(t, tools, "cron_history")
	assert.Equal(t, agent.ActivityQuery, historyTool.Capability.Activity)
	assert.True(t, historyTool.Capability.ReadOnly)
}

func findCronTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("tool %q not found", name)
	return nil
}
