// Package ontologybridge connects the P2P protocol handler to the ontology service.
// It implements protocol.OntologyHandler, bridging schema exchange messages
// to OntologyService.ExportSchema/ImportSchema calls.
package ontologybridge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/ontology"
	"github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/p2p/reputation"
)

// Config holds exchange behavior settings.
type Config struct {
	MinTrustForSchema float64
	MinTrustForFacts  float64
	AutoImportMode    string // "shadow", "governed", "disabled"
	MaxTypesPerImport int
}

// DefaultConfig returns conservative defaults.
func DefaultConfig() Config {
	return Config{
		MinTrustForSchema: 0.5,
		MinTrustForFacts:  0.7,
		AutoImportMode:    "shadow",
		MaxTypesPerImport: 10,
	}
}

// Bridge implements protocol.OntologyHandler by delegating to OntologyService.
type Bridge struct {
	svc        ontology.OntologyService
	reputation *reputation.Store
	cfg        Config
}

// Ensure Bridge implements the interface at compile time.
var _ protocol.OntologyHandler = (*Bridge)(nil)

// New creates an OntologyBridge.
func New(svc ontology.OntologyService, rep *reputation.Store, cfg Config) *Bridge {
	return &Bridge{svc: svc, reputation: rep, cfg: cfg}
}

// HandleSchemaQuery serves a peer's request for the local schema bundle.
func (b *Bridge) HandleSchemaQuery(ctx context.Context, peerDID string, req protocol.SchemaQueryRequest) (*protocol.SchemaQueryResponse, error) {
	// Trust check
	if err := b.checkTrust(ctx, peerDID, b.cfg.MinTrustForSchema); err != nil {
		return nil, err
	}

	// Set peer principal for ACL
	ctx = ctxkeys.WithPrincipal(ctx, "peer:"+peerDID)

	bundle, err := b.svc.ExportSchema(ctx)
	if err != nil {
		return nil, fmt.Errorf("export schema: %w", err)
	}

	// Filter by requested types if specified
	if len(req.RequestedTypes) > 0 {
		bundle = filterBundle(bundle, req.RequestedTypes, req.IncludePredicates)
	}

	data, err := json.Marshal(bundle)
	if err != nil {
		return nil, fmt.Errorf("marshal bundle: %w", err)
	}

	return &protocol.SchemaQueryResponse{Bundle: data}, nil
}

// HandleSchemaPropose processes a peer's schema proposal for local import.
func (b *Bridge) HandleSchemaPropose(ctx context.Context, peerDID string, req protocol.SchemaProposeRequest) (*protocol.SchemaProposeResponse, error) {
	// Trust check
	if err := b.checkTrust(ctx, peerDID, b.cfg.MinTrustForSchema); err != nil {
		return nil, err
	}

	if b.cfg.AutoImportMode == "disabled" {
		return &protocol.SchemaProposeResponse{Action: "rejected", Rejected: []string{"exchange disabled"}}, nil
	}

	// Decode bundle
	var bundle ontology.SchemaBundle
	if err := json.Unmarshal(req.Bundle, &bundle); err != nil {
		return nil, fmt.Errorf("unmarshal bundle: %w", err)
	}

	// Enforce max types limit
	if b.cfg.MaxTypesPerImport > 0 && len(bundle.Types) > b.cfg.MaxTypesPerImport {
		return &protocol.SchemaProposeResponse{
			Action:   "rejected",
			Rejected: []string{fmt.Sprintf("too many types: %d > %d", len(bundle.Types), b.cfg.MaxTypesPerImport)},
		}, nil
	}

	// Determine import mode
	mode := ontology.ImportShadow
	if b.cfg.AutoImportMode == "governed" {
		mode = ontology.ImportGoverned
	}

	// Set peer principal for ACL
	ctx = ctxkeys.WithPrincipal(ctx, "peer:"+peerDID)

	result, err := b.svc.ImportSchema(ctx, &bundle, ontology.ImportOptions{
		Mode:      mode,
		SourceDID: peerDID,
	})
	if err != nil {
		return nil, fmt.Errorf("import schema: %w", err)
	}

	// Build response
	resultData, _ := json.Marshal(result)
	action := "accepted"
	if len(result.TypesConflicting) > 0 || len(result.PredsConflicting) > 0 {
		action = "partial"
	}
	if result.TypesAdded == 0 && result.PredsAdded == 0 {
		if len(result.TypesConflicting) > 0 || len(result.PredsConflicting) > 0 {
			action = "rejected"
		}
	}

	return &protocol.SchemaProposeResponse{
		Action:   action,
		Accepted: acceptedNames(result),
		Rejected: append(result.TypesConflicting, result.PredsConflicting...),
		Result:   resultData,
	}, nil
}

func (b *Bridge) checkTrust(ctx context.Context, peerDID string, minTrust float64) error {
	if b.reputation == nil {
		return nil // no reputation store = trust everyone
	}
	score, err := b.reputation.GetScore(ctx, peerDID)
	if err != nil {
		return nil // unknown peer = allow (first interaction)
	}
	if score < minTrust {
		return fmt.Errorf("peer %q trust %.2f < required %.2f", peerDID, score, minTrust)
	}
	return nil
}

func filterBundle(bundle *ontology.SchemaBundle, requestedTypes []string, includePredicates bool) *ontology.SchemaBundle {
	requested := make(map[string]bool, len(requestedTypes))
	for _, name := range requestedTypes {
		requested[name] = true
	}

	filtered := &ontology.SchemaBundle{
		Version:       bundle.Version,
		SchemaVersion: bundle.SchemaVersion,
		ExportedAt:    bundle.ExportedAt,
		ExportedBy:    bundle.ExportedBy,
		Digest:        bundle.Digest,
	}

	for _, t := range bundle.Types {
		if requested[t.Name] {
			filtered.Types = append(filtered.Types, t)
		}
	}

	if includePredicates {
		for _, p := range bundle.Predicates {
			if predicateRelated(p, requested) {
				filtered.Predicates = append(filtered.Predicates, p)
			}
		}
	}

	return filtered
}

func predicateRelated(p ontology.SchemaPredicateSlim, requestedTypes map[string]bool) bool {
	for _, st := range p.SourceTypes {
		if requestedTypes[st] {
			return true
		}
	}
	for _, tt := range p.TargetTypes {
		if requestedTypes[tt] {
			return true
		}
	}
	return len(p.SourceTypes) == 0 && len(p.TargetTypes) == 0
}

func acceptedNames(r *ontology.ImportResult) []string {
	// We don't track individual names in ImportResult, just counts.
	// Return nil — the Result JSON has full details.
	return nil
}
