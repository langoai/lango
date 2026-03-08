// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "../interfaces/IERC20.sol";

/// @notice Minimal ERC-7579 account execution interface.
interface IERC7579Account {
    function execute(address target, uint256 value, bytes calldata callData) external;
}

/// @title LangoEscrowExecutor — ERC-7579 TYPE_EXECUTOR module for batched escrow operations.
/// @notice Creates a deal and deposits tokens into LangoEscrowHub in a single batched call
///         executed through the smart account.
contract LangoEscrowExecutor {
    // ERC-7579 module type constants
    uint256 internal constant TYPE_EXECUTOR = 2;

    struct BatchedEscrowParams {
        address seller;
        address token;
        uint256 amount;
        uint256 deadline;
    }

    // account => authorized session keys
    mapping(address => mapping(address => bool)) public authorizedKeys;

    event EscrowExecuted(address indexed account, address indexed escrowHub, uint256 dealId);
    event SessionKeyAuthorized(address indexed account, address indexed sessionKey);
    event SessionKeyDeauthorized(address indexed account, address indexed sessionKey);

    // ---- IERC7579Module ----

    /// @notice Called when this module is installed on an account.
    /// @param data Optional ABI-encoded list of authorized session keys.
    function onInstall(bytes calldata data) external {
        if (data.length > 0) {
            address[] memory keys = abi.decode(data, (address[]));
            for (uint256 i = 0; i < keys.length; i++) {
                authorizedKeys[msg.sender][keys[i]] = true;
                emit SessionKeyAuthorized(msg.sender, keys[i]);
            }
        }
    }

    /// @notice Called when this module is uninstalled.
    function onUninstall(bytes calldata data) external {
        if (data.length > 0) {
            address[] memory keys = abi.decode(data, (address[]));
            for (uint256 i = 0; i < keys.length; i++) {
                delete authorizedKeys[msg.sender][keys[i]];
                emit SessionKeyDeauthorized(msg.sender, keys[i]);
            }
        }
    }

    /// @notice Returns true if moduleTypeId == 2 (EXECUTOR).
    function isModuleType(uint256 moduleTypeId) external pure returns (bool) {
        return moduleTypeId == TYPE_EXECUTOR;
    }

    // ---- Executor ----

    /// @notice Execute a batched escrow operation: approve + createDeal + deposit.
    /// @dev This function is called by the smart account or an authorized session key.
    ///      It uses IERC7579Account.execute() to perform operations through the account.
    /// @param escrowHub The LangoEscrowHub contract address.
    /// @param params The escrow parameters (seller, token, amount, deadline).
    function executeBatchedEscrow(address escrowHub, BatchedEscrowParams calldata params) external {
        address account = msg.sender;

        require(escrowHub != address(0), "Executor: zero escrow hub");
        require(params.seller != address(0), "Executor: zero seller");
        require(params.amount > 0, "Executor: zero amount");

        // Step 1: Approve the escrow hub to spend tokens from the account
        bytes memory approveData = abi.encodeWithSelector(IERC20.approve.selector, escrowHub, params.amount);
        IERC7579Account(account).execute(params.token, 0, approveData);

        // Step 2: Create deal on the escrow hub
        bytes memory createDealData = abi.encodeWithSignature(
            "createDeal(address,address,uint256,uint256)", params.seller, params.token, params.amount, params.deadline
        );
        IERC7579Account(account).execute(escrowHub, 0, createDealData);

        // Step 3: Deposit — we need the deal ID.
        // The deal ID is nextDealId - 1 after createDeal.
        // Read nextDealId from escrow hub.
        (bool success, bytes memory result) =
            escrowHub.staticcall(abi.encodeWithSignature("nextDealId()"));
        require(success, "Executor: nextDealId call failed");
        uint256 nextId = abi.decode(result, (uint256));
        require(nextId > 0, "Executor: no deal created");
        uint256 dealId = nextId - 1;

        bytes memory depositData = abi.encodeWithSignature("deposit(uint256)", dealId);
        IERC7579Account(account).execute(escrowHub, 0, depositData);

        emit EscrowExecuted(account, escrowHub, dealId);
    }

    /// @notice Authorize a session key to use this executor.
    function authorizeSessionKey(address sessionKey) external {
        require(sessionKey != address(0), "Executor: zero key");
        authorizedKeys[msg.sender][sessionKey] = true;
        emit SessionKeyAuthorized(msg.sender, sessionKey);
    }

    /// @notice Deauthorize a session key.
    function deauthorizeSessionKey(address sessionKey) external {
        authorizedKeys[msg.sender][sessionKey] = false;
        emit SessionKeyDeauthorized(msg.sender, sessionKey);
    }

    /// @notice Check if a session key is authorized for an account.
    function isAuthorized(address account, address sessionKey) external view returns (bool) {
        return authorizedKeys[account][sessionKey];
    }

    // ---- ERC-165 ----

    function supportsInterface(bytes4 interfaceId) external pure returns (bool) {
        return interfaceId == 0x01ffc9a7; // ERC-165
    }
}
