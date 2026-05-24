package protocol

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/firewall"
	"github.com/langoai/lango/internal/p2p/handshake"
)

type settersAndBasicDispatchBranchesPayGate struct {
	result PayGateResult
	err    error
	calls  int
	seen   []string
}

func (g *settersAndBasicDispatchBranchesPayGate) Check(peerDID, toolName string, payload map[string]interface{}) (PayGateResult, error) {
	g.calls++
	g.seen = append(g.seen, peerDID+":"+toolName)
	return g.result, g.err
}

type settersAndBasicDispatchBranchesSecurityEvents struct {
	successes []string
	failures  []string
}

func (e *settersAndBasicDispatchBranchesSecurityEvents) RecordToolFailure(peerDID string) {
	e.failures = append(e.failures, peerDID)
}

func (e *settersAndBasicDispatchBranchesSecurityEvents) RecordToolSuccess(peerDID string) {
	e.successes = append(e.successes, peerDID)
}

type settersAndBasicDispatchBranchesOntologyHandler struct {
	queryResp   *SchemaQueryResponse
	proposeResp *SchemaProposeResponse
	err         error
	querySeen   SchemaQueryRequest
	propSeen    SchemaProposeRequest
	peerSeen    string
}

func (h *settersAndBasicDispatchBranchesOntologyHandler) HandleSchemaQuery(_ context.Context, peerDID string, req SchemaQueryRequest) (*SchemaQueryResponse, error) {
	h.peerSeen = peerDID
	h.querySeen = req
	return h.queryResp, h.err
}

func (h *settersAndBasicDispatchBranchesOntologyHandler) HandleSchemaPropose(_ context.Context, peerDID string, req SchemaProposeRequest) (*SchemaProposeResponse, error) {
	h.peerSeen = peerDID
	h.propSeen = req
	return h.proposeResp, h.err
}

type settersAndBasicDispatchBranchesProtocolStream struct {
	rw     io.ReadWriter
	closed bool
}

func (s *settersAndBasicDispatchBranchesProtocolStream) Read(p []byte) (int, error) {
	return s.rw.Read(p)
}
func (s *settersAndBasicDispatchBranchesProtocolStream) Write(p []byte) (int, error) {
	return s.rw.Write(p)
}
func (s *settersAndBasicDispatchBranchesProtocolStream) Close() error      { s.closed = true; return nil }
func (s *settersAndBasicDispatchBranchesProtocolStream) CloseWrite() error { return nil }
func (s *settersAndBasicDispatchBranchesProtocolStream) CloseRead() error  { return nil }
func (s *settersAndBasicDispatchBranchesProtocolStream) Reset() error      { s.closed = true; return nil }
func (s *settersAndBasicDispatchBranchesProtocolStream) ResetWithError(network.StreamErrorCode) error {
	s.closed = true
	return nil
}
func (s *settersAndBasicDispatchBranchesProtocolStream) SetDeadline(time.Time) error      { return nil }
func (s *settersAndBasicDispatchBranchesProtocolStream) SetReadDeadline(time.Time) error  { return nil }
func (s *settersAndBasicDispatchBranchesProtocolStream) SetWriteDeadline(time.Time) error { return nil }
func (s *settersAndBasicDispatchBranchesProtocolStream) ID() string {
	return "onChainEscrowToolsRunLifecycleAndQueryViews0-protocol-stream"
}
func (s *settersAndBasicDispatchBranchesProtocolStream) Protocol() protocol.ID         { return ProtocolID }
func (s *settersAndBasicDispatchBranchesProtocolStream) SetProtocol(protocol.ID) error { return nil }
func (s *settersAndBasicDispatchBranchesProtocolStream) Stat() network.Stats           { return network.Stats{} }
func (s *settersAndBasicDispatchBranchesProtocolStream) Conn() network.Conn {
	return &settersAndBasicDispatchBranchesProtocolConn{}
}
func (s *settersAndBasicDispatchBranchesProtocolStream) Scope() network.StreamScope {
	return &network.NullScope{}
}

type settersAndBasicDispatchBranchesProtocolConn struct{}

