package approval

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"sync"
)

type turnApprovalStateCtxKey struct{}

type TurnOutcome string

const (
	TurnOutcomeApproved    TurnOutcome = "approved"
	TurnOutcomeDenied      TurnOutcome = "denied"
	TurnOutcomeTimeout     TurnOutcome = "timeout"
	TurnOutcomeUnavailable TurnOutcome = "unavailable"
)

// MaxTurnApprovalTimeouts bounds how many approval timeouts the same
// canonical action can accumulate in a single turn before later retries are
// blocked without opening another approval prompt.
const MaxTurnApprovalTimeouts = 3

type TurnApprovalEntry struct {
	Outcome    TurnOutcome
	Provider   string
	RequestID  string
	Summary    string
	ParamsHash string
	Timeouts   int
}

type TurnApprovalState struct {
	mu      sync.RWMutex
	entries map[string]TurnApprovalEntry
}

func NewTurnApprovalState() *TurnApprovalState {
	return &TurnApprovalState{
		entries: make(map[string]TurnApprovalEntry),
	}
}

func WithTurnApprovalState(ctx context.Context, state *TurnApprovalState) context.Context {
	return context.WithValue(ctx, turnApprovalStateCtxKey{}, state)
}

func TurnApprovalStateFromContext(ctx context.Context) *TurnApprovalState {
	if state, ok := ctx.Value(turnApprovalStateCtxKey{}).(*TurnApprovalState); ok {
		return state
	}
	return nil
}

func (s *TurnApprovalState) Get(toolName string, params map[string]interface{}) (TurnApprovalEntry, bool, error) {
	key, hash, err := TurnApprovalKey(toolName, params)
	if err != nil {
		return TurnApprovalEntry{}, false, err
	}

	s.mu.RLock()
	entry, ok := s.entries[key]
	s.mu.RUnlock()
	if !ok {
		return TurnApprovalEntry{}, false, nil
	}
	if entry.ParamsHash == "" {
		entry.ParamsHash = hash
	}
	return entry, true, nil
}

func (s *TurnApprovalState) Put(toolName string, params map[string]interface{}, entry TurnApprovalEntry) error {
	key, hash, err := TurnApprovalKey(toolName, params)
	if err != nil {
		return err
	}
	if entry.ParamsHash == "" {
		entry.ParamsHash = hash
	}

	s.mu.Lock()
	s.entries[key] = entry
	s.mu.Unlock()
	return nil
}

func TurnApprovalKey(toolName string, params map[string]interface{}) (string, string, error) {
	data, err := json.Marshal(canonicalTurnApprovalParams(toolName, params))
	if err != nil {
		return "", "", err
	}
	sum := sha256.Sum256(append([]byte(toolName+"\x00"), data...))
	hash := hex.EncodeToString(sum[:])
	return toolName + "\x00" + string(data), hash, nil
}

func OutcomeFromError(err error) TurnOutcome {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrDenied):
		return TurnOutcomeDenied
	case errors.Is(err, ErrTimeout):
		return TurnOutcomeTimeout
	case errors.Is(err, ErrUnavailable):
		return TurnOutcomeUnavailable
	default:
		return ""
	}
}

func canonicalTurnApprovalParams(toolName string, params map[string]interface{}) map[string]interface{} {
	switch toolName {
	case "browser_search":
		return map[string]interface{}{
			"query": normalizeApprovalString(params["query"]),
		}
	default:
		return params
	}
}

func normalizeApprovalString(raw interface{}) string {
	s, _ := raw.(string)
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}
