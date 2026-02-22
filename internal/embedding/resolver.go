package embedding

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/memory"
)

// StoreResolver resolves content from knowledge and memory stores.
type StoreResolver struct {
	knowledgeStore *knowledge.Store
	memoryStore    *memory.Store
}

// NewStoreResolver creates a content resolver backed by the application stores.
func NewStoreResolver(ks *knowledge.Store, ms *memory.Store) *StoreResolver {
	return &StoreResolver{
		knowledgeStore: ks,
		memoryStore:    ms,
	}
}

// ResolveContent looks up the original text for a given collection and ID.
func (r *StoreResolver) ResolveContent(ctx context.Context, collection, id string) (string, error) {
	switch collection {
	case "knowledge":
		if r.knowledgeStore == nil {
			return "", fmt.Errorf("knowledge store not available")
		}
		entry, err := r.knowledgeStore.GetKnowledge(ctx, id)
		if err != nil {
			return "", err
		}
		return entry.Content, nil

	case "observation":
		if r.memoryStore == nil {
			return "", fmt.Errorf("memory store not available")
		}
		uid, err := uuid.Parse(id)
		if err != nil {
			return "", fmt.Errorf("parse observation id: %w", err)
		}
		obs, err := r.memoryStore.GetObservation(ctx, uid)
		if err != nil {
			return "", err
		}
		return obs.Content, nil

	case "reflection":
		if r.memoryStore == nil {
			return "", fmt.Errorf("memory store not available")
		}
		uid, err := uuid.Parse(id)
		if err != nil {
			return "", fmt.Errorf("parse reflection id: %w", err)
		}
		ref, err := r.memoryStore.GetReflection(ctx, uid)
		if err != nil {
			return "", err
		}
		return ref.Content, nil

	case "learning":
		if r.knowledgeStore == nil {
			return "", fmt.Errorf("knowledge store not available")
		}
		uid, err := uuid.Parse(id)
		if err != nil {
			return "", fmt.Errorf("parse learning id: %w", err)
		}
		entry, err := r.knowledgeStore.GetLearning(ctx, uid)
		if err != nil {
			return "", err
		}
		return entry.Trigger + "\n" + entry.Fix, nil

	default:
		return "", fmt.Errorf("unknown collection: %s", collection)
	}
}
