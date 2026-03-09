package observability

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordTokenUsage(t *testing.T) {
	tests := []struct {
		give         []TokenUsage
		wantInput    int64
		wantOutput   int64
		wantTotal    int64
		wantCache    int64
		wantSessions int
		wantAgents   int
	}{
		{
			give: []TokenUsage{
				{
					SessionKey:   "s1",
					AgentName:    "agent1",
					InputTokens:  100,
					OutputTokens: 50,
					TotalTokens:  150,
					CacheTokens:  10,
				},
				{
					SessionKey:   "s1",
					AgentName:    "agent1",
					InputTokens:  200,
					OutputTokens: 100,
					TotalTokens:  300,
					CacheTokens:  20,
				},
			},
			wantInput:    300,
			wantOutput:   150,
			wantTotal:    450,
			wantCache:    30,
			wantSessions: 1,
			wantAgents:   1,
		},
		{
			give: []TokenUsage{
				{
					SessionKey:   "s1",
					AgentName:    "agent1",
					InputTokens:  100,
					OutputTokens: 50,
					TotalTokens:  150,
				},
				{
					SessionKey:   "s2",
					AgentName:    "agent2",
					InputTokens:  200,
					OutputTokens: 100,
					TotalTokens:  300,
				},
			},
			wantInput:    300,
			wantOutput:   150,
			wantTotal:    450,
			wantSessions: 2,
			wantAgents:   2,
		},
		{
			give: []TokenUsage{
				{
					InputTokens:  100,
					OutputTokens: 50,
					TotalTokens:  150,
				},
			},
			wantInput:    100,
			wantOutput:   50,
			wantTotal:    150,
			wantSessions: 0,
			wantAgents:   0,
		},
	}

	for _, tt := range tests {
		c := NewCollector()
		for _, u := range tt.give {
			c.RecordTokenUsage(u)
		}

		snap := c.Snapshot()
		assert.Equal(t, tt.wantInput, snap.TokenUsageTotal.InputTokens)
		assert.Equal(t, tt.wantOutput, snap.TokenUsageTotal.OutputTokens)
		assert.Equal(t, tt.wantTotal, snap.TokenUsageTotal.TotalTokens)
		assert.Equal(t, tt.wantCache, snap.TokenUsageTotal.CacheTokens)
		assert.Len(t, snap.SessionBreakdown, tt.wantSessions)
		assert.Len(t, snap.AgentBreakdown, tt.wantAgents)
	}
}

func TestRecordToolExecution(t *testing.T) {
	tests := []struct {
		give          string
		giveAgent     string
		giveSuccess   bool
		giveDuration  time.Duration
		giveCount     int
		wantCount     int64
		wantErrors    int64
		wantToolExecs int64
	}{
		{
			give:          "search",
			giveAgent:     "agent1",
			giveSuccess:   true,
			giveDuration:  100 * time.Millisecond,
			giveCount:     3,
			wantCount:     3,
			wantErrors:    0,
			wantToolExecs: 3,
		},
		{
			give:          "fetch",
			giveAgent:     "agent1",
			giveSuccess:   false,
			giveDuration:  200 * time.Millisecond,
			giveCount:     2,
			wantCount:     2,
			wantErrors:    2,
			wantToolExecs: 2,
		},
	}

	for _, tt := range tests {
		c := NewCollector()
		for range tt.giveCount {
			c.RecordToolExecution(tt.give, tt.giveAgent, tt.giveDuration, tt.giveSuccess)
		}

		snap := c.Snapshot()
		assert.Equal(t, tt.wantToolExecs, snap.ToolExecutions)

		tm, ok := snap.ToolBreakdown[tt.give]
		require.True(t, ok)
		assert.Equal(t, tt.wantCount, tm.Count)
		assert.Equal(t, tt.wantErrors, tm.Errors)
		assert.Equal(t, tt.giveDuration, tm.AvgDuration)
	}
}

func TestRecordToolExecution_AvgDuration(t *testing.T) {
	c := NewCollector()
	c.RecordToolExecution("tool1", "", 100*time.Millisecond, true)
	c.RecordToolExecution("tool1", "", 300*time.Millisecond, true)

	snap := c.Snapshot()
	tm := snap.ToolBreakdown["tool1"]
	assert.Equal(t, 200*time.Millisecond, tm.AvgDuration)
	assert.Equal(t, 400*time.Millisecond, tm.TotalDuration)
}

func TestRecordToolExecution_AgentToolCalls(t *testing.T) {
	c := NewCollector()
	c.RecordToolExecution("tool1", "agent1", time.Millisecond, true)
	c.RecordToolExecution("tool2", "agent1", time.Millisecond, true)
	c.RecordToolExecution("tool1", "agent2", time.Millisecond, true)

	snap := c.Snapshot()
	assert.Equal(t, int64(2), snap.AgentBreakdown["agent1"].ToolCalls)
	assert.Equal(t, int64(1), snap.AgentBreakdown["agent2"].ToolCalls)
}

