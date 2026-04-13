package app

import (
	"context"
	"fmt"
	"time"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/channels/discord"
	"github.com/langoai/lango/internal/channels/slack"
	"github.com/langoai/lango/internal/channels/telegram"
	"github.com/langoai/lango/internal/deadline"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/turnrunner"
	"github.com/langoai/lango/internal/turntrace"
	"github.com/langoai/lango/internal/types"
)

// initChannels initializes all configured channels and wires them to the agent
func (a *App) initChannels() error {
	// Telegram
	if a.Config.Channels.Telegram.Enabled {
		tgConfig := telegram.Config{
			BotToken:           a.Config.Channels.Telegram.BotToken,
			Allowlist:          a.Config.Channels.Telegram.Allowlist,
			ApprovalTimeoutSec: a.Config.Security.Interceptor.ApprovalTimeoutSec,
		}
		tgChannel, err := telegram.New(tgConfig)
		if err != nil {
			logger().Errorw("create telegram channel", "error", err)
		} else {
			tgChannel.SetHandler(func(ctx context.Context, msg *telegram.IncomingMessage) (*telegram.OutgoingMessage, error) {
				return a.handleTelegramMessage(ctx, msg)
			})
			a.Channels = append(a.Channels, tgChannel)
			if composite, ok := a.ApprovalProvider.(*approval.CompositeProvider); ok {
				composite.Register(tgChannel.GetApprovalProvider())
			}
			logger().Info("telegram channel initialized")
		}
	}

	// Discord
	if a.Config.Channels.Discord.Enabled {
		dcConfig := discord.Config{
			BotToken:           a.Config.Channels.Discord.BotToken,
			ApplicationID:      a.Config.Channels.Discord.ApplicationID,
			AllowedGuilds:      a.Config.Channels.Discord.AllowedGuilds,
			ApprovalTimeoutSec: a.Config.Security.Interceptor.ApprovalTimeoutSec,
		}
		dcChannel, err := discord.New(dcConfig)
		if err != nil {
			logger().Errorw("create discord channel", "error", err)
		} else {
			dcChannel.SetHandler(func(ctx context.Context, msg *discord.IncomingMessage) (*discord.OutgoingMessage, error) {
				return a.handleDiscordMessage(ctx, msg)
			})
			a.Channels = append(a.Channels, dcChannel)
			if composite, ok := a.ApprovalProvider.(*approval.CompositeProvider); ok {
				composite.Register(dcChannel.GetApprovalProvider())
			}
			logger().Info("discord channel initialized")
		}
	}

	// Slack
	if a.Config.Channels.Slack.Enabled {
		slConfig := slack.Config{
			BotToken:           a.Config.Channels.Slack.BotToken,
			AppToken:           a.Config.Channels.Slack.AppToken,
			SigningSecret:      a.Config.Channels.Slack.SigningSecret,
			ApprovalTimeoutSec: a.Config.Security.Interceptor.ApprovalTimeoutSec,
		}
		slChannel, err := slack.New(slConfig)
		if err != nil {
			logger().Errorw("create slack channel", "error", err)
		} else {
			slChannel.SetHandler(func(ctx context.Context, msg *slack.IncomingMessage) (*slack.OutgoingMessage, error) {
				return a.handleSlackMessage(ctx, msg)
			})
			a.Channels = append(a.Channels, slChannel)
			if composite, ok := a.ApprovalProvider.(*approval.CompositeProvider); ok {
				composite.Register(slChannel.GetApprovalProvider())
			}
			logger().Info("slack channel initialized")
		}
	}

	return nil
}

func (a *App) handleTelegramMessage(ctx context.Context, msg *telegram.IncomingMessage) (*telegram.OutgoingMessage, error) {
	sessionKey := fmt.Sprintf("%s:%d:%d", types.ChannelTelegram, msg.ChatID, msg.UserID)

	if a.EventBus != nil {
		a.EventBus.Publish(eventbus.ChannelMessageReceivedEvent{
			Channel:    string(types.ChannelTelegram),
			SessionKey: sessionKey,
			SenderName: msg.Username,
			SenderID:   fmt.Sprint(msg.UserID),
			Text:       msg.Text,
			Timestamp:  time.Now(),
			Metadata:   map[string]string{"chatID": fmt.Sprint(msg.ChatID)},
		})
	}

	response, err := a.runAgent(ctx, sessionKey, msg.Text)
	if err != nil {
		return nil, err
	}

	out := &telegram.OutgoingMessage{Text: response}
	// Publish after constructing the outgoing message but before return.
	// Note: actual platform delivery happens in the adapter after this
	// handler returns, so this event reflects "response ready" rather
	// than "delivery confirmed". Adapter-level send failures are not
	// captured here.
	if a.EventBus != nil {
		a.EventBus.Publish(eventbus.ChannelMessageSentEvent{
			Channel:      string(types.ChannelTelegram),
			SessionKey:   sessionKey,
			ResponseText: response,
			Timestamp:    time.Now(),
		})
	}

	return out, nil
}

