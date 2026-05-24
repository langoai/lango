package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"testing"
	"time"
	"unsafe"

	p2pnet "github.com/langoai/lango/internal/p2p"
	"github.com/langoai/lango/internal/p2p/discovery"
	"github.com/langoai/lango/internal/p2p/handshake"
	"github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/p2p/reputation"
	"github.com/langoai/lango/internal/testutil"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	libp2pprotocol "github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestP2PStatusReportsHostIdentitySessionsAndLocalDID(t *testing.T) {
	t.Parallel()

	localDID := p2PToolsMetadataAndMissingDependencyBranchesP2PPeerDID(t)
	connectedDID := p2PToolsMetadataAndMissingDependencyBranchesP2PPeerDID(t)
	addr := p2PToolsDeterministicRemoteBranchesMultiaddr(t, "/ip4/127.0.0.1/tcp/4101")
	sessions := p2PToolsDeterministicRemoteBranchesSessions(t, connectedDID.ID)
	fakeHost := &p2PToolsDeterministicRemoteBranchesHost{
		id:    localDID.PeerID,
		addrs: []ma.Multiaddr{addr},
		conns: []network.Conn{p2PToolsDeterministicRemoteBranchesConn{remote: connectedDID.PeerID}},
	}

	tool := findP2PTool(t, buildP2PTools(&p2pComponents{
		node:     p2PToolsDeterministicRemoteBranchesNode(t, fakeHost),
		sessions: sessions,
		identity: p2PToolsDeterministicRemoteBranchesDIDProvider{did: localDID},
	}), "p2p_status")

	got, err := tool.Handler(context.Background(), nil)
	require.NoError(t, err)
	payload := p2PToolsMetadataAndMissingDependencyBranchesP2PPayload(t, got)
	assert.Equal(t, localDID.PeerID.String(), payload["peerID"])
	assert.Equal(t, localDID.ID, payload["did"])
	assert.Equal(t, []string{addr.String()}, payload["listenAddrs"])
	assert.Equal(t, []string{connectedDID.PeerID.String()}, payload["connectedPeers"])
	assert.Equal(t, 1, payload["peerCount"])
	assert.Equal(t, 1, payload["sessions"])
}

func TestP2PConnectReportsHandshakeStreamOpenFailureWithoutDialingNetwork(t *testing.T) {
	t.Parallel()

	remoteDID := p2PToolsMetadataAndMissingDependencyBranchesP2PPeerDID(t)
	fakeHost := &p2PToolsDeterministicRemoteBranchesHost{
		id:           p2PToolsMetadataAndMissingDependencyBranchesP2PPeerDID(t).PeerID,
		newStreamErr: errors.New("stream disabled"),
	}
	tool := findP2PTool(t, buildP2PTools(&p2pComponents{
		node:       p2PToolsDeterministicRemoteBranchesNode(t, fakeHost),
		sessions:   p2PToolsDeterministicRemoteBranchesSessions(t),
		kemEnabled: true,
	}), "p2p_connect")

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"multiaddr": "/ip4/127.0.0.1/tcp/4102/p2p/" + remoteDID.PeerID.String(),
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "open handshake stream: stream disabled")
	require.Len(t, fakeHost.connects, 1)
	assert.Equal(t, remoteDID.PeerID, fakeHost.connects[0].ID)
	assert.Equal(t, remoteDID.PeerID, fakeHost.newStreamPeer)
	assert.Equal(t, p2PToolsDeterministicRemoteBranchesProtocolIDs(handshake.PreferredProtocols(true)), fakeHost.newStreamProtocols)
}

func TestP2PQueryUsesSessionTokenAndDefaultsEmptyParams(t *testing.T) {
	t.Parallel()

	remoteDID := p2PToolsMetadataAndMissingDependencyBranchesP2PPeerDID(t)
	sessions := p2PToolsDeterministicRemoteBranchesSessions(t, remoteDID.ID)
	session := sessions.Get(remoteDID.ID)
	stream := p2PToolsDeterministicRemoteBranchesStreamWith(protocol.Response{
		Status: protocol.ResponseStatusOK,
		Result: map[string]interface{}{"answer": "indexed", "count": 2},
	})
	fakeHost := &p2PToolsDeterministicRemoteBranchesHost{
		id:      p2PToolsMetadataAndMissingDependencyBranchesP2PPeerDID(t).PeerID,
		streams: []*p2PToolsDeterministicRemoteBranchesStream{stream},
	}

	tool := findP2PTool(t, buildP2PTools(&p2pComponents{
		node:     p2PToolsDeterministicRemoteBranchesNode(t, fakeHost),
		sessions: sessions,
	}), "p2p_query")

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"peer_did":  remoteDID.ID,
		"tool_name": "search_knowledge",
	})

	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"answer": "indexed", "count": float64(2)}, got)
	assert.True(t, stream.closed)
	req := stream.request(t)
	assert.Equal(t, protocol.RequestToolInvoke, req.Type)
	assert.Equal(t, session.Token, req.SessionToken)
	assert.Equal(t, "search_knowledge", req.Payload["toolName"])
	assert.Equal(t, map[string]interface{}{}, req.Payload["params"])
}

