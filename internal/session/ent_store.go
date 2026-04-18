package session

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
	"unsafe"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/message"
	entschema "github.com/langoai/lango/internal/ent/schema"
	entsession "github.com/langoai/lango/internal/ent/session"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/sqlitedriver"
	"github.com/langoai/lango/internal/types"
)

// StoreOption defines the functional option pattern for EntStore
type StoreOption func(*EntStore)

// WithPassphrase sets the encryption passphrase for the database.
func WithPassphrase(passphrase string) StoreOption {
	return func(s *EntStore) {
		s.passphrase = passphrase
	}
}

// WithMaxHistoryTurns limits the number of messages kept per session.
func WithMaxHistoryTurns(n int) StoreOption {
	return func(s *EntStore) {
		s.maxHistoryTurns = n
	}
}

// WithTTL sets the session time-to-live.
func WithTTL(d time.Duration) StoreOption {
	return func(s *EntStore) {
		s.ttl = d
	}
}

// WithPayloadProtector enables broker-managed payload protection for session
// message content and tool-call payloads.
func WithPayloadProtector(protector security.PayloadProtector) StoreOption {
	return func(s *EntStore) {
		s.payloads = protector
	}
}

// EntStore implements Store using entgo.io
type EntStore struct {
	client          *ent.Client
	db              *sql.DB
	mu              sync.RWMutex
	passphrase      string
	payloads        security.PayloadProtector
	maxHistoryTurns int
	ttl             time.Duration
	endProcessor    SessionEndProcessor
	hardEndTimeout  time.Duration
}

// NewEntStore creates a new ent-backed session store
func NewEntStore(dbPath string, opts ...StoreOption) (*EntStore, error) {
	store := &EntStore{}
	for _, opt := range opts {
		opt(store)
	}

	dbPath = sqlitedriver.ExpandPath(dbPath)
	if err := sqlitedriver.CheckFileHeader(dbPath); err != nil {
		return nil, err
	}

	db, err := sqlitedriver.Open(dbPath, false)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := sqlitedriver.ConfigureConnection(db, false); err != nil {
		db.Close()
		return nil, err
	}

	// Set key immediately if provided (essential for SQLCipher)
	// Use hex-encoded key to avoid SQL injection via passphrase content.
	// SQLCipher accepts: PRAGMA key = "x'HEX_ENCODED_KEY'"
	if store.passphrase != "" {
		hexKey := hex.EncodeToString([]byte(store.passphrase))
		pragma := fmt.Sprintf(`PRAGMA key = "x'%s'"`, hexKey)
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("set encryption key: %w", err)
		}
	}

	// Check connectivity and enable foreign keys
	// This will fail if the DB is encrypted and key wasn't accepted, OR if file path is invalid
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys/unlock db: %w", err)
	}

	// Create ent driver with SQLite dialect
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))

	// Auto-migrate schema - skip FK check since we've enabled it manually
	if err := client.Schema.Create(context.Background(), schema.WithForeignKeys(false)); err != nil {
		client.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}

	store.client = client
	store.db = db
	return store, nil
}

// NewEntStoreWithClient creates a new ent-backed session store using an
// existing ent.Client. This avoids opening a second database connection when
// the client is already available (e.g., from the bootstrap process).
// Schema migration is assumed to be already complete.
func NewEntStoreWithClient(client *ent.Client, opts ...StoreOption) *EntStore {
	store := &EntStore{client: client}
	for _, opt := range opts {
		opt(store)
	}
	return store
}

// Client returns the ent client
func (s *EntStore) Client() *ent.Client {
	return s.client
}

// DB returns the underlying *sql.DB. Useful for sidecar consumers (FTS5
// indexes, custom SQL) that live in the same database as sessions.
func (s *EntStore) DB() *sql.DB {
	return s.db
}

func (s *EntStore) SetPayloadProtector(protector security.PayloadProtector) {
	s.payloads = protector
}

