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

func TestInitSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey(t *testing.T) {
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

func TestProvenanceAgentOptionsRegisterRootAndConfigCheckpoint(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	checkpoints := provenance.NewMemoryStore()
	tree := provenance.NewSessionTree(provenance.NewMemoryTreeStore())
	pv := &provenanceValues{
		checkpointService: provenance.NewCheckpointService(checkpoints, nil, config.CheckpointConfig{}),
		sessionTree:       tree,
		configMetadata:    map[string]string{"agent_model": "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7-model"},
	}
	adkAgent, err := adk.NewAgent(
		ctx,
		nil,
		adk.NewModelAdapter(&initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProvider{}, "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7-model"),
		"test system prompt",
		newInitSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeySessionStore(),
		buildProvenanceAgentOptions(pv, nil)...,
	)
	require.NoError(t, err)

	got, err := adkAgent.RunAndCollect(ctx, "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7-root", "hello")
	require.NoError(t, err)
	assert.Equal(t, "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7 response", got)

	node, err := tree.GetNode(ctx, "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7-root")
	require.NoError(t, err)
	assert.Equal(t, "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7-root", node.SessionKey)
	assert.Equal(t, "root", node.AgentName)
	assert.Equal(t, provenance.SessionStatusActive, node.Status)

	saved, err := checkpoints.ListBySession(ctx, "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7-root", 0)
	require.NoError(t, err)
	require.Len(t, saved, 1)
	assert.Equal(t, "session_config_snapshot", saved[0].Label)
	assert.Equal(t, "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7-model", saved[0].Metadata["agent_model"])
}

func TestWireMemoryAndTurnCallbacksTriggersAnalysisBufferAfterTurn(t *testing.T) {
	t.Parallel()

	analysis := &initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyTriggerRecorder{}
	runner := turnrunner.New(
		turnrunner.Config{HardCeiling: time.Second, StaleTimeout: -1},
		&initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyTurnExecutor{response: "turn complete"},
		newInitSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeySessionStore(),
		nil,
	)
	application := &App{TurnRunner: runner}

	wireMemoryAndTurnCallbacks(application, &intelligenceValues{AB: analysis}, &foundationValues{
		Store: &stubSessionStore{},
	})
	result, err := runner.Run(context.Background(), turnrunner.Request{
		SessionKey: "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7-session",
		Input:      "run turn",
	})

	require.NoError(t, err)
	assert.Equal(t, "turn complete", result.ResponseText)
	assert.Equal(t, []string{"initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7-session"}, analysis.sessions)
}

func TestWirePostAgentP2PApprovalAutoGrantsPricedSafeTool(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.ToolIsolation.Enabled = false
	sessions, err := handshake.NewSessionStore(time.Minute)
	require.NoError(t, err)
	handshakeSession, err := sessions.Create("did:lango:peer", false)
	require.NoError(t, err)
	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{Name: "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7", Description: "Bootstrap tools", Enabled: true})
	catalog.Register("initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7", []*agent.Tool{{
		Name:        "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey_safe",
		Description: "safe tool",
		SafetyLevel: agent.SafetyLevelSafe,
	}})
	handler := p2pproto.NewHandler(p2pproto.HandlerConfig{Sessions: sessions})
	grants := approval.NewGrantStore()
	limiter := &initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyLimiter{auto: true}
	application := &App{
		Config:      cfg,
		Gateway:     initGateway(cfg, nil, &stubSessionStore{}, nil),
		ToolCatalog: catalog,
	}

	wirePostAgent(
		application,
		staticResolver{
			appinit.ProvidesP2P:     &p2pComponents{handler: handler, pricingFn: initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyPricing},
			appinit.ProvidesPayment: &paymentComponents{limiter: limiter},
		},
		[]*agent.Tool{{
			Name:        "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey_safe",
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
		RequestID:    "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7-tool-invoke",
		Payload: map[string]interface{}{
			"toolName": "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey_safe",
			"params":   map[string]interface{}{"x": "y"},
		},
	}
	var input bytes.Buffer
	require.NoError(t, json.NewEncoder(&input).Encode(request))
	var output bytes.Buffer
	stream := &initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream{reader: &input, writer: &output}

	handler.StreamHandler()(stream)

	var response p2pproto.Response
	require.NoError(t, json.NewDecoder(&output).Decode(&response))
	assert.Equal(t, p2pproto.ResponseStatusDenied, response.Status)
	assert.Equal(t, p2pproto.ErrNoSandboxExecutor.Error(), response.Error)
	assert.True(t, stream.closed)
	assert.True(t, grants.IsGranted("p2p:did:lango:peer", "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey_safe"))
	require.NotNil(t, limiter.lastAmount)
	assert.Equal(t, big.NewInt(420000), limiter.lastAmount)
}

func initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyPricing(toolName string) (string, bool) {
	if toolName == "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey_safe" {
		return "0.42", false
	}
	return "", true
}

type initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProvider struct{}

func (p *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProvider) ID() string {
	return "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7"
}

func (p *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProvider) Generate(context.Context, provider.GenerateParams) (iter.Seq2[provider.StreamEvent, error], error) {
	return func(yield func(provider.StreamEvent, error) bool) {
		if !yield(provider.StreamEvent{Type: provider.StreamEventPlainText, Text: "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7 response"}, nil) {
			return
		}
		yield(provider.StreamEvent{Type: provider.StreamEventDone}, nil)
	}, nil
}

func (p *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProvider) ListModels(context.Context) ([]provider.ModelInfo, error) {
	return []provider.ModelInfo{{ID: "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7-model", Name: "Bootstrap Model"}}, nil
}

type initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyTurnExecutor struct {
	response string
}

func (e *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyTurnExecutor) RunStreamingDetailed(
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

type initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyTriggerRecorder struct {
	sessions []string
}

func (r *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyTriggerRecorder) Trigger(sessionKey string) {
	r.sessions = append(r.sessions, sessionKey)
}

type initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyLimiter struct {
	auto       bool
	lastAmount *big.Int
}

func (l *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyLimiter) Check(context.Context, *big.Int) error {
	return nil
}
func (l *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyLimiter) Record(context.Context, *big.Int) error {
	return nil
}
func (l *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyLimiter) DailySpent(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}
func (l *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyLimiter) DailyRemaining(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}
func (l *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyLimiter) IsAutoApprovable(_ context.Context, amount *big.Int) (bool, error) {
	l.lastAmount = new(big.Int).Set(amount)
	return l.auto, nil
}

type initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeySessionStore struct {
	sessions map[string]*session.Session
}

func newInitSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeySessionStore() *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeySessionStore {
	return &initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeySessionStore{sessions: make(map[string]*session.Session)}
}

func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeySessionStore) Create(sess *session.Session) error {
	s.sessions[sess.Key] = sess
	return nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeySessionStore) Get(key string) (*session.Session, error) {
	return s.sessions[key], nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeySessionStore) Update(sess *session.Session) error {
	s.sessions[sess.Key] = sess
	return nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeySessionStore) Delete(key string) error {
	delete(s.sessions, key)
	return nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeySessionStore) AppendMessage(key string, msg session.Message) error {
	sess := s.sessions[key]
	if sess == nil {
		sess = &session.Session{Key: key}
		s.sessions[key] = sess
	}
	sess.History = append(sess.History, msg)
	return nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeySessionStore) AnnotateTimeout(string, string) error {
	return nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeySessionStore) End(string) error {
	return nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeySessionStore) Close() error {
	return nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeySessionStore) GetSalt(string) ([]byte, error) {
	return nil, nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeySessionStore) SetSalt(string, []byte) error {
	return nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeySessionStore) ListSessions(context.Context) ([]session.SessionSummary, error) {
	return nil, nil
}

type initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream struct {
	reader io.Reader
	writer io.Writer
	closed bool
}

func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream) Read(p []byte) (int, error) {
	if s.reader == nil {
		return 0, io.EOF
	}
	return s.reader.Read(p)
}

func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream) Write(p []byte) (int, error) {
	if s.writer == nil {
		return len(p), nil
	}
	return s.writer.Write(p)
}

