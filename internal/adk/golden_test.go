package adk

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/provider"
	internal "github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// =============================================================================
// Test 1: Message Round-Trip
// Save via AppendEvent (eventToMessage) → restore via EventsAdapter.All().
// Verify FunctionCall and FunctionResponse fields survive exactly.
// =============================================================================

func TestGolden_MessageRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		giveEvents     []*session.Event
		wantRoles      []string // expected Content.Role in restored events
		wantTexts      []string // expected Text parts (empty string if none)
		wantFuncCalls  []roundTripFuncCall
		wantFuncResps  []roundTripFuncResp
	}{
		{
			name: "FunctionCall round-trip preserves ID, Name, Args, Thought, ThoughtSignature",
			giveEvents: []*session.Event{
				{
					Timestamp: time.Now(),
					Author:    "lango-agent",
					LLMResponse: model.LLMResponse{
						Content: &genai.Content{
							Role: "model",
							Parts: []*genai.Part{{
								FunctionCall: &genai.FunctionCall{
									ID:   "call_abc123",
									Name: "execute_command",
									Args: map[string]any{"cmd": "ls -la", "timeout": float64(30)},
								},
								Thought:          true,
								ThoughtSignature: []byte("sig-opaque-bytes-xyz"),
							}},
						},
					},
				},
			},
			wantRoles: []string{"assistant"},
			wantTexts: []string{""},
			wantFuncCalls: []roundTripFuncCall{
				{
					EventIdx: 0,
					ID:       "call_abc123",
					Name:     "execute_command",
					Args:     map[string]any{"cmd": "ls -la", "timeout": float64(30)},
					Thought:  true,
					ThoughtSig: []byte("sig-opaque-bytes-xyz"),
				},
			},
		},
		{
			name: "FunctionResponse round-trip preserves ID, Name, Response",
			giveEvents: []*session.Event{
				// Need a preceding FunctionCall so history is valid
				{
					Timestamp: time.Now(),
					Author:    "lango-agent",
					LLMResponse: model.LLMResponse{
						Content: &genai.Content{
							Role: "model",
							Parts: []*genai.Part{{
								FunctionCall: &genai.FunctionCall{
									ID:   "call_resp_test",
									Name: "read_file",
									Args: map[string]any{"path": "/tmp/test.txt"},
								},
							}},
						},
					},
				},
				{
					Timestamp: time.Now(),
					Author:    "tool",
					LLMResponse: model.LLMResponse{
						Content: &genai.Content{
							Role: "function",
							Parts: []*genai.Part{{
								FunctionResponse: &genai.FunctionResponse{
									ID:       "call_resp_test",
									Name:     "read_file",
									Response: map[string]any{"content": "file data here", "size": float64(1024)},
								},
							}},
						},
					},
				},
			},
			wantRoles: []string{"assistant", "function"},
			wantTexts: []string{"", ""},
			wantFuncResps: []roundTripFuncResp{
				{
					EventIdx: 1,
					ID:       "call_resp_test",
					Name:     "read_file",
					Response: map[string]any{"content": "file data here", "size": float64(1024)},
				},
			},
			wantFuncCalls: []roundTripFuncCall{
				{
					EventIdx: 0,
					ID:       "call_resp_test",
					Name:     "read_file",
					Args:     map[string]any{"path": "/tmp/test.txt"},
				},
			},
		},
		{
			name: "plain text user + assistant round-trip",
			giveEvents: []*session.Event{
				{
					Timestamp: time.Now(),
					Author:    "user",
					LLMResponse: model.LLMResponse{
						Content: &genai.Content{
							Role:  "user",
							Parts: []*genai.Part{{Text: "What is 2+2?"}},
						},
					},
				},
				{
					Timestamp: time.Now(),
					Author:    "lango-agent",
					LLMResponse: model.LLMResponse{
						Content: &genai.Content{
							Role:  "model",
							Parts: []*genai.Part{{Text: "The answer is 4."}},
						},
					},
				},
			},
			wantRoles: []string{"user", "assistant"},
			wantTexts: []string{"What is 2+2?", "The answer is 4."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store := newMockStore()
			sess := &internal.Session{
				Key:       "golden-rt-" + tt.name,
				Metadata:  make(map[string]string),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			require.NoError(t, store.Create(sess))

			adapter := NewSessionAdapter(sess, store, "lango-agent")
			svc := NewSessionServiceAdapter(store, "lango-agent")

			// Save all events
			for _, evt := range tt.giveEvents {
				require.NoError(t, svc.AppendEvent(context.Background(), adapter, evt))
			}

			// Restore via EventsAdapter.All()
			events := adapter.Events()
			var restored []*session.Event
			for evt := range events.All() {
				restored = append(restored, evt)
			}

			require.Len(t, restored, len(tt.wantRoles), "event count mismatch")

			for i, wantRole := range tt.wantRoles {
				assert.Equal(t, wantRole, restored[i].Content.Role, "role mismatch at event %d", i)
			}

			for i, wantText := range tt.wantTexts {
				gotText := ""
				for _, p := range restored[i].Content.Parts {
					if p.Text != "" {
						gotText += p.Text
					}
				}
				assert.Equal(t, wantText, gotText, "text mismatch at event %d", i)
			}

			for _, wfc := range tt.wantFuncCalls {
				evt := restored[wfc.EventIdx]
				var found *genai.Part
				for _, p := range evt.Content.Parts {
					if p.FunctionCall != nil && p.FunctionCall.Name == wfc.Name {
						found = p
						break
					}
				}
				require.NotNil(t, found, "FunctionCall %q not found in event %d", wfc.Name, wfc.EventIdx)
				assert.Equal(t, wfc.ID, found.FunctionCall.ID, "FunctionCall.ID mismatch")
				assert.Equal(t, wfc.Name, found.FunctionCall.Name, "FunctionCall.Name mismatch")
				for k, v := range wfc.Args {
					assert.Equal(t, v, found.FunctionCall.Args[k], "FunctionCall.Args[%s] mismatch", k)
				}
				assert.Equal(t, wfc.Thought, found.Thought, "Thought mismatch")
				if wfc.ThoughtSig != nil {
					assert.Equal(t, wfc.ThoughtSig, found.ThoughtSignature, "ThoughtSignature mismatch")
				}
			}

			for _, wfr := range tt.wantFuncResps {
				evt := restored[wfr.EventIdx]
				var found *genai.Part
				for _, p := range evt.Content.Parts {
					if p.FunctionResponse != nil && p.FunctionResponse.Name == wfr.Name {
						found = p
						break
					}
				}
				require.NotNil(t, found, "FunctionResponse %q not found in event %d", wfr.Name, wfr.EventIdx)
				assert.Equal(t, wfr.ID, found.FunctionResponse.ID, "FunctionResponse.ID mismatch")
				assert.Equal(t, wfr.Name, found.FunctionResponse.Name, "FunctionResponse.Name mismatch")
				for k, v := range wfr.Response {
					assert.Equal(t, v, found.FunctionResponse.Response[k], "FunctionResponse.Response[%s] mismatch", k)
				}
			}
		})
	}
}