// Create creates a new session
func (s *EntStore) Create(session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()
	now := time.Now()
	session.CreatedAt = now
	session.UpdatedAt = now

	// Create session entity
	builder := s.client.Session.Create().
		SetKey(session.Key).
		SetCreatedAt(now).
		SetUpdatedAt(now)

	if session.AgentID != "" {
		builder.SetAgentID(session.AgentID)
	}
	if session.ChannelType != "" {
		builder.SetChannelType(session.ChannelType)
	}
	if session.ChannelID != "" {
		builder.SetChannelID(session.ChannelID)
	}
	if session.Model != "" {
		builder.SetModel(session.Model)
	}
	if session.Metadata != nil {
		builder.SetMetadata(session.Metadata)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return fmt.Errorf("create session %q: %w", session.Key, ErrDuplicateSession)
		}
		return fmt.Errorf("create session %q: %w", session.Key, err)
	}

	// Create messages if any
	for _, msg := range session.History {
		msgBuilder, err := s.messageCreateBuilder(created, msg)
		if err != nil {
			return fmt.Errorf("prepare message: %w", err)
		}
		if _, err := msgBuilder.Save(ctx); err != nil {
			return fmt.Errorf("create message: %w", err)
		}
	}

	return nil
}

// Get retrieves a session by key
func (s *EntStore) Get(key string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ctx := context.Background()

	entSession, err := s.client.Session.
		Query().
		Where(entsession.Key(key)).
		WithMessages(func(q *ent.MessageQuery) {
			q.Order(message.ByTimestamp())
		}).
		Only(ctx)

	if ent.IsNotFound(err) {
		return nil, fmt.Errorf("get session %q: %w", key, ErrSessionNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("get session %q: %w", key, err)
	}

	// Check TTL
	if s.ttl > 0 && time.Since(entSession.UpdatedAt) > s.ttl {
		return nil, fmt.Errorf("get session %q: %w", key, ErrSessionExpired)
	}

	session, err := s.entToSession(entSession)
	if err != nil {
		return nil, fmt.Errorf("decode session %q: %w", key, err)
	}
	return session, nil
}

// Update updates an existing session
func (s *EntStore) Update(session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()
	now := time.Now()
	session.UpdatedAt = now

	// Find existing session
	entSession, err := s.client.Session.
		Query().
		Where(entsession.Key(session.Key)).
		Only(ctx)

	if ent.IsNotFound(err) {
		return fmt.Errorf("update session %q: %w", session.Key, ErrSessionNotFound)
	}
	if err != nil {
		return fmt.Errorf("update session %q: %w", session.Key, err)
	}

	// Update session
	builder := entSession.Update().SetUpdatedAt(now)

	if session.AgentID != "" {
		builder.SetAgentID(session.AgentID)
	}
	if session.ChannelType != "" {
		builder.SetChannelType(session.ChannelType)
	}
	if session.ChannelID != "" {
		builder.SetChannelID(session.ChannelID)
	}
	if session.Model != "" {
		builder.SetModel(session.Model)
	}
	if session.Metadata != nil {
		builder.SetMetadata(session.Metadata)
	}

	_, err = builder.Save(ctx)
	return err
}

// Delete removes a session
func (s *EntStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()

	// Delete messages first (cascade not automatic)
	entSession, err := s.client.Session.
		Query().
		Where(entsession.Key(key)).
		Only(ctx)

	if ent.IsNotFound(err) {
		return nil // Already deleted
	}
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}

	// Delete all messages
	_, err = s.client.Message.Delete().
		Where(message.HasSessionWith(entsession.ID(entSession.ID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete messages: %w", err)
	}

	// Delete session
	return s.client.Session.DeleteOne(entSession).Exec(ctx)
}

// AppendMessage adds a message to session history
func (s *EntStore) AppendMessage(key string, msg Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()

	// Get session
	entSession, err := s.client.Session.
		Query().
		Where(entsession.Key(key)).
		Only(ctx)

	if ent.IsNotFound(err) {
		return fmt.Errorf("append message to session %q: %w", key, ErrSessionNotFound)
	}
	if err != nil {
		return fmt.Errorf("append message to session %q: %w", key, err)
	}

	// Create message
	timestamp := msg.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	msg.Timestamp = timestamp
	msgBuilder, err := s.messageCreateBuilder(entSession, msg)
	if err != nil {
		return fmt.Errorf("prepare message: %w", err)
	}
	_, err = msgBuilder.Save(ctx)

	if err != nil {
		return fmt.Errorf("create message: %w", err)
	}

	// Update session timestamp
	_, err = entSession.Update().SetUpdatedAt(time.Now()).Save(ctx)
	if err != nil {
		return err
	}

	// Trim excess messages if maxHistoryTurns is configured
	if s.maxHistoryTurns > 0 {
		msgCount, err := s.client.Message.Query().
			Where(message.HasSessionWith(entsession.Key(key))).
			Count(ctx)
		if err == nil && msgCount > s.maxHistoryTurns {
			// Get IDs of oldest messages to delete
			oldest, err := s.client.Message.Query().
				Where(message.HasSessionWith(entsession.Key(key))).
				Order(message.ByTimestamp()).
				Limit(msgCount - s.maxHistoryTurns).
				IDs(ctx)
			if err == nil && len(oldest) > 0 {
				_, _ = s.client.Message.Delete().
					Where(message.IDIn(oldest...)).
					Exec(ctx)
			}
		}
	}

	return nil
}

// AnnotateTimeout appends a synthetic assistant message to mark a timed-out turn.
func (s *EntStore) AnnotateTimeout(key string, partial string) error {
	content := "[This response was interrupted due to a timeout]"

	return s.AppendMessage(key, Message{
		Role:      "assistant",
		Content:   content,
		Timestamp: time.Now(),
	})
}

// CompactMessages replaces messages up to (and including) upToIndex with a
// single summary message. This achieves compaction: the original messages are
// removed and replaced by a condensed version, preserving recent context.
func (s *EntStore) CompactMessages(key string, upToIndex int, summary string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()

	// Get session
	entSession, err := s.client.Session.
		Query().
		Where(entsession.Key(key)).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("get session %q: %w", key, err)
	}

	// Get ordered messages to identify which ones to compact
	messages, err := s.client.Message.
		Query().
		Where(message.HasSessionWith(entsession.Key(key))).
		Order(message.ByTimestamp()).
		All(ctx)
	if err != nil {
		return fmt.Errorf("list messages: %w", err)
	}

	if upToIndex >= len(messages) || upToIndex < 0 {
		return fmt.Errorf("compact index %d out of range (have %d messages)", upToIndex, len(messages))
	}

	// Collect IDs of messages to delete (0..upToIndex inclusive)
	toDelete := make([]int, 0, upToIndex+1)
	for i := 0; i <= upToIndex; i++ {
		toDelete = append(toDelete, messages[i].ID)
	}

	// Delete old messages in batch
	_, err = s.client.Message.Delete().
		Where(message.IDIn(toDelete...)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete compacted messages: %w", err)
	}

	// Insert summary message at the beginning (with early timestamp)
	msgBuilder, err := s.messageCreateBuilder(entSession, Message{
		Role:      "system",
		Content:   "[Compacted Summary]\n" + summary,
		Timestamp: time.Now().Add(-24 * time.Hour), // ensure it sorts before recent messages
	})
	if err != nil {
		return fmt.Errorf("prepare summary message: %w", err)
	}
	_, err = msgBuilder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create summary message: %w", err)
	}

	return nil
}

