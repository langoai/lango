package graph

import (
	"github.com/langoai/lango/internal/eventbus"
	"go.uber.org/zap"
)

type AdmissionSource string

const (
	AdmissionSourceConversationAnalysis  AdmissionSource = "conversation_analysis"
	AdmissionSourceSessionLearning       AdmissionSource = "session_learning"
	AdmissionSourceLearning              AdmissionSource = "learning"
	AdmissionSourceProactiveLibrarian    AdmissionSource = "proactive_librarian"
	AdmissionSourceContentSavedExtractor AdmissionSource = "content_saved_extractor"
)

type AdmissionProducerGroup string

const (
	AdmissionProducerGroupLearning  AdmissionProducerGroup = "learning"
	AdmissionProducerGroupLibrarian AdmissionProducerGroup = "librarian"
)

type AdmissionValidatorSource string

const (
	AdmissionValidatorSourceOntologyRegistry AdmissionValidatorSource = "ontology_registry"
	AdmissionValidatorSourceUnavailable      AdmissionValidatorSource = "unavailable"
)

type AdmissionDecision string

const (
	AdmissionDecisionKnown       AdmissionDecision = "known"
	AdmissionDecisionUnknown     AdmissionDecision = "unknown"
	AdmissionDecisionUnvalidated AdmissionDecision = "unvalidated"
)

type AdmissionSourceKind string

const (
	AdmissionSourceKindEventBus  AdmissionSourceKind = "event_bus"
	AdmissionSourceKindSynthetic AdmissionSourceKind = "synthetic"
)

type AdmissionBatch struct {
	SourceKind    AdmissionSourceKind
	Source        AdmissionSource
	ProducerGroup AdmissionProducerGroup
	Triples       []Triple
}

type AdmissionRecord struct {
	Triple   Triple
	Decision AdmissionDecision
}

type AdmissionObserveResult struct {
	Records       []AdmissionRecord
	Forwarded     []Triple
	Event         *eventbus.GraphAdmissionBatchEvent
	UnmappedEvent *eventbus.GraphAdmissionUnmappedSourceEvent
}

type AdmissionPolicyConfig struct {
	Validator       PredicateValidatorFunc
	ValidatorSource AdmissionValidatorSource
}

type AdmissionPolicy struct {
	cfg    AdmissionPolicyConfig
	logger *zap.SugaredLogger
}

func NewAdmissionPolicy(cfg AdmissionPolicyConfig, logger *zap.SugaredLogger) *AdmissionPolicy {
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}
	return &AdmissionPolicy{
		cfg:    cfg,
		logger: logger,
	}
}

func (p *AdmissionPolicy) ObserveBatch(batch AdmissionBatch) AdmissionObserveResult {
	result := AdmissionObserveResult{
		Forwarded: make([]Triple, 0, len(batch.Triples)),
	}
	for _, triple := range batch.Triples {
		result.Forwarded = append(result.Forwarded, triple)
	}

	sourceKind := normalizeAdmissionSourceKind(batch)

	if !isObservedAdmissionSource(sourceKind, batch.Source) {
		if sourceKind == AdmissionSourceKindEventBus {
			result.UnmappedEvent = &eventbus.GraphAdmissionUnmappedSourceEvent{
				RawSource:  string(batch.Source),
				BatchCount: 1,
			}
		}
		return result
	}

	result.Records = make([]AdmissionRecord, 0, len(batch.Triples))
	result.Event = newAdmissionBatchEvent(batch)

	validatorSource := p.cfg.ValidatorSource
	if p.cfg.Validator == nil {
		validatorSource = AdmissionValidatorSourceUnavailable
	} else if validatorSource == "" {
		validatorSource = AdmissionValidatorSourceOntologyRegistry
	}
	result.Event.ValidatorSource = string(validatorSource)

	for _, triple := range batch.Triples {
		decision := AdmissionDecisionUnvalidated

		switch {
		case p.cfg.Validator == nil:
			result.Event.UnvalidatedCount++
		case p.cfg.Validator(triple.Predicate):
			decision = AdmissionDecisionKnown
			result.Event.KnownCount++
		default:
			decision = AdmissionDecisionUnknown
			result.Event.UnknownCount++
		}

		result.Records = append(result.Records, AdmissionRecord{
			Triple:   triple,
			Decision: decision,
		})
	}

	return result
}

func normalizeAdmissionSourceKind(batch AdmissionBatch) AdmissionSourceKind {
	if batch.SourceKind != "" {
		return batch.SourceKind
	}
	if batch.Source == AdmissionSourceContentSavedExtractor {
		return AdmissionSourceKindSynthetic
	}
	return AdmissionSourceKindEventBus
}

func isObservedAdmissionSource(sourceKind AdmissionSourceKind, source AdmissionSource) bool {
	if sourceKind == AdmissionSourceKindSynthetic {
		return source == AdmissionSourceContentSavedExtractor
	}
	return IsSupportedAdmissionSource(source)
}

func IsSupportedAdmissionSource(source AdmissionSource) bool {
	switch source {
	case AdmissionSourceConversationAnalysis,
		AdmissionSourceSessionLearning,
		AdmissionSourceLearning,
		AdmissionSourceProactiveLibrarian:
		return true
	}
	return false
}

func newAdmissionBatchEvent(batch AdmissionBatch) *eventbus.GraphAdmissionBatchEvent {
	return &eventbus.GraphAdmissionBatchEvent{
		Source:        string(batch.Source),
		ProducerGroup: CanonicalAdmissionProducerGroup(batch.Source),
		BatchCount:    1,
	}
}

func CanonicalAdmissionProducerGroup(source AdmissionSource) string {
	switch source {
	case AdmissionSourceConversationAnalysis,
		AdmissionSourceSessionLearning,
		AdmissionSourceLearning:
		return string(AdmissionProducerGroupLearning)
	case AdmissionSourceProactiveLibrarian:
		return string(AdmissionProducerGroupLibrarian)
	case AdmissionSourceContentSavedExtractor:
		return ""
	}
	return ""
}
