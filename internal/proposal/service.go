package proposal

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type proposalRegistry interface {
	Upsert(in UpsertInput) (*Proposal, error)
	GetByID(proposalID string) (Proposal, bool)
	MarkPreparing(proposalID string) (*Proposal, error)
	MarkPrepared(proposalID string, brief PreparedBrief) (*Proposal, error)
	Dismiss(proposalID string) (*Proposal, error)
	Accept(proposalID string) (*Proposal, error)
	RestorePrepared(proposalID string) (*Proposal, error)
	PruneExpired() int
}

type proposalPreparer interface {
	PrepareLearningSuggestion(source LearningSuggestionSource) (PreparedBrief, error)
}

// Service is the explicit Wave 3 write boundary for transient proposal lifecycle.
// It owns proposal creation/update, preparation lifecycle, dismissal, acceptance,
// and expiration pruning through the registry.
type Service struct {
	registry proposalRegistry
	preparer proposalPreparer
}

// NewService creates a proposal service backed by the supplied registry.
func NewService(registry proposalRegistry, preparer proposalPreparer) *Service {
	if preparer == nil {
		preparer = NewDeterministicPreparer()
	}
	return &Service{
		registry: registry,
		preparer: preparer,
	}
}

// UpsertLearningSuggestion creates or updates a proposal from the active Wave 3
// producer, then deterministically prepares it.
func (s *Service) UpsertLearningSuggestion(ctx context.Context, source LearningSuggestionSource) (*Proposal, error) {
	_ = ctx

	registry, err := s.requireRegistry("upsert learning suggestion")
	if err != nil {
		return nil, err
	}
	preparer, err := s.requirePreparer()
	if err != nil {
		return nil, err
	}

	registry.PruneExpired()

	proposal, err := registry.Upsert(UpsertInput{
		SessionKey: strings.TrimSpace(source.SessionKey),
		Source: ProposalSource{
			Kind: learningProposalSourceKind,
			Ref:  strings.TrimSpace(source.SuggestionID),
		},
		Title:      learningSuggestionTitle(source),
		Summary:    learningSuggestionSourceSummary(source),
		Reason:     learningSuggestionReason(source),
		Confidence: source.Confidence,
		ExpiresAt:  source.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}

	preparing, err := registry.MarkPreparing(proposal.ProposalID)
	if err != nil {
		return nil, err
	}

	brief, err := preparer.PrepareLearningSuggestion(source)
	if err != nil {
		return nil, err
	}

	prepared, err := registry.MarkPrepared(preparing.ProposalID, brief)
	if err != nil {
		return nil, err
	}
	return prepared, nil
}

// Dismiss hides a proposal through the registry write boundary.
func (s *Service) Dismiss(ctx context.Context, proposalID string) (*Proposal, error) {
	_ = ctx

	registry, err := s.requireRegistry("dismiss proposal")
	if err != nil {
		return nil, err
	}
	return registry.Dismiss(strings.TrimSpace(proposalID))
}

// Accept marks a prepared proposal as accepted through the registry write boundary.
func (s *Service) Accept(ctx context.Context, proposalID string) (*Proposal, error) {
	_ = ctx

	registry, err := s.requireRegistry("accept proposal")
	if err != nil {
		return nil, err
	}
	proposalID = strings.TrimSpace(proposalID)
	registry.PruneExpired()
	current, ok := registry.GetByID(proposalID)
	if !ok {
		return nil, fmt.Errorf("proposal %q not found", proposalID)
	}
	if current.Status != ProposalStatusPrepared {
		return nil, fmt.Errorf("accept proposal: proposal %q is %s", proposalID, current.Status)
	}
	return registry.Accept(proposalID)
}

// RestorePrepared rolls back a prior acceptance attempt when downstream durable
// mission creation fails and the transient proposal must become visible again.
func (s *Service) RestorePrepared(ctx context.Context, proposalID string) (*Proposal, error) {
	_ = ctx

	registry, err := s.requireRegistry("restore prepared proposal")
	if err != nil {
		return nil, err
	}
	return registry.RestorePrepared(strings.TrimSpace(proposalID))
}

// PruneExpired runs proposal expiration through the service boundary.
func (s *Service) PruneExpired(ctx context.Context) (int, error) {
	_ = ctx

	registry, err := s.requireRegistry("prune proposals")
	if err != nil {
		return 0, err
	}
	return registry.PruneExpired(), nil
}

func (s *Service) requireRegistry(action string) (proposalRegistry, error) {
	if s == nil || s.registry == nil {
		return nil, fmt.Errorf("%s: proposal registry is required", action)
	}
	return s.registry, nil
}

func (s *Service) requirePreparer() (proposalPreparer, error) {
	if s == nil || s.preparer == nil {
		return nil, fmt.Errorf("proposal preparer is required")
	}
	return s.preparer, nil
}

// RestorePrepared returns an accepted proposal back to prepared state so it
// becomes visible again after a downstream failure.
func (r *Registry) RestorePrepared(proposalID string) (*Proposal, error) {
	return r.transition(strings.TrimSpace(proposalID), func(p *Proposal, now time.Time) error {
		if p.Status != ProposalStatusAccepted {
			return fmt.Errorf("restore prepared: proposal %q is %s", p.ProposalID, p.Status)
		}
		p.Status = ProposalStatusPrepared
		p.UpdatedAt = now
		return nil
	})
}
