# Package Dependency Graph

Generated: 2026-04-02

This diagram shows the top-level internal package dependencies in Lango.
Only cross-domain dependencies are shown; intra-domain and stdlib imports are omitted.
Edges are aggregated: if any sub-package of domain A imports any sub-package of domain B,
a single edge `A --> B` is drawn.

Source: `go list -f '{{.ImportPath}}|{{join .Imports ","}}' ./internal/...`

## Enforced Boundaries

After Phase 0 boundary fixes:
- `economy` no longer imports `p2p` (was: `escrow/address_resolver` -> `p2p/identity`)
- `p2p/paygate` no longer imports `wallet` (uses `finance` instead)
- Shared types (`DIDPrefix`, `ReputationQuerier`) moved to `types`
- Shared monetary utilities (`ParseUSDC`, `FormatUSDC`) moved to `finance`

## Dependency Diagram

```mermaid
flowchart TB
    %% ─────────────────────────────────────────────
    %% Domain grouping
    %% ─────────────────────────────────────────────

    subgraph Foundation ["Foundation Layer (zero or near-zero internal deps)"]
        types
        ctxkeys
        logging
        llm
        finance
        eventbus
        toolparam
        storeutil
        mdparse
        asyncbuf
        deadline
        keyring
        dbmigrate
        search
    end

    subgraph DataLayer ["Data / Persistence"]
        ent
        session
        config
        configstore
        security
    end

    subgraph AgentCore ["Agent Core"]
        agent
        provider
        prompt
        approval
        tooloutput
        toolcatalog
        agentmemory
    end

    subgraph Intelligence ["Intelligence & Knowledge"]
        knowledge
        embedding
        graph
        memory
        retrieval
        learning
        ontology
        librarian
    end

    subgraph Orchestration ["Orchestration & Runtime"]
        adk
        orchestration
        turnrunner
        turntrace
        agentrt
        agentregistry
        runledger
        workflow
        toolchain
        supervisor
        lifecycle
        background
        automation
        cron
        observability
        alerting
        skill
        mcp
        gatekeeper
    end

    subgraph Economy ["Economy & Payments"]
        economy
        wallet
        payment
        contract
        smartaccount
        x402
    end

    subgraph Network ["P2P & Gateway"]
        p2p
        gateway
        channels
    end

    subgraph AppWiring ["Application Wiring"]
        app
        appinit
        bootstrap
        sandbox
    end

    subgraph UI ["CLI / TUI"]
        cli
        testutil
    end

    %% ─────────────────────────────────────────────
    %% Key cross-domain edges (curated for readability)
    %%
    %% Only the MOST IMPORTANT edges are shown.
    %% Ubiquitous deps (types, logging, config, ent,
    %% ctxkeys, toolparam, eventbus) are NOT drawn
    %% individually to avoid a hairball.
    %% ─────────────────────────────────────────────

    %% --- Agent Core deps ---
    agent --> logging
    provider --> logging
    config --> provider
    config --> types

    %% --- Intelligence layer ---
    embedding --> agent
    embedding --> knowledge
    embedding --> memory
    embedding --> asyncbuf
    graph --> agent
    graph --> llm
    graph --> asyncbuf
    memory --> agent
    memory --> graph
    memory --> llm
    memory --> asyncbuf
    knowledge --> eventbus
    knowledge --> search
    retrieval --> embedding
    retrieval --> knowledge
    learning --> graph
    learning --> knowledge
    learning --> llm
    ontology --> agent
    ontology --> graph
    librarian --> agent
    librarian --> knowledge
    librarian --> memory

    %% --- Orchestration layer ---
    adk --> agent
    adk --> embedding
    adk --> graph
    adk --> knowledge
    adk --> memory
    adk --> prompt
    adk --> provider
    adk --> retrieval
    adk --> runledger
    adk --> approval
    turnrunner --> adk
    turnrunner --> approval
    turnrunner --> turntrace
    turnrunner --> deadline
    agentrt --> adk
    agentrt --> turnrunner
    agentregistry --> orchestration
    agentregistry --> mdparse
    orchestration --> agent
    orchestration --> p2p
    toolchain --> agent
    toolchain --> approval
    toolchain --> learning
    toolchain --> wallet
    toolchain --> tooloutput
    runledger --> agent
    runledger --> background
    runledger --> workflow
    runledger --> toolchain
    workflow --> agent
    workflow --> automation
    observability --> toolchain
    alerting --> agentrt
    supervisor --> provider
    cron --> agent
    cron --> approval
    cron --> automation
    background --> agent
    background --> approval
    background --> automation
    background --> toolchain
    skill --> agent
    skill --> mdparse
    mcp --> agent

    %% --- Economy layer ---
    economy --> agent
    economy --> wallet
    wallet --> finance
    wallet --> security
    payment --> wallet
    contract --> payment
    contract --> wallet
    smartaccount --> contract
    smartaccount --> wallet
    x402 --> wallet
    x402 --> security

    %% --- P2P / Network ---
    p2p --> security
    p2p --> wallet
    p2p --> finance
    p2p --> economy
    p2p --> ontology
    p2p --> payment
    gateway --> adk
    gateway --> approval
    gateway --> gatekeeper
    gateway --> security
    gateway --> turnrunner
    gateway --> turntrace
    gateway --> runledger
    channels --> approval

    %% --- Application Wiring ---
    app --> adk
    app --> agent
    app --> economy
    app --> p2p
    app --> payment
    app --> smartaccount
    app --> gateway
    app --> wallet
    app --> observability
    app --> ontology
    app --> learning
    app --> workflow
    app --> toolchain
    bootstrap --> config
    bootstrap --> configstore
    bootstrap --> security
    appinit --> agent
    appinit --> lifecycle

    %% --- CLI ---
    cli --> app
    cli --> bootstrap
    cli --> config
    cli --> configstore
    cli --> wallet
    cli --> p2p
    cli --> security
    cli --> payment
    cli --> smartaccount
    cli --> provider
    cli --> observability
    cli --> runledger
    cli --> provenance
    cli --> toolchain
    cli --> turnrunner
    cli --> turntrace
    cli --> agentregistry
    cli --> background
    cli --> cron
    cli --> workflow
    cli --> sandbox
    cli --> mcp
    cli --> graph
    cli --> memory

    %% --- Data layer cross-deps ---
    session --> ent
    configstore --> config
    configstore --> ent
    configstore --> security
    security --> config
    security --> ent

    %% --- Provenance ---
    provenance --> observability
    provenance --> p2p
    provenance --> runledger
    provenance --> storeutil
```

