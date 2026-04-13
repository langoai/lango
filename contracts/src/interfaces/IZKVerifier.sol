// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/// @title IZKVerifier — Interface for Groth16 ZK proof verification.
/// @notice Verifier contracts exported by gnark implement this interface.
/// @dev Uses gnark's compressed Groth16 format:
///      proof[8] = [a.x, a.y, b.x0, b.x1, b.y0, b.y1, c.x, c.y]
///      input[N] = public inputs as field elements.
///      Reverts with ProofInvalid() on verification failure (no bool return).
interface IZKVerifier {
    /// @notice Verifies a Groth16 proof against the circuit's verifying key.
    /// @param proof Compressed proof: [a.x, a.y, b.x0, b.x1, b.y0, b.y1, c.x, c.y].
    /// @param input Public inputs as BN254 field elements.
    /// @dev Reverts if the proof is invalid. Does not return false.
    function verifyProof(
        uint256[8] calldata proof,
        uint256[8] calldata input
    ) external view;
}