func TestP2PQueryWrapsRemoteToolError(t *testing.T) {
	t.Parallel()

	remoteDID := p2PToolsMetadataAndMissingDependencyBranchesP2PPeerDID(t)
	stream := p2PToolsDeterministicRemoteBranchesStreamWith(protocol.Response{
		Status: protocol.ResponseStatusError,
		Error:  "remote refused",
	})
	tool := findP2PTool(t, buildP2PTools(&p2pComponents{
		node: p2PToolsDeterministicRemoteBranchesNode(t, &p2PToolsDeterministicRemoteBranchesHost{
			id:      p2PToolsMetadataAndMissingDependencyBranchesP2PPeerDID(t).PeerID,
			streams: []*p2PToolsDeterministicRemoteBranchesStream{stream},
		}),
		sessions: p2PToolsDeterministicRemoteBranchesSessions(t, remoteDID.ID),
	}), "p2p_query")

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"peer_did":  remoteDID.ID,
		"tool_name": "summarize",
		"params":    `{"text":"hello"}`,
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "remote tool invoke")
	assert.ErrorContains(t, err, "remote tool summarize error: remote refused")
	assert.True(t, stream.closed)
}

func TestP2PPriceQueryReturnsQuoteFromRemoteAgent(t *testing.T) {
	t.Parallel()

	remoteDID := p2PToolsMetadataAndMissingDependencyBranchesP2PPeerDID(t)
	sessions := p2PToolsDeterministicRemoteBranchesSessions(t, remoteDID.ID)
	session := sessions.Get(remoteDID.ID)
	stream := p2PToolsDeterministicRemoteBranchesStreamWith(protocol.Response{
		Status: protocol.ResponseStatusOK,
		Result: map[string]interface{}{
			"toolName":     "rank_results",
			"price":        "0.25",
			"currency":     "USDC",
			"usdcContract": "0x1111111111111111111111111111111111111111",
			"chainId":      84532,
			"sellerAddr":   "0x2222222222222222222222222222222222222222",
			"quoteExpiry":  int64(1_779_321_600),
			"isFree":       false,
		},
	})
	tool := findP2PTool(t, buildP2PTools(&p2pComponents{
		node: p2PToolsDeterministicRemoteBranchesNode(t, &p2PToolsDeterministicRemoteBranchesHost{
			id:      p2PToolsMetadataAndMissingDependencyBranchesP2PPeerDID(t).PeerID,
			streams: []*p2PToolsDeterministicRemoteBranchesStream{stream},
		}),
		sessions: sessions,
	}), "p2p_price_query")

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"peer_did":  remoteDID.ID,
		"tool_name": "rank_results",
	})

	require.NoError(t, err)
	payload := p2PToolsMetadataAndMissingDependencyBranchesP2PPayload(t, got)
	assert.Equal(t, "rank_results", payload["toolName"])
	assert.Equal(t, "0.25", payload["price"])
	assert.Equal(t, "USDC", payload["currency"])
	assert.Equal(t, int64(84532), payload["chainId"])
	assert.Equal(t, false, payload["isFree"])
	req := stream.request(t)
	assert.Equal(t, protocol.RequestPriceQuery, req.Type)
	assert.Equal(t, session.Token, req.SessionToken)
	assert.Equal(t, "rank_results", req.Payload["toolName"])
}

func TestP2PDiscoverUsesNonNilEmptyGossipService(t *testing.T) {
	t.Parallel()

	tool := findP2PTool(t, buildP2PTools(&p2pComponents{
		gossip: &discovery.GossipService{},
	}), "p2p_discover")

	all, err := tool.Handler(context.Background(), nil)
	require.NoError(t, err)
	allPayload := p2PToolsMetadataAndMissingDependencyBranchesP2PPayload(t, all)
	assert.Equal(t, 0, allPayload["count"])
	assert.NotContains(t, allPayload, "message")

	filtered, err := tool.Handler(context.Background(), map[string]interface{}{"capability": "search"})
	require.NoError(t, err)
	filteredPayload := p2PToolsMetadataAndMissingDependencyBranchesP2PPayload(t, filtered)
	assert.Equal(t, 0, filteredPayload["count"])
	assert.NotContains(t, filteredPayload, "message")
}

func TestP2PReputationWrapsStoreLookupError(t *testing.T) {
	t.Parallel()

	client := testutil.TestEntClient(t)
	store := reputation.NewStore(client, testLog())
	require.NoError(t, client.Close())
	tool := findP2PTool(t, buildP2PTools(&p2pComponents{reputation: store}), "p2p_reputation")

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"peer_did": "did:lango:p2PToolsDeterministicRemoteBranches9-closed-store",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "get reputation")
}

type p2PToolsDeterministicRemoteBranchesDIDProvider struct {
	did *identity.DID
	err error
}

func (p p2PToolsDeterministicRemoteBranchesDIDProvider) DID(context.Context) (*identity.DID, error) {
	return p.did, p.err
}

