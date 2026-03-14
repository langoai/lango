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
	"github.com/langoai/lango/internal/deadline"
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
	idleTimeout, hardCeiling := a.resolveTimeouts()

	start := time.Now()
	logger().Debugw("agent request started",
		"session", sessionKey,
		"idleTimeout", idleTimeout.String(),
		"hardCeiling", hardCeiling.String(),
		"input_len", len(input))

	var cancel context.CancelFunc
	var extDeadline *deadline.ExtendableDeadline
	var runOpts []adk.RunOption

	if idleTimeout > 0 {
		ctx, extDeadline = deadline.New(ctx, idleTimeout, hardCeiling)
		cancel = extDeadline.Stop
		runOpts = append(runOpts, adk.WithOnActivity(extDeadline.Extend))
		logger().Debugw("idle timeout enabled",
			"session", sessionKey,
			"idleTimeout", idleTimeout.String(),
			"hardCeiling", hardCeiling.String())
	} else {
		ctx, cancel = context.WithTimeout(ctx, hardCeiling)
	}
	defer cancel()

	// Warn when approaching timeout (80% of hard ceiling).
	warnTimer := time.AfterFunc(time.Duration(float64(hardCeiling)*0.8), func() {
		logger().Warnw("agent request approaching timeout",
			"session", sessionKey,
			"elapsed", time.Since(start).String(),
			"hardCeiling", hardCeiling.String())
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
			// Annotate session so next turn doesn't see incomplete history.
			if a.Store != nil {
				_ = a.Store.AnnotateTimeout(sessionKey, agentErr.Partial)
			}
			return formatPartialResponse(agentErr.Partial, agentErr), nil
		}

		if ctx.Err() != nil {
			// Determine the specific timeout reason.
			errCode := adk.ErrTimeout
			errMsg := fmt.Sprintf("request timed out after %v", hardCeiling)
			if extDeadline != nil {
				switch extDeadline.Reason() {
				case deadline.ReasonIdle:
					errCode = adk.ErrIdleTimeout
					errMsg = fmt.Sprintf("no activity for %v", idleTimeout)
				case deadline.ReasonMaxTimeout:
					errMsg = fmt.Sprintf("maximum time limit (%v) exceeded", hardCeiling)
				}
			}
			logger().Errorw("agent request timed out",
				"session", sessionKey,
				"elapsed", elapsed.String(),
				"reason", string(errCode))
			// Annotate session to prevent error leak into next turn.
			if a.Store != nil {
				_ = a.Store.AnnotateTimeout(sessionKey, "")
			}
			return "", &adk.AgentError{
				Code:    errCode,
				Message: errMsg,
				Cause:   ctx.Err(),
				Elapsed: elapsed,
			}
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

// resolveTimeouts determines the idle timeout and hard ceiling based on config.
// Delegates to deadline.ResolveTimeouts for the actual logic.
func (a *App) resolveTimeouts() (idleTimeout, hardCeiling time.Duration) {
	cfg := a.Config.Agent
	return deadline.ResolveTimeouts(deadline.TimeoutConfig{
		IdleTimeout:       cfg.IdleTimeout,
		RequestTimeout:    cfg.RequestTimeout,
		AutoExtendTimeout: cfg.AutoExtendTimeout,
		MaxRequestTimeout: cfg.MaxRequestTimeout,
	})
}
