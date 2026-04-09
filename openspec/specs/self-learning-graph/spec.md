## Purpose

Capability spec for self-learning-graph. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: ToolResultObserver interface
The learning package SHALL define a `ToolResultObserver` interface with `OnToolResult(ctx, sessionKey, toolName, params, result, err)` method, implemented by both Engine and GraphEngine.

#### Scenario: Engine satisfies ToolResultObserver
- **WHEN** compile-time check `var _ ToolResultObserver = (*Engine)(nil)` is present
- **THEN** the code SHALL compile without errors

#### Scenario: GraphEngine satisfies ToolResultObserver
- **WHEN** compile-time check `var _ ToolResultObserver = (*GraphEngine)(nil)` is present
- **THEN** the code SHALL compile without errors

### Requirement: GraphEngine error-resolution graph
The GraphEngine SHALL create graph triples for error-resolution relationships: `CausedBy` (error → tool), `InSession` (error → session), `SimilarTo` (error → similar errors), `ResolvedBy` (error → fix), `LearnedFrom` (fix → session).

#### Scenario: Error produces CausedBy and InSession triples
- **WHEN** a tool execution produces an error
- **THEN** the GraphEngine SHALL create at least `CausedBy` and `InSession` triples

#### Scenario: Fix recorded with ResolvedBy and LearnedFrom
- **WHEN** RecordFix is called with an error pattern and fix
- **THEN** `ResolvedBy` and `LearnedFrom` triples SHALL be created

### Requirement: Confidence propagation
The GraphEngine SHALL propagate confidence boosts across similar learnings when a tool succeeds, using a configurable propagation rate (default 0.3).

#### Scenario: Success boosts similar learnings
- **WHEN** a tool succeeds and has previously failed with graph-linked similar errors
- **THEN** confidence of similar error learnings SHALL be boosted by the propagation rate

### Requirement: GraphEngine callback or direct store
The GraphEngine SHALL support two modes for triple persistence: async callback (preferred) or direct graph store writes (fallback).

#### Scenario: Callback mode
- **WHEN** a GraphCallback is set via SetGraphCallback
- **THEN** triples SHALL be sent to the callback instead of written directly

#### Scenario: Direct store mode
- **WHEN** no callback is set but graphStore is available
- **THEN** triples SHALL be written directly to the graph store

### Requirement: GraphEngine wiring in knowledge init
When graph store is available during knowledge initialization, a GraphEngine SHALL be created and used as the ToolResultObserver instead of the base Engine.

#### Scenario: Graph enabled → GraphEngine as observer
- **WHEN** `graph.enabled: true` and knowledge system is enabled
- **THEN** initKnowledge SHALL create a GraphEngine with the graph store and set it as the observer

#### Scenario: Graph disabled → Engine as observer
- **WHEN** `graph.enabled: false`
- **THEN** initKnowledge SHALL use the base Engine as the observer
