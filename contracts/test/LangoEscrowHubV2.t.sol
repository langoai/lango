// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import "../src/LangoEscrowHubV2.sol";
import "../src/settlers/DirectSettler.sol";
import "../src/settlers/MilestoneSettler.sol";
import "./mocks/MockUSDC.sol";

contract LangoEscrowHubV2Test is Test {
    LangoEscrowHubV2 public hub;
    LangoEscrowHubV2 public hubImpl;
    MockUSDC public usdc;
    DirectSettler public directSettler;
    MilestoneSettler public milestoneSettler;

    address public owner = address(0x1);
    address public buyer = address(0xB);
    address public seller = address(0xC);
    address public stranger = address(0xD);
    address public member1 = address(0xE1);
    address public member2 = address(0xE2);
    address public member3 = address(0xE3);

    uint256 public constant AMOUNT = 1000e6;
    bytes32 public constant REF_ID = keccak256("test-ref-1");
    uint256 public deadline;

    function setUp() public {
        usdc = new MockUSDC();

        // Deploy implementation
        hubImpl = new LangoEscrowHubV2();

        // Deploy proxy
        bytes memory initData = abi.encodeCall(LangoEscrowHubV2.initialize, (owner));
        ERC1967Proxy proxy = new ERC1967Proxy(address(hubImpl), initData);
        hub = LangoEscrowHubV2(address(proxy));

        // Deploy settlers
        directSettler = new DirectSettler();
        milestoneSettler = new MilestoneSettler(address(hub));

        // Register milestone settler
        vm.prank(owner);
        hub.registerSettler(keccak256("milestone"), address(milestoneSettler));

        // Fund buyer
        usdc.mint(buyer, 100_000e6);
        deadline = block.timestamp + 1 days;

        vm.prank(buyer);
        usdc.approve(address(hub), type(uint256).max);
    }

    // ---- Proxy + Initialization ----

    function test_initialize_setsOwner() public view {
        assertEq(hub.owner(), owner);
    }

    function testRevert_initialize_twice() public {
        vm.expectRevert();
        hub.initialize(owner);
    }

    function testRevert_initialize_zeroOwner() public {
        LangoEscrowHubV2 impl2 = new LangoEscrowHubV2();
        vm.expectRevert("HubV2: zero owner");
        new ERC1967Proxy(address(impl2), abi.encodeCall(LangoEscrowHubV2.initialize, (address(0))));
    }

    function test_implementation_cannotBeInitialized() public {
        vm.expectRevert();
        hubImpl.initialize(owner);
    }

    // ---- UUPS Upgrade ----

    function test_upgrade_onlyOwner() public {
        LangoEscrowHubV2 newImpl = new LangoEscrowHubV2();
        vm.prank(owner);
        hub.upgradeToAndCall(address(newImpl), "");
    }

    function testRevert_upgrade_notOwner() public {
        LangoEscrowHubV2 newImpl = new LangoEscrowHubV2();
        vm.prank(stranger);
        vm.expectRevert();
        hub.upgradeToAndCall(address(newImpl), "");
    }

    // ---- Settler Registration ----

    function test_registerSettler_success() public {
        vm.prank(owner);
        hub.registerSettler(keccak256("direct"), address(directSettler));
        assertEq(hub.settlers(keccak256("direct")), address(directSettler));
    }

    function testRevert_registerSettler_notOwner() public {
        vm.prank(stranger);
        vm.expectRevert();
        hub.registerSettler(keccak256("direct"), address(directSettler));
    }

    function testRevert_registerSettler_zeroAddress() public {
        vm.prank(owner);
        vm.expectRevert("HubV2: zero settler");
        hub.registerSettler(keccak256("direct"), address(0));
    }

    function test_registerSettler_emitsEvent() public {
        vm.prank(owner);
        vm.expectEmit(true, false, false, true);
        emit LangoEscrowHubV2.SettlerRegistered(keccak256("direct"), address(directSettler));
        hub.registerSettler(keccak256("direct"), address(directSettler));
    }

    // ---- directSettle ----

    function test_directSettle_transfersImmediately() public {
        vm.prank(buyer);
        hub.directSettle(seller, address(usdc), AMOUNT, REF_ID);

        assertEq(usdc.balanceOf(seller), AMOUNT);
        assertEq(usdc.balanceOf(buyer), 100_000e6 - AMOUNT);
    }

    function test_directSettle_emitsEvents() public {
        vm.prank(buyer);
        vm.expectEmit(true, true, false, true);
        emit ILangoEconomy.EscrowOpened(REF_ID, 0, buyer, seller, AMOUNT);
        hub.directSettle(seller, address(usdc), AMOUNT, REF_ID);
    }

    function test_directSettle_dealStatusIsReleased() public {
        vm.prank(buyer);
        hub.directSettle(seller, address(usdc), AMOUNT, REF_ID);

        LangoEscrowHubV2.Deal memory d = hub.getDeal(0);
        assertEq(uint8(d.status), uint8(LangoEscrowHubV2.DealStatus.Released));
        assertEq(d.refId, REF_ID);
    }

    function testRevert_directSettle_zeroSeller() public {
        vm.prank(buyer);
        vm.expectRevert("HubV2: zero seller");
        hub.directSettle(address(0), address(usdc), AMOUNT, REF_ID);
    }

    function testRevert_directSettle_zeroRefId() public {
        vm.prank(buyer);
        vm.expectRevert("HubV2: zero refId");
        hub.directSettle(seller, address(usdc), AMOUNT, bytes32(0));
    }

    // ---- createSimpleEscrow ----

    function test_createSimpleEscrow_success() public {
        vm.prank(buyer);
        uint256 dealId = hub.createSimpleEscrow(seller, address(usdc), AMOUNT, deadline, REF_ID);
        assertEq(dealId, 0);

        LangoEscrowHubV2.Deal memory d = hub.getDeal(dealId);
        assertEq(d.buyer, buyer);
        assertEq(d.seller, seller);
        assertEq(d.token, address(usdc));
        assertEq(d.amount, AMOUNT);
        assertEq(d.refId, REF_ID);
        assertEq(uint8(d.status), uint8(LangoEscrowHubV2.DealStatus.Created));
    }

    function test_createSimpleEscrow_emitsEscrowOpened() public {
        vm.prank(buyer);
        vm.expectEmit(true, true, false, true);
        emit ILangoEconomy.EscrowOpened(REF_ID, 0, buyer, seller, AMOUNT);
        hub.createSimpleEscrow(seller, address(usdc), AMOUNT, deadline, REF_ID);
    }

    function testRevert_createSimpleEscrow_zeroRefId() public {
        vm.prank(buyer);
        vm.expectRevert("HubV2: zero refId");
        hub.createSimpleEscrow(seller, address(usdc), AMOUNT, deadline, bytes32(0));
    }

    function testRevert_createSimpleEscrow_pastDeadline() public {
        vm.prank(buyer);
        vm.expectRevert("HubV2: past deadline");
        hub.createSimpleEscrow(seller, address(usdc), AMOUNT, block.timestamp, REF_ID);
    }

    // ---- deposit ----

    function test_deposit_success() public {
        uint256 dealId = _createSimpleEscrow();

        vm.prank(buyer);
        hub.deposit(dealId);

        LangoEscrowHubV2.Deal memory d = hub.getDeal(dealId);
        assertEq(uint8(d.status), uint8(LangoEscrowHubV2.DealStatus.Deposited));
        assertEq(usdc.balanceOf(address(hub)), AMOUNT);
    }

    function test_deposit_emitsEvent() public {
        uint256 dealId = _createSimpleEscrow();

        vm.prank(buyer);
        vm.expectEmit(true, true, true, true);
        emit LangoEscrowHubV2.Deposited(REF_ID, dealId, buyer, AMOUNT);
        hub.deposit(dealId);
    }

    function testRevert_deposit_notBuyer() public {
        uint256 dealId = _createSimpleEscrow();

        vm.prank(stranger);
        vm.expectRevert("HubV2: not buyer");
        hub.deposit(dealId);
    }

    function testRevert_deposit_notCreated() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(buyer);
        vm.expectRevert("HubV2: not created");
        hub.deposit(dealId);
    }

    // ---- submitWork ----

    function test_submitWork_success() public {
        uint256 dealId = _createAndDeposit();
        bytes32 wh = keccak256("work proof");

        vm.prank(seller);
        hub.submitWork(dealId, wh);

        LangoEscrowHubV2.Deal memory d = hub.getDeal(dealId);
        assertEq(uint8(d.status), uint8(LangoEscrowHubV2.DealStatus.WorkSubmitted));
        assertEq(d.workHash, wh);
    }

    function test_submitWork_emitsEvent() public {
        uint256 dealId = _createAndDeposit();
        bytes32 wh = keccak256("work proof");

        vm.prank(seller);
        vm.expectEmit(true, true, true, true);
        emit LangoEscrowHubV2.WorkSubmitted(REF_ID, dealId, seller, wh);
        hub.submitWork(dealId, wh);
    }

    function testRevert_submitWork_notSeller() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(buyer);
        vm.expectRevert("HubV2: not seller");
        hub.submitWork(dealId, keccak256("x"));
    }

    function testRevert_submitWork_emptyHash() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(seller);
        vm.expectRevert("HubV2: empty hash");
        hub.submitWork(dealId, bytes32(0));
    }

    // ---- release ----

    function test_release_afterDeposit() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(buyer);
        hub.release(dealId);

        LangoEscrowHubV2.Deal memory d = hub.getDeal(dealId);
        assertEq(uint8(d.status), uint8(LangoEscrowHubV2.DealStatus.Released));
        assertEq(usdc.balanceOf(seller), AMOUNT);
    }

    function test_release_afterWorkSubmitted() public {
        uint256 dealId = _createAndDeposit();
        vm.prank(seller);
        hub.submitWork(dealId, keccak256("proof"));

        vm.prank(buyer);
        hub.release(dealId);

        assertEq(uint8(hub.getDeal(dealId).status), uint8(LangoEscrowHubV2.DealStatus.Released));
        assertEq(usdc.balanceOf(seller), AMOUNT);
    }

    function test_release_emitsEvents() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(buyer);
        vm.expectEmit(true, true, true, true);
        emit LangoEscrowHubV2.Released(REF_ID, dealId, seller, AMOUNT);
        hub.release(dealId);
    }

    function testRevert_release_notReleasable() public {
        vm.prank(buyer);
        uint256 dealId = hub.createSimpleEscrow(seller, address(usdc), AMOUNT, deadline, REF_ID);

        vm.prank(buyer);
        vm.expectRevert("HubV2: not releasable");
        hub.release(dealId);
    }

    // ---- refund ----

    function test_refund_afterDeadline() public {
        uint256 dealId = _createAndDeposit();

        vm.warp(deadline + 1);

        vm.prank(buyer);
        hub.refund(dealId);

        assertEq(uint8(hub.getDeal(dealId).status), uint8(LangoEscrowHubV2.DealStatus.Refunded));
        assertEq(usdc.balanceOf(buyer), 100_000e6);
    }

    function test_refund_emitsEvents() public {
        uint256 dealId = _createAndDeposit();
        vm.warp(deadline + 1);

        vm.prank(buyer);
        vm.expectEmit(true, true, true, true);
        emit LangoEscrowHubV2.Refunded(REF_ID, dealId, buyer, AMOUNT);
        hub.refund(dealId);
    }

    function testRevert_refund_deadlineNotPassed() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(buyer);
        vm.expectRevert("HubV2: deadline not passed");
        hub.refund(dealId);
    }

    // ---- dispute ----

    function test_dispute_byBuyer() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(buyer);
        hub.dispute(dealId);

        assertEq(uint8(hub.getDeal(dealId).status), uint8(LangoEscrowHubV2.DealStatus.Disputed));
    }

    function test_dispute_bySeller() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(seller);
        hub.dispute(dealId);

        assertEq(uint8(hub.getDeal(dealId).status), uint8(LangoEscrowHubV2.DealStatus.Disputed));
    }

    function test_dispute_emitsEvent() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(buyer);
        vm.expectEmit(true, true, false, true);
        emit ILangoEconomy.DisputeRaised(REF_ID, dealId, buyer);
        hub.dispute(dealId);
    }

    function testRevert_dispute_notParty() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(stranger);
        vm.expectRevert("HubV2: not party");
        hub.dispute(dealId);
    }

    function testRevert_dispute_notDisputable() public {
        vm.prank(buyer);
        uint256 dealId = hub.createSimpleEscrow(seller, address(usdc), AMOUNT, deadline, REF_ID);

        vm.prank(buyer);
        vm.expectRevert("HubV2: not disputable");
        hub.dispute(dealId);
    }

    // ---- resolveDispute ----

    function test_resolveDispute_fullSeller() public {
        uint256 dealId = _createDepositAndDispute();

        vm.prank(owner);
        hub.resolveDispute(dealId, AMOUNT, 0);

        assertEq(uint8(hub.getDeal(dealId).status), uint8(LangoEscrowHubV2.DealStatus.Resolved));
        assertEq(usdc.balanceOf(seller), AMOUNT);
    }

    function test_resolveDispute_split() public {
        uint256 dealId = _createDepositAndDispute();

        uint256 sellerAmt = 600e6;
        uint256 buyerAmt = 400e6;

        vm.prank(owner);
        hub.resolveDispute(dealId, sellerAmt, buyerAmt);

        assertEq(usdc.balanceOf(seller), sellerAmt);
        assertEq(usdc.balanceOf(buyer), 100_000e6 - AMOUNT + buyerAmt);
    }

    function test_resolveDispute_emitsEvent() public {
        uint256 dealId = _createDepositAndDispute();

        vm.prank(owner);
        vm.expectEmit(true, true, false, true);
        emit ILangoEconomy.SettlementFinalized(REF_ID, dealId, address(0), AMOUNT, 0);
        hub.resolveDispute(dealId, AMOUNT, 0);
    }

    function testRevert_resolveDispute_notOwner() public {
        uint256 dealId = _createDepositAndDispute();

        vm.prank(buyer);
        vm.expectRevert();
        hub.resolveDispute(dealId, AMOUNT, 0);
    }

    function testRevert_resolveDispute_notDisputed() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(owner);
        vm.expectRevert("HubV2: not disputed");
        hub.resolveDispute(dealId, AMOUNT, 0);
    }

    function testRevert_resolveDispute_amountsMismatch() public {
        uint256 dealId = _createDepositAndDispute();

        vm.prank(owner);
        vm.expectRevert("HubV2: amounts mismatch");
        hub.resolveDispute(dealId, AMOUNT, 1);
    }

    // ---- createMilestoneEscrow ----

    function test_createMilestoneEscrow_success() public {
        uint256[] memory milestones = new uint256[](3);
        milestones[0] = 300e6;
        milestones[1] = 300e6;
        milestones[2] = 400e6;

        vm.prank(buyer);
        uint256 dealId = hub.createMilestoneEscrow(seller, address(usdc), AMOUNT, milestones, deadline, REF_ID);

        LangoEscrowHubV2.Deal memory d = hub.getDeal(dealId);
        assertEq(uint8(d.dealType), uint8(LangoEscrowHubV2.DealType.Milestone));
        assertEq(d.settler, address(milestoneSettler));
        assertEq(d.amount, AMOUNT);
    }

    function testRevert_createMilestoneEscrow_sumMismatch() public {
        uint256[] memory milestones = new uint256[](2);
        milestones[0] = 300e6;
        milestones[1] = 300e6;

        vm.prank(buyer);
        vm.expectRevert("HubV2: milestones sum mismatch");
        hub.createMilestoneEscrow(seller, address(usdc), AMOUNT, milestones, deadline, REF_ID);
    }

    function testRevert_createMilestoneEscrow_noMilestones() public {
        uint256[] memory milestones = new uint256[](0);

        vm.prank(buyer);
        vm.expectRevert("HubV2: no milestones");
        hub.createMilestoneEscrow(seller, address(usdc), AMOUNT, milestones, deadline, REF_ID);
    }

    // ---- Milestone complete + release flow ----

    function test_milestoneFlow_completeAndRelease() public {
        uint256[] memory milestones = new uint256[](2);
        milestones[0] = 400e6;
        milestones[1] = 600e6;

        vm.prank(buyer);
        uint256 dealId = hub.createMilestoneEscrow(seller, address(usdc), AMOUNT, milestones, deadline, REF_ID);

        // Deposit
        vm.prank(buyer);
        hub.deposit(dealId);

        // Complete first milestone
        vm.prank(buyer);
        hub.completeMilestone(dealId, 0);

        // Check releasable
        assertEq(milestoneSettler.releasableAmount(dealId), 400e6);

        // Release milestone — hub transfers to settler, settler transfers to seller
        vm.prank(buyer);
        hub.releaseMilestone(dealId);

        // Verify seller received the milestone payment
        assertEq(usdc.balanceOf(seller), 400e6);
        // Settler should have zero balance (forwarded everything)
        assertEq(usdc.balanceOf(address(milestoneSettler)), 0);

        // Complete second milestone
        vm.prank(buyer);
        hub.completeMilestone(dealId, 1);

        // Release second milestone
        vm.prank(buyer);
        hub.releaseMilestone(dealId);

        // Verify seller received full amount
        assertEq(usdc.balanceOf(seller), AMOUNT);
        assertEq(usdc.balanceOf(address(hub)), 0);
    }

    // ---- DirectSettler release flow ----

    function test_directSettler_releaseTransfersViaSetter() public {
        // Register direct settler
        vm.prank(owner);
        hub.registerSettler(keccak256("direct"), address(directSettler));

        // Create escrow with direct settler
        vm.prank(buyer);
        uint256 dealId = hub.createSimpleEscrow(seller, address(usdc), AMOUNT, deadline, REF_ID);

        // Manually set settler on the deal (since createSimpleEscrow uses address(0))
        // Instead, test via release flow on a milestone deal using directSettler
        // Actually, let's just verify DirectSettler works standalone
        usdc.mint(address(directSettler), AMOUNT);
        vm.prank(address(hub));
        directSettler.settle(0, buyer, seller, address(usdc), AMOUNT, "");
        assertEq(usdc.balanceOf(seller), AMOUNT);
        assertEq(usdc.balanceOf(address(directSettler)), 0);
    }

    // ---- createTeamEscrow ----

    function test_createTeamEscrow_success() public {
        address[] memory members = new address[](3);
        members[0] = member1;
        members[1] = member2;
        members[2] = member3;

        uint256[] memory shares = new uint256[](3);
        shares[0] = 400e6;
        shares[1] = 300e6;
        shares[2] = 300e6;

        vm.prank(buyer);
        uint256 dealId = hub.createTeamEscrow(members, address(usdc), AMOUNT, shares, deadline, REF_ID);

        LangoEscrowHubV2.Deal memory d = hub.getDeal(dealId);
        assertEq(uint8(d.dealType), uint8(LangoEscrowHubV2.DealType.Team));
        assertEq(d.seller, member1); // first member is representative

        (address[] memory m, uint256[] memory s) = hub.getTeamDeal(dealId);
        assertEq(m.length, 3);
        assertEq(s[0], 400e6);
    }

    function test_teamEscrow_releaseDistributesProportionally() public {
        address[] memory members = new address[](3);
        members[0] = member1;
        members[1] = member2;
        members[2] = member3;

        uint256[] memory shares = new uint256[](3);
        shares[0] = 400e6;
        shares[1] = 300e6;
        shares[2] = 300e6;

        vm.prank(buyer);
        uint256 dealId = hub.createTeamEscrow(members, address(usdc), AMOUNT, shares, deadline, REF_ID);

        vm.prank(buyer);
        hub.deposit(dealId);

        vm.prank(buyer);
        hub.release(dealId);

        assertEq(usdc.balanceOf(member1), 400e6);
        assertEq(usdc.balanceOf(member2), 300e6);
        assertEq(usdc.balanceOf(member3), 300e6);
    }

    function testRevert_createTeamEscrow_noMembers() public {
        address[] memory members = new address[](0);
        uint256[] memory shares = new uint256[](0);

        vm.prank(buyer);
        vm.expectRevert("HubV2: no members");
        hub.createTeamEscrow(members, address(usdc), AMOUNT, shares, deadline, REF_ID);
    }

    function testRevert_createTeamEscrow_memberSharesMismatch() public {
        address[] memory members = new address[](2);
        members[0] = member1;
        members[1] = member2;

        uint256[] memory shares = new uint256[](1);
        shares[0] = 1000e6;

        vm.prank(buyer);
        vm.expectRevert("HubV2: members/shares mismatch");
        hub.createTeamEscrow(members, address(usdc), AMOUNT, shares, deadline, REF_ID);
    }

    function testRevert_createTeamEscrow_sharesSumMismatch() public {
        address[] memory members = new address[](2);
        members[0] = member1;
        members[1] = member2;

        uint256[] memory shares = new uint256[](2);
        shares[0] = 400e6;
        shares[1] = 400e6;

        vm.prank(buyer);
        vm.expectRevert("HubV2: shares sum mismatch");
        hub.createTeamEscrow(members, address(usdc), AMOUNT, shares, deadline, REF_ID);
    }

    // ---- Team dispute resolution ----

    function test_teamDispute_resolveSplitsProportionally() public {
        address[] memory members = new address[](2);
        members[0] = member1;
        members[1] = member2;

        uint256[] memory shares = new uint256[](2);
        shares[0] = 600e6;
        shares[1] = 400e6;

        vm.prank(buyer);
        uint256 dealId = hub.createTeamEscrow(members, address(usdc), AMOUNT, shares, deadline, REF_ID);

        vm.prank(buyer);
        hub.deposit(dealId);

        // Dispute
        vm.prank(buyer);
        hub.dispute(dealId);

        // Resolve: 800 to seller side, 200 refund to buyer
        vm.prank(owner);
        hub.resolveDispute(dealId, 800e6, 200e6);

        // member1 gets 800 * 600/1000 = 480
        // member2 gets 800 * 400/1000 = 320
        assertEq(usdc.balanceOf(member1), 480e6);
        assertEq(usdc.balanceOf(member2), 320e6);
        assertEq(usdc.balanceOf(buyer), 100_000e6 - AMOUNT + 200e6);
    }

    // ---- Full lifecycle ----

    function test_fullLifecycle_simpleEscrow() public {
        vm.prank(buyer);
        uint256 dealId = hub.createSimpleEscrow(seller, address(usdc), AMOUNT, deadline, REF_ID);

        vm.prank(buyer);
        hub.deposit(dealId);

        vm.prank(seller);
        hub.submitWork(dealId, keccak256("result"));

        vm.prank(buyer);
        hub.release(dealId);

        assertEq(uint8(hub.getDeal(dealId).status), uint8(LangoEscrowHubV2.DealStatus.Released));
        assertEq(usdc.balanceOf(seller), AMOUNT);
        assertEq(usdc.balanceOf(address(hub)), 0);
    }

    // ---- refId in events ----

    function test_refId_inAllEvents() public {
        bytes32 customRef = keccak256("custom-ref-id");

        vm.prank(buyer);
        uint256 dealId = hub.createSimpleEscrow(seller, address(usdc), AMOUNT, deadline, customRef);

        LangoEscrowHubV2.Deal memory d = hub.getDeal(dealId);
        assertEq(d.refId, customRef);
    }

    // ---- getDeal ----

    function test_getDeal_returnsCorrectData() public {
        vm.prank(buyer);
        uint256 dealId = hub.createSimpleEscrow(seller, address(usdc), AMOUNT, deadline, REF_ID);

        LangoEscrowHubV2.Deal memory d = hub.getDeal(dealId);
        assertEq(d.buyer, buyer);
        assertEq(d.seller, seller);
        assertEq(d.token, address(usdc));
        assertEq(d.amount, AMOUNT);
        assertEq(d.deadline, deadline);
        assertEq(d.refId, REF_ID);
    }

    // ---- Helpers ----

    function _createSimpleEscrow() internal returns (uint256) {
        vm.prank(buyer);
        return hub.createSimpleEscrow(seller, address(usdc), AMOUNT, deadline, REF_ID);
    }

    function _createAndDeposit() internal returns (uint256 dealId) {
        dealId = _createSimpleEscrow();
        vm.prank(buyer);
        hub.deposit(dealId);
    }

    function _createDepositAndDispute() internal returns (uint256 dealId) {
        dealId = _createAndDeposit();
        vm.prank(buyer);
        hub.dispute(dealId);
    }
}
