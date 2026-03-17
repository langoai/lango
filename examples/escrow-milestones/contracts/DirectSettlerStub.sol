// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title DirectSettlerStub — minimal DirectSettler for integration tests.
contract DirectSettlerStub {
    function version() external pure returns (string memory) {
        return "direct-stub";
    }
}
