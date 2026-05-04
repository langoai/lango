package eventbus

// Event name constants for observability domain events.
const (
	EventTokenUsage                   = "token.usage"
	EventGraphAdmissionBatch          = "graph.admission.batch"
	EventGraphAdmissionUnmappedSource = "graph.admission.unmapped_source"
	EventGraphExtractorDroppedUnknown = "graph.extractor.dropped_unknown"
	EventGraphAdmissionWriteFailure   = "graph.admission.write_failure"
)

// TokenUsageEvent is published when an LLM provider returns token usage data.
// The observability TokenTracker subscribes to this event.
type TokenUsageEvent struct {
	Provider         string
	Model            string
	SessionKey       string
	AgentName        string
	InputTokens      int64
	OutputTokens     int64
	TotalTokens      int64
	CacheTokens      int64
	EstimatedCostUSD float64 // 0 when model has no pricing entry; populated by emitter
}

// EventName implements Event.
func (e TokenUsageEvent) EventName() string { return EventTokenUsage }

// GraphAdmissionBatchEvent is published for one observe-only admission batch.
type GraphAdmissionBatchEvent struct {
	Source           string
	ProducerGroup    *string
	ValidatorSource  string
	BatchCount       int
	KnownCount       int
	UnknownCount     int
	UnvalidatedCount int
}

// EventName implements Event.
func (e GraphAdmissionBatchEvent) EventName() string { return EventGraphAdmissionBatch }

// GraphAdmissionUnmappedSourceEvent is published for one unsupported source batch.
type GraphAdmissionUnmappedSourceEvent struct {
	RawSource  string
	BatchCount int
}

// EventName implements Event.
func (e GraphAdmissionUnmappedSourceEvent) EventName() string {
	return EventGraphAdmissionUnmappedSource
}

// GraphExtractorDroppedUnknownEvent is published for one extractor dropped-unknown baseline.
type GraphExtractorDroppedUnknownEvent struct {
	Source    string
	SourceID  string
	Subject   string
	Predicate string
	Object    string
}

// EventName implements Event.
func (e GraphExtractorDroppedUnknownEvent) EventName() string {
	return EventGraphExtractorDroppedUnknown
}

// GraphAdmissionWriteFailureEvent is published for one aggregate graph write-failure baseline.
type GraphAdmissionWriteFailureEvent struct {
	BatchCount int
}

// EventName implements Event.
func (e GraphAdmissionWriteFailureEvent) EventName() string {
	return EventGraphAdmissionWriteFailure
}
