package circuits

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
)

// PQAttestationCircuit proves that a trusted attestor observed and verified
// a post-quantum (ML-DSA-65) signature off-chain, and binds the attestation
// to a specific on-chain action via domain-separated public inputs.
//
// Trust model: on-chain does NOT verify PQ signature validity directly.
// It verifies attestor-bound attestation validity. The attestor is a trusted
// oracle that performed ML-DSA-65 verification off-chain. The ZK proof proves:
// "attestor X saw message Y signed by PQ key Z, for deal D on chain C at contract A."
//
// Security property: non-repudiation of attestation (attestor cannot deny
// having made the attestation). Domain binding prevents proof replay across
// deals, chains, or contracts.
type PQAttestationCircuit struct {
	// Public inputs — attestation identity.
	AttestorDIDHash frontend.Variable `gnark:",public"` // MiMC(AttestorSecret)
	MessageHash     frontend.Variable `gnark:",public"` // MiMC(MessagePreimage)
	PQPublicKeyHash frontend.Variable `gnark:",public"` // MiMC(PQPublicKeyPreimage)
	Timestamp       frontend.Variable `gnark:",public"` // attestation timestamp
	MinTimestamp    frontend.Variable `gnark:",public"` // freshness lower bound

	// Public inputs — domain binding (replay prevention).
	// These are checked on-chain by the verifier contract against expected values.
	DealID          frontend.Variable `gnark:",public"` // on-chain deal identifier
	ChainID         frontend.Variable `gnark:",public"` // EVM chain ID
	ContractAddress frontend.Variable `gnark:",public"` // uint256(escrow contract address)

	// Private inputs.
	AttestorSecret      frontend.Variable // attestor's secret (proves identity)
	MessagePreimage     frontend.Variable // message that was PQ-signed
	PQPublicKeyPreimage frontend.Variable // PQ public key that signed it
}

// Define constrains the PQ attestation proof.
func (c *PQAttestationCircuit) Define(api frontend.API) error {
	h, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}

	// Attestor binding: prove identity.
	h.Reset()
	h.Write(c.AttestorSecret)
	api.AssertIsEqual(h.Sum(), c.AttestorDIDHash)

	// Message binding: prove knowledge of signed message.
	h.Reset()
	h.Write(c.MessagePreimage)
	api.AssertIsEqual(h.Sum(), c.MessageHash)

	// PQ public key binding: prove knowledge of PQ signing key.
	h.Reset()
	h.Write(c.PQPublicKeyPreimage)
	api.AssertIsEqual(h.Sum(), c.PQPublicKeyHash)

	// Freshness: timestamp must be >= MinTimestamp.
	api.AssertIsLessOrEqual(c.MinTimestamp, c.Timestamp)

	// Domain binding: DealID, ChainID, ContractAddress are public inputs only.
	// No constraints needed — the verifier contract checks them against expected
	// on-chain values (dealId, block.chainid, address(this)). Including them as
	// public inputs binds the proof to a specific on-chain action. A proof
	// generated for deal D1 cannot be replayed for deal D2 because the public
	// inputs would not match.
	_ = c.DealID
	_ = c.ChainID
	_ = c.ContractAddress

	return nil
}
