package background

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubMissionLinker struct {
	taskID string
	origin Origin
	prompt string
	calls  int
}

func (s *stubMissionLinker) LinkBackgroundTask(_ context.Context, taskID string, origin Origin, prompt string) error {
	s.calls++
	s.taskID = taskID
	s.origin = origin
	s.prompt = prompt
	return nil
}

func TestBuildTools_SubmitInvokesMissionLinker(t *testing.T) {
	t.Parallel()

	mgr := NewManager(&mockRunner{result: "done"}, nil, 2, time.Minute, testLogger())
	linker := &stubMissionLinker{}
	tools := BuildTools(mgr, []string{"telegram:default"}, linker)

	var submitToolFound bool
	for _, tool := range tools {
		if tool.Name != "bg_submit" {
			continue
		}
		submitToolFound = true
		result, err := tool.Handler(context.Background(), map[string]interface{}{
			"prompt": "ship mission",
		})
		require.NoError(t, err)
		payload := result.(map[string]interface{})
		require.NotEmpty(t, payload["task_id"])
		assert.Equal(t, 1, linker.calls)
		assert.Equal(t, payload["task_id"], linker.taskID)
		assert.Equal(t, "ship mission", linker.prompt)
		assert.Equal(t, "telegram:default", linker.origin.Channel)
		break
	}

	require.True(t, submitToolFound)
}
