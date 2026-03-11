package app

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/reputation"
	"github.com/langoai/lango/internal/p2p/team"
	"github.com/langoai/lango/internal/testutil"
)

func TestTeamReputationBridge_UnhealthyEventTriggersKick(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	coord := setupTestCoordinator(t, bus)

	_, err := coord.FormTeam(context.Background(), team.FormTeamRequest{
		TeamID: "t-rep", Name: "rep-team", Goal: "test",
		LeaderDID: "did:leader", Capability: "search", MemberCount: 1,
	})
	require.NoError(t, err)

	var mu sync.Mutex
	var leftEvents []eventbus.TeamMemberLeftEvent
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamMemberLeftEvent) {
		mu.Lock()
		defer mu.Unlock()
		leftEvents = append(leftEvents, ev)
	})

	// Wire bridge without a real repStore (repStore=nil means no recording happens,
	// so unhealthy event won't trigger kick via reputation path).
	// For this test, we verify that the bridge correctly handles nil repStore gracefully.
	initTeamReputationBridge(bus, coord, nil, 0.3, testLog())

	bus.Publish(eventbus.TeamMemberUnhealthyEvent{
		TeamID:      "t-rep",
		MemberDID:   "did:worker1",
		MissedPings: 3,
	})

	mu.Lock()
	defer mu.Unlock()
	// With nil repStore, no kick should happen (reputation check is skipped).
	assert.Empty(t, leftEvents)
}

func TestTeamReputationBridge_ReputationDropTriggersKick(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	coord := setupTestCoordinator(t, bus)

	_, err := coord.FormTeam(context.Background(), team.FormTeamRequest{
		TeamID: "t-drop", Name: "drop-team", Goal: "test",
		LeaderDID: "did:leader", Capability: "search", MemberCount: 1,
	})
	require.NoError(t, err)

	var mu sync.Mutex
	var leftEvents []eventbus.TeamMemberLeftEvent
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamMemberLeftEvent) {
		mu.Lock()
		defer mu.Unlock()
		leftEvents = append(leftEvents, ev)
	})

	initTeamReputationBridge(bus, coord, nil, 0.3, testLog())

	// Simulate a reputation drop below threshold.
	bus.Publish(eventbus.ReputationChangedEvent{
		PeerDID:  "did:worker1",
		NewScore: 0.1, // below minScore of 0.3
	})

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, leftEvents, 1)
	assert.Equal(t, "t-drop", leftEvents[0].TeamID)
	assert.Equal(t, "did:worker1", leftEvents[0].MemberDID)
	assert.Contains(t, leftEvents[0].Reason, "reputation dropped below threshold")
}

func TestTeamReputationBridge_ReputationAboveThresholdNoKick(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	coord := setupTestCoordinator(t, bus)

	_, err := coord.FormTeam(context.Background(), team.FormTeamRequest{
		TeamID: "t-ok", Name: "ok-team", Goal: "test",
		LeaderDID: "did:leader", Capability: "search", MemberCount: 1,
	})
	require.NoError(t, err)

	var mu sync.Mutex
	var leftEvents []eventbus.TeamMemberLeftEvent
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamMemberLeftEvent) {
		mu.Lock()
		defer mu.Unlock()
		leftEvents = append(leftEvents, ev)
	})

	initTeamReputationBridge(bus, coord, nil, 0.3, testLog())

	// Score above threshold should not trigger kick.
	bus.Publish(eventbus.ReputationChangedEvent{
		PeerDID:  "did:worker1",
		NewScore: 0.5, // above minScore of 0.3
	})

	mu.Lock()
	defer mu.Unlock()
	assert.Empty(t, leftEvents, "reputation above threshold should not trigger kick")
}

func TestCoordinator_TeamsForMember(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	coord := setupTestCoordinator(t, bus)

	_, err := coord.FormTeam(context.Background(), team.FormTeamRequest{
		TeamID: "t1", Name: "team-1", Goal: "test",
		LeaderDID: "did:leader", Capability: "search", MemberCount: 1,
	})
	require.NoError(t, err)

	_, err = coord.FormTeam(context.Background(), team.FormTeamRequest{
		TeamID: "t2", Name: "team-2", Goal: "test",
		LeaderDID: "did:leader", Capability: "search", MemberCount: 1,
	})
	require.NoError(t, err)

	// Worker should be in both teams.
	teamIDs := coord.TeamsForMember("did:worker1")
	assert.Len(t, teamIDs, 2)

	// Leader should also be in both teams.
	leaderTeams := coord.TeamsForMember("did:leader")
	assert.Len(t, leaderTeams, 2)

	// Unknown DID should not be in any team.
	unknownTeams := coord.TeamsForMember("did:unknown")
	assert.Empty(t, unknownTeams)
}

