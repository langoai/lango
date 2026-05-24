## Summary

Propagate Cobra command cancellation into ephemeral P2P node startup used by `lango p2p` inspection and management commands.

## Problem

Several P2P CLI commands create an ephemeral libp2p node through `initP2PDeps`, but that helper starts the node without a caller context. `Node.Start` creates its own `context.Background()` root, so parent cancellation from wrappers, Ctrl+C handling, or command deadlines does not reach DHT bootstrap, bootstrap peer dial goroutines, or mDNS discovered-peer connection attempts.

The previous connect timeout work bounded the final host connect attempt, but shared node startup still remains outside the command lifecycle.

## Goals

- Make `internal/p2p.Node.Start` derive its internal lifecycle context from a caller-provided context.
- Thread `cmd.Context()` into shared P2P CLI dependency initialization.
- Cover representative inspection, discovery, session, connect, and disconnect paths so regressions cannot reintroduce command-detached startup.
- Document that ephemeral P2P CLI commands honor command cancellation during node startup.

## Non-Goals

- Change the public `lango p2p` command surface.
- Add new P2P configuration keys.
- Redesign long-running server-owned P2P lifecycle management.
