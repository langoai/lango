// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title ILangoEconomy — Unified entry points for P2P agent economy.
interface ILangoEconomy {
    // ---- Events ----

    event EscrowOpened(
        bytes32 indexed refId, uint256 indexed dealId, address buyer, address seller, uint256 amount
    );
    event MilestoneReached(bytes32 indexed refId, uint256 indexed dealId, uint256 milestoneIndex, uint256 amount);
    event DisputeRaised(bytes32 indexed refId, uint256 indexed dealId, address initiator);
    event SettlementFinalized(
        bytes32 indexed refId, uint256 indexed dealId, address settler, uint256 sellerAmount, uint256 buyerRefund
    );

    // ---- Entry Points ----

    /// @notice Immediate transfer — no escrow period.
    function directSettle(address seller, address token, uint256 amount, bytes32 refId) external;

    /// @notice Create a simple escrow deal with a deadline.
    function createSimpleEscrow(address seller, address token, uint256 amount, uint256 deadline, bytes32 refId)
        external
        returns (uint256 dealId);

    /// @notice Create a milestone-based escrow deal.
    function createMilestoneEscrow(
        address seller,
        address token,
        uint256 totalAmount,
        uint256[] calldata milestoneAmounts,
        uint256 deadline,
        bytes32 refId
    ) external returns (uint256 dealId);

    /// @notice Create a team escrow with proportional shares.
    function createTeamEscrow(
        address[] calldata members,
        address token,
        uint256 totalAmount,
        uint256[] calldata shares,
        uint256 deadline,
        bytes32 refId
    ) external returns (uint256 dealId);
}
