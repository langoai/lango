// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/LangoEscrowHub.sol";
import "./mocks/MockUSDC.sol";

contract LangoEscrowHubTest is Test {
    LangoEscrowHub public hub;
    MockUSDC public usdc;

    address public arbitrator = address(0xA);
    address public buyer = address(0xB);
    address public seller = address(0xC);
    address public stranger = address(0xD);

    uint256 public constant AMOUNT = 1000e6; // 1000 USDC
    uint256 public deadline;

    function setUp() public {
        usdc = new MockUSDC();
        hub = new LangoEscrowHub(arbitrator);

        usdc.mint(buyer, 10_000e6);
        deadline = block.timestamp + 1 days;

        vm.prank(buyer);
        usdc.approve(address(hub), type(uint256).max);
    }

    // ---- constructor ----

    function test_constructor_setsArbitrator() public view {
        assertEq(hub.arbitrator(), arbitrator);
    }

    function testRevert_constructor_zeroArbitrator() public {
        vm.expectRevert("Hub: zero arbitrator");
        new LangoEscrowHub(address(0));
    }

    // ---- createDeal ----

    function test_createDeal_success() public {
        vm.prank(buyer);
        uint256 dealId = hub.createDeal(seller, address(usdc), AMOUNT, deadline);
        assertEq(dealId, 0);
        assertEq(hub.nextDealId(), 1);
    }

    function test_createDeal_incrementsId() public {
        vm.startPrank(buyer);
        hub.createDeal(seller, address(usdc), AMOUNT, deadline);
        uint256 second = hub.createDeal(seller, address(usdc), AMOUNT, deadline);
        vm.stopPrank();
        assertEq(second, 1);
        assertEq(hub.nextDealId(), 2);
    }

    function test_createDeal_emitsDealCreated() public {
        vm.prank(buyer);
        vm.expectEmit(true, true, true, true);
        emit LangoEscrowHub.DealCreated(0, buyer, seller, address(usdc), AMOUNT, deadline);
        hub.createDeal(seller, address(usdc), AMOUNT, deadline);
    }

    function test_createDeal_storesDealData() public {
        vm.prank(buyer);
        uint256 dealId = hub.createDeal(seller, address(usdc), AMOUNT, deadline);

        LangoEscrowHub.Deal memory d = hub.getDeal(dealId);
        assertEq(d.buyer, buyer);
        assertEq(d.seller, seller);
        assertEq(d.token, address(usdc));
        assertEq(d.amount, AMOUNT);
        assertEq(d.deadline, deadline);
        assertEq(uint8(d.status), uint8(LangoEscrowHub.DealStatus.Created));
    }

    function testRevert_createDeal_zeroSeller() public {
        vm.prank(buyer);
        vm.expectRevert("Hub: zero seller");
        hub.createDeal(address(0), address(usdc), AMOUNT, deadline);
    }

    function testRevert_createDeal_zeroToken() public {
        vm.prank(buyer);
        vm.expectRevert("Hub: zero token");
        hub.createDeal(seller, address(0), AMOUNT, deadline);
    }

    function testRevert_createDeal_zeroAmount() public {
        vm.prank(buyer);
        vm.expectRevert("Hub: zero amount");
        hub.createDeal(seller, address(usdc), 0, deadline);
    }

    function testRevert_createDeal_pastDeadline() public {
        vm.prank(buyer);
        vm.expectRevert("Hub: past deadline");
        hub.createDeal(seller, address(usdc), AMOUNT, block.timestamp);
    }

    // ---- deposit ----

    function test_deposit_success() public {
        vm.prank(buyer);
        uint256 dealId = hub.createDeal(seller, address(usdc), AMOUNT, deadline);

        vm.prank(buyer);
        hub.deposit(dealId);

        LangoEscrowHub.Deal memory d = hub.getDeal(dealId);
        assertEq(uint8(d.status), uint8(LangoEscrowHub.DealStatus.Deposited));
        assertEq(usdc.balanceOf(address(hub)), AMOUNT);
    }

    function test_deposit_emitsEvent() public {
        vm.prank(buyer);
        uint256 dealId = hub.createDeal(seller, address(usdc), AMOUNT, deadline);

        vm.prank(buyer);
        vm.expectEmit(true, true, false, true);
        emit LangoEscrowHub.Deposited(dealId, buyer, AMOUNT);
        hub.deposit(dealId);
    }

    function testRevert_deposit_notBuyer() public {
        vm.prank(buyer);
        uint256 dealId = hub.createDeal(seller, address(usdc), AMOUNT, deadline);

        vm.prank(stranger);
        vm.expectRevert("Hub: not buyer");
        hub.deposit(dealId);
    }

    function testRevert_deposit_notCreated() public {
        vm.prank(buyer);
        uint256 dealId = hub.createDeal(seller, address(usdc), AMOUNT, deadline);
        vm.prank(buyer);
        hub.deposit(dealId);

        vm.prank(buyer);
        vm.expectRevert("Hub: not created");
        hub.deposit(dealId);
    }

    // ---- submitWork ----

    function test_submitWork_success() public {
        uint256 dealId = _createAndDeposit();
        bytes32 wh = keccak256("work proof");

        vm.prank(seller);
        hub.submitWork(dealId, wh);

        LangoEscrowHub.Deal memory d = hub.getDeal(dealId);
        assertEq(uint8(d.status), uint8(LangoEscrowHub.DealStatus.WorkSubmitted));
        assertEq(d.workHash, wh);
    }

    function test_submitWork_emitsEvent() public {
        uint256 dealId = _createAndDeposit();
        bytes32 wh = keccak256("work proof");

        vm.prank(seller);
        vm.expectEmit(true, true, false, true);
        emit LangoEscrowHub.WorkSubmitted(dealId, seller, wh);
        hub.submitWork(dealId, wh);
    }

    function testRevert_submitWork_notSeller() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(buyer);
        vm.expectRevert("Hub: not seller");
        hub.submitWork(dealId, keccak256("x"));
    }

    function testRevert_submitWork_notDeposited() public {
        vm.prank(buyer);
        uint256 dealId = hub.createDeal(seller, address(usdc), AMOUNT, deadline);

        vm.prank(seller);
        vm.expectRevert("Hub: not deposited");
        hub.submitWork(dealId, keccak256("x"));
    }

    function testRevert_submitWork_emptyHash() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(seller);
        vm.expectRevert("Hub: empty hash");
        hub.submitWork(dealId, bytes32(0));
    }

    // ---- release ----

    function test_release_afterDeposit() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(buyer);
        hub.release(dealId);

        LangoEscrowHub.Deal memory d = hub.getDeal(dealId);
        assertEq(uint8(d.status), uint8(LangoEscrowHub.DealStatus.Released));
        assertEq(usdc.balanceOf(seller), AMOUNT);
    }

    function test_release_afterWorkSubmitted() public {
        uint256 dealId = _createAndDeposit();
        vm.prank(seller);
        hub.submitWork(dealId, keccak256("proof"));

        vm.prank(buyer);
        hub.release(dealId);

        assertEq(uint8(hub.getDeal(dealId).status), uint8(LangoEscrowHub.DealStatus.Released));
        assertEq(usdc.balanceOf(seller), AMOUNT);
    }

    function test_release_emitsEvent() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(buyer);
        vm.expectEmit(true, true, false, true);
        emit LangoEscrowHub.Released(dealId, seller, AMOUNT);
        hub.release(dealId);
    }

    function testRevert_release_notReleasable() public {
        vm.prank(buyer);
        uint256 dealId = hub.createDeal(seller, address(usdc), AMOUNT, deadline);

        vm.prank(buyer);
        vm.expectRevert("Hub: not releasable");
        hub.release(dealId);
    }

    // ---- refund ----

    function test_refund_afterDeadline() public {
        uint256 dealId = _createAndDeposit();

        vm.warp(deadline + 1);

        vm.prank(buyer);
        hub.refund(dealId);

        assertEq(uint8(hub.getDeal(dealId).status), uint8(LangoEscrowHub.DealStatus.Refunded));
        assertEq(usdc.balanceOf(buyer), 10_000e6); // full balance restored
    }

    function testRevert_refund_deadlineNotPassed() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(buyer);
        vm.expectRevert("Hub: deadline not passed");
        hub.refund(dealId);
    }

    // ---- dispute ----

    function test_dispute_byBuyer() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(buyer);
        hub.dispute(dealId);

        assertEq(uint8(hub.getDeal(dealId).status), uint8(LangoEscrowHub.DealStatus.Disputed));
    }

    function test_dispute_bySeller() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(seller);
        hub.dispute(dealId);

        assertEq(uint8(hub.getDeal(dealId).status), uint8(LangoEscrowHub.DealStatus.Disputed));
    }

    function test_dispute_emitsEvent() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(buyer);
        vm.expectEmit(true, true, false, false);
        emit LangoEscrowHub.Disputed(dealId, buyer);
        hub.dispute(dealId);
    }

    function testRevert_dispute_notParty() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(stranger);
        vm.expectRevert("Hub: not party");
        hub.dispute(dealId);
    }

    function testRevert_dispute_notDisputable() public {
        vm.prank(buyer);
        uint256 dealId = hub.createDeal(seller, address(usdc), AMOUNT, deadline);

        vm.prank(buyer);
        vm.expectRevert("Hub: not disputable");
        hub.dispute(dealId);
    }

    // ---- resolveDispute ----

    function test_resolveDispute_fullSeller() public {
        uint256 dealId = _createDepositAndDispute();

        vm.prank(arbitrator);
        hub.resolveDispute(dealId, true, AMOUNT, 0);

        assertEq(uint8(hub.getDeal(dealId).status), uint8(LangoEscrowHub.DealStatus.Resolved));
        assertEq(usdc.balanceOf(seller), AMOUNT);
    }

    function test_resolveDispute_split() public {
        uint256 dealId = _createDepositAndDispute();

        uint256 sellerAmt = 600e6;
        uint256 buyerAmt = 400e6;

        vm.prank(arbitrator);
        hub.resolveDispute(dealId, true, sellerAmt, buyerAmt);

        assertEq(usdc.balanceOf(seller), sellerAmt);
        assertEq(usdc.balanceOf(buyer), 10_000e6 - AMOUNT + buyerAmt);
    }

    function test_resolveDispute_emitsEvent() public {
        uint256 dealId = _createDepositAndDispute();

        vm.prank(arbitrator);
        vm.expectEmit(true, false, false, true);
        emit LangoEscrowHub.DealResolved(dealId, true, AMOUNT, 0);
        hub.resolveDispute(dealId, true, AMOUNT, 0);
    }

    function testRevert_resolveDispute_notArbitrator() public {
        uint256 dealId = _createDepositAndDispute();

        vm.prank(buyer);
        vm.expectRevert("Hub: not arbitrator");
        hub.resolveDispute(dealId, true, AMOUNT, 0);
    }

    function testRevert_resolveDispute_notDisputed() public {
        uint256 dealId = _createAndDeposit();

        vm.prank(arbitrator);
        vm.expectRevert("Hub: not disputed");
        hub.resolveDispute(dealId, true, AMOUNT, 0);
    }

    function testRevert_resolveDispute_amountsMismatch() public {
        uint256 dealId = _createDepositAndDispute();

        vm.prank(arbitrator);
        vm.expectRevert("Hub: amounts mismatch");
        hub.resolveDispute(dealId, true, AMOUNT, 1);
    }

    // ---- getDeal ----

    function test_getDeal_returnsCorrectData() public {
        vm.prank(buyer);
        uint256 dealId = hub.createDeal(seller, address(usdc), AMOUNT, deadline);

        LangoEscrowHub.Deal memory d = hub.getDeal(dealId);
        assertEq(d.buyer, buyer);
        assertEq(d.seller, seller);
        assertEq(d.token, address(usdc));
        assertEq(d.amount, AMOUNT);
        assertEq(d.deadline, deadline);
    }

    // ---- full lifecycle ----

    function test_fullLifecycle_createDepositSubmitRelease() public {
        vm.prank(buyer);
        uint256 dealId = hub.createDeal(seller, address(usdc), AMOUNT, deadline);

        vm.prank(buyer);
        hub.deposit(dealId);

        vm.prank(seller);
        hub.submitWork(dealId, keccak256("result"));

        vm.prank(buyer);
        hub.release(dealId);

        assertEq(uint8(hub.getDeal(dealId).status), uint8(LangoEscrowHub.DealStatus.Released));
        assertEq(usdc.balanceOf(seller), AMOUNT);
        assertEq(usdc.balanceOf(address(hub)), 0);
    }

    // ---- helpers ----

    function _createAndDeposit() internal returns (uint256 dealId) {
        vm.prank(buyer);
        dealId = hub.createDeal(seller, address(usdc), AMOUNT, deadline);
        vm.prank(buyer);
        hub.deposit(dealId);
    }

    function _createDepositAndDispute() internal returns (uint256 dealId) {
        dealId = _createAndDeposit();
        vm.prank(buyer);
        hub.dispute(dealId);
    }
}
