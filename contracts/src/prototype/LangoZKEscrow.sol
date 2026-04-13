// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import {IERC20} from "../interfaces/IERC20.sol";
import {IZKVerifier} from "../interfaces/IZKVerifier.sol";

/// @title LangoZKEscrow — Standalone ZK-gated escrow prototype.
/// @notice Minimal escrow that gates fund release on ZK proof verification.
/// @dev This is a standalone prototype — does NOT inherit from LangoEscrowHubV2.
///      It demonstrates the attestor-bound PQ attestation pattern where:
///      - On-chain verifies attestation validity (NOT PQ signature validity directly)
///      - Domain binding (dealId, chainId, contractAddress) prevents proof replay
///      - Trusted attestor allowlist restricts who can produce valid attestations
///
/// Trust model: the attestor is a trusted oracle that performed ML-DSA-65
/// verification off-chain. The ZK proof proves the attestor made the attestation.
/// Security property: non-repudiation of attestation.
contract LangoZKEscrow {
    // --- Errors ---
    error InvalidZKProof();
    error DomainBindingMismatch();
    error UntrustedAttestor();
    error DealNotFound();
    error DealAlreadyReleased();
    error DealAlreadyExists();
    error InsufficientDeposit();
    error TransferFailed();
    error NotBuyer();

    // --- Events ---
    event DealCreated(uint256 indexed dealId, address buyer, address seller, address token, uint256 amount);
    event DealReleased(uint256 indexed dealId, address seller, uint256 amount);

    // --- Constants: public input indices in PQAttestationCircuit ---
    // Order must match gnark circuit field ordering.
    uint256 constant IDX_ATTESTOR_DID_HASH = 0;
    uint256 constant IDX_MESSAGE_HASH = 1;
    uint256 constant IDX_PQ_PUBKEY_HASH = 2;
    uint256 constant IDX_TIMESTAMP = 3;
    uint256 constant IDX_MIN_TIMESTAMP = 4;
    uint256 constant IDX_DEAL_ID = 5;
    uint256 constant IDX_CHAIN_ID = 6;
    uint256 constant IDX_CONTRACT_ADDR = 7;
    uint256 constant NUM_PUBLIC_INPUTS = 8;

    // --- State ---
    struct Deal {
        address buyer;
        address seller;
        address token;
        uint256 amount;
        bool released;
    }

    mapping(uint256 => Deal) public deals;
    mapping(uint256 => bool) public trustedAttestors; // attestorDIDHash => trusted
    address public owner;
    address public immutable zkVerifier; // pinned Groth16 verifier — never caller-supplied

    constructor(uint256[] memory trustedAttestorHashes_, address verifier_) {
        require(verifier_ != address(0), "zero verifier");
        owner = msg.sender;
        zkVerifier = verifier_;
        for (uint256 i = 0; i < trustedAttestorHashes_.length; i++) {
            trustedAttestors[trustedAttestorHashes_[i]] = true;
        }
    }

    /// @notice Add a trusted attestor hash (owner only).
    function addTrustedAttestor(uint256 attestorHash) external {
        require(msg.sender == owner, "not owner");
        trustedAttestors[attestorHash] = true;
    }

    /// @notice Create a deal and deposit funds.
    function createDeal(
        uint256 dealId,
        address seller,
        address token,
        uint256 amount
    ) external {
        if (deals[dealId].buyer != address(0)) revert DealAlreadyExists();
        if (amount == 0) revert InsufficientDeposit();

        deals[dealId] = Deal({
            buyer: msg.sender,
            seller: seller,
            token: token,
            amount: amount,
            released: false
        });

        bool ok = IERC20(token).transferFrom(msg.sender, address(this), amount);
        if (!ok) revert TransferFailed();

        emit DealCreated(dealId, msg.sender, seller, token, amount);
    }

    /// @notice Release funds to seller with ZK proof verification.
    /// @param dealId The deal identifier (must match proof's domain binding).
    /// @param proof Compressed Groth16 proof [a.x, a.y, b.x0, b.x1, b.y0, b.y1, c.x, c.y].
    /// @param publicInputs Public inputs to the PQAttestationCircuit (8 field elements).
    function releaseWithProof(
        uint256 dealId,
        uint256[8] calldata proof,
        uint256[8] calldata publicInputs
    ) external {
        Deal storage deal = deals[dealId];
        if (deal.buyer == address(0)) revert DealNotFound();
        if (deal.released) revert DealAlreadyReleased();

        // Domain binding: check deal/chain/contract match on-chain state.
        if (publicInputs[IDX_DEAL_ID] != dealId) revert DomainBindingMismatch();
        if (publicInputs[IDX_CHAIN_ID] != block.chainid) revert DomainBindingMismatch();
        if (publicInputs[IDX_CONTRACT_ADDR] != uint256(uint160(address(this)))) revert DomainBindingMismatch();

        // Attestor trust: check attestor is in allowlist.
        if (!trustedAttestors[publicInputs[IDX_ATTESTOR_DID_HASH]]) revert UntrustedAttestor();

        // ZK proof verification via external verifier contract.
        // gnark's verifyProof reverts with ProofInvalid() on failure (no bool return).
        // If this call does not revert, the proof is valid.
        try IZKVerifier(zkVerifier).verifyProof(proof, publicInputs) {
            // Proof valid — continue to release.
        } catch {
            revert InvalidZKProof();
        }

        // Release funds.
        deal.released = true;
        bool ok = IERC20(deal.token).transfer(deal.seller, deal.amount);
        if (!ok) revert TransferFailed();

        emit DealReleased(dealId, deal.seller, deal.amount);
    }

    /// @notice Refund funds to buyer (buyer only, for unreleased deals).
    function refund(uint256 dealId) external {
        Deal storage deal = deals[dealId];
        if (deal.buyer == address(0)) revert DealNotFound();
        if (deal.released) revert DealAlreadyReleased();
        if (msg.sender != deal.buyer) revert NotBuyer();

        deal.released = true;
        bool ok = IERC20(deal.token).transfer(deal.buyer, deal.amount);
        if (!ok) revert TransferFailed();
    }
}
