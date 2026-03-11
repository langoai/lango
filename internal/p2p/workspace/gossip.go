package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"
)

const topicPrefix = "/lango/workspace/"

// MessageHandler is called when a workspace message is received via GossipSub.
type MessageHandler func(msg Message)

// WorkspaceGossip manages per-workspace GossipSub topics.
type WorkspaceGossip struct {
	ps      *pubsub.PubSub
	localID peer.ID
	logger  *zap.SugaredLogger
	handler MessageHandler

	mu     sync.RWMutex
	topics map[string]*topicState // workspaceID → topic state
}

type topicState struct {
	topic *pubsub.Topic
	sub   *pubsub.Subscription
	stop  context.CancelFunc
}

// GossipConfig configures the workspace gossip.
type GossipConfig struct {
	PubSub  *pubsub.PubSub
	LocalID peer.ID
	Handler MessageHandler
	Logger  *zap.SugaredLogger
}

// NewWorkspaceGossip creates a new workspace gossip manager.
func NewWorkspaceGossip(cfg GossipConfig) *WorkspaceGossip {
	return &WorkspaceGossip{
		ps:      cfg.PubSub,
		localID: cfg.LocalID,
		handler: cfg.Handler,
		logger:  cfg.Logger,
		topics:  make(map[string]*topicState),
	}
}

// Subscribe joins the GossipSub topic for a workspace and starts listening.
func (g *WorkspaceGossip) Subscribe(workspaceID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, ok := g.topics[workspaceID]; ok {
		return nil // already subscribed
	}

	topicName := topicPrefix + workspaceID
	topic, err := g.ps.Join(topicName)
	if err != nil {
		return fmt.Errorf("join topic %s: %w", topicName, err)
	}

	sub, err := topic.Subscribe()
	if err != nil {
		topic.Close()
		return fmt.Errorf("subscribe to %s: %w", topicName, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	state := &topicState{
		topic: topic,
		sub:   sub,
		stop:  cancel,
	}
	g.topics[workspaceID] = state

	go g.readLoop(ctx, workspaceID, sub)

	g.logger.Infow("subscribed to workspace topic", "workspace", workspaceID, "topic", topicName)
	return nil
}

// Unsubscribe leaves the GossipSub topic for a workspace.
func (g *WorkspaceGossip) Unsubscribe(workspaceID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	state, ok := g.topics[workspaceID]
	if !ok {
		return
	}

	state.stop()
	state.sub.Cancel()
	state.topic.Close()
	delete(g.topics, workspaceID)

	g.logger.Infow("unsubscribed from workspace topic", "workspace", workspaceID)
}

// Publish sends a message to the workspace's GossipSub topic.
func (g *WorkspaceGossip) Publish(ctx context.Context, workspaceID string, msg Message) error {
	g.mu.RLock()
	state, ok := g.topics[workspaceID]
	g.mu.RUnlock()
	if !ok {
		return fmt.Errorf("not subscribed to workspace %s", workspaceID)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	if err := state.topic.Publish(ctx, data); err != nil {
		return fmt.Errorf("publish to workspace %s: %w", workspaceID, err)
	}

	return nil
}

// SubscribedWorkspaces returns the list of workspace IDs currently subscribed.
func (g *WorkspaceGossip) SubscribedWorkspaces() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	ids := make([]string, 0, len(g.topics))
	for id := range g.topics {
		ids = append(ids, id)
	}
	return ids
}

// Stop unsubscribes from all workspace topics.
func (g *WorkspaceGossip) Stop() {
	g.mu.Lock()
	defer g.mu.Unlock()

	for id, state := range g.topics {
		state.stop()
		state.sub.Cancel()
		state.topic.Close()
		delete(g.topics, id)
	}

	g.logger.Info("workspace gossip stopped")
}

func (g *WorkspaceGossip) readLoop(ctx context.Context, workspaceID string, sub *pubsub.Subscription) {
	for {
		raw, err := sub.Next(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			g.logger.Warnw("workspace gossip subscription error", "workspace", workspaceID, "error", err)
			continue
		}

		// Skip own messages.
		if raw.ReceivedFrom == g.localID {
			continue
		}

		var msg Message
		if err := json.Unmarshal(raw.Data, &msg); err != nil {
			g.logger.Debugw("unmarshal workspace message", "workspace", workspaceID, "error", err)
			continue
		}

		if g.handler != nil {
			g.handler(msg)
		}
	}
}
