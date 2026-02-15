package supervisor

import (
	"context"
	"fmt"
	"iter"

	"github.com/langowarny/lango/internal/config"
	"github.com/langowarny/lango/internal/logging"
	"github.com/langowarny/lango/internal/provider"
	"github.com/langowarny/lango/internal/provider/anthropic"
	"github.com/langowarny/lango/internal/provider/gemini"
	"github.com/langowarny/lango/internal/provider/openai"
	"github.com/langowarny/lango/internal/tools/exec"
)

var logger = logging.SubsystemSugar("supervisor")

// Supervisor is the root component that manages secrets and lifecycle.
type Supervisor struct {
	Config   *config.Config
	registry *provider.Registry
	execTool *exec.Tool
}

// New creates a new Supervisor.
func New(cfg *config.Config) (*Supervisor, error) {
	s := &Supervisor{
		Config:   cfg,
		registry: provider.NewRegistry(),
	}

	// Initialize Exec Tool with Whitelist
	execConfig := exec.Config{
		DefaultTimeout:  cfg.Tools.Exec.DefaultTimeout,
		AllowBackground: cfg.Tools.Exec.AllowBackground,
		WorkDir:         cfg.Tools.Exec.WorkDir,
		EnvWhitelist: []string{
			"PATH", "HOME", "USER", "LANG", "LC_ALL", "LC_CTYPE", "TERM",
			"SHELL", "TMPDIR", "SSH_AUTH_SOCK",
		},
	}
	s.execTool = exec.New(execConfig)

	if err := s.initializeProviders(); err != nil {
		return nil, err
	}

	return s, nil
}

// initializeProviders sets up the AI providers with secrets from config.
func (s *Supervisor) initializeProviders() error {
	if len(s.Config.Providers) > 0 {
		for id, pCfg := range s.Config.Providers {
			var p provider.Provider
			var err error

			apiKey := pCfg.APIKey
			if apiKey == "" {
				logger.Warnw("provider has no API key configured", "id", id)
			}

			switch pCfg.Type {
			case "openai":
				p = openai.NewProvider(id, apiKey, pCfg.BaseURL)
			case "anthropic":
				p = anthropic.NewProvider(id, apiKey)
			case "gemini", "google": // Support "google" as alias
				p, err = gemini.NewProvider(context.Background(), id, apiKey, "")
			case "ollama":
				baseURL := pCfg.BaseURL
				if baseURL == "" {
					baseURL = "http://localhost:11434/v1"
				}
				p = openai.NewProvider(id, apiKey, baseURL)
			case "github":
				// GitHub Models uses OpenAI compatible endpoint
				baseURL := pCfg.BaseURL
				if baseURL == "" {
					baseURL = "https://models.inference.ai.azure.com"
				}
				p = openai.NewProvider(id, apiKey, baseURL)
			default:
				logger.Warnw("unknown provider type", "id", id, "type", pCfg.Type)
				continue
			}

			if err != nil {
				logger.Warnw("failed to initialize provider", "id", id, "error", err)
				continue
			}
			s.registry.Register(p)
		}
	}

	// Verify that the default provider is configured
	if s.Config.Agent.Provider != "" {
		if _, ok := s.registry.Get(s.Config.Agent.Provider); !ok {
			return fmt.Errorf("configured default provider '%s' not found in providers list", s.Config.Agent.Provider)
		}
	} else if len(s.Config.Providers) == 1 {
		// Auto-select the only provider if default not set
		for id := range s.Config.Providers {
			s.Config.Agent.Provider = id
			logger.Infow("auto-selected default provider", "provider", id)
		}
	}

	return nil
}

// Generate forwards a generation request to the appropriate provider.
// This is called by the Runtime via the Proxy.
func (s *Supervisor) Generate(ctx context.Context, providerID, model string, params provider.GenerateParams) (iter.Seq2[provider.StreamEvent, error], error) {
	// If providerID is empty, try to use default from config if available (though runtime should usually know)
	if providerID == "" {
		providerID = s.Config.Agent.Provider
	}

	p, ok := s.registry.Get(providerID)
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", providerID)
	}

	// Ensure model is set if not provided in params
	if params.Model == "" {
		params.Model = model // Use the default model for this provider if known, or from config
	}

	logger.Infow("proxying generation request", "provider", providerID, "model", params.Model)
	return p.Generate(ctx, params)
}

// ExecuteTool forwards a command execution request to the internal exec tool.
// Importantly, the exec tool is configured with an environment whitelist in New(),
// ensuring that sensitive secrets (like API keys) are NOT passed to the command.
func (s *Supervisor) ExecuteTool(ctx context.Context, cmd string) (string, error) {
	logger.Infow("supervisor executing potentially privileged command", "command", cmd)

	// We use the default timeout configured in s.execTool
	res, err := s.execTool.Run(ctx, cmd, 0)
	if err != nil {
		return "", err
	}

	// If exit code is non-zero but no error returned (e.g. command failed),
	// we should probably return stdout+stderr anyway so the agent sees why.
	// exec.Tool.Run returns nil error on exit code != 0?
	// Checking exec.go: "if err != nil { ... returns result.ExitCode ... returns nil"
	// So Run returns result and nil error even if exit code is non-zero, unless it's a different error.
	// Wait, "if err != nil { if exitErr ... result.ExitCode = ... return result, nil } else { return nil, err }"
	// So if it's an ExitError, it returns (result, nil).

	if res.ExitCode != 0 {
		return fmt.Sprintf("Command failed with exit code %d\nOutput:\n%s\n%s", res.ExitCode, res.Stdout, res.Stderr), nil
	}

	return res.Stdout + res.Stderr, nil
}

// StartBackground starts a command in the background.
func (s *Supervisor) StartBackground(cmd string) (string, error) {
	logger.Infow("supervisor starting background process", "command", cmd)
	return s.execTool.StartBackground(cmd)
}

// GetBackgroundStatus returns the status of a background process.
func (s *Supervisor) GetBackgroundStatus(id string) (map[string]interface{}, error) {
	bp, err := s.execTool.GetBackgroundStatus(id)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":        bp.ID,
		"command":   bp.Command,
		"done":      bp.Done,
		"exitCode":  bp.ExitCode,
		"error":     bp.Error,
		"output":    bp.Output.String(),
		"startTime": bp.StartTime,
	}, nil
}

// StopBackground stops a background process.
func (s *Supervisor) StopBackground(id string) error {
	logger.Infow("supervisor stopping background process", "id", id)
	return s.execTool.StopBackground(id)
}
