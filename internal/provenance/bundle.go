package provenance

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"

	"github.com/langoai/lango/internal/security"
)

// AlgorithmSecp256k1Keccak256 re-exports the canonical algorithm constant for backward compatibility.
const AlgorithmSecp256k1Keccak256 = security.AlgorithmSecp256k1Keccak256

// BundleSigner signs a canonical provenance payload and declares its algorithm.
type BundleSigner interface {
	Sign(ctx context.Context, payload []byte) ([]byte, error)
	Algorithm() string
}

// PQBundleSigner is an optional interface that BundleSigners can implement
// to provide ML-DSA-65 post-quantum dual signatures.
type PQBundleSigner interface {
	SignPQ(ctx context.Context, payload []byte) ([]byte, error)
	PQAlgorithm() string
	PQPublicKey() []byte
}

// SignatureVerifyFunc verifies a signature against a signer DID.
// Implementations are injected at the app/cli wiring layer.
type SignatureVerifyFunc func(signerDID string, payload, signature []byte) error

// BundleService exports, verifies, and imports provenance bundles.
type BundleService struct {
	checkpoints  CheckpointStore
	treeStore    SessionTreeStore
	attributions AttributionStore
	attrService  *AttributionService
	verifiers    map[string]SignatureVerifyFunc
}

// NewBundleService creates a new provenance bundle service.
// verifiers maps algorithm names to verification functions; the provenance
// package does not contain any built-in verifier implementation.
func NewBundleService(
	checkpoints CheckpointStore,
	treeStore SessionTreeStore,
	attributions AttributionStore,
	attrService *AttributionService,
	verifiers map[string]SignatureVerifyFunc,
) *BundleService {
	return &BundleService{
		checkpoints:  checkpoints,
		treeStore:    treeStore,
		attributions: attributions,
		attrService:  attrService,
		verifiers:    verifiers,
	}
}

// Export builds a provenance bundle, applies redaction, and signs it.
func (s *BundleService) Export(
	ctx context.Context,
	sessionKey string,
	level RedactionLevel,
	signerDID string,
	signer BundleSigner,
) (*ProvenanceBundle, []byte, error) {
	if sessionKey == "" {
		return nil, nil, ErrInvalidSessionKey
	}
	if !level.Valid() {
		return nil, nil, fmt.Errorf("%w: %q", ErrInvalidRedaction, level)
	}
	if signerDID == "" {
		return nil, nil, fmt.Errorf("signer DID is required")
	}
	if signer == nil {
		return nil, nil, fmt.Errorf("bundle signer is required")
	}

	bundle, err := s.buildBundle(ctx, sessionKey, level)
	if err != nil {
		return nil, nil, err
	}
	bundle.SignerDID = signerDID
	bundle.SignatureAlgorithm = signer.Algorithm()

	// PQ dual-sign: embed PQ public key before computing canonical payload.
	// The classical signature covers PQSignerPublicKey (trust chain binding).
	if pqs, ok := signer.(PQBundleSigner); ok {
		bundle.PQSignerPublicKey = pqs.PQPublicKey()
		bundle.PQSignatureAlgorithm = pqs.PQAlgorithm()
	}

	payload, err := canonicalBundlePayload(bundle)
	if err != nil {
		return nil, nil, err
	}

	// Classical signature (required).
	sig, err := signer.Sign(ctx, payload)
	if err != nil {
		return nil, nil, fmt.Errorf("sign bundle: %w", err)
	}
	bundle.Signature = sig

	// PQ signature (optional, over the same canonical payload).
	if pqs, ok := signer.(PQBundleSigner); ok {
		pqSig, pqErr := pqs.SignPQ(ctx, payload)
		if pqErr != nil {
			return nil, nil, fmt.Errorf("PQ sign bundle: %w", pqErr)
		}
		bundle.PQSignature = pqSig
	}

	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("marshal bundle: %w", err)
	}
	return bundle, data, nil
}

// Verify validates the signer DID and signature of a bundle.
// The verifier is looked up from the map injected at construction time.
func (s *BundleService) Verify(bundle *ProvenanceBundle) error {
	if bundle == nil {
		return fmt.Errorf("nil provenance bundle")
	}
	if !bundle.RedactionLevel.Valid() {
		return fmt.Errorf("%w: %q", ErrInvalidRedaction, bundle.RedactionLevel)
	}
	if bundle.SignerDID == "" {
		return fmt.Errorf("bundle signer DID is required")
	}
	verifier, ok := s.verifiers[bundle.SignatureAlgorithm]
	if !ok {
		return fmt.Errorf("unsupported signature algorithm %q", bundle.SignatureAlgorithm)
	}
	if len(bundle.Signature) == 0 {
		return fmt.Errorf("bundle signature is required")
	}
	payload, err := canonicalBundlePayload(bundle)
	if err != nil {
		return err
	}
	// Classical signature verification (required).
	if err := verifier(bundle.SignerDID, payload, bundle.Signature); err != nil {
		return fmt.Errorf("verify bundle signature: %w", err)
	}
	// PQ signature verification (optional — backward compat with classical-only bundles).
	// Uses the embedded PQ public key directly (self-contained, rotation-safe).
	if len(bundle.PQSignature) > 0 && len(bundle.PQSignerPublicKey) > 0 {
		if err := security.VerifyMLDSA65(bundle.PQSignerPublicKey, payload, bundle.PQSignature); err != nil {
			return fmt.Errorf("verify PQ bundle signature: %w", err)
		}
	}
	return nil
}

