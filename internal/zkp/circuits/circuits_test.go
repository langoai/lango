package circuits

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	nativemimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark/test"
)

// mimcHash computes the native MiMC hash of the given field elements.
// Each element is written as a 32-byte big-endian representation.
func mimcHash(elems ...*big.Int) *big.Int {
	h := nativemimc.NewMiMC()
	for _, e := range elems {
		var elem fr.Element
		elem.SetBigInt(e)
		b := elem.Bytes() // 32-byte big-endian
		h.Write(b[:])
	}
	result := h.Sum(nil)
	var out big.Int
	out.SetBytes(result)
	return &out
}

// --- WalletOwnership Tests ---

func TestWalletOwnership_Valid(t *testing.T) {
	assert := test.NewAssert(t)

	response := big.NewInt(42)
	challenge := big.NewInt(123)
	publicKeyHash := mimcHash(response, challenge)

	circuit := &WalletOwnershipCircuit{}
	assignment := &WalletOwnershipCircuit{
		PublicKeyHash: publicKeyHash,
		Challenge:     challenge,
		Response:      response,
	}

	assert.ProverSucceeded(circuit, assignment, test.WithCurves(ecc.BN254))
}

func TestWalletOwnership_InvalidResponse(t *testing.T) {
	assert := test.NewAssert(t)

	response := big.NewInt(42)
	challenge := big.NewInt(123)
	publicKeyHash := mimcHash(response, challenge)

	wrongResponse := big.NewInt(99)

	circuit := &WalletOwnershipCircuit{}
	assignment := &WalletOwnershipCircuit{
		PublicKeyHash: publicKeyHash,
		Challenge:     challenge,
		Response:      wrongResponse,
	}

	assert.ProverFailed(circuit, assignment, test.WithCurves(ecc.BN254))
}

func TestWalletOwnership_WrongChallenge(t *testing.T) {
	assert := test.NewAssert(t)

	response := big.NewInt(42)
	challenge := big.NewInt(123)
	publicKeyHash := mimcHash(response, challenge)

	wrongChallenge := big.NewInt(456)

	circuit := &WalletOwnershipCircuit{}
	assignment := &WalletOwnershipCircuit{
		PublicKeyHash: publicKeyHash,
		Challenge:     wrongChallenge,
		Response:      response,
	}

	assert.ProverFailed(circuit, assignment, test.WithCurves(ecc.BN254))
}

// --- ResponseAttestation Tests ---

func TestResponseAttestation_Valid(t *testing.T) {
	assert := test.NewAssert(t)

	agentKeyProof := big.NewInt(777)
	sourceDataHash := big.NewInt(555)
	timestamp := big.NewInt(1700000000)
	minTimestamp := big.NewInt(1699999700) // 5 min before
	maxTimestamp := big.NewInt(1700000030) // 30s after

	agentDIDHash := mimcHash(agentKeyProof)
	responseHash := mimcHash(sourceDataHash, agentKeyProof, timestamp)

	circuit := &ResponseAttestationCircuit{}
	assignment := &ResponseAttestationCircuit{
		ResponseHash:   responseHash,
		AgentDIDHash:   agentDIDHash,
		Timestamp:      timestamp,
		MinTimestamp:    minTimestamp,
		MaxTimestamp:    maxTimestamp,
		SourceDataHash: sourceDataHash,
		AgentKeyProof:  agentKeyProof,
	}

	assert.ProverSucceeded(circuit, assignment, test.WithCurves(ecc.BN254))
}

func TestResponseAttestation_WrongAgentKey(t *testing.T) {
	assert := test.NewAssert(t)

	agentKeyProof := big.NewInt(777)
	sourceDataHash := big.NewInt(555)
	timestamp := big.NewInt(1700000000)
	minTimestamp := big.NewInt(1699999700)
	maxTimestamp := big.NewInt(1700000030)

	agentDIDHash := mimcHash(agentKeyProof)
	responseHash := mimcHash(sourceDataHash, agentKeyProof, timestamp)

	wrongAgentKey := big.NewInt(888)

	circuit := &ResponseAttestationCircuit{}
	assignment := &ResponseAttestationCircuit{
		ResponseHash:   responseHash,
		AgentDIDHash:   agentDIDHash,
		Timestamp:      timestamp,
		MinTimestamp:    minTimestamp,
		MaxTimestamp:    maxTimestamp,
		SourceDataHash: sourceDataHash,
		AgentKeyProof:  wrongAgentKey,
	}

	assert.ProverFailed(circuit, assignment, test.WithCurves(ecc.BN254))
}

