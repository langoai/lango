// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Script.sol";
import "../src/LangoEscrowHub.sol";
import "../src/LangoVault.sol";
import "../src/LangoVaultFactory.sol";
import "../src/modules/LangoSessionValidator.sol";
import "../src/modules/LangoSpendingHook.sol";
import "../src/modules/LangoEscrowExecutor.sol";
import "../test/mocks/MockUSDC.sol";

/// @title Deploy — deploy all Lango infrastructure contracts.
/// @notice Outputs deployed addresses to deployments/<chainId>.json.
contract DeployScript is Script {
    // Base Sepolia canonical USDC
    address constant CANONICAL_USDC = 0x036CbD53842c5426634e7929541eC2318f3dCF7e;

    function run() external {
        bool deployMockUsdc = vm.envOr("DEPLOY_MOCK_USDC", false);

        // Signing method is determined by CLI flags:
        //   --account <name>   → Foundry encrypted keystore (recommended)
        //   --interactive      → prompt for private key at runtime
        //   --ledger / --trezor→ hardware wallet
        //   --private-key $KEY → direct key (CI only)
        vm.startBroadcast();

        address deployer = msg.sender;

        // 1. Token — MockUSDC or canonical
        address tokenAddress;
        if (deployMockUsdc) {
            MockUSDC mockUsdc = new MockUSDC();
            tokenAddress = address(mockUsdc);
        } else {
            tokenAddress = CANONICAL_USDC;
        }

        // 2. Escrow Hub — deployer is testnet arbitrator
        LangoEscrowHub escrowHub = new LangoEscrowHub(deployer);

        // 3. Vault implementation (clone target, no constructor args)
        LangoVault vaultImpl = new LangoVault();

        // 4. Vault Factory — needs vault implementation address
        LangoVaultFactory vaultFactory = new LangoVaultFactory(address(vaultImpl));

        // 5. ERC-7579 Modules (no constructor args)
        LangoSessionValidator sessionValidator = new LangoSessionValidator();
        LangoSpendingHook spendingHook = new LangoSpendingHook();
        LangoEscrowExecutor escrowExecutor = new LangoEscrowExecutor();

        vm.stopBroadcast();

        // Write deployment addresses to JSON
        string memory obj = "deployment";
        vm.serializeAddress(obj, "deployer", deployer);
        vm.serializeUint(obj, "chainId", block.chainid);
        vm.serializeAddress(obj, "tokenAddress", tokenAddress);
        vm.serializeAddress(obj, "escrowHub", address(escrowHub));
        vm.serializeAddress(obj, "vaultImplementation", address(vaultImpl));
        vm.serializeAddress(obj, "vaultFactory", address(vaultFactory));
        vm.serializeAddress(obj, "sessionValidator", address(sessionValidator));
        vm.serializeAddress(obj, "spendingHook", address(spendingHook));
        string memory json = vm.serializeAddress(obj, "escrowExecutor", address(escrowExecutor));

        string memory path = string.concat("deployments/", vm.toString(block.chainid), ".json");
        vm.writeJson(json, path);
    }
}
