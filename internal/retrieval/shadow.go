package retrieval

import (
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/knowledge"
)

// factualLayers identifies layers that are covered by the factual retrieval path
// (FactSearchAgent). Used to separate factual overlap from new-context findings
// in shadow comparison.
var factualLayers = map[knowledge.ContextLayer]bool{
	knowledge.LayerUserKnowledge:    true,
	knowledge.LayerAgentLearnings:   true,
	knowledge.LayerExternalKnowledge: true,
}

// CompareShadowResults logs comparison metrics between old retrieval path and new coordinator findings.
// Logs both overall metrics and factual-layer-only metrics to prevent structural dilution
// when new agents (e.g., ContextSearchAgent) add items from non-factual layers.
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
	var factualOverlap, factualOldOnly, factualNewOnly int

	for k := range oldSet {
		if _, ok := newSet[k]; ok {
			overlap++
			if factualLayers[k.Layer] {
				factualOverlap++
			}
		} else {
			oldOnly++
			if factualLayers[k.Layer] {
				factualOldOnly++
			}
		}
	}
	for k := range newSet {
		if _, ok := oldSet[k]; !ok {
			newOnly++
			if factualLayers[k.Layer] {
				factualNewOnly++
			}
		}
	}

	logger.Infow("shadow comparison",
		"overlap", overlap,
		"old_only", oldOnly,
		"new_only", newOnly,
		"total_old", len(oldSet),
		"total_new", len(newSet),
		"factual_overlap", factualOverlap,
		"factual_old_only", factualOldOnly,
		"factual_new_only", factualNewOnly,
	)
}
