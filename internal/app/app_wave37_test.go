package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"iter"
	"math/big"
	"testing"
	"time"

	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	libp2pproto "github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/handshake"
	p2pproto "github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/provenance"
	"github.com/langoai/lango/internal/provider"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/testutil"
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/langoai/lango/internal/turnrunner"
)

func TestWave37InitSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey(t *testing.T) {
	ctx := context.Background()
	client := testutil.TestEntClient(t)
	keys := security.NewKeyRegistry(client)
	crypto := security.NewLocalCryptoProvider()
	facade := storage.NewFacade(nil, nil,
		storage.WithKeyRegistryFactory(func() *security.KeyRegistry { return keys }),
		storage.WithSecretsStoreFactory(func(provider security.CryptoProvider) *security.SecretsStore {
			return security.NewSecretsStore(client, keys, provider)
		}),
	)
	cfg := config.DefaultConfig()
	cfg.Security.Signer.Provider = "local"

	gotCrypto, gotKeys, secrets, err := initSecurity(cfg, &stubSessionStore{}, &bootstrap.Result{
		Crypto:  crypto,
		Storage: facade,
	})

	require.NoError(t, err)
	assert.Same(t, crypto, gotCrypto)
	assert.Same(t, keys, gotKeys)
	require.NotNil(t, secrets)
	defaultKey, err := gotKeys.GetKey(ctx, "default")
	require.NoError(t, err)
	assert.Equal(t, "local", defaultKey.RemoteKeyID)
	assert.Equal(t, security.KeyTypeEncryption, defaultKey.Type)
}

func TestWave37ProvenanceAgentOptionsRegisterRootAndConfigCheckpoint(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	checkpoints := provenance.NewMemoryStore()
	tree := provenance.NewSessionTree(provenance.NewMemoryTreeStore())
	pv := &provenanceValues{
		checkpointService: provenance.NewCheckpointService(checkpoints, nil, config.CheckpointConfig{}),
		sessionTree:       tree,
		configMetadata:    map[string]string{"agent_model": "wave37-model"},
	}
	adkAgent, err := adk.NewAgent(
		ctx,
		nil,
		adk.NewModelAdapter(&wave37Provider{}, "wave37-model"),
		"test system prompt",
		newWave37SessionStore(),
		buildProvenanceAgentOptions(pv, nil)...,
	)
	require.NoError(t, err)

	got, err := adkAgent.RunAndCollect(ctx, "wave37-root", "hello")
	require.NoError(t, err)
	assert.Equal(t, "wave37 response", got)

	node, err := tree.GetNode(ctx, "wave37-root")
	require.NoError(t, err)
	assert.Equal(t, "wave37-root", node.SessionKey)
	assert.Equal(t, "root", node.AgentName)
	assert.Equal(t, provenance.SessionStatusActive, node.Status)

	saved, err := checkpoints.ListBySession(ctx, "wave37-root", 0)
	require.NoError(t, err)
	require.Len(t, saved, 1)
	assert.Equal(t, "session_config_snapshot", saved[0].Label)
	assert.Equal(t, "wave37-model", saved[0].Metadata["agent_model"])
}

func TestWave37WireMemoryAndTurnCallbacksTriggersAnalysisBufferAfterTurn(t *testing.T) {
	t.Parallel()

	analysis := &wave37TriggerRecorder{}
	runner := turnrunner.New(
		turnrunner.Config{HardCeiling: time.Second, StaleTimeout: -1},
		&wave37TurnExecutor{response: "turn complete"},
		newWave37SessionStore(),
		nil,
	)
	application := &App{TurnRunner: runner}

	wireMemoryAndTurnCallbacks(application, &intelligenceValues{AB: analysis}, &foundationValues{
		Store: &stubSessionStore{},
	})
	result, err := runner.Run(context.Background(), turnrunner.Request{
		SessionKey: "wave37-session",
		Input:      "run turn",
	})

	require.NoError(t, err)
	assert.Equal(t, "turn complete", result.ResponseText)
	assert.Equal(t, []string{"wave37-session"}, analysis.sessions)
}

