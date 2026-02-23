package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// PeerReputation holds the schema definition for tracking peer trust scores.
type PeerReputation struct {
	ent.Schema
}

// Fields of the PeerReputation.
func (PeerReputation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("peer_did").
			Unique().
			NotEmpty().
			Comment("DID of the peer"),
		field.Int("successful_exchanges").
			Default(0).
			Comment("Count of successful paid exchanges"),
		field.Int("failed_exchanges").
			Default(0).
			Comment("Count of failed exchanges"),
		field.Int("timeout_count").
			Default(0).
			Comment("Count of timed-out exchanges"),
		field.Float("trust_score").
			Default(0.0).
			Comment("Computed trust score"),
		field.Time("first_seen").
			Default(time.Now).
			Immutable().
			Comment("When this peer was first observed"),
		field.Time("last_interaction").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Most recent interaction timestamp"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the PeerReputation.
func (PeerReputation) Edges() []ent.Edge {
	return nil
}

// Indexes of the PeerReputation.
func (PeerReputation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("trust_score"),
		index.Fields("last_interaction"),
	}
}
