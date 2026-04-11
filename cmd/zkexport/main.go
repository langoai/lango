// Command zkexport compiles gnark circuits and exports Groth16 verifying keys
// as Solidity contracts for on-chain proof verification.
//
// Usage:
//
//	go run cmd/zkexport/main.go --circuit ownership --output contracts/src/verifiers/OwnershipVerifier.sol
//	go run cmd/zkexport/main.go --all --outdir contracts/src/verifiers/
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/consensys/gnark/frontend"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/p2p/zkp"
	"github.com/langoai/lango/internal/p2p/zkp/circuits"
)

// registeredCircuits maps circuit IDs to their empty struct definitions.
var registeredCircuits = map[string]frontend.Circuit{
	"ownership":      &circuits.WalletOwnershipCircuit{},
	"attestation":    &circuits.ResponseAttestationCircuit{},
	"balance":        &circuits.BalanceRangeCircuit{},
	"capability":     &circuits.AgentCapabilityCircuit{},
	"pq_attestation": &circuits.PQAttestationCircuit{},
}

func main() {
	circuitName := flag.String("circuit", "", "Circuit ID to export (ownership, attestation, balance, capability)")
	output := flag.String("output", "", "Output Solidity file path")
	all := flag.Bool("all", false, "Export all circuits")
	outDir := flag.String("outdir", "", "Output directory for --all mode")
	flag.Parse()

	if !*all && (*circuitName == "" || *output == "") {
		fmt.Fprintln(os.Stderr, "Usage: zkexport --circuit <name> --output <path>")
		fmt.Fprintln(os.Stderr, "       zkexport --all --outdir <dir>")
		fmt.Fprintln(os.Stderr, "\nAvailable circuits:", strings.Join(circuitIDs(), ", "))
		os.Exit(1)
	}

	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	svc, err := zkp.NewProverService(zkp.Config{
		Scheme: zkp.SchemeGroth16,
		Logger: sugar,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "create prover service: %v\n", err)
		os.Exit(1)
	}

	if *all {
		dir := *outDir
		if dir == "" {
			dir = "contracts/src/verifiers"
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "create output dir: %v\n", err)
			os.Exit(1)
		}
		for id, circuit := range registeredCircuits {
			filename := circuitToFilename(id)
			path := filepath.Join(dir, filename)
			if err := exportCircuit(svc, id, circuit, path); err != nil {
				fmt.Fprintf(os.Stderr, "export %s: %v\n", id, err)
				os.Exit(1)
			}
		}
		fmt.Printf("Exported %d verifier contracts to %s/\n", len(registeredCircuits), dir)
		return
	}

	circuit, ok := registeredCircuits[*circuitName]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown circuit %q. Available: %s\n", *circuitName, strings.Join(circuitIDs(), ", "))
		os.Exit(1)
	}

	if err := exportCircuit(svc, *circuitName, circuit, *output); err != nil {
		fmt.Fprintf(os.Stderr, "export: %v\n", err)
		os.Exit(1)
	}
}

func exportCircuit(svc *zkp.ProverService, id string, circuit frontend.Circuit, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file %q: %w", path, err)
	}
	defer f.Close()

	if err := svc.ExportGroth16Verifier(id, circuit, f); err != nil {
		os.Remove(path) // Clean up on error.
		return err
	}

	fmt.Printf("  %s → %s\n", id, path)
	return nil
}

func circuitToFilename(id string) string {
	parts := strings.Split(id, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "") + "Verifier.sol"
}

func circuitIDs() []string {
	ids := make([]string, 0, len(registeredCircuits))
	for id := range registeredCircuits {
		ids = append(ids, id)
	}
	return ids
}
