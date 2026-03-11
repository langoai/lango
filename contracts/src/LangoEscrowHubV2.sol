// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Initializable} from "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts/proxy/utils/UUPSUpgradeable.sol";
import {OwnableUpgradeable} from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "./interfaces/IERC20.sol";
import "./interfaces/ILangoEconomy.sol";
import "./interfaces/ISettler.sol";

/// @title LangoEscrowHubV2 — UUPS-upgradeable master escrow hub with settler pattern.
/// @notice Holds multiple deals in a single contract. Supports direct settle, simple escrow,
///         milestone escrow, and team escrow via pluggable ISettler implementations.
contract LangoEscrowHubV2 is
    Initializable,
    UUPSUpgradeable,
    OwnableUpgradeable,
    ReentrancyGuard,
    ILangoEconomy
{
    // ---- Enums ----

    enum DealStatus {
        Created, // 0
        Deposited, // 1
        WorkSubmitted, // 2
        Released, // 3
        Refunded, // 4
        Disputed, // 5
        Resolved // 6
    }

    enum DealType {
        Simple, // 0
        Milestone, // 1
        Team // 2
    }

    // ---- Structs ----

    struct Deal {
        address buyer;
        address seller;
        address token;
        uint256 amount;
        uint256 deadline;
        DealStatus status;
        DealType dealType;
        bytes32 workHash;
        bytes32 refId;
        address settler;
    }

    struct TeamDeal {
        address[] members;
        uint256[] shares;
    }

    // ---- State ----

    uint256 public nextDealId;
    mapping(uint256 => Deal) public deals;
    mapping(bytes32 => address) public settlers; // settlerType => settler address
    mapping(uint256 => TeamDeal) internal _teamDeals;

    // ---- Events (beyond ILangoEconomy) ----

    event Deposited(bytes32 indexed refId, uint256 indexed dealId, address indexed buyer, uint256 amount);
    event WorkSubmitted(bytes32 indexed refId, uint256 indexed dealId, address indexed seller, bytes32 workHash);
    event Released(bytes32 indexed refId, uint256 indexed dealId, address indexed seller, uint256 amount);
    event Refunded(bytes32 indexed refId, uint256 indexed dealId, address indexed buyer, uint256 amount);
    event SettlerRegistered(bytes32 indexed settlerType, address settler);

    // ---- Modifiers ----

    modifier onlyBuyer(uint256 dealId) {
        require(msg.sender == deals[dealId].buyer, "HubV2: not buyer");
        _;
    }

    modifier onlySeller(uint256 dealId) {
        require(msg.sender == deals[dealId].seller, "HubV2: not seller");
        _;
    }

    // ---- Initializer ----

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(address owner_) external initializer {
        require(owner_ != address(0), "HubV2: zero owner");
        __Ownable_init(owner_);
    }

    // ---- UUPS ----

    function _authorizeUpgrade(address) internal override onlyOwner {}

    // ---- Settler Management ----

    /// @notice Register a settler implementation for a given type.
    function registerSettler(bytes32 settlerType, address settler) external onlyOwner {
        require(settler != address(0), "HubV2: zero settler");
        settlers[settlerType] = settler;
        emit SettlerRegistered(settlerType, settler);
    }

    // ---- ILangoEconomy Entry Points ----

    /// @inheritdoc ILangoEconomy
    function directSettle(address seller, address token, uint256 amount, bytes32 refId)
        external
        override
        nonReentrant
    {
        require(seller != address(0), "HubV2: zero seller");
        require(token != address(0), "HubV2: zero token");
        require(amount > 0, "HubV2: zero amount");
        require(refId != bytes32(0), "HubV2: zero refId");

        uint256 dealId = nextDealId++;
        deals[dealId] = Deal({
            buyer: msg.sender,
            seller: seller,
            token: token,
            amount: amount,
            deadline: 0,
            status: DealStatus.Released,
            dealType: DealType.Simple,
            workHash: bytes32(0),
            refId: refId,
            settler: address(0)
        });

        bool ok = IERC20(token).transferFrom(msg.sender, seller, amount);
        require(ok, "HubV2: transfer failed");

        emit EscrowOpened(refId, dealId, msg.sender, seller, amount);
        emit SettlementFinalized(refId, dealId, address(0), amount, 0);
    }

    /// @inheritdoc ILangoEconomy
    function createSimpleEscrow(address seller, address token, uint256 amount, uint256 deadline, bytes32 refId)
        external
        override
        nonReentrant
        returns (uint256 dealId)
    {
        dealId = _createDeal(seller, token, amount, deadline, refId, DealType.Simple, address(0));
    }

    /// @inheritdoc ILangoEconomy
    function createMilestoneEscrow(
        address seller,
        address token,
        uint256 totalAmount,
        uint256[] calldata milestoneAmounts,
        uint256 deadline,
        bytes32 refId
    ) external override nonReentrant returns (uint256 dealId) {
        require(milestoneAmounts.length > 0, "HubV2: no milestones");

        uint256 sum;
        for (uint256 i; i < milestoneAmounts.length; ++i) {
            require(milestoneAmounts[i] > 0, "HubV2: zero milestone amount");
            sum += milestoneAmounts[i];
        }
        require(sum == totalAmount, "HubV2: milestones sum mismatch");

        address settler = settlers[keccak256("milestone")];
        require(settler != address(0), "HubV2: milestone settler not set");

        dealId = _createDeal(seller, token, totalAmount, deadline, refId, DealType.Milestone, settler);

        MilestoneSettlerLike(settler).initMilestones(dealId, milestoneAmounts);
    }

    /// @inheritdoc ILangoEconomy
    function createTeamEscrow(
        address[] calldata members,
        address token,
        uint256 totalAmount,
        uint256[] calldata shares,
        uint256 deadline,
        bytes32 refId
    ) external override nonReentrant returns (uint256 dealId) {
        require(members.length > 0, "HubV2: no members");
        require(members.length == shares.length, "HubV2: members/shares mismatch");

        uint256 sum;
        for (uint256 i; i < members.length; ++i) {
            require(members[i] != address(0), "HubV2: zero member");
            require(shares[i] > 0, "HubV2: zero share");
            sum += shares[i];
        }
        require(sum == totalAmount, "HubV2: shares sum mismatch");

        // Use first member as the "seller" representative
        dealId = _createDeal(members[0], token, totalAmount, deadline, refId, DealType.Team, address(0));

        _teamDeals[dealId].members = members;
        _teamDeals[dealId].shares = shares;
    }

    // ---- Escrow Operations ----

    /// @notice Buyer deposits ERC-20 tokens into the escrow.
    function deposit(uint256 dealId) external onlyBuyer(dealId) nonReentrant {
        Deal storage d = deals[dealId];
        require(d.status == DealStatus.Created, "HubV2: not created");

        bool ok = IERC20(d.token).transferFrom(msg.sender, address(this), d.amount);
        require(ok, "HubV2: transfer failed");

        d.status = DealStatus.Deposited;
        emit Deposited(d.refId, dealId, msg.sender, d.amount);
    }

    /// @notice Seller submits work proof hash.
    function submitWork(uint256 dealId, bytes32 workHash) external onlySeller(dealId) {
        Deal storage d = deals[dealId];
        require(d.status == DealStatus.Deposited, "HubV2: not deposited");
        require(workHash != bytes32(0), "HubV2: empty hash");

        d.workHash = workHash;
        d.status = DealStatus.WorkSubmitted;
        emit WorkSubmitted(d.refId, dealId, msg.sender, workHash);
    }

    /// @notice Buyer releases funds to seller after accepting work.
    function release(uint256 dealId) external onlyBuyer(dealId) nonReentrant {
        Deal storage d = deals[dealId];
        require(
            d.status == DealStatus.Deposited || d.status == DealStatus.WorkSubmitted, "HubV2: not releasable"
        );

        d.status = DealStatus.Released;

        if (d.dealType == DealType.Team) {
            _releaseTeamFunds(dealId, d);
        } else if (d.settler != address(0) && ISettler(d.settler).canSettle(dealId)) {
            bool ok2 = IERC20(d.token).transfer(d.settler, d.amount);
            require(ok2, "HubV2: settler transfer failed");
            ISettler(d.settler).settle(dealId, d.buyer, d.seller, d.token, d.amount, "");
        } else {
            bool ok = IERC20(d.token).transfer(d.seller, d.amount);
            require(ok, "HubV2: transfer failed");
        }

        emit Released(d.refId, dealId, d.seller, d.amount);
        emit SettlementFinalized(d.refId, dealId, d.settler, d.amount, 0);
    }

    /// @notice Buyer requests refund after deadline passes.
    function refund(uint256 dealId) external onlyBuyer(dealId) nonReentrant {
        Deal storage d = deals[dealId];
        require(
            d.status == DealStatus.Deposited || d.status == DealStatus.WorkSubmitted, "HubV2: not refundable"
        );
        require(block.timestamp > d.deadline, "HubV2: deadline not passed");

        d.status = DealStatus.Refunded;
        bool ok = IERC20(d.token).transfer(d.buyer, d.amount);
        require(ok, "HubV2: transfer failed");

        emit Refunded(d.refId, dealId, d.buyer, d.amount);
        emit SettlementFinalized(d.refId, dealId, address(0), 0, d.amount);
    }

    /// @notice Either party raises a dispute.
    function dispute(uint256 dealId) external {
        Deal storage d = deals[dealId];
        require(msg.sender == d.buyer || msg.sender == d.seller, "HubV2: not party");
        require(
            d.status == DealStatus.Deposited || d.status == DealStatus.WorkSubmitted, "HubV2: not disputable"
        );

        d.status = DealStatus.Disputed;
        emit DisputeRaised(d.refId, dealId, msg.sender);
    }

    /// @notice Owner resolves a dispute by splitting funds.
    function resolveDispute(uint256 dealId, uint256 sellerAmount, uint256 buyerAmount)
        external
        onlyOwner
        nonReentrant
    {
        Deal storage d = deals[dealId];
        require(d.status == DealStatus.Disputed, "HubV2: not disputed");
        require(sellerAmount + buyerAmount == d.amount, "HubV2: amounts mismatch");

        d.status = DealStatus.Resolved;

        if (sellerAmount > 0) {
            if (d.dealType == DealType.Team) {
                _distributeTeamFunds(dealId, d, sellerAmount);
            } else {
                bool ok = IERC20(d.token).transfer(d.seller, sellerAmount);
                require(ok, "HubV2: seller transfer failed");
            }
        }
        if (buyerAmount > 0) {
            bool ok = IERC20(d.token).transfer(d.buyer, buyerAmount);
            require(ok, "HubV2: buyer transfer failed");
        }

        emit SettlementFinalized(d.refId, dealId, address(0), sellerAmount, buyerAmount);
    }

    // ---- Milestone Operations ----

    /// @notice Complete a milestone for a milestone-type deal.
    function completeMilestone(uint256 dealId, uint256 index) external onlyBuyer(dealId) {
        Deal storage d = deals[dealId];
        require(d.dealType == DealType.Milestone, "HubV2: not milestone deal");
        require(d.status == DealStatus.Deposited || d.status == DealStatus.WorkSubmitted, "HubV2: invalid status");
        require(d.settler != address(0), "HubV2: no settler");

        MilestoneSettlerLike(d.settler).completeMilestone(dealId, index);

        (uint256 milestoneAmount) = MilestoneSettlerLike(d.settler).getMilestoneAmount(dealId, index);
        emit MilestoneReached(d.refId, dealId, index, milestoneAmount);
    }

    /// @notice Release milestone funds for completed milestones.
    function releaseMilestone(uint256 dealId) external onlyBuyer(dealId) nonReentrant {
        Deal storage d = deals[dealId];
        require(d.dealType == DealType.Milestone, "HubV2: not milestone deal");
        require(d.status == DealStatus.Deposited || d.status == DealStatus.WorkSubmitted, "HubV2: invalid status");
        require(d.settler != address(0), "HubV2: no settler");
        require(ISettler(d.settler).canSettle(dealId), "HubV2: cannot settle");

        uint256 releasable = MilestoneSettlerLike(d.settler).releasableAmount(dealId);
        require(releasable > 0, "HubV2: nothing to release");

        bool ok = IERC20(d.token).transfer(d.settler, releasable);
        require(ok, "HubV2: settler transfer failed");
        ISettler(d.settler).settle(dealId, d.buyer, d.seller, d.token, releasable, "");

        emit Released(d.refId, dealId, d.seller, releasable);
    }

    // ---- View ----

    /// @notice Get deal details.
    function getDeal(uint256 dealId) external view returns (Deal memory) {
        return deals[dealId];
    }

    /// @notice Get team deal details.
    function getTeamDeal(uint256 dealId) external view returns (address[] memory members, uint256[] memory shares) {
        TeamDeal storage td = _teamDeals[dealId];
        return (td.members, td.shares);
    }

    // ---- Internal ----

    function _createDeal(
        address seller,
        address token,
        uint256 amount,
        uint256 deadline,
        bytes32 refId,
        DealType dealType,
        address settler
    ) internal returns (uint256 dealId) {
        require(seller != address(0), "HubV2: zero seller");
        require(token != address(0), "HubV2: zero token");
        require(amount > 0, "HubV2: zero amount");
        require(deadline > block.timestamp, "HubV2: past deadline");
        require(refId != bytes32(0), "HubV2: zero refId");

        dealId = nextDealId++;
        deals[dealId] = Deal({
            buyer: msg.sender,
            seller: seller,
            token: token,
            amount: amount,
            deadline: deadline,
            status: DealStatus.Created,
            dealType: dealType,
            workHash: bytes32(0),
            refId: refId,
            settler: settler
        });

        emit EscrowOpened(refId, dealId, msg.sender, seller, amount);
    }

    function _releaseTeamFunds(uint256 dealId, Deal storage d) internal {
        TeamDeal storage td = _teamDeals[dealId];
        for (uint256 i; i < td.members.length; ++i) {
            bool ok = IERC20(d.token).transfer(td.members[i], td.shares[i]);
            require(ok, "HubV2: team transfer failed");
        }
    }

    function _distributeTeamFunds(uint256 dealId, Deal storage d, uint256 totalSellerAmount) internal {
        TeamDeal storage td = _teamDeals[dealId];
        uint256 totalShares = d.amount;
        for (uint256 i; i < td.members.length; ++i) {
            uint256 memberAmount = (totalSellerAmount * td.shares[i]) / totalShares;
            if (memberAmount > 0) {
                bool ok = IERC20(d.token).transfer(td.members[i], memberAmount);
                require(ok, "HubV2: team transfer failed");
            }
        }
    }
}

/// @dev Minimal interface for MilestoneSettler calls from the hub.
interface MilestoneSettlerLike {
    function initMilestones(uint256 dealId, uint256[] calldata amounts) external;
    function completeMilestone(uint256 dealId, uint256 index) external;
    function getMilestoneAmount(uint256 dealId, uint256 index) external view returns (uint256);
    function releasableAmount(uint256 dealId) external view returns (uint256);
}
