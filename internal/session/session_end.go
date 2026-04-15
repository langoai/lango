package session

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	entsession "github.com/langoai/lango/internal/ent/session"
)

// SessionEndProcessor is invoked when a session ends (hard or soft path).
// Implementations should be idempotent — the same key may be processed more
// than once across sweeps. An error return is logged but does not prevent the
// pending flag from staying set; the next sweep will retry.
type SessionEndProcessor func(ctx context.Context, key string) error

// SetSessionEndProcessor registers the processor used by End() and by
// ProcessEndPending sweeps. Passing nil clears the processor.
func (s *EntStore) SetSessionEndProcessor(fn SessionEndProcessor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.endProcessor = fn
}

// End implements Store.End. It marks the session with
// MetadataKeyEndPending=true and, if a processor is registered, invokes it
// synchronously bounded by hardEndTimeout (default 3s). If the processor
// succeeds within the bound, the pending flag is cleared. On timeout or
// error, the flag stays set so the next sweep can retry.
func (s *EntStore) End(key string) error {
	if err := s.MarkEndPending(key); err != nil {
		return err
	}

	processor := s.snapshotProcessor()
	if processor == nil {
		return nil
	}

	timeout := s.hardEndTimeout
	if timeout == 0 {
		timeout = 3 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- processor(ctx, key)
	}()

	select {
	case err := <-done:
		if err != nil {
			slog.Warn("session end processor failed", "key", key, "error", err)
			return nil
		}
		if clearErr := s.ClearEndPending(key); clearErr != nil {
			slog.Warn("session clear end-pending failed", "key", key, "error", clearErr)
		}
		return nil
	case <-ctx.Done():
		slog.Warn("session end processor timeout", "key", key, "timeout", timeout)
		return nil
	}
}

// SetHardEndTimeout overrides the default 3s synchronous bound for End().
func (s *EntStore) SetHardEndTimeout(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hardEndTimeout = d
}

// MarkEndPending sets MetadataKeyEndPending=true on the session's metadata.
// Returns ErrSessionNotFound if the session does not exist.
func (s *EntStore) MarkEndPending(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()
	return s.setMetadataKey(ctx, key, MetadataKeyEndPending, MetadataValueTrue)
}

// ClearEndPending removes MetadataKeyEndPending from the session metadata.
// A no-op if the session does not have the flag set.
func (s *EntStore) ClearEndPending(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()
	return s.deleteMetadataKey(ctx, key, MetadataKeyEndPending)
}

// ListEndPending returns the keys of all sessions currently marked with
// MetadataKeyEndPending=true, ordered by most recent update first.
func (s *EntStore) ListEndPending() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ctx := context.Background()
	rows, err := s.client.Session.
		Query().
		Order(entsession.ByUpdatedAt()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	var out []string
	for _, r := range rows {
		if r.Metadata != nil && r.Metadata[MetadataKeyEndPending] == MetadataValueTrue {
			out = append(out, r.Key)
		}
	}
	return out, nil
}

// snapshotProcessor returns the current processor under read lock.
func (s *EntStore) snapshotProcessor() SessionEndProcessor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.endProcessor
}

// setMetadataKey upserts a metadata key on the session. Caller must hold s.mu.
func (s *EntStore) setMetadataKey(ctx context.Context, key, metaKey, metaValue string) error {
	row, err := s.client.Session.
		Query().
		Where(entsession.Key(key)).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("get session %q: %w", key, err)
	}
	meta := row.Metadata
	if meta == nil {
		meta = make(map[string]string)
	}
	if meta[metaKey] == metaValue {
		return nil
	}
	meta[metaKey] = metaValue
	_, err = row.Update().SetMetadata(meta).SetUpdatedAt(time.Now()).Save(ctx)
	if err != nil {
		return fmt.Errorf("update session %q metadata: %w", key, err)
	}
	return nil
}

// deleteMetadataKey removes a metadata key. Caller must hold s.mu.
func (s *EntStore) deleteMetadataKey(ctx context.Context, key, metaKey string) error {
	row, err := s.client.Session.
		Query().
		Where(entsession.Key(key)).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("get session %q: %w", key, err)
	}
	if row.Metadata == nil {
		return nil
	}
	if _, ok := row.Metadata[metaKey]; !ok {
		return nil
	}
	delete(row.Metadata, metaKey)
	_, err = row.Update().SetMetadata(row.Metadata).SetUpdatedAt(time.Now()).Save(ctx)
	if err != nil {
		return fmt.Errorf("update session %q metadata: %w", key, err)
	}
	return nil
}
