// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title ISessionValidator — ERC-7579 session key validator interface.
/// @notice Defines session key management for modular smart accounts.
interface ISessionValidator {
    struct SessionPolicy {
        address[] allowedTargets;
        bytes4[] allowedFunctions;
        uint256 spendLimit;
        uint256 spentAmount;
        uint48 validAfter;
        uint48 validUntil;
        bool active;
    }

    event SessionKeyRegistered(address indexed account, address indexed sessionKey, uint48 validUntil);
    event SessionKeyRevoked(address indexed account, address indexed sessionKey);

    function registerSessionKey(address sessionKey, SessionPolicy calldata policy) external;
    function revokeSessionKey(address sessionKey) external;
    function getSessionKeyPolicy(address sessionKey) external view returns (SessionPolicy memory);
    function isSessionKeyActive(address sessionKey) external view returns (bool);
}
