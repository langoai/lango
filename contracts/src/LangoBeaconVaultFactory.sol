// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {UpgradeableBeacon} from "@openzeppelin/contracts/proxy/beacon/UpgradeableBeacon.sol";
import {BeaconProxy} from "@openzeppelin/contracts/proxy/beacon/BeaconProxy.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import "./LangoVaultV2.sol";

/// @title LangoBeaconVaultFactory — Creates BeaconProxy vaults pointing to a shared UpgradeableBeacon.
/// @notice Owner can upgrade the beacon implementation, upgrading all vaults simultaneously.
contract LangoBeaconVaultFactory is Ownable {
    UpgradeableBeacon public immutable beacon;

    uint256 public vaultCount;
    mapping(uint256 => address) public vaults;

    event VaultCreated(address indexed vault, bytes32 indexed refId, address indexed buyer, address seller);

    constructor(address beaconAddress, address owner_) Ownable(owner_) {
        require(beaconAddress != address(0), "Factory: zero beacon");
        beacon = UpgradeableBeacon(beaconAddress);
    }

    /// @notice Create a new vault via BeaconProxy and initialize it.
    function createVault(address seller, address token, uint256 amount, address arbiter, bytes32 refId)
        external
        returns (address vault)
    {
        bytes memory initData = abi.encodeCall(
            LangoVaultV2.initialize, (msg.sender, seller, token, amount, arbiter, refId)
        );

        BeaconProxy proxy = new BeaconProxy(address(beacon), initData);
        vault = address(proxy);

        uint256 vaultId = vaultCount++;
        vaults[vaultId] = vault;

        emit VaultCreated(vault, refId, msg.sender, seller);
    }

    /// @notice Upgrade the beacon implementation (upgrades ALL vaults).
    function upgradeImplementation(address newImpl) external onlyOwner {
        beacon.upgradeTo(newImpl);
    }

    /// @notice Get vault address by ID.
    function getVault(uint256 vaultId) external view returns (address) {
        return vaults[vaultId];
    }
}
