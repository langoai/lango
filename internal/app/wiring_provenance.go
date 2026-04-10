package app

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/p2p/gitbundle"
	"github.com/langoai/lango/internal/p2p/provenanceproto"
	"github.com/langoai/lango/internal/provenance"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/langoai/lango/internal/wallet"
)

// walletBundleSigner wraps a WalletProvider to satisfy provenance.BundleSigner.
type walletBundleSigner struct {
	wp wallet.WalletProvider
}

func (s *walletBundleSigner) Sign(ctx context.Context, payload []byte) ([]byte, error) {
	return s.wp.SignMessage(ctx, payload)
}

func (s *walletBundleSigner) Algorithm() string {
	return security.AlgorithmSecp256k1Keccak256
}

func wireProvenanceRuntime(app *App, r appinit.Resolver) {
	pv, _ := r.Resolve(appinit.ProvidesProvenance).(*provenanceValues)
	if pv == nil || pv.attribution == nil || pv.bundle == nil {
		return
	}

	if wsc, ok := r.Resolve(appinit.ProvidesWorkspace).(*wsComponents); ok && wsc != nil && wsc.gitService != nil {
		wsc.gitService.SetBundleCreatedHook(func(ctx context.Context, workspaceID, headCommit string, bundleSize int) {
			authorType, authorID := provenanceAuthorFromContext(ctx, wsc.localDID)
			if err := pv.attribution.RecordWorkspaceOperation(
				ctx,
				session.SessionKeyFromContext(ctx),
				"",
				workspaceID,
				authorType,
				authorID,
				headCommit,
				"",
				provenance.AttributionSourceWorkspaceBundlePush,
				nil,
			); err != nil {
				logger().Debugw("record provenance bundle push", "workspace", workspaceID, "error", err)
			}
			if wsc.tracker != nil && wsc.localDID != "" {
				wsc.tracker.RecordCommit(workspaceID, wsc.localDID, int64(bundleSize))
			}
		})
		wsc.gitService.SetMergeHook(func(ctx context.Context, ev gitbundle.MergeHookEvent) {
			stats := make([]provenance.GitFileStat, 0, len(ev.Files))
			for _, fs := range ev.Files {
				stats = append(stats, provenance.GitFileStat{
					FilePath:     fs.FilePath,
					LinesAdded:   fs.LinesAdded,
					LinesRemoved: fs.LinesRemoved,
				})
			}
			authorType, authorID := provenanceAuthorFromContext(ctx, wsc.localDID)
			if err := pv.attribution.RecordWorkspaceOperation(
				ctx,
				session.SessionKeyFromContext(ctx),
				"",
				ev.WorkspaceID,
				authorType,
				authorID,
				ev.MergeCommit,
				ev.TaskID,
				provenance.AttributionSourceWorkspaceMerge,
				stats,
			); err != nil {
				logger().Debugw("record provenance merge", "workspace", ev.WorkspaceID, "error", err)
			}
		})
	}

	p2pc, _ := r.Resolve(appinit.ProvidesP2P).(*p2pComponents)
	if p2pc == nil || p2pc.node == nil || p2pc.sessions == nil {
		return
	}

	validator := provenanceproto.SessionValidator(func(token string) (string, bool) {
		return p2pc.sessions.GetByToken(token)
	})

	handler := provenanceproto.NewHandler(provenanceproto.HandlerConfig{
		Validator: validator,
		Importer: func(ctx context.Context, peerDID string, data []byte) error {
			if _, err := pv.bundle.Import(ctx, data); err != nil {
				return err
			}
			if err := pv.attribution.Save(ctx, provenance.Attribution{
				SessionKey: fmt.Sprintf("p2p:%s", peerDID),
				AuthorType: provenance.AuthorRemotePeer,
				AuthorID:   peerDID,
				Source:     provenance.AttributionSourceBundleImport,
			}); err != nil {
				logger().Debugw("record provenance bundle import", "peerDID", peerDID, "error", err)
			}
			return nil
		},
		Exporter: func(ctx context.Context, peerDID, sessionKey, redaction string) ([]byte, error) {
			if app == nil || app.WalletProvider == nil || p2pc.identity == nil {
				return nil, fmt.Errorf("wallet-backed DID identity is required for provenance bundle export")
			}
			did, err := p2pc.identity.DID(ctx)
			if err != nil {
				return nil, err
			}
			signer := &walletBundleSigner{wp: app.WalletProvider}
			_, data, err := pv.bundle.Export(ctx, sessionKey, provenance.RedactionLevel(redaction), did.ID, signer)
			if err != nil {
				return nil, err
			}
			return data, nil
		},
		Logger: logger().Desugar(),
	})
	p2pc.node.SetStreamHandler(provenanceproto.ProtocolID, handler.StreamHandler())
	logger().Info("registered provenance protocol handler")
}

func provenanceAuthorFromContext(ctx context.Context, fallbackDID string) (provenance.AuthorType, string) {
	if agentName := toolchain.AgentNameFromContext(ctx); agentName != "" && agentName != "user" {
		return provenance.AuthorAgent, agentName
	}
	if fallbackDID != "" {
		return provenance.AuthorRemotePeer, fallbackDID
	}
	return provenance.AuthorHuman, "unknown"
}
