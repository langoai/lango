package browser

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrSearchLimitReached is returned when browser_search exceeds MaxSearchesPerRequest.
// The error message instructs the model to stop searching and use existing results.
var ErrSearchLimitReached = errors.New(
	"browser_search is no longer available: search limit (2) reached. " +
		"Present your existing results to the user. Do NOT call browser_search again.",
)

type requestStateCtxKey struct{}

// MaxSearchesPerRequest is the hard limit on browser_search calls per request.
// The agent gets one initial search plus one reformulation (total 2).
// The 3rd attempt is blocked with a structured stop response.
const MaxSearchesPerRequest = 2

// RequestState tracks browser-search churn for a single agent request.
type RequestState struct {
	ID string

	mu          sync.Mutex
	searchCount int
	queries     []string
	currentURL  string
	warned      bool
}

func NewRequestState() *RequestState {
	return &RequestState{
		ID: "browser-req-" + time.Now().Format("150405.000000000"),
	}
}

func WithRequestState(ctx context.Context, state *RequestState) context.Context {
	return context.WithValue(ctx, requestStateCtxKey{}, state)
}

func RequestStateFromContext(ctx context.Context) *RequestState {
	if state, ok := ctx.Value(requestStateCtxKey{}).(*RequestState); ok {
		return state
	}
	return nil
}

func (s *RequestState) RecordSearch(query, currentURL string) (count int, queries []string, shouldWarn bool, limitReached bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.searchCount++
	s.queries = append(s.queries, query)
	if currentURL != "" {
		s.currentURL = currentURL
	}

	if s.searchCount > MaxSearchesPerRequest {
		limitReached = true
	}

	if s.searchCount >= 3 && !s.warned {
		s.warned = true
		shouldWarn = true
	}

	queries = append([]string(nil), s.queries...)
	return s.searchCount, queries, shouldWarn, limitReached
}

func (s *RequestState) CurrentURL() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentURL
}

// IsLimitReached reports whether the search limit has already been reached
// without recording a new search attempt.
func (s *RequestState) IsLimitReached() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.searchCount > MaxSearchesPerRequest
}
