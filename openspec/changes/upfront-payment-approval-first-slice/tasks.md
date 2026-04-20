## 1. Upfront Payment Approval Domain

- [x] 1.1 Add the `internal/paymentapproval` package with structured decision inputs, outcomes, suggested payment modes, and amount/risk classes
- [x] 1.2 Add the core upfront payment approval evaluator with approve/reject/escalate behavior

## 2. Approval Receipt And Transaction Update

- [x] 2.1 Extend receipt models with canonical payment approval state
- [x] 2.2 Add an `approve_upfront_payment` meta tool and payment approval event updates

## 3. Operator Surface And Docs

- [ ] 3.1 Add minimal operator-facing upfront payment approval docs and truth-align surrounding surfaces

## 4. Verification And OpenSpec Closeout

- [ ] 4.1 Run targeted tests while implementing each slice
- [ ] 4.2 Run `go test ./...`, `go build ./...`, and `python3 -m mkdocs build --strict`
- [ ] 4.3 Sync main specs and archive the completed change