type roundTripFuncCall struct {
	EventIdx   int
	ID         string
	Name       string
	Args       map[string]any
	Thought    bool
	ThoughtSig []byte
}

type roundTripFuncResp struct {
	EventIdx int
	ID       string
	Name     string
	Response map[string]any
}

// =============================================================================
// Test 2: Streaming Partial/Final Deduplication
// Verify that when streaming produces partial events followed by a final done
// event, text is NOT duplicated in RunAndCollect/runAndCollectOnce output.
// =============================================================================

func TestGolden_StreamingPartialFinalDedup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		giveEvents []provider.StreamEvent
		wantText   string
	}{
		{
			name: "partial chunks then done: text not duplicated",
			giveEvents: []provider.StreamEvent{
				{Type: provider.StreamEventPlainText, Text: "Hello "},
				{Type: provider.StreamEventPlainText, Text: "world"},
				{Type: provider.StreamEventDone},
			},
			wantText: "Hello world",
		},
		{
			name: "single partial then done with full text",
			giveEvents: []provider.StreamEvent{
				{Type: provider.StreamEventPlainText, Text: "Complete response"},
				{Type: provider.StreamEventDone},
			},
			wantText: "Complete response",
		},
		{
			name: "many small partials then done",
			giveEvents: []provider.StreamEvent{
				{Type: provider.StreamEventPlainText, Text: "A"},
				{Type: provider.StreamEventPlainText, Text: "B"},
				{Type: provider.StreamEventPlainText, Text: "C"},
				{Type: provider.StreamEventPlainText, Text: "D"},
				{Type: provider.StreamEventDone},
			},
			wantText: "ABCD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &mockProvider{id: "test", events: tt.giveEvents}
			adapter := NewModelAdapter(p, "test-model")

			req := &model.LLMRequest{Model: "test-model"}
			seq := adapter.GenerateContent(context.Background(), req, true)

			// Simulate what runAndCollectOnce does:
			// collect partial text only, skip final done text.
			var collected string
			sawPartial := false
			for resp, err := range seq {
				require.NoError(t, err)
				if resp.Partial {
					sawPartial = true
					for _, part := range resp.Content.Parts {
						if part.Text != "" {
							collected += part.Text
						}
					}
				} else if !sawPartial {
					for _, part := range resp.Content.Parts {
						if part.Text != "" {
							collected += part.Text
						}
					}
				}
				// sawPartial && !resp.Partial: skip (done event text duplicates partials)
			}

			assert.Equal(t, tt.wantText, collected, "streaming dedup failed")
		})
	}
}

// =============================================================================
// Test 3: Orphaned Tool Response (Delta before Start)
// Verify delta chunk with only Arguments arriving before any start chunk
// is dropped by toolCallAccumulator.
// =============================================================================

