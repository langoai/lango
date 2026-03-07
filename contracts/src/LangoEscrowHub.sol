// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./interfaces/IERC20.sol";

/// @title LangoEscrowHub — Master escrow hub for P2P agent deals.
/// @notice Holds multiple deals in a single contract. Pull-over-push pattern.
contract LangoEscrowHub {
    enum DealStatus {
        Created,    // 0
        Deposited,  // 1
        WorkSubmitted, // 2
        Released,   // 3
        Refunded,   // 4
        Disputed,   // 5
        Resolved    // 6
    }

    struct Deal {
        address buyer;
        address seller;
        address token;
        uint256 amount;
        uint256 deadline;
        DealStatus status;
        bytes32 workHash;
    }

    uint256 public nextDealId;
    address public arbitrator;
    mapping(uint256 => Deal) public deals;

    event DealCreated(uint256 indexed dealId, address indexed buyer, address indexed seller, address token, uint256 amount, uint256 deadline);
    event Deposited(uint256 indexed dealId, address indexed buyer, uint256 amount);
    event WorkSubmitted(uint256 indexed dealId, address indexed seller, bytes32 workHash);
    event Released(uint256 indexed dealId, address indexed seller, uint256 amount);
    event Refunded(uint256 indexed dealId, address indexed buyer, uint256 amount);
    event Disputed(uint256 indexed dealId, address indexed initiator);
    event DealResolved(uint256 indexed dealId, bool sellerFavor, uint256 sellerAmount, uint256 buyerAmount);

    modifier onlyBuyer(uint256 dealId) {
        require(msg.sender == deals[dealId].buyer, "Hub: not buyer");
        _;
    }

    modifier onlySeller(uint256 dealId) {
        require(msg.sender == deals[dealId].seller, "Hub: not seller");
        _;
    }

    modifier onlyArbitrator() {
        require(msg.sender == arbitrator, "Hub: not arbitrator");
        _;
    }

    constructor(address _arbitrator) {
        require(_arbitrator != address(0), "Hub: zero arbitrator");
        arbitrator = _arbitrator;
    }

    /// @notice Create a new escrow deal.
    function createDeal(
        address seller,
        address token,
        uint256 amount,
        uint256 deadline
    ) external returns (uint256 dealId) {
        require(seller != address(0), "Hub: zero seller");
        require(token != address(0), "Hub: zero token");
        require(amount > 0, "Hub: zero amount");
        require(deadline > block.timestamp, "Hub: past deadline");

        dealId = nextDealId++;
        deals[dealId] = Deal({
            buyer: msg.sender,
            seller: seller,
            token: token,
            amount: amount,
            deadline: deadline,
            status: DealStatus.Created,
            workHash: bytes32(0)
        });

        emit DealCreated(dealId, msg.sender, seller, token, amount, deadline);
    }

    /// @notice Buyer deposits ERC-20 tokens into the escrow.
    function deposit(uint256 dealId) external onlyBuyer(dealId) {
        Deal storage d = deals[dealId];
        require(d.status == DealStatus.Created, "Hub: not created");

        bool ok = IERC20(d.token).transferFrom(msg.sender, address(this), d.amount);
        require(ok, "Hub: transfer failed");

        d.status = DealStatus.Deposited;
        emit Deposited(dealId, msg.sender, d.amount);
    }

    /// @notice Seller submits work proof hash.
    function submitWork(uint256 dealId, bytes32 workHash) external onlySeller(dealId) {
        Deal storage d = deals[dealId];
        require(d.status == DealStatus.Deposited, "Hub: not deposited");
        require(workHash != bytes32(0), "Hub: empty hash");

        d.workHash = workHash;
        d.status = DealStatus.WorkSubmitted;
        emit WorkSubmitted(dealId, msg.sender, workHash);
    }

    /// @notice Buyer releases funds to seller after accepting work.
    function release(uint256 dealId) external onlyBuyer(dealId) {
        Deal storage d = deals[dealId];
        require(
            d.status == DealStatus.Deposited || d.status == DealStatus.WorkSubmitted,
            "Hub: not releasable"
        );

        d.status = DealStatus.Released;
        bool ok = IERC20(d.token).transfer(d.seller, d.amount);
        require(ok, "Hub: transfer failed");

        emit Released(dealId, d.seller, d.amount);
    }

    /// @notice Buyer requests refund after deadline passes.
    function refund(uint256 dealId) external onlyBuyer(dealId) {
        Deal storage d = deals[dealId];
        require(
            d.status == DealStatus.Deposited || d.status == DealStatus.WorkSubmitted,
            "Hub: not refundable"
        );
        require(block.timestamp > d.deadline, "Hub: deadline not passed");

        d.status = DealStatus.Refunded;
        bool ok = IERC20(d.token).transfer(d.buyer, d.amount);
        require(ok, "Hub: transfer failed");

        emit Refunded(dealId, d.buyer, d.amount);
    }

    /// @notice Either party raises a dispute.
    function dispute(uint256 dealId) external {
        Deal storage d = deals[dealId];
        require(msg.sender == d.buyer || msg.sender == d.seller, "Hub: not party");
        require(
            d.status == DealStatus.Deposited || d.status == DealStatus.WorkSubmitted,
            "Hub: not disputable"
        );

        d.status = DealStatus.Disputed;
        emit Disputed(dealId, msg.sender);
    }

    /// @notice Arbitrator resolves a dispute by splitting funds.
    function resolveDispute(
        uint256 dealId,
        bool sellerFavor,
        uint256 sellerAmount,
        uint256 buyerAmount
    ) external onlyArbitrator {
        Deal storage d = deals[dealId];
        require(d.status == DealStatus.Disputed, "Hub: not disputed");
        require(sellerAmount + buyerAmount == d.amount, "Hub: amounts mismatch");

        d.status = DealStatus.Resolved;

        if (sellerAmount > 0) {
            bool ok = IERC20(d.token).transfer(d.seller, sellerAmount);
            require(ok, "Hub: seller transfer failed");
        }
        if (buyerAmount > 0) {
            bool ok = IERC20(d.token).transfer(d.buyer, buyerAmount);
            require(ok, "Hub: buyer transfer failed");
        }

        emit DealResolved(dealId, sellerFavor, sellerAmount, buyerAmount);
    }

    /// @notice Get deal details.
    function getDeal(uint256 dealId) external view returns (Deal memory) {
        return deals[dealId];
    }
}
