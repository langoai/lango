// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title EntryPointStub — minimal ERC-4337 EntryPoint for integration tests.
/// @dev Provides a deployable contract address. Real EntryPoint logic is not needed
///      for Lango's smart account tool testing as the tools use local simulation.
contract EntryPointStub {
    // Stub — provides a valid contract address for config injection.
    function getNonce(address, uint192) external pure returns (uint256) {
        return 0;
    }
}
