package provenance

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/ent"
	entsp "github.com/langoai/lango/internal/ent/sessionprovenance"
)

var _ SessionTreeStore = (*EntSessionTreeStore)(nil)

// EntSessionTreeStore persists session provenance nodes in Ent.
type EntSessionTreeStore struct {
	client *ent.Client
}

// NewEntSessionTreeStore creates a new Ent-backed session tree store.
func NewEntSessionTreeStore(client *ent.Client) *EntSessionTreeStore {
	return &EntSessionTreeStore{client: client}
}

func (s *EntSessionTreeStore) SaveNode(ctx context.Context, node SessionNode) error {
	existing, err := s.client.SessionProvenance.Query().
		Where(entsp.SessionKeyEQ(node.SessionKey)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return fmt.Errorf("query session node: %w", err)
	}

	if ent.IsNotFound(err) {
		builder := s.client.SessionProvenance.Create().
			SetID(uuid.New()).
			SetSessionKey(node.SessionKey).
			SetAgentName(node.AgentName).
			SetDepth(node.Depth).
			SetStatus(entsp.Status(node.Status)).
			SetCreatedAt(node.CreatedAt)
		if node.ParentKey != "" {
			builder = builder.SetParentKey(node.ParentKey)
		}
		if node.Goal != "" {
			builder = builder.SetGoal(node.Goal)
		}
		if node.RunID != "" {
			builder = builder.SetRunID(node.RunID)
		}
		if node.WorkspaceID != "" {
			builder = builder.SetWorkspaceID(node.WorkspaceID)
		}
		if node.ClosedAt != nil {
			builder = builder.SetClosedAt(*node.ClosedAt)
		}
		if _, err := builder.Save(ctx); err != nil {
			return fmt.Errorf("save session node: %w", err)
		}
		return nil
	}

	upd := existing.Update().
		SetAgentName(node.AgentName).
		SetDepth(node.Depth).
		SetStatus(entsp.Status(node.Status))
	if node.ParentKey != "" {
		upd = upd.SetParentKey(node.ParentKey)
	} else {
		upd = upd.ClearParentKey()
	}
	if node.Goal != "" {
		upd = upd.SetGoal(node.Goal)
	} else {
		upd = upd.ClearGoal()
	}
	if node.RunID != "" {
		upd = upd.SetRunID(node.RunID)
	} else {
		upd = upd.ClearRunID()
	}
	if node.WorkspaceID != "" {
		upd = upd.SetWorkspaceID(node.WorkspaceID)
	} else {
		upd = upd.ClearWorkspaceID()
	}
	if node.ClosedAt != nil {
		upd = upd.SetClosedAt(*node.ClosedAt)
	} else {
		upd = upd.ClearClosedAt()
	}
	if _, err := upd.Save(ctx); err != nil {
		return fmt.Errorf("update session node: %w", err)
	}
	return nil
}

func (s *EntSessionTreeStore) GetNode(ctx context.Context, sessionKey string) (*SessionNode, error) {
	row, err := s.client.SessionProvenance.Query().
		Where(entsp.SessionKeyEQ(sessionKey)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get session node: %w", err)
	}
	return entRowToSessionNode(row), nil
}

func (s *EntSessionTreeStore) GetChildren(ctx context.Context, parentKey string) ([]SessionNode, error) {
	rows, err := s.client.SessionProvenance.Query().
		Where(entsp.ParentKeyEQ(parentKey)).
		Order(entsp.ByCreatedAt()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("get session children: %w", err)
	}
	return entRowsToSessionNodes(rows), nil
}

func (s *EntSessionTreeStore) ListAll(ctx context.Context, limit int) ([]SessionNode, error) {
	query := s.client.SessionProvenance.Query().
		Order(entsp.ByCreatedAt(sql.OrderDesc()))
	if limit > 0 {
		query = query.Limit(limit)
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list session nodes: %w", err)
	}
	return entRowsToSessionNodes(rows), nil
}

func (s *EntSessionTreeStore) UpdateStatus(ctx context.Context, sessionKey string, status SessionStatus, closedAt *time.Time) error {
	row, err := s.client.SessionProvenance.Query().
		Where(entsp.SessionKeyEQ(sessionKey)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return ErrSessionNotFound
	}
	if err != nil {
		return fmt.Errorf("query session node: %w", err)
	}

	upd := row.Update().
		SetStatus(entsp.Status(status))
	if closedAt != nil {
		upd = upd.SetClosedAt(*closedAt)
	} else {
		upd = upd.ClearClosedAt()
	}
	if _, err := upd.Save(ctx); err != nil {
		return fmt.Errorf("update session status: %w", err)
	}
	return nil
}

func entRowToSessionNode(row *ent.SessionProvenance) *SessionNode {
	node := &SessionNode{
		SessionKey:  row.SessionKey,
		ParentKey:   row.ParentKey,
		AgentName:   row.AgentName,
		Goal:        row.Goal,
		RunID:       row.RunID,
		WorkspaceID: row.WorkspaceID,
		Depth:       row.Depth,
		Status:      SessionStatus(row.Status),
		CreatedAt:   row.CreatedAt,
	}
	if row.ClosedAt != nil {
		closed := *row.ClosedAt
		node.ClosedAt = &closed
	}
	return node
}

func entRowsToSessionNodes(rows []*ent.SessionProvenance) []SessionNode {
	out := make([]SessionNode, 0, len(rows))
	for _, row := range rows {
		out = append(out, *entRowToSessionNode(row))
	}
	return out
}
