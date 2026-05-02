package proposal

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"
)

type Registry struct {
	mu        sync.RWMutex
	nowFn     func() time.Time
	nextID    int
	byID      map[string]*Proposal
	bySource  map[string]string
	bySession map[string]map[string]struct{}
}

func NewRegistry(nowFn func() time.Time) *Registry {
	if nowFn == nil {
		nowFn = time.Now
	}
	return &Registry{
		nowFn:     nowFn,
		byID:      make(map[string]*Proposal),
		bySource:  make(map[string]string),
		bySession: make(map[string]map[string]struct{}),
	}
}

func (r *Registry) Upsert(in UpsertInput) (*Proposal, error) {
	sessionKey := strings.TrimSpace(in.SessionKey)
	if sessionKey == "" {
		return nil, fmt.Errorf("upsert proposal: session_key is required")
	}
	sourceKind := strings.TrimSpace(in.Source.Kind)
	sourceRef := strings.TrimSpace(in.Source.Ref)
	if sourceKind == "" || sourceRef == "" {
		return nil, fmt.Errorf("upsert proposal: source identity is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.nowFn()
	sourceKey := stableSourceKey(sessionKey, ProposalSource{Kind: sourceKind, Ref: sourceRef})
	if id, ok := r.bySource[sourceKey]; ok {
		existing := r.byID[id]
		existing.Title = strings.TrimSpace(in.Title)
		existing.Summary = strings.TrimSpace(in.Summary)
		existing.Reason = strings.TrimSpace(in.Reason)
		existing.Confidence = in.Confidence
		if !in.ExpiresAt.IsZero() {
			existing.ExpiresAt = in.ExpiresAt
		}
		existing.UpdatedAt = now
		return cloneProposal(existing), nil
	}

	r.nextID++
	id := fmt.Sprintf("proposal-%d", r.nextID)
	proposal := &Proposal{
		ProposalID: id,
		SessionKey: sessionKey,
		Source: ProposalSource{
			Kind: sourceKind,
			Ref:  sourceRef,
		},
		Title:      strings.TrimSpace(in.Title),
		Summary:    strings.TrimSpace(in.Summary),
		Reason:     strings.TrimSpace(in.Reason),
		Confidence: in.Confidence,
		Status:     ProposalStatusSuggested,
		CreatedAt:  now,
		UpdatedAt:  now,
		ExpiresAt:  in.ExpiresAt,
	}
	r.byID[id] = proposal
	r.bySource[sourceKey] = id
	if _, ok := r.bySession[sessionKey]; !ok {
		r.bySession[sessionKey] = make(map[string]struct{})
	}
	r.bySession[sessionKey][id] = struct{}{}
	return cloneProposal(proposal), nil
}

func (r *Registry) GetByID(proposalID string) (Proposal, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	proposal, ok := r.byID[strings.TrimSpace(proposalID)]
	if !ok {
		return Proposal{}, false
	}
	return *cloneProposal(proposal), true
}

func (r *Registry) ListBySession(sessionKey string) []Proposal {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.bySession[strings.TrimSpace(sessionKey)]
	if len(ids) == 0 {
		return nil
	}

	now := r.nowFn()
	items := make([]Proposal, 0, len(ids))
	for id := range ids {
		proposal := r.byID[id]
		if proposal == nil || !isVisibleActiveStatus(proposal.Status) || isExpiredAt(proposal, now) {
			continue
		}
		items = append(items, *cloneProposal(proposal))
	}

	slices.SortFunc(items, func(a, b Proposal) int {
		switch {
		case a.UpdatedAt.After(b.UpdatedAt):
			return -1
		case a.UpdatedAt.Before(b.UpdatedAt):
			return 1
		case a.CreatedAt.After(b.CreatedAt):
			return -1
		case a.CreatedAt.Before(b.CreatedAt):
			return 1
		case a.ProposalID < b.ProposalID:
			return -1
		case a.ProposalID > b.ProposalID:
			return 1
		default:
			return 0
		}
	})

	return items
}

func (r *Registry) MarkPreparing(proposalID string) (*Proposal, error) {
	return r.transition(strings.TrimSpace(proposalID), func(p *Proposal, now time.Time) error {
		if isTerminalStatus(p.Status) {
			return fmt.Errorf("mark preparing: proposal %q is terminal", p.ProposalID)
		}
		p.Status = ProposalStatusPreparing
		p.UpdatedAt = now
		return nil
	})
}

func (r *Registry) MarkPrepared(proposalID string, brief PreparedBrief) (*Proposal, error) {
	return r.transition(strings.TrimSpace(proposalID), func(p *Proposal, now time.Time) error {
		if isTerminalStatus(p.Status) {
			return fmt.Errorf("mark prepared: proposal %q is terminal", p.ProposalID)
		}
		p.Status = ProposalStatusPrepared
		p.PreparedBrief = clonePreparedBrief(&brief)
		p.UpdatedAt = now
		return nil
	})
}

func (r *Registry) Dismiss(proposalID string) (*Proposal, error) {
	return r.transition(strings.TrimSpace(proposalID), func(p *Proposal, now time.Time) error {
		if isTerminalStatus(p.Status) {
			return fmt.Errorf("dismiss proposal: proposal %q is terminal", p.ProposalID)
		}
		p.Status = ProposalStatusDismissed
		p.UpdatedAt = now
		return nil
	})
}

func (r *Registry) Accept(proposalID string) (*Proposal, error) {
	return r.transition(strings.TrimSpace(proposalID), func(p *Proposal, now time.Time) error {
		if isTerminalStatus(p.Status) {
			return fmt.Errorf("accept proposal: proposal %q is terminal", p.ProposalID)
		}
		p.Status = ProposalStatusAccepted
		p.UpdatedAt = now
		return nil
	})
}

func (r *Registry) PruneExpired() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.nowFn()
	expired := 0
	for _, proposal := range r.byID {
		if proposal == nil || isTerminalStatus(proposal.Status) {
			continue
		}
		if proposal.ExpiresAt.IsZero() || proposal.ExpiresAt.After(now) {
			continue
		}
		proposal.Status = ProposalStatusExpired
		proposal.UpdatedAt = now
		expired++
	}
	return expired
}

func (r *Registry) transition(proposalID string, fn func(*Proposal, time.Time) error) (*Proposal, error) {
	if proposalID == "" {
		return nil, fmt.Errorf("proposal id is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	proposal, ok := r.byID[proposalID]
	if !ok {
		return nil, fmt.Errorf("proposal %q not found", proposalID)
	}
	if err := fn(proposal, r.nowFn()); err != nil {
		return nil, err
	}
	return cloneProposal(proposal), nil
}

func stableSourceKey(sessionKey string, source ProposalSource) string {
	return sessionKey + "|" + strings.TrimSpace(source.Kind) + "|" + strings.TrimSpace(source.Ref)
}

func isVisibleActiveStatus(status ProposalStatus) bool {
	switch status {
	case ProposalStatusSuggested, ProposalStatusPreparing, ProposalStatusPrepared:
		return true
	default:
		return false
	}
}

func isTerminalStatus(status ProposalStatus) bool {
	switch status {
	case ProposalStatusDismissed, ProposalStatusAccepted, ProposalStatusExpired:
		return true
	default:
		return false
	}
}

func isExpiredAt(proposal *Proposal, now time.Time) bool {
	if proposal == nil || proposal.ExpiresAt.IsZero() {
		return false
	}
	return !proposal.ExpiresAt.After(now)
}

func cloneProposal(p *Proposal) *Proposal {
	if p == nil {
		return nil
	}
	copyProposal := *p
	copyProposal.Source = ProposalSource{
		Kind: p.Source.Kind,
		Ref:  p.Source.Ref,
	}
	copyProposal.PreparedBrief = clonePreparedBrief(p.PreparedBrief)
	return &copyProposal
}

func clonePreparedBrief(brief *PreparedBrief) *PreparedBrief {
	if brief == nil {
		return nil
	}
	copyBrief := *brief
	copyBrief.SupportingEvidence = append([]string(nil), brief.SupportingEvidence...)
	return &copyBrief
}
