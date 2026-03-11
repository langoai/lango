// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "forge-std/Script.sol";
import "../src/LangoEscrowHubV2.sol";
import "../src/LangoVaultV2.sol";
import "../src/LangoBeaconVaultFactory.sol";

/// @title UpgradeV2 — Template for upgrading V2 contracts.
/// @notice Upgrades EscrowHubV2 (UUPS) and/or VaultV2 (Beacon) to new implementations.
contract UpgradeV2Script is Script {
    function run() external {
        // Read addresses from environment
        address hubProxy = vm.envAddress("HUB_PROXY");
        address vaultFactory = vm.envAddress("VAULT_FACTORY");
        bool upgradeHub = vm.envOr("UPGRADE_HUB", false);
        bool upgradeVault = vm.envOr("UPGRADE_VAULT", false);

        vm.startBroadcast();

        // 1. Upgrade EscrowHubV2 (UUPS)
        if (upgradeHub) {
            LangoEscrowHubV2 newHubImpl = new LangoEscrowHubV2();
            console.log("New EscrowHubV2 implementation:", address(newHubImpl));

            LangoEscrowHubV2 hub = LangoEscrowHubV2(hubProxy);
            hub.upgradeToAndCall(address(newHubImpl), "");
            console.log("EscrowHubV2 upgraded successfully");
        }

        // 2. Upgrade VaultV2 (Beacon via Factory)
        if (upgradeVault) {
            LangoVaultV2 newVaultImpl = new LangoVaultV2();
            console.log("New VaultV2 implementation:", address(newVaultImpl));

            LangoBeaconVaultFactory factory = LangoBeaconVaultFactory(vaultFactory);
            factory.upgradeImplementation(address(newVaultImpl));
            console.log("VaultV2 beacon upgraded successfully (all vaults updated)");
        }

        vm.stopBroadcast();
    }
}
