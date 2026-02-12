package app

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/agent"
	"github.com/langowarny/lango/internal/config"
	"github.com/langowarny/lango/internal/logging"
	"github.com/langowarny/lango/internal/tools/browser"
	"github.com/langowarny/lango/internal/tools/filesystem"
)

func logger() *zap.SugaredLogger { return logging.App() }

// New creates a new application instance
func New(cfg *config.Config) (*App, error) {
	app := &App{
		Config: cfg,
	}

	// 1. Supervisor (holds provider secrets, exec tool)
	sv, err := initSupervisor(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create supervisor: %w", err)
	}

	// 2. Session Store
	store, err := initSessionStore(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create session store: %w", err)
	}
	app.Store = store

	// 3. Security (optional, non-blocking)
	if cfg.Security.Passphrase != "" {
		logger().Warn("security.passphrase in config is deprecated; use LANGO_PASSPHRASE env var")
	}
	if cfg.Security.Signer.Provider == "" {
		logger().Info("security disabled, set security.signer.provider to enable")
	}

	// 4. Base tools (exec + filesystem + optional browser)
	fsConfig := filesystem.Config{
		MaxReadSize:  cfg.Tools.Filesystem.MaxReadSize,
		AllowedPaths: cfg.Tools.Filesystem.AllowedPaths,
	}

	var browserSM *browser.SessionManager
	if cfg.Tools.Browser.Enabled {
		bt, err := browser.New(browser.Config{
			Headless:       cfg.Tools.Browser.Headless,
			SessionTimeout: cfg.Tools.Browser.SessionTimeout,
		})
		if err != nil {
			return nil, fmt.Errorf("create browser tool: %w", err)
		}
		browserSM = browser.NewSessionManager(bt)
		app.Browser = browserSM
		logger().Info("browser tools enabled")
	}

	tools := buildTools(sv, fsConfig, browserSM)

	// 5. Knowledge system (optional, non-blocking)
	kc := initKnowledge(cfg, store, tools)
	if kc != nil {
		app.KnowledgeStore = kc.store
		app.LearningEngine = kc.engine
		app.SkillRegistry = kc.registry

		// Wrap base tools with learning engine
		wrapped := make([]*agent.Tool, len(tools))
		for i, t := range tools {
			wrapped[i] = wrapWithLearning(t, kc.engine)
		}
		tools = wrapped

		// Add dynamic skills from registry
		tools = append(tools, kc.registry.AllTools()...)

		// Add meta-tools
		metaTools := buildMetaTools(kc.store, kc.engine, kc.registry, cfg.Knowledge.AutoApproveSkills)
		tools = append(tools, metaTools...)
	}

	// 6. ADK Agent
	adkAgent, err := initAgent(context.Background(), sv, cfg, store, tools, kc)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}
	app.Agent = adkAgent

	// 7. Channels
	if err := app.initChannels(); err != nil {
		logger().Errorw("failed to initialize channels", "error", err)
	}

	// 8. Gateway
	app.Gateway = initGateway(cfg, app.Agent, app.Store)

	return app, nil
}

// Start starts the application services
func (a *App) Start(ctx context.Context) error {
	logger().Info("starting application")

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.Gateway.Start(); err != nil {
			logger().Errorw("gateway server error", "error", err)
		}
	}()

	logger().Info("starting channels...")
	for _, ch := range a.Channels {
		a.wg.Add(1)
		go func(c Channel) {
			defer a.wg.Done()
			if err := c.Start(ctx); err != nil {
				logger().Errorw("channel start error", "error", err)
			}
		}(ch)
	}

	return nil
}

// Stop stops the application services and waits for all goroutines to exit.
func (a *App) Stop(ctx context.Context) error {
	logger().Info("stopping application")

	// Signal gateway and channels to stop
	if err := a.Gateway.Shutdown(ctx); err != nil {
		logger().Errorw("gateway shutdown error", "error", err)
	}

	for _, ch := range a.Channels {
		ch.Stop()
	}

	// Wait for all background goroutines to finish
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger().Info("all services stopped")
	case <-ctx.Done():
		logger().Warnw("shutdown timed out waiting for services", "error", ctx.Err())
	}

	if a.Browser != nil {
		if err := a.Browser.Close(); err != nil {
			logger().Errorw("browser close error", "error", err)
		}
	}

	if a.Store != nil {
		if err := a.Store.Close(); err != nil {
			logger().Errorw("session store close error", "error", err)
		}
	}

	return nil
}
