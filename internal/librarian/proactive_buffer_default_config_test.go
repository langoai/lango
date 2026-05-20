package librarian

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	entinquiry "github.com/langoai/lango/internal/ent/inquiry"
	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/testutil"
	"github.com/langoai/lango/internal/types"
)

type proactiveBufferDefaultConfigTextGenerator struct {
	responses []string
	err       error
	calls     int
}

func (g *proactiveBufferDefaultConfigTextGenerator) GenerateText(_ context.Context, _, _ string) (string, error) {
	g.calls++
	if g.err != nil {
		return "", g.err
	}
	if len(g.responses) == 0 {
		return "{}", nil
	}
	response := g.responses[0]
	g.responses = g.responses[1:]
	return response, nil
}

func TestProactiveBufferDefaultConfig(t *testing.T) {
	buffer := NewProactiveBuffer(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		ProactiveBufferConfig{},
		testutil.NopLogger(),
	)

	require.NotNil(t, buffer)
	require.NotNil(t, buffer.inner)
	require.NotNil(t, buffer.turnCounter)
	assert.Equal(t, 2, buffer.observationThreshold)
	assert.Equal(t, 3, buffer.cooldownTurns)
	assert.Equal(t, 2, buffer.maxPending)
	assert.Equal(t, types.ConfidenceHigh, buffer.autoSaveConfidence)

	bus := eventbus.New()
	buffer.SetEventBus(bus)
	assert.Same(t, bus, buffer.bus)
}

func TestProactiveBufferProcessProviderErrors(t *testing.T) {
	t.Run("message provider error stops before observations", func(t *testing.T) {
		var messageCalls int
		var observationCalls int
		buffer := NewProactiveBuffer(
			nil,
			nil,
			nil,
			nil,
			func(string) ([]session.Message, error) {
				messageCalls++
				return nil, errors.New("messages unavailable")
			},
			func(context.Context, string) ([]memory.Observation, error) {
				observationCalls++
				return nil, nil
			},
			ProactiveBufferConfig{},
			testutil.NopLogger(),
		)

		buffer.process("session-errors")

		assert.Equal(t, 1, messageCalls)
		assert.Zero(t, observationCalls)
		assert.Empty(t, buffer.turnCounter)
	})

	t.Run("observation provider error stops before analysis", func(t *testing.T) {
		generator := &proactiveBufferDefaultConfigTextGenerator{responses: []string{`{"extractions":[],"gaps":[]}`}}
		buffer := NewProactiveBuffer(
			NewObservationAnalyzer(generator, testutil.NopLogger()),
			nil,
			nil,
			nil,
			func(string) ([]session.Message, error) {
				return nil, nil
			},
			func(context.Context, string) ([]memory.Observation, error) {
				return nil, errors.New("observations unavailable")
			},
			ProactiveBufferConfig{},
			testutil.NopLogger(),
		)

		buffer.process("session-errors")

		assert.Zero(t, generator.calls)
		assert.Empty(t, buffer.turnCounter)
	})
}

func TestProactiveBufferProcessSkipsBelowObservationThreshold(t *testing.T) {
	generator := &proactiveBufferDefaultConfigTextGenerator{responses: []string{`{"extractions":[],"gaps":[]}`}}
	buffer := NewProactiveBuffer(
		NewObservationAnalyzer(generator, testutil.NopLogger()),
		nil,
		nil,
		nil,
		func(string) ([]session.Message, error) {
			return nil, nil
		},
		func(context.Context, string) ([]memory.Observation, error) {
			return []memory.Observation{
				{Content: "The user prefers focused tests."},
				{Content: "The user wants deterministic behavior."},
			}, nil
		},
		ProactiveBufferConfig{ObservationThreshold: 3},
		testutil.NopLogger(),
	)

	buffer.process("session-threshold")

	assert.Zero(t, generator.calls)
	assert.Empty(t, buffer.turnCounter)
}