func (c *settersAndBasicDispatchBranchesProtocolConn) Close() error { return nil }
func (c *settersAndBasicDispatchBranchesProtocolConn) CloseWithError(network.ConnErrorCode) error {
	return nil
}
func (c *settersAndBasicDispatchBranchesProtocolConn) ID() string {
	return "onChainEscrowToolsRunLifecycleAndQueryViews0-protocol-conn"
}
func (c *settersAndBasicDispatchBranchesProtocolConn) NewStream(context.Context) (network.Stream, error) {
	return nil, errors.New("not implemented")
}
func (c *settersAndBasicDispatchBranchesProtocolConn) GetStreams() []network.Stream { return nil }
func (c *settersAndBasicDispatchBranchesProtocolConn) IsClosed() bool               { return false }
func (c *settersAndBasicDispatchBranchesProtocolConn) As(any) bool                  { return false }
func (c *settersAndBasicDispatchBranchesProtocolConn) LocalPeer() peer.ID {
	return peer.ID("onChainEscrowToolsRunLifecycleAndQueryViews0-local")
}
func (c *settersAndBasicDispatchBranchesProtocolConn) RemotePeer() peer.ID {
	return peer.ID("onChainEscrowToolsRunLifecycleAndQueryViews0-remote")
}
func (c *settersAndBasicDispatchBranchesProtocolConn) RemotePublicKey() crypto.PubKey { return nil }
func (c *settersAndBasicDispatchBranchesProtocolConn) ConnState() network.ConnectionState {
	return network.ConnectionState{}
}
func (c *settersAndBasicDispatchBranchesProtocolConn) LocalMultiaddr() ma.Multiaddr {
	return ma.StringCast("/ip4/127.0.0.1/tcp/11001")
}
func (c *settersAndBasicDispatchBranchesProtocolConn) RemoteMultiaddr() ma.Multiaddr {
	return ma.StringCast("/ip4/127.0.0.1/tcp/11002")
}
func (c *settersAndBasicDispatchBranchesProtocolConn) Stat() network.ConnStats {
	return network.ConnStats{}
}
func (c *settersAndBasicDispatchBranchesProtocolConn) Scope() network.ConnScope {
	return &network.NullScope{}
}

type settersAndBasicDispatchBranchesSplitReadWriter struct {
	reader io.Reader
	writer io.Writer
}

func (rw *settersAndBasicDispatchBranchesSplitReadWriter) Read(p []byte) (int, error) {
	return rw.reader.Read(p)
}

func (rw *settersAndBasicDispatchBranchesSplitReadWriter) Write(p []byte) (int, error) {
	return rw.writer.Write(p)
}

func settersAndBasicDispatchBranchesHandler(t *testing.T, peerDID string) (*Handler, *handshake.SessionStore, string) {
	t.Helper()
	sessions, err := handshake.NewSessionStore(time.Hour)
	require.NoError(t, err)
	fw := firewall.New([]firewall.ACLRule{
		{PeerDID: peerDID, Action: firewall.ACLActionAllow, Tools: []string{firewall.WildcardAll}},
	}, zap.NewNop().Sugar())
	h := NewHandler(HandlerConfig{
		Sessions: sessions,
		Firewall: fw,
		Executor: func(_ context.Context, toolName string, _ map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{"defaultExecutor": toolName}, nil
		},
		LocalDID: "did:key:local-onChainEscrowToolsRunLifecycleAndQueryViews0",
		Logger:   zap.NewNop().Sugar(),
	})
	sess, err := sessions.Create(peerDID, false)
	require.NoError(t, err)
	return h, sessions, sess.Token
}

