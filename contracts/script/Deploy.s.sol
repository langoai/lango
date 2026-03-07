// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

// Foundry Script — deploy all contracts to local Anvil or testnet.
// Usage: forge script script/Deploy.s.sol --rpc-url http://localhost:8545 --broadcast

// NOTE: This script requires forge-std. Install with: forge install foundry-rs/forge-std
// import "forge-std/Script.sol";
// import "../src/LangoEscrowHub.sol";
// import "../src/LangoVault.sol";
// import "../src/LangoVaultFactory.sol";
// import "../test/mocks/MockUSDC.sol";

// contract DeployScript is Script {
//     function run() external {
//         uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
//         vm.startBroadcast(deployerPrivateKey);
//
//         // 1. Deploy MockUSDC (testnet only)
//         MockUSDC usdc = new MockUSDC();
//
//         // 2. Deploy Hub with deployer as arbitrator
//         LangoEscrowHub hub = new LangoEscrowHub(msg.sender);
//
//         // 3. Deploy Vault implementation + Factory
//         LangoVault vaultImpl = new LangoVault();
//         LangoVaultFactory factory = new LangoVaultFactory(address(vaultImpl));
//
//         vm.stopBroadcast();
//     }
// }

// Placeholder — uncomment above after `forge install foundry-rs/forge-std`
contract DeployScript {}