func TestGolden_OrphanedToolResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		giveDeltas []*provider.ToolCall
		wantParts  int
	}{
		{
			name: "orphan delta with only arguments is dropped",
			giveDeltas: []*provider.ToolCall{
				{Arguments: `{"key":"value"}`}, // no Index, ID, or Name
			},
			wantParts: 0,
		},
		{
			name: "orphan delta followed by valid start produces one part",
			giveDeltas: []*provider.ToolCall{
				{Arguments: `{"orphan":"true"}`}, // dropped
				{ID: "call_1", Name: "exec"},     // valid start
				{Arguments: `{"cmd":"ls"}`},       // appended to call_1
			},
			wantParts: 1,
		},
		{
			name: "nil tool call is safely ignored",
			giveDeltas: []*provider.ToolCall{nil},
			wantParts:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var acc toolCallAccumulator
			for _, delta := range tt.giveDeltas {
				acc.add(delta)
			}
			parts := acc.done()
			assert.Len(t, parts, tt.wantParts)
		})
	}
}

// =============================================================================
// Test 4: Delegation Event Preservation
// Verify delegation events with author field survive save/restore with
// author preserved.
// =============================================================================

func TestGolden_DelegationEventPreservation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		giveAuthor string
		giveRole   string
		giveText   string
		wantAuthor string
	}{
		{
			name:       "orchestrator author preserved",
			giveAuthor: "lango-orchestrator",
			giveRole:   "model",
			giveText:   "Delegating to researcher",
			wantAuthor: "lango-orchestrator",
		},
		{
			name:       "sub-agent author preserved",
			giveAuthor: "researcher-agent",
			giveRole:   "model",
			giveText:   "Research results here",
			wantAuthor: "researcher-agent",
		},
		{
			name:       "executor author preserved",
			giveAuthor: "executor",
			giveRole:   "model",
			giveText:   "Executing task",
			wantAuthor: "executor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store := newMockStore()
			sess := &internal.Session{
				Key:       "golden-deleg-" + tt.name,
				Metadata:  make(map[string]string),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			require.NoError(t, store.Create(sess))

			adapter := NewSessionAdapter(sess, store, "lango-orchestrator")
			svc := NewSessionServiceAdapter(store, "lango-orchestrator")

			// Append a user message first so we don't get merged with the delegation event
			userEvt := &session.Event{
				Timestamp: time.Now(),
				Author:    "user",
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Role:  "user",
						Parts: []*genai.Part{{Text: "do something"}},
					},
				},
			}
			require.NoError(t, svc.AppendEvent(context.Background(), adapter, userEvt))

			// Append delegation event
			evt := &session.Event{
				Timestamp: time.Now(),
				Author:    tt.giveAuthor,
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Role:  tt.giveRole,
						Parts: []*genai.Part{{Text: tt.giveText}},
					},
				},
			}
			require.NoError(t, svc.AppendEvent(context.Background(), adapter, evt))

			// Verify stored author in internal message
			require.Len(t, adapter.sess.History, 2)
			assert.Equal(t, tt.wantAuthor, adapter.sess.History[1].Author,
				"author not preserved in internal message")

			// Verify author survives EventsAdapter.All() restore
			events := adapter.Events()
			var restored []*session.Event
			for evt := range events.All() {
				restored = append(restored, evt)
			}
			require.Len(t, restored, 2)
			assert.Equal(t, tt.wantAuthor, restored[1].Author,
				"author not preserved in restored event")
		})
	}
}

// =============================================================================
// Test 5: Isolated-Agent Child History
// Verify isolated agent child session messages are segregated from parent.
// =============================================================================

func TestGolden_IsolatedAgentChildHistory(t *testing.T) {
	t.Parallel()

	parentStore := newMockStore()
	parentSess := &internal.Session{
		Key:       "parent-sess",
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		History: []internal.Message{
			{Role: types.RoleUser, Content: "parent question", Timestamp: time.Now()},
			{Role: types.RoleAssistant, Content: "parent answer", Timestamp: time.Now()},
		},
	}
	require.NoError(t, parentStore.Create(parentSess))

	// Create child store
	childStore := internal.NewInMemoryChildStore(parentStore)

	// Fork a child session with isolation (no history inheritance)
	child, err := childStore.ForkChild("parent-sess", "isolated-agent", internal.ChildSessionConfig{
		InheritHistory: 0, // isolated — no parent history
	})
	require.NoError(t, err)

	// Add messages to child
	child.History = append(child.History, internal.Message{
		Role:    types.RoleUser,
		Content: "child-only question",
	})
	child.History = append(child.History, internal.Message{
		Role:    types.RoleAssistant,
		Content: "child-only answer",
	})

	// Verify child has only child messages (not parent)
	assert.Len(t, child.History, 2, "child should have only its own messages")
	assert.Equal(t, "child-only question", child.History[0].Content)
	assert.Equal(t, "child-only answer", child.History[1].Content)

	// Verify parent still has only parent messages
	parentAdapter := NewSessionAdapter(parentSess, parentStore, "lango-agent")
	parentEvents := parentAdapter.Events()
	parentCount := 0
	for range parentEvents.All() {
		parentCount++
	}
	assert.Equal(t, 2, parentCount, "parent should still have only 2 messages")

	// Verify child messages don't appear in parent's EventsAdapter
	for evt := range parentAdapter.Events().All() {
		for _, p := range evt.Content.Parts {
			assert.NotContains(t, p.Text, "child-only",
				"child message leaked into parent events")
		}
	}
}

