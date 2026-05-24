package zkp

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type proofSchemeAndSrsModeValidationCompileErrorCircuit struct{}

func (*proofSchemeAndSrsModeValidationCompileErrorCircuit) Define(frontend.API) error {
	return errors.New("define failed")
}

type proofSchemeAndSrsModeValidationBadWitnessCircuit struct {
	Public frontend.Variable `gnark:",public"`
}

func (*proofSchemeAndSrsModeValidationBadWitnessCircuit) Define(frontend.API) error {
	return nil
}

func newProofSchemeAndSrsModeValidationService(scheme ProofScheme) *ProverService {
	return &ProverService{
		scheme:   scheme,
		srsMode:  SRSModeUnsafe,
		logger:   newTestLogger(),
		compiled: make(map[string]*CompiledCircuit),
	}
}

func TestProofSchemeAndSRSModeValidation(t *testing.T) {
	assert.True(t, SchemePlonk.Valid())
	assert.True(t, SchemeGroth16.Valid())
	assert.False(t, ProofScheme("").Valid())
	assert.False(t, ProofScheme("snark").Valid())

	assert.True(t, SRSModeUnsafe.Valid())
	assert.True(t, SRSModeFile.Valid())
	assert.False(t, SRSMode("").Valid())
	assert.False(t, SRSMode("remote").Valid())
}

func TestNewProverServiceDefaultsAndCacheDirErrors(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "nested", "cache")
	svc, err := NewProverService(Config{
		CacheDir: cacheDir,
		Logger:   newTestLogger(),
	})
	require.NoError(t, err)
	require.NotNil(t, svc)
	assert.Equal(t, SchemePlonk, svc.Scheme())
	assert.Equal(t, SRSModeUnsafe, svc.srsMode)
	assert.NotNil(t, svc.compiled)

	info, err := os.Stat(cacheDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())

	filePath := filepath.Join(t.TempDir(), "cache-file")
	require.NoError(t, os.WriteFile(filePath, []byte("not a directory"), 0o600))

	_, err = NewProverService(Config{
		CacheDir: filePath,
		Logger:   newTestLogger(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create ZKP cache dir")
}

func TestCompileRejectsUnsupportedSchemeAndWrapsCircuitErrors(t *testing.T) {
	circuit, _ := validOwnershipAssignment()

	unsupported := newProofSchemeAndSrsModeValidationService(ProofScheme("snark"))
	err := unsupported.Compile("ownership", circuit)
	require.ErrorIs(t, err, ErrUnsupportedScheme)
	assert.Contains(t, err.Error(), "snark")
	assert.False(t, unsupported.IsCompiled("ownership"))

	broken := newProofSchemeAndSrsModeValidationService(SchemeGroth16)
	err = broken.Compile("broken", &proofSchemeAndSrsModeValidationCompileErrorCircuit{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `compile circuit "broken"`)
	assert.Contains(t, err.Error(), "define failed")
	assert.False(t, broken.IsCompiled("broken"))
}

func TestProveRejectsInvalidStateBeforeBackendProvers(t *testing.T) {
	_, assignment := validOwnershipAssignment()

	t.Run("witness error", func(t *testing.T) {
		svc := newProofSchemeAndSrsModeValidationService(SchemePlonk)
		svc.compiled["ownership"] = &CompiledCircuit{}

		proof, err := svc.Prove(context.Background(), "ownership", &proofSchemeAndSrsModeValidationBadWitnessCircuit{
			Public: struct{}{},
		})
		require.Error(t, err)
		assert.Nil(t, proof)
		assert.Contains(t, err.Error(), `create witness for "ownership"`)
	})

	t.Run("invalid plonk proving key", func(t *testing.T) {
		svc := newProofSchemeAndSrsModeValidationService(SchemePlonk)
		svc.compiled["ownership"] = &CompiledCircuit{ProvingKey: "not a plonk key"}

		proof, err := svc.Prove(context.Background(), "ownership", assignment)
		require.Error(t, err)
		assert.Nil(t, proof)
		assert.Contains(t, err.Error(), `invalid plonk proving key for "ownership"`)
	})

	t.Run("invalid groth16 proving key", func(t *testing.T) {
		svc := newProofSchemeAndSrsModeValidationService(SchemeGroth16)
		svc.compiled["ownership"] = &CompiledCircuit{ProvingKey: "not a groth16 key"}

		proof, err := svc.Prove(context.Background(), "ownership", assignment)
		require.Error(t, err)
		assert.Nil(t, proof)
		assert.Contains(t, err.Error(), `invalid groth16 proving key for "ownership"`)
	})

	t.Run("unsupported scheme", func(t *testing.T) {
		svc := newProofSchemeAndSrsModeValidationService(ProofScheme("snark"))
		svc.compiled["ownership"] = &CompiledCircuit{}

		proof, err := svc.Prove(context.Background(), "ownership", assignment)
		require.ErrorIs(t, err, ErrUnsupportedScheme)
		assert.Nil(t, proof)
		assert.Contains(t, err.Error(), "snark")
	})
}

func TestVerifyRejectsInvalidStateBeforeBackendVerification(t *testing.T) {
	_, assignment := validOwnershipAssignment()

	t.Run("nil proof", func(t *testing.T) {
		svc := newProofSchemeAndSrsModeValidationService(SchemePlonk)
		valid, err := svc.Verify(context.Background(), nil, assignment)
		require.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "empty proof")
	})

	t.Run("empty proof data", func(t *testing.T) {
		svc := newProofSchemeAndSrsModeValidationService(SchemePlonk)
		valid, err := svc.Verify(context.Background(), &Proof{CircuitID: "ownership"}, assignment)
		require.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "empty proof")
	})

	t.Run("uncompiled circuit", func(t *testing.T) {
		svc := newProofSchemeAndSrsModeValidationService(SchemePlonk)
		valid, err := svc.Verify(context.Background(), &Proof{
			Data:      []byte("proof"),
			CircuitID: "missing",
		}, assignment)
		require.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), `circuit "missing" not compiled`)
	})

	t.Run("invalid plonk verifying key", func(t *testing.T) {
		svc := newProofSchemeAndSrsModeValidationService(SchemePlonk)
		svc.compiled["ownership"] = &CompiledCircuit{VerifyingKey: "not a plonk key"}

		valid, err := svc.Verify(context.Background(), &Proof{
			Data:      []byte("proof"),
			CircuitID: "ownership",
		}, assignment)
		require.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), `invalid plonk verifying key for "ownership"`)
	})

	t.Run("invalid groth16 verifying key", func(t *testing.T) {
		svc := newProofSchemeAndSrsModeValidationService(SchemeGroth16)
		svc.compiled["ownership"] = &CompiledCircuit{VerifyingKey: "not a groth16 key"}

		valid, err := svc.Verify(context.Background(), &Proof{
			Data:      []byte("proof"),
			CircuitID: "ownership",
		}, assignment)
		require.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), `invalid groth16 verifying key for "ownership"`)
	})

	t.Run("unsupported scheme", func(t *testing.T) {
		svc := newProofSchemeAndSrsModeValidationService(ProofScheme("snark"))
		svc.compiled["ownership"] = &CompiledCircuit{}

		valid, err := svc.Verify(context.Background(), &Proof{
			Data:      []byte("proof"),
			CircuitID: "ownership",
		}, assignment)
		require.ErrorIs(t, err, ErrUnsupportedScheme)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "snark")
	})
}