// ListSessions returns lightweight summaries of all sessions,
// ordered by most recent update first.
func (s *EntStore) ListSessions(ctx context.Context) ([]SessionSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.client.Session.Query().
		Order(entsession.ByUpdatedAt(entsql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	summaries := make([]SessionSummary, len(rows))
	for i, r := range rows {
		summaries[i] = SessionSummary{
			Key:       r.Key,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		}
	}
	return summaries, nil
}

// Close closes the ent client and underlying database connection.
// When the client was provided externally via NewEntStoreWithClient, only the
// ent client is closed; the raw DB connection is managed by the caller.
func (s *EntStore) Close() error {
	if s.client == nil {
		return nil
	}
	return s.client.Close()
}

// entToSession converts ent Session to domain Session
func (s *EntStore) entToSession(e *ent.Session) (*Session, error) {
	session := &Session{
		Key:         e.Key,
		AgentID:     e.AgentID,
		ChannelType: e.ChannelType,
		ChannelID:   e.ChannelID,
		Model:       e.Model,
		Metadata:    e.Metadata,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
		History:     make([]Message, 0, len(e.Edges.Messages)),
	}

	for _, m := range e.Edges.Messages {
		content, err := s.resolveMessageContent(m)
		if err != nil {
			return nil, err
		}
		toolCalls, err := s.resolveMessageToolCalls(m)
		if err != nil {
			return nil, err
		}

		session.History = append(session.History, Message{
			Role:      types.MessageRole(m.Role),
			Content:   content,
			Timestamp: m.Timestamp,
			ToolCalls: toolCalls,
			Author:    m.Author,
		})
	}

	return session, nil
}

func (s *EntStore) messageCreateBuilder(sessionEntity *ent.Session, msg Message) (*ent.MessageCreate, error) {
	projection, ciphertext, nonce, keyVersion, err := security.ProtectText(s.payloads, msg.Content, 512)
	if err != nil {
		return nil, err
	}
	toolCalls, toolCiphertext, toolNonce, toolKeyVersion, err := s.prepareToolCallPayload(msg.ToolCalls)
	if err != nil {
		return nil, err
	}

	builder := s.client.Message.Create().
		SetSession(sessionEntity).
		SetRole(string(msg.Role)).
		SetContent(projection).
		SetTimestamp(msg.Timestamp).
		SetToolCalls(toolCalls)
	if ciphertext != nil {
		builder.SetContentCiphertext(ciphertext)
		builder.SetContentNonce(nonce)
		builder.SetContentKeyVersion(keyVersion)
	}
	if toolCiphertext != nil {
		builder.SetToolCallsCiphertext(toolCiphertext)
		builder.SetToolCallsNonce(toolNonce)
		builder.SetToolCallsKeyVersion(toolKeyVersion)
	}
	if msg.Author != "" {
		builder.SetAuthor(msg.Author)
	}
	return builder, nil
}

func (s *EntStore) prepareToolCallPayload(toolCalls []ToolCall) ([]entschema.ToolCall, []byte, []byte, int, error) {
	projection := make([]entschema.ToolCall, len(toolCalls))
	if len(toolCalls) == 0 {
		return projection, nil, nil, 0, nil
	}
	for i, tc := range toolCalls {
		projection[i] = entschema.ToolCall{
			ID:               tc.ID,
			Name:             tc.Name,
			ThoughtSignature: tc.ThoughtSignature,
		}
	}
	if s.payloads == nil {
		for i, tc := range toolCalls {
			projection[i].Input = tc.Input
			projection[i].Output = tc.Output
			projection[i].Thought = tc.Thought
		}
		return projection, nil, nil, 0, nil
	}
	ciphertext, nonce, keyVersion, err := security.ProtectJSONBundle(s.payloads, toolCalls)
	if err != nil {
		return nil, nil, nil, 0, err
	}
	return projection, ciphertext, nonce, keyVersion, nil
}

func (s *EntStore) resolveMessageContent(m *ent.Message) (string, error) {
	if m == nil {
		return "", nil
	}
	if s.payloads == nil || m.ContentCiphertext == nil || m.ContentNonce == nil || m.ContentKeyVersion == nil {
		return m.Content, nil
	}
	plaintext, err := s.payloads.DecryptPayload(*m.ContentCiphertext, *m.ContentNonce, *m.ContentKeyVersion)
	if err != nil {
		return "", fmt.Errorf("decrypt message content: %w", err)
	}
	return string(plaintext), nil
}

func (s *EntStore) resolveMessageToolCalls(m *ent.Message) ([]ToolCall, error) {
	if m == nil {
		return nil, nil
	}
	if s.payloads == nil || m.ToolCallsCiphertext == nil || m.ToolCallsNonce == nil || m.ToolCallsKeyVersion == nil {
		return convertSessionToolCalls(m.ToolCalls), nil
	}
	decrypted, err := security.UnprotectJSONBundle[[]ToolCall](s.payloads, *m.ToolCallsCiphertext, *m.ToolCallsNonce, *m.ToolCallsKeyVersion)
	if err != nil {
		return nil, fmt.Errorf("decrypt message tool calls: %w", err)
	}
	return decrypted, nil
}

func convertSessionToolCalls(in []entschema.ToolCall) []ToolCall {
	out := make([]ToolCall, len(in))
	for i, tc := range in {
		out[i] = ToolCall{
			ID:               tc.ID,
			Name:             tc.Name,
			Input:            tc.Input,
			Output:           tc.Output,
			Thought:          tc.Thought,
			ThoughtSignature: tc.ThoughtSignature,
		}
	}
	return out
}

// GetSalt retrieves the encryption salt by name.
// Delegates to SecurityConfigStore for unified access.
func (s *EntStore) GetSalt(name string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	store := security.NewSecurityConfigStore(s.db)
	salt, err := store.LoadSaltNamed(name)
	if err != nil {
		return nil, err
	}
	if salt == nil {
		return nil, fmt.Errorf("salt not found: %s", name)
	}
	return salt, nil
}

// SetSalt stores the encryption salt by name.
func (s *EntStore) SetSalt(name string, salt []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return security.NewSecurityConfigStore(s.db).StoreSaltNamed(name, salt)
}

// GetChecksum retrieves the passphrase checksum by name.
func (s *EntStore) GetChecksum(name string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	store := security.NewSecurityConfigStore(s.db)
	sum, err := store.LoadChecksumNamed(name)
	if err != nil {
		return nil, err
	}
	if sum == nil {
		// Distinguish "row missing" from "checksum null" by re-reading salt.
		salt, saltErr := store.LoadSaltNamed(name)
		if saltErr == nil && salt == nil {
			return nil, fmt.Errorf("checksum not found: %s", name)
		}
		return nil, fmt.Errorf("checksum not set for: %s", name)
	}
	return sum, nil
}

// SetChecksum stores the passphrase checksum by name.
func (s *EntStore) SetChecksum(name string, checksum []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := security.NewSecurityConfigStore(s.db).StoreChecksumNamed(name, checksum)
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("cannot set checksum: salt entry '%s' does not exist", name)
	}
	return nil
}

// MigrateSecrets performs the secret migration using callbacks to avoid import cycles.
// reencryptFn typically decrypts with old key and encrypts with new key.
func (s *EntStore) MigrateSecrets(ctx context.Context, reencryptFn func([]byte) ([]byte, error), newSalt, newChecksum []byte) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Start Ent Transaction
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 2. Iterate and Re-encrypt Secrets
	secrets, err := tx.Secret.Query().All(ctx)
	if err != nil {
		return fmt.Errorf("query secrets: %w", err)
	}

	for _, sec := range secrets {
		newVal, err := reencryptFn(sec.EncryptedValue)
		if err != nil {
			return fmt.Errorf("re-encrypt secret %s: %w", sec.Name, err)
		}

		if _, err := tx.Secret.UpdateOne(sec).SetEncryptedValue(newVal).Save(ctx); err != nil {
			return fmt.Errorf("update secret %s: %w", sec.Name, err)
		}
	}

	// 3. Update Salt & Checksum using Raw SQL via underlying driver
	// Access the driver using reflection as it is not exposed by ent.Tx
	// tx.config.driver is the txDriver.

	// Get tx value
	v := reflect.ValueOf(tx).Elem()
	cfgField := v.FieldByName("config")
	drvField := cfgField.FieldByName("driver")

	// Access unexported field
	drvField = reflect.NewAt(drvField.Type(), unsafe.Pointer(drvField.UnsafeAddr())).Elem()

	drv, ok := drvField.Interface().(dialect.Driver)
	if !ok {
		return fmt.Errorf("resolve transaction driver")
	}

	// Exec Raw SQL
	// Update Salt
	err = drv.Exec(ctx, `INSERT OR REPLACE INTO security_config (name, value) VALUES (?, ?)`, []interface{}{"default", newSalt}, nil)
	if err != nil {
		return fmt.Errorf("update salt: %w", err)
	}

	// Update Checksum
	err = drv.Exec(ctx, `UPDATE security_config SET checksum = ? WHERE name = ?`, []interface{}{newChecksum, "default"}, nil)
	if err != nil {
		return fmt.Errorf("update checksum: %w", err)
	}

	// 4. Commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