// =============================================================================
// Test 6: thought_signature Round-Trip
// Verify ThoughtSignature (opaque []byte) preserved through
// save (session_service.go) → store → restore (state.go).
// =============================================================================

func TestGolden_ThoughtSignatureRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		giveThought      bool
		giveThoughtSig   []byte
		wantThought      bool
		wantThoughtSig   []byte
	}{
		{
			name:           "binary thought signature preserved",
			giveThought:    true,
			giveThoughtSig: []byte{0x01, 0x02, 0xFF, 0xFE, 0x00, 0xAB},
			wantThought:    true,
			wantThoughtSig: []byte{0x01, 0x02, 0xFF, 0xFE, 0x00, 0xAB},
		},
		{
			name:           "long base64-like signature preserved",
			giveThought:    true,
			giveThoughtSig: []byte("aGVsbG8gd29ybGQgdGhpcyBpcyBhIGxvbmcgc2lnbmF0dXJl"),
			wantThought:    true,
			wantThoughtSig: []byte("aGVsbG8gd29ybGQgdGhpcyBpcyBhIGxvbmcgc2lnbmF0dXJl"),
		},
		{
			name:           "nil signature stays nil",
			giveThought:    false,
			giveThoughtSig: nil,
			wantThought:    false,
			wantThoughtSig: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store := newMockStore()
			sess := &internal.Session{
				Key:       "golden-sig-" + tt.name,
				Metadata:  make(map[string]string),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			require.NoError(t, store.Create(sess))

			adapter := NewSessionAdapter(sess, store, "lango-agent")
			svc := NewSessionServiceAdapter(store, "lango-agent")

			// Save event with ThoughtSignature
			evt := &session.Event{
				Timestamp: time.Now(),
				Author:    "lango-agent",
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Role: "model",
						Parts: []*genai.Part{{
							FunctionCall: &genai.FunctionCall{
								ID:   "call_thought_test",
								Name: "think_tool",
								Args: map[string]any{"query": "test"},
							},
							Thought:          tt.giveThought,
							ThoughtSignature: tt.giveThoughtSig,
						}},
					},
				},
			}
			require.NoError(t, svc.AppendEvent(context.Background(), adapter, evt))

			// Verify internal storage preserved the fields
			require.Len(t, adapter.sess.History, 1)
			storedTC := adapter.sess.History[0].ToolCalls[0]
			assert.Equal(t, tt.giveThought, storedTC.Thought, "Thought not preserved in store")
			assert.Equal(t, tt.giveThoughtSig, storedTC.ThoughtSignature, "ThoughtSignature not preserved in store")

			// Restore via EventsAdapter
			events := adapter.Events()
			var restored []*session.Event
			for evt := range events.All() {
				restored = append(restored, evt)
			}
			require.Len(t, restored, 1)

			var foundPart *genai.Part
			for _, p := range restored[0].Content.Parts {
				if p.FunctionCall != nil {
					foundPart = p
					break
				}
			}
			require.NotNil(t, foundPart, "no FunctionCall part found in restored event")
			assert.Equal(t, tt.wantThought, foundPart.Thought, "Thought not preserved in restore")
			if tt.wantThoughtSig == nil {
				assert.Nil(t, foundPart.ThoughtSignature, "expected nil ThoughtSignature")
			} else {
				assert.Equal(t, tt.wantThoughtSig, foundPart.ThoughtSignature, "ThoughtSignature not preserved in restore")
			}
		})
	}
}

// =============================================================================
// Test 7: FunctionResponse Role Correction Regression
// Verify ADK-stored FunctionResponse with role="user" is corrected to "tool"
// on both save (AppendEvent) and restore (EventsAdapter.All()).
// =============================================================================

