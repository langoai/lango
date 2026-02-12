package app

import (
	"context"
	"io"
	"sync"

	"github.com/langowarny/lango/internal/adk"
	"github.com/langowarny/lango/internal/config"
	"github.com/langowarny/lango/internal/gateway"
	"github.com/langowarny/lango/internal/knowledge"
	"github.com/langowarny/lango/internal/learning"
	"github.com/langowarny/lango/internal/session"
	"github.com/langowarny/lango/internal/skill"
)

// App is the root application structure
type App struct {
	Config *config.Config

	// Core Components
	Agent   *adk.Agent
	Gateway *gateway.Server
	Store   session.Store

	// Browser (optional, io.Closer)
	Browser io.Closer

	// Self-Learning Components
	KnowledgeStore  *knowledge.Store
	LearningEngine  *learning.Engine
	SkillRegistry   *skill.Registry

	// Channels
	Channels []Channel

	// wg tracks background goroutines for graceful shutdown
	wg sync.WaitGroup
}

// Channel represents a communication channel (Telegram, Discord, Slack)
type Channel interface {
	Start(ctx context.Context) error
	Stop()
}
