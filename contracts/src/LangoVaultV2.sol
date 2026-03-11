// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Initializable} from "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "./interfaces/IERC20.sol";
import "./interfaces/ISettler.sol";

/// @title LangoVaultV2 — Beacon-compatible individual escrow vault with refId and settler support.
/// @notice Designed as a BeaconProxy implementation. initialize() replaces constructor.
contract LangoVaultV2 is Initializable, ReentrancyGuard {
    enum VaultStatus {
        Uninitialized, // 0
        Created, // 1
        Deposited, // 2
        WorkSubmitted, // 3
        Released, // 4
        Refunded, // 5
        Disputed, // 6
        Resolved // 7
    }

    address public buyer;
    address public seller;
    address public token;
    uint256 public amount;
    uint256 public deadline;
    address public arbiter;
    address public settler;
    bytes32 public refId;
    VaultStatus public status;
    bytes32 public workHash;

    // ---- Events (all with indexed refId) ----

    event VaultInitialized(
        bytes32 indexed refId, address indexed buyer, address indexed seller, address token, uint256 amount
    );
    event Deposited(bytes32 indexed refId, address indexed buyer, uint256 amount);
    event WorkSubmitted(bytes32 indexed refId, address indexed seller, bytes32 workHash);
    event Released(bytes32 indexed refId, address indexed seller, uint256 amount);
    event Refunded(bytes32 indexed refId, address indexed buyer, uint256 amount);
    event Disputed(bytes32 indexed refId, address indexed initiator);
    event VaultResolved(bytes32 indexed refId, uint256 sellerAmount, uint256 buyerAmount);

    // ---- Modifiers ----

    modifier onlyBuyer() {
        require(msg.sender == buyer, "VaultV2: not buyer");
        _;
    }

    modifier onlySeller() {
        require(msg.sender == seller, "VaultV2: not seller");
        _;
    }

    modifier onlyArbiter() {
        require(msg.sender == arbiter, "VaultV2: not arbiter");
        _;
    }

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    /// @notice Initialize the vault (called once by factory via BeaconProxy).
    function initialize(
        address buyer_,
        address seller_,
        address token_,
        uint256 amount_,
        address arbiter_,
        bytes32 refId_
    ) external initializer {
        require(buyer_ != address(0), "VaultV2: zero buyer");
        require(seller_ != address(0), "VaultV2: zero seller");
        require(token_ != address(0), "VaultV2: zero token");
        require(amount_ > 0, "VaultV2: zero amount");
        require(arbiter_ != address(0), "VaultV2: zero arbiter");
        require(refId_ != bytes32(0), "VaultV2: zero refId");

        buyer = buyer_;
        seller = seller_;
        token = token_;
        amount = amount_;
        deadline = block.timestamp + 30 days;
        arbiter = arbiter_;
        refId = refId_;
        status = VaultStatus.Created;

        emit VaultInitialized(refId_, buyer_, seller_, token_, amount_);
    }

    /// @notice Buyer deposits tokens.
    function deposit() external onlyBuyer nonReentrant {
        require(status == VaultStatus.Created, "VaultV2: not created");
        bool ok = IERC20(token).transferFrom(msg.sender, address(this), amount);
        require(ok, "VaultV2: transfer failed");
        status = VaultStatus.Deposited;
        emit Deposited(refId, msg.sender, amount);
    }

    /// @notice Seller submits work hash.
    function submitWork(bytes32 workHash_) external onlySeller {
        require(status == VaultStatus.Deposited, "VaultV2: not deposited");
        require(workHash_ != bytes32(0), "VaultV2: empty hash");
        workHash = workHash_;
        status = VaultStatus.WorkSubmitted;
        emit WorkSubmitted(refId, msg.sender, workHash_);
    }

    /// @notice Buyer releases funds to seller.
    function release() external onlyBuyer nonReentrant {
        require(
            status == VaultStatus.Deposited || status == VaultStatus.WorkSubmitted, "VaultV2: not releasable"
        );
        status = VaultStatus.Released;

        if (settler != address(0) && ISettler(settler).canSettle(0)) {
            bool ok = IERC20(token).transfer(settler, amount);
            require(ok, "VaultV2: settler transfer failed");
            ISettler(settler).settle(0, buyer, seller, token, amount, "");
        } else {
            bool ok = IERC20(token).transfer(seller, amount);
            require(ok, "VaultV2: transfer failed");
        }

        emit Released(refId, seller, amount);
    }

    /// @notice Buyer refunds after deadline.
    function refund() external onlyBuyer nonReentrant {
        require(
            status == VaultStatus.Deposited || status == VaultStatus.WorkSubmitted, "VaultV2: not refundable"
        );
        require(block.timestamp > deadline, "VaultV2: deadline not passed");
        status = VaultStatus.Refunded;
        bool ok = IERC20(token).transfer(buyer, amount);
        require(ok, "VaultV2: transfer failed");
        emit Refunded(refId, buyer, amount);
    }

    /// @notice Either party raises a dispute.
    function dispute() external {
        require(msg.sender == buyer || msg.sender == seller, "VaultV2: not party");
        require(
            status == VaultStatus.Deposited || status == VaultStatus.WorkSubmitted, "VaultV2: not disputable"
        );
        status = VaultStatus.Disputed;
        emit Disputed(refId, msg.sender);
    }

    /// @notice Arbiter resolves dispute.
    function resolve(uint256 sellerAmount, uint256 buyerAmount) external onlyArbiter nonReentrant {
        require(status == VaultStatus.Disputed, "VaultV2: not disputed");
        require(sellerAmount + buyerAmount == amount, "VaultV2: amounts mismatch");
        status = VaultStatus.Resolved;

        if (sellerAmount > 0) {
            bool ok = IERC20(token).transfer(seller, sellerAmount);
            require(ok, "VaultV2: seller transfer failed");
        }
        if (buyerAmount > 0) {
            bool ok = IERC20(token).transfer(buyer, buyerAmount);
            require(ok, "VaultV2: buyer transfer failed");
        }
        emit VaultResolved(refId, sellerAmount, buyerAmount);
    }

    /// @notice Set settler address (can only be set once by buyer).
    function setSettler(address settler_) external onlyBuyer {
        require(settler == address(0), "VaultV2: settler already set");
        settler = settler_;
    }
}
