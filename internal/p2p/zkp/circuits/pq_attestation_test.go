package circuits

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/test"
)

func TestPQAttestation_ValidProof(t *testing.T) {
	attestorSecret := big.NewInt(42)
	messagePreimage := big.NewInt(123456)
	pqKeyPreimage := big.NewInt(789012)

	attestorHash := mimcHash(attestorSecret)
	messageHash := mimcHash(messagePreimage)
	pqKeyHash := mimcHash(pqKeyPreimage)

	assignment := &PQAttestationCircuit{
		AttestorDIDHash:     attestorHash,
		MessageHash:         messageHash,
		PQPublicKeyHash:     pqKeyHash,
		Timestamp:           1000,
		MinTimestamp:        900,
		DealID:              1,
		ChainID:             8453,
		ContractAddress:     0xDEADBEEF,
		AttestorSecret:      attestorSecret,
		MessagePreimage:     messagePreimage,
		PQPublicKeyPreimage: pqKeyPreimage,
	}

	assert := test.NewAssert(t)
	assert.ProverSucceeded(
		&PQAttestationCircuit{},
		assignment,
		test.WithCurves(ecc.BN254),
		test.WithBackends(backend.GROTH16),
	)
}

func TestPQAttestation_WrongAttestor(t *testing.T) {
	attestorHash := mimcHash(big.NewInt(42))
	messageHash := mimcHash(big.NewInt(123))
	pqKeyHash := mimcHash(big.NewInt(456))

	assignment := &PQAttestationCircuit{
		AttestorDIDHash:     attestorHash,
		MessageHash:         messageHash,
		PQPublicKeyHash:     pqKeyHash,
		Timestamp:           1000,
		MinTimestamp:        900,
		DealID:              1,
		ChainID:             1,
		ContractAddress:     0xABCD,
		AttestorSecret:      big.NewInt(99), // wrong secret
		MessagePreimage:     big.NewInt(123),
		PQPublicKeyPreimage: big.NewInt(456),
	}

	assert := test.NewAssert(t)
	assert.ProverFailed(
		&PQAttestationCircuit{},
		assignment,
		test.WithCurves(ecc.BN254),
		test.WithBackends(backend.GROTH16),
	)
}

func TestPQAttestation_ExpiredTimestamp(t *testing.T) {
	attestorHash := mimcHash(big.NewInt(42))
	messageHash := mimcHash(big.NewInt(123))
	pqKeyHash := mimcHash(big.NewInt(456))

	assignment := &PQAttestationCircuit{
		AttestorDIDHash:     attestorHash,
		MessageHash:         messageHash,
		PQPublicKeyHash:     pqKeyHash,
		Timestamp:           800, // before MinTimestamp
		MinTimestamp:        900,
		DealID:              1,
		ChainID:             1,
		ContractAddress:     0xABCD,
		AttestorSecret:      big.NewInt(42),
		MessagePreimage:     big.NewInt(123),
		PQPublicKeyPreimage: big.NewInt(456),
	}

	assert := test.NewAssert(t)
	assert.ProverFailed(
		&PQAttestationCircuit{},
		assignment,
		test.WithCurves(ecc.BN254),
		test.WithBackends(backend.GROTH16),
	)
}

func TestPQAttestation_DomainBinding(t *testing.T) {
	attestorSecret := big.NewInt(42)
	attestorHash := mimcHash(attestorSecret)
	messageHash := mimcHash(big.NewInt(123))
	pqKeyHash := mimcHash(big.NewInt(456))

	// Deal 1 on Base chain.
	assignment := &PQAttestationCircuit{
		AttestorDIDHash:     attestorHash,
		MessageHash:         messageHash,
		PQPublicKeyHash:     pqKeyHash,
		Timestamp:           1000,
		MinTimestamp:        900,
		DealID:              1,
		ChainID:             8453,
		ContractAddress:     0x1234,
		AttestorSecret:      attestorSecret,
		MessagePreimage:     big.NewInt(123),
		PQPublicKeyPreimage: big.NewInt(456),
	}

	assert := test.NewAssert(t)
	assert.ProverSucceeded(&PQAttestationCircuit{}, assignment,
		test.WithCurves(ecc.BN254), test.WithBackends(backend.GROTH16))
}
