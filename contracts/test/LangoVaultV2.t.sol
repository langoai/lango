// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import {UpgradeableBeacon} from "@openzeppelin/contracts/proxy/beacon/UpgradeableBeacon.sol";
import {BeaconProxy} from "@openzeppelin/contracts/proxy/beacon/BeaconProxy.sol";
import "../src/LangoVaultV2.sol";
import "../src/LangoBeaconVaultFactory.sol";
import "./mocks/MockUSDC.sol";

contract LangoVaultV2Test is Test {
    LangoVaultV2 public vaultImpl;
    UpgradeableBeacon public beacon;
    LangoBeaconVaultFactory public factory;
    MockUSDC public usdc;

    address public factoryOwner = address(0x1);
    address public buyer = address(0xB);
    address public seller = address(0xC);
    address public arbiter = address(0xA);
    address public stranger = address(0xD);

    uint256 public constant AMOUNT = 1000e6;
    bytes32 public constant REF_ID = keccak256("vault-ref-1");

    function setUp() public {
        usdc = new MockUSDC();

        // Deploy implementation
        vaultImpl = new LangoVaultV2();

        // Deploy beacon — owned by address(this) temporarily, will transfer to factory
        beacon = new UpgradeableBeacon(address(vaultImpl), address(this));

        // Deploy factory
        factory = new LangoBeaconVaultFactory(address(beacon), factoryOwner);

        // Transfer beacon ownership to factory so upgradeImplementation works
        beacon.transferOwnership(address(factory));

        // Fund buyer
        usdc.mint(buyer, 100_000e6);
    }

    // ---- Beacon + Factory Deployment ----

    function test_beacon_pointsToImplementation() public view {
        assertEq(beacon.implementation(), address(vaultImpl));
    }

    function test_factory_storesBeacon() public view {
        assertEq(address(factory.beacon()), address(beacon));
    }

    function test_factory_ownerIsCorrect() public view {
        assertEq(factory.owner(), factoryOwner);
    }

    // ---- Vault Creation via Factory ----

    function test_createVault_success() public {
        vm.prank(buyer);
        address vault = factory.createVault(seller, address(usdc), AMOUNT, arbiter, REF_ID);

        assertTrue(vault != address(0));
        assertEq(factory.vaultCount(), 1);
        assertEq(factory.getVault(0), vault);

        LangoVaultV2 v = LangoVaultV2(vault);
        assertEq(v.buyer(), buyer);
        assertEq(v.seller(), seller);
        assertEq(v.token(), address(usdc));
        assertEq(v.amount(), AMOUNT);
        assertEq(v.arbiter(), arbiter);
        assertEq(v.refId(), REF_ID);
        assertEq(uint8(v.status()), uint8(LangoVaultV2.VaultStatus.Created));
    }

    function test_createVault_emitsEvent() public {
        vm.prank(buyer);
        vm.expectEmit(false, true, true, true); // vault address unknown before creation
        emit LangoBeaconVaultFactory.VaultCreated(address(0), REF_ID, buyer, seller);
        factory.createVault(seller, address(usdc), AMOUNT, arbiter, REF_ID);
    }

    function test_createVault_multipleVaults() public {
        vm.startPrank(buyer);
        address v1 = factory.createVault(seller, address(usdc), AMOUNT, arbiter, REF_ID);
        address v2 = factory.createVault(seller, address(usdc), AMOUNT, arbiter, keccak256("ref-2"));
        vm.stopPrank();

        assertTrue(v1 != v2);
        assertEq(factory.vaultCount(), 2);
        assertEq(factory.getVault(0), v1);
        assertEq(factory.getVault(1), v2);
    }

    // ---- Vault Initialization Validation ----

    function testRevert_initialize_zeroRefId() public {
        vm.prank(buyer);
        vm.expectRevert("VaultV2: zero refId");
        factory.createVault(seller, address(usdc), AMOUNT, arbiter, bytes32(0));
    }

    function testRevert_initialize_zeroSeller() public {
        vm.prank(buyer);
        vm.expectRevert("VaultV2: zero seller");
        factory.createVault(address(0), address(usdc), AMOUNT, arbiter, REF_ID);
    }

    function test_implementation_cannotBeInitialized() public {
        vm.expectRevert();
        vaultImpl.initialize(buyer, seller, address(usdc), AMOUNT, arbiter, REF_ID);
    }

    // ---- Full Vault Lifecycle ----

    function test_fullLifecycle_depositWorkRelease() public {
        address vault = _createVault();
        LangoVaultV2 v = LangoVaultV2(vault);

        // Approve and deposit
        vm.prank(buyer);
        usdc.approve(vault, type(uint256).max);

        vm.prank(buyer);
        v.deposit();
        assertEq(uint8(v.status()), uint8(LangoVaultV2.VaultStatus.Deposited));
        assertEq(usdc.balanceOf(vault), AMOUNT);

        // Submit work
        bytes32 wh = keccak256("work-proof");
        vm.prank(seller);
        v.submitWork(wh);
        assertEq(uint8(v.status()), uint8(LangoVaultV2.VaultStatus.WorkSubmitted));
        assertEq(v.workHash(), wh);

        // Release
        vm.prank(buyer);
        v.release();
        assertEq(uint8(v.status()), uint8(LangoVaultV2.VaultStatus.Released));
        assertEq(usdc.balanceOf(seller), AMOUNT);
        assertEq(usdc.balanceOf(vault), 0);
    }

    // ---- Deposit ----

    function test_deposit_emitsEvent() public {
        address vault = _createVault();
        LangoVaultV2 v = LangoVaultV2(vault);

        vm.prank(buyer);
        usdc.approve(vault, type(uint256).max);

        vm.prank(buyer);
        vm.expectEmit(true, true, false, true);
        emit LangoVaultV2.Deposited(REF_ID, buyer, AMOUNT);
        v.deposit();
    }

    function testRevert_deposit_notBuyer() public {
        address vault = _createVault();
        LangoVaultV2 v = LangoVaultV2(vault);

        vm.prank(stranger);
        vm.expectRevert("VaultV2: not buyer");
        v.deposit();
    }

    // ---- Release ----

    function test_release_afterDeposit() public {
        address vault = _createAndDeposit();
        LangoVaultV2 v = LangoVaultV2(vault);

        vm.prank(buyer);
        v.release();

        assertEq(uint8(v.status()), uint8(LangoVaultV2.VaultStatus.Released));
        assertEq(usdc.balanceOf(seller), AMOUNT);
    }

    function test_release_emitsEvent() public {
        address vault = _createAndDeposit();
        LangoVaultV2 v = LangoVaultV2(vault);

        vm.prank(buyer);
        vm.expectEmit(true, true, false, true);
        emit LangoVaultV2.Released(REF_ID, seller, AMOUNT);
        v.release();
    }

    // ---- Refund ----

    function test_refund_afterDeadline() public {
        address vault = _createAndDeposit();
        LangoVaultV2 v = LangoVaultV2(vault);

        vm.warp(block.timestamp + 31 days);

        vm.prank(buyer);
        v.refund();

        assertEq(uint8(v.status()), uint8(LangoVaultV2.VaultStatus.Refunded));
        assertEq(usdc.balanceOf(buyer), 100_000e6);
    }

    function testRevert_refund_deadlineNotPassed() public {
        address vault = _createAndDeposit();
        LangoVaultV2 v = LangoVaultV2(vault);

        vm.prank(buyer);
        vm.expectRevert("VaultV2: deadline not passed");
        v.refund();
    }

    // ---- Dispute + Resolve ----

    function test_dispute_byBuyer() public {
        address vault = _createAndDeposit();
        LangoVaultV2 v = LangoVaultV2(vault);

        vm.prank(buyer);
        v.dispute();

        assertEq(uint8(v.status()), uint8(LangoVaultV2.VaultStatus.Disputed));
    }

    function test_dispute_bySeller() public {
        address vault = _createAndDeposit();
        LangoVaultV2 v = LangoVaultV2(vault);

        vm.prank(seller);
        v.dispute();

        assertEq(uint8(v.status()), uint8(LangoVaultV2.VaultStatus.Disputed));
    }

    function test_dispute_emitsEvent() public {
        address vault = _createAndDeposit();
        LangoVaultV2 v = LangoVaultV2(vault);

        vm.prank(buyer);
        vm.expectEmit(true, true, false, false);
        emit LangoVaultV2.Disputed(REF_ID, buyer);
        v.dispute();
    }

    function testRevert_dispute_notParty() public {
        address vault = _createAndDeposit();
        LangoVaultV2 v = LangoVaultV2(vault);

        vm.prank(stranger);
        vm.expectRevert("VaultV2: not party");
        v.dispute();
    }

    function test_resolve_fullSeller() public {
        address vault = _createDepositAndDispute();
        LangoVaultV2 v = LangoVaultV2(vault);

        vm.prank(arbiter);
        v.resolve(AMOUNT, 0);

        assertEq(uint8(v.status()), uint8(LangoVaultV2.VaultStatus.Resolved));
        assertEq(usdc.balanceOf(seller), AMOUNT);
    }

    function test_resolve_split() public {
        address vault = _createDepositAndDispute();
        LangoVaultV2 v = LangoVaultV2(vault);

        vm.prank(arbiter);
        v.resolve(600e6, 400e6);

        assertEq(usdc.balanceOf(seller), 600e6);
        assertEq(usdc.balanceOf(buyer), 100_000e6 - AMOUNT + 400e6);
    }

    function test_resolve_emitsEvent() public {
        address vault = _createDepositAndDispute();
        LangoVaultV2 v = LangoVaultV2(vault);

        vm.prank(arbiter);
        vm.expectEmit(true, false, false, true);
        emit LangoVaultV2.VaultResolved(REF_ID, AMOUNT, 0);
        v.resolve(AMOUNT, 0);
    }

    function testRevert_resolve_notArbiter() public {
        address vault = _createDepositAndDispute();
        LangoVaultV2 v = LangoVaultV2(vault);

        vm.prank(buyer);
        vm.expectRevert("VaultV2: not arbiter");
        v.resolve(AMOUNT, 0);
    }

    function testRevert_resolve_amountsMismatch() public {
        address vault = _createDepositAndDispute();
        LangoVaultV2 v = LangoVaultV2(vault);

        vm.prank(arbiter);
        vm.expectRevert("VaultV2: amounts mismatch");
        v.resolve(AMOUNT, 1);
    }

    // ---- refId in events ----

    function test_refId_storedCorrectly() public {
        address vault = _createVault();
        LangoVaultV2 v = LangoVaultV2(vault);
        assertEq(v.refId(), REF_ID);
    }

    // ---- Beacon Upgrade ----

    function test_beaconUpgrade_allVaultsPointToNewImpl() public {
        // Create two vaults
        vm.prank(buyer);
        address v1 = factory.createVault(seller, address(usdc), AMOUNT, arbiter, REF_ID);
        vm.prank(buyer);
        address v2 = factory.createVault(seller, address(usdc), AMOUNT, arbiter, keccak256("ref-2"));

        // Deploy new implementation
        LangoVaultV2 newImpl = new LangoVaultV2();

        // Verify both vaults still work before upgrade
        assertEq(LangoVaultV2(v1).buyer(), buyer);
        assertEq(LangoVaultV2(v2).buyer(), buyer);

        // Upgrade via factory (which owns the beacon)
        vm.prank(factoryOwner);
        factory.upgradeImplementation(address(newImpl));

        assertEq(beacon.implementation(), address(newImpl));

        // Both vaults still work after upgrade
        assertEq(LangoVaultV2(v1).buyer(), buyer);
        assertEq(LangoVaultV2(v2).buyer(), buyer);
        assertEq(LangoVaultV2(v1).refId(), REF_ID);
    }

    function test_factoryUpgrade_callsBeacon() public {
        LangoVaultV2 newImpl = new LangoVaultV2();

        vm.prank(factoryOwner);
        factory.upgradeImplementation(address(newImpl));

        assertEq(beacon.implementation(), address(newImpl));
    }

    function testRevert_beaconUpgrade_notOwner() public {
        LangoVaultV2 newImpl = new LangoVaultV2();

        // Direct beacon upgrade should fail (owned by factory)
        vm.prank(factoryOwner);
        vm.expectRevert();
        beacon.upgradeTo(address(newImpl));
    }

    function testRevert_factoryUpgrade_notOwner() public {
        LangoVaultV2 newImpl = new LangoVaultV2();

        vm.prank(stranger);
        vm.expectRevert();
        factory.upgradeImplementation(address(newImpl));
    }

    // ---- Helpers ----

    function _createVault() internal returns (address) {
        vm.prank(buyer);
        return factory.createVault(seller, address(usdc), AMOUNT, arbiter, REF_ID);
    }

    function _createAndDeposit() internal returns (address vault) {
        vault = _createVault();

        vm.prank(buyer);
        usdc.approve(vault, type(uint256).max);

        vm.prank(buyer);
        LangoVaultV2(vault).deposit();
    }

    function _createDepositAndDispute() internal returns (address vault) {
        vault = _createAndDeposit();

        vm.prank(buyer);
        LangoVaultV2(vault).dispute();
    }
}
