// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title FactoryStub — minimal Safe factory for integration tests.
contract FactoryStub {
    // Stub — provides a valid contract address for config injection.
    function proxyCreationCode() external pure returns (bytes memory) {
        return "";
    }
}
