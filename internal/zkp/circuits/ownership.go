// Package circuits provides gnark circuit definitions for zero-knowledge proofs
// used in agent identity, attestation, and capability verification.
package circuits

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
)

// WalletOwnershipCircuit proves knowledge of a secret response that, when
// hashed with a public challenge, produces the expected public key hash.
// This is a simplified commitment scheme using MiMC:
//
//	MiMC(Response, Challenge) == PublicKeyHash
type WalletOwnershipCircuit struct {
	// Public inputs
	PublicKeyHash frontend.Variable `gnark:",public"`
	Challenge     frontend.Variable `gnark:",public"`

	// Private witness
	Response frontend.Variable `gnark:""`
}

// Define implements frontend.Circuit and constrains the ownership proof.
func (c *WalletOwnershipCircuit) Define(api frontend.API) error {
	h, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}

	h.Write(c.Response, c.Challenge)
	computed := h.Sum()

	api.AssertIsEqual(computed, c.PublicKeyHash)
	return nil
}