func TestSettersAndBasicDispatchBranches(t *testing.T) {
	t.Parallel()

	peerDID := "did:key:onChainEscrowToolsRunLifecycleAndQueryViews0-setters"
	h, _, token := settersAndBasicDispatchBranchesHandler(t, peerDID)

	exec := func(context.Context, string, map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"ok": true}, nil
	}
	sandbox := func(context.Context, string, map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"sandbox": true}, nil
	}
	events := &settersAndBasicDispatchBranchesSecurityEvents{}
	bus := eventbus.New()
	payGate := &settersAndBasicDispatchBranchesPayGate{result: PayGateResult{Status: payGateStatusFree}}
	negotiator := func(context.Context, string, NegotiatePayload) (map[string]interface{}, error) {
		return map[string]interface{}{"negotiated": true}, nil
	}
	team := func(context.Context, string, RequestType, map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"team": true}, nil
	}
	ontology := &settersAndBasicDispatchBranchesOntologyHandler{}
	approval := func(context.Context, string, string, map[string]interface{}) (bool, error) { return true, nil }
	safety := func(toolName string) (int, bool) { return 2, true }

	h.SetExecutor(exec)
	h.SetSandboxExecutor(sandbox)
	h.SetSecurityEvents(events)
	h.SetEventBus(bus)
	h.SetPayGate(payGate)
	h.SetApprovalFunc(approval)
	h.SetNegotiator(negotiator)
	h.SetTeamHandler(team)
	h.SetOntologyHandler(ontology)
	h.SetSafetyGate(safety, 1, []string{"allowed-tool"})

	assert.NotNil(t, h.executor)
	assert.NotNil(t, h.sandboxExec)
	assert.Same(t, events, h.securityEvents)
	assert.Same(t, bus, h.eventBus)
	assert.Same(t, payGate, h.payGate)
	assert.NotNil(t, h.approvalFn)
	assert.NotNil(t, h.negotiator)
	assert.NotNil(t, h.teamHandler)
	assert.Same(t, ontology, h.ontologyHandler)
	assert.False(t, h.checkSafetyGate("blocked-tool"))
	assert.True(t, h.checkSafetyGate("allowed-tool"))

	invalid := h.handleRequest(context.Background(), nil, &Request{Type: RequestAgentCard, SessionToken: "bad", RequestID: "bad"})
	assert.Equal(t, ResponseStatusDenied, invalid.Status)
	assert.Equal(t, ErrInvalidSession.Error(), invalid.Error)

	cardMissing := h.handleRequest(context.Background(), nil, &Request{Type: RequestAgentCard, SessionToken: token, RequestID: "card-missing"})
	assert.Equal(t, ResponseStatusError, cardMissing.Status)
	assert.Equal(t, ErrAgentCardUnavailable.Error(), cardMissing.Error)

	capFallback := h.handleRequest(context.Background(), nil, &Request{Type: RequestCapabilityQuery, SessionToken: token, RequestID: "cap"})
	assert.Equal(t, ResponseStatusOK, capFallback.Status)
	assert.Equal(t, []string{}, capFallback.Result["capabilities"])

	h.cardFn = func() map[string]interface{} {
		return map[string]interface{}{"name": "onChainEscrowToolsRunLifecycleAndQueryViews0-agent", "capabilities": []interface{}{"search"}}
	}
	card := h.handleRequest(context.Background(), nil, &Request{Type: RequestAgentCard, SessionToken: token, RequestID: "card-ok"})
	assert.Equal(t, ResponseStatusOK, card.Status)
	assert.Equal(t, "onChainEscrowToolsRunLifecycleAndQueryViews0-agent", card.Result["name"])

	unknown := h.handleRequest(context.Background(), nil, &Request{Type: RequestType("unknown_request"), SessionToken: token, RequestID: "unknown"})
	assert.Equal(t, ResponseStatusError, unknown.Status)
	assert.Contains(t, unknown.Error, "unknown request type")
}

func TestPriceQueryBranches(t *testing.T) {
	t.Parallel()

	peerDID := "did:key:onChainEscrowToolsRunLifecycleAndQueryViews0-price"
	h, _, token := settersAndBasicDispatchBranchesHandler(t, peerDID)

	missingTool := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestPriceQuery,
		SessionToken: token,
		RequestID:    "price-missing",
		Payload:      map[string]interface{}{},
	})
	assert.Equal(t, ResponseStatusError, missingTool.Status)
	assert.Equal(t, ErrMissingToolName.Error(), missingTool.Error)

	freeDefault := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestPriceQuery,
		SessionToken: token,
		RequestID:    "price-free-default",
		Payload:      map[string]interface{}{"toolName": "free-tool"},
	})
	assert.Equal(t, ResponseStatusOK, freeDefault.Status)
	assert.Equal(t, true, freeDefault.Result["isFree"])

	h.SetPayGate(&settersAndBasicDispatchBranchesPayGate{err: errors.New("meter down")})
	payGateError := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestPriceQuery,
		SessionToken: token,
		RequestID:    "price-error",
		Payload:      map[string]interface{}{"toolName": "priced-tool"},
	})
	assert.Equal(t, ResponseStatusError, payGateError.Status)
	assert.Contains(t, payGateError.Error, "meter down")

	h.SetPayGate(&settersAndBasicDispatchBranchesPayGate{result: PayGateResult{Status: payGateStatusFree}})
	freeGate := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestPriceQuery,
		SessionToken: token,
		RequestID:    "price-free-gate",
		Payload:      map[string]interface{}{"toolName": "priced-tool"},
	})
	assert.Equal(t, ResponseStatusOK, freeGate.Status)
	assert.Equal(t, true, freeGate.Result["isFree"])

	quote := map[string]interface{}{"toolName": "priced-tool", "price": "1.25", "currency": "USDC"}
	h.SetPayGate(&settersAndBasicDispatchBranchesPayGate{result: PayGateResult{Status: payGateStatusPaymentRequired, PriceQuote: quote}})
	quoted := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestPriceQuery,
		SessionToken: token,
		RequestID:    "price-quote",
		Payload:      map[string]interface{}{"toolName": "priced-tool"},
	})
	assert.Equal(t, ResponseStatusOK, quoted.Status)
	assert.Equal(t, "1.25", quoted.Result["price"])
}

