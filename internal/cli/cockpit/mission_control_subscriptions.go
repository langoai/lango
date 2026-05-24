package cockpit

import (
	"time"

	"github.com/langoai/lango/internal/eventbus"
)

// SubscribeMissionControlEvents wires cockpit-lifetime EventBus subscriptions
// into the shared state used by Mission Control projection.
func SubscribeMissionControlEvents(bus *eventbus.Bus, sessionKey string, learning *LearningSuggestionBuffer, activity *MissionActivityBuffer) {
	if bus == nil {
		return
	}

	eventbus.SubscribeTyped(bus, func(e eventbus.LearningSuggestionEvent) {
		if e.SessionKey != sessionKey {
			return
		}
		if learning != nil {
			learning.Append(e)
		}
		if activity != nil {
			activity.Append(newLearningSuggestionActivity(e))
		}
	})

	eventbus.SubscribeTyped(bus, func(e eventbus.CompactionCompletedEvent) {
		if e.SessionKey != sessionKey || activity == nil {
			return
		}
		activity.Append(newCompactionCompletedActivity(e))
	})

	eventbus.SubscribeTyped(bus, func(e eventbus.CompactionSlowEvent) {
		if e.SessionKey != sessionKey || activity == nil {
			return
		}
		activity.Append(newCompactionSlowActivity(e))
	})

	eventbus.SubscribeTyped(bus, func(e eventbus.ModeChangedEvent) {
		if e.SessionKey != sessionKey || activity == nil {
			return
		}
		activity.Append(newModeChangedActivity(e, time.Now()))
	})

	eventbus.SubscribeTyped(bus, func(e eventbus.TurnCompletedEvent) {
		if e.SessionKey != sessionKey || activity == nil {
			return
		}
		activity.Append(newTurnCompletedActivity(e, time.Now()))
	})

	eventbus.SubscribeTyped(bus, func(e eventbus.PolicyDecisionEvent) {
		if e.SessionKey != sessionKey || activity == nil {
			return
		}
		activity.Append(newPolicyDecisionActivity(e, time.Now()))
	})

	eventbus.SubscribeTyped(bus, func(e eventbus.AlertEvent) {
		if e.SessionKey != sessionKey || activity == nil {
			return
		}
		activity.Append(newAlertActivity(e))
	})

}
