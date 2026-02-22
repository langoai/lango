// Package zkp provides zero-knowledge proof generation and verification
// using the gnark library with support for PlonK and Groth16 proving schemes.
package zkp

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/test/unsafekzg"
	"go.uber.org/zap"
)

// Proof holds the serialized proof data and metadata.
type Proof struct {
	Data         []byte `json:"data"`
	PublicInputs []byte `json:"publicInputs"`
	CircuitID    string `json:"circuitId"`
	Scheme       string `json:"scheme"`
}

// CompiledCircuit stores a compiled constraint system with its proving and verifying keys.
type CompiledCircuit struct {
	CCS          constraint.ConstraintSystem
	ProvingKey   any // groth16.ProvingKey or plonk.ProvingKey
	VerifyingKey any // groth16.VerifyingKey or plonk.VerifyingKey
}

// Config configures the ProverService.
type Config struct {
	CacheDir string
	Scheme   string // "plonk" (default) or "groth16"
	Logger   *zap.SugaredLogger
}

// ProverService manages circuit compilation, proof generation, and verification.
type ProverService struct {
	cacheDir string
	scheme   string // "plonk" or "groth16"
	logger   *zap.SugaredLogger
	mu       sync.RWMutex
	compiled map[string]*CompiledCircuit
}

// NewProverService creates a new ZKP prover service.
func NewProverService(cfg Config) (*ProverService, error) {
	scheme := cfg.Scheme
	if scheme == "" {
		scheme = "plonk"
	}

	cacheDir := cfg.CacheDir
	if cacheDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home dir: %w", err)
		}
		cacheDir = filepath.Join(home, ".lango", "zkp", "cache")
	}

	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		return nil, fmt.Errorf("create ZKP cache dir: %w", err)
	}

	svc := &ProverService{
		cacheDir: cacheDir,
		scheme:   scheme,
		logger:   cfg.Logger,
		compiled: make(map[string]*CompiledCircuit),
	}

	cfg.Logger.Infow("ZKP prover service initialized",
		"scheme", scheme,
		"cacheDir", cacheDir,
	)

	return svc, nil
}

// Scheme returns the proving scheme.
func (s *ProverService) Scheme() string { return s.scheme }

// Compile compiles the given circuit and caches the result under circuitID.
func (s *ProverService) Compile(circuitID string, circuit frontend.Circuit) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.compiled[circuitID]; ok {
		s.logger.Debugw("circuit already compiled", "circuitID", circuitID)
		return nil
	}

	s.logger.Infow("compiling circuit", "circuitID", circuitID, "scheme", s.scheme)

	var (
		ccs constraint.ConstraintSystem
		err error
	)

	switch s.scheme {
	case "plonk":
		ccs, err = frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, circuit)
	case "groth16":
		ccs, err = frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, circuit)
	default:
		return fmt.Errorf("unsupported proving scheme: %s", s.scheme)
	}
	if err != nil {
		return fmt.Errorf("compile circuit %q: %w", circuitID, err)
	}

	compiled := &CompiledCircuit{CCS: ccs}

	switch s.scheme {
	case "plonk":
		canonical, lagrange, err := unsafekzg.NewSRS(ccs)
		if err != nil {
			return fmt.Errorf("generate SRS for %q: %w", circuitID, err)
		}
		pk, vk, err := plonk.Setup(ccs, canonical, lagrange)
		if err != nil {
			return fmt.Errorf("plonk setup for %q: %w", circuitID, err)
		}
		compiled.ProvingKey = pk
		compiled.VerifyingKey = vk

	case "groth16":
		pk, vk, err := groth16.Setup(ccs)
		if err != nil {
			return fmt.Errorf("groth16 setup for %q: %w", circuitID, err)
		}
		compiled.ProvingKey = pk
		compiled.VerifyingKey = vk
	}

	s.compiled[circuitID] = compiled
	s.logger.Infow("circuit compiled",
		"circuitID", circuitID,
		"constraints", ccs.GetNbConstraints(),
	)
	return nil
}