func TestWave37WirePostAgentP2PApprovalAutoGrantsPricedSafeTool(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.ToolIsolation.Enabled = false
	sessions, err := handshake.NewSessionStore(time.Minute)
	require.NoError(t, err)
	handshakeSession, err := sessions.Create("did:lango:peer", false)
	require.NoError(t, err)
	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{Name: "wave37", Description: "Wave 37 tools", Enabled: true})
	catalog.Register("wave37", []*agent.Tool{{
		Name:        "wave37_safe",
		Description: "safe tool",
		SafetyLevel: agent.SafetyLevelSafe,
	}})
	handler := p2pproto.NewHandler(p2pproto.HandlerConfig{Sessions: sessions})
	grants := approval.NewGrantStore()
	limiter := &wave37Limiter{auto: true}
	application := &App{
		Config:      cfg,
		Gateway:     initGateway(cfg, nil, &stubSessionStore{}, nil),
		ToolCatalog: catalog,
	}

	wirePostAgent(
		application,
		staticResolver{
			appinit.ProvidesP2P:     &p2pComponents{handler: handler, pricingFn: wave37Pricing},
			appinit.ProvidesPayment: &paymentComponents{limiter: limiter},
		},
		[]*agent.Tool{{
			Name:        "wave37_safe",
			Description: "safe tool",
			SafetyLevel: agent.SafetyLevelSafe,
			Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
				return "ok", nil
			},
		}},
		eventbus.New(),
		approval.NewCompositeProvider(),
		grants,
		nil,
		nil,
	)

	request := p2pproto.Request{
		Type:         p2pproto.RequestToolInvoke,
		SessionToken: handshakeSession.Token,
		RequestID:    "wave37-tool-invoke",
		Payload: map[string]interface{}{
			"toolName": "wave37_safe",
			"params":   map[string]interface{}{"x": "y"},
		},
	}
	var input bytes.Buffer
	require.NoError(t, json.NewEncoder(&input).Encode(request))
	var output bytes.Buffer
	stream := &wave37ProtocolStream{reader: &input, writer: &output}

	handler.StreamHandler()(stream)

	var response p2pproto.Response
	require.NoError(t, json.NewDecoder(&output).Decode(&response))
	assert.Equal(t, p2pproto.ResponseStatusDenied, response.Status)
	assert.Equal(t, p2pproto.ErrNoSandboxExecutor.Error(), response.Error)
	assert.True(t, stream.closed)
	assert.True(t, grants.IsGranted("p2p:did:lango:peer", "wave37_safe"))
	require.NotNil(t, limiter.lastAmount)
	assert.Equal(t, big.NewInt(420000), limiter.lastAmount)
}

func wave37Pricing(toolName string) (string, bool) {
	if toolName == "wave37_safe" {
		return "0.42", false
	}
	return "", true
}

type wave37Provider struct{}

func (p *wave37Provider) ID() string { return "wave37" }

func (p *wave37Provider) Generate(context.Context, provider.GenerateParams) (iter.Seq2[provider.StreamEvent, error], error) {
	return func(yield func(provider.StreamEvent, error) bool) {
		if !yield(provider.StreamEvent{Type: provider.StreamEventPlainText, Text: "wave37 response"}, nil) {
			return
		}
		yield(provider.StreamEvent{Type: provider.StreamEventDone}, nil)
	}, nil
}

func (p *wave37Provider) ListModels(context.Context) ([]provider.ModelInfo, error) {
	return []provider.ModelInfo{{ID: "wave37-model", Name: "Wave 37 Model"}}, nil
}

type wave37TurnExecutor struct {
	response string
}

func (e *wave37TurnExecutor) RunStreamingDetailed(
	_ context.Context,
	_, _ string,
	onChunk adk.ChunkCallback,
	opts ...adk.RunOption,
) (adk.RunReport, error) {
	hooks := adk.ResolveRunHooks(opts...)
	defer func() {
		if hooks.OnFinish != nil {
			hooks.OnFinish()
		}
	}()
	if onChunk != nil {
		onChunk(e.response)
	}
	return adk.RunReport{Response: e.response}, nil
}

type wave37TriggerRecorder struct {
	sessions []string
}

func (r *wave37TriggerRecorder) Trigger(sessionKey string) {
	r.sessions = append(r.sessions, sessionKey)
}

type wave37Limiter struct {
	auto       bool
	lastAmount *big.Int
}

func (l *wave37Limiter) Check(context.Context, *big.Int) error  { return nil }
func (l *wave37Limiter) Record(context.Context, *big.Int) error { return nil }
func (l *wave37Limiter) DailySpent(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}
func (l *wave37Limiter) DailyRemaining(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}
func (l *wave37Limiter) IsAutoApprovable(_ context.Context, amount *big.Int) (bool, error) {
	l.lastAmount = new(big.Int).Set(amount)
	return l.auto, nil
}

type wave37SessionStore struct {
	sessions map[string]*session.Session
}

func newWave37SessionStore() *wave37SessionStore {
	return &wave37SessionStore{sessions: make(map[string]*session.Session)}
}

