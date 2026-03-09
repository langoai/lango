package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// EscrowDeal holds the schema definition for the EscrowDeal entity.
// EscrowDeal records escrow agreements between peers with on-chain tracking.
type EscrowDeal struct {
	ent.Schema
}

// Fields of the EscrowDeal.
func (EscrowDeal) Fields() []ent.Field {
	return []ent.Field{
		field.String("escrow_id").
			Unique().
			NotEmpty().
			Comment("Unique escrow identifier"),
		field.String("buyer_did").
			NotEmpty().
			Comment("Buyer DID"),
		field.String("seller_did").
			NotEmpty().
			Comment("Seller DID"),
		field.String("total_amount").
			NotEmpty().
			Comment("Total escrow amount as decimal string (big.Int)"),
		field.String("status").
			NotEmpty().
			Default("pending").
			Comment("Escrow lifecycle status"),
		field.Bytes("milestones").
			Optional().
			Comment("JSON-serialized milestone data"),
		field.String("task_id").
			Optional().
			Comment("Associated task identifier"),
		field.String("reason").
			Optional().
			Comment("Reason for the escrow"),
		field.String("dispute_note").
			Optional().
			Comment("Dispute description if disputed"),
		field.Int64("chain_id").
			Optional().
			Default(0).
			Comment("EVM chain ID for on-chain tracking"),
		field.String("hub_address").
			Optional().
			Comment("On-chain escrow hub contract address"),
		field.String("on_chain_deal_id").
			Optional().
			Comment("Deal ID on the escrow contract"),
		field.String("deposit_tx_hash").
			Optional().
			Comment("Deposit transaction hash"),
		field.String("release_tx_hash").
			Optional().
			Comment("Release transaction hash"),
		field.String("refund_tx_hash").
			Optional().
			Comment("Refund transaction hash"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("expires_at").
			Comment("Escrow expiration time"),
	}
}

// Edges of the EscrowDeal.
func (EscrowDeal) Edges() []ent.Edge {
	return nil
}

// Indexes of the EscrowDeal.
func (EscrowDeal) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("buyer_did"),
		index.Fields("seller_did"),
		index.Fields("status"),
		index.Fields("on_chain_deal_id"),
	}
}