func TestSnapshot(t *testing.T) {
	c := NewCollector()
	c.RecordTokenUsage(TokenUsage{
		SessionKey:   "s1",
		AgentName:    "a1",
		InputTokens:  500,
		OutputTokens: 200,
		TotalTokens:  700,
	})
	c.RecordToolExecution("search", "a1", 50*time.Millisecond, true)

	snap := c.Snapshot()

	assert.False(t, snap.StartedAt.IsZero())
	assert.True(t, snap.Uptime > 0)
	assert.Equal(t, int64(500), snap.TokenUsageTotal.InputTokens)
	assert.Equal(t, int64(200), snap.TokenUsageTotal.OutputTokens)
	assert.Equal(t, int64(700), snap.TokenUsageTotal.TotalTokens)
	assert.Equal(t, int64(1), snap.ToolExecutions)
	assert.Len(t, snap.ToolBreakdown, 1)
	assert.Len(t, snap.AgentBreakdown, 1)
	assert.Len(t, snap.SessionBreakdown, 1)

	// Verify snapshot is a copy (mutations don't affect collector)
	snap.ToolBreakdown["injected"] = ToolMetric{Name: "injected"}
	snap2 := c.Snapshot()
	_, exists := snap2.ToolBreakdown["injected"]
	assert.False(t, exists)
}

func TestSessionMetrics(t *testing.T) {
	c := NewCollector()

	// Unknown session returns nil
	assert.Nil(t, c.SessionMetrics("unknown"))

	c.RecordTokenUsage(TokenUsage{
		SessionKey:  "s1",
		InputTokens: 100,
		TotalTokens: 100,
	})

	sm := c.SessionMetrics("s1")
	require.NotNil(t, sm)
	assert.Equal(t, "s1", sm.SessionKey)
	assert.Equal(t, int64(100), sm.InputTokens)
	assert.Equal(t, int64(1), sm.RequestCount)

	// Verify it's a copy
	sm.InputTokens = 9999
	sm2 := c.SessionMetrics("s1")
	assert.Equal(t, int64(100), sm2.InputTokens)
}

func TestTopSessions(t *testing.T) {
	c := NewCollector()
	c.RecordTokenUsage(TokenUsage{SessionKey: "s1", TotalTokens: 100})
	c.RecordTokenUsage(TokenUsage{SessionKey: "s2", TotalTokens: 300})
	c.RecordTokenUsage(TokenUsage{SessionKey: "s3", TotalTokens: 200})

	top := c.TopSessions(2)
	require.Len(t, top, 2)
	assert.Equal(t, "s2", top[0].SessionKey)
	assert.Equal(t, "s3", top[1].SessionKey)

	// No limit
	all := c.TopSessions(0)
	assert.Len(t, all, 3)

	// Limit larger than count
	all2 := c.TopSessions(10)
	assert.Len(t, all2, 3)
}

func TestReset(t *testing.T) {
	c := NewCollector()
	c.RecordTokenUsage(TokenUsage{
		SessionKey:   "s1",
		AgentName:    "a1",
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
	})
	c.RecordToolExecution("tool1", "a1", time.Millisecond, true)

	c.Reset()

	snap := c.Snapshot()
	assert.Equal(t, int64(0), snap.TokenUsageTotal.InputTokens)
	assert.Equal(t, int64(0), snap.TokenUsageTotal.OutputTokens)
	assert.Equal(t, int64(0), snap.TokenUsageTotal.TotalTokens)
	assert.Equal(t, int64(0), snap.ToolExecutions)
	assert.Empty(t, snap.ToolBreakdown)
	assert.Empty(t, snap.AgentBreakdown)
	assert.Empty(t, snap.SessionBreakdown)
}

func TestConcurrency(t *testing.T) {
	c := NewCollector()
	var wg sync.WaitGroup

	// Parallel writers
	for i := range 100 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			c.RecordTokenUsage(TokenUsage{
				SessionKey:  "s1",
				AgentName:   "a1",
				InputTokens: 10,
				TotalTokens: 10,
			})
			c.RecordToolExecution("tool1", "a1", time.Millisecond, idx%5 != 0)
		}(i)
	}

	// Parallel readers
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.Snapshot()
			_ = c.SessionMetrics("s1")
			_ = c.TopSessions(5)
		}()
	}

	wg.Wait()

	snap := c.Snapshot()
	assert.Equal(t, int64(1000), snap.TokenUsageTotal.InputTokens)
	assert.Equal(t, int64(100), snap.ToolExecutions)
}