func TestGolden_FunctionResponseRoleCorrection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		giveRole       string // role as ADK sends it
		giveParts      []*genai.Part
		wantStoreRole  types.MessageRole // expected role in internal message
		wantEventRole  string            // expected role in restored event
	}{
		{
			name:     "ADK sends role=user for FunctionResponse — corrected to tool on save, function on restore",
			giveRole: "user",
			giveParts: []*genai.Part{{
				FunctionResponse: &genai.FunctionResponse{
					ID:       "call_fix_1",
					Name:     "exec",
					Response: map[string]any{"output": "done"},
				},
			}},
			wantStoreRole: types.RoleTool,
			wantEventRole: "function",
		},
		{
			name:     "correct role=function preserved",
			giveRole: "function",
			giveParts: []*genai.Part{{
				FunctionResponse: &genai.FunctionResponse{
					ID:       "call_fix_2",
					Name:     "read",
					Response: map[string]any{"content": "data"},
				},
			}},
			wantStoreRole: types.RoleTool,
			wantEventRole: "function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store := newMockStore()
			sess := &internal.Session{
				Key:       "golden-rolecorr-" + tt.name,
				Metadata:  make(map[string]string),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			require.NoError(t, store.Create(sess))

			adapter := NewSessionAdapter(sess, store, "lango-agent")
			svc := NewSessionServiceAdapter(store, "lango-agent")

			// Prepend an assistant FunctionCall so the tool response has context
			fcEvt := &session.Event{
				Timestamp: time.Now(),
				Author:    "lango-agent",
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Role: "model",
						Parts: []*genai.Part{{
							FunctionCall: &genai.FunctionCall{
								ID:   tt.giveParts[0].FunctionResponse.ID,
								Name: tt.giveParts[0].FunctionResponse.Name,
								Args: map[string]any{"input": "test"},
							},
						}},
					},
				},
			}
			require.NoError(t, svc.AppendEvent(context.Background(), adapter, fcEvt))

			// Save FunctionResponse with potentially wrong role
			frEvt := &session.Event{
				Timestamp: time.Now(),
				Author:    "tool",
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Role:  tt.giveRole,
						Parts: tt.giveParts,
					},
				},
			}
			require.NoError(t, svc.AppendEvent(context.Background(), adapter, frEvt))

			// Verify stored role is corrected
			require.Len(t, adapter.sess.History, 2)
			assert.Equal(t, tt.wantStoreRole, adapter.sess.History[1].Role,
				"stored role not corrected")

			// Verify restored event has correct role
			events := adapter.Events()
			var restored []*session.Event
			for evt := range events.All() {
				restored = append(restored, evt)
			}
			require.Len(t, restored, 2)
			assert.Equal(t, tt.wantEventRole, restored[1].Content.Role,
				"restored event role incorrect")
		})
	}
}

// =============================================================================
// Test 8: Get() Auto-Create/Renew Regression
// Verify SessionServiceAdapter.Get() auto-creates when missing,
// auto-renews when expired.
// =============================================================================

func TestGolden_GetAutoCreateRenew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		giveSetup     func(store *mockStore)
		giveSessionID string
		wantErr       bool
		wantSessionID string
	}{
		{
			name:          "auto-creates when session does not exist",
			giveSetup:     func(store *mockStore) { /* nothing — session missing */ },
			giveSessionID: "new-session",
			wantErr:       false,
			wantSessionID: "new-session",
		},
		{
			name: "auto-creates when store returns nil session",
			giveSetup: func(store *mockStore) {
				// Mock Get returns nil, nil (no session found but no error)
				// This is already the default behavior of mockStore
			},
			giveSessionID: "nil-session",
			wantErr:       false,
			wantSessionID: "nil-session",
		},
		{
			name: "auto-renews when session is expired",
			giveSetup: func(store *mockStore) {
				store.sessions["expired-sess"] = &internal.Session{
					Key:      "expired-sess",
					Metadata: map[string]string{"old": "data"},
				}
				store.expiredKeys["expired-sess"] = true
			},
			giveSessionID: "expired-sess",
			wantErr:       false,
			wantSessionID: "expired-sess",
		},
		{
			name: "expired session delete failure returns error",
			giveSetup: func(store *mockStore) {
				store.sessions["fail-del"] = &internal.Session{Key: "fail-del"}
				store.expiredKeys["fail-del"] = true
				store.deleteErr = fmt.Errorf("disk full")
			},
			giveSessionID: "fail-del",
			wantErr:       true,
		},
		{
			name: "returns existing session when found",
			giveSetup: func(store *mockStore) {
				store.sessions["existing-sess"] = &internal.Session{
					Key:       "existing-sess",
					Metadata:  map[string]string{"key": "value"},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
			},
			giveSessionID: "existing-sess",
			wantErr:       false,
			wantSessionID: "existing-sess",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store := newMockStore()
			tt.giveSetup(store)

			svc := NewSessionServiceAdapter(store, "lango-agent")
			resp, err := svc.Get(context.Background(), &session.GetRequest{
				SessionID: tt.giveSessionID,
			})

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Session)
			assert.Equal(t, tt.wantSessionID, resp.Session.ID())

			// Verify session actually exists in store after auto-create/renew
			stored, storeErr := store.Get(tt.wantSessionID)
			require.NoError(t, storeErr)
			require.NotNil(t, stored, "session should exist in store after Get")
		})
	}
}

// =============================================================================
// Test: Full Pipeline — FunctionCall + FunctionResponse + Text Round-Trip
// End-to-end: save a multi-turn conversation with tool use, restore, verify all fields.
// =============================================================================

