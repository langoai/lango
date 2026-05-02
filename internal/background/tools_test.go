package background

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ctxkeys"
)

type stubMissionLinker struct {
	taskID    string
	origin    Origin
	prompt    string
	calls     int
	missionID string
	err       error
}

func (s *stubMissionLinker) LinkBackgroundTask(ctx context.Context, taskID string, origin Origin, prompt string) error {
	s.calls++
	s.taskID = taskID
	s.origin = origin
	s.prompt = prompt
	s.missionID = ctxkeys.MissionIDFromContext(ctx)
	return s.err
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
		ctx := ctxkeys.WithMissionID(context.Background(), "mission-bg-1")
		result, err := tool.Handler(ctx, map[string]interface{}{
			"prompt": "ship mission",
		})
		require.NoError(t, err)
		payload := result.(map[string]interface{})
		require.NotEmpty(t, payload["task_id"])
		assert.Equal(t, 1, linker.calls)
		assert.Equal(t, payload["task_id"], linker.taskID)
		assert.Equal(t, "ship mission", linker.prompt)
		assert.Equal(t, "telegram:default", linker.origin.Channel)
		assert.Equal(t, "mission-bg-1", linker.missionID)
		break
	}

	require.True(t, submitToolFound)
}

func TestBuildTools_SubmitPropagatesMissionLinkFailure(t *testing.T) {
	t.Parallel()

	mgr := NewManager(&mockRunner{result: "done"}, nil, 2, time.Minute, testLogger())
	linker := &stubMissionLinker{err: errors.New("link failed")}
	tools := BuildTools(mgr, []string{"telegram:default"}, linker)

	for _, tool := range tools {
		if tool.Name != "bg_submit" {
			continue
		}
		ctx := ctxkeys.WithMissionID(context.Background(), "mission-bg-fail")
		result, err := tool.Handler(ctx, map[string]interface{}{
			"prompt": "ship mission",
		})
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "mission link failed")
		require.Len(t, mgr.List(), 1)
		assert.Equal(t, Cancelled, mgr.List()[0].Status)
		return
	}

	t.Fatal("bg_submit tool not found")
}
