package circuits

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
)

// ResponseAttestationCircuit proves that an agent produced a response derived
// from specific source data, without revealing the source data or agent key.
//
// Public inputs: ResponseHash, AgentDIDHash, Timestamp, MinTimestamp, MaxTimestamp
// Private witness: SourceDataHash, AgentKeyProof
//
// Constraints:
//   - MiMC(AgentKeyProof) == AgentDIDHash
//   - MiMC(SourceDataHash, AgentKeyProof, Timestamp) == ResponseHash
//   - MinTimestamp <= Timestamp <= MaxTimestamp (freshness)
type ResponseAttestationCircuit struct {
	ResponseHash   frontend.Variable `gnark:",public"`
	AgentDIDHash   frontend.Variable `gnark:",public"`
	Timestamp      frontend.Variable `gnark:",public"`
	MinTimestamp   frontend.Variable `gnark:",public"`
	MaxTimestamp   frontend.Variable `gnark:",public"`

	SourceDataHash frontend.Variable `gnark:""`
	AgentKeyProof  frontend.Variable `gnark:""`
}

// Define implements frontend.Circuit and constrains the attestation proof.
func (c *ResponseAttestationCircuit) Define(api frontend.API) error {
	// Prove agent authority: MiMC(AgentKeyProof) == AgentDIDHash
	hAgent, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}
	hAgent.Write(c.AgentKeyProof)
	computedDID := hAgent.Sum()
	api.AssertIsEqual(computedDID, c.AgentDIDHash)

	// Prove response derivation: MiMC(SourceDataHash, AgentKeyProof, Timestamp) == ResponseHash
	hResp, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}
	hResp.Write(c.SourceDataHash, c.AgentKeyProof, c.Timestamp)
	computedResp := hResp.Sum()
	api.AssertIsEqual(computedResp, c.ResponseHash)

	// Prove timestamp freshness: MinTimestamp <= Timestamp <= MaxTimestamp
	api.AssertIsLessOrEqual(c.MinTimestamp, c.Timestamp)
	api.AssertIsLessOrEqual(c.Timestamp, c.MaxTimestamp)

	return nil
}
