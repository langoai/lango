## REMOVED Requirements

### Requirement: Legacy ApprovalRequired migration
**Reason**: The `ApprovalRequired` bool field and `migrateApprovalPolicy()` function provided backward compatibility for an older config format. As a pre-release project, no migration is needed.
**Migration**: Use `approvalPolicy` field directly in config. Default is `"dangerous"`.
