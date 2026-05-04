package eventbus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func stringPtr(v string) *string {
	return &v
}

func TestGraphAdmissionBatchEvent_EventName(t *testing.T) {
	t.Parallel()

	evt := GraphAdmissionBatchEvent{}

	assert.Equal(t, EventGraphAdmissionBatch, evt.EventName())
}

func TestGraphAdmissionBatchEvent_RoundTrip(t *testing.T) {
	t.Parallel()

	bus := New()

	var got GraphAdmissionBatchEvent
	SubscribeTyped(bus, func(evt GraphAdmissionBatchEvent) {
		got = evt
	})

	bus.Publish(GraphAdmissionBatchEvent{
		Source:           "conversation_analysis",
		ProducerGroup:    stringPtr("learning"),
		ValidatorSource:  "ontology_registry",
		BatchCount:       1,
		KnownCount:       2,
		UnknownCount:     1,
		UnvalidatedCount: 0,
	})

	assert.Equal(t, GraphAdmissionBatchEvent{
		Source:           "conversation_analysis",
		ProducerGroup:    stringPtr("learning"),
		ValidatorSource:  "ontology_registry",
		BatchCount:       1,
		KnownCount:       2,
		UnknownCount:     1,
		UnvalidatedCount: 0,
	}, got)
}

func TestGraphAdmissionUnmappedSourceEvent_EventName(t *testing.T) {
	t.Parallel()

	evt := GraphAdmissionUnmappedSourceEvent{}

	assert.Equal(t, EventGraphAdmissionUnmappedSource, evt.EventName())
}

func TestGraphAdmissionUnmappedSourceEvent_RoundTrip(t *testing.T) {
	t.Parallel()

	bus := New()

	var got GraphAdmissionUnmappedSourceEvent
	SubscribeTyped(bus, func(evt GraphAdmissionUnmappedSourceEvent) {
		got = evt
	})

	bus.Publish(GraphAdmissionUnmappedSourceEvent{
		RawSource:  "new_source",
		BatchCount: 1,
	})

	assert.Equal(t, GraphAdmissionUnmappedSourceEvent{
		RawSource:  "new_source",
		BatchCount: 1,
	}, got)
}

func TestGraphExtractorDroppedUnknownEvent_EventName(t *testing.T) {
	t.Parallel()

	evt := GraphExtractorDroppedUnknownEvent{}

	assert.Equal(t, EventGraphExtractorDroppedUnknown, evt.EventName())
}

func TestGraphExtractorDroppedUnknownEvent_RoundTrip(t *testing.T) {
	t.Parallel()

	bus := New()

	var got GraphExtractorDroppedUnknownEvent
	SubscribeTyped(bus, func(evt GraphExtractorDroppedUnknownEvent) {
		got = evt
	})

	bus.Publish(GraphExtractorDroppedUnknownEvent{
		Source:    "content_saved_extractor",
		SourceID:  "doc-1",
		Subject:   "a",
		Predicate: "invented_rel",
		Object:    "b",
	})

	assert.Equal(t, GraphExtractorDroppedUnknownEvent{
		Source:    "content_saved_extractor",
		SourceID:  "doc-1",
		Subject:   "a",
		Predicate: "invented_rel",
		Object:    "b",
	}, got)
}

func TestGraphAdmissionWriteFailureEvent_EventName(t *testing.T) {
	t.Parallel()

	evt := GraphAdmissionWriteFailureEvent{}

	assert.Equal(t, EventGraphAdmissionWriteFailure, evt.EventName())
}

func TestGraphAdmissionWriteFailureEvent_RoundTrip(t *testing.T) {
	t.Parallel()

	bus := New()

	var got GraphAdmissionWriteFailureEvent
	SubscribeTyped(bus, func(evt GraphAdmissionWriteFailureEvent) {
		got = evt
	})

	bus.Publish(GraphAdmissionWriteFailureEvent{
		BatchCount: 1,
	})

	assert.Equal(t, GraphAdmissionWriteFailureEvent{
		BatchCount: 1,
	}, got)
}