func TestResponseAttestation_WrongTimestamp(t *testing.T) {
	assert := test.NewAssert(t)

	agentKeyProof := big.NewInt(777)
	sourceDataHash := big.NewInt(555)
	timestamp := big.NewInt(1700000000)
	minTimestamp := big.NewInt(1699999700)
	maxTimestamp := big.NewInt(1700000030)

	agentDIDHash := mimcHash(agentKeyProof)
	responseHash := mimcHash(sourceDataHash, agentKeyProof, timestamp)

	wrongTimestamp := big.NewInt(1700000001)

	circuit := &ResponseAttestationCircuit{}
	assignment := &ResponseAttestationCircuit{
		ResponseHash:   responseHash,
		AgentDIDHash:   agentDIDHash,
		Timestamp:      wrongTimestamp,
		MinTimestamp:    minTimestamp,
		MaxTimestamp:    maxTimestamp,
		SourceDataHash: sourceDataHash,
		AgentKeyProof:  agentKeyProof,
	}

	assert.ProverFailed(circuit, assignment, test.WithCurves(ecc.BN254))
}

func TestResponseAttestation_TimestampBelowMin(t *testing.T) {
	assert := test.NewAssert(t)

	agentKeyProof := big.NewInt(777)
	sourceDataHash := big.NewInt(555)
	timestamp := big.NewInt(1699999600) // before minTimestamp
	minTimestamp := big.NewInt(1699999700)
	maxTimestamp := big.NewInt(1700000030)

	agentDIDHash := mimcHash(agentKeyProof)
	responseHash := mimcHash(sourceDataHash, agentKeyProof, timestamp)

	circuit := &ResponseAttestationCircuit{}
	assignment := &ResponseAttestationCircuit{
		ResponseHash:   responseHash,
		AgentDIDHash:   agentDIDHash,
		Timestamp:      timestamp,
		MinTimestamp:    minTimestamp,
		MaxTimestamp:    maxTimestamp,
		SourceDataHash: sourceDataHash,
		AgentKeyProof:  agentKeyProof,
	}

	assert.ProverFailed(circuit, assignment, test.WithCurves(ecc.BN254))
}

func TestResponseAttestation_TimestampAboveMax(t *testing.T) {
	assert := test.NewAssert(t)

	agentKeyProof := big.NewInt(777)
	sourceDataHash := big.NewInt(555)
	timestamp := big.NewInt(1700000100) // after maxTimestamp
	minTimestamp := big.NewInt(1699999700)
	maxTimestamp := big.NewInt(1700000030)

	agentDIDHash := mimcHash(agentKeyProof)
	responseHash := mimcHash(sourceDataHash, agentKeyProof, timestamp)

	circuit := &ResponseAttestationCircuit{}
	assignment := &ResponseAttestationCircuit{
		ResponseHash:   responseHash,
		AgentDIDHash:   agentDIDHash,
		Timestamp:      timestamp,
		MinTimestamp:    minTimestamp,
		MaxTimestamp:    maxTimestamp,
		SourceDataHash: sourceDataHash,
		AgentKeyProof:  agentKeyProof,
	}

	assert.ProverFailed(circuit, assignment, test.WithCurves(ecc.BN254))
}

// --- BalanceRange Tests ---

func TestBalanceRange_Above(t *testing.T) {
	assert := test.NewAssert(t)

	circuit := &BalanceRangeCircuit{}
	assignment := &BalanceRangeCircuit{
		Threshold: big.NewInt(50),
		Balance:   big.NewInt(100),
	}

	assert.ProverSucceeded(circuit, assignment, test.WithCurves(ecc.BN254))
}

