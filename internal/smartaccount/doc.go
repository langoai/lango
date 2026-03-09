// Package smartaccount provides ERC-7579 modular smart account management
// with session key-based controlled autonomy for blockchain agents.
//
// Architecture:
//   - Account Manager: Safe deployment and UserOp construction
//   - Session Manager: Hierarchical session key lifecycle
//   - Policy Engine: Off-chain pre-flight validation
//   - Module Registry: ERC-7579 module management
//   - Bundler Client: External bundler RPC communication
package smartaccount
