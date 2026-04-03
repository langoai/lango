package agentrt

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
)

// --- TaskStore unit tests ---

func TestInMemoryTaskStore_CreateAndGet(t *testing.T) {
	store := NewInMemoryTaskStore()

	now := time.Now()
	entry := &TaskEntry{
		ID:          "task-001",
		Title:       "write tests",
		Status:      "todo",
		Description: "table-driven tests",
		ParentID:    "parent-1",
		AgentID:     "agent-x",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	require.NoError(t, store.Create(entry))

	got, err := store.Get("task-001")
	require.NoError(t, err)
	assert.Equal(t, "task-001", got.ID)
	assert.Equal(t, "write tests", got.Title)
	assert.Equal(t, "todo", got.Status)
	assert.Equal(t, "table-driven tests", got.Description)
	assert.Equal(t, "parent-1", got.ParentID)
	assert.Equal(t, "agent-x", got.AgentID)
}

func TestInMemoryTaskStore_CreateDuplicateID(t *testing.T) {
	store := NewInMemoryTaskStore()

	entry := &TaskEntry{ID: "dup-1", Title: "first", Status: "todo"}
	require.NoError(t, store.Create(entry))

	err := store.Create(&TaskEntry{ID: "dup-1", Title: "second", Status: "todo"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestInMemoryTaskStore_CreateNil(t *testing.T) {
	store := NewInMemoryTaskStore()
	err := store.Create(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil entry")
}

func TestInMemoryTaskStore_GetNotFound(t *testing.T) {
	store := NewInMemoryTaskStore()
	_, err := store.Get("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestInMemoryTaskStore_GetReturnsCopy(t *testing.T) {
	store := NewInMemoryTaskStore()
	entry := &TaskEntry{ID: "copy-1", Title: "original", Status: "todo"}
	require.NoError(t, store.Create(entry))

	got, err := store.Get("copy-1")
	require.NoError(t, err)

	// Mutate the returned copy.
	got.Title = "mutated"
	got.Status = "done"

	// Verify the internal state is unaffected.
	original, err := store.Get("copy-1")
	require.NoError(t, err)
	assert.Equal(t, "original", original.Title)
	assert.Equal(t, "todo", original.Status)
}

func TestInMemoryTaskStore_List(t *testing.T) {
	store := NewInMemoryTaskStore()

	// Empty store returns empty slice.
	assert.Empty(t, store.List("", ""))

	require.NoError(t, store.Create(&TaskEntry{ID: "a", Title: "task a", Status: "todo", ParentID: "p1"}))
	require.NoError(t, store.Create(&TaskEntry{ID: "b", Title: "task b", Status: "in_progress", ParentID: "p1"}))
	require.NoError(t, store.Create(&TaskEntry{ID: "c", Title: "task c", Status: "done", ParentID: "p2"}))
	require.NoError(t, store.Create(&TaskEntry{ID: "d", Title: "task d", Status: "blocked", ParentID: "p2"}))

	tests := []struct {
		give         string
		giveStatus   string
		giveParent   string
		wantCount    int
		wantContains []string
	}{
		{
			give:      "no filter",
			wantCount: 4,
		},
		{
			give:         "filter by status todo",
			giveStatus:   "todo",
			wantCount:    1,
			wantContains: []string{"a"},
		},
		{
			give:         "filter by status in_progress",
			giveStatus:   "in_progress",
			wantCount:    1,
			wantContains: []string{"b"},
		},
		{
			give:         "filter by parent p1",
			giveParent:   "p1",
			wantCount:    2,
			wantContains: []string{"a", "b"},
		},
		{
			give:         "filter by parent p2",
			giveParent:   "p2",
			wantCount:    2,
			wantContains: []string{"c", "d"},
		},
		{
			give:         "filter by status and parent",
			giveStatus:   "done",
			giveParent:   "p2",
			wantCount:    1,
			wantContains: []string{"c"},
		},
		{
			give:       "no match",
			giveStatus: "todo",
			giveParent: "p2",
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result := store.List(tt.giveStatus, tt.giveParent)
			assert.Len(t, result, tt.wantCount)

			if len(tt.wantContains) > 0 {
				ids := make(map[string]bool, len(result))
				for _, r := range result {
					ids[r.ID] = true
				}
				for _, id := range tt.wantContains {
					assert.True(t, ids[id], "expected ID %q in results", id)
				}
			}
		})
	}
}

func TestInMemoryTaskStore_ListReturnsCopies(t *testing.T) {
	store := NewInMemoryTaskStore()
	require.NoError(t, store.Create(&TaskEntry{ID: "list-cp", Title: "original", Status: "todo"}))

	tasks := store.List("", "")
	require.Len(t, tasks, 1)
	tasks[0].Status = "done"

	got, err := store.Get("list-cp")
	require.NoError(t, err)
	assert.Equal(t, "todo", got.Status)
}

func TestInMemoryTaskStore_Update(t *testing.T) {
	tests := []struct {
		give            string
		giveStatus      string
		giveDescription string
		wantStatus      string
		wantDescription string
	}{
		{
			give:       "update status only",
			giveStatus: "in_progress",
			wantStatus: "in_progress",
		},
		{
			give:            "update description only",
			giveDescription: "updated desc",
			wantDescription: "updated desc",
		},
		{
			give:            "update both",
			giveStatus:      "done",
			giveDescription: "completed",
			wantStatus:      "done",
			wantDescription: "completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			store := NewInMemoryTaskStore()
			require.NoError(t, store.Create(&TaskEntry{
				ID:          "u-" + tt.give,
				Title:       "test",
				Status:      "todo",
				Description: "initial",
			}))

			err := store.Update("u-"+tt.give, tt.giveStatus, tt.giveDescription)
			require.NoError(t, err)

			got, err := store.Get("u-" + tt.give)
			require.NoError(t, err)

			if tt.wantStatus != "" {
				assert.Equal(t, tt.wantStatus, got.Status)
			}
			if tt.wantDescription != "" {
				assert.Equal(t, tt.wantDescription, got.Description)
			}
			assert.False(t, got.UpdatedAt.IsZero(), "UpdatedAt should be set")
		})
	}
}

func TestInMemoryTaskStore_UpdateNotFound(t *testing.T) {
	store := NewInMemoryTaskStore()
	err := store.Update("ghost", "done", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGenerateTaskID(t *testing.T) {
	id, err := generateTaskID()
	require.NoError(t, err)
	assert.Contains(t, id, "task-")
	// "task-" (5 chars) + 16 hex chars = 21 total.
	assert.Len(t, id, 21)

	// IDs should be unique.
	id2, err := generateTaskID()
	require.NoError(t, err)
	assert.NotEqual(t, id, id2)
}

// --- Tool handler tests ---

func TestTaskCreateTool(t *testing.T) {
	store := NewInMemoryTaskStore()
	tools := BuildTaskTools(store)

	var createTool *taskToolHelper
	for _, tool := range tools {
		if tool.Name == "task_create" {
			createTool = &taskToolHelper{tool}
			break
		}
	}
	require.NotNil(t, createTool, "task_create tool not found")

	tests := []struct {
		give       string
		giveParams map[string]interface{}
		wantErr    bool
		wantStatus string
	}{
		{
			give: "create with title only",
			giveParams: map[string]interface{}{
				"title": "my task",
			},
			wantStatus: "todo",
		},
		{
			give: "create with all fields",
			giveParams: map[string]interface{}{
				"title":       "full task",
				"description": "a detailed task",
				"parent_id":   "parent-1",
			},
			wantStatus: "todo",
		},
		{
			give:       "missing title",
			giveParams: map[string]interface{}{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result, err := createTool.call(tt.giveParams)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			m := result.(map[string]interface{})
			assert.NotEmpty(t, m["task_id"])
			assert.Equal(t, tt.wantStatus, m["status"])
		})
	}
}

func TestTaskGetTool(t *testing.T) {
	store := NewInMemoryTaskStore()
	now := time.Now()
	require.NoError(t, store.Create(&TaskEntry{
		ID:          "get-1",
		Title:       "fetch me",
		Status:      "in_progress",
		Description: "test desc",
		ParentID:    "p-1",
		AgentID:     "ag-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}))

	tools := BuildTaskTools(store)
	getTool := findTool(t, tools, "task_get")

	tests := []struct {
		give       string
		giveParams map[string]interface{}
		wantErr    bool
		wantTitle  string
	}{
		{
			give:       "existing task",
			giveParams: map[string]interface{}{"task_id": "get-1"},
			wantTitle:  "fetch me",
		},
		{
			give:       "not found",
			giveParams: map[string]interface{}{"task_id": "nonexistent"},
			wantErr:    true,
		},
		{
			give:       "missing task_id",
			giveParams: map[string]interface{}{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result, err := getTool.call(tt.giveParams)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			m := result.(map[string]interface{})
			assert.Equal(t, tt.wantTitle, m["title"])
			assert.Equal(t, "in_progress", m["status"])
			assert.Equal(t, "ag-1", m["agent_id"])
			assert.Equal(t, "p-1", m["parent_id"])
		})
	}
}

func TestTaskListTool(t *testing.T) {
	store := NewInMemoryTaskStore()
	require.NoError(t, store.Create(&TaskEntry{ID: "l1", Title: "t1", Status: "todo", ParentID: "p1"}))
	require.NoError(t, store.Create(&TaskEntry{ID: "l2", Title: "t2", Status: "done", ParentID: "p1"}))
	require.NoError(t, store.Create(&TaskEntry{ID: "l3", Title: "t3", Status: "todo", ParentID: "p2"}))

	tools := BuildTaskTools(store)
	listTool := findTool(t, tools, "task_list")

	tests := []struct {
		give       string
		giveParams map[string]interface{}
		wantCount  int
	}{
		{
			give:       "no filter",
			giveParams: map[string]interface{}{},
			wantCount:  3,
		},
		{
			give:       "filter by status todo",
			giveParams: map[string]interface{}{"status": "todo"},
			wantCount:  2,
		},
		{
			give:       "filter by parent p1",
			giveParams: map[string]interface{}{"parent_id": "p1"},
			wantCount:  2,
		},
		{
			give:       "filter by both",
			giveParams: map[string]interface{}{"status": "todo", "parent_id": "p1"},
			wantCount:  1,
		},
		{
			give:       "no match",
			giveParams: map[string]interface{}{"status": "blocked"},
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result, err := listTool.call(tt.giveParams)
			require.NoError(t, err)

			m := result.(map[string]interface{})
			// count comes back as int from our handler.
			assert.Equal(t, tt.wantCount, m["count"])
			tasks := m["tasks"].([]map[string]interface{})
			assert.Len(t, tasks, tt.wantCount)
		})
	}
}

func TestTaskUpdateTool(t *testing.T) {
	store := NewInMemoryTaskStore()
	require.NoError(t, store.Create(&TaskEntry{
		ID:          "upd-1",
		Title:       "updatable",
		Status:      "todo",
		Description: "initial",
	}))

	tools := BuildTaskTools(store)
	updateTool := findTool(t, tools, "task_update")

	tests := []struct {
		give            string
		giveParams      map[string]interface{}
		wantErr         bool
		wantStatus      string
		wantDescription string
	}{
		{
			give:       "update status",
			giveParams: map[string]interface{}{"task_id": "upd-1", "status": "in_progress"},
			wantStatus: "in_progress",
		},
		{
			give:            "update description",
			giveParams:      map[string]interface{}{"task_id": "upd-1", "description": "new desc"},
			wantDescription: "new desc",
		},
		{
			give:       "not found",
			giveParams: map[string]interface{}{"task_id": "ghost", "status": "done"},
			wantErr:    true,
		},
		{
			give:       "missing task_id",
			giveParams: map[string]interface{}{"status": "done"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result, err := updateTool.call(tt.giveParams)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			m := result.(map[string]interface{})
			if tt.wantStatus != "" {
				assert.Equal(t, tt.wantStatus, m["status"])
			}
			if tt.wantDescription != "" {
				assert.Equal(t, tt.wantDescription, m["description"])
			}
		})
	}
}

func TestBuildTaskTools_ToolCount(t *testing.T) {
	store := NewInMemoryTaskStore()
	tools := BuildTaskTools(store)
	assert.Len(t, tools, 4)

	names := make(map[string]bool, 4)
	for _, tool := range tools {
		names[tool.Name] = true
	}
	assert.True(t, names["task_create"])
	assert.True(t, names["task_get"])
	assert.True(t, names["task_list"])
	assert.True(t, names["task_update"])
}

// --- Helpers ---

type taskToolHelper struct {
	tool *agent.Tool
}

func (h *taskToolHelper) call(params map[string]interface{}) (interface{}, error) {
	return h.tool.Handler(context.Background(), params)
}

func findTool(t *testing.T, tools []*agent.Tool, name string) *taskToolHelper {
	t.Helper()
	for _, tool := range tools {
		if tool.Name == name {
			return &taskToolHelper{tool}
		}
	}
	t.Fatalf("tool %q not found", name)
	return nil
}
