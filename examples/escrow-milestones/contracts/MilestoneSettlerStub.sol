// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title MilestoneSettlerStub — minimal MilestoneSettler for integration tests.
contract MilestoneSettlerStub {
    function version() external pure returns (string memory) {
        return "milestone-stub";
    }
}
