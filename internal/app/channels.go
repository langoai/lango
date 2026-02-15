package app

import (
	"context"
	"fmt"

	"github.com/langowarny/lango/internal/approval"
	"github.com/langowarny/lango/internal/channels/discord"
	"github.com/langowarny/lango/internal/channels/slack"
	"github.com/langowarny/lango/internal/channels/telegram"
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
			logger().Errorw("failed to create telegram channel", "error", err)
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
			logger().Errorw("failed to create discord channel", "error", err)
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
			logger().Errorw("failed to create slack channel", "error", err)
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
	sessionKey := fmt.Sprintf("telegram:%d:%d", msg.ChatID, msg.UserID)
	response, err := a.runAgent(ctx, sessionKey, msg.Text)
	if err != nil {
		return nil, err
	}
	return &telegram.OutgoingMessage{Text: response}, nil
}

func (a *App) handleDiscordMessage(ctx context.Context, msg *discord.IncomingMessage) (*discord.OutgoingMessage, error) {
	sessionKey := fmt.Sprintf("discord:%s:%s", msg.ChannelID, msg.AuthorID)
	response, err := a.runAgent(ctx, sessionKey, msg.Content)
	if err != nil {
		return nil, err
	}
	return &discord.OutgoingMessage{Content: response}, nil
}

func (a *App) handleSlackMessage(ctx context.Context, msg *slack.IncomingMessage) (*slack.OutgoingMessage, error) {
	sessionKey := fmt.Sprintf("slack:%s:%s", msg.ChannelID, msg.UserID)
	response, err := a.runAgent(ctx, sessionKey, msg.Text)
	if err != nil {
		return nil, err
	}
	return &slack.OutgoingMessage{Text: response}, nil
}

// runAgent executes the agent and aggregates the response
func (a *App) runAgent(ctx context.Context, sessionKey, input string) (string, error) {
	return a.Agent.RunAndCollect(ctx, sessionKey, input)
}
