// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./LangoVault.sol";

/// @title LangoVaultFactory — EIP-1167 Minimal Proxy factory for LangoVault.
/// @notice Creates lightweight clones of the LangoVault implementation.
contract LangoVaultFactory {
    address public immutable implementation;
    uint256 public vaultCount;

    mapping(uint256 => address) public vaults;

    event VaultCreated(uint256 indexed vaultId, address indexed vault, address indexed buyer, address seller);

    constructor(address _implementation) {
        require(_implementation != address(0), "Factory: zero implementation");
        implementation = _implementation;
    }

    /// @notice Create a new vault clone and initialize it.
    function createVault(
        address seller,
        address token,
        uint256 amount,
        uint256 deadline,
        address arbitrator
    ) external returns (uint256 vaultId, address vault) {
        vaultId = vaultCount++;
        vault = _clone(implementation);
        vaults[vaultId] = vault;

        LangoVault(vault).initialize(
            msg.sender,
            seller,
            token,
            amount,
            deadline,
            arbitrator
        );

        emit VaultCreated(vaultId, vault, msg.sender, seller);
    }

    /// @notice Get vault address by ID.
    function getVault(uint256 vaultId) external view returns (address) {
        return vaults[vaultId];
    }

    /// @dev EIP-1167 Minimal Proxy clone.
    function _clone(address impl) internal returns (address instance) {
        assembly {
            let ptr := mload(0x40)
            mstore(ptr, 0x3d602d80600a3d3981f3363d3d373d3d3d363d73000000000000000000000000)
            mstore(add(ptr, 0x14), shl(0x60, impl))
            mstore(add(ptr, 0x28), 0x5af43d82803e903d91602b57fd5bf30000000000000000000000000000000000)
            instance := create(0, ptr, 0x37)
            if iszero(instance) { revert(0, 0) }
        }
    }
}
