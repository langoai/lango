// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./ISessionValidator.sol";

/// @notice ERC-4337 PackedUserOperation struct.
struct PackedUserOperation {
    address sender;
    uint256 nonce;
    bytes initCode;
    bytes callData;
    bytes32 accountGasLimits;
    uint256 preVerificationGas;
    bytes32 gasFees;
    bytes paymasterAndData;
    bytes signature;
}

/// @notice Minimal ERC-7579 module interface.
interface IERC7579Module {
    function onInstall(bytes calldata data) external;
    function onUninstall(bytes calldata data) external;
    function isModuleType(uint256 moduleTypeId) external view returns (bool);
}

/// @title LangoSessionValidator — ERC-7579 TYPE_VALIDATOR module for session key management.
/// @notice Validates user operations against registered session key policies.
///         Enforces target/function allow-lists and spending limits.
contract LangoSessionValidator is IERC7579Module, ISessionValidator {
    // ERC-7579 module type constants
    uint256 internal constant TYPE_VALIDATOR = 1;

    // account => sessionKey => policy
    mapping(address => mapping(address => SessionPolicy)) internal _sessions;

    // ---- IERC7579Module ----

    /// @notice Called when this module is installed on an account.
    /// @param data Optional encoded session key + policy to register on install.
    function onInstall(bytes calldata data) external override {
        if (data.length > 0) {
            (address sessionKey, SessionPolicy memory policy) = abi.decode(data, (address, SessionPolicy));
            _setSession(msg.sender, sessionKey, policy);
            emit SessionKeyRegistered(msg.sender, sessionKey, policy.validUntil);
        }
    }

    /// @notice Called when this module is uninstalled. Cleans up given session key.
    /// @param data ABI-encoded session key address to revoke.
    function onUninstall(bytes calldata data) external override {
        if (data.length > 0) {
            address sessionKey = abi.decode(data, (address));
            delete _sessions[msg.sender][sessionKey];
            emit SessionKeyRevoked(msg.sender, sessionKey);
        }
    }

    /// @notice Returns true if moduleTypeId == 1 (VALIDATOR).
    function isModuleType(uint256 moduleTypeId) external pure override returns (bool) {
        return moduleTypeId == TYPE_VALIDATOR;
    }

    // ---- ISessionValidator ----

    /// @notice Register a session key with a given policy. Only callable by the account itself.
    function registerSessionKey(address sessionKey, SessionPolicy calldata policy) external override {
        require(sessionKey != address(0), "SV: zero session key");
        require(policy.validUntil > policy.validAfter, "SV: invalid validity window");

        SessionPolicy memory p = policy;
        p.active = true;
        p.spentAmount = 0;
        _setSession(msg.sender, sessionKey, p);

        emit SessionKeyRegistered(msg.sender, sessionKey, policy.validUntil);
    }

    /// @notice Revoke a session key. Only callable by the account itself.
    function revokeSessionKey(address sessionKey) external override {
        require(_sessions[msg.sender][sessionKey].active, "SV: not active");
        _sessions[msg.sender][sessionKey].active = false;
        emit SessionKeyRevoked(msg.sender, sessionKey);
    }

    /// @notice Get the session key policy for the calling account.
    function getSessionKeyPolicy(address sessionKey) external view override returns (SessionPolicy memory) {
        return _sessions[msg.sender][sessionKey];
    }

    /// @notice Check whether a session key is active and not expired.
    function isSessionKeyActive(address sessionKey) external view override returns (bool) {
        return _isActive(msg.sender, sessionKey);
    }

    // ---- Validation ----

    /// @notice Validate a user operation signed by a session key.
    /// @param userOp The packed user operation.
    /// @param userOpHash The hash of the user operation (signed by session key).
    /// @return validationData 0 on success, 1 on failure. Packed with validAfter/validUntil per ERC-4337.
    function validateUserOp(PackedUserOperation calldata userOp, bytes32 userOpHash) external returns (uint256) {
        address account = userOp.sender;

        // Recover signer from signature
        address signer = _recoverSigner(userOpHash, userOp.signature);
        if (signer == address(0)) {
            return 1; // SIG_VALIDATION_FAILED
        }

        SessionPolicy storage session = _sessions[account][signer];

        // Check session is active and not expired
        if (!session.active) {
            return 1;
        }
        if (block.timestamp < session.validAfter || block.timestamp > session.validUntil) {
            return 1;
        }

        // Extract target and function selector from callData
        if (userOp.callData.length >= 4) {
            (address target, uint256 value, bytes memory innerData) = _decodeExecuteCallData(userOp.callData);

            // Check allowed targets
            if (session.allowedTargets.length > 0) {
                bool targetAllowed = false;
                for (uint256 i = 0; i < session.allowedTargets.length; i++) {
                    if (session.allowedTargets[i] == target) {
                        targetAllowed = true;
                        break;
                    }
                }
                if (!targetAllowed) {
                    return 1;
                }
            }

            // Check allowed function selectors
            if (session.allowedFunctions.length > 0 && innerData.length >= 4) {
                bytes4 selector;
                assembly {
                    selector := mload(add(innerData, 32))
                }
                bool funcAllowed = false;
                for (uint256 i = 0; i < session.allowedFunctions.length; i++) {
                    if (session.allowedFunctions[i] == selector) {
                        funcAllowed = true;
                        break;
                    }
                }
                if (!funcAllowed) {
                    return 1;
                }
            }

            // Check and update spend limit
            if (value > 0 && session.spendLimit > 0) {
                if (session.spentAmount + value > session.spendLimit) {
                    return 1;
                }
                session.spentAmount += value;
            }
        }

        // Check paymaster allowlist
        if (session.allowedPaymasters.length > 0 && userOp.paymasterAndData.length >= 20) {
            address paymaster = address(bytes20(userOp.paymasterAndData[:20]));
            bool paymasterAllowed = false;
            for (uint256 i = 0; i < session.allowedPaymasters.length; i++) {
                if (session.allowedPaymasters[i] == paymaster) {
                    paymasterAllowed = true;
                    break;
                }
            }
            if (!paymasterAllowed) {
                return 1;
            }
        }

        // Pack validAfter and validUntil into validationData
        // validationData = sigFailed (0) | validUntil (6 bytes) | validAfter (6 bytes)
        return _packValidationData(session.validAfter, session.validUntil);
    }

    // ---- ERC-165 ----

    /// @notice ERC-165 interface support.
    function supportsInterface(bytes4 interfaceId) external pure returns (bool) {
        return interfaceId == type(ISessionValidator).interfaceId || interfaceId == type(IERC7579Module).interfaceId
            || interfaceId == 0x01ffc9a7; // ERC-165
    }

    // ---- Internal ----

    function _setSession(address account, address sessionKey, SessionPolicy memory policy) internal {
        SessionPolicy storage s = _sessions[account][sessionKey];
        s.allowedTargets = policy.allowedTargets;
        s.allowedFunctions = policy.allowedFunctions;
        s.spendLimit = policy.spendLimit;
        s.spentAmount = policy.spentAmount;
        s.validAfter = policy.validAfter;
        s.validUntil = policy.validUntil;
        s.active = policy.active;
        s.allowedPaymasters = policy.allowedPaymasters;
    }

    function _isActive(address account, address sessionKey) internal view returns (bool) {
        SessionPolicy storage s = _sessions[account][sessionKey];
        return s.active && block.timestamp >= s.validAfter && block.timestamp <= s.validUntil;
    }

    /// @dev Recover signer from ECDSA signature (v, r, s packed as 65 bytes).
    function _recoverSigner(bytes32 hash, bytes memory signature) internal pure returns (address) {
        if (signature.length != 65) {
            return address(0);
        }

        bytes32 r;
        bytes32 s;
        uint8 v;

        assembly {
            r := mload(add(signature, 32))
            s := mload(add(signature, 64))
            v := byte(0, mload(add(signature, 96)))
        }

        // EIP-2: s-value constraint
        if (uint256(s) > 0x7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF5D576E7357A4501DDFE92F46681B20A0) {
            return address(0);
        }

        if (v != 27 && v != 28) {
            return address(0);
        }

        return ecrecover(hash, v, r, s);
    }

    /// @dev Decode execute(address,uint256,bytes) call data.
    function _decodeExecuteCallData(bytes calldata callData) internal pure returns (address target, uint256 value, bytes memory data) {
        // Skip the 4-byte function selector, then decode (address, uint256, bytes)
        if (callData.length < 68) {
            return (address(0), 0, "");
        }
        (target, value, data) = abi.decode(callData[4:], (address, uint256, bytes));
    }

    /// @dev Pack validAfter and validUntil into ERC-4337 validationData format.
    function _packValidationData(uint48 validAfter, uint48 validUntil) internal pure returns (uint256) {
        return (uint256(validUntil) << 160) | (uint256(validAfter) << (160 + 48));
    }
}
