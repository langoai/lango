// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title EscrowHubV2Stub — minimal EscrowHubV2 for integration tests.
contract EscrowHubV2Stub {
    function version() external pure returns (string memory) {
        return "v2-stub";
    }
}
