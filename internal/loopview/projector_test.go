package loopview

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type projectorClock struct {
	now time.Time
}

func (c *projectorClock) Now() time.Time { return c.now }

func TestLoopStatusOrdering(t *testing.T) {
	t.Parallel()

	clock := &projectorClock{now: time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)}
	projector := NewProjector(clock.Now)

	agenda := projector.Project(ProjectionInput{
		SessionKey: "sess-1",
		Missions: []MissionSource{
			{MissionID: "m-active", SessionKey: "sess-1", Title: "Active", Status: "active", UpdatedAt: clock.now.Add(-3 * time.Minute)},
			{MissionID: "m-blocked", SessionKey: "sess-1", Title: "Blocked", Status: "blocked", UpdatedAt: clock.now.Add(-2 * time.Minute)},
			{MissionID: "m-wait", SessionKey: "sess-1", Title: "Waiting", Status: "waiting_decision", UpdatedAt: clock.now.Add(-1 * time.Minute)},
			{MissionID: "m-done", SessionKey: "sess-1", Title: "Done", Status: "done", UpdatedAt: clock.now.Add(-4 * time.Minute)},
		},
		CronJobs: []CronSource{
			{JobID: "cron-1", Name: "Nightly digest", Enabled: true, NextRunAt: clock.now.Add(time.Hour)},
		},
	})

	require.Len(t, agenda.Loops, 5)
	assert.Equal(t, "mission:m-wait", agenda.Loops[0].LoopID)
	assert.Equal(t, "mission:m-blocked", agenda.Loops[1].LoopID)
	assert.Equal(t, "mission:m-active", agenda.Loops[2].LoopID)
	assert.Equal(t, "cron:cron-1", agenda.Loops[3].LoopID)
	assert.Equal(t, "mission:m-done", agenda.Loops[4].LoopID)
}

func TestSessionScoping(t *testing.T) {
	t.Parallel()

	clock := &projectorClock{now: time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)}
	projector := NewProjector(clock.Now)

	agenda := projector.Project(ProjectionInput{
		SessionKey: "sess-1",
		Missions: []MissionSource{
			{MissionID: "m-1", SessionKey: "sess-1", Title: "Mine", Status: "active", UpdatedAt: clock.now},
			{MissionID: "m-2", SessionKey: "sess-2", Title: "Other", Status: "active", UpdatedAt: clock.now},
		},
		Inquiries: []InquirySource{
			{InquiryID: "i-1", SessionKey: "sess-1", Topic: "Mine inquiry", Question: "Q1", CreatedAt: clock.now},
			{InquiryID: "i-2", SessionKey: "sess-2", Topic: "Other inquiry", Question: "Q2", CreatedAt: clock.now},
		},
		DeadLetters: []DeadLetterSource{
			{ReferenceID: "d-1", Title: "Global dead letter", Retryable: true, UpdatedAt: clock.now},
		},
		CronJobs: []CronSource{
			{JobID: "c-1", Name: "Global cron", Enabled: true, NextRunAt: clock.now.Add(time.Hour)},
		},
	})

	ids := make([]string, 0, len(agenda.Loops))
	for _, loop := range agenda.Loops {
		ids = append(ids, loop.LoopID)
	}

	assert.Contains(t, ids, "mission:m-1")
	assert.NotContains(t, ids, "mission:m-2")
	assert.Contains(t, ids, "inquiry:i-1")
	assert.NotContains(t, ids, "inquiry:i-2")
	assert.Contains(t, ids, "dead-letter:d-1")
	assert.Contains(t, ids, "cron:c-1")
}

func TestMissionLoops(t *testing.T) {
	t.Parallel()

	clock := &projectorClock{now: time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)}
	projector := NewProjector(clock.Now)

	agenda := projector.Project(ProjectionInput{
		SessionKey: "sess-1",
		Missions: []MissionSource{
			{MissionID: "m-1", SessionKey: "sess-1", Title: "Wait", Status: "waiting_decision", UpdatedAt: clock.now},
			{MissionID: "m-2", SessionKey: "sess-1", Title: "Block", Status: "blocked", UpdatedAt: clock.now},
			{MissionID: "m-3", SessionKey: "sess-1", Title: "Done", Status: "done", UpdatedAt: clock.now},
		},
	})

	require.Len(t, agenda.Loops, 3)
	assert.Equal(t, LoopKindMissionCluster, agenda.Loops[0].LoopKind)
	assert.Equal(t, LoopStatusWaitingUser, agenda.Loops[0].Status)
	assert.Equal(t, LoopStatusBlocked, agenda.Loops[1].Status)
	assert.Equal(t, LoopStatusResolved, agenda.Loops[2].Status)
}

func TestInquiryLoops(t *testing.T) {
	t.Parallel()

	clock := &projectorClock{now: time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)}
	projector := NewProjector(clock.Now)

	agenda := projector.Project(ProjectionInput{
		SessionKey: "sess-1",
		Inquiries: []InquirySource{
			{InquiryID: "inq-1", SessionKey: "sess-1", Topic: "Need answer", Question: "What should we store?", CreatedAt: clock.now.Add(-time.Hour)},
		},
	})

	require.Len(t, agenda.Loops, 1)
	assert.Equal(t, LoopKindInquiry, agenda.Loops[0].LoopKind)
	assert.Equal(t, LoopStatusWaitingUser, agenda.Loops[0].Status)
}