type p2PToolsDeterministicRemoteBranchesHost struct {
	host.Host
	id                 peer.ID
	addrs              []ma.Multiaddr
	conns              []network.Conn
	connectErr         error
	connects           []peer.AddrInfo
	streams            []*p2PToolsDeterministicRemoteBranchesStream
	newStreamErr       error
	newStreamPeer      peer.ID
	newStreamProtocols []libp2pprotocol.ID
}

func (h *p2PToolsDeterministicRemoteBranchesHost) ID() peer.ID {
	return h.id
}

func (h *p2PToolsDeterministicRemoteBranchesHost) Addrs() []ma.Multiaddr {
	return append([]ma.Multiaddr(nil), h.addrs...)
}

func (h *p2PToolsDeterministicRemoteBranchesHost) Network() network.Network {
	return p2PToolsDeterministicRemoteBranchesNetwork{conns: h.conns}
}

func (h *p2PToolsDeterministicRemoteBranchesHost) Connect(ctx context.Context, info peer.AddrInfo) error {
	h.connects = append(h.connects, info)
	if err := ctx.Err(); err != nil {
		return err
	}
	return h.connectErr
}

func (h *p2PToolsDeterministicRemoteBranchesHost) NewStream(ctx context.Context, p peer.ID, protocols ...libp2pprotocol.ID) (network.Stream, error) {
	h.newStreamPeer = p
	h.newStreamProtocols = append([]libp2pprotocol.ID(nil), protocols...)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if h.newStreamErr != nil {
		return nil, h.newStreamErr
	}
	if len(h.streams) == 0 {
		return nil, errors.New("unexpected stream open")
	}
	stream := h.streams[0]
	h.streams = h.streams[1:]
	return stream, nil
}

type p2PToolsDeterministicRemoteBranchesNetwork struct {
	network.Network
	conns []network.Conn
}

func (n p2PToolsDeterministicRemoteBranchesNetwork) Conns() []network.Conn {
	return append([]network.Conn(nil), n.conns...)
}

type p2PToolsDeterministicRemoteBranchesConn struct {
	network.Conn
	remote peer.ID
}

func (c p2PToolsDeterministicRemoteBranchesConn) RemotePeer() peer.ID {
	return c.remote
}

type p2PToolsDeterministicRemoteBranchesStream struct {
	network.Stream
	reader   io.Reader
	sent     bytes.Buffer
	response protocol.Response
	closed   bool
}

func p2PToolsDeterministicRemoteBranchesStreamWith(response protocol.Response) *p2PToolsDeterministicRemoteBranchesStream {
	return &p2PToolsDeterministicRemoteBranchesStream{response: response}
}

func (s *p2PToolsDeterministicRemoteBranchesStream) Read(p []byte) (int, error) {
	if s.reader == nil {
		return 0, io.EOF
	}
	return s.reader.Read(p)
}

func (s *p2PToolsDeterministicRemoteBranchesStream) Write(p []byte) (int, error) {
	n, err := s.sent.Write(p)
	if err != nil {
		return n, err
	}
	var req protocol.Request
	if err := json.Unmarshal(p, &req); err == nil {
		resp := s.response
		resp.RequestID = req.RequestID
		var encoded bytes.Buffer
		if err := json.NewEncoder(&encoded).Encode(resp); err == nil {
			s.reader = &encoded
		}
	}
	return n, nil
}

func (s *p2PToolsDeterministicRemoteBranchesStream) Close() error {
	s.closed = true
	return nil
}

func (s *p2PToolsDeterministicRemoteBranchesStream) request(t *testing.T) protocol.Request {
	t.Helper()

	var req protocol.Request
	require.NoError(t, json.NewDecoder(bytes.NewReader(s.sent.Bytes())).Decode(&req))
	require.NotEmpty(t, req.RequestID)
	return req
}

func p2PToolsDeterministicRemoteBranchesNode(t *testing.T, fakeHost host.Host) *p2pnet.Node {
	t.Helper()

	node := &p2pnet.Node{}
	field := reflect.ValueOf(node).Elem().FieldByName("host")
	require.True(t, field.IsValid())
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(reflect.ValueOf(fakeHost))
	return node
}

func p2PToolsDeterministicRemoteBranchesSessions(t *testing.T, peerDIDs ...string) *handshake.SessionStore {
	t.Helper()

	sessions, err := handshake.NewSessionStore(time.Hour)
	require.NoError(t, err)
	for _, peerDID := range peerDIDs {
		_, err := sessions.Create(peerDID, true)
		require.NoError(t, err)
	}
	return sessions
}

func p2PToolsDeterministicRemoteBranchesMultiaddr(t *testing.T, value string) ma.Multiaddr {
	t.Helper()

	addr, err := ma.NewMultiaddr(value)
	require.NoError(t, err)
	return addr
}

func p2PToolsDeterministicRemoteBranchesProtocolIDs(values []string) []libp2pprotocol.ID {
	ids := make([]libp2pprotocol.ID, len(values))
	for i, value := range values {
		ids[i] = libp2pprotocol.ID(value)
	}
	return ids
}
