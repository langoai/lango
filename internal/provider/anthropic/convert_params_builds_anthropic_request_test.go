package anthropic

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	anthropicsdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/langoai/lango/internal/provider"
)

func TestConvertParamsBuildsAnthropicRequest(t *testing.T) {
	p := NewProvider("anthropic", "test-key")

	req, err := p.convertParams(provider.GenerateParams{
		Model:       "claude-3-5-sonnet-20241022",
		MaxTokens:   321,
		Temperature: 0.7,
		Messages: []provider.Message{
			{Role: "system", Content: "Follow the system policy."},
			{Role: "user", Content: "Plan a trip."},
			{Role: "tool", Content: "must be skipped"},
			{Role: "assistant", Content: "Where to?"},
		},
		Tools: []provider.Tool{{
			Name:        "lookup_weather",
			Description: "Look up weather.",
			Parameters: map[string]interface{}{
				"properties": map[string]interface{}{
					"city": map[string]interface{}{"type": "string"},
				},
				"required": []interface{}{"city"},
			},
		}},
	})
	if err != nil {
		t.Fatalf("convertParams returned error: %v", err)
	}

	body := marshalAnthropicParams(t, req)

	if got := body["model"]; got != "claude-3-5-sonnet-20241022" {
		t.Fatalf("model = %v, want claude-3-5-sonnet-20241022", got)
	}
	if got := body["max_tokens"]; got != float64(321) {
		t.Fatalf("max_tokens = %v, want 321", got)
	}
	if got := body["temperature"]; got != 0.7 {
		t.Fatalf("temperature = %v, want 0.7", got)
	}

	system := body["system"].([]interface{})
	if len(system) != 1 || system[0].(map[string]interface{})["text"] != "Follow the system policy." {
		t.Fatalf("system = %#v, want one text system block", system)
	}

	messages := body["messages"].([]interface{})
	if len(messages) != 2 {
		t.Fatalf("messages len = %d, want 2 after skipping system and unknown roles", len(messages))
	}
	assertConvertParamsBuildsAnthropicRequestMessage(t, messages[0], "user", "Plan a trip.")
	assertConvertParamsBuildsAnthropicRequestMessage(t, messages[1], "assistant", "Where to?")

	tools := body["tools"].([]interface{})
	if len(tools) != 1 {
		t.Fatalf("tools len = %d, want 1", len(tools))
	}
	tool := tools[0].(map[string]interface{})
	if tool["name"] != "lookup_weather" || tool["description"] != "Look up weather." {
		t.Fatalf("tool identity = %#v, want lookup_weather with description", tool)
	}
	schema := tool["input_schema"].(map[string]interface{})
	if schema["type"] != "object" {
		t.Fatalf("tool schema type = %v, want object", schema["type"])
	}
	required := schema["required"].([]interface{})
	if len(required) != 1 || required[0] != "city" {
		t.Fatalf("tool required = %#v, want [city]", required)
	}
}

