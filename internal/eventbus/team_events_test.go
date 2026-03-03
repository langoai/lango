package eventbus

import (
	"testing"
	"time"
)

func TestTeamEventNames(t *testing.T) {
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
			if got := tt.give.EventName(); got != tt.want {
				t.Errorf("EventName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTeamEvents_PublishSubscribe(t *testing.T) {
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

	if len(received) != 3 {
		t.Fatalf("received %d events, want 3", len(received))
	}

	// Verify ordering.
	if _, ok := received[0].(TeamFormedEvent); !ok {
		t.Errorf("event[0] type = %T, want TeamFormedEvent", received[0])
	}
	if _, ok := received[1].(TeamTaskCompletedEvent); !ok {
		t.Errorf("event[1] type = %T, want TeamTaskCompletedEvent", received[1])
	}
	if _, ok := received[2].(TeamDisbandedEvent); !ok {
		t.Errorf("event[2] type = %T, want TeamDisbandedEvent", received[2])
	}
}