func (s *wave37SessionStore) Create(sess *session.Session) error {
	s.sessions[sess.Key] = sess
	return nil
}
func (s *wave37SessionStore) Get(key string) (*session.Session, error) {
	return s.sessions[key], nil
}
func (s *wave37SessionStore) Update(sess *session.Session) error {
	s.sessions[sess.Key] = sess
	return nil
}
func (s *wave37SessionStore) Delete(key string) error {
	delete(s.sessions, key)
	return nil
}
func (s *wave37SessionStore) AppendMessage(key string, msg session.Message) error {
	sess := s.sessions[key]
	if sess == nil {
		sess = &session.Session{Key: key}
		s.sessions[key] = sess
	}
	sess.History = append(sess.History, msg)
	return nil
}
func (s *wave37SessionStore) AnnotateTimeout(string, string) error { return nil }
func (s *wave37SessionStore) End(string) error                     { return nil }
func (s *wave37SessionStore) Close() error                         { return nil }
func (s *wave37SessionStore) GetSalt(string) ([]byte, error)       { return nil, nil }
func (s *wave37SessionStore) SetSalt(string, []byte) error         { return nil }
func (s *wave37SessionStore) ListSessions(context.Context) ([]session.SessionSummary, error) {
	return nil, nil
}

type wave37ProtocolStream struct {
	reader io.Reader
	writer io.Writer
	closed bool
}

func (s *wave37ProtocolStream) Read(p []byte) (int, error) {
	if s.reader == nil {
		return 0, io.EOF
	}
	return s.reader.Read(p)
}

func (s *wave37ProtocolStream) Write(p []byte) (int, error) {
	if s.writer == nil {
		return len(p), nil
	}
	return s.writer.Write(p)
}

func (s *wave37ProtocolStream) Close() error {
	s.closed = true
	return nil
}

func (s *wave37ProtocolStream) CloseWrite() error { return nil }
func (s *wave37ProtocolStream) CloseRead() error  { return nil }
func (s *wave37ProtocolStream) Reset() error      { s.closed = true; return nil }
func (s *wave37ProtocolStream) ResetWithError(network.StreamErrorCode) error {
	s.closed = true
	return nil
}
func (s *wave37ProtocolStream) SetDeadline(time.Time) error      { return nil }
func (s *wave37ProtocolStream) SetReadDeadline(time.Time) error  { return nil }
func (s *wave37ProtocolStream) SetWriteDeadline(time.Time) error { return nil }
func (s *wave37ProtocolStream) ID() string                       { return "wave37-protocol-stream" }
func (s *wave37ProtocolStream) Protocol() libp2pproto.ID         { return p2pproto.ProtocolID }
func (s *wave37ProtocolStream) SetProtocol(libp2pproto.ID) error { return nil }
func (s *wave37ProtocolStream) Stat() network.Stats              { return network.Stats{} }
func (s *wave37ProtocolStream) Conn() network.Conn               { return &wave37ProtocolConn{} }
func (s *wave37ProtocolStream) Scope() network.StreamScope       { return &network.NullScope{} }

type wave37ProtocolConn struct{}

func (c *wave37ProtocolConn) Close() error                               { return nil }
func (c *wave37ProtocolConn) CloseWithError(network.ConnErrorCode) error { return nil }
func (c *wave37ProtocolConn) ID() string                                 { return "wave37-protocol-conn" }
func (c *wave37ProtocolConn) NewStream(context.Context) (network.Stream, error) {
	return nil, errors.New("not implemented")
}
func (c *wave37ProtocolConn) GetStreams() []network.Stream { return nil }
func (c *wave37ProtocolConn) IsClosed() bool               { return false }
func (c *wave37ProtocolConn) As(any) bool                  { return false }
func (c *wave37ProtocolConn) LocalPeer() peer.ID           { return peer.ID("wave37-local") }
func (c *wave37ProtocolConn) RemotePeer() peer.ID          { return peer.ID("wave37-remote") }
func (c *wave37ProtocolConn) RemotePublicKey() libp2pcrypto.PubKey {
	return nil
}
func (c *wave37ProtocolConn) ConnState() network.ConnectionState { return network.ConnectionState{} }
func (c *wave37ProtocolConn) LocalMultiaddr() ma.Multiaddr       { return nil }
func (c *wave37ProtocolConn) RemoteMultiaddr() ma.Multiaddr      { return nil }
func (c *wave37ProtocolConn) Stat() network.ConnStats            { return network.ConnStats{} }
func (c *wave37ProtocolConn) Scope() network.ConnScope           { return &network.NullScope{} }
