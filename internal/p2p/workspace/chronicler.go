package workspace

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Triple represents a subject-predicate-object triple for the graph store.
type Triple struct {
	Subject   string
	Predicate string
	Object    string
}

// TripleAdder persists triples to a graph store.
type TripleAdder func(ctx context.Context, triples []Triple) error

// Chronicler persists workspace messages as triples in the graph store.
type Chronicler struct {
	addTriples TripleAdder
	logger     *zap.SugaredLogger
}

// NewChronicler creates a new message chronicler.
func NewChronicler(adder TripleAdder, logger *zap.SugaredLogger) *Chronicler {
	return &Chronicler{
		addTriples: adder,
		logger:     logger,
	}
}

// Record persists a workspace message as graph triples.
func (c *Chronicler) Record(ctx context.Context, msg Message) error {
	if c.addTriples == nil {
		return nil
	}

	subject := fmt.Sprintf("workspace:message:%s", msg.ID)
	triples := []Triple{
		{Subject: subject, Predicate: "type", Object: string(msg.Type)},
		{Subject: subject, Predicate: "workspace", Object: msg.WorkspaceID},
		{Subject: subject, Predicate: "sender", Object: msg.SenderDID},
		{Subject: subject, Predicate: "content", Object: msg.Content},
		{Subject: subject, Predicate: "timestamp", Object: msg.Timestamp.Format(time.RFC3339)},
	}

	if msg.ParentID != "" {
		triples = append(triples, Triple{
			Subject: subject, Predicate: "replyTo", Object: fmt.Sprintf("workspace:message:%s", msg.ParentID),
		})
	}

	for k, v := range msg.Metadata {
		triples = append(triples, Triple{
			Subject: subject, Predicate: "meta:" + k, Object: v,
		})
	}

	if err := c.addTriples(ctx, triples); err != nil {
		return fmt.Errorf("persist message triples: %w", err)
	}

	c.logger.Debugw("chronicled workspace message",
		"messageID", msg.ID,
		"type", msg.Type,
		"workspace", msg.WorkspaceID,
		"triples", len(triples),
	)
	return nil
}

// HandleMessage is a MessageHandler that records messages via the chronicler.
func (c *Chronicler) HandleMessage(msg Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := c.Record(ctx, msg); err != nil {
		c.logger.Warnw("chronicler record error", "messageID", msg.ID, "error", err)
	}
}