func TestGenerateStreamsTextToolUsageAndBuildsRequest(t *testing.T) {
	var captured map[string]interface{}
	var gotPath, gotAPIKey, gotVersion, gotContentType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAPIKey = r.Header.Get("X-Api-Key")
		gotVersion = r.Header.Get("anthropic-version")
		gotContentType = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Errorf("decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		writeConvertParamsBuildsAnthropicRequestSSE(w, "message_start", `{"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","model":"claude-3-5-sonnet-20241022","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":11,"output_tokens":0,"cache_creation_input_tokens":2,"cache_read_input_tokens":3}}}`)
		writeConvertParamsBuildsAnthropicRequestSSE(w, "content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`)
		writeConvertParamsBuildsAnthropicRequestSSE(w, "content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}`)
		writeConvertParamsBuildsAnthropicRequestSSE(w, "content_block_stop", `{"type":"content_block_stop","index":0}`)
		writeConvertParamsBuildsAnthropicRequestSSE(w, "content_block_start", `{"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"toolu_1","name":"lookup_weather","input":{}}}`)
		writeConvertParamsBuildsAnthropicRequestSSE(w, "content_block_delta", `{"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"{\"city\":\"Seoul\"}"}}`)
		writeConvertParamsBuildsAnthropicRequestSSE(w, "content_block_stop", `{"type":"content_block_stop","index":1}`)
		writeConvertParamsBuildsAnthropicRequestSSE(w, "message_delta", `{"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":7}}`)
		writeConvertParamsBuildsAnthropicRequestSSE(w, "message_stop", `{"type":"message_stop"}`)
	}))
	defer server.Close()

	p := newConvertParamsBuildsAnthropicRequestTestProvider(server.URL, "test-key")
	seq, err := p.Generate(context.Background(), provider.GenerateParams{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 64,
		Messages:  []provider.Message{{Role: "user", Content: "Say hello."}},
	})
	if err != nil {
		t.Fatalf("Generate returned setup error: %v", err)
	}
	events := collectConvertParamsBuildsAnthropicRequestEvents(t, seq)

	if gotPath != "/v1/messages" {
		t.Fatalf("request path = %q, want /v1/messages", gotPath)
	}
	if gotAPIKey != "test-key" {
		t.Fatalf("X-Api-Key = %q, want test-key", gotAPIKey)
	}
	if gotVersion != "2023-06-01" {
		t.Fatalf("anthropic-version = %q, want 2023-06-01", gotVersion)
	}
	if !strings.HasPrefix(gotContentType, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", gotContentType)
	}
	if captured["stream"] != true {
		t.Fatalf("stream = %v, want true", captured["stream"])
	}
	if captured["model"] != "claude-3-5-sonnet-20241022" || captured["max_tokens"] != float64(64) {
		t.Fatalf("request model/max_tokens = %#v", captured)
	}

	if len(events) != 4 {
		t.Fatalf("events len = %d, want 4: %#v", len(events), events)
	}
	if events[0].Type != provider.StreamEventPlainText || events[0].Text != "Hello" {
		t.Fatalf("first event = %#v, want text delta Hello", events[0])
	}
	if events[1].Type != provider.StreamEventToolCall || events[1].ToolCall.ID != "toolu_1" || events[1].ToolCall.Name != "lookup_weather" {
		t.Fatalf("tool start event = %#v, want lookup_weather tool call", events[1])
	}
	if events[2].Type != provider.StreamEventToolCall || events[2].ToolCall.Arguments != `{"city":"Seoul"}` {
		t.Fatalf("tool delta event = %#v, want JSON arguments", events[2])
	}
	if events[3].Type != provider.StreamEventDone {
		t.Fatalf("done event type = %q, want done", events[3].Type)
	}
	if events[3].Usage == nil {
		t.Fatal("done usage is nil, want token usage")
	}
	if got := *events[3].Usage; got.InputTokens != 11 || got.OutputTokens != 7 || got.TotalTokens != 18 || got.CacheTokens != 5 {
		t.Fatalf("usage = %#v, want input=11 output=7 total=18 cache=5", got)
	}
}

func TestGenerateMapsHTTPErrorJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"type":"error","error":{"type":"invalid_request_error","message":"bad request"}}`))
	}))
	defer server.Close()

	p := newConvertParamsBuildsAnthropicRequestTestProvider(server.URL, "test-key")
	seq, err := p.Generate(context.Background(), provider.GenerateParams{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 32,
		Messages:  []provider.Message{{Role: "user", Content: "fail"}},
	})
	if err != nil {
		t.Fatalf("Generate returned setup error: %v", err)
	}
	events := collectConvertParamsBuildsAnthropicRequestEventsAllowError(t, seq)

	if len(events) != 1 {
		t.Fatalf("events len = %d, want one error event", len(events))
	}
	if events[0].Type != provider.StreamEventError || events[0].Error == nil {
		t.Fatalf("event = %#v, want provider error event", events[0])
	}
	var apiErr *anthropicsdk.Error
	if !errors.As(events[0].Error, &apiErr) {
		t.Fatalf("error type = %T, want *anthropic.Error", events[0].Error)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", apiErr.StatusCode)
	}
	if !strings.Contains(events[0].Error.Error(), "invalid_request_error") {
		t.Fatalf("error = %q, want invalid_request_error body", events[0].Error.Error())
	}
}

func TestGenerateMapsMalformedStreamToErrorEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		writeConvertParamsBuildsAnthropicRequestSSE(w, "content_block_delta", `{"type":"content_block_delta","index":0,"delta":`)
	}))
	defer server.Close()

	p := newConvertParamsBuildsAnthropicRequestTestProvider(server.URL, "test-key")
	seq, err := p.Generate(context.Background(), provider.GenerateParams{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 32,
		Messages:  []provider.Message{{Role: "user", Content: "malformed"}},
	})
	if err != nil {
		t.Fatalf("Generate returned setup error: %v", err)
	}
	events := collectConvertParamsBuildsAnthropicRequestEventsAllowError(t, seq)

	if len(events) != 1 {
		t.Fatalf("events len = %d, want one malformed-stream error event", len(events))
	}
	if events[0].Type != provider.StreamEventError || events[0].Error == nil {
		t.Fatalf("event = %#v, want error event", events[0])
	}
	if !strings.Contains(events[0].Error.Error(), "unexpected end of JSON input") {
		t.Fatalf("error = %q, want JSON parse failure", events[0].Error.Error())
	}
}

func TestGenerateRejectsMismatchedModelBeforeRequest(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	}))
	defer server.Close()

	p := newConvertParamsBuildsAnthropicRequestTestProvider(server.URL, "test-key")
	seq, err := p.Generate(context.Background(), provider.GenerateParams{
		Model:     "gpt-5.3-codex",
		MaxTokens: 32,
		Messages:  []provider.Message{{Role: "user", Content: "must not send"}},
	})

	if err == nil {
		t.Fatal("Generate error is nil, want model-provider mismatch")
	}
	if seq != nil {
		t.Fatalf("seq = %#v, want nil on setup error", seq)
	}
	if called {
		t.Fatal("server was called for mismatched model; want validation before request")
	}
	if !errors.Is(err, provider.ErrModelProviderMismatch) {
		t.Fatalf("error = %v, want ErrModelProviderMismatch", err)
	}
}

func marshalAnthropicParams(t *testing.T, params anthropicsdk.MessageNewParams) map[string]interface{} {
	t.Helper()

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("unmarshal params: %v", err)
	}
	return body
}

func assertConvertParamsBuildsAnthropicRequestMessage(t *testing.T, got interface{}, role, text string) {
	t.Helper()

	msg := got.(map[string]interface{})
	if msg["role"] != role {
		t.Fatalf("message role = %v, want %s", msg["role"], role)
	}
	content := msg["content"].([]interface{})
	if len(content) != 1 {
		t.Fatalf("content len = %d, want 1", len(content))
	}
	block := content[0].(map[string]interface{})
	if block["type"] != "text" || block["text"] != text {
		t.Fatalf("content block = %#v, want text %q", block, text)
	}
}

func newConvertParamsBuildsAnthropicRequestTestProvider(baseURL, apiKey string) *AnthropicProvider {
	client := anthropicsdk.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey(apiKey),
	)
	return &AnthropicProvider{
		client: &client,
		id:     "anthropic",
	}
}

func writeConvertParamsBuildsAnthropicRequestSSE(w http.ResponseWriter, event, data string) {
	_, _ = w.Write([]byte("event: " + event + "\n"))
	_, _ = w.Write([]byte("data: " + data + "\n\n"))
}

func collectConvertParamsBuildsAnthropicRequestEvents(t *testing.T, seq func(func(provider.StreamEvent, error) bool)) []provider.StreamEvent {
	t.Helper()

	var events []provider.StreamEvent
	for evt, err := range seq {
		if err != nil {
			t.Fatalf("unexpected stream error: %v", err)
		}
		events = append(events, evt)
	}
	return events
}

func collectConvertParamsBuildsAnthropicRequestEventsAllowError(t *testing.T, seq func(func(provider.StreamEvent, error) bool)) []provider.StreamEvent {
	t.Helper()

	var events []provider.StreamEvent
	for evt := range seq {
		events = append(events, evt)
	}
	return events
}