func TestSafetyGateBlocksBeforeApprovalPaymentAndExecution(t *testing.T) {
	t.Parallel()

	peerDID := "did:key:onChainEscrowToolsRunLifecycleAndQueryViews0-safety"
	h, _, token := settersAndBasicDispatchBranchesHandler(t, peerDID)
	payGate := &settersAndBasicDispatchBranchesPayGate{result: PayGateResult{Status: payGateStatusVerified, Auth: "auth"}}
	approvalCalls := 0
	execCalls := 0

	h.SetPayGate(payGate)
	h.SetApprovalFunc(func(context.Context, string, string, map[string]interface{}) (bool, error) {
		approvalCalls++
		return true, nil
	})
	h.SetSandboxExecutor(func(context.Context, string, map[string]interface{}) (map[string]interface{}, error) {
		execCalls++
		return map[string]interface{}{"ok": true}, nil
	})
	h.SetSafetyGate(func(string) (int, bool) { return 3, true }, 2, nil)

	resp := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestToolInvoke,
		SessionToken: token,
		RequestID:    "tool-blocked",
		Payload:      map[string]interface{}{"toolName": "danger"},
	})
	assert.Equal(t, ResponseStatusDenied, resp.Status)
	assert.Equal(t, ErrToolSafetyBlocked.Error(), resp.Error)

	paidResp := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestToolInvokePaid,
		SessionToken: token,
		RequestID:    "paid-blocked",
		Payload:      map[string]interface{}{"toolName": "danger"},
	})
	assert.Equal(t, ResponseStatusDenied, paidResp.Status)
	assert.Equal(t, ErrToolSafetyBlocked.Error(), paidResp.Error)
	assert.Zero(t, payGate.calls)
	assert.Zero(t, approvalCalls)
	assert.Zero(t, execCalls)
}

