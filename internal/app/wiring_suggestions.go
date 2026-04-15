package app

import (
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/learning"
)

// wireLearningSuggestions constructs the suggestion emitter, wires it into
// the engine, and subscribes it to TurnCompletedEvent for per-session turn
// ticking. A nil learning engine or disabled config is a no-op.
func wireLearningSuggestions(app *App) {
	if app == nil || app.LearningEngine == nil || app.EventBus == nil {
		return
	}
	cfg := app.Config.Learning.Suggestions.ResolveSuggestions()
	if cfg.Enabled == nil || !*cfg.Enabled {
		return
	}

	em := learning.NewSuggestionEmitter(
		app.EventBus,
		cfg.Threshold,
		cfg.RateLimit,
		cfg.DedupWindow,
	)
	app.LearningEngine.WithSuggestionEmitter(em)
	app.LearningSuggestionEmitter = em

	// Tick per-session turn counters on TurnCompletedEvent.
	eventbus.SubscribeTyped(app.EventBus, func(e eventbus.TurnCompletedEvent) {
		if e.SessionKey == "" {
			return
		}
		em.TickTurn(e.SessionKey)
	})
}
