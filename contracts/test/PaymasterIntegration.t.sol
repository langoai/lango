// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/modules/LangoSessionValidator.sol";

/// @notice Mock ERC-20 token for testing.
contract MockUSDC {
    mapping(address => uint256) public balanceOf;
    mapping(address => mapping(address => uint256)) public allowance;

    function mint(address to, uint256 amount) external {
        balanceOf[to] += amount;
    }

    function approve(address spender, uint256 amount) external returns (bool) {
        allowance[msg.sender][spender] = amount;
        return true;
    }

    function transfer(address to, uint256 amount) external returns (bool) {
        require(balanceOf[msg.sender] >= amount, "insufficient balance");
        balanceOf[msg.sender] -= amount;
        balanceOf[to] += amount;
        return true;
    }
}

/// @notice Mock paymaster that simply validates the paymasterAndData.
contract MockPaymaster {
    address public token;

    constructor(address _token) {
        token = _token;
    }

    function validatePaymasterUserOp(
        PackedUserOperation calldata,
        bytes32,
        uint256
    ) external pure returns (bytes memory context, uint256 validationData) {
        return ("", 0);
    }
}

contract PaymasterIntegrationTest is Test {
    MockUSDC public usdc;
    MockPaymaster public paymaster;
    LangoSessionValidator public validator;

    address public account = address(this);
    uint256 internal sessionKeyPk = 0xB0B;
    address public sessionKey;

    function setUp() public {
        usdc = new MockUSDC();
        paymaster = new MockPaymaster(address(usdc));
        validator = new LangoSessionValidator();
        sessionKey = vm.addr(sessionKeyPk);

        // Mint USDC to account
        usdc.mint(account, 1000 * 1e6);
    }

    function testApproveUSDCToPaymaster() public {
        // Account approves USDC to paymaster
        usdc.approve(address(paymaster), type(uint256).max);

        uint256 allowed = usdc.allowance(account, address(paymaster));
        assertEq(allowed, type(uint256).max, "approval should be max uint256");
    }

    function testPaymasterWithSessionKey() public {
        // Register session key with paymaster in allowlist
        address[] memory targets = new address[](1);
        targets[0] = address(usdc);

        bytes4[] memory functions = new bytes4[](1);
        functions[0] = bytes4(keccak256("transfer(address,uint256)"));

        address[] memory paymasters = new address[](1);
        paymasters[0] = address(paymaster);

        ISessionValidator.SessionPolicy memory policy = ISessionValidator.SessionPolicy({
            allowedTargets: targets,
            allowedFunctions: functions,
            spendLimit: 100 * 1e6,
            spentAmount: 0,
            validAfter: uint48(block.timestamp),
            validUntil: uint48(block.timestamp + 1 days),
            active: true,
            allowedPaymasters: paymasters
        });

        validator.registerSessionKey(sessionKey, policy);

        // Build UserOp with paymaster data
        bytes memory innerData = abi.encodeWithSelector(
            bytes4(keccak256("transfer(address,uint256)")),
            address(0x9999),
            uint256(10 * 1e6)
        );
        bytes memory callData = abi.encodeWithSignature(
            "execute(address,uint256,bytes)",
            address(usdc),
            uint256(0),
            innerData
        );

        bytes32 opHash = keccak256("test_pm_session");
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(sessionKeyPk, opHash);
        bytes memory sig = abi.encodePacked(r, s, v);

        bytes memory pmData = abi.encodePacked(address(paymaster), hex"aabbccdd");

        PackedUserOperation memory userOp = PackedUserOperation({
            sender: account,
            nonce: 0,
            initCode: "",
            callData: callData,
            accountGasLimits: bytes32(0),
            preVerificationGas: 0,
            gasFees: bytes32(0),
            paymasterAndData: pmData,
            signature: sig
        });

        uint256 result = validator.validateUserOp(userOp, opHash);
        assertTrue(result != 1, "session key with allowed paymaster should pass");
    }

    function testPaymasterNotInAllowlist_Fails() public {
        address[] memory targets = new address[](1);
        targets[0] = address(usdc);

        bytes4[] memory functions = new bytes4[](1);
        functions[0] = bytes4(keccak256("transfer(address,uint256)"));

        // Only allow a different paymaster
        address[] memory paymasters = new address[](1);
        paymasters[0] = address(0xDEAD);

        ISessionValidator.SessionPolicy memory policy = ISessionValidator.SessionPolicy({
            allowedTargets: targets,
            allowedFunctions: functions,
            spendLimit: 100 * 1e6,
            spentAmount: 0,
            validAfter: uint48(block.timestamp),
            validUntil: uint48(block.timestamp + 1 days),
            active: true,
            allowedPaymasters: paymasters
        });

        validator.registerSessionKey(sessionKey, policy);

        bytes memory innerData = abi.encodeWithSelector(
            bytes4(keccak256("transfer(address,uint256)")),
            address(0x9999),
            uint256(10 * 1e6)
        );
        bytes memory callData = abi.encodeWithSignature(
            "execute(address,uint256,bytes)",
            address(usdc),
            uint256(0),
            innerData
        );

        bytes32 opHash = keccak256("test_pm_wrong");
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(sessionKeyPk, opHash);
        bytes memory sig = abi.encodePacked(r, s, v);

        // Use actual paymaster, not the allowed one
        bytes memory pmData = abi.encodePacked(address(paymaster), hex"aabbccdd");

        PackedUserOperation memory userOp = PackedUserOperation({
            sender: account,
            nonce: 0,
            initCode: "",
            callData: callData,
            accountGasLimits: bytes32(0),
            preVerificationGas: 0,
            gasFees: bytes32(0),
            paymasterAndData: pmData,
            signature: sig
        });

        uint256 result = validator.validateUserOp(userOp, opHash);
        assertEq(result, 1, "wrong paymaster should fail");
    }
}