func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream) Close() error {
	s.closed = true
	return nil
}

func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream) CloseWrite() error {
	return nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream) CloseRead() error {
	return nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream) Reset() error {
	s.closed = true
	return nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream) ResetWithError(network.StreamErrorCode) error {
	s.closed = true
	return nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream) SetDeadline(time.Time) error {
	return nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream) SetReadDeadline(time.Time) error {
	return nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream) SetWriteDeadline(time.Time) error {
	return nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream) ID() string {
	return "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7-protocol-stream"
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream) Protocol() libp2pproto.ID {
	return p2pproto.ProtocolID
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream) SetProtocol(libp2pproto.ID) error {
	return nil
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream) Stat() network.Stats {
	return network.Stats{}
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream) Conn() network.Conn {
	return &initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolConn{}
}
func (s *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolStream) Scope() network.StreamScope {
	return &network.NullScope{}
}

type initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolConn struct{}

func (c *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolConn) Close() error {
	return nil
}
func (c *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolConn) CloseWithError(network.ConnErrorCode) error {
	return nil
}
func (c *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolConn) ID() string {
	return "initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7-protocol-conn"
}
func (c *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolConn) NewStream(context.Context) (network.Stream, error) {
	return nil, errors.New("not implemented")
}
func (c *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolConn) GetStreams() []network.Stream {
	return nil
}
func (c *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolConn) IsClosed() bool {
	return false
}
func (c *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolConn) As(any) bool {
	return false
}
func (c *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolConn) LocalPeer() peer.ID {
	return peer.ID("initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7-local")
}
func (c *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolConn) RemotePeer() peer.ID {
	return peer.ID("initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKey7-remote")
}
func (c *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolConn) RemotePublicKey() libp2pcrypto.PubKey {
	return nil
}
func (c *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolConn) ConnState() network.ConnectionState {
	return network.ConnectionState{}
}
func (c *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolConn) LocalMultiaddr() ma.Multiaddr {
	return nil
}
func (c *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolConn) RemoteMultiaddr() ma.Multiaddr {
	return nil
}
func (c *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolConn) Stat() network.ConnStats {
	return network.ConnStats{}
}
func (c *initSecurityLocalReusesBootstrapCryptoAndRegistersDefaultKeyProtocolConn) Scope() network.ConnScope {
	return &network.NullScope{}
}
