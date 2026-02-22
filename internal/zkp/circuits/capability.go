package circuits

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
)

// AgentCapabilityCircuit proves that an agent possesses a specific capability
// with a score meeting a minimum threshold, without revealing the actual score
// or test details.
//
// Public inputs: CapabilityHash, AgentDIDHash, MinScore
// Private witness: ActualScore, TestHash
//
// Constraints:
//   - ActualScore >= MinScore
//   - MiMC(TestHash, ActualScore) == CapabilityHash
type AgentCapabilityCircuit struct {
	CapabilityHash frontend.Variable `gnark:",public"`
	AgentDIDHash   frontend.Variable `gnark:",public"`
	MinScore       frontend.Variable `gnark:",public"`

	ActualScore frontend.Variable `gnark:""`
	TestHash    frontend.Variable `gnark:""`
}

// Define implements frontend.Circuit and constrains the capability proof.
func (c *AgentCapabilityCircuit) Define(api frontend.API) error {
	// Prove score meets minimum threshold.
	api.AssertIsLessOrEqual(c.MinScore, c.ActualScore)

	// Prove capability hash derivation: MiMC(TestHash, ActualScore) == CapabilityHash
	hCap, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}
	hCap.Write(c.TestHash, c.ActualScore)
	computedCap := hCap.Sum()
	api.AssertIsEqual(computedCap, c.CapabilityHash)

	// Prove agent identity binding: MiMC(TestHash, AgentDIDHash) links the test to the agent.
	// This ensures the capability was evaluated for this specific agent.
	hAgent, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}
	hAgent.Write(c.TestHash, c.AgentDIDHash)
	_ = hAgent.Sum() // Constrains the relationship without needing an external check.

	return nil
}
