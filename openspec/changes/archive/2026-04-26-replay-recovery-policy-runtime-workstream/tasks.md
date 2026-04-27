## 1. Runtime architecture docs

- [x] 1.1 Update the automatic post-adjudication execution page with the shared execution-mode default
- [x] 1.2 Update the background post-adjudication execution page with manual/background/inline alignment
- [x] 1.3 Update the retry / dead-letter page with the normalized retry policy shape
- [x] 1.4 Update the operator replay / manual retry page with shared recovery-substrate semantics
- [x] 1.5 Update the policy-driven replay controls page with shared recovery-gate wording

## 2. Track alignment

- [x] 2.1 Update the P2P knowledge exchange track to mark replay/recovery runtime alignment as landed
- [x] 2.2 Narrow the remaining-work lists to actual follow-on gaps

## 3. OpenSpec sync

- [x] 3.1 Update `openspec/specs/docs-only/spec.md` for the landed runtime default behavior
- [x] 3.2 Update `openspec/specs/docs-only/spec.md` for the normalized retry/dead-letter policy shape
- [x] 3.3 Update `openspec/specs/docs-only/spec.md` for replay / recovery substrate alignment

## 4. Archive and validation

- [x] 4.1 Archive proposal, design, tasks, and docs-only delta spec under `2026-04-26-replay-recovery-policy-runtime-workstream`
- [x] 4.2 Run `.venv/bin/zensical build`
- [x] 4.3 Run `openspec validate docs-only --type spec --strict --no-interactive`