func TestGolden_FullPipelineRoundTrip(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	sess := &internal.Session{
		Key:       "golden-pipeline",
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, store.Create(sess))

	adapter := NewSessionAdapter(sess, store, "lango-agent")
	svc := NewSessionServiceAdapter(store, "lango-agent")

	// Step 1: User asks a question
	userEvt := &session.Event{
		Timestamp: time.Now(),
		Author:    "user",
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{
				Role:  "user",
				Parts: []*genai.Part{{Text: "List files in /tmp"}},
			},
		},
	}
	require.NoError(t, svc.AppendEvent(context.Background(), adapter, userEvt))

	// Step 2: Assistant calls a tool
	assistantFCEvt := &session.Event{
		Timestamp: time.Now(),
		Author:    "lango-agent",
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{
				Role: "model",
				Parts: []*genai.Part{{
					FunctionCall: &genai.FunctionCall{
						ID:   "call_ls_1",
						Name: "exec",
						Args: map[string]any{"command": "ls /tmp"},
					},
					Thought:          true,
					ThoughtSignature: []byte("thinking-about-listing-files"),
				}},
			},
		},
	}
	require.NoError(t, svc.AppendEvent(context.Background(), adapter, assistantFCEvt))

	// Step 3: Tool returns result
	toolEvt := &session.Event{
		Timestamp: time.Now(),
		Author:    "tool",
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{
				Role: "function",
				Parts: []*genai.Part{{
					FunctionResponse: &genai.FunctionResponse{
						ID:       "call_ls_1",
						Name:     "exec",
						Response: map[string]any{"output": "file1.txt\nfile2.txt"},
					},
				}},
			},
		},
	}
	require.NoError(t, svc.AppendEvent(context.Background(), adapter, toolEvt))

	// Step 4: Assistant responds with text
	assistantTextEvt := &session.Event{
		Timestamp: time.Now(),
		Author:    "lango-agent",
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{
				Role:  "model",
				Parts: []*genai.Part{{Text: "Found: file1.txt and file2.txt"}},
			},
		},
	}
	require.NoError(t, svc.AppendEvent(context.Background(), adapter, assistantTextEvt))

	// Restore and verify
	events := adapter.Events()
	var restored []*session.Event
	for evt := range events.All() {
		restored = append(restored, evt)
	}
	require.Len(t, restored, 4, "expected 4 events in full pipeline")

	// Event 0: User
	assert.Equal(t, "user", restored[0].Content.Role)
	assert.Equal(t, "List files in /tmp", restored[0].Content.Parts[0].Text)

	// Event 1: Assistant FunctionCall
	assert.Equal(t, "assistant", restored[1].Content.Role)
	require.NotNil(t, restored[1].Content.Parts[0].FunctionCall)
	assert.Equal(t, "call_ls_1", restored[1].Content.Parts[0].FunctionCall.ID)
	assert.Equal(t, "exec", restored[1].Content.Parts[0].FunctionCall.Name)
	assert.Equal(t, "ls /tmp", restored[1].Content.Parts[0].FunctionCall.Args["command"])
	assert.True(t, restored[1].Content.Parts[0].Thought, "Thought flag not preserved")
	assert.Equal(t, []byte("thinking-about-listing-files"),
		restored[1].Content.Parts[0].ThoughtSignature, "ThoughtSignature not preserved")

	// Event 2: FunctionResponse
	assert.Equal(t, "function", restored[2].Content.Role)
	require.NotNil(t, restored[2].Content.Parts[0].FunctionResponse)
	assert.Equal(t, "call_ls_1", restored[2].Content.Parts[0].FunctionResponse.ID)
	assert.Equal(t, "exec", restored[2].Content.Parts[0].FunctionResponse.Name)
	assert.Equal(t, "file1.txt\nfile2.txt",
		restored[2].Content.Parts[0].FunctionResponse.Response["output"])

	// Event 3: Assistant text
	assert.Equal(t, "assistant", restored[3].Content.Role)
	assert.Equal(t, "Found: file1.txt and file2.txt", restored[3].Content.Parts[0].Text)
}

// =============================================================================
// Test: toolCallAccumulator preserves ThoughtSignature through streaming
// =============================================================================