## Reading the Diagram

| Subgraph | Description |
|---|---|
| **Foundation** | Zero or near-zero internal deps; leaf nodes that many packages import |
| **Data / Persistence** | Ent ORM, session store, config, security vault |
| **Agent Core** | Agent abstraction, LLM providers, prompt management, tool I/O |
| **Intelligence** | Knowledge graph, embeddings, RAG retrieval, learning, ontology |
| **Orchestration** | ADK integration, turn runner, run ledger, workflow engine, observability |
| **Economy** | Token economy, wallet, on-chain payments, smart accounts |
| **P2P & Gateway** | libp2p networking, HTTP gateway, chat channels (Slack/Discord/Telegram) |
| **Application Wiring** | `app` (God object), bootstrap, sandbox |
| **CLI / TUI** | Cobra commands, Bubbletea TUI, test utilities |

## Ubiquitous Dependencies (omitted from diagram for clarity)

The following packages are imported by 10+ other top-level packages and their edges are
**not drawn** to keep the diagram readable:

| Package | Role | Approx. importers |
|---|---|---|
| `types` | Shared value types | ~20 |
| `logging` | Structured logger | ~25 |
| `config` | Configuration | ~20 |
| `ent` | ORM / data access | ~15 |
| `ctxkeys` | Context key constants | ~10 |
| `toolparam` | Tool parameter types | ~15 |
| `eventbus` | Event pub/sub bus | ~12 |
| `session` | Session management | ~12 |

## Notes

- **`app` is a God object**: it imports nearly every other top-level package. Decomposition is a future goal.
- **`cli` fans out widely**: expected for a CLI layer, but it should only depend on `app` + a few infra packages.
- **`orchestration` --> `p2p`**: the `p2p/agentpool` dependency creates a coupling between orchestration and networking.
- **Foundation packages have no internal deps**: `types`, `finance`, `eventbus`, `toolparam`, `llm` are true leaf nodes.
