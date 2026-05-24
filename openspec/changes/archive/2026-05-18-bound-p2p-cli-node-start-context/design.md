## Overview

The P2P CLI commands are thin wrappers around short-lived ephemeral P2P nodes. Their startup should be scoped to the Cobra command context just like the peer connect operation is scoped today.

## Design

### Core Node Startup

Change `internal/p2p.Node.Start` to accept a parent `context.Context`:

- If the parent is already canceled, startup should fail through existing DHT/bootstrap context handling rather than creating a detached background context.
- The node should still keep its own cancel function so `Node.Stop` can stop bootstrap peer goroutines and mDNS notifees.
- Bootstrap peer goroutines should keep using the node lifecycle context derived from the caller.

### CLI Dependency Initialization

Change `initP2PDeps` to accept a context and call `node.Start(ctx, &wg)`.

Every P2P CLI command path that uses `initP2PDeps` should pass `cmd.Context()` through its package-level loader seam. This includes status, peers, discover, identity, session list/revoke/revoke-all, connect, and disconnect.

### Testability

Keep package-level loader seams so command tests can verify context propagation without starting a real libp2p node. Add a lower-level startup seam for `initP2PDeps` so tests can assert it passes the parent context to `Node.Start`.

### Documentation

Update public P2P CLI docs to explain that ephemeral-node commands honor command cancellation while starting their temporary P2P node.
