// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./interfaces/IERC20.sol";

/// @title LangoVault — Individual escrow vault for a single deal.
/// @notice Designed as an EIP-1167 clone target. initialize() replaces constructor.
contract LangoVault {
    enum VaultStatus {
        Uninitialized, // 0
        Created,       // 1
        Deposited,     // 2
        WorkSubmitted, // 3
        Released,      // 4
        Refunded,      // 5
        Disputed,      // 6
        Resolved       // 7
    }

    address public buyer;
    address public seller;
    address public token;
    uint256 public amount;
    uint256 public deadline;
    address public arbitrator;
    VaultStatus public status;
    bytes32 public workHash;

    event VaultInitialized(address indexed buyer, address indexed seller, address token, uint256 amount);
    event Deposited(address indexed buyer, uint256 amount);
    event WorkSubmitted(address indexed seller, bytes32 workHash);
    event Released(address indexed seller, uint256 amount);
    event Refunded(address indexed buyer, uint256 amount);
    event Disputed(address indexed initiator);
    event VaultResolved(bool sellerFavor, uint256 sellerAmount, uint256 buyerAmount);

    modifier onlyBuyer() {
        require(msg.sender == buyer, "Vault: not buyer");
        _;
    }

    modifier onlySeller() {
        require(msg.sender == seller, "Vault: not seller");
        _;
    }

    modifier onlyArbitrator() {
        require(msg.sender == arbitrator, "Vault: not arbitrator");
        _;
    }

    /// @notice Initialize the vault (called once by factory via clone).
    function initialize(
        address _buyer,
        address _seller,
        address _token,
        uint256 _amount,
        uint256 _deadline,
        address _arbitrator
    ) external {
        require(status == VaultStatus.Uninitialized, "Vault: already initialized");
        require(_buyer != address(0), "Vault: zero buyer");
        require(_seller != address(0), "Vault: zero seller");
        require(_token != address(0), "Vault: zero token");
        require(_amount > 0, "Vault: zero amount");
        require(_deadline > block.timestamp, "Vault: past deadline");
        require(_arbitrator != address(0), "Vault: zero arbitrator");

        buyer = _buyer;
        seller = _seller;
        token = _token;
        amount = _amount;
        deadline = _deadline;
        arbitrator = _arbitrator;
        status = VaultStatus.Created;

        emit VaultInitialized(_buyer, _seller, _token, _amount);
    }

    /// @notice Buyer deposits tokens.
    function deposit() external onlyBuyer {
        require(status == VaultStatus.Created, "Vault: not created");
        bool ok = IERC20(token).transferFrom(msg.sender, address(this), amount);
        require(ok, "Vault: transfer failed");
        status = VaultStatus.Deposited;
        emit Deposited(msg.sender, amount);
    }

    /// @notice Seller submits work hash.
    function submitWork(bytes32 _workHash) external onlySeller {
        require(status == VaultStatus.Deposited, "Vault: not deposited");
        require(_workHash != bytes32(0), "Vault: empty hash");
        workHash = _workHash;
        status = VaultStatus.WorkSubmitted;
        emit WorkSubmitted(msg.sender, _workHash);
    }

    /// @notice Buyer releases funds to seller.
    function release() external onlyBuyer {
        require(
            status == VaultStatus.Deposited || status == VaultStatus.WorkSubmitted,
            "Vault: not releasable"
        );
        status = VaultStatus.Released;
        bool ok = IERC20(token).transfer(seller, amount);
        require(ok, "Vault: transfer failed");
        emit Released(seller, amount);
    }

    /// @notice Buyer refunds after deadline.
    function refund() external onlyBuyer {
        require(
            status == VaultStatus.Deposited || status == VaultStatus.WorkSubmitted,
            "Vault: not refundable"
        );
        require(block.timestamp > deadline, "Vault: deadline not passed");
        status = VaultStatus.Refunded;
        bool ok = IERC20(token).transfer(buyer, amount);
        require(ok, "Vault: transfer failed");
        emit Refunded(buyer, amount);
    }

    /// @notice Either party raises a dispute.
    function dispute() external {
        require(msg.sender == buyer || msg.sender == seller, "Vault: not party");
        require(
            status == VaultStatus.Deposited || status == VaultStatus.WorkSubmitted,
            "Vault: not disputable"
        );
        status = VaultStatus.Disputed;
        emit Disputed(msg.sender);
    }

    /// @notice Arbitrator resolves dispute.
    function resolve(bool sellerFavor, uint256 sellerAmount, uint256 buyerAmount) external onlyArbitrator {
        require(status == VaultStatus.Disputed, "Vault: not disputed");
        require(sellerAmount + buyerAmount == amount, "Vault: amounts mismatch");
        status = VaultStatus.Resolved;

        if (sellerAmount > 0) {
            bool ok = IERC20(token).transfer(seller, sellerAmount);
            require(ok, "Vault: seller transfer failed");
        }
        if (buyerAmount > 0) {
            bool ok = IERC20(token).transfer(buyer, buyerAmount);
            require(ok, "Vault: buyer transfer failed");
        }
        emit VaultResolved(sellerFavor, sellerAmount, buyerAmount);
    }
}