func (a *App) handleDiscordMessage(ctx context.Context, msg *discord.IncomingMessage) (*discord.OutgoingMessage, error) {
	sessionKey := fmt.Sprintf("%s:%s:%s", types.ChannelDiscord, msg.ChannelID, msg.AuthorID)

	if a.EventBus != nil {
		meta := map[string]string{"channelID": msg.ChannelID}
		if msg.GuildID != "" {
			meta["guildID"] = msg.GuildID
		}
		a.EventBus.Publish(eventbus.ChannelMessageReceivedEvent{
			Channel:    string(types.ChannelDiscord),
			SessionKey: sessionKey,
			SenderName: msg.AuthorName,
			SenderID:   msg.AuthorID,
			Text:       msg.Content,
			Timestamp:  time.Now(),
			Metadata:   meta,
		})
	}

	response, err := a.runAgent(ctx, sessionKey, msg.Content)
	if err != nil {
		return nil, err
	}

	out := &discord.OutgoingMessage{Content: response}
	if a.EventBus != nil {
		a.EventBus.Publish(eventbus.ChannelMessageSentEvent{
			Channel:      string(types.ChannelDiscord),
			SessionKey:   sessionKey,
			ResponseText: response,
			Timestamp:    time.Now(),
		})
	}

	return out, nil
}

func (a *App) handleSlackMessage(ctx context.Context, msg *slack.IncomingMessage) (*slack.OutgoingMessage, error) {
	sessionKey := fmt.Sprintf("%s:%s:%s", types.ChannelSlack, msg.ChannelID, msg.UserID)

	if a.EventBus != nil {
		meta := map[string]string{"channelID": msg.ChannelID}
		if msg.ThreadTS != "" {
			meta["threadTS"] = msg.ThreadTS
		}
		a.EventBus.Publish(eventbus.ChannelMessageReceivedEvent{
			Channel:    string(types.ChannelSlack),
			SessionKey: sessionKey,
			SenderName: msg.UserID,
			SenderID:   msg.UserID,
			Text:       msg.Text,
			Timestamp:  time.Now(),
			Metadata:   meta,
		})
	}

	response, err := a.runAgent(ctx, sessionKey, msg.Text)
	if err != nil {
		return nil, err
	}

	if a.EventBus != nil {
		a.EventBus.Publish(eventbus.ChannelMessageSentEvent{
			Channel:      string(types.ChannelSlack),
			SessionKey:   sessionKey,
			ResponseText: response,
			Timestamp:    time.Now(),
		})
	}

	return &slack.OutgoingMessage{Text: response}, nil
}

// runAgent executes the agent and aggregates the response.
func (a *App) runAgent(ctx context.Context, sessionKey, input string) (string, error) {
	idleTimeout, hardCeiling := a.resolveTimeouts()

	start := time.Now()
	logger().Debugw("agent request started",
		"session", sessionKey,
		"idleTimeout", idleTimeout.String(),
		"hardCeiling", hardCeiling.String(),
		"input_len", len(input))

	if a.TurnRunner == nil {
		return "", fmt.Errorf("turn runner is not initialized")
	}
	result, err := a.TurnRunner.Run(ctx, turnrunner.Request{
		SessionKey: sessionKey,
		Input:      input,
		Entrypoint: "channel",
	})
	elapsed := time.Since(start)
	if err != nil {
		logger().Warnw("agent request failed",
			"session", sessionKey,
			"elapsed", elapsed.String(),
			"error", err)
		return "", err
	}

	if result.Outcome != turntrace.OutcomeSuccess {
		logger().Warnw("agent request completed with failure",
			"session", sessionKey,
			"elapsed", elapsed.String(),
			"response_len", len(result.ResponseText),
			"outcome", string(result.Outcome),
			"trace_id", result.TraceID,
			"error_code", result.ErrorCode,
			"cause_class", result.CauseClass,
			"summary", result.Summary)
		return result.ResponseText, nil
	}

	logger().Infow("agent request completed",
		"session", sessionKey,
		"elapsed", elapsed.String(),
		"response_len", len(result.ResponseText),
		"outcome", string(result.Outcome),
		"trace_id", result.TraceID)
	return result.ResponseText, nil
}

// resolveTimeouts determines the idle timeout and hard ceiling based on config.
func (a *App) resolveTimeouts() (idleTimeout, hardCeiling time.Duration) {
	cfg := a.Config.Agent
	return deadline.ResolveTimeouts(deadline.TimeoutConfig{
		IdleTimeout:       cfg.IdleTimeout,
		RequestTimeout:    cfg.RequestTimeout,
		AutoExtendTimeout: cfg.AutoExtendTimeout,
		MaxRequestTimeout: cfg.MaxRequestTimeout,
	})
}
