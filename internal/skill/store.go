package skill

import "context"

// SkillStore defines the persistence interface for skills.
type SkillStore interface {
	Save(ctx context.Context, entry SkillEntry) error
	Get(ctx context.Context, name string) (*SkillEntry, error)
	ListActive(ctx context.Context) ([]SkillEntry, error)
	Activate(ctx context.Context, name string) error
	Delete(ctx context.Context, name string) error
}
