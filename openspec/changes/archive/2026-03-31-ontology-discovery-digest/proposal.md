# Proposal: Ontology Discovery Digest

## Problem

Peers in the P2P network currently have no way to know whether another agent has an ontology schema they might want to query or align with. This forces blind requests or manual coordination.

## Solution

Add an `OntologyDigest` optional field to `GossipCard` and `AgentAd` structs. This lightweight summary (schema version, hash digest, type/predicate counts, and optionally type names) lets peers discover ontology-capable agents and assess schema compatibility before initiating heavier exchange protocols.

## Scope

- Define `OntologyDigest` struct in `internal/p2p/discovery/`
- Add optional field to `GossipCard` and `AgentAd`
- Backward compatible: field is pointer + omitempty, old peers ignore it

## Non-Goals

- Ontology exchange protocol (future change)
- Schema alignment/merge logic
- Populating the digest from the ontology subsystem (wiring is a separate change)
