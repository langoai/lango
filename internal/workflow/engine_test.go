package workflow

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockAgentRunner struct {
	mu       sync.Mutex
	result   string
	err      error
	delay    time.Duration
	sessions []string
}

func (m *mockAgentRunner) Run(ctx context.Context, sessionKey string, _ string) (string, error) {
	m.mu.Lock()
	m.sessions = append(m.sessions, sessionKey)
	m.mu.Unlock()
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return m.result, m.err
}

func TestEngine_ExecuteStep_ChecksCancellation(t *testing.T) {
	t.Parallel()

	runner := &mockAgentRunner{result: "ok"}
	logger := zap.NewNop().Sugar()

	e := &Engine{
		runner:         runner,
		maxConcurrent:  4,
		defaultTimeout: 5 * time.Minute,
		logger:         logger,
		cancels:        make(map[string]context.CancelFunc),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	step := &Step{ID: "step-1", Prompt: "do something"}
	_, err := e.executeStep(ctx, "run-1", "wf", step, nil)
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	runner.mu.Lock()
	assert.Empty(t, runner.sessions)
	runner.mu.Unlock()
}

func TestEngine_SessionKeyFormat(t *testing.T) {
	t.Parallel()

	key1 := fmt.Sprintf("workflow:%s:%s:%s", "my-wf", "run-1", "step-a")
	key2 := fmt.Sprintf("workflow:%s:%s:%s", "my-wf", "run-2", "step-a")

	assert.Equal(t, "workflow:my-wf:run-1:step-a", key1)
	assert.Equal(t, "workflow:my-wf:run-2:step-a", key2)
	assert.NotEqual(t, key1, key2, "different runIDs must produce different session keys")
	assert.True(t, strings.Contains(key1, "run-1"))
	assert.True(t, strings.Contains(key2, "run-2"))
}

func TestEngine_ExecuteStep_RunnerError(t *testing.T) {
	t.Parallel()

	runner := &mockAgentRunner{err: fmt.Errorf("agent failed")}
	logger := zap.NewNop().Sugar()

	e := &Engine{
		runner:         runner,
		maxConcurrent:  4,
		defaultTimeout: 5 * time.Minute,
		logger:         logger,
		cancels:        make(map[string]context.CancelFunc),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	step := &Step{ID: "step-1", Prompt: "fail"}
	_, err := e.executeStep(ctx, "run-1", "wf", step, nil)
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestEngineShutdownHonorsContextDeadline(t *testing.T) {
	t.Parallel()

	engine := NewEngine(nil, nil, nil, 1, time.Minute, zap.NewNop().Sugar())

	release := make(chan struct{})
	engine.wg.Add(1)
	go func() {
		defer engine.wg.Done()
		<-release
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := engine.Shutdown(ctx)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	close(release)
	require.NoError(t, engine.Shutdown(context.Background()))
}