// Prove generates a zero-knowledge proof for the given circuit assignment.
func (s *ProverService) Prove(ctx context.Context, circuitID string, assignment frontend.Circuit) (*Proof, error) {
	s.mu.RLock()
	compiled, ok := s.compiled[circuitID]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("circuit %q not compiled", circuitID)
	}

	fullWitness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
	if err != nil {
		return nil, fmt.Errorf("create witness for %q: %w", circuitID, err)
	}

	publicWitness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField(), frontend.PublicOnly())
	if err != nil {
		return nil, fmt.Errorf("create public witness for %q: %w", circuitID, err)
	}

	var proofBuf bytes.Buffer

	switch s.scheme {
	case "plonk":
		pk, ok := compiled.ProvingKey.(plonk.ProvingKey)
		if !ok {
			return nil, fmt.Errorf("invalid plonk proving key for %q", circuitID)
		}
		proof, err := plonk.Prove(compiled.CCS, pk, fullWitness)
		if err != nil {
			return nil, fmt.Errorf("plonk prove for %q: %w", circuitID, err)
		}
		if _, err := proof.WriteTo(&proofBuf); err != nil {
			return nil, fmt.Errorf("serialize plonk proof for %q: %w", circuitID, err)
		}

	case "groth16":
		pk, ok := compiled.ProvingKey.(groth16.ProvingKey)
		if !ok {
			return nil, fmt.Errorf("invalid groth16 proving key for %q", circuitID)
		}
		proof, err := groth16.Prove(compiled.CCS, pk, fullWitness)
		if err != nil {
			return nil, fmt.Errorf("groth16 prove for %q: %w", circuitID, err)
		}
		if _, err := proof.WriteTo(&proofBuf); err != nil {
			return nil, fmt.Errorf("serialize groth16 proof for %q: %w", circuitID, err)
		}

	default:
		return nil, fmt.Errorf("unsupported proving scheme: %s", s.scheme)
	}

	var publicBuf bytes.Buffer
	if _, err := publicWitness.WriteTo(&publicBuf); err != nil {
		return nil, fmt.Errorf("serialize public witness for %q: %w", circuitID, err)
	}

	s.logger.Debugw("proof generated",
		"circuitID", circuitID,
		"proofSize", proofBuf.Len(),
	)

	return &Proof{
		Data:         proofBuf.Bytes(),
		PublicInputs: publicBuf.Bytes(),
		CircuitID:    circuitID,
		Scheme:       s.scheme,
	}, nil
}

// Verify checks whether the given proof is valid for the circuit's public inputs.
func (s *ProverService) Verify(ctx context.Context, proof *Proof, circuit frontend.Circuit) (bool, error) {
	if proof == nil || len(proof.Data) == 0 {
		return false, fmt.Errorf("empty proof")
	}

	s.mu.RLock()
	compiled, ok := s.compiled[proof.CircuitID]
	s.mu.RUnlock()
	if !ok {
		return false, fmt.Errorf("circuit %q not compiled", proof.CircuitID)
	}

	publicWitness, err := frontend.NewWitness(circuit, ecc.BN254.ScalarField(), frontend.PublicOnly())
	if err != nil {
		return false, fmt.Errorf("create public witness for %q: %w", proof.CircuitID, err)
	}

	switch s.scheme {
	case "plonk":
		vk, ok := compiled.VerifyingKey.(plonk.VerifyingKey)
		if !ok {
			return false, fmt.Errorf("invalid plonk verifying key for %q", proof.CircuitID)
		}
		p := plonk.NewProof(ecc.BN254)
		if _, err := p.ReadFrom(bytes.NewReader(proof.Data)); err != nil {
			return false, fmt.Errorf("deserialize plonk proof for %q: %w", proof.CircuitID, err)
		}
		if err := plonk.Verify(p, vk, publicWitness); err != nil {
			return false, nil
		}
		return true, nil

	case "groth16":
		vk, ok := compiled.VerifyingKey.(groth16.VerifyingKey)
		if !ok {
			return false, fmt.Errorf("invalid groth16 verifying key for %q", proof.CircuitID)
		}
		p := groth16.NewProof(ecc.BN254)
		if _, err := p.ReadFrom(bytes.NewReader(proof.Data)); err != nil {
			return false, fmt.Errorf("deserialize groth16 proof for %q: %w", proof.CircuitID, err)
		}
		if err := groth16.Verify(p, vk, publicWitness); err != nil {
			return false, nil
		}
		return true, nil

	default:
		return false, fmt.Errorf("unsupported proving scheme: %s", s.scheme)
	}
}

// IsCompiled reports whether the circuit with the given ID has been compiled.
func (s *ProverService) IsCompiled(circuitID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.compiled[circuitID]
	return ok
}
