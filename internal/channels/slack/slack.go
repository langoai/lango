package slack

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/langoai/lango/internal/logging"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

var logger = logging.SubsystemSugar("channel.slack")

// Config holds Slack channel configuration
type Config struct {
	BotToken           string // xoxb-...
	AppToken           string // xapp-... (for Socket Mode)
	SigningSecret      string
	ApprovalTimeoutSec int          // 0 = default 30s
	APIURL             string       // optional, for testing
	HTTPClient         *http.Client // optional, for testing
	Client             Client       // optional, for testing
	Socket             Socket       // optional, for testing
}

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, msg *IncomingMessage) (*OutgoingMessage, error)

// IncomingMessage represents a message from Slack
type IncomingMessage struct {
	EventType string // app_mention, message
	ChannelID string
	UserID    string
	Text      string
	ThreadTS  string
	IsThread  bool
}

// OutgoingMessage represents a message to send
type OutgoingMessage struct {
	Text     string
	ThreadTS string
	Blocks   []Block
}

// Block represents a Slack Block Kit block
type Block struct {
	Type string
	Text *TextBlock
}

// TextBlock represents text in a block
type TextBlock struct {
	Type string // mrkdwn, plain_text
	Text string
}

// Channel implements Slack bot
type Channel struct {
	config   Config
	api      Client
	socket   Socket
	handler  MessageHandler
	approval *ApprovalProvider
	botID    string
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// New creates a new Slack channel
func New(cfg Config) (*Channel, error) {
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("bot token is required")
	}
	if cfg.AppToken == "" {
		return nil, fmt.Errorf("app token is required for Socket Mode")
	}

	opts := []slack.Option{
		slack.OptionAppLevelToken(cfg.AppToken),
	}

	if cfg.APIURL != "" {
		opts = append(opts, slack.OptionAPIURL(cfg.APIURL))
	}
	if cfg.HTTPClient != nil {
		opts = append(opts, slack.OptionHTTPClient(cfg.HTTPClient))
	}

	var apiClient Client
	var socketClient Socket

	if cfg.Client != nil {
		apiClient = cfg.Client
	} else {
		api := slack.New(cfg.BotToken, opts...)
		apiClient = NewSlackClient(api)
	}

	if cfg.Socket != nil {
		socketClient = cfg.Socket
	} else {
		// If we created a real client (wrapped in adapter), we can unwrap it or just recreate the ref locally?
		// But Wait, NewSlackClient wraps *slack.Client.
		// If we are here, apiClient is *SlackClient.

		if adapter, ok := apiClient.(*SlackClient); ok {
			socket := socketmode.New(
				adapter.Client,
				socketmode.OptionDebug(false),
			)
			socketClient = NewSlackSocket(socket)
		} else {
			// This happens if cfg.Client is provided (mock) but cfg.Socket is NOT.
			return nil, fmt.Errorf("must provide Socket if Client is mocked")
		}
	}

	ch := &Channel{
		config:   cfg,
		api:      apiClient,
		socket:   socketClient,
		stopChan: make(chan struct{}),
	}
	ch.approval = NewApprovalProvider(apiClient, time.Duration(cfg.ApprovalTimeoutSec)*time.Second)

	return ch, nil
}

// SetHandler sets the message handler
func (c *Channel) SetHandler(handler MessageHandler) {
	c.handler = handler
}

// GetApprovalProvider returns the channel's approval provider for composite registration.
func (c *Channel) GetApprovalProvider() *ApprovalProvider {
	return c.approval
}

// Start starts the Slack bot
func (c *Channel) Start(ctx context.Context) error {
	if c.handler == nil {
		return fmt.Errorf("message handler not set")
	}

	// Get bot user ID
	authResp, err := c.api.AuthTest()
	if err != nil {
		return fmt.Errorf("auth test failed: %w", err)
	}
	c.botID = authResp.UserID
	logger.Infow("slack bot connected", "botId", c.botID, "team", authResp.Team)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.handleEvents(ctx)
	}()

	go func() {
		if err := c.socket.Run(); err != nil {
			logger.Errorw("socket mode error", "error", err)
		}
	}()

	return nil
}

// handleEvents processes events from Socket Mode
func (c *Channel) handleEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopChan:
			return
		case event := <-c.socket.Events():
			switch event.Type {
			case socketmode.EventTypeEventsAPI:
				c.handleEventsAPI(ctx, event)
			case socketmode.EventTypeInteractive:
				c.handleInteractiveEvent(event)
			case socketmode.EventTypeSlashCommand:
				// Handle slash commands if needed
			}
		}
	}
}

// handleEventsAPI processes Events API events
func (c *Channel) handleEventsAPI(ctx context.Context, event socketmode.Event) {
	eventsAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
	if !ok {
		return
	}

	c.socket.Ack(*event.Request)

	switch eventsAPIEvent.Type {
	case slackevents.CallbackEvent:
		c.handleCallbackEvent(ctx, eventsAPIEvent.InnerEvent)
	}
}

// handleInteractiveEvent processes interactive events (button clicks)
func (c *Channel) handleInteractiveEvent(event socketmode.Event) {
	callback, ok := event.Data.(slack.InteractionCallback)
	if !ok {
		return
	}

	c.socket.Ack(*event.Request)

	if callback.Type == slack.InteractionTypeBlockActions {
		for _, action := range callback.ActionCallback.BlockActions {
			c.approval.HandleInteractive(action.ActionID)
		}
	}
}

