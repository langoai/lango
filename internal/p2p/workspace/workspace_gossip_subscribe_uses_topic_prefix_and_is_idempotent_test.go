package workspace

import (
	"context"
	"errors"
	"testing"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentPubSub struct {
	joinErr error
	joins   []string
	topics  map[string]*workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentTopic
}

func (p *workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentPubSub) Join(topic string) (workspaceTopic, error) {
	p.joins = append(p.joins, topic)
	if p.joinErr != nil {
		return nil, p.joinErr
	}
	if p.topics == nil {
		p.topics = make(map[string]*workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentTopic)
	}
	t, ok := p.topics[topic]
	if !ok {
		t = &workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentTopic{}
		p.topics[topic] = t
	}
	return t, nil
}

type workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentTopic struct {
	subscribeErr error
	publishErr   error
	sub          *workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentSubscription
	published    [][]byte
	closeCount   int
}

func (t *workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentTopic) Subscribe() (workspaceSubscription, error) {
	if t.subscribeErr != nil {
		return nil, t.subscribeErr
	}
	if t.sub == nil {
		t.sub = &workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentSubscription{}
	}
	return t.sub, nil
}

func (t *workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentTopic) Publish(ctx context.Context, data []byte) error {
	if t.publishErr != nil {
		return t.publishErr
	}
	t.published = append(t.published, append([]byte(nil), data...))
	return nil
}

func (t *workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentTopic) Close() error {
	t.closeCount++
	return nil
}

type workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentNextResult struct {
	msg *pubsub.Message
	err error
}

type workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentSubscription struct {
	results     []workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentNextResult
	cancelCount int
}

func (s *workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentSubscription) Next(ctx context.Context) (*pubsub.Message, error) {
	if len(s.results) == 0 {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	next := s.results[0]
	s.results = s.results[1:]
	return next.msg, next.err
}

func (s *workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentSubscription) Cancel() {
	s.cancelCount++
}

func newWorkspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentGossip(ps workspacePubSub, handler MessageHandler) *WorkspaceGossip {
	return newWorkspaceGossipWithPubSub(ps, GossipConfig{
		LocalID: peer.ID("local-peer"),
		Handler: handler,
		Logger:  zap.NewNop().Sugar(),
	})
}

func TestWorkspaceGossipSubscribeUsesTopicPrefixAndIsIdempotent(t *testing.T) {
	ps := &workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentPubSub{}
	gossip := newWorkspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentGossip(ps, nil)

	err := gossip.Subscribe("workspace-1")
	require.NoError(t, err)
	err = gossip.Subscribe("workspace-1")
	require.NoError(t, err)

	assert.Equal(t, []string{topicPrefix + "workspace-1"}, ps.joins)
	assert.Equal(t, []string{"workspace-1"}, gossip.SubscribedWorkspaces())

	gossip.Unsubscribe("workspace-1")
	assert.Equal(t, 1, ps.topics[topicPrefix+"workspace-1"].sub.cancelCount)
	assert.Equal(t, 1, ps.topics[topicPrefix+"workspace-1"].closeCount)
	assert.Empty(t, gossip.SubscribedWorkspaces())
}

func TestWorkspaceGossipSubscribeReturnsJoinAndSubscribeErrors(t *testing.T) {
	joinErr := errors.New("join denied")
	gossip := newWorkspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentGossip(&workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentPubSub{joinErr: joinErr}, nil)

	err := gossip.Subscribe("workspace-join")
	require.Error(t, err)
	assert.ErrorIs(t, err, joinErr)
	assert.Contains(t, err.Error(), "join topic /lango/workspace/workspace-join")

	subscribeErr := errors.New("subscription refused")
	topic := &workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentTopic{subscribeErr: subscribeErr}
	gossip = newWorkspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentGossip(&workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentPubSub{
		topics: map[string]*workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentTopic{topicPrefix + "workspace-sub": topic},
	}, nil)

	err = gossip.Subscribe("workspace-sub")
	require.Error(t, err)
	assert.ErrorIs(t, err, subscribeErr)
	assert.Contains(t, err.Error(), "subscribe to /lango/workspace/workspace-sub")
	assert.Equal(t, 1, topic.closeCount)
	assert.Empty(t, gossip.SubscribedWorkspaces())
}

func TestWorkspaceGossipPublishBranches(t *testing.T) {
	msg := Message{
		ID:          "msg-1",
		Type:        MessageTypeKnowledgeShare,
		WorkspaceID: "workspace-1",
		SenderDID:   "did:lango:agent",
		Content:     "hello peers",
		Metadata:    map[string]string{"kind": "note"},
		Timestamp:   time.Date(2026, 5, 19, 8, 0, 0, 0, time.UTC),
	}

	gossip := newWorkspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentGossip(&workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentPubSub{}, nil)
	err := gossip.Publish(context.Background(), "missing", msg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not subscribed to workspace missing")

	topic := &workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentTopic{}
	gossip = newWorkspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentGossip(&workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentPubSub{
		topics: map[string]*workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentTopic{topicPrefix + "workspace-1": topic},
	}, nil)
	require.NoError(t, gossip.Subscribe("workspace-1"))
	t.Cleanup(gossip.Stop)

	err = gossip.Publish(context.Background(), "workspace-1", msg)
	require.NoError(t, err)
	require.Len(t, topic.published, 1)
	assert.JSONEq(t, `{
		"id":"msg-1",
		"type":"KNOWLEDGE_SHARE",
		"workspaceId":"workspace-1",
		"senderDid":"did:lango:agent",
		"content":"hello peers",
		"metadata":{"kind":"note"},
		"timestamp":"2026-05-19T08:00:00Z"
	}`, string(topic.published[0]))

	publishErr := errors.New("publish rejected")
	topic.publishErr = publishErr
	err = gossip.Publish(context.Background(), "workspace-1", msg)
	require.Error(t, err)
	assert.ErrorIs(t, err, publishErr)
	assert.Contains(t, err.Error(), "publish to workspace workspace-1")
}

func TestWorkspaceGossipReadLoopSkipsOwnInvalidAndErroredMessages(t *testing.T) {
	received := make(chan Message, 1)
	gossip := newWorkspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentGossip(&workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentPubSub{}, func(msg Message) {
		received <- msg
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := &workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentSubscription{results: []workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentNextResult{
		{msg: &pubsub.Message{
			Message:      &pb.Message{Data: []byte(`{"id":"own"}`)},
			ReceivedFrom: peer.ID("local-peer"),
		}},
		{msg: &pubsub.Message{
			Message:      &pb.Message{Data: []byte(`{bad json`)},
			ReceivedFrom: peer.ID("remote-peer"),
		}},
		{err: errors.New("temporary read failure")},
		{msg: &pubsub.Message{
			Message: &pb.Message{Data: []byte(`{
			"id":"remote",
			"type":"LOG_STREAM",
			"workspaceId":"workspace-1",
			"senderDid":"did:lango:remote",
			"content":"remote update",
			"timestamp":"2026-05-19T08:30:00Z"
		}`)},
			ReceivedFrom: peer.ID("remote-peer"),
		}},
	}}

	done := make(chan struct{})
	go func() {
		defer close(done)
		gossip.readLoop(ctx, "workspace-1", sub)
	}()

	var got Message
	select {
	case got = <-received:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for remote message")
	}

	assert.Equal(t, "remote", got.ID)
	assert.Equal(t, MessageTypeLogStream, got.Type)
	assert.Equal(t, "remote update", got.Content)

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for readLoop exit")
	}
	assert.Empty(t, received)
}

func TestWorkspaceGossipStopClosesAllTopics(t *testing.T) {
	topicOne := &workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentTopic{}
	topicTwo := &workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentTopic{}
	gossip := newWorkspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentGossip(&workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentPubSub{
		topics: map[string]*workspaceGossipSubscribeUsesTopicPrefixAndIsIdempotentTopic{
			topicPrefix + "one": topicOne,
			topicPrefix + "two": topicTwo,
		},
	}, nil)
	require.NoError(t, gossip.Subscribe("one"))
	require.NoError(t, gossip.Subscribe("two"))

	gossip.Stop()

	assert.Empty(t, gossip.SubscribedWorkspaces())
	assert.Equal(t, 1, topicOne.sub.cancelCount)
	assert.Equal(t, 1, topicOne.closeCount)
	assert.Equal(t, 1, topicTwo.sub.cancelCount)
	assert.Equal(t, 1, topicTwo.closeCount)

	gossip.Unsubscribe("missing")
}
