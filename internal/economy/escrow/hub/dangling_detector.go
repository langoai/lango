package hub

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/eventbus"
)

// DanglingDetector periodically scans for escrows stuck in Pending status too long.
type DanglingDetector struct {
	store        escrow.Store
	engine       *escrow.Engine
	bus          *eventbus.Bus
	logger       *zap.SugaredLogger
	scanInterval time.Duration
	maxPending   time.Duration
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

// DanglingOption configures a DanglingDetector.
type DanglingOption func(*DanglingDetector)

// WithScanInterval sets the scan interval.
func WithScanInterval(d time.Duration) DanglingOption {
	return func(dd *DanglingDetector) {
		if d > 0 {
			dd.scanInterval = d
		}
	}
}

// WithMaxPending sets the maximum time an escrow can stay in Pending.
func WithMaxPending(d time.Duration) DanglingOption {
	return func(dd *DanglingDetector) {
		if d > 0 {
			dd.maxPending = d
		}
	}
}

// WithDanglingLogger sets a structured logger.
func WithDanglingLogger(l *zap.SugaredLogger) DanglingOption {
	return func(dd *DanglingDetector) {
		if l != nil {
			dd.logger = l
		}
	}
}

// NewDanglingDetector creates a new dangling escrow detector.
func NewDanglingDetector(store escrow.Store, engine *escrow.Engine, bus *eventbus.Bus, opts ...DanglingOption) *DanglingDetector {
	dd := &DanglingDetector{
		store:        store,
		engine:       engine,
		bus:          bus,
		logger:       zap.NewNop().Sugar(),
		scanInterval: 5 * time.Minute,
		maxPending:   10 * time.Minute,
		stopCh:       make(chan struct{}),
	}
	for _, o := range opts {
		o(dd)
	}
	return dd
}

// Name implements lifecycle.Component.
func (dd *DanglingDetector) Name() string { return "dangling-detector" }

// Start launches the periodic scan goroutine.
func (dd *DanglingDetector) Start(_ context.Context, wg *sync.WaitGroup) error {
	dd.wg.Add(1)
	go func() {
		defer dd.wg.Done()
		if wg != nil {
			wg.Done()
		}
		dd.run()
	}()
	dd.logger.Infow("dangling detector started", "scanInterval", dd.scanInterval, "maxPending", dd.maxPending)
	return nil
}

// Stop signals the detector to stop and waits for completion.
func (dd *DanglingDetector) Stop(_ context.Context) error {
	close(dd.stopCh)
	dd.wg.Wait()
	dd.logger.Info("dangling detector stopped")
	return nil
}

// run is the main loop.
func (dd *DanglingDetector) run() {
	ticker := time.NewTicker(dd.scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-dd.stopCh:
			return
		case <-ticker.C:
			dd.scan()
		}
	}
}

// scan iterates pending escrows and expires those stuck too long.
func (dd *DanglingDetector) scan() {
	entries := dd.store.ListByStatus(escrow.StatusPending)
	now := time.Now()

	for _, entry := range entries {
		if now.Sub(entry.CreatedAt) < dd.maxPending {
			continue
		}

		dd.logger.Warnw("dangling escrow detected",
			"escrowID", entry.ID,
			"buyerDID", entry.BuyerDID,
			"pendingSince", entry.CreatedAt,
			"age", now.Sub(entry.CreatedAt),
		)

		if _, err := dd.engine.Expire(context.Background(), entry.ID); err != nil {
			dd.logger.Warnw("expire dangling escrow", "escrowID", entry.ID, "error", err)
			continue
		}

		if dd.bus != nil {
			dd.bus.Publish(eventbus.EscrowDanglingEvent{
				EscrowID:     entry.ID,
				BuyerDID:     entry.BuyerDID,
				SellerDID:    entry.SellerDID,
				Amount:       entry.TotalAmount.String(),
				PendingSince: entry.CreatedAt,
				Action:       "expired",
			})
		}
	}
}
