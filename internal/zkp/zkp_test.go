package zkp

import (
	"context"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	nativemimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/langoai/lango/internal/zkp/circuits"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func mimcHash(elems ...*big.Int) *big.Int {
	h := nativemimc.NewMiMC()
	for _, e := range elems {
		var elem fr.Element
		elem.SetBigInt(e)
		b := elem.Bytes()
		h.Write(b[:])
	}
	result := h.Sum(nil)
	var out big.Int
	out.SetBytes(result)
	return &out
}

func newTestLogger() *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	return logger.Sugar()
}

func validOwnershipAssignment() (*circuits.WalletOwnershipCircuit, *circuits.WalletOwnershipCircuit) {
	response := big.NewInt(42)
	challenge := big.NewInt(123)
	publicKeyHash := mimcHash(response, challenge)

	circuit := &circuits.WalletOwnershipCircuit{}
	assignment := &circuits.WalletOwnershipCircuit{
		PublicKeyHash: publicKeyHash,
		Challenge:     challenge,
		Response:      response,
	}
	return circuit, assignment
}

func TestProverService_CompileAndProve_PlonK(t *testing.T) {
	cfg := Config{
		CacheDir: t.TempDir(),
		Scheme:   "plonk",
		Logger:   newTestLogger(),
	}
	ps, err := NewProverService(cfg)
	require.NoError(t, err)

	circuit, assignment := validOwnershipAssignment()

	err = ps.Compile("wallet_ownership", circuit)
	require.NoError(t, err)

	proof, err := ps.Prove(context.Background(), "wallet_ownership", assignment)
	require.NoError(t, err)
	require.NotNil(t, proof)
	assert.Equal(t, "wallet_ownership", proof.CircuitID)
	assert.Equal(t, "plonk", proof.Scheme)
	assert.NotEmpty(t, proof.Data)
	assert.NotEmpty(t, proof.PublicInputs)
}

func TestProverService_CompileAndProve_Groth16(t *testing.T) {
	cfg := Config{
		CacheDir: t.TempDir(),
		Scheme:   "groth16",
		Logger:   newTestLogger(),
	}
	ps, err := NewProverService(cfg)
	require.NoError(t, err)

	circuit, assignment := validOwnershipAssignment()

	err = ps.Compile("wallet_ownership", circuit)
	require.NoError(t, err)

	proof, err := ps.Prove(context.Background(), "wallet_ownership", assignment)
	require.NoError(t, err)
	require.NotNil(t, proof)
	assert.Equal(t, "wallet_ownership", proof.CircuitID)
	assert.Equal(t, "groth16", proof.Scheme)
	assert.NotEmpty(t, proof.Data)
}

func TestProverService_Verify_Valid(t *testing.T) {
	cfg := Config{
		CacheDir: t.TempDir(),
		Scheme:   "plonk",
		Logger:   newTestLogger(),
	}
	ps, err := NewProverService(cfg)
	require.NoError(t, err)

	circuit, assignment := validOwnershipAssignment()

	err = ps.Compile("wallet_ownership", circuit)
	require.NoError(t, err)

	proof, err := ps.Prove(context.Background(), "wallet_ownership", assignment)
	require.NoError(t, err)

	// Build a verification circuit with only the public inputs set.
	verifyCircuit := &circuits.WalletOwnershipCircuit{
		PublicKeyHash: assignment.PublicKeyHash,
		Challenge:     assignment.Challenge,
	}

	valid, err := ps.Verify(context.Background(), proof, verifyCircuit)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestProverService_Verify_Invalid(t *testing.T) {
	cfg := Config{
		CacheDir: t.TempDir(),
		Scheme:   "plonk",
		Logger:   newTestLogger(),
	}
	ps, err := NewProverService(cfg)
	require.NoError(t, err)

	circuit, assignment := validOwnershipAssignment()

	err = ps.Compile("wallet_ownership", circuit)
	require.NoError(t, err)

	proof, err := ps.Prove(context.Background(), "wallet_ownership", assignment)
	require.NoError(t, err)

	// Tamper with proof bytes.
	tampered := make([]byte, len(proof.Data))
	copy(tampered, proof.Data)
	for i := 0; i < len(tampered) && i < 32; i++ {
		tampered[i] ^= 0xFF
	}
	tamperedProof := &Proof{
		Data:         tampered,
		PublicInputs: proof.PublicInputs,
		CircuitID:    proof.CircuitID,
		Scheme:       proof.Scheme,
	}

	verifyCircuit := &circuits.WalletOwnershipCircuit{
		PublicKeyHash: assignment.PublicKeyHash,
		Challenge:     assignment.Challenge,
	}

	valid, err := ps.Verify(context.Background(), tamperedProof, verifyCircuit)
	// Tampered proof should either fail verification (valid=false) or return an error.
	if err == nil {
		assert.False(t, valid)
	}
}

func TestProverService_DoubleCompile_Idempotent(t *testing.T) {
	cfg := Config{
		CacheDir: t.TempDir(),
		Scheme:   "plonk",
		Logger:   newTestLogger(),
	}
	ps, err := NewProverService(cfg)
	require.NoError(t, err)

	circuit := &circuits.WalletOwnershipCircuit{}

	err = ps.Compile("wallet_ownership", circuit)
	require.NoError(t, err)
	assert.True(t, ps.IsCompiled("wallet_ownership"))

	// Compile the same circuit again — should succeed silently.
	err = ps.Compile("wallet_ownership", circuit)
	require.NoError(t, err)
	assert.True(t, ps.IsCompiled("wallet_ownership"))
}

func TestProverService_ProveUncompiled_Error(t *testing.T) {
	cfg := Config{
		CacheDir: t.TempDir(),
		Scheme:   "plonk",
		Logger:   newTestLogger(),
	}
	ps, err := NewProverService(cfg)
	require.NoError(t, err)

	_, assignment := validOwnershipAssignment()

	_, err = ps.Prove(context.Background(), "nonexistent", assignment)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not compiled")
}
