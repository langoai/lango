package session

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
	"unsafe"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/langowarny/lango/internal/ent"
	"github.com/langowarny/lango/internal/ent/message"
	entschema "github.com/langowarny/lango/internal/ent/schema"
	entsession "github.com/langowarny/lango/internal/ent/session"
	"github.com/langowarny/lango/internal/logging"
	_ "github.com/mattn/go-sqlite3" // Use cgo driver for SQLCipher support
)

var logger = logging.SubsystemSugar("session")

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

// EntStore implements Store using entgo.io
type EntStore struct {
	client          *ent.Client
	db              *sql.DB
	mu              sync.RWMutex
	passphrase      string
	maxHistoryTurns int
	ttl             time.Duration
}

// NewEntStore creates a new ent-backed session store
func NewEntStore(dbPath string, opts ...StoreOption) (*EntStore, error) {
	store := &EntStore{}
	for _, opt := range opts {
		opt(store)
	}

	// Expand tilde in path if present
	if strings.HasPrefix(dbPath, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			dbPath = filepath.Join(home, dbPath[2:])
		}
	}

	// Build connection string
	// We use "file:" prefix to ensure parameters are respected
	connStr := dbPath
	if !strings.HasPrefix(connStr, "file:") {
		connStr = "file:" + connStr
	}

	// Add parameters. Note: _fk=1 is standard but we'll enable it manually to be safe with encryption
	if !strings.Contains(connStr, "?") {
		connStr += "?cache=shared"
	} else {
		connStr += "&cache=shared"
	}

	// Open with sqlite3 driver (mattn/go-sqlite3)
	// modernc.org/sqlite registers as "sqlite", mattn/go-sqlite3 registers as "sqlite3"
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
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
		return nil, fmt.Errorf("failed to enable foreign keys/unlock db: %w", err)
	}

	// Create ent driver with SQLite dialect
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))

	// Auto-migrate schema - skip FK check since we've enabled it manually
	if err := client.Schema.Create(context.Background(), schema.WithForeignKeys(false)); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
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
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Create messages if any
	for _, msg := range session.History {
		toolCalls := make([]entschema.ToolCall, len(msg.ToolCalls))
		for i, tc := range msg.ToolCalls {
			toolCalls[i] = entschema.ToolCall{
				ID:     tc.ID,
				Name:   tc.Name,
				Input:  tc.Input,
				Output: tc.Output,
			}
		}

		_, err := s.client.Message.Create().
			SetSession(created).
			SetRole(msg.Role).
			SetContent(msg.Content).
			SetTimestamp(msg.Timestamp).
			SetToolCalls(toolCalls).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create message: %w", err)
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
		return nil, fmt.Errorf("session not found: %s", key)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Check TTL
	if s.ttl > 0 && time.Since(entSession.UpdatedAt) > s.ttl {
		return nil, fmt.Errorf("session expired: %s", key)
	}

	return s.entToSession(entSession), nil
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
		return fmt.Errorf("session not found: %s", session.Key)
	}
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
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
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Delete all messages
	_, err = s.client.Message.Delete().
		Where(message.HasSessionWith(entsession.ID(entSession.ID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
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
		return fmt.Errorf("session not found: %s", key)
	}
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Convert tool calls
	toolCalls := make([]entschema.ToolCall, len(msg.ToolCalls))
	for i, tc := range msg.ToolCalls {
		toolCalls[i] = entschema.ToolCall{
			ID:     tc.ID,
			Name:   tc.Name,
			Input:  tc.Input,
			Output: tc.Output,
		}
	}

	// Create message
	timestamp := msg.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	_, err = s.client.Message.Create().
		SetSession(entSession).
		SetRole(msg.Role).
		SetContent(msg.Content).
		SetTimestamp(timestamp).
		SetToolCalls(toolCalls).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
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
func (s *EntStore) entToSession(e *ent.Session) *Session {
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
		toolCalls := make([]ToolCall, len(m.ToolCalls))
		for i, tc := range m.ToolCalls {
			toolCalls[i] = ToolCall{
				ID:     tc.ID,
				Name:   tc.Name,
				Input:  tc.Input,
				Output: tc.Output,
			}
		}

		session.History = append(session.History, Message{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp,
			ToolCalls: toolCalls,
		})
	}

	return session
}

// ensureSecurityTable ensures the security_config table exists with correct schema
func (s *EntStore) ensureSecurityTable() error {
	// Create table if not exists with basic schema
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS security_config (
			name TEXT PRIMARY KEY,
			value BLOB NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create security_config table: %w", err)
	}

	// Check if checksum column exists
	// This is a simple migration check for SQLite
	var count int
	err = s.db.QueryRow(`
		SELECT count(*) FROM pragma_table_info('security_config') WHERE name='checksum'
	`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check table schema: %w", err)
	}

	if count == 0 {
		// Add checksum column
		_, err = s.db.Exec(`ALTER TABLE security_config ADD COLUMN checksum BLOB`)
		if err != nil {
			return fmt.Errorf("failed to add checksum column: %w", err)
		}
	}

	return nil
}

// GetSalt retrieves the encryption salt by name
func (s *EntStore) GetSalt(name string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureSecurityTable(); err != nil {
		return nil, err
	}

	var salt []byte
	err := s.db.QueryRow(`SELECT value FROM security_config WHERE name = ?`, name).Scan(&salt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("salt not found: %s", name)
		}
		return nil, fmt.Errorf("failed to query salt: %w", err)
	}

	return salt, nil
}

