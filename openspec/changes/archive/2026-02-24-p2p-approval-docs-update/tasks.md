## 1. P2P Feature Docs (docs/features/p2p-network.md)

- [x] 1.1 Add "Approval Pipeline" section with Mermaid flowchart after Knowledge Firewall section
- [x] 1.2 Add "Auto-Approval for Small Amounts" subsection in Paid Value Exchange section
- [x] 1.3 Add reputation and pricing endpoints to REST API table with curl examples
- [x] 1.4 Add reputation and pricing CLI commands to CLI Commands listing

## 2. README.md

- [x] 2.1 Add "Approval Pipeline" bullet to P2P feature list
- [x] 2.2 Add auto-approval note after Paid Value Exchange flow
- [x] 2.3 Add missing P2P config fields (autoApproveKnownPeers, minTrustScore, pricing.enabled, pricing.perQuery) to config reference table
- [x] 2.4 Add reputation and pricing curl examples to REST API section

## 3. Prompts (prompts/TOOL_USAGE.md)

- [x] 3.1 Update p2p_pay description with auto-approval behavior
- [x] 3.2 Update p2p_query description with remote owner approval pipeline
- [x] 3.3 Update paid tool workflow with auto-approval and approval pipeline notes
- [x] 3.4 Add inbound tool invocation three-stage gate description

## 4. HTTP API Docs (docs/gateway/http-api.md)

- [x] 4.1 Add GET /api/p2p/reputation section with query params, JSON response, and curl example
- [x] 4.2 Add GET /api/p2p/pricing section with query params, JSON response, and curl examples

## 5. Payment Docs (docs/payments/usdc.md)

- [x] 5.1 Add P2P integration note after config table explaining cross-cutting autoApproveBelow behavior

## 6. Example Docs (examples/p2p-trading/README.md)

- [x] 6.1 Add "Configuration Highlights" section with approval and payment settings table
- [x] 6.2 Add reputation and pricing endpoints to REST API table

## 7. Build (Makefile)

- [x] 7.1 Add test-p2p target running P2P and wallet tests with race detector
- [x] 7.2 Add test-p2p to .PHONY declaration

## 8. Verification

- [x] 8.1 Run go build ./... to verify no build errors
- [x] 8.2 Run make test-p2p to verify new Makefile target works