func TestProactiveBufferProcessCooldownAndMaxPendingInquiries(t *testing.T) {
	ctx := context.Background()
	client := testutil.TestEntClient(t)
	inquiryStore := NewInquiryStore(client, testutil.NopLogger())
	knowledgeStore := knowledge.NewStore(client, testutil.NopLogger())
	generator := &proactiveBufferDefaultConfigTextGenerator{responses: []string{
		proactiveBufferDefaultConfigAnalysisWithGapsJSON(),
		proactiveBufferDefaultConfigAnalysisWithGapsJSON(),
	}}
	buffer := NewProactiveBuffer(
		NewObservationAnalyzer(generator, testutil.NopLogger()),
		NewInquiryProcessor(&proactiveBufferDefaultConfigTextGenerator{}, inquiryStore, knowledgeStore, testutil.NopLogger()),
		inquiryStore,
		knowledgeStore,
		func(string) ([]session.Message, error) {
			return nil, nil
		},
		func(context.Context, string) ([]memory.Observation, error) {
			return []memory.Observation{{Content: "Missing runtime detail."}}, nil
		},
		ProactiveBufferConfig{
			ObservationThreshold: 1,
			CooldownTurns:        2,
			MaxPending:           2,
		},
		testutil.NopLogger(),
	)

	buffer.process("session-gaps")
	pending, err := inquiryStore.ListPendingInquiries(ctx, "session-gaps", 10)
	require.NoError(t, err)
	assert.Empty(t, pending)
	assert.Equal(t, 1, buffer.turnCounter["session-gaps"])

	buffer.process("session-gaps")
	pending, err = inquiryStore.ListPendingInquiries(ctx, "session-gaps", 10)
	require.NoError(t, err)
	require.Len(t, pending, 2)
	questionsByTopic := map[string]string{}
	for _, inquiry := range pending {
		questionsByTopic[inquiry.Topic] = inquiry.Question
	}
	assert.Equal(t, map[string]string{
		"runtime": "Which runtime should the agent prefer?",
		"testing": "Which test command should be authoritative?",
	}, questionsByTopic)
	assert.Zero(t, buffer.turnCounter["session-gaps"])
}

func TestProactiveBufferProcessAutoSaveUnknownTypeAndEvents(t *testing.T) {
	ctx := context.Background()
	client := testutil.TestEntClient(t)
	inquiryStore := NewInquiryStore(client, testutil.NopLogger())
	knowledgeStore := knowledge.NewStore(client, testutil.NopLogger())
	generator := &proactiveBufferDefaultConfigTextGenerator{responses: []string{`{
		"extractions": [
			{
				"type": "pattern",
				"content": "Use table-driven tests for branch-heavy helpers.",
				"confidence": "high",
				"key": "table_driven_branch_tests",
				"subject": "tests",
				"predicate": "prefer",
				"object": "table-driven branch coverage",
				"temporal": "evergreen"
			},
			{
				"type": "fact",
				"content": "Medium confidence facts need confirmation.",
				"confidence": "medium",
				"key": "medium_fact_skip"
			},
			{
				"type": "mystery",
				"content": "Unknown extraction types must be skipped.",
				"confidence": "high",
				"key": "unknown_type_skip",
				"subject": "unknown",
				"predicate": "should_not",
				"object": "publish"
			}
		],
		"gaps": []
	}`}}
	bus := eventbus.New()
	var tripleEvents []eventbus.TriplesExtractedEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.TriplesExtractedEvent) {
		tripleEvents = append(tripleEvents, evt)
	})
	buffer := NewProactiveBuffer(
		NewObservationAnalyzer(generator, testutil.NopLogger()),
		NewInquiryProcessor(&proactiveBufferDefaultConfigTextGenerator{}, inquiryStore, knowledgeStore, testutil.NopLogger()),
		inquiryStore,
		knowledgeStore,
		func(string) ([]session.Message, error) {
			return nil, nil
		},
		func(context.Context, string) ([]memory.Observation, error) {
			return []memory.Observation{{Content: "Use table-driven tests."}}, nil
		},
		ProactiveBufferConfig{
			ObservationThreshold: 1,
			CooldownTurns:        10,
		},
		testutil.NopLogger(),
	)
	buffer.SetEventBus(bus)

	buffer.process("session-autosave")

	got, err := knowledgeStore.GetKnowledge(ctx, "table_driven_branch_tests")
	require.NoError(t, err)
	assert.Equal(t, entknowledge.CategoryPattern, got.Category)
	assert.Equal(t, "Use table-driven tests for branch-heavy helpers.", got.Content)
	assert.Equal(t, "proactive_librarian", got.Source)
	assert.Contains(t, got.Tags, "temporal:evergreen")

	_, err = knowledgeStore.GetKnowledge(ctx, "medium_fact_skip")
	require.ErrorIs(t, err, knowledge.ErrKnowledgeNotFound)
	_, err = knowledgeStore.GetKnowledge(ctx, "unknown_type_skip")
	require.ErrorIs(t, err, knowledge.ErrKnowledgeNotFound)

	learnings, err := client.Learning.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, learnings, 1)
	assert.Equal(t, "proactive:table_driven_branch_tests", learnings[0].Trigger)
	assert.Equal(t, "Use table-driven tests for branch-heavy helpers.", learnings[0].Diagnosis)
	assert.Equal(t, entlearning.CategoryGeneral, learnings[0].Category)

	require.Len(t, tripleEvents, 1)
	require.Len(t, tripleEvents[0].Triples, 1)
	assert.Equal(t, "proactive_librarian", tripleEvents[0].Source)
	assert.Equal(t, "tests", tripleEvents[0].Triples[0].Subject)
	assert.Equal(t, "prefer", tripleEvents[0].Triples[0].Predicate)
	assert.Equal(t, "table-driven branch coverage", tripleEvents[0].Triples[0].Object)
}

