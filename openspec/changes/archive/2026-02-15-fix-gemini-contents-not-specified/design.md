## Context

Two bugs were found in ADK v0.4.0's message processing flow:

1. `SessionServiceAdapter.AppendEvent` saves messages to DB via `store.AppendMessage()` but does not update `SessionAdapter.sess.History` (in-memory). Since the ADK runner reads from the same session object's in-memory history, the user message just added is not visible, causing empty contents to be sent to the Gemini API.

2. `ModelAdapter.GenerateContent` ignores `req.Config.SystemInstruction` and only passes `req.Contents` to the provider. The system prompt set by ADK does not reach the LLM.

## Goals / Non-Goals

**Goals:**
- After AppendEvent call, in-memory history is immediately updated so the ADK runner can read the current turn's messages
- ADK SystemInstruction is passed to the provider as a system message
- Both fixes verified with tests

**Non-Goals:**
- Modifying the ADK library itself
- Changing the session store interface
- Changing the Provider interface

## Decisions

### Decision 1: Directly update in-memory history in AppendEvent

After successful DB save, append the message to `SessionAdapter.sess.History`. Access via type assertion of the `sess` parameter to `*SessionAdapter`.

**Alternative considered:** Re-read from DB → unnecessary I/O, performance degradation. Direct append is simpler and more efficient.

### Decision 2: Convert SystemInstruction to system role message

Combine text parts from `genai.Content` into a single `provider.Message{Role: "system"}` and prepend to the messages array.

**Alternative considered:** Add SystemInstruction field to GenerateParams → requires provider interface changes, all provider implementations need modification. Inserting as a system role in the existing Messages array minimizes the change scope.

## Risks / Trade-offs

- **[In-memory/DB inconsistency]** → If AppendEvent fails, the message may be in DB but not in memory. However, since the error is returned for the caller to handle, and in-memory is only updated after successful DB save, consistency is maintained.
- **[System message duplication]** → Some providers may already handle system messages separately. Checking current provider implementations confirmed they properly process system role messages from the Messages array.
