## MODIFIED Requirements

### Requirement: WithApproval middleware
The approval middleware SHALL gate tool execution behind the approval flow. It SHALL populate `ApprovalRequest.SafetyLevel`, `ApprovalRequest.Category`, and `ApprovalRequest.Activity` from the tool's metadata before sending the request.

#### Scenario: Dangerous tool requires approval
- **WHEN** a tool with `SafetyLevel: Dangerous` is invoked and approval policy requires it
- **THEN** the middleware sends an `ApprovalRequest` and blocks until response

#### Scenario: Exempt tool bypasses approval
- **WHEN** a tool is listed in `ExemptTools`
- **THEN** the middleware skips approval and executes directly

#### Scenario: SafetyLevel populated from tool metadata
- **WHEN** the middleware creates an `ApprovalRequest`
- **THEN** `req.SafetyLevel` is set to `tool.SafetyLevel.String()`

#### Scenario: Category and Activity populated from tool capability
- **WHEN** the middleware creates an `ApprovalRequest` for a tool with `Capability.Category: "filesystem"` and `Capability.Activity: "write"`
- **THEN** `req.Category` is `"filesystem"` and `req.Activity` is `"write"`
