package app

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/langowarny/lango/internal/channels/discord"
	"github.com/langowarny/lango/internal/channels/slack"
	"github.com/langowarny/lango/internal/channels/telegram"
	"github.com/langowarny/lango/internal/types"
)

// channelSender implements cron.ChannelSender, background.ChannelNotifier, and
// workflow.ChannelSender by dispatching messages to the configured channel adapters.
type channelSender struct {
	app *App
}

func newChannelSender(a *App) *channelSender {
	return &channelSender{app: a}
}

// parseDeliveryTarget splits "channel:id" into channel type and optional target ID.
func parseDeliveryTarget(target string) (types.ChannelType, string) {
	target = strings.TrimSpace(target)
	if idx := strings.IndexByte(target, ':'); idx >= 0 {
		return types.ChannelType(strings.ToLower(target[:idx])), target[idx+1:]
	}
	return types.ChannelType(strings.ToLower(target)), ""
}

// SendMessage sends a text message to the specified delivery target.
// Target format: "channel" (bare name) or "channel:id" (with routing ID).
// For telegram: uses chatID from target, falls back to first allowlisted chat ID.
// For discord/slack: uses channelID from target, returns error if not provided.
func (s *channelSender) SendMessage(_ context.Context, channel, message string) error {
	chName, targetID := parseDeliveryTarget(channel)

	for _, c := range s.app.Channels {
		switch chName {
		case types.ChannelTelegram:
			if tg, ok := c.(*telegram.Channel); ok {
				var chatID int64
				if targetID != "" {
					parsed, err := strconv.ParseInt(targetID, 10, 64)
					if err != nil {
						return fmt.Errorf("parse telegram chat ID %q: %w", targetID, err)
					}
					chatID = parsed
				} else {
					chatID = s.firstTelegramChatID()
					if chatID == 0 {
						return fmt.Errorf("telegram delivery requires a chat ID (use telegram:CHAT_ID) or at least one allowlisted chat ID")
					}
				}
				return tg.Send(chatID, &telegram.OutgoingMessage{Text: message})
			}
		case types.ChannelDiscord:
			if dc, ok := c.(*discord.Channel); ok {
				if targetID == "" {
					return fmt.Errorf("discord delivery requires a channel ID (use discord:CHANNEL_ID)")
				}
				return dc.Send(targetID, &discord.OutgoingMessage{Content: message})
			}
		case types.ChannelSlack:
			if sl, ok := c.(*slack.Channel); ok {
				if targetID == "" {
					return fmt.Errorf("slack delivery requires a channel ID (use slack:CHANNEL_ID)")
				}
				return sl.Send(targetID, &slack.OutgoingMessage{Text: message})
			}
		}
	}

	return fmt.Errorf("channel %q not available", channel)
}

// StartTyping starts a typing indicator on the specified delivery target.
// The returned stop function ends the typing indicator. It is always non-nil.
// Typing failures are logged but never block execution.
func (s *channelSender) StartTyping(ctx context.Context, channel string) (func(), error) {
	chName, targetID := parseDeliveryTarget(channel)
	noop := func() {}

	for _, c := range s.app.Channels {
		switch chName {
		case types.ChannelTelegram:
			if tg, ok := c.(*telegram.Channel); ok {
				var chatID int64
				if targetID != "" {
					parsed, err := strconv.ParseInt(targetID, 10, 64)
					if err != nil {
						return noop, fmt.Errorf("parse telegram chat ID %q: %w", targetID, err)
					}
					chatID = parsed
				} else {
					chatID = s.firstTelegramChatID()
					if chatID == 0 {
						return noop, nil
					}
				}
				return tg.StartTyping(ctx, chatID), nil
			}
		case types.ChannelDiscord:
			if dc, ok := c.(*discord.Channel); ok {
				if targetID == "" {
					return noop, nil
				}
				return dc.StartTyping(ctx, targetID), nil
			}
		case types.ChannelSlack:
			if sl, ok := c.(*slack.Channel); ok {
				if targetID == "" {
					return noop, nil
				}
				return sl.StartTyping(targetID), nil
			}
		}
	}

	return noop, nil
}

func (s *channelSender) firstTelegramChatID() int64 {
	if len(s.app.Config.Channels.Telegram.Allowlist) > 0 {
		return s.app.Config.Channels.Telegram.Allowlist[0]
	}
	return 0
}