// Import verifies a bundle and stores only provenance-owned records.
func (s *BundleService) Import(ctx context.Context, data []byte) (*ProvenanceBundle, error) {
	var bundle ProvenanceBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return nil, fmt.Errorf("unmarshal bundle: %w", err)
	}
	if err := s.Verify(&bundle); err != nil {
		return nil, err
	}

	for _, cp := range bundle.Checkpoints {
		if err := s.checkpoints.SaveCheckpoint(ctx, cp); err != nil {
			return nil, fmt.Errorf("import checkpoint %s: %w", cp.ID, err)
		}
	}
	for _, node := range bundle.SessionTree {
		if err := s.treeStore.SaveNode(ctx, node); err != nil {
			return nil, fmt.Errorf("import session node %s: %w", node.SessionKey, err)
		}
	}
	for _, attr := range bundle.Attributions {
		if err := s.attributions.SaveAttribution(ctx, attr); err != nil {
			return nil, fmt.Errorf("import attribution %s: %w", attr.ID, err)
		}
	}

	return &bundle, nil
}

func (s *BundleService) buildBundle(ctx context.Context, sessionKey string, level RedactionLevel) (*ProvenanceBundle, error) {
	var checkpoints []Checkpoint
	if s.checkpoints != nil {
		cps, err := s.checkpoints.ListBySession(ctx, sessionKey, 0)
		if err != nil {
			return nil, fmt.Errorf("list session checkpoints: %w", err)
		}
		checkpoints = SortedCheckpoints(cps)
	}

	var tree []SessionNode
	if s.treeStore != nil {
		service := NewSessionTree(s.treeStore)
		nodes, err := service.GetTree(ctx, sessionKey, 64)
		if err == nil {
			tree = SortedSessionNodes(nodes)
		} else if err != ErrSessionNotFound {
			return nil, fmt.Errorf("get session tree: %w", err)
		}
	}

	view, err := s.attrService.View(ctx, sessionKey, 0)
	if err != nil {
		return nil, err
	}
	report, err := s.attrService.Report(ctx, sessionKey)
	if err != nil {
		return nil, err
	}

	bundle := &ProvenanceBundle{
		Version:        "1",
		Checkpoints:    checkpoints,
		SessionTree:    tree,
		Attributions:   SortedAttributions(view.Attributions),
		Report:         report,
		RedactionLevel: level,
	}

	return redactBundle(bundle, level), nil
}

func canonicalBundlePayload(bundle *ProvenanceBundle) ([]byte, error) {
	clone := *bundle
	clone.Signature = nil
	clone.PQSignature = nil
	// PQSignerPublicKey and PQSignatureAlgorithm are INCLUDED —
	// the classical signature authenticates the embedded PQ public key.
	return json.Marshal(clone)
}

func redactBundle(bundle *ProvenanceBundle, level RedactionLevel) *ProvenanceBundle {
	clone := *bundle
	clone.Checkpoints = append([]Checkpoint(nil), bundle.Checkpoints...)
	clone.SessionTree = append([]SessionNode(nil), bundle.SessionTree...)
	clone.Attributions = append([]Attribution(nil), bundle.Attributions...)
	if bundle.Report != nil {
		cp := *bundle.Report
		cp.ByAuthor = maps.Clone(bundle.Report.ByAuthor)
		cp.ByFile = maps.Clone(bundle.Report.ByFile)
		clone.Report = &cp
	}

	switch level {
	case RedactionNone:
		return &clone
	case RedactionContent:
		for i := range clone.Checkpoints {
			clone.Checkpoints[i].Metadata = nil
			clone.Checkpoints[i].GitRef = ""
		}
		for i := range clone.SessionTree {
			clone.SessionTree[i].Goal = ""
		}
		for i := range clone.Attributions {
			clone.Attributions[i].FilePath = ""
			clone.Attributions[i].CommitHash = ""
			clone.Attributions[i].StepID = ""
		}
		if clone.Report != nil {
			clone.Report.ByFile = map[string]FileStats{}
		}
		return &clone
	case RedactionFull:
		clone = *redactBundle(&clone, RedactionContent)
		clone.Checkpoints = nil
		clone.SessionTree = nil
		clone.Attributions = nil
		if clone.Report != nil {
			clone.Report.ByAuthor = map[string]AuthorStats{}
			clone.Report.ByFile = map[string]FileStats{}
		}
		return &clone
	default:
		return redactBundle(bundle, RedactionContent)
	}
}

