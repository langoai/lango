// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title ISettler — Settlement strategy interface for escrow deals.
interface ISettler {
    /// @notice Execute settlement for a deal.
    /// @param dealId  The deal identifier.
    /// @param buyer   The buyer address.
    /// @param seller  The seller address.
    /// @param token   The ERC-20 token address for payment.
    /// @param amount  The total deal amount.
    /// @param data    Settler-specific encoded data.
    function settle(uint256 dealId, address buyer, address seller, address token, uint256 amount, bytes calldata data) external;

    /// @notice Check if a deal can be settled.
    /// @param dealId The deal identifier.
    /// @return True if settlement conditions are met.
    function canSettle(uint256 dealId) external view returns (bool);
}
