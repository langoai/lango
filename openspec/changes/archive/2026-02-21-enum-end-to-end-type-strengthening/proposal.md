## Why

Typed enums (`ChannelType`, `ProviderType`, `Confidence`, `MessageRole`, `KnowledgeCategory`) were introduced but receiving fields/parameters remain `string`, forcing `string(types.XXX)` casts throughout internal code. The `string()` cast should only occur at system boundaries (JSON parsing, DB reads, external input), not in internal logic.

## What Changes

- Change `session.Message.Role` from `string` to `types.MessageRole` end-to-end, with `string()` casts only at DB (Ent) and ADK (genai) boundaries.
- Change `parseDeliveryTarget()` return type from `string` to `types.ChannelType`, removing 6 `string(types.ChannelXXX)` casts in sender switch statements.
- Change `analysisResult.Confidence`, `ObservationKnowledge.Confidence`, `answerMatch.Confidence`, `ProactiveBuffer.autoSaveConfidence`, and `LibrarianConfig.AutoSaveConfidence` from `string` to `types.Confidence`.
- Change `KnowledgeEntry.Category` from `string` to `knowledge.KnowledgeCategory`, updating `mapKnowledgeCategory()` and `mapCategory()` return types.
- CLI boundary casts (wizard.go, forms_impl.go) are intentionally preserved as valid system boundary casts.

## Capabilities

### New Capabilities

### Modified Capabilities
- `session-store`: `Message.Role` field type changes from `string` to `types.MessageRole`.
- `knowledge-store`: `KnowledgeEntry.Category` field type changes from `string` to `knowledge.KnowledgeCategory`.

## Impact

- **session package**: `store.go` (Message struct), `ent_store.go` (DB boundary casts)
- **adk package**: `session_service.go` (role assignment), `state.go` (event mapping switch)
- **app package**: `sender.go` (parseDeliveryTarget, switch statements), `tools.go` (save_knowledge handler)
- **learning package**: `parse.go` (analysisResult, mapKnowledgeCategory), `session_learner.go` (confidence comparison)
- **librarian package**: `proactive_buffer.go` (shouldAutoSave, mapCategory), `inquiry_processor.go` (confidence/category), `types.go`, `parse.go`
- **knowledge package**: `types.go` (KnowledgeEntry), `store.go` (DB boundary casts), `retriever.go` (ContextItem boundary)
- **config package**: `types.go` (LibrarianConfig.AutoSaveConfidence), `loader.go` (default value)
- **cli package**: `state_update.go` (UI boundary cast), `forms_impl.go` (UI boundary cast)
