package retrieval

import (
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/knowledge"
)

// CompareShadowResults logs comparison metrics between old retrieval path and new coordinator findings.
func CompareShadowResults(old *knowledge.RetrievalResult, new []Finding, logger *zap.SugaredLogger) {
	if logger == nil {
		return
	}

	oldSet := make(map[dedupKey]struct{})
	if old != nil {
		for layer, items := range old.Items {
			for _, item := range items {
				oldSet[dedupKey{Layer: layer, Key: item.Key}] = struct{}{}
			}
		}
	}

	newSet := make(map[dedupKey]struct{}, len(new))
	for _, f := range new {
		newSet[dedupKey{Layer: f.Layer, Key: f.Key}] = struct{}{}
	}

	var overlap, oldOnly, newOnly int
	for k := range oldSet {
		if _, ok := newSet[k]; ok {
			overlap++
		} else {
			oldOnly++
		}
	}
	for k := range newSet {
		if _, ok := oldSet[k]; !ok {
			newOnly++
		}
	}

	logger.Infow("shadow comparison",
		"overlap", overlap,
		"old_only", oldOnly,
		"new_only", newOnly,
		"total_old", len(oldSet),
		"total_new", len(newSet),
	)
}
