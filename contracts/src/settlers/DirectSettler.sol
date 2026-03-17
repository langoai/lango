// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "../interfaces/ISettler.sol";
import "../interfaces/IERC20.sol";

/// @title DirectSettler — Immediate transfer settlement with no escrow period.
/// @notice Receives funds from the hub and transfers them directly to the seller.
contract DirectSettler is ISettler {
    /// @inheritdoc ISettler
    function settle(uint256, address, address seller, address token, uint256 amount, bytes calldata) external override {
        // The hub transfers tokens to this contract before calling settle().
        // Forward everything to the seller.
        bool ok = IERC20(token).transfer(seller, amount);
        require(ok, "DirectSettler: transfer failed");
    }

    /// @inheritdoc ISettler
    function canSettle(uint256) external pure override returns (bool) {
        return true;
    }
}