func TestDeadLetterLoops(t *testing.T) {
	t.Parallel()

	clock := &projectorClock{now: time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)}
	projector := NewProjector(clock.Now)

	agenda := projector.Project(ProjectionInput{
		SessionKey: "sess-1",
		DeadLetters: []DeadLetterSource{
			{ReferenceID: "d-retry", Title: "Retryable", Summary: "Needs retry", Retryable: true, UpdatedAt: clock.now},
			{ReferenceID: "d-done", Title: "Reviewed", Summary: "Already reviewed", Retryable: false, UpdatedAt: clock.now.Add(-time.Minute)},
		},
	})

	require.Len(t, agenda.Loops, 2)
	assert.Equal(t, LoopKindDeadLetter, agenda.Loops[0].LoopKind)
	assert.Equal(t, LoopStatusBlocked, agenda.Loops[0].Status)
	assert.Equal(t, LoopStatusResolved, agenda.Loops[1].Status)
}

func TestFollowUpLoopsOnlyAllowedPredicates(t *testing.T) {
	t.Parallel()

	clock := &projectorClock{now: time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)}
	projector := NewProjector(clock.Now)

	agenda := projector.Project(ProjectionInput{
		SessionKey: "sess-1",
		Missions: []MissionSource{
			{MissionID: "m-review", SessionKey: "sess-1", Title: "Review me", Status: "done", NeedsReview: true, UpdatedAt: clock.now.Add(-2 * time.Hour)},
			{MissionID: "m-old", SessionKey: "sess-1", Title: "Too old", Status: "done", NeedsReview: true, UpdatedAt: clock.now.Add(-48 * time.Hour)},
			{MissionID: "m-no-review", SessionKey: "sess-1", Title: "Done no review", Status: "done", NeedsReview: false, UpdatedAt: clock.now},
		},
		Proposals: []ProposalSource{
			{ProposalID: "p-accepted", SessionKey: "sess-1", Title: "Accepted proposal", Status: "accepted", UpdatedAt: clock.now.Add(-10 * time.Minute), HasActiveExecution: false},
			{ProposalID: "p-running", SessionKey: "sess-1", Title: "Accepted with execution", Status: "accepted", UpdatedAt: clock.now.Add(-5 * time.Minute), HasActiveExecution: true},
		},
		Inquiries: []InquirySource{
			{InquiryID: "i-old", SessionKey: "sess-1", Topic: "Old inquiry", Question: "Answer me", CreatedAt: clock.now.Add(-30 * time.Hour)},
			{InquiryID: "i-fresh", SessionKey: "sess-1", Topic: "Fresh inquiry", Question: "Later", CreatedAt: clock.now.Add(-2 * time.Hour)},
		},
	})

	followUps := make([]LoopView, 0)
	for _, loop := range agenda.Loops {
		if loop.LoopKind == LoopKindFollowUp {
			followUps = append(followUps, loop)
		}
	}

	require.Len(t, followUps, 3)
	ids := []string{followUps[0].LoopID, followUps[1].LoopID, followUps[2].LoopID}
	assert.Contains(t, ids, "follow-up:proposal:p-accepted")
	assert.Contains(t, ids, "follow-up:mission:m-review")
	assert.Contains(t, ids, "follow-up:inquiry:i-old")
	assert.NotContains(t, ids, "follow-up:proposal:p-running")
	assert.NotContains(t, ids, "follow-up:mission:m-old")
}

func TestCronBasedScheduledLoopsOnly(t *testing.T) {
	t.Parallel()

	clock := &projectorClock{now: time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)}
	projector := NewProjector(clock.Now)

	agenda := projector.Project(ProjectionInput{
		SessionKey: "sess-1",
		CronJobs: []CronSource{
			{JobID: "cron-ok", Name: "Digest", Enabled: true, NextRunAt: clock.now.Add(2 * time.Hour)},
			{JobID: "cron-failed", Name: "Sync", Enabled: true, LastRunStatus: "failed", LastRunAt: clock.now.Add(-time.Hour)},
			{JobID: "cron-disabled", Name: "Disabled", Enabled: false, NextRunAt: clock.now.Add(2 * time.Hour)},
		},
	})

	require.Len(t, agenda.Loops, 2)
	assert.Equal(t, "cron:cron-failed", agenda.Loops[0].LoopID)
	assert.Equal(t, LoopStatusBlocked, agenda.Loops[0].Status)
	assert.Equal(t, "cron:cron-ok", agenda.Loops[1].LoopID)
	assert.Equal(t, LoopStatusScheduled, agenda.Loops[1].Status)
}

func TestNoFabricatedScheduledLoopsWhenSourceUnavailable(t *testing.T) {
	t.Parallel()

	clock := &projectorClock{now: time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)}
	projector := NewProjector(clock.Now)

	agenda := projector.Project(ProjectionInput{
		SessionKey: "sess-1",
		CronJobs:   nil,
	})

	for _, loop := range agenda.Loops {
		assert.NotEqual(t, LoopKindScheduledAutomation, loop.LoopKind)
	}
	assert.Empty(t, agenda.Loops)
}
