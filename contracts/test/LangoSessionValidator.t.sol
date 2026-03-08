// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/modules/LangoSessionValidator.sol";

contract LangoSessionValidatorTest is Test {
    LangoSessionValidator public validator;

    address public account = address(this);
    uint256 internal sessionKeyPk = 0xA11CE;
    address public sessionKey;
    address public target1 = address(0x1111);
    address public target2 = address(0x2222);

    uint48 public validAfter;
    uint48 public validUntil;

    function setUp() public {
        validator = new LangoSessionValidator();
        sessionKey = vm.addr(sessionKeyPk);
        validAfter = uint48(block.timestamp);
        validUntil = uint48(block.timestamp + 1 days);
    }

    // ---- registerSessionKey ----

    function test_registerSessionKey_storesPolicy() public {
        ISessionValidator.SessionPolicy memory policy = _defaultPolicy();

        validator.registerSessionKey(sessionKey, policy);

        ISessionValidator.SessionPolicy memory stored = validator.getSessionKeyPolicy(sessionKey);
        assertEq(stored.allowedTargets.length, 2);
        assertEq(stored.allowedTargets[0], target1);
        assertEq(stored.allowedTargets[1], target2);
        assertEq(stored.allowedFunctions.length, 1);
        assertEq(stored.allowedFunctions[0], bytes4(0xdeadbeef));
        assertEq(stored.spendLimit, 1 ether);
        assertEq(stored.spentAmount, 0);
        assertEq(stored.validAfter, validAfter);
        assertEq(stored.validUntil, validUntil);
        assertTrue(stored.active);
    }

    function test_registerSessionKey_emitsEvent() public {
        ISessionValidator.SessionPolicy memory policy = _defaultPolicy();

        vm.expectEmit(true, true, false, true);
        emit ISessionValidator.SessionKeyRegistered(account, sessionKey, validUntil);
        validator.registerSessionKey(sessionKey, policy);
    }

    function testRevert_registerSessionKey_zeroAddress() public {
        ISessionValidator.SessionPolicy memory policy = _defaultPolicy();

        vm.expectRevert("SV: zero session key");
        validator.registerSessionKey(address(0), policy);
    }

    function testRevert_registerSessionKey_invalidWindow() public {
        ISessionValidator.SessionPolicy memory policy = _defaultPolicy();
        policy.validAfter = uint48(block.timestamp + 2 days);
        policy.validUntil = uint48(block.timestamp + 1 days);

        vm.expectRevert("SV: invalid validity window");
        validator.registerSessionKey(sessionKey, policy);
    }

    // ---- revokeSessionKey ----

    function test_revokeSessionKey_deactivates() public {
        validator.registerSessionKey(sessionKey, _defaultPolicy());
        assertTrue(validator.isSessionKeyActive(sessionKey));

        validator.revokeSessionKey(sessionKey);
        assertFalse(validator.isSessionKeyActive(sessionKey));
    }

    function test_revokeSessionKey_emitsEvent() public {
        validator.registerSessionKey(sessionKey, _defaultPolicy());

        vm.expectEmit(true, true, false, false);
        emit ISessionValidator.SessionKeyRevoked(account, sessionKey);
        validator.revokeSessionKey(sessionKey);
    }

    function testRevert_revokeSessionKey_notActive() public {
        vm.expectRevert("SV: not active");
        validator.revokeSessionKey(sessionKey);
    }

    // ---- isSessionKeyActive ----

    function test_isSessionKeyActive_returnsTrue() public {
        validator.registerSessionKey(sessionKey, _defaultPolicy());
        assertTrue(validator.isSessionKeyActive(sessionKey));
    }

    function test_isSessionKeyActive_expiredReturnsFalse() public {
        ISessionValidator.SessionPolicy memory policy = _defaultPolicy();
        policy.validAfter = uint48(block.timestamp);
        policy.validUntil = uint48(block.timestamp + 100);
        validator.registerSessionKey(sessionKey, policy);

        vm.warp(block.timestamp + 200);
        assertFalse(validator.isSessionKeyActive(sessionKey));
    }

    function test_isSessionKeyActive_notYetValidReturnsFalse() public {
        ISessionValidator.SessionPolicy memory policy = _defaultPolicy();
        policy.validAfter = uint48(block.timestamp + 1000);
        policy.validUntil = uint48(block.timestamp + 2000);
        validator.registerSessionKey(sessionKey, policy);

        assertFalse(validator.isSessionKeyActive(sessionKey));
    }

    // ---- validateUserOp ----

    function test_validateUserOp_validSession() public {
        validator.registerSessionKey(sessionKey, _defaultPolicy());

        // Build a user operation with callData = execute(target1, 0, 0xdeadbeef...)
        bytes memory innerData = abi.encodeWithSelector(bytes4(0xdeadbeef), uint256(42));
        bytes memory callData =
            abi.encodeWithSignature("execute(address,uint256,bytes)", target1, uint256(0), innerData);

        bytes32 opHash = keccak256("test_op");
        bytes memory sig = _sign(sessionKeyPk, opHash);

        PackedUserOperation memory userOp = _buildUserOp(account, callData, sig);

        uint256 result = validator.validateUserOp(userOp, opHash);
        // result should contain packed validAfter/validUntil, not 1 (failure)
        assertTrue(result != 1, "validation should succeed");
    }

    function test_validateUserOp_revokedSessionFails() public {
        validator.registerSessionKey(sessionKey, _defaultPolicy());
        validator.revokeSessionKey(sessionKey);

        bytes memory callData =
            abi.encodeWithSignature("execute(address,uint256,bytes)", target1, uint256(0), hex"deadbeef");
        bytes32 opHash = keccak256("test_op");
        bytes memory sig = _sign(sessionKeyPk, opHash);

        PackedUserOperation memory userOp = _buildUserOp(account, callData, sig);

        uint256 result = validator.validateUserOp(userOp, opHash);
        assertEq(result, 1, "revoked session should fail");
    }

    function test_validateUserOp_expiredSessionFails() public {
        ISessionValidator.SessionPolicy memory policy = _defaultPolicy();
        policy.validAfter = uint48(block.timestamp);
        policy.validUntil = uint48(block.timestamp + 100);
        validator.registerSessionKey(sessionKey, policy);

        vm.warp(block.timestamp + 200);

        bytes memory callData =
            abi.encodeWithSignature("execute(address,uint256,bytes)", target1, uint256(0), hex"deadbeef");
        bytes32 opHash = keccak256("test_op");
        bytes memory sig = _sign(sessionKeyPk, opHash);

        PackedUserOperation memory userOp = _buildUserOp(account, callData, sig);

        uint256 result = validator.validateUserOp(userOp, opHash);
        assertEq(result, 1, "expired session should fail");
    }

    function test_validateUserOp_disallowedTargetFails() public {
        validator.registerSessionKey(sessionKey, _defaultPolicy());

        address disallowedTarget = address(0x9999);
        bytes memory callData =
            abi.encodeWithSignature("execute(address,uint256,bytes)", disallowedTarget, uint256(0), hex"deadbeef");
        bytes32 opHash = keccak256("test_op");
        bytes memory sig = _sign(sessionKeyPk, opHash);

        PackedUserOperation memory userOp = _buildUserOp(account, callData, sig);

        uint256 result = validator.validateUserOp(userOp, opHash);
        assertEq(result, 1, "disallowed target should fail");
    }

    function test_validateUserOp_disallowedFunctionFails() public {
        validator.registerSessionKey(sessionKey, _defaultPolicy());

        bytes memory innerData = abi.encodeWithSelector(bytes4(0x11111111), uint256(42));
        bytes memory callData =
            abi.encodeWithSignature("execute(address,uint256,bytes)", target1, uint256(0), innerData);
        bytes32 opHash = keccak256("test_op");
        bytes memory sig = _sign(sessionKeyPk, opHash);

        PackedUserOperation memory userOp = _buildUserOp(account, callData, sig);

        uint256 result = validator.validateUserOp(userOp, opHash);
        assertEq(result, 1, "disallowed function should fail");
    }

    function test_validateUserOp_spendLimitEnforced() public {
        ISessionValidator.SessionPolicy memory policy = _defaultPolicy();
        policy.spendLimit = 0.5 ether;
        validator.registerSessionKey(sessionKey, policy);

        // First call with 0.3 ether — should pass
        bytes memory innerData = abi.encodeWithSelector(bytes4(0xdeadbeef));
        bytes memory callData =
            abi.encodeWithSignature("execute(address,uint256,bytes)", target1, uint256(0.3 ether), innerData);
        bytes32 opHash1 = keccak256("op1");
        bytes memory sig1 = _sign(sessionKeyPk, opHash1);

        PackedUserOperation memory userOp1 = _buildUserOp(account, callData, sig1);
        uint256 result1 = validator.validateUserOp(userOp1, opHash1);
        assertTrue(result1 != 1, "first spend should pass");

        // Second call with 0.3 ether — should fail (total 0.6 > 0.5 limit)
        bytes32 opHash2 = keccak256("op2");
        bytes memory sig2 = _sign(sessionKeyPk, opHash2);

        bytes memory callData2 =
            abi.encodeWithSignature("execute(address,uint256,bytes)", target1, uint256(0.3 ether), innerData);
        PackedUserOperation memory userOp2 = _buildUserOp(account, callData2, sig2);
        uint256 result2 = validator.validateUserOp(userOp2, opHash2);
        assertEq(result2, 1, "exceeding spend limit should fail");
    }

    // ---- onInstall / onUninstall ----

    function test_onInstall_registersSession() public {
        ISessionValidator.SessionPolicy memory policy = _defaultPolicy();
        bytes memory data = abi.encode(sessionKey, policy);

        validator.onInstall(data);

        ISessionValidator.SessionPolicy memory stored = validator.getSessionKeyPolicy(sessionKey);
        assertTrue(stored.active);
        assertEq(stored.spendLimit, policy.spendLimit);
    }

    function test_onUninstall_removesSession() public {
        validator.registerSessionKey(sessionKey, _defaultPolicy());
        assertTrue(validator.isSessionKeyActive(sessionKey));

        validator.onUninstall(abi.encode(sessionKey));
        assertFalse(validator.isSessionKeyActive(sessionKey));
    }

    // ---- isModuleType ----

    function test_isModuleType_validator() public view {
        assertTrue(validator.isModuleType(1));
        assertFalse(validator.isModuleType(2));
        assertFalse(validator.isModuleType(4));
    }

    // ---- supportsInterface ----

    function test_supportsInterface() public view {
        assertTrue(validator.supportsInterface(0x01ffc9a7)); // ERC-165
        assertTrue(validator.supportsInterface(type(ISessionValidator).interfaceId));
        assertTrue(validator.supportsInterface(type(IERC7579Module).interfaceId));
    }

    // ---- Helpers ----

    function _defaultPolicy() internal view returns (ISessionValidator.SessionPolicy memory) {
        address[] memory targets = new address[](2);
        targets[0] = target1;
        targets[1] = target2;

        bytes4[] memory functions = new bytes4[](1);
        functions[0] = bytes4(0xdeadbeef);

        address[] memory emptyPaymasters = new address[](0);

        return ISessionValidator.SessionPolicy({
            allowedTargets: targets,
            allowedFunctions: functions,
            spendLimit: 1 ether,
            spentAmount: 0,
            validAfter: validAfter,
            validUntil: validUntil,
            active: true,
            allowedPaymasters: emptyPaymasters
        });
    }

    function _sign(uint256 pk, bytes32 hash) internal returns (bytes memory) {
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(pk, hash);
        return abi.encodePacked(r, s, v);
    }

    function _buildUserOp(address sender, bytes memory callData, bytes memory sig)
        internal
        pure
        returns (PackedUserOperation memory)
    {
        return PackedUserOperation({
            sender: sender,
            nonce: 0,
            initCode: "",
            callData: callData,
            accountGasLimits: bytes32(0),
            preVerificationGas: 0,
            gasFees: bytes32(0),
            paymasterAndData: "",
            signature: sig
        });
    }
}
