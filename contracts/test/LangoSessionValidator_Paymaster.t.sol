// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/modules/LangoSessionValidator.sol";

contract LangoSessionValidator_PaymasterTest is Test {
    LangoSessionValidator public validator;

    address public account = address(this);
    uint256 internal sessionKeyPk = 0xA11CE;
    address public sessionKey;
    address public target1 = address(0x1111);

    address public paymaster1 = address(0xAA01);
    address public paymaster2 = address(0xAA02);
    address public paymaster3 = address(0xAA03);

    uint48 public validAfter;
    uint48 public validUntil;

    function setUp() public {
        validator = new LangoSessionValidator();
        sessionKey = vm.addr(sessionKeyPk);
        validAfter = uint48(block.timestamp);
        validUntil = uint48(block.timestamp + 1 days);
    }

    // ---- Paymaster allowlist tests ----

    function testValidateUserOp_PaymasterAllowed() public {
        ISessionValidator.SessionPolicy memory policy = _policyWithPaymasters();
        validator.registerSessionKey(sessionKey, policy);

        bytes memory innerData = abi.encodeWithSelector(bytes4(0xdeadbeef), uint256(42));
        bytes memory callData =
            abi.encodeWithSignature("execute(address,uint256,bytes)", target1, uint256(0), innerData);

        bytes32 opHash = keccak256("test_pm_allowed");
        bytes memory sig = _sign(sessionKeyPk, opHash);

        // paymaster1 is in the allowlist
        bytes memory pmData = abi.encodePacked(paymaster1, hex"0011223344");

        PackedUserOperation memory userOp = _buildUserOp(account, callData, sig, pmData);
        uint256 result = validator.validateUserOp(userOp, opHash);
        assertTrue(result != 1, "allowed paymaster should pass");
    }

    function testValidateUserOp_PaymasterNotAllowed() public {
        ISessionValidator.SessionPolicy memory policy = _policyWithPaymasters();
        validator.registerSessionKey(sessionKey, policy);

        bytes memory innerData = abi.encodeWithSelector(bytes4(0xdeadbeef), uint256(42));
        bytes memory callData =
            abi.encodeWithSignature("execute(address,uint256,bytes)", target1, uint256(0), innerData);

        bytes32 opHash = keccak256("test_pm_not_allowed");
        bytes memory sig = _sign(sessionKeyPk, opHash);

        // paymaster3 is NOT in the allowlist
        bytes memory pmData = abi.encodePacked(paymaster3, hex"0011223344");

        PackedUserOperation memory userOp = _buildUserOp(account, callData, sig, pmData);
        uint256 result = validator.validateUserOp(userOp, opHash);
        assertEq(result, 1, "disallowed paymaster should fail");
    }

    function testValidateUserOp_EmptyAllowedPaymasters() public {
        // Empty allowedPaymasters = all paymasters allowed
        ISessionValidator.SessionPolicy memory policy = _policyNoPaymasterRestriction();
        validator.registerSessionKey(sessionKey, policy);

        bytes memory innerData = abi.encodeWithSelector(bytes4(0xdeadbeef), uint256(42));
        bytes memory callData =
            abi.encodeWithSignature("execute(address,uint256,bytes)", target1, uint256(0), innerData);

        bytes32 opHash = keccak256("test_pm_empty");
        bytes memory sig = _sign(sessionKeyPk, opHash);

        bytes memory pmData = abi.encodePacked(paymaster3, hex"aabbccdd");

        PackedUserOperation memory userOp = _buildUserOp(account, callData, sig, pmData);
        uint256 result = validator.validateUserOp(userOp, opHash);
        assertTrue(result != 1, "empty allowlist should allow any paymaster");
    }

    function testValidateUserOp_NoPaymaster_WithAllowlist() public {
        // paymasterAndData is empty, but allowlist is set — should still pass
        ISessionValidator.SessionPolicy memory policy = _policyWithPaymasters();
        validator.registerSessionKey(sessionKey, policy);

        bytes memory innerData = abi.encodeWithSelector(bytes4(0xdeadbeef), uint256(42));
        bytes memory callData =
            abi.encodeWithSignature("execute(address,uint256,bytes)", target1, uint256(0), innerData);

        bytes32 opHash = keccak256("test_no_pm");
        bytes memory sig = _sign(sessionKeyPk, opHash);

        PackedUserOperation memory userOp = _buildUserOp(account, callData, sig, "");
        uint256 result = validator.validateUserOp(userOp, opHash);
        assertTrue(result != 1, "no paymaster with allowlist should pass");
    }

    function testValidateUserOp_ShortPaymasterData() public {
        // paymasterAndData < 20 bytes — should not trigger allowlist check
        ISessionValidator.SessionPolicy memory policy = _policyWithPaymasters();
        validator.registerSessionKey(sessionKey, policy);

        bytes memory innerData = abi.encodeWithSelector(bytes4(0xdeadbeef), uint256(42));
        bytes memory callData =
            abi.encodeWithSignature("execute(address,uint256,bytes)", target1, uint256(0), innerData);

        bytes32 opHash = keccak256("test_short_pm");
        bytes memory sig = _sign(sessionKeyPk, opHash);

        // Only 10 bytes — too short for paymaster address
        bytes memory pmData = hex"00112233445566778899";

        PackedUserOperation memory userOp = _buildUserOp(account, callData, sig, pmData);
        uint256 result = validator.validateUserOp(userOp, opHash);
        assertTrue(result != 1, "short paymasterData should not trigger check");
    }

    function testRegisterSessionKey_WithPaymasterAllowlist() public {
        ISessionValidator.SessionPolicy memory policy = _policyWithPaymasters();
        validator.registerSessionKey(sessionKey, policy);

        ISessionValidator.SessionPolicy memory stored = validator.getSessionKeyPolicy(sessionKey);
        assertEq(stored.allowedPaymasters.length, 2);
        assertEq(stored.allowedPaymasters[0], paymaster1);
        assertEq(stored.allowedPaymasters[1], paymaster2);
    }

    // ---- Helpers ----

    function _policyWithPaymasters() internal view returns (ISessionValidator.SessionPolicy memory) {
        address[] memory targets = new address[](1);
        targets[0] = target1;

        bytes4[] memory functions = new bytes4[](1);
        functions[0] = bytes4(0xdeadbeef);

        address[] memory paymasters = new address[](2);
        paymasters[0] = paymaster1;
        paymasters[1] = paymaster2;

        return ISessionValidator.SessionPolicy({
            allowedTargets: targets,
            allowedFunctions: functions,
            spendLimit: 1 ether,
            spentAmount: 0,
            validAfter: validAfter,
            validUntil: validUntil,
            active: true,
            allowedPaymasters: paymasters
        });
    }

    function _policyNoPaymasterRestriction() internal view returns (ISessionValidator.SessionPolicy memory) {
        address[] memory targets = new address[](1);
        targets[0] = target1;

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

    function _buildUserOp(address sender, bytes memory callData, bytes memory sig, bytes memory pmData)
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
            paymasterAndData: pmData,
            signature: sig
        });
    }
}