// SetSalt stores the encryption salt by name
func (s *EntStore) SetSalt(name string, salt []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureSecurityTable(); err != nil {
		return err
	}

	// Upsert salt (value column)
	// We use ON CONFLICT DO UPDATE to preserve checksum if it exists
	_, err := s.db.Exec(`
		INSERT INTO security_config (name, value) VALUES (?, ?)
		ON CONFLICT(name) DO UPDATE SET value=excluded.value
	`, name, salt)
	if err != nil {
		return fmt.Errorf("failed to store salt: %w", err)
	}

	return nil
}

// GetChecksum retrieves the passphrase checksum by name
func (s *EntStore) GetChecksum(name string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureSecurityTable(); err != nil {
		return nil, err
	}

	// Scan directly into interface{} specific for SQLite type handling
	var raw interface{}

	err := s.db.QueryRow(`SELECT checksum FROM security_config WHERE name = ?`, name).Scan(&raw)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("checksum not found: %s", name)
		}
		return nil, fmt.Errorf("failed to query checksum: %w", err)
	}

	if raw == nil {
		return nil, fmt.Errorf("checksum not set for: %s", name)
	}

	switch v := raw.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("unexpected type for checksum: %T", v)
	}
}

// SetChecksum stores the passphrase checksum by name
func (s *EntStore) SetChecksum(name string, checksum []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureSecurityTable(); err != nil {
		return err
	}

	// Update checksum
	// The row MUST exist (salt should be set first or together)
	// But let's support upsert just in case, though value (salt) cannot be NULL if inserted fresh?
	// Schema says value is NOT NULL.
	// So we assume row exists or we fail?
	// Or we insert with empty salt? No.
	// Task 2.5 says "Store checksum on first-time passphrase setup".
	// Typically we set Salt AND Checksum.
	// Let's allow updating just checksum if row exists.

	res, err := s.db.Exec(`
		UPDATE security_config SET checksum = ? WHERE name = ?
	`, checksum, name)
	if err != nil {
		return fmt.Errorf("failed to update checksum: %w", err)
	}

	rows, _ := res.RowsAffected()
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
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	// 2. Iterate and Re-encrypt Secrets
	secrets, err := tx.Secret.Query().All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query secrets: %w", err)
	}

	for _, sec := range secrets {
		newVal, err := reencryptFn(sec.EncryptedValue)
		if err != nil {
			return fmt.Errorf("failed to re-encrypt secret %s: %w", sec.Name, err)
		}

		if _, err := tx.Secret.UpdateOne(sec).SetEncryptedValue(newVal).Save(ctx); err != nil {
			return fmt.Errorf("failed to update secret %s: %w", sec.Name, err)
		}
	}

	// 3. Update Salt & Checksum using Raw SQL via underlying driver
	// Access the driver using reflection as it is not exposed by ent.Tx
	// tx.config.driver is the txDriver.

	// Get tx value
	v := reflect.ValueOf(tx).Elem()
	// Get config field (embedded)
	// config is struct, driver is field 0 usually? Or named "driver".
	// config has "driver" field.
	// We need to access "driver" field of "config" struct embedded in "Tx".

	// "Tx" struct:
	// type Tx struct {
	//    config
	//    ...
	// }
	// "config" struct:
	// type config struct {
	//    driver dialect.Driver
	//    ...
	// }

	// Note: config might be embedded via name "config" or anonymous?
	// In generated code: `type Tx struct { config ... }`. It is named field "config".

	cfgField := v.FieldByName("config")
	drvField := cfgField.FieldByName("driver")

	// Access unexported field
	drvField = reflect.NewAt(drvField.Type(), unsafe.Pointer(drvField.UnsafeAddr())).Elem()

	drv, ok := drvField.Interface().(dialect.Driver)
	if !ok {
		return fmt.Errorf("failed to resolve transaction driver")
	}

	// Exec Raw SQL
	// Update Salt
	err = drv.Exec(ctx, `INSERT OR REPLACE INTO security_config (name, value) VALUES (?, ?)`, []interface{}{"default", newSalt}, nil)
	if err != nil {
		return fmt.Errorf("failed to update salt: %w", err)
	}

	// Update Checksum
	err = drv.Exec(ctx, `UPDATE security_config SET checksum = ? WHERE name = ?`, []interface{}{newChecksum, "default"}, nil)
	if err != nil {
		return fmt.Errorf("failed to update checksum: %w", err)
	}

	// 4. Commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
