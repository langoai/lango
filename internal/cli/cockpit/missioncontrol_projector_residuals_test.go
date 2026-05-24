package cockpit

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/mission"
)

func TestMissionControlProjectorNilReceiverReturnsZeroSnapshot(t *testing.T) {
	t.Parallel()

	var projector *MissionControlProjector
	snapshot := projector.Project(nil)

	assert.Equal(t, MissionControlSnapshot{}, snapshot)
}

func TestMissionControlProjectorLearningSuggestionFallbackTitleUsesInspectDefault(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 21, 9, 0, 0, 0, time.UTC)
	buffer := NewLearningSuggestionBuffer(func() time.Time { return now })
	buffer.Append(eventbus.LearningSuggestionEvent{
		SuggestionID: "  suggestion-83  ",
		ProposedRule: " \t\n ",
		Pattern:      "   ",
		Rationale:    "Inspect raw evidence before applying a rule.",
		Timestamp:    now,
	})

	projector := NewMissionControlProjector(Deps{
		SessionKey:     "sess-83",
		LearningBuffer: buffer,
		MissionReader:  stubMissionControlMissionReader{missions: map[string][]*mission.Mission{"sess-83": {}}},
		RunLedgerStore: stubMissionControlRunLedgerReader{},
		AgentRunStore:  stubMissionControlAgentRunReader{},
	})

	snapshot := projector.Project(nil)

	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, "learn:suggestion-83", snapshot.Missions[0].ID)
	assert.Equal(t, MissionKindProposed, snapshot.Missions[0].Kind)
	assert.Equal(t, MissionStatusPending, snapshot.Missions[0].Status)
	assert.Equal(t, "Apply learning rule: Inspect learning suggestion", snapshot.Missions[0].Title)
	assert.Equal(t, "proposed_learning", snapshot.Missions[0].SourceKind)
	assert.Equal(t, "suggestion-83", snapshot.Missions[0].SourceRef)
	assert.Equal(t, now, snapshot.Missions[0].UpdatedAt)
}

func TestMissionControlProjectorCollaborationLinkReaderErrorDegradesSnapshot(t *testing.T) {
	t.Parallel()

	missionID := uuid.MustParse("83838383-8383-8383-8383-838383838383")
	projector := NewMissionControlProjector(Deps{
		SessionKey: "sess-83",
		MissionReader: stubMissionControlMissionReader{
			missions: map[string][]*mission.Mission{
				"sess-83": {{
					ID:         missionID,
					SessionKey: "sess-83",
					Title:      "Durable collaboration mission",
					Status:     mission.StatusActive,
					SourceKind: "user",
					UpdatedAt:  time.Date(2026, 5, 21, 9, 5, 0, 0, time.UTC),
					CreatedAt:  time.Date(2026, 5, 21, 9, 0, 0, 0, time.UTC),
				}},
			},
		},
		CollabMissionLinks: stubMissionControlCollabMissionLinkReader{err: errors.New("collaboration links unavailable")},
		RunLedgerStore:     stubMissionControlRunLedgerReader{},
		AgentRunStore:      stubMissionControlAgentRunReader{},
	})

	snapshot := projector.Project(nil)

	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, missionID.String(), snapshot.Missions[0].ID)
	assert.True(t, snapshot.Degraded)
	assert.Contains(t, snapshot.Header.DegradedNote, "Collaboration context unavailable")
}

func TestMissionControlProjectorActivityProjectionEmptyBufferReturnsNil(t *testing.T) {
	t.Parallel()

	projector := NewMissionControlProjector(Deps{
		SessionKey:     "sess-83",
		ActivityBuffer: NewMissionActivityBuffer(),
		MissionReader:  stubMissionControlMissionReader{missions: map[string][]*mission.Mission{"sess-83": {}}},
		RunLedgerStore: stubMissionControlRunLedgerReader{},
		AgentRunStore:  stubMissionControlAgentRunReader{},
	})

	snapshot := projector.Project(nil)

	assert.Empty(t, snapshot.Activities)
	assert.False(t, snapshot.Degraded)
	assert.Empty(t, snapshot.Header.DegradedNote)
}
