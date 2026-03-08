// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/modules/LangoSpendingHook.sol";

contract LangoSpendingHookTest is Test {
    LangoSpendingHook public hook;

    address public account;
    address public sessionKey = address(0xBEEF);

    uint256 public constant PER_TX = 1 ether;
    uint256 public constant DAILY = 5 ether;
    uint256 public constant CUMULATIVE = 20 ether;

    function setUp() public {
        hook = new LangoSpendingHook();
        account = address(this);

        // Set limits via the account (msg.sender)
        hook.setLimits(PER_TX, DAILY, CUMULATIVE);
    }

    // ---- Per-Tx Limit ----

    function test_preCheck_withinPerTxLimit() public {
        hook.preCheck(sessionKey, 0.5 ether, "");
        // No revert means success
    }

    function testRevert_preCheck_exceedsPerTxLimit() public {
        vm.expectRevert("Hook: exceeds per-tx limit");
        hook.preCheck(sessionKey, 1.5 ether, "");
    }

    function test_preCheck_exactPerTxLimit() public {
        hook.preCheck(sessionKey, PER_TX, "");
        // Exact limit should pass
    }

    // ---- Daily Limit ----

    function test_preCheck_withinDailyLimit() public {
        // 5 calls of 1 ether each = 5 ether = exact daily limit
        for (uint256 i = 0; i < 5; i++) {
            hook.preCheck(sessionKey, PER_TX, "");
        }
    }

    function testRevert_preCheck_exceedsDailyLimit() public {
        // First 5 calls succeed (5 ether)
        for (uint256 i = 0; i < 5; i++) {
            hook.preCheck(sessionKey, PER_TX, "");
        }

        // 6th call should fail (daily limit exceeded)
        vm.expectRevert("Hook: exceeds daily limit");
        hook.preCheck(sessionKey, PER_TX, "");
    }

    function test_preCheck_dailyLimitResetsAfterDay() public {
        // Spend to daily limit
        for (uint256 i = 0; i < 5; i++) {
            hook.preCheck(sessionKey, PER_TX, "");
        }

        // Warp forward 1 day
        vm.warp(block.timestamp + 86401);

        // Should work again after daily reset
        hook.preCheck(sessionKey, PER_TX, "");
    }

    // ---- Cumulative Limit ----

    function test_preCheck_withinCumulativeLimit() public {
        // Spend across multiple days within cumulative limit
        for (uint256 day = 0; day < 4; day++) {
            for (uint256 i = 0; i < 5; i++) {
                hook.preCheck(sessionKey, PER_TX, "");
            }
            vm.warp(block.timestamp + 86401);
        }
        // Total: 20 ether = exact cumulative limit
    }

    function testRevert_preCheck_exceedsCumulativeLimit() public {
        // Spend across multiple days to hit cumulative limit
        for (uint256 day = 0; day < 4; day++) {
            for (uint256 i = 0; i < 5; i++) {
                hook.preCheck(sessionKey, PER_TX, "");
            }
            vm.warp(block.timestamp + 86401);
        }

        // Next spend should fail (cumulative limit exceeded)
        vm.expectRevert("Hook: exceeds cumulative limit");
        hook.preCheck(sessionKey, PER_TX, "");
    }

    // ---- setLimits ----

    function test_setLimits_updatesConfig() public {
        hook.setLimits(2 ether, 10 ether, 50 ether);

        LangoSpendingHook.SpendingConfig memory cfg = hook.getConfig(account);
        assertEq(cfg.perTxLimit, 2 ether);
        assertEq(cfg.dailyLimit, 10 ether);
        assertEq(cfg.cumulativeLimit, 50 ether);
        assertTrue(cfg.configured);
    }

    function test_setLimits_emitsEvent() public {
        vm.expectEmit(true, false, false, true);
        emit LangoSpendingHook.LimitsUpdated(account, 2 ether, 10 ether, 50 ether);
        hook.setLimits(2 ether, 10 ether, 50 ether);
    }

    function test_setLimits_allowsHigherPerTxAfterUpdate() public {
        // Initially per-tx is 1 ether
        vm.expectRevert("Hook: exceeds per-tx limit");
        hook.preCheck(sessionKey, 1.5 ether, "");

        // Update to 2 ether per-tx
        hook.setLimits(2 ether, DAILY, CUMULATIVE);

        // Now 1.5 ether should pass
        hook.preCheck(sessionKey, 1.5 ether, "");
    }

    // ---- onInstall / onUninstall ----

    function test_onInstall_setsConfig() public {
        LangoSpendingHook freshHook = new LangoSpendingHook();
        bytes memory data = abi.encode(uint256(0.5 ether), uint256(3 ether), uint256(10 ether));

        freshHook.onInstall(data);

        LangoSpendingHook.SpendingConfig memory cfg = freshHook.getConfig(address(this));
        assertEq(cfg.perTxLimit, 0.5 ether);
        assertEq(cfg.dailyLimit, 3 ether);
        assertEq(cfg.cumulativeLimit, 10 ether);
        assertTrue(cfg.configured);
    }

    function test_onUninstall_clearsConfig() public {
        hook.onUninstall("");

        LangoSpendingHook.SpendingConfig memory cfg = hook.getConfig(account);
        assertFalse(cfg.configured);
    }

    // ---- isModuleType ----

    function test_isModuleType_hook() public view {
        assertTrue(hook.isModuleType(4));
        assertFalse(hook.isModuleType(1));
        assertFalse(hook.isModuleType(2));
    }

    // ---- postCheck ----

    function test_postCheck_noOp() public view {
        // postCheck is pure and does nothing — verify the selector exists
        bytes4 selector = hook.postCheck.selector;
        assertTrue(selector != bytes4(0));
    }

    // ---- Unconfigured account ----

    function test_preCheck_unconfiguredAccountPassesThrough() public {
        LangoSpendingHook freshHook = new LangoSpendingHook();
        // Should not revert — no config means no limits
        freshHook.preCheck(sessionKey, 100 ether, "");
    }

    // ---- getSpendState ----

    function test_getSpendState_tracksCorrectly() public {
        hook.preCheck(sessionKey, 0.5 ether, "");

        LangoSpendingHook.SpendState memory state = hook.getSpendState(account, sessionKey);
        assertEq(state.dailySpent, 0.5 ether);
        assertEq(state.cumulativeSpent, 0.5 ether);
    }
}
