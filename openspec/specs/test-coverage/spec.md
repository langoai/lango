# Spec: Test Coverage

## Overview
Comprehensive test suite covering core business logic, security tools, and existing package enhancements for the Lango project.

## Requirements

### REQ-1: Knowledge Package Tests
The `internal/knowledge/` package must have test coverage for all Store CRUD operations and the ContextRetriever.

**Scenarios:**
- Save and retrieve knowledge entries (create, upsert, get, search, delete)
- Rate limiting per session for knowledge, learning, and skill creation
- Context retrieval across multiple layers (knowledge, skills, learnings, external refs)
- Keyword extraction from queries with stop word filtering
- Prompt assembly from retrieval results

### REQ-2: Learning Package Tests
The `internal/learning/` package must have test coverage for error pattern analysis and the learning engine.

**Scenarios:**
- Extract error patterns (UUID removal, timestamp removal, path/port normalization)
- Categorize errors by type (timeout, permission, provider_error, tool_error, general)
- Detect `context.DeadlineExceeded` through wrapped errors
- Summarize params (truncation, slice counting, type preservation)
- Engine records audit logs and creates learnings from tool errors
- Engine boosts confidence on successful tool executions
- Engine returns known fixes for high-confidence learnings
- Engine records user corrections as learnings

### REQ-3: Skill Package Tests
The `internal/skill/` package must have test coverage for building, executing, and managing skills.

**Scenarios:**
- Build composite, script, and template skills with correct type and definition
- Validate scripts against dangerous patterns (rm -rf, fork bombs, curl|bash, etc.)
- Execute composite skills (plan generation), template skills (rendering), script skills (shell execution)
- Registry validates skill creation (name, type, definition, script safety)
- Registry loads active skills and exposes them as agent tools with "skill_" prefix

### REQ-4: Security Tool Tests
The `internal/tools/crypto/` and `internal/tools/secrets/` packages must have test coverage.

**Scenarios:**
- Hash computation with sha256/sha512 (known values), default algorithm, unsupported algorithm
- Encrypt/decrypt round-trip via mock provider
- Sign with explicit and default key IDs
- Key listing via real KeyRegistry
- Secret store/get/list/delete lifecycle with real LocalCryptoProvider
- Secret update (upsert) returns latest value

### REQ-5: Existing Test Enhancements
Existing test files must be expanded with additional scenarios.

**Scenarios:**
- Session store: individual CRUD tests, TTL expiration, max history turns trimming, salt/checksum management
- Anthropic provider: ListModels returns expected model list
- OpenAI provider: ListModels returns error for unavailable server
- App: creation fails gracefully with no providers or invalid provider type

### REQ-6: Channel Mock Thread Safety
Channel test mock types SHALL use mutex synchronization to protect shared slices from concurrent access by handler goroutines and test assertions.

**Scenarios:**
- Slack mock concurrent access: serialized via mutex when handler goroutine appends to PostMessages/UpdateMessages while test goroutine reads
- Telegram mock concurrent access: serialized via mutex when handler goroutine appends to SentMessages/RequestCalls while test goroutine reads
- Safe mock data retrieval: helper methods return defensive copies of underlying slices
