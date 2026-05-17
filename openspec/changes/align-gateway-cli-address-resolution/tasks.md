## 1. Planning

- [x] 1.1 Create a single OpenSpec change for gateway CLI address resolution
- [x] 1.2 Define affected capabilities and downstream artifacts
- [ ] 1.3 Validate the active OpenSpec change
- [ ] 1.4 Commit planning artifacts as a scoped commit

## 2. Tests First

- [ ] 2.1 Add failing metrics tests for configured gateway default and explicit `--addr` override
- [ ] 2.2 Add failing alerts tests for configured gateway default and explicit `--addr` override
- [ ] 2.3 Add failing status tests proving the live probe uses the configured gateway when `--addr` is omitted
- [ ] 2.4 Add or update executable docs guards for configured-gateway wording

## 3. Implementation

- [ ] 3.1 Add a shared CLI gateway address resolver
- [ ] 3.2 Wire metrics commands to use explicit `--addr`, then config, then fallback
- [ ] 3.3 Wire alerts commands to use explicit `--addr`, then config, then fallback
- [ ] 3.4 Wire status root command to probe the configured gateway by default
- [ ] 3.5 Wire root `cmd/lango` constructors with the config loader

## 4. Downstream Artifacts

- [ ] 4.1 Update metrics, alerts, and status CLI docs after verifying actual help and behavior
- [ ] 4.2 Sync main OpenSpec specs after implementation is verified

## 5. Review And Verification

- [ ] 5.1 Run subagent implementation or review checkpoints for the changed CLI surface
- [ ] 5.2 Run focused package tests for metrics, alerts, status, and docs guards
- [ ] 5.3 Run `go build ./...`
- [ ] 5.4 Run `go test ./...`
- [ ] 5.5 Run `git diff --check`
- [ ] 5.6 Run `openspec validate --all --strict`
- [ ] 5.7 Archive the completed OpenSpec change