func TestGolden_AccumulatorThoughtSignatureStreaming(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		giveChunks     []*provider.ToolCall
		wantThought    bool
		wantThoughtSig []byte
		wantName       string
		wantArgs       map[string]any
	}{
		{
			name: "OpenAI-style streaming preserves thought fields",
			giveChunks: []*provider.ToolCall{
				{
					Index:            intPtr(0),
					ID:               "call_think",
					Name:             "deep_think",
					Arguments:        `{"query":`,
					Thought:          true,
					ThoughtSignature: []byte("sig-data"),
				},
				{
					Index:     intPtr(0),
					Arguments: `"test"}`,
				},
			},
			wantThought:    true,
			wantThoughtSig: []byte("sig-data"),
			wantName:       "deep_think",
			wantArgs:       map[string]any{"query": "test"},
		},
		{
			name: "Anthropic-style streaming preserves thought fields",
			giveChunks: []*provider.ToolCall{
				{
					ID:               "tool_think",
					Name:             "analyze",
					Thought:          true,
					ThoughtSignature: []byte("anthropic-sig"),
				},
				{
					Arguments: `{"data":"value"}`,
				},
			},
			wantThought:    true,
			wantThoughtSig: []byte("anthropic-sig"),
			wantName:       "analyze",
			wantArgs:       map[string]any{"data": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var acc toolCallAccumulator
			for _, chunk := range tt.giveChunks {
				acc.add(chunk)
			}

			parts := acc.done()
			require.Len(t, parts, 1)

			p := parts[0]
			assert.Equal(t, tt.wantThought, p.Thought)
			assert.Equal(t, tt.wantThoughtSig, p.ThoughtSignature)
			assert.Equal(t, tt.wantName, p.FunctionCall.Name)
			for k, v := range tt.wantArgs {
				assert.Equal(t, v, p.FunctionCall.Args[k])
			}
		})
	}
}

// =============================================================================
// Test: EventsAdapter role correction for user-role with tool output
// Specifically targets the role correction block in state.go:246-257.
// =============================================================================

func TestGolden_EventsAdapter_UserRoleWithToolOutput(t *testing.T) {
	t.Parallel()

	now := time.Now()

	// Create a session where a message has role="user" but contains
	// ToolCalls with Output set — this is the ADK bug scenario.
	sess := &internal.Session{
		Key: "golden-rolecorr-events",
		History: []internal.Message{
			// Preceding assistant FunctionCall
			{
				Role:      types.RoleAssistant,
				Content:   "",
				Timestamp: now,
				ToolCalls: []internal.ToolCall{
					{
						ID:    "call_fix_evt",
						Name:  "exec",
						Input: `{"cmd":"ls"}`,
					},
				},
			},
			// FunctionResponse incorrectly stored as "user" role
			{
				Role:      types.RoleUser,
				Content:   `{"output":"files"}`,
				Timestamp: now.Add(time.Second),
				ToolCalls: []internal.ToolCall{
					{
						ID:     "call_fix_evt",
						Name:   "exec",
						Output: `{"output":"files"}`,
					},
				},
			},
		},
	}

	adapter := NewSessionAdapter(sess, newMockStore(), "lango-agent")
	events := adapter.Events()

	var restored []*session.Event
	for evt := range events.All() {
		restored = append(restored, evt)
	}

	require.Len(t, restored, 2)

	// The user-role message with tool output should be corrected to "function"
	assert.Equal(t, "function", restored[1].Content.Role,
		"expected role correction from user to function for FunctionResponse")

	// Should have a FunctionResponse part, not a text part
	require.NotEmpty(t, restored[1].Content.Parts)
	found := false
	for _, p := range restored[1].Content.Parts {
		if p.FunctionResponse != nil {
			found = true
			assert.Equal(t, "call_fix_evt", p.FunctionResponse.ID)
			assert.Equal(t, "exec", p.FunctionResponse.Name)
		}
	}
	assert.True(t, found, "expected FunctionResponse part in corrected event")
}

// =============================================================================
// Test: Multiple FunctionCalls in a single assistant message
// =============================================================================

func TestGolden_MultipleFunctionCallsInOneMessage(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	sess := &internal.Session{
		Key:       "golden-multi-fc",
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, store.Create(sess))

	adapter := NewSessionAdapter(sess, store, "lango-agent")
	svc := NewSessionServiceAdapter(store, "lango-agent")

	// Save event with two FunctionCalls
	evt := &session.Event{
		Timestamp: time.Now(),
		Author:    "lango-agent",
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{
				Role: "model",
				Parts: []*genai.Part{
					{
						FunctionCall: &genai.FunctionCall{
							ID:   "call_a",
							Name: "exec",
							Args: map[string]any{"cmd": "ls"},
						},
					},
					{
						FunctionCall: &genai.FunctionCall{
							ID:   "call_b",
							Name: "read_file",
							Args: map[string]any{"path": "/tmp/file.txt"},
						},
						Thought:          true,
						ThoughtSignature: []byte("sig-b"),
					},
				},
			},
		},
	}
	require.NoError(t, svc.AppendEvent(context.Background(), adapter, evt))

	// Verify stored
	require.Len(t, adapter.sess.History, 1)
	msg := adapter.sess.History[0]
	require.Len(t, msg.ToolCalls, 2)
	assert.Equal(t, "call_a", msg.ToolCalls[0].ID)
	assert.Equal(t, "call_b", msg.ToolCalls[1].ID)

	// Restore and verify
	events := adapter.Events()
	var restored []*session.Event
	for evt := range events.All() {
		restored = append(restored, evt)
	}
	require.Len(t, restored, 1)

	fcParts := 0
	for _, p := range restored[0].Content.Parts {
		if p.FunctionCall != nil {
			fcParts++
		}
	}
	assert.Equal(t, 2, fcParts, "expected 2 FunctionCall parts in restored event")

	// Verify second call's thought signature
	var secondFC *genai.Part
	for _, p := range restored[0].Content.Parts {
		if p.FunctionCall != nil && p.FunctionCall.Name == "read_file" {
			secondFC = p
			break
		}
	}
	require.NotNil(t, secondFC)
	assert.True(t, secondFC.Thought)
	assert.Equal(t, []byte("sig-b"), secondFC.ThoughtSignature)
}

