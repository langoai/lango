// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/LangoVault.sol";
import "./mocks/MockUSDC.sol";

contract LangoVaultTest is Test {
    LangoVault public vault;
    MockUSDC public usdc;

    address public buyer = address(0xB);
    address public seller = address(0xC);
    address public arbitrator = address(0xA);
    address public stranger = address(0xD);

    uint256 public constant AMOUNT = 500e6;
    uint256 public vaultDeadline;

    function setUp() public {
        usdc = new MockUSDC();
        vault = new LangoVault();

        usdc.mint(buyer, 10_000e6);
        vaultDeadline = block.timestamp + 1 days;

        vault.initialize(buyer, seller, address(usdc), AMOUNT, vaultDeadline, arbitrator);

        vm.prank(buyer);
        usdc.approve(address(vault), type(uint256).max);
    }

    // ---- initialize ----

    function test_initialize_setsFields() public view {
        assertEq(vault.buyer(), buyer);
        assertEq(vault.seller(), seller);
        assertEq(vault.token(), address(usdc));
        assertEq(vault.amount(), AMOUNT);
        assertEq(vault.deadline(), vaultDeadline);
        assertEq(vault.arbitrator(), arbitrator);
        assertEq(uint8(vault.status()), uint8(LangoVault.VaultStatus.Created));
    }

    function testRevert_initialize_doubleInit() public {
        vm.expectRevert("Vault: already initialized");
        vault.initialize(buyer, seller, address(usdc), AMOUNT, vaultDeadline, arbitrator);
    }

    function testRevert_initialize_zeroBuyer() public {
        LangoVault v = new LangoVault();
        vm.expectRevert("Vault: zero buyer");
        v.initialize(address(0), seller, address(usdc), AMOUNT, vaultDeadline, arbitrator);
    }

    function testRevert_initialize_zeroSeller() public {
        LangoVault v = new LangoVault();
        vm.expectRevert("Vault: zero seller");
        v.initialize(buyer, address(0), address(usdc), AMOUNT, vaultDeadline, arbitrator);
    }

    function testRevert_initialize_zeroToken() public {
        LangoVault v = new LangoVault();
        vm.expectRevert("Vault: zero token");
        v.initialize(buyer, seller, address(0), AMOUNT, vaultDeadline, arbitrator);
    }

    function testRevert_initialize_zeroAmount() public {
        LangoVault v = new LangoVault();
        vm.expectRevert("Vault: zero amount");
        v.initialize(buyer, seller, address(usdc), 0, vaultDeadline, arbitrator);
    }

    function testRevert_initialize_pastDeadline() public {
        LangoVault v = new LangoVault();
        vm.expectRevert("Vault: past deadline");
        v.initialize(buyer, seller, address(usdc), AMOUNT, block.timestamp, arbitrator);
    }

    function testRevert_initialize_zeroArbitrator() public {
        LangoVault v = new LangoVault();
        vm.expectRevert("Vault: zero arbitrator");
        v.initialize(buyer, seller, address(usdc), AMOUNT, vaultDeadline, address(0));
    }

    // ---- deposit ----

    function test_deposit_success() public {
        vm.prank(buyer);
        vault.deposit();

        assertEq(uint8(vault.status()), uint8(LangoVault.VaultStatus.Deposited));
        assertEq(usdc.balanceOf(address(vault)), AMOUNT);
    }

    function testRevert_deposit_notBuyer() public {
        vm.prank(stranger);
        vm.expectRevert("Vault: not buyer");
        vault.deposit();
    }

    function testRevert_deposit_notCreated() public {
        vm.prank(buyer);
        vault.deposit();

        vm.prank(buyer);
        vm.expectRevert("Vault: not created");
        vault.deposit();
    }

    // ---- submitWork ----

    function test_submitWork_success() public {
        _deposit();
        bytes32 wh = keccak256("work");

        vm.prank(seller);
        vault.submitWork(wh);

        assertEq(uint8(vault.status()), uint8(LangoVault.VaultStatus.WorkSubmitted));
        assertEq(vault.workHash(), wh);
    }

    function testRevert_submitWork_notSeller() public {
        _deposit();
        vm.prank(buyer);
        vm.expectRevert("Vault: not seller");
        vault.submitWork(keccak256("x"));
    }

    function testRevert_submitWork_emptyHash() public {
        _deposit();
        vm.prank(seller);
        vm.expectRevert("Vault: empty hash");
        vault.submitWork(bytes32(0));
    }

    // ---- release ----

    function test_release_success() public {
        _deposit();

        vm.prank(buyer);
        vault.release();

        assertEq(uint8(vault.status()), uint8(LangoVault.VaultStatus.Released));
        assertEq(usdc.balanceOf(seller), AMOUNT);
    }

    function testRevert_release_notBuyer() public {
        _deposit();
        vm.prank(stranger);
        vm.expectRevert("Vault: not buyer");
        vault.release();
    }

    // ---- refund ----

    function test_refund_afterDeadline() public {
        _deposit();
        vm.warp(vaultDeadline + 1);

        vm.prank(buyer);
        vault.refund();

        assertEq(uint8(vault.status()), uint8(LangoVault.VaultStatus.Refunded));
        assertEq(usdc.balanceOf(buyer), 10_000e6);
    }

    function testRevert_refund_deadlineNotPassed() public {
        _deposit();

        vm.prank(buyer);
        vm.expectRevert("Vault: deadline not passed");
        vault.refund();
    }

    // ---- dispute ----

    function test_dispute_byBuyer() public {
        _deposit();
        vm.prank(buyer);
        vault.dispute();
        assertEq(uint8(vault.status()), uint8(LangoVault.VaultStatus.Disputed));
    }

    function test_dispute_bySeller() public {
        _deposit();
        vm.prank(seller);
        vault.dispute();
        assertEq(uint8(vault.status()), uint8(LangoVault.VaultStatus.Disputed));
    }

    function testRevert_dispute_notParty() public {
        _deposit();
        vm.prank(stranger);
        vm.expectRevert("Vault: not party");
        vault.dispute();
    }

    // ---- resolve ----

    function test_resolve_success() public {
        _depositAndDispute();

        vm.prank(arbitrator);
        vault.resolve(true, AMOUNT, 0);

        assertEq(uint8(vault.status()), uint8(LangoVault.VaultStatus.Resolved));
        assertEq(usdc.balanceOf(seller), AMOUNT);
    }

    function test_resolve_split() public {
        _depositAndDispute();

        vm.prank(arbitrator);
        vault.resolve(false, 200e6, 300e6);

        assertEq(usdc.balanceOf(seller), 200e6);
        assertEq(usdc.balanceOf(buyer), 10_000e6 - AMOUNT + 300e6);
    }

    function testRevert_resolve_notArbitrator() public {
        _depositAndDispute();
        vm.prank(buyer);
        vm.expectRevert("Vault: not arbitrator");
        vault.resolve(true, AMOUNT, 0);
    }

    function testRevert_resolve_amountsMismatch() public {
        _depositAndDispute();
        vm.prank(arbitrator);
        vm.expectRevert("Vault: amounts mismatch");
        vault.resolve(true, AMOUNT, 1);
    }

    // ---- full lifecycle ----

    function test_fullLifecycle() public {
        vm.prank(buyer);
        vault.deposit();

        vm.prank(seller);
        vault.submitWork(keccak256("result"));

        vm.prank(buyer);
        vault.release();

        assertEq(uint8(vault.status()), uint8(LangoVault.VaultStatus.Released));
        assertEq(usdc.balanceOf(seller), AMOUNT);
        assertEq(usdc.balanceOf(address(vault)), 0);
    }

    // ---- helpers ----

    function _deposit() internal {
        vm.prank(buyer);
        vault.deposit();
    }

    function _depositAndDispute() internal {
        _deposit();
        vm.prank(buyer);
        vault.dispute();
    }
}