func TestVerifyReportsMalformedProofDeserialization(t *testing.T) {
	_, assignment := validOwnershipAssignment()

	tests := []struct {
		name         string
		scheme       ProofScheme
		verifyingKey any
		want         string
	}{
		{
			name:         "plonk",
			scheme:       SchemePlonk,
			verifyingKey: plonk.NewVerifyingKey(ecc.BN254),
			want:         `deserialize plonk proof for "ownership"`,
		},
		{
			name:         "groth16",
			scheme:       SchemeGroth16,
			verifyingKey: groth16.NewVerifyingKey(ecc.BN254),
			want:         `deserialize groth16 proof for "ownership"`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			svc := newProofSchemeAndSrsModeValidationService(tt.scheme)
			svc.compiled["ownership"] = &CompiledCircuit{VerifyingKey: tt.verifyingKey}

			valid, err := svc.Verify(context.Background(), &Proof{
				Data:      []byte("not a serialized proof"),
				CircuitID: "ownership",
			}, assignment)
			require.Error(t, err)
			assert.False(t, valid)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestExportGroth16VerifierWrapsCompileErrors(t *testing.T) {
	svc := newProofSchemeAndSrsModeValidationService(SchemePlonk)
	err := svc.ExportGroth16Verifier("broken", &proofSchemeAndSrsModeValidationCompileErrorCircuit{}, os.Stdout)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `compile circuit "broken" for groth16`)
	assert.Contains(t, err.Error(), "define failed")
}

func TestExportGroth16VerifierWritesSolidity(t *testing.T) {
	svc := newProofSchemeAndSrsModeValidationService(SchemePlonk)
	circuit, _ := validOwnershipAssignment()

	var out bytes.Buffer
	err := svc.ExportGroth16Verifier("ownership", circuit, &out)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "contract Verifier")
	assert.Contains(t, out.String(), "function verifyProof")
}

func TestCompilePlonkFileSRSFallsBackWhenFileMissing(t *testing.T) {
	svc, err := NewProverService(Config{
		CacheDir: t.TempDir(),
		Scheme:   SchemePlonk,
		Logger:   newTestLogger(),
		SRSMode:  SRSModeFile,
		SRSPath:  filepath.Join(t.TempDir(), "missing.srs"),
	})
	require.NoError(t, err)

	circuit, _ := validOwnershipAssignment()
	require.NoError(t, svc.Compile("ownership", circuit))
	assert.True(t, svc.IsCompiled("ownership"))
}

func TestLoadSRSFromFileReportsMissingFile(t *testing.T) {
	canonical, lagrange, err := loadSRSFromFile(filepath.Join(t.TempDir(), "missing.srs"))
	require.ErrorIs(t, err, os.ErrNotExist)
	assert.Nil(t, canonical)
	assert.Nil(t, lagrange)
	assert.Contains(t, err.Error(), "open SRS file")
}

func TestLoadSRSFromFileReportsReadErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.srs")
	require.NoError(t, os.WriteFile(path, nil, 0o600))

	canonical, lagrange, err := loadSRSFromFile(path)
	require.Error(t, err)
	assert.Nil(t, canonical)
	assert.Nil(t, lagrange)
	assert.Contains(t, err.Error(), "read canonical SRS")
}