func TestProactiveBufferProcessResolvesInquiryAnswers(t *testing.T) {
	ctx := context.Background()
	client := testutil.TestEntClient(t)
	inquiryStore := NewInquiryStore(client, testutil.NopLogger())
	knowledgeStore := knowledge.NewStore(client, testutil.NopLogger())
	inquiryID := uuid.New()
	require.NoError(t, inquiryStore.SaveInquiry(ctx, Inquiry{
		ID:         inquiryID,
		SessionKey: "session-answers",
		Topic:      "language",
		Question:   "Which language should examples use?",
		Context:    "Documentation examples need a default language.",
		Priority:   "high",
	}))

	generator := &proactiveBufferDefaultConfigTextGenerator{responses: []string{`[{
		"inquiry_id": "` + inquiryID.String() + `",
		"answer": "Use Go for examples.",
		"confidence": "high",
		"knowledge": {
			"key": "preferred_example_language",
			"category": "preference",
			"content": "Use Go for examples.",
			"temporal": "current_state"
		}
	}]`}}
	buffer := NewProactiveBuffer(
		NewObservationAnalyzer(generator, testutil.NopLogger()),
		NewInquiryProcessor(generator, inquiryStore, knowledgeStore, testutil.NopLogger()),
		inquiryStore,
		knowledgeStore,
		func(string) ([]session.Message, error) {
			return []session.Message{{Role: "user", Content: "Use Go for examples."}}, nil
		},
		func(context.Context, string) ([]memory.Observation, error) {
			return nil, nil
		},
		ProactiveBufferConfig{},
		testutil.NopLogger(),
	)

	buffer.process("session-answers")

	pending, err := inquiryStore.ListPendingInquiries(ctx, "session-answers", 10)
	require.NoError(t, err)
	assert.Empty(t, pending)

	row, err := client.Inquiry.Get(ctx, inquiryID)
	require.NoError(t, err)
	decoded, err := inquiryStore.entToInquiry(row)
	require.NoError(t, err)
	assert.Equal(t, string(entinquiry.StatusResolved), decoded.Status)
	assert.Equal(t, "Use Go for examples.", decoded.Answer)
	assert.Equal(t, "preferred_example_language", decoded.KnowledgeKey)

	got, err := knowledgeStore.GetKnowledge(ctx, "preferred_example_language")
	require.NoError(t, err)
	assert.Equal(t, entknowledge.CategoryPreference, got.Category)
	assert.Equal(t, "Use Go for examples.", got.Content)
	assert.Contains(t, got.Tags, "temporal:current_state")
	assert.Equal(t, 1, generator.calls)
}

func proactiveBufferDefaultConfigAnalysisWithGapsJSON() string {
	return `{
		"extractions": [],
		"gaps": [
			{
				"topic": "runtime",
				"question": "Which runtime should the agent prefer?",
				"context": "Runtime preference affects local verification.",
				"priority": "high"
			},
			{
				"topic": "testing",
				"question": "Which test command should be authoritative?",
				"context": "The package has several verification scopes.",
				"priority": "medium"
			},
			{
				"topic": "docs",
				"question": "Should docs be updated for this test-only change?",
				"context": "Avoid documenting unexposed behavior.",
				"priority": "low"
			}
		]
	}`
}
