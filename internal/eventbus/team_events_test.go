package eventbus

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamEventNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give Event
		want string
	}{
		{give: TeamFormedEvent{}, want: "team.formed"},
		{give: TeamDisbandedEvent{}, want: "team.disbanded"},
		{give: TeamMemberJoinedEvent{}, want: "team.member.joined"},
		{give: TeamMemberLeftEvent{}, want: "team.member.left"},
		{give: TeamTaskDelegatedEvent{}, want: "team.task.delegated"},
		{give: TeamTaskCompletedEvent{}, want: "team.task.completed"},
		{give: TeamConflictDetectedEvent{}, want: "team.conflict.detected"},
		{give: TeamPaymentAgreedEvent{}, want: "team.payment.agreed"},
		{give: TeamHealthCheckEvent{}, want: "team.health.check"},
		{give: TeamLeaderChangedEvent{}, want: "team.leader.changed"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.give.EventName())
		})
	}
}

func TestTeamEvents_PublishSubscribe(t *testing.T) {
	t.Parallel()

	bus := New()

	var received []Event

	SubscribeTyped(bus, func(e TeamFormedEvent) {
		received = append(received, e)
	})
	SubscribeTyped(bus, func(e TeamDisbandedEvent) {
		received = append(received, e)
	})
	SubscribeTyped(bus, func(e TeamTaskCompletedEvent) {
		received = append(received, e)
	})

	bus.Publish(TeamFormedEvent{TeamID: "t1", Name: "alpha", Members: 3})
	bus.Publish(TeamTaskCompletedEvent{TeamID: "t1", ToolName: "search", Successful: 2, Failed: 1, Duration: time.Second})
	bus.Publish(TeamDisbandedEvent{TeamID: "t1", Reason: "task complete"})

	require.Len(t, received, 3)

	// Verify ordering.
	assert.IsType(t, TeamFormedEvent{}, received[0])
	assert.IsType(t, TeamTaskCompletedEvent{}, received[1])
	assert.IsType(t, TeamDisbandedEvent{}, received[2])
}
