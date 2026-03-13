package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/channels/discord"
	"github.com/langoai/lango/internal/channels/slack"
	"github.com/langoai/lango/internal/channels/telegram"
	"github.com/langoai/lango/internal/session"
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
	response, err := a.runAgent(ctx, sessionKey, msg.Text)
	if err != nil {
		return nil, err
	}
	return &telegram.OutgoingMessage{Text: response}, nil
}

func (a *App) handleDiscordMessage(ctx context.Context, msg *discord.IncomingMessage) (*discord.OutgoingMessage, error) {
	sessionKey := fmt.Sprintf("%s:%s:%s", types.ChannelDiscord, msg.ChannelID, msg.AuthorID)
	response, err := a.runAgent(ctx, sessionKey, msg.Content)
	if err != nil {
		return nil, err
	}
	return &discord.OutgoingMessage{Content: response}, nil
}

func (a *App) handleSlackMessage(ctx context.Context, msg *slack.IncomingMessage) (*slack.OutgoingMessage, error) {
	sessionKey := fmt.Sprintf("%s:%s:%s", types.ChannelSlack, msg.ChannelID, msg.UserID)
	response, err := a.runAgent(ctx, sessionKey, msg.Text)
	if err != nil {
		return nil, err
	}
	return &slack.OutgoingMessage{Text: response}, nil
}

// emptyResponseFallback is returned to the user when the agent succeeds
// but produces no visible text (e.g. Gemini thought-only responses).
const emptyResponseFallback = "I processed your message but couldn't formulate a visible response. Could you try rephrasing your question?"

// runAgent executes the agent and aggregates the response.
// It injects the session key into the context so that downstream components
// (approval providers, learning engine, etc.) can route by channel.
// After each agent turn, buffers (memory, analysis) are triggered for async processing.
func (a *App) runAgent(ctx context.Context, sessionKey, input string) (string, error) {
	timeout := a.Config.Agent.RequestTimeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}

	start := time.Now()
	logger().Debugw("agent request started",
		"session", sessionKey,
		"timeout", timeout.String(),
		"input_len", len(input))

	var cancel context.CancelFunc
	var extDeadline *ExtendableDeadline
	var runOpts []adk.RunOption

	if a.Config.Agent.AutoExtendTimeout {
		maxTimeout := a.Config.Agent.MaxRequestTimeout
		if maxTimeout <= 0 {
			maxTimeout = timeout * 3
		}
		ctx, extDeadline = NewExtendableDeadline(ctx, timeout, maxTimeout)
		cancel = extDeadline.Stop
		runOpts = append(runOpts, adk.WithOnActivity(func() {
			extDeadline.Extend()
		}))
		logger().Debugw("auto-extend timeout enabled",
			"session", sessionKey,
			"baseTimeout", timeout.String(),
			"maxTimeout", maxTimeout.String())
	} else {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	// Warn when approaching timeout (80%).
	warnTimer := time.AfterFunc(time.Duration(float64(timeout)*0.8), func() {
		logger().Warnw("agent request approaching timeout",
			"session", sessionKey,
			"elapsed", time.Since(start).String(),
			"timeout", timeout.String())
	})
	defer warnTimer.Stop()

	ctx = session.WithSessionKey(ctx, sessionKey)
	response, err := a.Agent.RunAndCollect(ctx, sessionKey, input, runOpts...)

	// Trigger async buffers after agent turn regardless of error.
	if a.MemoryBuffer != nil {
		a.MemoryBuffer.Trigger(sessionKey)
	}
	if a.AnalysisBuffer != nil {
		a.AnalysisBuffer.Trigger(sessionKey)
	}

	elapsed := time.Since(start)
	if err != nil {
		// Check if the error carries a partial result we can recover.
		var agentErr *adk.AgentError
		if errors.As(err, &agentErr) && agentErr.Partial != "" {
			logger().Warnw("agent request failed with partial result",
				"session", sessionKey,
				"elapsed", elapsed.String(),
				"code", string(agentErr.Code),
				"partial_len", len(agentErr.Partial))
			return formatPartialResponse(agentErr.Partial, agentErr), nil
		}

		if ctx.Err() == context.DeadlineExceeded {
			logger().Errorw("agent request timed out",
				"session", sessionKey,
				"elapsed", elapsed.String(),
				"timeout", timeout.String())
			return "", fmt.Errorf("request timed out after %v", timeout)
		}

		logger().Warnw("agent request failed",
			"session", sessionKey,
			"elapsed", elapsed.String(),
			"error", err)
		return "", err
	}

	if response == "" {
		logger().Warnw("empty agent response, using fallback",
			"session", sessionKey,
			"elapsed", elapsed.String())
		response = emptyResponseFallback
	}

	// Apply response sanitization.
	if a.Sanitizer != nil && a.Sanitizer.Enabled() {
		response = a.Sanitizer.Sanitize(response)
	}

	logger().Infow("agent request completed",
		"session", sessionKey,
		"elapsed", elapsed.String(),
		"response_len", len(response))
	return response, nil
}
