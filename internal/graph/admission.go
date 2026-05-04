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

type AdmissionBatch struct {
	Source        AdmissionSource
	ProducerGroup AdmissionProducerGroup
	Triples       []Triple
}

type AdmissionRecord struct {
	Triple   Triple
	Decision AdmissionDecision
}

type AdmissionObserveResult struct {
	Records   []AdmissionRecord
	Forwarded []Triple
	Event     eventbus.GraphAdmissionBatchEvent
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
		Records:   make([]AdmissionRecord, 0, len(batch.Triples)),
		Forwarded: make([]Triple, 0, len(batch.Triples)),
		Event: eventbus.GraphAdmissionBatchEvent{
			Source:        string(batch.Source),
			ProducerGroup: string(batch.ProducerGroup),
			BatchCount:    1,
		},
	}

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
		result.Forwarded = append(result.Forwarded, triple)
	}

	return result
}