func TestBalanceRange_Below(t *testing.T) {
	assert := test.NewAssert(t)

	circuit := &BalanceRangeCircuit{}
	assignment := &BalanceRangeCircuit{
		Threshold: big.NewInt(50),
		Balance:   big.NewInt(30),
	}

	assert.ProverFailed(circuit, assignment, test.WithCurves(ecc.BN254))
}

func TestBalanceRange_Equal(t *testing.T) {
	assert := test.NewAssert(t)

	circuit := &BalanceRangeCircuit{}
	assignment := &BalanceRangeCircuit{
		Threshold: big.NewInt(50),
		Balance:   big.NewInt(50),
	}

	assert.ProverSucceeded(circuit, assignment, test.WithCurves(ecc.BN254))
}

// --- AgentCapability Tests ---

func TestAgentCapability_Valid(t *testing.T) {
	assert := test.NewAssert(t)

	testHash := big.NewInt(1234)
	actualScore := big.NewInt(85)
	minScore := big.NewInt(70)
	agentDIDHash := big.NewInt(9999)

	capabilityHash := mimcHash(testHash, actualScore)
	agentTestBinding := mimcHash(testHash, agentDIDHash)

	circuit := &AgentCapabilityCircuit{}
	assignment := &AgentCapabilityCircuit{
		CapabilityHash:   capabilityHash,
		AgentDIDHash:     agentDIDHash,
		MinScore:         minScore,
		AgentTestBinding: agentTestBinding,
		ActualScore:      actualScore,
		TestHash:         testHash,
	}

	assert.ProverSucceeded(circuit, assignment, test.WithCurves(ecc.BN254))
}

func TestAgentCapability_BelowMinimum(t *testing.T) {
	assert := test.NewAssert(t)

	testHash := big.NewInt(1234)
	actualScore := big.NewInt(40)
	minScore := big.NewInt(70)
	agentDIDHash := big.NewInt(9999)

	capabilityHash := mimcHash(testHash, actualScore)
	agentTestBinding := mimcHash(testHash, agentDIDHash)

	circuit := &AgentCapabilityCircuit{}
	assignment := &AgentCapabilityCircuit{
		CapabilityHash:   capabilityHash,
		AgentDIDHash:     agentDIDHash,
		MinScore:         minScore,
		AgentTestBinding: agentTestBinding,
		ActualScore:      actualScore,
		TestHash:         testHash,
	}

	assert.ProverFailed(circuit, assignment, test.WithCurves(ecc.BN254))
}

func TestAgentCapability_WrongBinding(t *testing.T) {
	assert := test.NewAssert(t)

	testHash := big.NewInt(1234)
	actualScore := big.NewInt(85)
	minScore := big.NewInt(70)
	agentDIDHash := big.NewInt(9999)

	wrongCapabilityHash := big.NewInt(111111)
	agentTestBinding := mimcHash(testHash, agentDIDHash)

	circuit := &AgentCapabilityCircuit{}
	assignment := &AgentCapabilityCircuit{
		CapabilityHash:   wrongCapabilityHash,
		AgentDIDHash:     agentDIDHash,
		MinScore:         minScore,
		AgentTestBinding: agentTestBinding,
		ActualScore:      actualScore,
		TestHash:         testHash,
	}

	assert.ProverFailed(circuit, assignment, test.WithCurves(ecc.BN254))
}

func TestAgentCapability_WrongAgentTestBinding(t *testing.T) {
	assert := test.NewAssert(t)

	testHash := big.NewInt(1234)
	actualScore := big.NewInt(85)
	minScore := big.NewInt(70)
	agentDIDHash := big.NewInt(9999)

	capabilityHash := mimcHash(testHash, actualScore)
	wrongBinding := big.NewInt(111111) // wrong agent-test binding

	circuit := &AgentCapabilityCircuit{}
	assignment := &AgentCapabilityCircuit{
		CapabilityHash:   capabilityHash,
		AgentDIDHash:     agentDIDHash,
		MinScore:         minScore,
		AgentTestBinding: wrongBinding,
		ActualScore:      actualScore,
		TestHash:         testHash,
	}

	assert.ProverFailed(circuit, assignment, test.WithCurves(ecc.BN254))
}
