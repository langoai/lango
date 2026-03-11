// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "../interfaces/ISettler.sol";
import "../interfaces/IERC20.sol";

/// @title MilestoneSettler — Milestone-based settlement strategy.
/// @notice Tracks milestone completion and releases proportional amounts to the seller.
contract MilestoneSettler is ISettler {
    struct MilestoneData {
        uint256[] amounts;
        bool[] completed;
        uint256 releasedTotal;
    }

    /// @notice The escrow hub that is authorized to call this settler.
    address public hub;

    mapping(uint256 => MilestoneData) internal _milestones;

    event MilestoneInitialized(uint256 indexed dealId, uint256 milestoneCount);
    event MilestoneCompleted(uint256 indexed dealId, uint256 indexed index, uint256 amount);

    modifier onlyHub() {
        require(msg.sender == hub, "MilestoneSettler: not hub");
        _;
    }

    constructor(address hub_) {
        require(hub_ != address(0), "MilestoneSettler: zero hub");
        hub = hub_;
    }

    /// @notice Initialize milestones for a deal. Called by hub during createMilestoneEscrow.
    function initMilestones(uint256 dealId, uint256[] calldata amounts) external onlyHub {
        require(_milestones[dealId].amounts.length == 0, "MilestoneSettler: already initialized");

        _milestones[dealId].amounts = amounts;
        _milestones[dealId].completed = new bool[](amounts.length);

        emit MilestoneInitialized(dealId, amounts.length);
    }

    /// @notice Mark a milestone as completed. Called by hub.
    function completeMilestone(uint256 dealId, uint256 index) external onlyHub {
        MilestoneData storage md = _milestones[dealId];
        require(index < md.amounts.length, "MilestoneSettler: invalid index");
        require(!md.completed[index], "MilestoneSettler: already completed");

        md.completed[index] = true;
        emit MilestoneCompleted(dealId, index, md.amounts[index]);
    }

    /// @inheritdoc ISettler
    function settle(uint256 dealId, address, address seller, address token, uint256 amount, bytes calldata) external override onlyHub {
        MilestoneData storage md = _milestones[dealId];
        uint256 releasable = _releasableAmount(md);
        require(releasable > 0, "MilestoneSettler: nothing to release");
        require(amount >= releasable, "MilestoneSettler: insufficient amount");

        md.releasedTotal += releasable;

        bool ok = IERC20(token).transfer(seller, releasable);
        require(ok, "MilestoneSettler: transfer failed");
    }

    /// @inheritdoc ISettler
    function canSettle(uint256 dealId) external view override returns (bool) {
        return _releasableAmount(_milestones[dealId]) > 0;
    }

    /// @notice Get the releasable amount for completed but unreleased milestones.
    function releasableAmount(uint256 dealId) external view returns (uint256) {
        return _releasableAmount(_milestones[dealId]);
    }

    /// @notice Get the amount for a specific milestone.
    function getMilestoneAmount(uint256 dealId, uint256 index) external view returns (uint256) {
        require(index < _milestones[dealId].amounts.length, "MilestoneSettler: invalid index");
        return _milestones[dealId].amounts[index];
    }

    /// @notice Get milestone data for a deal.
    function getMilestones(uint256 dealId)
        external
        view
        returns (uint256[] memory amounts, bool[] memory completed, uint256 releasedTotal)
    {
        MilestoneData storage md = _milestones[dealId];
        return (md.amounts, md.completed, md.releasedTotal);
    }

    function _releasableAmount(MilestoneData storage md) internal view returns (uint256) {
        uint256 completedTotal;
        for (uint256 i; i < md.amounts.length; ++i) {
            if (md.completed[i]) {
                completedTotal += md.amounts[i];
            }
        }
        return completedTotal - md.releasedTotal;
    }
}
