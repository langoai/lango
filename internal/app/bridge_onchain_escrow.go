package app

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/eventbus"
)

// escrowTransitionFunc matches the signature of most escrow.Engine transition methods.
type escrowTransitionFunc func(context.Context, string) (*escrow.EscrowEntry, error)

// tryEscrowTransition calls fn and logs the result idempotently.
// If the escrow is already in the target state, it logs at debug level.
func tryEscrowTransition(ctx context.Context, log *zap.SugaredLogger, escrowID, label string, fn escrowTransitionFunc) {
	if _, err := fn(ctx, escrowID); err != nil {
		if isAlreadyTransitioned(err) {
			log.Debugw(label+": already transitioned", "escrowID", escrowID)
		} else {
			log.Warnw(label, "escrowID", escrowID, "error", err)
		}
	}
}

// initOnChainEscrowBridge wires on-chain events to escrow engine state transitions.
// All transitions are idempotent (check current state before transitioning).
func initOnChainEscrowBridge(bus *eventbus.Bus, engine *escrow.Engine, log *zap.SugaredLogger) {
	// Deposit → Fund then Activate.
	eventbus.SubscribeTyped(bus, func(ev eventbus.EscrowOnChainDepositEvent) {
		if ev.EscrowID == "" {
			log.Debugw("deposit event without escrow ID", "dealID", ev.DealID)
			return
		}
		ctx := context.Background()
		tryEscrowTransition(ctx, log, ev.EscrowID, "deposit: fund", engine.Fund)
		tryEscrowTransition(ctx, log, ev.EscrowID, "deposit: activate", engine.Activate)
	})

	// Release → Release.
	eventbus.SubscribeTyped(bus, func(ev eventbus.EscrowOnChainReleaseEvent) {
		if ev.EscrowID == "" {
			return
		}
		tryEscrowTransition(context.Background(), log, ev.EscrowID, "release", engine.Release)
	})

	// Refund → Refund.
	eventbus.SubscribeTyped(bus, func(ev eventbus.EscrowOnChainRefundEvent) {
		if ev.EscrowID == "" {
			return
		}
		tryEscrowTransition(context.Background(), log, ev.EscrowID, "refund", engine.Refund)
	})

	// Dispute → Dispute.
	eventbus.SubscribeTyped(bus, func(ev eventbus.EscrowOnChainDisputeEvent) {
		if ev.EscrowID == "" {
			return
		}
		disputeFn := func(ctx context.Context, id string) (*escrow.EscrowEntry, error) {
			return engine.Dispute(ctx, id, "on-chain dispute")
		}
		tryEscrowTransition(context.Background(), log, ev.EscrowID, "dispute", disputeFn)
	})

	// Resolved → Release or Refund based on SellerFavor.
	eventbus.SubscribeTyped(bus, func(ev eventbus.EscrowOnChainResolvedEvent) {
		if ev.EscrowID == "" {
			return
		}
		ctx := context.Background()
		if ev.SellerFavor {
			tryEscrowTransition(ctx, log, ev.EscrowID, "resolved: release", engine.Release)
		} else {
			tryEscrowTransition(ctx, log, ev.EscrowID, "resolved: refund", engine.Refund)
		}
	})

	// Reorg detection alert.
	eventbus.SubscribeTyped(bus, func(ev eventbus.EscrowReorgDetectedEvent) {
		if ev.ExceedsDepth {
			log.Errorw("CRITICAL: deep reorg exceeds confirmation depth",
				"previousBlock", ev.PreviousBlock, "newBlock", ev.NewBlock,
				"depth", ev.Depth)
		} else {
			log.Warnw("reorg detected, rolled back to safe block",
				"previousBlock", ev.PreviousBlock, "newBlock", ev.NewBlock,
				"depth", ev.Depth)
		}
	})

	log.Info("on-chain escrow bridge initialized")
}

// isAlreadyTransitioned returns true if the error indicates the escrow
// is already in the target state (invalid transition).
func isAlreadyTransitioned(err error) bool {
	return errors.Is(err, escrow.ErrInvalidTransition)
}
