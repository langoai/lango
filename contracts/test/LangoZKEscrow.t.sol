// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "../src/prototype/LangoZKEscrow.sol";
import "../src/interfaces/IZKVerifier.sol";
import "./mocks/MockUSDC.sol";

/// @dev Mock verifier that always succeeds (does not revert).
contract MockVerifierValid is IZKVerifier {
    function verifyProof(uint256[8] calldata, uint256[8] calldata) external pure override {
        // No revert = proof valid.
    }
}

/// @dev Mock verifier that always fails (reverts).
contract MockVerifierInvalid is IZKVerifier {
    error ProofInvalid();
    function verifyProof(uint256[8] calldata, uint256[8] calldata) external pure override {
        revert ProofInvalid();
    }
}

contract LangoZKEscrowTest is Test {
    LangoZKEscrow public escrow;
    MockUSDC public usdc;
    MockVerifierValid public validVerifier;
    MockVerifierInvalid public invalidVerifier;

    address public buyer = address(0xB);
    address public seller = address(0xC);
    address public stranger = address(0xD);

    uint256 public constant AMOUNT = 1000e6;
    uint256 public constant DEAL_ID = 1;
    uint256 public constant ATTESTOR_HASH = 0x42;

    function setUp() public {
        usdc = new MockUSDC();
        validVerifier = new MockVerifierValid();
        invalidVerifier = new MockVerifierInvalid();

        // Deploy escrow with one trusted attestor.
        uint256[] memory attestors = new uint256[](1);
        attestors[0] = ATTESTOR_HASH;
        escrow = new LangoZKEscrow(attestors);

        // Fund buyer.
        usdc.mint(buyer, 100_000e6);
        vm.prank(buyer);
        usdc.approve(address(escrow), type(uint256).max);
    }

    // --- Helper: build public inputs matching circuit field order ---
    function _buildPublicInputs(
        uint256 attestorHash,
        uint256 dealId,
        uint256 chainId,
        uint256 contractAddr
    ) internal pure returns (uint256[8] memory inputs) {
        inputs[0] = attestorHash;       // AttestorDIDHash
        inputs[1] = 0xAABB;            // MessageHash
        inputs[2] = 0xCCDD;            // PQPublicKeyHash
        inputs[3] = 1000;              // Timestamp
        inputs[4] = 900;               // MinTimestamp
        inputs[5] = dealId;            // DealID
        inputs[6] = chainId;           // ChainID
        inputs[7] = contractAddr;      // ContractAddress
    }

    function _dummyProof() internal pure returns (uint256[8] memory) {
        // Dummy proof bytes — mock verifier ignores content.
        return [uint256(1), 2, 3, 4, 5, 6, 7, 8];
    }

    // ==================== Deal Lifecycle ====================

    function testCreateDeal() public {
        vm.prank(buyer);
        escrow.createDeal(DEAL_ID, seller, address(usdc), AMOUNT);

        (address b, address s, address t, uint256 a, bool r) = escrow.deals(DEAL_ID);
        assertEq(b, buyer);
        assertEq(s, seller);
        assertEq(t, address(usdc));
        assertEq(a, AMOUNT);
        assertFalse(r);
        assertEq(usdc.balanceOf(address(escrow)), AMOUNT);
    }

    function testCreateDealDuplicate() public {
        vm.prank(buyer);
        escrow.createDeal(DEAL_ID, seller, address(usdc), AMOUNT);

        vm.prank(buyer);
        vm.expectRevert(LangoZKEscrow.DealAlreadyExists.selector);
        escrow.createDeal(DEAL_ID, seller, address(usdc), AMOUNT);
    }

    function testCreateDealZeroAmount() public {
        vm.prank(buyer);
        vm.expectRevert(LangoZKEscrow.InsufficientDeposit.selector);
        escrow.createDeal(DEAL_ID, seller, address(usdc), 0);
    }

    // ==================== Release with ZK Proof ====================

    function testReleaseWithValidProof() public {
        vm.prank(buyer);
        escrow.createDeal(DEAL_ID, seller, address(usdc), AMOUNT);

        uint256[8] memory inputs = _buildPublicInputs(
            ATTESTOR_HASH,
            DEAL_ID,
            block.chainid,
            uint256(uint160(address(escrow)))
        );

        escrow.releaseWithProof(DEAL_ID, _dummyProof(), inputs, address(validVerifier));

        (, , , , bool released) = escrow.deals(DEAL_ID);
        assertTrue(released);
        assertEq(usdc.balanceOf(seller), AMOUNT);
    }

    function testReleaseAlreadyReleased() public {
        vm.prank(buyer);
        escrow.createDeal(DEAL_ID, seller, address(usdc), AMOUNT);

        uint256[8] memory inputs = _buildPublicInputs(
            ATTESTOR_HASH, DEAL_ID, block.chainid, uint256(uint160(address(escrow)))
        );
        escrow.releaseWithProof(DEAL_ID, _dummyProof(), inputs, address(validVerifier));

        vm.expectRevert(LangoZKEscrow.DealAlreadyReleased.selector);
        escrow.releaseWithProof(DEAL_ID, _dummyProof(), inputs, address(validVerifier));
    }

    function testReleaseNonexistentDeal() public {
        uint256[8] memory inputs = _buildPublicInputs(
            ATTESTOR_HASH, 999, block.chainid, uint256(uint160(address(escrow)))
        );
        vm.expectRevert(LangoZKEscrow.DealNotFound.selector);
        escrow.releaseWithProof(999, _dummyProof(), inputs, address(validVerifier));
    }

    // ==================== Domain Binding ====================

    function testDomainBindingDealIdMismatch() public {
        vm.prank(buyer);
        escrow.createDeal(DEAL_ID, seller, address(usdc), AMOUNT);

        uint256[8] memory inputs = _buildPublicInputs(
            ATTESTOR_HASH,
            999, // wrong deal ID
            block.chainid,
            uint256(uint160(address(escrow)))
        );

        vm.expectRevert(LangoZKEscrow.DomainBindingMismatch.selector);
        escrow.releaseWithProof(DEAL_ID, _dummyProof(), inputs, address(validVerifier));
    }

    function testDomainBindingChainIdMismatch() public {
        vm.prank(buyer);
        escrow.createDeal(DEAL_ID, seller, address(usdc), AMOUNT);

        uint256[8] memory inputs = _buildPublicInputs(
            ATTESTOR_HASH,
            DEAL_ID,
            999, // wrong chain ID
            uint256(uint160(address(escrow)))
        );

        vm.expectRevert(LangoZKEscrow.DomainBindingMismatch.selector);
        escrow.releaseWithProof(DEAL_ID, _dummyProof(), inputs, address(validVerifier));
    }

    function testDomainBindingContractMismatch() public {
        vm.prank(buyer);
        escrow.createDeal(DEAL_ID, seller, address(usdc), AMOUNT);

        uint256[8] memory inputs = _buildPublicInputs(
            ATTESTOR_HASH,
            DEAL_ID,
            block.chainid,
            uint256(uint160(address(0xDEAD))) // wrong contract
        );

        vm.expectRevert(LangoZKEscrow.DomainBindingMismatch.selector);
        escrow.releaseWithProof(DEAL_ID, _dummyProof(), inputs, address(validVerifier));
    }

    // ==================== Attestor Trust ====================

    function testUntrustedAttestorRejected() public {
        vm.prank(buyer);
        escrow.createDeal(DEAL_ID, seller, address(usdc), AMOUNT);

        uint256[8] memory inputs = _buildPublicInputs(
            0xBAD, // untrusted attestor hash
            DEAL_ID,
            block.chainid,
            uint256(uint160(address(escrow)))
        );

        vm.expectRevert(LangoZKEscrow.UntrustedAttestor.selector);
        escrow.releaseWithProof(DEAL_ID, _dummyProof(), inputs, address(validVerifier));
    }

    function testAddTrustedAttestor() public {
        uint256 newAttestor = 0xABCDEF;
        escrow.addTrustedAttestor(newAttestor);
        assertTrue(escrow.trustedAttestors(newAttestor));
    }

    function testAddTrustedAttestorNotOwner() public {
        vm.prank(stranger);
        vm.expectRevert("not owner");
        escrow.addTrustedAttestor(0x123);
    }

    // ==================== Invalid ZK Proof ====================

    function testInvalidProofRejected() public {
        vm.prank(buyer);
        escrow.createDeal(DEAL_ID, seller, address(usdc), AMOUNT);

        uint256[8] memory inputs = _buildPublicInputs(
            ATTESTOR_HASH, DEAL_ID, block.chainid, uint256(uint160(address(escrow)))
        );

        vm.expectRevert(LangoZKEscrow.InvalidZKProof.selector);
        escrow.releaseWithProof(DEAL_ID, _dummyProof(), inputs, address(invalidVerifier));
    }

    // ==================== Refund ====================

    function testRefundByBuyer() public {
        vm.prank(buyer);
        escrow.createDeal(DEAL_ID, seller, address(usdc), AMOUNT);

        uint256 balBefore = usdc.balanceOf(buyer);
        vm.prank(buyer);
        escrow.refund(DEAL_ID);

        assertEq(usdc.balanceOf(buyer), balBefore + AMOUNT);
        (, , , , bool released) = escrow.deals(DEAL_ID);
        assertTrue(released);
    }

    function testRefundNotBuyer() public {
        vm.prank(buyer);
        escrow.createDeal(DEAL_ID, seller, address(usdc), AMOUNT);

        vm.prank(stranger);
        vm.expectRevert(LangoZKEscrow.NotBuyer.selector);
        escrow.refund(DEAL_ID);
    }
}
