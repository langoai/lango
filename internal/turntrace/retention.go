package turntrace

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/logging"
)

func retentionLogger() *zap.SugaredLogger { return logging.SubsystemSugar("turntrace.retention") }

// RetentionConfig controls automatic trace cleanup.
type RetentionConfig struct {
	MaxAge                time.Duration
	MaxTraces             int
	FailedTraceMultiplier int
	CleanupInterval       time.Duration
}

// RetentionCleaner periodically purges old traces.
type RetentionCleaner struct {
	store  Store
	config RetentionConfig
	stopCh chan struct{}
	doneCh chan struct{}
}

// NewRetentionCleaner creates a retention cleaner.
func NewRetentionCleaner(store Store, cfg RetentionConfig) *RetentionCleaner {
	if cfg.CleanupInterval <= 0 {
		cfg.CleanupInterval = time.Hour
	}
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = 30 * 24 * time.Hour // 30 days
	}
	if cfg.MaxTraces <= 0 {
		cfg.MaxTraces = 10000
	}
	if cfg.FailedTraceMultiplier <= 0 {
		cfg.FailedTraceMultiplier = 2
	}
	return &RetentionCleaner{
		store:  store,
		config: cfg,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

func (c *RetentionCleaner) Name() string { return "turntrace-retention" }

func (c *RetentionCleaner) Start(_ context.Context, _ *sync.WaitGroup) error {
	go c.run()
	return nil
}

func (c *RetentionCleaner) Stop(_ context.Context) error {
	close(c.stopCh)
	<-c.doneCh
	return nil
}

func (c *RetentionCleaner) run() {
	defer close(c.doneCh)
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.cleanup()
		}
	}
}

func (c *RetentionCleaner) cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log := retentionLogger()

	// 1. Purge old successful traces.
	successCutoff := time.Now().Add(-c.config.MaxAge)
	ids, err := c.store.OldTraces(ctx, successCutoff, true, 500)
	if err != nil {
		log.Warnw("query old successful traces", "error", err)
	} else if len(ids) > 0 {
		if err := c.store.PurgeTraces(ctx, ids); err != nil {
			log.Warnw("purge old successful traces", "count", len(ids), "error", err)
		} else {
			log.Infow("purged old successful traces", "count", len(ids))
		}
	}

	// 2. Purge old failed traces (retained longer).
	failedCutoff := time.Now().Add(-c.config.MaxAge * time.Duration(c.config.FailedTraceMultiplier))
	ids, err = c.store.OldTraces(ctx, failedCutoff, false, 500)
	if err != nil {
		log.Warnw("query old failed traces", "error", err)
	} else if len(ids) > 0 {
		if err := c.store.PurgeTraces(ctx, ids); err != nil {
			log.Warnw("purge old failed traces", "count", len(ids), "error", err)
		} else {
			log.Infow("purged old failed traces", "count", len(ids))
		}
	}

	// 3. Enforce max trace count.
	count, err := c.store.TraceCount(ctx)
	if err != nil {
		log.Warnw("query trace count", "error", err)
		return
	}
	if count > c.config.MaxTraces {
		excess := count - c.config.MaxTraces
		oldestCutoff := time.Now() // get oldest regardless of age
		ids, err := c.store.OldTraces(ctx, oldestCutoff, false, excess)
		if err != nil {
			log.Warnw("query excess traces", "error", err)
		} else if len(ids) > 0 {
			if err := c.store.PurgeTraces(ctx, ids); err != nil {
				log.Warnw("purge excess traces", "count", len(ids), "error", err)
			} else {
				log.Infow("purged excess traces", "count", len(ids), "total", count)
			}
		}
	}
}
