## 1. Tests

- [x] 1.1 Add failing chat rendering coverage for incomplete setup state.
- [x] 1.2 Add failing chat submission coverage proving incomplete setup does not call the turn runner and keeps the draft.
- [x] 1.3 Add slash-command and ready-provider regressions for pre-setup slash access and normal ready submission.

## 2. Implementation

- [x] 2.1 Add setup-readiness helpers to the focused chat model using `config.EvaluateAgentSetup`.
- [x] 2.2 Update header, turn strip, and help/footer affordances so incomplete profiles do not appear ready.
- [x] 2.3 Gate normal non-slash submission before user transcript/pending/runner side effects while preserving slash commands.

## 3. Documentation And Specs

- [x] 3.1 Update public core CLI docs and main OpenSpec specs for focused-chat readiness gating.
- [x] 3.2 Run focused tests, full Go build/tests, and strict OpenSpec validation.
- [x] 3.3 Run subagent-driven review and address required findings.
- [x] 3.4 Sync specs, archive the change, and commit the scoped unit.
