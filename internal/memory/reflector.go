package memory

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const reflectorPrompt = `You are a conversation memory assistant. Your task is to condense these observation notes into a single, comprehensive summary.

Merge overlapping information, resolve any contradictions (prefer later observations), and create a coherent narrative of the conversation so far.

Write a concise summary (3-8 sentences) that captures all essential context.`

// Reflector condenses accumulated observations into reflections.
type Reflector struct {
	generator TextGenerator
	store     *Store
	logger    *zap.SugaredLogger
}

// NewReflector creates a new Reflector.
func NewReflector(generator TextGenerator, store *Store, logger *zap.SugaredLogger) *Reflector {
	return &Reflector{
		generator: generator,
		store:     store,
		logger:    logger,
	}
}

// Reflect condenses all observations for a session into a single reflection.
// Returns nil if there are no observations to condense.
func (r *Reflector) Reflect(ctx context.Context, sessionKey string) (*Reflection, error) {
	observations, err := r.store.ListObservations(ctx, sessionKey)
	if err != nil {
		return nil, fmt.Errorf("list observations: %w", err)
	}

	if len(observations) == 0 {
		return nil, nil
	}

	userPrompt := formatObservations(observations)

	response, err := r.generator.GenerateText(ctx, reflectorPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("generate reflection: %w", err)
	}

	ref := Reflection{
		ID:         uuid.New(),
		SessionKey: sessionKey,
		Content:    response,
		TokenCount: EstimateTokens(response),
		Generation: 1,
		CreatedAt:  time.Now(),
	}

	if err := r.store.SaveReflection(ctx, ref); err != nil {
		return nil, fmt.Errorf("save reflection: %w", err)
	}

	// Delete the condensed observations.
	ids := make([]uuid.UUID, 0, len(observations))
	for _, obs := range observations {
		ids = append(ids, obs.ID)
	}
	if err := r.store.DeleteObservations(ctx, ids); err != nil {
		return nil, fmt.Errorf("delete condensed observations: %w", err)
	}

	r.logger.Debugw("reflection created",
		"sessionKey", sessionKey,
		"generation", ref.Generation,
		"condensedObservations", len(observations),
		"tokens", ref.TokenCount,
	)

	return &ref, nil
}

// ReflectOnReflections condenses accumulated reflections into a higher-generation reflection.
// Returns nil if there are no reflections to condense.
func (r *Reflector) ReflectOnReflections(ctx context.Context, sessionKey string) (*Reflection, error) {
	reflections, err := r.store.ListReflections(ctx, sessionKey)
	if err != nil {
		return nil, fmt.Errorf("list reflections: %w", err)
	}

	if len(reflections) == 0 {
		return nil, nil
	}

	// Determine the next generation number.
	maxGen := 0
	for _, ref := range reflections {
		if ref.Generation > maxGen {
			maxGen = ref.Generation
		}
	}

	userPrompt := formatReflections(reflections)

	response, err := r.generator.GenerateText(ctx, reflectorPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("generate meta-reflection: %w", err)
	}

	ref := Reflection{
		ID:         uuid.New(),
		SessionKey: sessionKey,
		Content:    response,
		TokenCount: EstimateTokens(response),
		Generation: maxGen + 1,
		CreatedAt:  time.Now(),
	}

	if err := r.store.SaveReflection(ctx, ref); err != nil {
		return nil, fmt.Errorf("save meta-reflection: %w", err)
	}

	// Delete the condensed reflections.
	ids := make([]uuid.UUID, 0, len(reflections))
	for _, old := range reflections {
		ids = append(ids, old.ID)
	}
	if err := r.store.DeleteReflections(ctx, ids); err != nil {
		return nil, fmt.Errorf("delete condensed reflections: %w", err)
	}

	r.logger.Debugw("meta-reflection created",
		"sessionKey", sessionKey,
		"generation", ref.Generation,
		"condensedReflections", len(reflections),
		"tokens", ref.TokenCount,
	)

	return &ref, nil
}

// formatObservations formats observations into a text block for the reflector prompt.
func formatObservations(observations []Observation) string {
	var b strings.Builder
	for i, obs := range observations {
		fmt.Fprintf(&b, "Observation %d:\n%s\n\n", i+1, obs.Content)
	}
	return b.String()
}

// formatReflections formats reflections into a text block for the meta-reflector prompt.
func formatReflections(reflections []Reflection) string {
	var b strings.Builder
	for i, ref := range reflections {
		fmt.Fprintf(&b, "Reflection %d (gen %d):\n%s\n\n", i+1, ref.Generation, ref.Content)
	}
	return b.String()
}
