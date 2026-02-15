package slack

import (
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// Client defines the interface for Slack REST API operations.
type Client interface {
	AuthTest() (*slack.AuthTestResponse, error)
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
	UpdateMessage(channelID, timestamp string, options ...slack.MsgOption) (string, string, string, error)
}

// Socket defines the interface for Slack Socket Mode operations.
type Socket interface {
	Run() error
	Ack(req socketmode.Request, payload ...interface{})
	Events() <-chan socketmode.Event
}

// SlackClient is an adapter for *slack.Client.
type SlackClient struct {
	*slack.Client
}

var _ Client = (*SlackClient)(nil)

// NewSlackClient creates a new SlackClient adapter.
func NewSlackClient(c *slack.Client) *SlackClient {
	return &SlackClient{Client: c}
}

// SlackSocket is an adapter for *socketmode.Client.
type SlackSocket struct {
	*socketmode.Client
}

var _ Socket = (*SlackSocket)(nil)

// NewSlackSocket creates a new SlackSocket adapter.
func NewSlackSocket(c *socketmode.Client) *SlackSocket {
	return &SlackSocket{Client: c}
}

// Events returns the events channel.
func (s *SlackSocket) Events() <-chan socketmode.Event {
	return s.Client.Events
}
