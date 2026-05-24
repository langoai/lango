package app

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/lifecycle"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
)

// wireCompactionBuffer constructs the CompactionBuffer, registers it in the
// lifecycle, and subscribes it to TurnCompletedEvent for post-turn hygiene.
// A nil session store or disabled config is a no-op.
func wireCompactionBuffer(app *App, modelWindow int) {
	if app == nil || app.Store == nil {
		return
	}
	cfg := app.Config.Context.Compaction.ResolveCompaction()
	if cfg.Enabled == nil || !*cfg.Enabled {
		return
	}
	entStore, ok := app.Store.(*session.EntStore)
	if !ok {
		return
	}

	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()

	buf := session.NewCompactionBuffer(entStore, app.EventBus, sugar)
	app.CompactionBuffer = buf
	if app.compactionSync != nil {
		app.compactionSync.SetWaiter(buf)
	}

	// Register as a Buffer-priority lifecycle component so it participates in
	// Start/Stop with drain.
	app.registry.Register(lifecycle.NewFuncComponent("compaction-buffer",
		func(_ context.Context, wg *sync.WaitGroup) error {
			buf.Start(wg)
			return nil
		},
		func(_ context.Context) error {
			buf.Stop()
			return nil
		},
	), lifecycle.PriorityBuffer)

	// Post-turn trigger: estimate tokens, enqueue if over threshold.
	threshold := int(float64(modelWindow) * cfg.Threshold)
	if modelWindow <= 0 {
		threshold = 0
	}

	eventbus.SubscribeTyped(app.EventBus, func(e eventbus.TurnCompletedEvent) {
		if e.SessionKey == "" || threshold <= 0 {
			return
		}
		sess, err := entStore.Get(e.SessionKey)
		if err != nil || sess == nil {
			return
		}
		total := 0
		for _, m := range sess.History {
			total += 4 + types.EstimateTokens(m.Content)
		}
		if total <= threshold {
			return
		}
		// Compact the first half of history; leave the recent half intact.
		upTo := len(sess.History)/2 - 1
		if upTo < 0 {
			return
		}
		buf.EnqueueCompaction(e.SessionKey, upTo)
	})
}