// handleCallbackEvent handles inner callback events
func (c *Channel) handleCallbackEvent(ctx context.Context, innerEvent slackevents.EventsAPIInnerEvent) {
	switch ev := innerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		c.handleMessage(ctx, "app_mention", ev.Channel, ev.User, ev.Text, ev.ThreadTimeStamp)
	case *slackevents.MessageEvent:
		// Only handle DMs (channel type = "im")
		if ev.ChannelType == "im" {
			c.handleMessage(ctx, "message", ev.Channel, ev.User, ev.Text, ev.ThreadTimeStamp)
		}
	}
}

// handleMessage processes a message event
func (c *Channel) handleMessage(ctx context.Context, eventType, channelID, userID, text, threadTS string) {
	// Ignore bot's own messages
	if userID == c.botID {
		return
	}

	// Clean text (remove bot mention)
	text = c.cleanText(text)

	incoming := &IncomingMessage{
		EventType: eventType,
		ChannelID: channelID,
		UserID:    userID,
		Text:      text,
		ThreadTS:  threadTS,
		IsThread:  threadTS != "",
	}

	logger.Infow("received message",
		"eventType", eventType,
		"channelId", channelID,
		"userId", userID,
	)

	// Run handler in a separate goroutine to avoid blocking the event loop.
	// Without this, a blocking handler (e.g. waiting for approval) would prevent
	// interactive events (button clicks) from being processed, causing a deadlock.
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		// Post a placeholder "Thinking..." message while processing
		placeholderTS, placeholderErr := c.postThinking(channelID, threadTS)

		response, err := c.handler(ctx, incoming)
		if err != nil {
			logger.Errorw("handler error", "error", err)
			c.sendError(channelID, threadTS, err)
			// Clean up placeholder on error
			if placeholderErr == nil {
				c.updateThinking(channelID, placeholderTS, fmt.Sprintf("Error: %s", err.Error()))
			}
			return
		}

		if response != nil {
			// Replace placeholder with actual response
			if placeholderErr == nil {
				formattedText := FormatMrkdwn(response.Text)
				if updateErr := c.updateThinking(channelID, placeholderTS, formattedText); updateErr != nil {
					logger.Warnw("placeholder update failed, sending new message", "error", updateErr)
					if err := c.Send(channelID, response); err != nil {
						logger.Errorw("send error", "error", err)
					}
				}
			} else {
				// Placeholder failed, send normally
				if err := c.Send(channelID, response); err != nil {
					logger.Errorw("send error", "error", err)
				}
			}
		}
	}()
}

// StartTyping posts a "_Processing..._ " placeholder message.
// The returned stop function deletes the placeholder on call.
// If posting fails, a no-op stop function is returned.
func (c *Channel) StartTyping(channelID string) func() {
	options := []slack.MsgOption{
		slack.MsgOptionText("_Processing..._", false),
	}

	_, ts, err := c.api.PostMessage(channelID, options...)
	if err != nil {
		logger.Warnw("start typing placeholder error", "error", err)
		return func() {}
	}

	var once sync.Once
	return func() {
		once.Do(func() {
			if _, _, err := c.api.DeleteMessage(channelID, ts); err != nil {
				logger.Warnw("delete typing placeholder error", "error", err)
			}
		})
	}
}

// postThinking posts a placeholder "Thinking..." message and returns its timestamp.
func (c *Channel) postThinking(channelID, threadTS string) (string, error) {
	options := []slack.MsgOption{
		slack.MsgOptionText("_Thinking..._", false),
	}
	if threadTS != "" {
		options = append(options, slack.MsgOptionTS(threadTS))
	}

	_, ts, err := c.api.PostMessage(channelID, options...)
	if err != nil {
		logger.Warnw("post thinking placeholder error", "error", err)
		return "", err
	}
	return ts, nil
}

// updateThinking replaces a placeholder message with the given text.
func (c *Channel) updateThinking(channelID, messageTS, text string) error {
	_, _, _, err := c.api.UpdateMessage(channelID, messageTS, slack.MsgOptionText(text, false))
	return err
}

// Send sends a message.
// Standard Markdown in msg.Text is auto-converted to Slack mrkdwn before sending.
func (c *Channel) Send(channelID string, msg *OutgoingMessage) error {
	formattedText := FormatMrkdwn(msg.Text)

	options := []slack.MsgOption{
		slack.MsgOptionText(formattedText, false),
	}

	// Reply in thread if specified
	if msg.ThreadTS != "" {
		options = append(options, slack.MsgOptionTS(msg.ThreadTS))
	}

	// Add blocks if specified
	if len(msg.Blocks) > 0 {
		var blocks []slack.Block
		for _, b := range msg.Blocks {
			if b.Type == "section" && b.Text != nil {
				textBlock := slack.NewTextBlockObject(b.Text.Type, b.Text.Text, false, false)
				blocks = append(blocks, slack.NewSectionBlock(textBlock, nil, nil))
			}
		}
		if len(blocks) > 0 {
			options = append(options, slack.MsgOptionBlocks(blocks...))
		}
	}

	_, _, err := c.api.PostMessage(channelID, options...)
	return err
}

// cleanText removes bot mention from text
func (c *Channel) cleanText(text string) string {
	// Remove <@BOTID> mentions
	text = strings.ReplaceAll(text, "<@"+c.botID+">", "")
	return strings.TrimSpace(text)
}

// sendError sends an error message
func (c *Channel) sendError(channelID, threadTS string, err error) {
	c.Send(channelID, &OutgoingMessage{
		Text:     fmt.Sprintf("❌ Error: %s", err.Error()),
		ThreadTS: threadTS,
	})
}

// Stop stops the Slack bot
func (c *Channel) Stop() {
	close(c.stopChan)
	c.wg.Wait()
	logger.Info("slack channel stopped")
}
