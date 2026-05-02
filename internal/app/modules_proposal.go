package app

import (
	"context"
	"time"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/proposal"
)

const proposalTTL = 30 * time.Minute

type proposalValues struct {
	registry *proposal.Registry
	preparer proposal.Preparer
	service  *proposal.Service
}

type proposalModule struct {
	bus *eventbus.Bus
}

func (m *proposalModule) Name() string { return "proposal" }

func (m *proposalModule) Provides() []appinit.Provides {
	return []appinit.Provides{appinit.ProvidesProposal}
}

func (m *proposalModule) DependsOn() []appinit.Provides { return nil }

func (m *proposalModule) Enabled() bool { return m != nil && m.bus != nil }

func (m *proposalModule) Init(_ context.Context, _ appinit.Resolver) (*appinit.ModuleResult, error) {
	registry := proposal.NewRegistry(time.Now)
	preparer := proposal.NewDeterministicPreparer()
	service := proposal.NewService(registry, preparer)

	eventbus.SubscribeTyped(m.bus, func(e eventbus.LearningSuggestionEvent) {
		expiresAt := e.Timestamp.Add(proposalTTL)
		if e.Timestamp.IsZero() {
			expiresAt = time.Now().Add(proposalTTL)
		}
		if _, err := service.UpsertLearningSuggestion(context.Background(), proposal.LearningSuggestionSource{
			SessionKey:   e.SessionKey,
			SuggestionID: e.SuggestionID,
			Pattern:      e.Pattern,
			ProposedRule: e.ProposedRule,
			Confidence:   e.Confidence,
			Rationale:    e.Rationale,
			ExpiresAt:    expiresAt,
		}); err != nil {
			logger().Warnw("upsert proposal from learning suggestion", "suggestion_id", e.SuggestionID, "error", err)
		}
	})

	return &appinit.ModuleResult{
		Values: map[appinit.Provides]interface{}{
			appinit.ProvidesProposal: &proposalValues{
				registry: registry,
				preparer: preparer,
				service:  service,
			},
		},
	}, nil
}