func TestPaidInvokePaymentBranchesAndEvents(t *testing.T) {
	t.Parallel()

	peerDID := "did:key:onChainEscrowToolsRunLifecycleAndQueryViews0-paid"
	h, _, token := settersAndBasicDispatchBranchesHandler(t, peerDID)
	h.SetApprovalFunc(func(context.Context, string, string, map[string]interface{}) (bool, error) { return true, nil })
	events := &settersAndBasicDispatchBranchesSecurityEvents{}
	h.SetSecurityEvents(events)

	quote := map[string]interface{}{"price": "2.00", "currency": "USDC"}
	h.SetPayGate(&settersAndBasicDispatchBranchesPayGate{result: PayGateResult{Status: payGateStatusPaymentRequired, PriceQuote: quote}})
	paymentRequired := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestToolInvokePaid,
		SessionToken: token,
		RequestID:    "paid-required",
		Payload:      map[string]interface{}{"toolName": "paid-tool"},
	})
	assert.Equal(t, ResponseStatusPaymentRequired, paymentRequired.Status)
	assert.Equal(t, "2.00", paymentRequired.Result["price"])

	h.SetPayGate(&settersAndBasicDispatchBranchesPayGate{result: PayGateResult{Status: payGateStatusInvalid}})
	invalid := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestToolInvokePaid,
		SessionToken: token,
		RequestID:    "paid-invalid",
		Payload:      map[string]interface{}{"toolName": "paid-tool"},
	})
	assert.Equal(t, ResponseStatusError, invalid.Status)
	assert.Equal(t, ErrInvalidPaymentAuth.Error(), invalid.Error)

	bus := eventbus.New()
	var paidEvents []eventbus.ToolExecutionPaidEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.ToolExecutionPaidEvent) {
		paidEvents = append(paidEvents, evt)
	})
	h.SetEventBus(bus)
	h.SetSandboxExecutor(func(context.Context, string, map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"ok": true, "secret": "remove-me"}, nil
	})
	h.firewall.SetZKAttestFunc(func(responseHash, agentDIDHash []byte) (*firewall.AttestationResult, error) {
		require.Len(t, responseHash, 32)
		require.Len(t, agentDIDHash, 32)
		return &firewall.AttestationResult{
			Proof:        []byte("proof"),
			PublicInputs: []byte("inputs"),
			CircuitID:    "onChainEscrowToolsRunLifecycleAndQueryViews0-circuit",
			Scheme:       "test",
		}, nil
	})

	h.SetPayGate(&settersAndBasicDispatchBranchesPayGate{result: PayGateResult{Status: payGateStatusVerified, Auth: "verified-auth"}})
	verified := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestToolInvokePaid,
		SessionToken: token,
		RequestID:    "paid-verified",
		Payload:      map[string]interface{}{"toolName": "paid-tool", "params": map[string]interface{}{"x": "y"}},
	})
	require.Equal(t, ResponseStatusOK, verified.Status)
	assert.Equal(t, true, verified.Result["ok"])
	assert.NotContains(t, verified.Result, "secret")
	require.NotNil(t, verified.Attestation)
	assert.Equal(t, []byte("proof"), verified.AttestationProof)
	require.Len(t, paidEvents, 1)
	assert.Equal(t, "verified-auth", paidEvents[0].Auth)
	assert.Equal(t, peerDID, paidEvents[0].PeerDID)
	assert.Equal(t, []string{peerDID}, events.successes)

	h.SetPayGate(&settersAndBasicDispatchBranchesPayGate{result: PayGateResult{Status: payGateStatusPostPayApproved, SettlementID: "settlement-1"}})
	postpay := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestToolInvokePaid,
		SessionToken: token,
		RequestID:    "paid-postpay",
		Payload:      map[string]interface{}{"toolName": "paid-tool"},
	})
	assert.Equal(t, ResponseStatusOK, postpay.Status)
	require.Len(t, paidEvents, 2)
	assert.Equal(t, "settlement-1", paidEvents[1].SettlementID)

	h.SetSandboxExecutor(func(context.Context, string, map[string]interface{}) (map[string]interface{}, error) {
		return nil, errors.New("tool failed")
	})
	failed := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestToolInvokePaid,
		SessionToken: token,
		RequestID:    "paid-failed",
		Payload:      map[string]interface{}{"toolName": "paid-tool"},
	})
	assert.Equal(t, ResponseStatusError, failed.Status)
	assert.Contains(t, failed.Error, "tool failed")
	assert.Equal(t, []string{peerDID}, events.failures)
}

