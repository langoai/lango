## 1. Planning

- [x] 1.1 Create a single OpenSpec change for gateway CLI address resolution
- [x] 1.2 Define affected capabilities and downstream artifacts
- [x] 1.3 Validate the active OpenSpec change
- [x] 1.4 Commit planning artifacts as a scoped commit

## 2. Tests First

- [x] 2.1 Add failing metrics tests for configured gateway default and explicit `--addr` override
- [x] 2.2 Add failing alerts tests for configured gateway default and explicit `--addr` override
- [x] 2.3 Add failing status tests proving the live probe uses the configured gateway when `--addr` is omitted
- [x] 2.4 Add or update executable docs guards for configured-gateway wording

## 3. Implementation

- [x] 3.1 Add a shared CLI gateway address resolver
- [x] 3.2 Wire metrics commands to use explicit `--addr`, then config, then fallback
- [x] 3.3 Wire alerts commands to use explicit `--addr`, then config, then fallback
- [x] 3.4 Wire status root command to probe the configured gateway by default
- [x] 3.5 Wire root `cmd/lango` constructors with the config loader

## 4. Downstream Artifacts

- [x] 4.1 Update metrics, alerts, and status CLI docs after verifying actual help and behavior
- [x] 4.2 Sync main OpenSpec specs after implementation is verified

## 5. Review And Verification

- [x] 5.1 Run subagent implementation or review checkpoints for the changed CLI surface
- [x] 5.2 Run focused package tests for metrics, alerts, status, and docs guards
- [x] 5.3 Run `go build ./...`
- [x] 5.4 Run `go test ./...`
- [x] 5.5 Run `git diff --check`
- [x] 5.6 Run `openspec validate --all --strict`
- [x] 5.7 Archive the completed OpenSpec change
