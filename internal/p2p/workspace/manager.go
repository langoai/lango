package workspace

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
)

var (
	bucketWorkspaces = []byte("workspaces")
	bucketMessages   = []byte("workspace_messages")
)

// ManagerConfig configures the WorkspaceManager.
type ManagerConfig struct {
	DB            *bolt.DB
	LocalDID      string
	MaxWorkspaces int
	Logger        *zap.SugaredLogger
}

// Manager manages workspace lifecycle with BoltDB persistence.
type Manager struct {
	db            *bolt.DB
	localDID      string
	maxWorkspaces int
	logger        *zap.SugaredLogger

	mu         sync.RWMutex
	workspaces map[string]*Workspace
}

// NewManager creates a new workspace manager.
func NewManager(cfg ManagerConfig) (*Manager, error) {
	if cfg.DB == nil {
		return nil, fmt.Errorf("BoltDB instance required")
	}
	if cfg.MaxWorkspaces <= 0 {
		cfg.MaxWorkspaces = 10
	}

	err := cfg.DB.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(bucketWorkspaces); err != nil {
			return fmt.Errorf("create workspaces bucket: %w", err)
		}
		if _, err := tx.CreateBucketIfNotExists(bucketMessages); err != nil {
			return fmt.Errorf("create messages bucket: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("init workspace buckets: %w", err)
	}

	m := &Manager{
		db:            cfg.DB,
		localDID:      cfg.LocalDID,
		maxWorkspaces: cfg.MaxWorkspaces,
		logger:        cfg.Logger,
		workspaces:    make(map[string]*Workspace),
	}

	// Load persisted workspaces into memory.
	if err := m.loadAll(); err != nil {
		return nil, fmt.Errorf("load workspaces: %w", err)
	}

	return m, nil
}

// Create creates a new workspace.
func (m *Manager) Create(ctx context.Context, req CreateRequest) (*Workspace, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Count active workspaces.
	active := 0
	for _, ws := range m.workspaces {
		if ws.Status != StatusArchived {
			active++
		}
	}
	if active >= m.maxWorkspaces {
		return nil, fmt.Errorf("max workspaces reached (%d)", m.maxWorkspaces)
	}

	now := time.Now()
	ws := &Workspace{
		ID:     uuid.New().String(),
		Name:   req.Name,
		Goal:   req.Goal,
		Status: StatusForming,
		Members: []*Member{
			{
				DID:      m.localDID,
				Role:     RoleCreator,
				JoinedAt: now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  req.Metadata,
	}

	if err := m.persist(ws); err != nil {
		return nil, fmt.Errorf("persist workspace: %w", err)
	}

	m.workspaces[ws.ID] = ws
	m.logger.Infow("workspace created", "id", ws.ID, "name", ws.Name)
	return ws, nil
}

// Join adds the local agent to a workspace.
func (m *Manager) Join(ctx context.Context, workspaceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ws, ok := m.workspaces[workspaceID]
	if !ok {
		return fmt.Errorf("workspace %s: %w", workspaceID, ErrWorkspaceNotFound)
	}
	if ws.Status == StatusArchived {
		return fmt.Errorf("workspace %s is archived", workspaceID)
	}

	// Check if already a member.
	for _, mem := range ws.Members {
		if mem.DID == m.localDID {
			return nil // already a member
		}
	}

	ws.Members = append(ws.Members, &Member{
		DID:      m.localDID,
		Role:     RoleMember,
		JoinedAt: time.Now(),
	})
	ws.UpdatedAt = time.Now()

	if err := m.persist(ws); err != nil {
		return fmt.Errorf("persist workspace: %w", err)
	}

	m.logger.Infow("joined workspace", "id", workspaceID)
	return nil
}

// Leave removes the local agent from a workspace.
func (m *Manager) Leave(ctx context.Context, workspaceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ws, ok := m.workspaces[workspaceID]
	if !ok {
		return fmt.Errorf("workspace %s: %w", workspaceID, ErrWorkspaceNotFound)
	}

	members := make([]*Member, 0, len(ws.Members))
	for _, mem := range ws.Members {
		if mem.DID != m.localDID {
			members = append(members, mem)
		}
	}
	ws.Members = members
	ws.UpdatedAt = time.Now()

	if err := m.persist(ws); err != nil {
		return fmt.Errorf("persist workspace: %w", err)
	}

	m.logger.Infow("left workspace", "id", workspaceID)
	return nil
}

// List returns all known workspaces.
func (m *Manager) List(ctx context.Context) ([]*Workspace, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Workspace, 0, len(m.workspaces))
	for _, ws := range m.workspaces {
		result = append(result, ws)
	}
	return result, nil
}

// Get returns a workspace by ID.
func (m *Manager) Get(ctx context.Context, workspaceID string) (*Workspace, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ws, ok := m.workspaces[workspaceID]
	if !ok {
		return nil, fmt.Errorf("workspace %s: %w", workspaceID, ErrWorkspaceNotFound)
	}
	return ws, nil
}

// Activate transitions a workspace from forming to active.
func (m *Manager) Activate(ctx context.Context, workspaceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ws, ok := m.workspaces[workspaceID]
	if !ok {
		return fmt.Errorf("workspace %s: %w", workspaceID, ErrWorkspaceNotFound)
	}
	if ws.Status != StatusForming {
		return fmt.Errorf("workspace %s is not in forming state", workspaceID)
	}

	ws.Status = StatusActive
	ws.UpdatedAt = time.Now()

	if err := m.persist(ws); err != nil {
		return fmt.Errorf("persist workspace: %w", err)
	}

	m.logger.Infow("workspace activated", "id", workspaceID)
	return nil
}

// Archive transitions a workspace to archived status.
func (m *Manager) Archive(ctx context.Context, workspaceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ws, ok := m.workspaces[workspaceID]
	if !ok {
		return fmt.Errorf("workspace %s: %w", workspaceID, ErrWorkspaceNotFound)
	}

	ws.Status = StatusArchived
	ws.UpdatedAt = time.Now()

	if err := m.persist(ws); err != nil {
		return fmt.Errorf("persist workspace: %w", err)
	}

	m.logger.Infow("workspace archived", "id", workspaceID)
	return nil
}

// Post adds a message to a workspace.
func (m *Manager) Post(ctx context.Context, workspaceID string, msg Message) error {
	m.mu.RLock()
	ws, ok := m.workspaces[workspaceID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("workspace %s: %w", workspaceID, ErrWorkspaceNotFound)
	}
	if ws.Status == StatusArchived {
		return fmt.Errorf("workspace %s is archived", workspaceID)
	}

	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	msg.WorkspaceID = workspaceID
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	return m.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketMessages)
		key := []byte(workspaceID + "/" + msg.ID)
		return b.Put(key, data)
	})
}

