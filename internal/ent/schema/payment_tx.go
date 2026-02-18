package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// PaymentTx holds the schema definition for the PaymentTx entity.
// PaymentTx records blockchain payment transactions for audit and tracking.
type PaymentTx struct {
	ent.Schema
}

// Fields of the PaymentTx.
func (PaymentTx) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("tx_hash").
			Optional().
			Comment("On-chain transaction hash"),
		field.String("from_address").
			NotEmpty().
			Comment("Sender wallet address"),
		field.String("to_address").
			NotEmpty().
			Comment("Recipient wallet address"),
		field.String("amount").
			NotEmpty().
			Comment("Transaction amount in USDC (decimal string)"),
		field.Int64("chain_id").
			Comment("EVM chain ID"),
		field.Enum("status").
			Values("pending", "submitted", "confirmed", "failed").
			Default("pending").
			Comment("Transaction lifecycle status"),
		field.String("session_key").
			Optional().
			Comment("Agent session key that initiated this transaction"),
		field.String("purpose").
			Optional().
			Comment("Human-readable purpose of the payment"),
		field.String("x402_url").
			Optional().
			Comment("URL that triggered X402 payment (if applicable)"),
		field.Enum("payment_method").
			Values("direct_transfer", "x402_v2").
			Default("direct_transfer").
			Comment("How the payment was made: direct ERC-20 transfer or X402 V2 auto-payment"),
		field.String("error_message").
			Optional().
			Comment("Error details if transaction failed"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the PaymentTx.
func (PaymentTx) Edges() []ent.Edge {
	return nil
}

// Indexes of the PaymentTx.
func (PaymentTx) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tx_hash"),
		index.Fields("from_address"),
		index.Fields("status"),
		index.Fields("created_at"),
		index.Fields("session_key"),
	}
}