func TestSchemaAndTeamBranches(t *testing.T) {
	t.Parallel()

	peerDID := "did:key:onChainEscrowToolsRunLifecycleAndQueryViews0-schema-team"
	h, _, token := settersAndBasicDispatchBranchesHandler(t, peerDID)

	noSchema := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestSchemaQuery,
		SessionToken: token,
		RequestID:    "schema-none",
		Payload:      map[string]interface{}{},
	})
	assert.Equal(t, ResponseStatusError, noSchema.Status)
	assert.Equal(t, "ontology handler not configured", noSchema.Error)

	ontology := &settersAndBasicDispatchBranchesOntologyHandler{
		queryResp: &SchemaQueryResponse{Bundle: json.RawMessage(`{"types":["task"]}`)},
		proposeResp: &SchemaProposeResponse{
			Action:   OntologyActionAccepted,
			Accepted: []string{"Task"},
			Result:   json.RawMessage(`{"imported":1}`),
		},
	}
	h.SetOntologyHandler(ontology)

	query := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestSchemaQuery,
		SessionToken: token,
		RequestID:    "schema-query",
		Payload:      map[string]interface{}{"requestedTypes": []string{"Task"}, "includePredicates": true},
	})
	assert.Equal(t, ResponseStatusOK, query.Status)
	assert.Equal(t, peerDID, ontology.peerSeen)
	assert.Equal(t, []string{"Task"}, ontology.querySeen.RequestedTypes)
	assert.Equal(t, map[string]interface{}{"types": []interface{}{"task"}}, query.Result["bundle"])

	propose := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestSchemaPropose,
		SessionToken: token,
		RequestID:    "schema-propose",
		Payload:      map[string]interface{}{"bundle": map[string]interface{}{"types": []interface{}{"Task"}}, "reason": "share"},
	})
	assert.Equal(t, ResponseStatusOK, propose.Status)
	assert.Equal(t, "share", ontology.propSeen.Reason)
	assert.Equal(t, OntologyActionAccepted, propose.Result["action"])

	h.SetOntologyHandler(&settersAndBasicDispatchBranchesOntologyHandler{err: errors.New("schema backend unavailable")})
	schemaErr := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestSchemaQuery,
		SessionToken: token,
		RequestID:    "schema-error",
		Payload:      map[string]interface{}{},
	})
	assert.Equal(t, ResponseStatusError, schemaErr.Status)
	assert.Equal(t, "schema backend unavailable", schemaErr.Error)

	noTeam := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestTeamInvite,
		SessionToken: token,
		RequestID:    "team-none",
		Payload:      map[string]interface{}{},
	})
	assert.Equal(t, ResponseStatusError, noTeam.Status)
	assert.Equal(t, "team handler not configured", noTeam.Error)

	h.SetTeamHandler(func(_ context.Context, gotPeer string, gotType RequestType, payload map[string]interface{}) (map[string]interface{}, error) {
		assert.Equal(t, peerDID, gotPeer)
		assert.Equal(t, RequestTeamInvite, gotType)
		assert.Equal(t, "onChainEscrowToolsRunLifecycleAndQueryViews0-team", payload["teamId"])
		return map[string]interface{}{"accepted": true}, nil
	})
	teamOK := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestTeamInvite,
		SessionToken: token,
		RequestID:    "team-ok",
		Payload:      map[string]interface{}{"teamId": "onChainEscrowToolsRunLifecycleAndQueryViews0-team"},
	})
	assert.Equal(t, ResponseStatusOK, teamOK.Status)
	assert.Equal(t, true, teamOK.Result["accepted"])

	h.SetTeamHandler(func(context.Context, string, RequestType, map[string]interface{}) (map[string]interface{}, error) {
		return nil, errors.New("team rejected")
	})
	teamErr := h.handleRequest(context.Background(), nil, &Request{
		Type:         RequestTeamTask,
		SessionToken: token,
		RequestID:    "team-error",
		Payload:      map[string]interface{}{},
	})
	assert.Equal(t, ResponseStatusError, teamErr.Status)
	assert.Equal(t, "team rejected", teamErr.Error)
}

func TestStreamHandlerDecodeErrorAndSendRequest(t *testing.T) {
	t.Parallel()

	h, _, _ := settersAndBasicDispatchBranchesHandler(t, "did:key:onChainEscrowToolsRunLifecycleAndQueryViews0-stream")
	var invalidIO bytes.Buffer
	invalidIO.WriteString("{not-json")
	stream := &settersAndBasicDispatchBranchesProtocolStream{rw: &invalidIO}

	h.StreamHandler()(stream)

	var resp Response
	require.NoError(t, json.NewDecoder(&invalidIO).Decode(&resp))
	assert.True(t, stream.closed)
	assert.Equal(t, ResponseStatusError, resp.Status)
	assert.Contains(t, resp.Error, "decode request")

	var responseBytes bytes.Buffer
	expected := Response{RequestID: "server-response", Status: ResponseStatusOK, Result: map[string]interface{}{"pong": true}}
	require.NoError(t, json.NewEncoder(&responseBytes).Encode(expected))
	var sentBytes bytes.Buffer
	sendResp, err := SendRequest(context.Background(), &settersAndBasicDispatchBranchesProtocolStream{
		rw: &settersAndBasicDispatchBranchesSplitReadWriter{reader: &responseBytes, writer: &sentBytes},
	}, RequestCapabilityQuery, "token", map[string]interface{}{"ping": true})
	require.NoError(t, err)
	assert.Equal(t, ResponseStatusOK, sendResp.Status)
	assert.Equal(t, true, sendResp.Result["pong"])

	var sent Request
	require.NoError(t, json.NewDecoder(&sentBytes).Decode(&sent))
	assert.Equal(t, RequestCapabilityQuery, sent.Type)
	assert.Equal(t, "token", sent.SessionToken)
	assert.Equal(t, true, sent.Payload["ping"])
}
