// Command zkexport compiles gnark circuits and exports Groth16 verifying keys
// as Solidity contracts for on-chain proof verification.
//
// Usage:
//
//	go run cmd/zkexport/main.go --circuit ownership --output contracts/src/verifiers/OwnershipVerifier.sol
//	go run cmd/zkexport/main.go --all --outdir contracts/src/verifiers/
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/consensys/gnark/frontend"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/p2p/zkp"
	"github.com/langoai/lango/internal/p2p/zkp/circuits"
)

type verifierExporter interface {
	ExportGroth16Verifier(circuitID string, circuit frontend.Circuit, w io.Writer) error
}

type zkexportOptions struct {
	circuitName string
	output      string
	all         bool
	outDir      string
}

var newZKExportProverService = func() (verifierExporter, error) {
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	return zkp.NewProverService(zkp.Config{
		Scheme: zkp.SchemeGroth16,
		Logger: sugar,
	})
}

var (
	zkexportArgs             = func() []string { return os.Args[1:] }
	zkexportStdout io.Writer = os.Stdout
	zkexportStderr io.Writer = os.Stderr
	zkexportExitFn           = os.Exit
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
	zkexportExitFn(runZKExport(zkexportArgs(), zkexportStdout, zkexportStderr))
}

func runZKExport(args []string, stdout, stderr io.Writer) int {
	opts, err := parseZKExportArgs(args, stderr)
	if err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 1
	}

	if opts.all {
		svc, err := newZKExportProverService()
		if err != nil {
			fmt.Fprintf(stderr, "create prover service: %v\n", err)
			return 1
		}
		dir := opts.outDir
		if dir == "" {
			dir = "contracts/src/verifiers"
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(stderr, "create output dir: %v\n", err)
			return 1
		}
		var progress bytes.Buffer
		var createdPaths []string
		for _, id := range circuitIDs() {
			circuit := registeredCircuits[id]
			filename := circuitToFilename(id)
			path := filepath.Join(dir, filename)
			if err := exportCircuit(svc, id, circuit, path, &progress); err != nil {
				for _, created := range createdPaths {
					_ = os.Remove(created)
				}
				fmt.Fprintf(stderr, "export %s: %v\n", id, err)
				return 1
			}
			createdPaths = append(createdPaths, path)
		}
		_, _ = io.Copy(stdout, &progress)
		fmt.Fprintf(stdout, "Exported %d verifier contracts to %s/\n", len(registeredCircuits), dir)
		return 0
	}

	circuit, ok := registeredCircuits[opts.circuitName]
	if !ok {
		fmt.Fprintf(stderr, "unknown circuit %q. Available: %s\n", opts.circuitName, strings.Join(circuitIDs(), ", "))
		return 1
	}

	svc, err := newZKExportProverService()
	if err != nil {
		fmt.Fprintf(stderr, "create prover service: %v\n", err)
		return 1
	}

	if err := exportCircuit(svc, opts.circuitName, circuit, opts.output, stdout); err != nil {
		fmt.Fprintf(stderr, "export: %v\n", err)
		return 1
	}
	return 0
}

func parseZKExportArgs(args []string, stderr io.Writer) (zkexportOptions, error) {
	var opts zkexportOptions

	fs := flag.NewFlagSet("zkexport", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&opts.circuitName, "circuit", "", "Circuit ID to export (ownership, attestation, balance, capability)")
	fs.StringVar(&opts.output, "output", "", "Output Solidity file path")
	fs.BoolVar(&opts.all, "all", false, "Export all circuits")
	fs.StringVar(&opts.outDir, "outdir", "", "Output directory for --all mode")

	if err := fs.Parse(args); err != nil {
		return opts, err
	}

	if !opts.all && (opts.circuitName == "" || opts.output == "") {
		writeZKExportUsage(stderr)
		return opts, fmt.Errorf("missing required flags")
	}

	return opts, nil
}

func writeZKExportUsage(stderr io.Writer) {
	fmt.Fprintln(stderr, "Usage: zkexport --circuit <name> --output <path>")
	fmt.Fprintln(stderr, "       zkexport --all --outdir <dir>")
	fmt.Fprintln(stderr, "\nAvailable circuits:", strings.Join(circuitIDs(), ", "))
}

func exportCircuit(svc verifierExporter, id string, circuit frontend.Circuit, path string, stdout io.Writer) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file %q: %w", path, err)
	}
	defer f.Close()

	if err := svc.ExportGroth16Verifier(id, circuit, f); err != nil {
		_ = os.Remove(path) // Clean up on error.
		return err
	}

	fmt.Fprintf(stdout, "  %s → %s\n", id, path)
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
	sort.Strings(ids)
	return ids
}
