package storage

import (
	"context"
	"time"

	"github.com/langoai/lango/internal/search"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/storagebroker"
)

type brokerSessionStore struct {
	broker         storagebroker.API
	endProcessor   session.SessionEndProcessor
	hardEndTimeout time.Duration
}

func NewBrokerSessionStore(broker storagebroker.API, _ ...session.StoreOption) session.Store {
	if broker == nil {
		return nil
	}
	return &brokerSessionStore{broker: broker}
}

func (s *brokerSessionStore) Create(sess *session.Session) error {
	return s.broker.SessionCreate(context.Background(), sess)
}

func (s *brokerSessionStore) Get(key string) (*session.Session, error) {
	return s.broker.SessionGet(context.Background(), key)
}

func (s *brokerSessionStore) Update(sess *session.Session) error {
	return s.broker.SessionUpdate(context.Background(), sess)
}

func (s *brokerSessionStore) Delete(key string) error {
	return s.broker.SessionDelete(context.Background(), key)
}

func (s *brokerSessionStore) AppendMessage(key string, msg session.Message) error {
	return s.broker.SessionAppendMessage(context.Background(), key, msg)
}

func (s *brokerSessionStore) AnnotateTimeout(key string, partial string) error {
	content := "[This response was interrupted due to a timeout]"
	return s.AppendMessage(key, session.Message{
		Role:      "assistant",
		Content:   content,
		Timestamp: time.Now(),
	})
}

func (s *brokerSessionStore) End(key string) error {
	if err := s.broker.SessionEnd(context.Background(), key); err != nil {
		return err
	}
	if s.endProcessor == nil {
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
		done <- s.endProcessor(ctx, key)
	}()
	select {
	case <-ctx.Done():
		return nil
	case err := <-done:
		return err
	}
}

func (s *brokerSessionStore) Close() error { return nil }

func (s *brokerSessionStore) ListSessions(ctx context.Context) ([]session.SessionSummary, error) {
	return s.broker.SessionList(ctx)
}

func (s *brokerSessionStore) GetSalt(name string) ([]byte, error) {
	return s.broker.SessionGetSalt(context.Background(), name)
}

func (s *brokerSessionStore) SetSalt(name string, salt []byte) error {
	return s.broker.SessionSetSalt(context.Background(), name, salt)
}

func (s *brokerSessionStore) SetSessionEndProcessor(fn session.SessionEndProcessor) {
	s.endProcessor = fn
}

func (s *brokerSessionStore) SetHardEndTimeout(d time.Duration) {
	s.hardEndTimeout = d
}

func (s *brokerSessionStore) IndexSession(ctx context.Context, key string) error {
	return s.broker.RecallIndexSession(ctx, key)
}

func (s *brokerSessionStore) ProcessPending(ctx context.Context) error {
	return s.broker.RecallProcessPending(ctx)
}

func (s *brokerSessionStore) Search(ctx context.Context, query string, limit int) ([]search.SearchResult, error) {
	return s.broker.RecallSearch(ctx, query, limit)
}

func (s *brokerSessionStore) GetSummary(ctx context.Context, key string) (string, error) {
	return s.broker.RecallGetSummary(ctx, key)
}
