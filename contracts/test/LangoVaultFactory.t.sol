// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/LangoVault.sol";
import "../src/LangoVaultFactory.sol";
import "./mocks/MockUSDC.sol";

contract LangoVaultFactoryTest is Test {
    LangoVaultFactory public factory;
    LangoVault public impl;
    MockUSDC public usdc;

    address public buyer = address(0xB);
    address public seller = address(0xC);
    address public arbitrator = address(0xA);

    uint256 public constant AMOUNT = 500e6;
    uint256 public factoryDeadline;

    function setUp() public {
        usdc = new MockUSDC();
        impl = new LangoVault();
        factory = new LangoVaultFactory(address(impl));

        usdc.mint(buyer, 10_000e6);
        factoryDeadline = block.timestamp + 1 days;
    }

    // ---- constructor ----

    function test_constructor_setsImplementation() public view {
        assertEq(factory.implementation(), address(impl));
    }

    function testRevert_constructor_zeroImpl() public {
        vm.expectRevert("Factory: zero implementation");
        new LangoVaultFactory(address(0));
    }

    // ---- createVault ----

    function test_createVault_success() public {
        vm.prank(buyer);
        (uint256 vaultId, address vaultAddr) = factory.createVault(
            seller, address(usdc), AMOUNT, factoryDeadline, arbitrator
        );

        assertEq(vaultId, 0);
        assertTrue(vaultAddr != address(0));
        assertEq(factory.vaultCount(), 1);
    }

    function test_createVault_cloneIsUsable() public {
        vm.prank(buyer);
        (, address vaultAddr) = factory.createVault(
            seller, address(usdc), AMOUNT, factoryDeadline, arbitrator
        );

        LangoVault v = LangoVault(vaultAddr);
        assertEq(v.buyer(), buyer);
        assertEq(v.seller(), seller);
        assertEq(v.token(), address(usdc));
        assertEq(v.amount(), AMOUNT);
        assertEq(uint8(v.status()), uint8(LangoVault.VaultStatus.Created));

        // Deposit should work on the clone.
        vm.prank(buyer);
        usdc.approve(vaultAddr, AMOUNT);
        vm.prank(buyer);
        v.deposit();
        assertEq(uint8(v.status()), uint8(LangoVault.VaultStatus.Deposited));
    }

    function test_createVault_multiple() public {
        vm.startPrank(buyer);
        (uint256 id0,) = factory.createVault(seller, address(usdc), AMOUNT, factoryDeadline, arbitrator);
        (uint256 id1,) = factory.createVault(seller, address(usdc), AMOUNT, factoryDeadline, arbitrator);
        (uint256 id2,) = factory.createVault(seller, address(usdc), AMOUNT, factoryDeadline, arbitrator);
        vm.stopPrank();

        assertEq(id0, 0);
        assertEq(id1, 1);
        assertEq(id2, 2);
        assertEq(factory.vaultCount(), 3);
    }

    function test_createVault_emitsEvent() public {
        vm.prank(buyer);
        vm.expectEmit(true, false, true, false);
        emit LangoVaultFactory.VaultCreated(0, address(0), buyer, seller);
        factory.createVault(seller, address(usdc), AMOUNT, factoryDeadline, arbitrator);
    }

    // ---- getVault ----

    function test_getVault_returnsCorrectAddress() public {
        vm.prank(buyer);
        (, address vaultAddr) = factory.createVault(
            seller, address(usdc), AMOUNT, factoryDeadline, arbitrator
        );

        assertEq(factory.getVault(0), vaultAddr);
    }

    function test_getVault_unknownId_returnsZero() public view {
        assertEq(factory.getVault(999), address(0));
    }

    // ---- vaultCount ----

    function test_vaultCount_startsAtZero() public view {
        assertEq(factory.vaultCount(), 0);
    }
}
