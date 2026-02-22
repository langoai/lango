package memory

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/graph"
)

// TripleCallback is a hook that receives graph triples for asynchronous storage.
type TripleCallback func(triples []graph.Triple)

// GraphHooks adds graph relationship tracking to the memory system.
// It generates triples for temporal ordering, session membership,
// and reflection-observation links.
type GraphHooks struct {
	callback TripleCallback
	logger   *zap.SugaredLogger
}

// NewGraphHooks creates a new graph hooks instance.
func NewGraphHooks(callback TripleCallback, logger *zap.SugaredLogger) *GraphHooks {
	return &GraphHooks{
		callback: callback,
		logger:   logger,
	}
}

// OnObservation generates graph triples when an observation is saved.
// Triples created:
//   - observation:ID --[in_session]--> session:SessionKey
//   - observation:ID --[follows]--> observation:PreviousID (if provided)
func (h *GraphHooks) OnObservation(obs Observation, previousObsID string) {
	if h.callback == nil {
		return
	}

	obsNode := fmt.Sprintf("observation:%s", obs.ID.String())
	sessionNode := fmt.Sprintf("session:%s", obs.SessionKey)

	triples := []graph.Triple{
		{
			Subject:   obsNode,
			Predicate: graph.InSession,
			Object:    sessionNode,
		},
	}

	if previousObsID != "" {
		triples = append(triples, graph.Triple{
			Subject:   obsNode,
			Predicate: graph.Follows,
			Object:    fmt.Sprintf("observation:%s", previousObsID),
		})
	}

	h.callback(triples)
}

// OnReflection generates graph triples when a reflection is saved.
// Triples created:
//   - reflection:ID --[in_session]--> session:SessionKey
//   - reflection:ID --[reflects_on]--> observation:ObsID (for each related observation)
func (h *GraphHooks) OnReflection(ref Reflection, observationIDs []string) {
	if h.callback == nil {
		return
	}

	refNode := fmt.Sprintf("reflection:%s", ref.ID.String())
	sessionNode := fmt.Sprintf("session:%s", ref.SessionKey)

	triples := []graph.Triple{
		{
			Subject:   refNode,
			Predicate: graph.InSession,
			Object:    sessionNode,
		},
	}

	for _, obsID := range observationIDs {
		triples = append(triples, graph.Triple{
			Subject:   refNode,
			Predicate: graph.ReflectsOn,
			Object:    fmt.Sprintf("observation:%s", obsID),
		})
	}

	h.callback(triples)
}
