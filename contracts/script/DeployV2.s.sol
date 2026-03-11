// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "forge-std/Script.sol";
import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import {UpgradeableBeacon} from "@openzeppelin/contracts/proxy/beacon/UpgradeableBeacon.sol";
import "../src/LangoEscrowHubV2.sol";
import "../src/LangoVaultV2.sol";
import "../src/LangoBeaconVaultFactory.sol";
import "../src/settlers/DirectSettler.sol";
import "../src/settlers/MilestoneSettler.sol";

/// @title DeployV2 — Deploy all V2 Lango economy contracts.
/// @notice Deploys UUPS proxy for EscrowHubV2, Beacon + Factory for VaultV2, and settler contracts.
contract DeployV2Script is Script {
    function run() external {
        vm.startBroadcast();

        address deployer = msg.sender;

        // 1. Deploy LangoEscrowHubV2 implementation
        LangoEscrowHubV2 hubImpl = new LangoEscrowHubV2();
        console.log("EscrowHubV2 implementation:", address(hubImpl));

        // 2. Deploy ERC1967Proxy pointing to LangoEscrowHubV2
        bytes memory hubInitData = abi.encodeCall(LangoEscrowHubV2.initialize, (deployer));
        ERC1967Proxy hubProxy = new ERC1967Proxy(address(hubImpl), hubInitData);
        address hubProxyAddr = address(hubProxy);
        console.log("EscrowHubV2 proxy:", hubProxyAddr);

        // 3. Verify initialization
        LangoEscrowHubV2 hub = LangoEscrowHubV2(hubProxyAddr);
        require(hub.owner() == deployer, "DeployV2: hub owner mismatch");

        // 4. Deploy DirectSettler
        DirectSettler directSettler = new DirectSettler();
        console.log("DirectSettler:", address(directSettler));

        // 5. Deploy MilestoneSettler (linked to hub proxy)
        MilestoneSettler milestoneSettler = new MilestoneSettler(hubProxyAddr);
        console.log("MilestoneSettler:", address(milestoneSettler));

        // 6. Register settlers on hub
        hub.registerSettler(keccak256("direct"), address(directSettler));
        hub.registerSettler(keccak256("milestone"), address(milestoneSettler));
        console.log("Settlers registered");

        // 7. Deploy LangoVaultV2 implementation
        LangoVaultV2 vaultImpl = new LangoVaultV2();
        console.log("VaultV2 implementation:", address(vaultImpl));

        // 8. Deploy UpgradeableBeacon pointing to LangoVaultV2
        UpgradeableBeacon vaultBeacon = new UpgradeableBeacon(address(vaultImpl), deployer);
        console.log("VaultV2 beacon:", address(vaultBeacon));

        // 9. Deploy LangoBeaconVaultFactory with beacon address
        LangoBeaconVaultFactory vaultFactory = new LangoBeaconVaultFactory(address(vaultBeacon), deployer);
        console.log("BeaconVaultFactory:", address(vaultFactory));

        // 10. Transfer beacon ownership to factory for upgrades via factory
        vaultBeacon.transferOwnership(address(vaultFactory));
        console.log("Beacon ownership transferred to factory");

        vm.stopBroadcast();

        // Write deployment addresses to JSON
        string memory obj = "v2deployment";
        vm.serializeAddress(obj, "deployer", deployer);
        vm.serializeUint(obj, "chainId", block.chainid);
        vm.serializeAddress(obj, "escrowHubV2Implementation", address(hubImpl));
        vm.serializeAddress(obj, "escrowHubV2Proxy", hubProxyAddr);
        vm.serializeAddress(obj, "directSettler", address(directSettler));
        vm.serializeAddress(obj, "milestoneSettler", address(milestoneSettler));
        vm.serializeAddress(obj, "vaultV2Implementation", address(vaultImpl));
        vm.serializeAddress(obj, "vaultV2Beacon", address(vaultBeacon));
        string memory json = vm.serializeAddress(obj, "beaconVaultFactory", address(vaultFactory));

        string memory path = string.concat("deployments/", vm.toString(block.chainid), "-v2.json");
        vm.writeJson(json, path);
        console.log("Deployment saved to:", path);
    }
}