// =============================================================================
// Test: FunctionResponse ID fallback (when ID is empty, uses "call_" + Name)
// =============================================================================

func TestGolden_FunctionResponseIDFallback(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	sess := &internal.Session{
		Key:       "golden-id-fallback",
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, store.Create(sess))

	adapter := NewSessionAdapter(sess, store, "lango-agent")
	svc := NewSessionServiceAdapter(store, "lango-agent")

	// FunctionResponse with empty ID
	evt := &session.Event{
		Timestamp: time.Now(),
		Author:    "tool",
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{
				Role: "function",
				Parts: []*genai.Part{{
					FunctionResponse: &genai.FunctionResponse{
						ID:       "", // empty
						Name:     "search",
						Response: map[string]any{"results": "found"},
					},
				}},
			},
		},
	}
	require.NoError(t, svc.AppendEvent(context.Background(), adapter, evt))

	// Verify fallback ID in store
	require.Len(t, adapter.sess.History, 1)
	tc := adapter.sess.History[0].ToolCalls[0]
	assert.Equal(t, "call_search", tc.ID, "expected fallback ID")
}

// =============================================================================
// Test: Args are JSON-round-tripped correctly (nested objects, arrays)
// =============================================================================

func TestGolden_ArgsJSONRoundTrip(t *testing.T) {
	t.Parallel()

	giveArgs := map[string]any{
		"simple":  "value",
		"number":  float64(42),
		"boolean": true,
		"nested": map[string]any{
			"inner_key": "inner_value",
			"inner_num": float64(3.14),
		},
		"array": []any{"a", "b", "c"},
	}

	store := newMockStore()
	sess := &internal.Session{
		Key:       "golden-args-json",
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, store.Create(sess))

	adapter := NewSessionAdapter(sess, store, "lango-agent")
	svc := NewSessionServiceAdapter(store, "lango-agent")

	evt := &session.Event{
		Timestamp: time.Now(),
		Author:    "lango-agent",
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{
				Role: "model",
				Parts: []*genai.Part{{
					FunctionCall: &genai.FunctionCall{
						ID:   "call_complex_args",
						Name: "complex_tool",
						Args: giveArgs,
					},
				}},
			},
		},
	}
	require.NoError(t, svc.AppendEvent(context.Background(), adapter, evt))

	// Restore
	events := adapter.Events()
	var restored []*session.Event
	for evt := range events.All() {
		restored = append(restored, evt)
	}
	require.Len(t, restored, 1)

	fc := restored[0].Content.Parts[0].FunctionCall
	require.NotNil(t, fc)

	// Verify args are semantically equal via JSON comparison
	giveJSON, _ := json.Marshal(giveArgs)
	gotJSON, _ := json.Marshal(fc.Args)
	assert.JSONEq(t, string(giveJSON), string(gotJSON), "Args not preserved through JSON round-trip")
}

// =============================================================================
// Test: Consecutive same-role events are merged
// =============================================================================

func TestGolden_ConsecutiveSameRoleMerge(t *testing.T) {
	t.Parallel()

	now := time.Now()
	sess := &internal.Session{
		Key: "golden-merge",
		History: []internal.Message{
			{Role: types.RoleAssistant, Content: "part1", Timestamp: now},
			{Role: types.RoleAssistant, Content: "part2", Timestamp: now.Add(time.Second)},
			{Role: types.RoleUser, Content: "question", Timestamp: now.Add(2 * time.Second)},
		},
	}

	adapter := NewSessionAdapter(sess, newMockStore(), "lango-agent")
	events := adapter.Events()

	var restored []*session.Event
	for evt := range events.All() {
		restored = append(restored, evt)
	}

	// Two consecutive assistant messages should be merged into one event
	require.Len(t, restored, 2, "expected 2 events after merging consecutive assistant messages")

	// Merged event should have both text parts
	assert.Equal(t, "assistant", restored[0].Content.Role)
	require.Len(t, restored[0].Content.Parts, 2)
	assert.Equal(t, "part1", restored[0].Content.Parts[0].Text)
	assert.Equal(t, "part2", restored[0].Content.Parts[1].Text)

	// User event unchanged
	assert.Equal(t, "user", restored[1].Content.Role)
	assert.Equal(t, "question", restored[1].Content.Parts[0].Text)
}
