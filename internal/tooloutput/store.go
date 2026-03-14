package tooloutput

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/logging"
)

// OutputStore is an in-memory TTL store for tool output.
// It implements lifecycle.Component for app lifecycle integration.
type OutputStore struct {
	mu      sync.RWMutex
	entries map[string]entry
	ttl     time.Duration
	stopCh  chan struct{}
}

type entry struct {
	toolName  string
	content   string
	createdAt time.Time
}

// NewOutputStore creates a store with the given TTL for entries.
func NewOutputStore(ttl time.Duration) *OutputStore {
	return &OutputStore{
		entries: make(map[string]entry),
		ttl:     ttl,
		stopCh:  make(chan struct{}),
	}
}

// Store saves content and returns a UUID reference.
func (s *OutputStore) Store(toolName, content string) string {
	ref := uuid.New().String()
	s.mu.Lock()
	s.entries[ref] = entry{
		toolName:  toolName,
		content:   content,
		createdAt: time.Now(),
	}
	s.mu.Unlock()
	logger().Debugw("stored tool output", "ref", ref, "tool", toolName, "bytes", len(content))
	return ref
}

// Get retrieves the full content by reference.
func (s *OutputStore) Get(ref string) (string, bool) {
	s.mu.RLock()
	e, ok := s.entries[ref]
	s.mu.RUnlock()
	return e.content, ok
}

// GetRange retrieves a line range. Returns (lines, totalLines, found).
func (s *OutputStore) GetRange(ref string, offset, limit int) (string, int, bool) {
	s.mu.RLock()
	e, ok := s.entries[ref]
	s.mu.RUnlock()
	if !ok {
		return "", 0, false
	}

	lines := strings.Split(e.content, "\n")
	total := len(lines)

	if offset >= total {
		return "", total, true
	}
	if offset < 0 {
		offset = 0
	}

	end := total
	if limit > 0 && offset+limit < total {
		end = offset + limit
	}

	return strings.Join(lines[offset:end], "\n"), total, true
}

// Grep searches content by regex pattern. Returns (matchingLines, found).
func (s *OutputStore) Grep(ref, pattern string) (string, bool) {
	s.mu.RLock()
	e, ok := s.entries[ref]
	s.mu.RUnlock()
	if !ok {
		return "", false
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		logger().Warnw("invalid grep pattern", "pattern", pattern, "error", err)
		return "", true
	}

	lines := strings.Split(e.content, "\n")
	matches := make([]string, 0, len(lines)/4)
	for _, line := range lines {
		if re.MatchString(line) {
			matches = append(matches, line)
		}
	}

	return strings.Join(matches, "\n"), true
}

// Name implements lifecycle.Component.
func (s *OutputStore) Name() string { return "output-store" }

// Start implements lifecycle.Component. Starts the cleanup goroutine.
func (s *OutputStore) Start(_ context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.cleanupLoop()
	}()
	logger().Infow("output store started", "ttl", s.ttl)
	return nil
}

// Stop implements lifecycle.Component. Stops the cleanup goroutine.
func (s *OutputStore) Stop(_ context.Context) error {
	close(s.stopCh)
	logger().Infow("output store stopped")
	return nil
}

func (s *OutputStore) cleanupLoop() {
	ticker := time.NewTicker(s.ttl / 2)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.evictExpired()
		}
	}
}

func (s *OutputStore) evictExpired() {
	now := time.Now()
	s.mu.Lock()
	for ref, e := range s.entries {
		if now.Sub(e.createdAt) > s.ttl {
			delete(s.entries, ref)
			logger().Debugw("evicted expired output", "ref", ref, "tool", e.toolName)
		}
	}
	s.mu.Unlock()
}

func logger() *zap.SugaredLogger {
	return logging.SubsystemSugar("tooloutput.store")
}
