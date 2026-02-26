package librarian

import (
	"testing"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- stripCodeFence ---

func TestStripCodeFence_NoFence(t *testing.T) {
	input := `[{"inquiry_id":"abc","answer":"yes"}]`
	assert.Equal(t, input, stripCodeFence(input))
}

func TestStripCodeFence_JSONFence(t *testing.T) {
	input := "```json\n[{\"inquiry_id\":\"abc\"}]\n```"
	assert.Equal(t, `[{"inquiry_id":"abc"}]`, stripCodeFence(input))
}

func TestStripCodeFence_PlainFence(t *testing.T) {
	input := "```\n{\"key\":\"val\"}\n```"
	assert.Equal(t, `{"key":"val"}`, stripCodeFence(input))
}

func TestStripCodeFence_TrailingWhitespace(t *testing.T) {
	input := "  ```json\n  content  \n```  "
	result := stripCodeFence(input)
	assert.Equal(t, "content", result)
}

// --- parseAnswerMatches ---

func TestParseAnswerMatches_EmptyArray(t *testing.T) {
	matches, err := parseAnswerMatches("[]")
	require.NoError(t, err)
	assert.Empty(t, matches)
}

func TestParseAnswerMatches_SingleMatch(t *testing.T) {
	raw := `[{"inquiry_id":"550e8400-e29b-41d4-a716-446655440000","answer":"Go 1.21","confidence":"high"}]`
	matches, err := parseAnswerMatches(raw)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", matches[0].InquiryID)
	assert.Equal(t, "Go 1.21", matches[0].Answer)
	assert.Equal(t, types.ConfidenceHigh, matches[0].Confidence)
	assert.Nil(t, matches[0].Knowledge)
}

func TestParseAnswerMatches_WithKnowledge(t *testing.T) {
	raw := `[{
		"inquiry_id":"abc123",
		"answer":"Python",
		"confidence":"medium",
		"knowledge":{
			"key":"preferred_language",
			"category":"preference",
			"content":"User prefers Python"
		}
	}]`
	matches, err := parseAnswerMatches(raw)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	require.NotNil(t, matches[0].Knowledge)
	assert.Equal(t, "preferred_language", matches[0].Knowledge.Key)
	assert.Equal(t, "preference", matches[0].Knowledge.Category)
	assert.Equal(t, "User prefers Python", matches[0].Knowledge.Content)
}

func TestParseAnswerMatches_WithCodeFence(t *testing.T) {
	raw := "```json\n[{\"inquiry_id\":\"abc\",\"answer\":\"yes\",\"confidence\":\"high\"}]\n```"
	matches, err := parseAnswerMatches(raw)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	assert.Equal(t, "yes", matches[0].Answer)
}

func TestParseAnswerMatches_InvalidJSON(t *testing.T) {
	_, err := parseAnswerMatches("not json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse answer matches")
}

// --- parseAnalysisOutput ---

func TestParseAnalysisOutput_Valid(t *testing.T) {
	raw := `{
		"extractions": [
			{
				"type": "fact",
				"category": "tech",
				"content": "Uses Go",
				"confidence": "high",
				"key": "tech_go"
			}
		],
		"gaps": [
			{
				"topic": "testing",
				"question": "What test framework do you prefer?",
				"priority": "medium"
			}
		]
	}`
	out, err := parseAnalysisOutput(raw)
	require.NoError(t, err)
	require.Len(t, out.Extractions, 1)
	assert.Equal(t, "fact", out.Extractions[0].Type)
	assert.Equal(t, "tech_go", out.Extractions[0].Key)
	require.Len(t, out.Gaps, 1)
	assert.Equal(t, "testing", out.Gaps[0].Topic)
}

func TestParseAnalysisOutput_Empty(t *testing.T) {
	raw := `{"extractions":[],"gaps":[]}`
	out, err := parseAnalysisOutput(raw)
	require.NoError(t, err)
	assert.Empty(t, out.Extractions)
	assert.Empty(t, out.Gaps)
}

func TestParseAnalysisOutput_InvalidJSON(t *testing.T) {
	_, err := parseAnalysisOutput("{bad")
	assert.Error(t, err)
}

func TestParseAnalysisOutput_WithCodeFence(t *testing.T) {
	raw := "```json\n{\"extractions\":[],\"gaps\":[]}\n```"
	out, err := parseAnalysisOutput(raw)
	require.NoError(t, err)
	assert.Empty(t, out.Extractions)
}

// --- buildMatchPrompt ---

func TestBuildMatchPrompt_EmptyInputs(t *testing.T) {
	p := &InquiryProcessor{}
	result := p.buildMatchPrompt(nil, nil)
	assert.Contains(t, result, "## Pending Inquiries")
	assert.Contains(t, result, "## Recent Messages")
}

func TestBuildMatchPrompt_WithData(t *testing.T) {
	p := &InquiryProcessor{}

	id1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	id2 := uuid.MustParse("660e8400-e29b-41d4-a716-446655440000")

	pending := []Inquiry{
		{ID: id1, Topic: "language", Question: "What language do you prefer?"},
		{ID: id2, Topic: "framework", Question: "Which framework?"},
	}

	messages := []session.Message{
		{Role: "user", Content: "I prefer Go"},
		{Role: "assistant", Content: "Got it!"},
	}

	result := p.buildMatchPrompt(pending, messages)

	assert.Contains(t, result, id1.String())
	assert.Contains(t, result, id2.String())
	assert.Contains(t, result, "language")
	assert.Contains(t, result, "What language do you prefer?")
	assert.Contains(t, result, "[user]: I prefer Go")
	assert.Contains(t, result, "[assistant]: Got it!")
}

// --- NewInquiryProcessor ---

func TestNewInquiryProcessor(t *testing.T) {
	p := NewInquiryProcessor(nil, nil, nil, nil)
	require.NotNil(t, p)
	assert.Nil(t, p.generator)
	assert.Nil(t, p.inquiryStore)
	assert.Nil(t, p.knowledgeStore)
	assert.Nil(t, p.logger)
}

// --- answerMatch type ---

func TestAnswerMatch_Confidence(t *testing.T) {
	tests := []struct {
		conf types.Confidence
		low  bool
	}{
		{types.ConfidenceHigh, false},
		{types.ConfidenceMedium, false},
		{types.ConfidenceLow, true},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.low, tt.conf == types.ConfidenceLow,
			"confidence %q", tt.conf)
	}
}
