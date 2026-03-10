package app

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/p2p"
	"github.com/langoai/lango/internal/p2p/gitbundle"
	"github.com/langoai/lango/internal/p2p/workspace"
	bolt "go.etcd.io/bbolt"
)

// wsComponents holds initialized workspace components.
type wsComponents struct {
	manager    *workspace.Manager
	gitService *gitbundle.Service
	gitHandler *gitbundle.Handler
	gossip     *workspace.WorkspaceGossip
	chronicler *workspace.Chronicler
	tracker    *workspace.ContributionTracker
	db         *bolt.DB
}

// initWorkspace creates workspace and git bundle components if enabled.
func initWorkspace(cfg *config.Config, node *p2p.Node, localDID string, sessionValidator gitbundle.SessionValidator) *wsComponents {
	wsCfg := cfg.P2P.Workspace
	if !wsCfg.Enabled {
		logger().Info("P2P workspace disabled")
		return nil
	}

	log := logger()

	// Resolve data directory.
	dataDir := wsCfg.DataDir
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Warnw("resolve home dir for workspace data", "error", err)
			return nil
		}
		dataDir = filepath.Join(home, ".lango", "workspaces")
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		log.Warnw("create workspace data dir", "error", err)
		return nil
	}

	// Open BoltDB for workspace persistence.
	dbPath := filepath.Join(dataDir, "workspaces.db")
	db, err := bolt.Open(dbPath, 0o600, nil)
	if err != nil {
		log.Warnw("open workspace BoltDB", "path", dbPath, "error", err)
		return nil
	}

	// Create workspace manager.
	maxWS := wsCfg.MaxWorkspaces
	if maxWS <= 0 {
		maxWS = 10
	}
	mgr, err := workspace.NewManager(workspace.ManagerConfig{
		DB:            db,
		LocalDID:      localDID,
		MaxWorkspaces: maxWS,
		Logger:        log,
	})
	if err != nil {
		log.Warnw("create workspace manager", "error", err)
		db.Close()
		return nil
	}

	// Create bare repo store and git bundle service.
	zapLogger := zap.L()
	repoStore := gitbundle.NewBareRepoStore(dataDir, zapLogger)
	gitSvc := gitbundle.NewService(repoStore, zapLogger)

	// Create git protocol handler.
	maxBundle := wsCfg.MaxBundleSizeBytes
	if maxBundle <= 0 {
		maxBundle = 50 * 1024 * 1024 // 50MB
	}
	gitHdl := gitbundle.NewHandler(gitbundle.HandlerConfig{
		Service:       gitSvc,
		Validator:     sessionValidator,
		MaxBundleSize: maxBundle,
		Logger:        zapLogger,
	})

	// Register git protocol stream handler on the P2P node.
	node.SetStreamHandler(gitbundle.ProtocolID, gitHdl.StreamHandler())
	log.Infow("registered git protocol handler", "protocol", gitbundle.ProtocolID)

	// Create workspace gossip (per-workspace GossipSub topics).
	var wsGossip *workspace.WorkspaceGossip
	ps, err := node.PubSub()
	if err != nil {
		log.Warnw("get PubSub for workspace gossip", "error", err)
	} else {
		wsGossip = workspace.NewWorkspaceGossip(workspace.GossipConfig{
			PubSub:  ps,
			LocalID: node.PeerID(),
			Logger:  log,
		})
	}

	// Create contribution tracker if enabled.
	var tracker *workspace.ContributionTracker
	if wsCfg.ContributionTracking {
		tracker = workspace.NewContributionTracker()
		log.Info("workspace contribution tracking enabled")
	}

	// Create chronicler if enabled and gossip handler is available.
	var chronicler *workspace.Chronicler
	if wsCfg.ChroniclerEnabled {
		// Chronicler uses a callback to avoid direct graph store import.
		// Actual triple adder will be wired in app.go if graph store is available.
		chronicler = workspace.NewChronicler(nil, log)
		log.Info("workspace chronicler enabled (triple adder pending)")
	}

	// Wire gossip message handler to chronicler and tracker.
	if wsGossip != nil {
		wsGossip = workspace.NewWorkspaceGossip(workspace.GossipConfig{
			PubSub:  ps,
			LocalID: node.PeerID(),
			Handler: func(msg workspace.Message) {
				if chronicler != nil {
					chronicler.HandleMessage(msg)
				}
				if tracker != nil {
					tracker.RecordMessage(msg.WorkspaceID, msg.SenderDID)
				}
			},
			Logger: log,
		})
	}

	log.Infow("P2P workspace initialized",
		"dataDir", dataDir,
		"maxWorkspaces", maxWS,
		"contributionTracking", wsCfg.ContributionTracking,
		"chronicler", wsCfg.ChroniclerEnabled,
	)

	return &wsComponents{
		manager:    mgr,
		gitService: gitSvc,
		gitHandler: gitHdl,
		gossip:     wsGossip,
		chronicler: chronicler,
		tracker:    tracker,
		db:         db,
	}
}