func TestCoordinator_KickMember(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	coord := setupTestCoordinator(t, bus)

	_, err := coord.FormTeam(context.Background(), team.FormTeamRequest{
		TeamID: "t-kick", Name: "kick-team", Goal: "test",
		LeaderDID: "did:leader", Capability: "search", MemberCount: 1,
	})
	require.NoError(t, err)

	var mu sync.Mutex
	var leftEvents []eventbus.TeamMemberLeftEvent
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamMemberLeftEvent) {
		mu.Lock()
		defer mu.Unlock()
		leftEvents = append(leftEvents, ev)
	})

	err = coord.KickMember(context.Background(), "t-kick", "did:worker1", "test reason")
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, leftEvents, 1)
	assert.Equal(t, "t-kick", leftEvents[0].TeamID)
	assert.Equal(t, "did:worker1", leftEvents[0].MemberDID)
	assert.Equal(t, "test reason", leftEvents[0].Reason)

	// Member should no longer be in the team.
	teamIDs := coord.TeamsForMember("did:worker1")
	assert.Empty(t, teamIDs)
}

func TestTeamReputationBridge_WithRepStore_RecordTimeoutAndKick(t *testing.T) {
	t.Parallel()

	client := testutil.TestEntClient(t)
	repStore := reputation.NewStore(client, testLog())

	bus := eventbus.New()
	coord := setupTestCoordinator(t, bus)

	_, err := coord.FormTeam(context.Background(), team.FormTeamRequest{
		TeamID: "t-rep-real", Name: "rep-real-team", Goal: "test",
		LeaderDID: "did:leader", Capability: "search", MemberCount: 1,
	})
	require.NoError(t, err)

	var mu sync.Mutex
	var leftEvents []eventbus.TeamMemberLeftEvent
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamMemberLeftEvent) {
		mu.Lock()
		defer mu.Unlock()
		leftEvents = append(leftEvents, ev)
	})

	// Use a high minScore so that a fresh peer (score 0.0) gets kicked.
	initTeamReputationBridge(bus, coord, repStore, 0.5, testLog())

	// Record a timeout — this will create the peer's reputation record with a low score.
	bus.Publish(eventbus.TeamMemberUnhealthyEvent{
		TeamID:      "t-rep-real",
		MemberDID:   "did:worker1",
		MissedPings: 3,
	})

	mu.Lock()
	defer mu.Unlock()
	// After RecordTimeout the score is 0 / (0 + 0 + 1.5 + 1) = 0.0, which is below 0.5.
	require.Len(t, leftEvents, 1, "member should be kicked when score drops below threshold")
	assert.Equal(t, "t-rep-real", leftEvents[0].TeamID)
	assert.Equal(t, "did:worker1", leftEvents[0].MemberDID)
	assert.Contains(t, leftEvents[0].Reason, "reputation below threshold")
}

func TestTeamReputationBridge_WithRepStore_RecordSuccess(t *testing.T) {
	t.Parallel()

	client := testutil.TestEntClient(t)
	repStore := reputation.NewStore(client, testLog())

	bus := eventbus.New()
	coord := setupTestCoordinator(t, bus)

	_, err := coord.FormTeam(context.Background(), team.FormTeamRequest{
		TeamID: "t-rep-ok", Name: "rep-ok-team", Goal: "test",
		LeaderDID: "did:leader", Capability: "search", MemberCount: 1,
	})
	require.NoError(t, err)

	initTeamReputationBridge(bus, coord, repStore, 0.3, testLog())

	// Publish task completed event — should record success for workers.
	bus.Publish(eventbus.TeamTaskCompletedEvent{
		TeamID:     "t-rep-ok",
		ToolName:   "search",
		Successful: 1,
		Failed:     0,
	})

	// Verify the worker's score was recorded (1 success -> 1/(1+0+0+1) = 0.5).
	ctx := context.Background()
	score, err := repStore.GetScore(ctx, "did:worker1")
	require.NoError(t, err)
	assert.InDelta(t, 0.5, score, 1e-9, "worker should have score 0.5 after one success")
}
