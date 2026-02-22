package p2p

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	ma "github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/config"
)

const nodeKeyFile = "node.key"

// Node wraps a libp2p host with DHT-based peer discovery.
type Node struct {
	host   host.Host
	dht    *dht.IpfsDHT
	cfg    config.P2PConfig
	logger *zap.SugaredLogger
	cancel context.CancelFunc

	mdnsSvc mdns.Service
}

// NewNode creates a libp2p node with Noise encryption and TCP/QUIC transports.
// The node key is persisted in cfg.KeyDir so the peer identity survives restarts.
func NewNode(cfg config.P2PConfig, logger *zap.SugaredLogger) (*Node, error) {
	privKey, err := loadOrGenerateKey(cfg.KeyDir)
	if err != nil {
		return nil, fmt.Errorf("load node key: %w", err)
	}

	lowWatermark := cfg.MaxPeers * 80 / 100
	cm, err := connmgr.NewConnManager(lowWatermark, cfg.MaxPeers)
	if err != nil {
		return nil, fmt.Errorf("new conn manager: %w", err)
	}

	opts := []libp2p.Option{
		libp2p.Identity(privKey),
		libp2p.ListenAddrStrings(cfg.ListenAddrs...),
		libp2p.ConnectionManager(cm),
	}

	if cfg.EnableRelay {
		opts = append(opts, libp2p.EnableRelayService())
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("new libp2p host: %w", err)
	}

	logger.Infow("libp2p node created",
		"peerID", h.ID(),
		"addrs", h.Addrs(),
	)

	return &Node{
		host:   h,
		cfg:    cfg,
		logger: logger,
	}, nil
}

// Start bootstraps the Kademlia DHT and optionally starts mDNS discovery.
// The WaitGroup is incremented so callers can wait for graceful shutdown.
func (n *Node) Start(wg *sync.WaitGroup) error {
	ctx, cancel := context.WithCancel(context.Background())
	n.cancel = cancel

	// Bootstrap the DHT.
	kadDHT, err := dht.New(ctx, n.host, dht.Mode(dht.ModeAutoServer))
	if err != nil {
		cancel()
		return fmt.Errorf("new DHT: %w", err)
	}
	n.dht = kadDHT

	if err := n.dht.Bootstrap(ctx); err != nil {
		cancel()
		return fmt.Errorf("DHT bootstrap: %w", err)
	}

	// Connect to bootstrap peers.
	for _, addr := range n.cfg.BootstrapPeers {
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			n.logger.Warnw("invalid bootstrap multiaddr", "addr", addr, "err", err)
			continue
		}
		pi, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			n.logger.Warnw("parse bootstrap peer info", "addr", addr, "err", err)
			continue
		}
		wg.Add(1)
		go func(pi peer.AddrInfo) {
			defer wg.Done()
			if err := n.host.Connect(ctx, pi); err != nil {
				n.logger.Warnw("connect bootstrap peer", "peer", pi.ID, "err", err)
			} else {
				n.logger.Infow("connected to bootstrap peer", "peer", pi.ID)
			}
		}(*pi)
	}

	// Optional mDNS discovery for LAN peers.
	if n.cfg.EnableMDNS {
		svc := mdns.NewMdnsService(n.host, "", &mdnsNotifee{
			host:   n.host,
			ctx:    ctx,
			logger: n.logger,
		})
		if err := svc.Start(); err != nil {
			n.logger.Warnw("start mDNS", "err", err)
		} else {
			n.mdnsSvc = svc
			n.logger.Info("mDNS discovery started")
		}
	}

	n.logger.Infow("P2P node started",
		"peerID", n.host.ID(),
		"listenAddrs", n.host.Addrs(),
	)

	return nil
}

// Stop shuts down the DHT, mDNS service, and libp2p host.
func (n *Node) Stop() error {
	if n.cancel != nil {
		n.cancel()
	}

	if n.mdnsSvc != nil {
		if err := n.mdnsSvc.Close(); err != nil {
			n.logger.Warnw("close mDNS", "err", err)
		}
	}

	if n.dht != nil {
		if err := n.dht.Close(); err != nil {
			return fmt.Errorf("close DHT: %w", err)
		}
	}

	if err := n.host.Close(); err != nil {
		return fmt.Errorf("close host: %w", err)
	}

	n.logger.Info("P2P node stopped")
	return nil
}

// PeerID returns the node's libp2p peer ID.
func (n *Node) PeerID() peer.ID { return n.host.ID() }

// Multiaddrs returns the listen addresses of the underlying host.
func (n *Node) Multiaddrs() []ma.Multiaddr { return n.host.Addrs() }

// ConnectedPeers returns the peer IDs of all currently connected peers.
func (n *Node) ConnectedPeers() []peer.ID {
	conns := n.host.Network().Conns()
	seen := make(map[peer.ID]struct{}, len(conns))
	peers := make([]peer.ID, 0, len(conns))
	for _, c := range conns {
		pid := c.RemotePeer()
		if _, ok := seen[pid]; !ok {
			seen[pid] = struct{}{}
			peers = append(peers, pid)
		}
	}
	return peers
}

// Host returns the underlying libp2p host for protocol registration.
func (n *Node) Host() host.Host { return n.host }

// SetStreamHandler registers a protocol stream handler on the host.
func (n *Node) SetStreamHandler(protocolID string, handler network.StreamHandler) {
	n.host.SetStreamHandler(protocol.ID(protocolID), handler)
}

// loadOrGenerateKey loads an Ed25519 node key from keyDir/node.key,
// generating a new one if it does not exist.
func loadOrGenerateKey(keyDir string) (crypto.PrivKey, error) {
	keyDir = expandHome(keyDir)
	if keyDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home dir: %w", err)
		}
		keyDir = filepath.Join(home, ".lango", "p2p")
	}

	if err := os.MkdirAll(keyDir, 0o700); err != nil {
		return nil, fmt.Errorf("create key dir %q: %w", keyDir, err)
	}

	keyPath := filepath.Join(keyDir, nodeKeyFile)
	data, err := os.ReadFile(keyPath)
	if err == nil {
		key, err := crypto.UnmarshalPrivateKey(data)
		if err != nil {
			return nil, fmt.Errorf("unmarshal node key: %w", err)
		}
		return key, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read node key: %w", err)
	}

	// Generate new key.
	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate ed25519 key: %w", err)
	}

	raw, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("marshal node key: %w", err)
	}

	if err := os.WriteFile(keyPath, raw, 0o600); err != nil {
		return nil, fmt.Errorf("write node key: %w", err)
	}

	return privKey, nil
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}

// mdnsNotifee handles mDNS peer discovery events.
type mdnsNotifee struct {
	host   host.Host
	ctx    context.Context
	logger *zap.SugaredLogger
}

func (n *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == n.host.ID() {
		return
	}
	n.logger.Infow("mDNS peer discovered", "peer", pi.ID)
	if err := n.host.Connect(n.ctx, pi); err != nil {
		n.logger.Warnw("connect mDNS peer", "peer", pi.ID, "err", err)
	}
}
