## Context

The previous refactoring introduced typed enums (e.g., `types.MessageRole`, `types.Confidence`, `types.ChannelType`, `knowledge.KnowledgeCategory`) but only changed constant definitions. Receiving fields and function parameters remained `string`, forcing `string(types.XXX)` casts at every usage site. This undermines the type safety that enums were designed to provide.

All affected enum types are `string`-based (`type X string`), which means JSON marshal/unmarshal works transparently. Ent ORM stores these as plain strings in SQLite, so no DB migration is needed.

## Goals / Non-Goals

**Goals:**
- Eliminate `string(types.XXX)` casts from internal code by changing receiving fields/params to typed enums.
- Confine `string()` casts to system boundaries: DB reads/writes (Ent), external API (genai), CLI/TUI (UI frameworks requiring `[]string`), and JSON-parsed LLM responses.
- Maintain full backward compatibility — no behavioral changes, no DB migration, no JSON format changes.

**Non-Goals:**
- Changing CLI/TUI code where UI frameworks require `string` (these are valid boundary casts).
- Adding `String()` methods to string-based enums (redundant — they're already strings).
- Changing `LearningEntry.Category` or `ContextItem.Category` to typed enums (these have broader, cross-domain usage).

## Decisions

**1. Phased approach by type scope**
Change types from most impactful (Message.Role used everywhere) to least. This allows incremental build verification after each phase.
- Phase 1: `Message.Role` → `types.MessageRole` (highest impact, touches ADK/session/app/learning/memory)
- Phase 2: `parseDeliveryTarget()` → `types.ChannelType` (localized to sender.go)
- Phase 3: Confidence fields → `types.Confidence` (learning/librarian/config)
- Phase 4: `KnowledgeEntry.Category` → `knowledge.KnowledgeCategory` (knowledge/learning/librarian/app)
- Phase 5: CLI — no changes (boundary casts are correct)

**2. DB boundary: cast at Ent interface**
Ent `SetRole(string)` and `entknowledge.Category(string)` require string input. We add `string()` casts at these two write points and `types.MessageRole(m.Role)` / `KnowledgeCategory(k.Category)` at read points. This keeps casts confined to the store layer.

**3. Config/CLI boundary: cast at edges**
`mapstructure` handles `type X string` aliases correctly, so config YAML deserialization works without changes. CLI `tuicore.Field.Value` is `string`, so we cast at the UI boundary (`string(cfg.AutoSaveConfidence)` and `types.Confidence(val)`).

## Risks / Trade-offs

- **Risk**: Ent-generated code expects `string` for Role/Category fields → **Mitigation**: Add explicit `string()` casts only at the 2 DB write points per type. Verified with `go build`.
- **Risk**: `mapstructure` might not decode YAML strings into `types.Confidence` → **Mitigation**: Tested — `mapstructure` handles `type X string` aliases natively.
- **Trade-off**: Phase 4 adds `string()` casts at knowledge store boundary (3 metadata map sites) to remove casts at 5+ internal usage sites. Net reduction in casts.
- **Trade-off**: `ContextItem.Category` remains `string` because it serves as a cross-domain container (knowledge categories, learning categories, etc). Typing it would require a union type or separate fields.