// Read returns messages from a workspace.
func (m *Manager) Read(ctx context.Context, workspaceID string, opts ReadOptions) ([]Message, error) {
	m.mu.RLock()
	_, ok := m.workspaces[workspaceID]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("workspace %s: %w", workspaceID, ErrWorkspaceNotFound)
	}

	if opts.Limit <= 0 {
		opts.Limit = 50
	}

	var messages []Message
	prefix := []byte(workspaceID + "/")

	err := m.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketMessages)
		c := b.Cursor()

		for k, v := c.Seek(prefix); k != nil && len(messages) < opts.Limit; k, v = c.Next() {
			if !bytes.HasPrefix(k, prefix) {
				break
			}

			var msg Message
			if err := json.Unmarshal(v, &msg); err != nil {
				continue
			}

			// Apply filters.
			if !opts.Before.IsZero() && !msg.Timestamp.Before(opts.Before) {
				continue
			}
			if !opts.After.IsZero() && !msg.Timestamp.After(opts.After) {
				continue
			}
			if opts.SenderDID != "" && msg.SenderDID != opts.SenderDID {
				continue
			}
			if opts.ParentID != "" && msg.ParentID != opts.ParentID {
				continue
			}
			if len(opts.Types) > 0 {
				typeMatch := false
				for _, t := range opts.Types {
					if string(msg.Type) == t {
						typeMatch = true
						break
					}
				}
				if !typeMatch {
					continue
				}
			}

			messages = append(messages, msg)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read messages: %w", err)
	}

	return messages, nil
}

// AddMember adds a remote member to a workspace.
func (m *Manager) AddMember(ctx context.Context, workspaceID string, member *Member) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ws, ok := m.workspaces[workspaceID]
	if !ok {
		return fmt.Errorf("workspace %s: %w", workspaceID, ErrWorkspaceNotFound)
	}

	for _, mem := range ws.Members {
		if mem.DID == member.DID {
			return nil // already a member
		}
	}

	ws.Members = append(ws.Members, member)
	ws.UpdatedAt = time.Now()

	return m.persist(ws)
}

func (m *Manager) persist(ws *Workspace) error {
	data, err := json.Marshal(ws)
	if err != nil {
		return fmt.Errorf("marshal workspace: %w", err)
	}
	return m.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketWorkspaces).Put([]byte(ws.ID), data)
	})
}

func (m *Manager) loadAll() error {
	return m.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketWorkspaces)
		return b.ForEach(func(k, v []byte) error {
			var ws Workspace
			if err := json.Unmarshal(v, &ws); err != nil {
				m.logger.Warnw("skip corrupt workspace", "key", string(k), "error", err)
				return nil
			}
			m.workspaces[ws.ID] = &ws
			return nil
		})
	})
}
