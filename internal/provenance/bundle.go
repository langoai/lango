package provenance

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"

	"github.com/langoai/lango/internal/p2p/identity"
)

const signatureAlgorithmSecp256k1Keccak256 = "secp256k1-keccak256"

// BundleSignFunc signs a canonical provenance payload.
type BundleSignFunc func(ctx context.Context, payload []byte) ([]byte, error)

// BundleService exports, verifies, and imports provenance bundles.
type BundleService struct {
	checkpoints  CheckpointStore
	treeStore    SessionTreeStore
	attributions AttributionStore
	attrService  *AttributionService
}

// NewBundleService creates a new provenance bundle service.
func NewBundleService(
	checkpoints CheckpointStore,
	treeStore SessionTreeStore,
	attributions AttributionStore,
	attrService *AttributionService,
) *BundleService {
	return &BundleService{
		checkpoints:  checkpoints,
		treeStore:    treeStore,
		attributions: attributions,
		attrService:  attrService,
	}
}

// Export builds a provenance bundle, applies redaction, and signs it.
func (s *BundleService) Export(
	ctx context.Context,
	sessionKey string,
	level RedactionLevel,
	signerDID string,
	signFn BundleSignFunc,
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
	if signFn == nil {
		return nil, nil, fmt.Errorf("bundle signer is required")
	}

	bundle, err := s.buildBundle(ctx, sessionKey, level)
	if err != nil {
		return nil, nil, err
	}
	bundle.SignerDID = signerDID
	bundle.SignatureAlgorithm = signatureAlgorithmSecp256k1Keccak256

	payload, err := canonicalBundlePayload(bundle)
	if err != nil {
		return nil, nil, err
	}
	sig, err := signFn(ctx, payload)
	if err != nil {
		return nil, nil, fmt.Errorf("sign bundle: %w", err)
	}
	bundle.Signature = sig

	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("marshal bundle: %w", err)
	}
	return bundle, data, nil
}

// Verify validates the signer DID and signature of a bundle.
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
	if bundle.SignatureAlgorithm != signatureAlgorithmSecp256k1Keccak256 {
		return fmt.Errorf("unsupported signature algorithm %q", bundle.SignatureAlgorithm)
	}
	if len(bundle.Signature) == 0 {
		return fmt.Errorf("bundle signature is required")
	}
	payload, err := canonicalBundlePayload(bundle)
	if err != nil {
		return err
	}
	if err := identity.VerifyMessageSignature(bundle.SignerDID, payload, bundle.Signature); err != nil {
		return fmt.Errorf("verify bundle signature: %w", err)
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

