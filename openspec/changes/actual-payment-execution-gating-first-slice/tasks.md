## 1. Shared Payment Gate Service

- [x] 1.1 Add the `internal/paymentgate` package with direct payment allow/deny results and deny reason codes
- [x] 1.2 Add the shared gate service that consumes receipt-backed canonical payment approval state

## 2. Direct Payment Handler Integration

- [ ] 2.1 Gate `payment_send` with the shared payment execution gate
- [ ] 2.2 Gate `p2p_pay` with the shared payment execution gate
- [ ] 2.3 Emit allow/deny execution evidence into audit and receipt trails

## 3. Operator Surface And Docs

- [ ] 3.1 Add minimal operator-facing payment execution gating docs and truth-align surrounding surfaces

## 4. Verification And OpenSpec Closeout

- [x] 4.1 Run targeted tests while implementing each slice
- [ ] 4.2 Run `go test ./...`, `go build ./...`, and `python3 -m mkdocs build --strict`
- [ ] 4.3 Sync main specs and archive the completed change
