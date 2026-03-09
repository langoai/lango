// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/modules/LangoEscrowExecutor.sol";
import "../src/LangoEscrowHub.sol";
import "./mocks/MockUSDC.sol";

/// @notice Mock smart account that implements IERC7579Account.
///         Executes calls on behalf of the account (address(this)).
contract MockSmartAccount is IERC7579Account {
    function execute(address target, uint256 value, bytes calldata callData) external override {
        (bool success, bytes memory ret) = target.call{value: value}(callData);
        if (!success) {
            assembly {
                revert(add(ret, 32), mload(ret))
            }
        }
    }

    receive() external payable {}
}

contract LangoEscrowExecutorTest is Test {
    LangoEscrowExecutor public executor;
    LangoEscrowHub public hub;
    MockUSDC public usdc;
    MockSmartAccount public smartAccount;

    address public arbitrator = address(0xA);
    address public seller = address(0xC);
    address public stranger = address(0xD);
    uint256 public constant AMOUNT = 1000e6;

    function setUp() public {
        executor = new LangoEscrowExecutor();
        hub = new LangoEscrowHub(arbitrator);
        usdc = new MockUSDC();
        smartAccount = new MockSmartAccount();

        // Mint tokens to the smart account
        usdc.mint(address(smartAccount), 10_000e6);
    }

    // ---- executeBatchedEscrow ----

    function test_executeBatchedEscrow_success() public {
        uint256 deadline = block.timestamp + 1 days;

        LangoEscrowExecutor.BatchedEscrowParams memory params = LangoEscrowExecutor.BatchedEscrowParams({
            seller: seller,
            token: address(usdc),
            amount: AMOUNT,
            deadline: deadline
        });

        // Call executor from the smart account context
        vm.prank(address(smartAccount));
        executor.executeBatchedEscrow(address(hub), params);

        // Verify deal was created and deposited
        LangoEscrowHub.Deal memory deal = hub.getDeal(0);
        assertEq(deal.buyer, address(smartAccount));
        assertEq(deal.seller, seller);
        assertEq(deal.token, address(usdc));
        assertEq(deal.amount, AMOUNT);
        assertEq(uint8(deal.status), uint8(LangoEscrowHub.DealStatus.Deposited));

        // Verify tokens moved to escrow hub
        assertEq(usdc.balanceOf(address(hub)), AMOUNT);
        assertEq(usdc.balanceOf(address(smartAccount)), 10_000e6 - AMOUNT);
    }

    function test_executeBatchedEscrow_emitsEvent() public {
        uint256 deadline = block.timestamp + 1 days;

        LangoEscrowExecutor.BatchedEscrowParams memory params = LangoEscrowExecutor.BatchedEscrowParams({
            seller: seller,
            token: address(usdc),
            amount: AMOUNT,
            deadline: deadline
        });

        vm.prank(address(smartAccount));
        vm.expectEmit(true, true, false, true);
        emit LangoEscrowExecutor.EscrowExecuted(address(smartAccount), address(hub), 0);
        executor.executeBatchedEscrow(address(hub), params);
    }

    // ---- Validation ----

    function testRevert_executeBatchedEscrow_zeroEscrowHub() public {
        LangoEscrowExecutor.BatchedEscrowParams memory params = LangoEscrowExecutor.BatchedEscrowParams({
            seller: seller,
            token: address(usdc),
            amount: AMOUNT,
            deadline: block.timestamp + 1 days
        });

        vm.prank(address(smartAccount));
        vm.expectRevert("Executor: zero escrow hub");
        executor.executeBatchedEscrow(address(0), params);
    }

    function testRevert_executeBatchedEscrow_zeroSeller() public {
        LangoEscrowExecutor.BatchedEscrowParams memory params = LangoEscrowExecutor.BatchedEscrowParams({
            seller: address(0),
            token: address(usdc),
            amount: AMOUNT,
            deadline: block.timestamp + 1 days
        });

        vm.prank(address(smartAccount));
        vm.expectRevert("Executor: zero seller");
        executor.executeBatchedEscrow(address(hub), params);
    }

    function testRevert_executeBatchedEscrow_zeroAmount() public {
        LangoEscrowExecutor.BatchedEscrowParams memory params = LangoEscrowExecutor.BatchedEscrowParams({
            seller: seller,
            token: address(usdc),
            amount: 0,
            deadline: block.timestamp + 1 days
        });

        vm.prank(address(smartAccount));
        vm.expectRevert("Executor: zero amount");
        executor.executeBatchedEscrow(address(hub), params);
    }

    // ---- Session key authorization ----

    function test_authorizeSessionKey() public {
        executor.authorizeSessionKey(sessionKey());
        assertTrue(executor.isAuthorized(address(this), sessionKey()));
    }

    function test_deauthorizeSessionKey() public {
        executor.authorizeSessionKey(sessionKey());
        executor.deauthorizeSessionKey(sessionKey());
        assertFalse(executor.isAuthorized(address(this), sessionKey()));
    }

    function testRevert_authorizeSessionKey_zeroAddress() public {
        vm.expectRevert("Executor: zero key");
        executor.authorizeSessionKey(address(0));
    }

    // ---- isModuleType ----

    function test_isModuleType_executor() public view {
        assertTrue(executor.isModuleType(2));
        assertFalse(executor.isModuleType(1));
        assertFalse(executor.isModuleType(4));
    }

    // ---- onInstall / onUninstall ----

    function test_onInstall_authorizesKeys() public {
        address[] memory keys = new address[](2);
        keys[0] = address(0x1111);
        keys[1] = address(0x2222);

        executor.onInstall(abi.encode(keys));

        assertTrue(executor.isAuthorized(address(this), address(0x1111)));
        assertTrue(executor.isAuthorized(address(this), address(0x2222)));
    }

    function test_onUninstall_deauthorizesKeys() public {
        address[] memory keys = new address[](1);
        keys[0] = address(0x1111);

        executor.onInstall(abi.encode(keys));
        assertTrue(executor.isAuthorized(address(this), address(0x1111)));

        executor.onUninstall(abi.encode(keys));
        assertFalse(executor.isAuthorized(address(this), address(0x1111)));
    }

    // ---- Helpers ----

    function sessionKey() internal pure returns (address) {
        return address(0xBEEF);
    }
}
