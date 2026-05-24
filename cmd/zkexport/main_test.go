package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/consensys/gnark/frontend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubVerifierExporter struct {
}

func (s stubVerifierExporter) ExportGroth16Verifier(circuitID string, circuit frontend.Circuit, w io.Writer) error {
	buf, ok := w.(*os.File)
	if ok {
		_, err := buf.Write([]byte("// verifier"))
		return err
	}
	_, err := w.Write([]byte("// verifier"))
	return err
}

type failAfterWriteExporter struct{}

func (failAfterWriteExporter) ExportGroth16Verifier(circuitID string, circuit frontend.Circuit, w io.Writer) error {
	_, _ = w.Write([]byte("// partial verifier"))
	return errors.New("export failed")
}

type failOnSecondExporter struct {
	callCount int
}

func (f *failOnSecondExporter) ExportGroth16Verifier(circuitID string, circuit frontend.Circuit, w io.Writer) error {
	f.callCount++
	_, _ = w.Write([]byte("// verifier"))
	if f.callCount == 2 {
		return errors.New("second export failed")
	}
	return nil
}

func TestRunZKExport_MissingRequiredFlagsWritesUsageToStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := runZKExport(nil, &stdout, &stderr)

	assert.Equal(t, 1, code)
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), "Usage: zkexport --circuit <name> --output <path>")
	assert.Contains(t, stderr.String(), "Available circuits: attestation, balance, capability, ownership, pq_attestation")
}

func TestRunZKExport_HelpReturnsSuccessWithoutProverSetup(t *testing.T) {
	origNewService := newZKExportProverService
	t.Cleanup(func() { newZKExportProverService = origNewService })

	called := false
	newZKExportProverService = func() (verifierExporter, error) {
		called = true
		return nil, errors.New("should not be called")
	}

	var stdout, stderr bytes.Buffer
	code := runZKExport([]string{"--help"}, &stdout, &stderr)

	assert.Equal(t, 0, code)
	assert.False(t, called)
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), "Usage of zkexport:")
}

func TestRunZKExport_SingleCircuitWritesProgressToStdout(t *testing.T) {
	origNewService := newZKExportProverService
	t.Cleanup(func() { newZKExportProverService = origNewService })

	newZKExportProverService = func() (verifierExporter, error) {
		return stubVerifierExporter{}, nil
	}

	dir := t.TempDir()
	output := filepath.Join(dir, "OwnershipVerifier.sol")
	var stdout, stderr bytes.Buffer

	code := runZKExport([]string{"--circuit", "ownership", "--output", output}, &stdout, &stderr)

	assert.Equal(t, 0, code)
	assert.Empty(t, stderr.String())
	assert.Contains(t, stdout.String(), "ownership → "+output)
	data, err := os.ReadFile(output)
	require.NoError(t, err)
	assert.Equal(t, "// verifier", string(data))
}

func TestRunZKExport_AllModeWritesSummaryToStdout(t *testing.T) {
	origNewService := newZKExportProverService
	t.Cleanup(func() { newZKExportProverService = origNewService })

	newZKExportProverService = func() (verifierExporter, error) {
		return stubVerifierExporter{}, nil
	}

	dir := t.TempDir()
	var stdout, stderr bytes.Buffer

	code := runZKExport([]string{"--all", "--outdir", dir}, &stdout, &stderr)

	assert.Equal(t, 0, code)
	assert.Empty(t, stderr.String())
	out := stdout.String()
	assert.Contains(t, out, "Exported ")
	assert.Contains(t, out, dir+"/")
	assert.Less(t, strings.Index(out, "attestation →"), strings.Index(out, "balance →"))
	assert.Less(t, strings.Index(out, "balance →"), strings.Index(out, "capability →"))
	assert.Less(t, strings.Index(out, "capability →"), strings.Index(out, "ownership →"))
	assert.Less(t, strings.Index(out, "ownership →"), strings.Index(out, "pq_attestation →"))
}

func TestRunZKExport_ProverServiceErrorWritesToStderr(t *testing.T) {
	origNewService := newZKExportProverService
	t.Cleanup(func() { newZKExportProverService = origNewService })

	newZKExportProverService = func() (verifierExporter, error) {
		return nil, errors.New("boom")
	}

	var stdout, stderr bytes.Buffer
	code := runZKExport([]string{"--circuit", "ownership", "--output", "out.sol"}, &stdout, &stderr)

	assert.Equal(t, 1, code)
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), "create prover service: boom")
}

func TestRunZKExport_UnknownCircuitWritesDeterministicErrorToStderr(t *testing.T) {
	origNewService := newZKExportProverService
	t.Cleanup(func() { newZKExportProverService = origNewService })

	newZKExportProverService = func() (verifierExporter, error) {
		return stubVerifierExporter{}, nil
	}

	var stdout, stderr bytes.Buffer
	code := runZKExport([]string{"--circuit", "missing", "--output", "out.sol"}, &stdout, &stderr)

	assert.Equal(t, 1, code)
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), `unknown circuit "missing". Available: attestation, balance, capability, ownership, pq_attestation`)
}

func TestRunZKExport_UnknownCircuitSkipsProverServiceSetup(t *testing.T) {
	origNewService := newZKExportProverService
	t.Cleanup(func() { newZKExportProverService = origNewService })

	called := false
	newZKExportProverService = func() (verifierExporter, error) {
		called = true
		return nil, errors.New("should not be called")
	}

	var stdout, stderr bytes.Buffer
	code := runZKExport([]string{"--circuit", "missing", "--output", "out.sol"}, &stdout, &stderr)

	assert.Equal(t, 1, code)
	assert.False(t, called)
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), `unknown circuit "missing". Available: attestation, balance, capability, ownership, pq_attestation`)
}

func TestRunZKExport_SingleCircuitFailureRemovesPartialFileAndWritesToStderr(t *testing.T) {
	origNewService := newZKExportProverService
	t.Cleanup(func() { newZKExportProverService = origNewService })

	newZKExportProverService = func() (verifierExporter, error) {
		return failAfterWriteExporter{}, nil
	}

	dir := t.TempDir()
	output := filepath.Join(dir, "OwnershipVerifier.sol")
	var stdout, stderr bytes.Buffer

	code := runZKExport([]string{"--circuit", "ownership", "--output", output}, &stdout, &stderr)

	assert.Equal(t, 1, code)
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), "export: export failed")
	_, err := os.Stat(output)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestRunZKExport_AllModeFailureRemovesPartialFileAndWritesToStderr(t *testing.T) {
	origNewService := newZKExportProverService
	t.Cleanup(func() { newZKExportProverService = origNewService })

	newZKExportProverService = func() (verifierExporter, error) {
		return failAfterWriteExporter{}, nil
	}

	dir := t.TempDir()
	var stdout, stderr bytes.Buffer

	code := runZKExport([]string{"--all", "--outdir", dir}, &stdout, &stderr)

	assert.Equal(t, 1, code)
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), "export ")
	assert.Contains(t, stderr.String(), "export failed")

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 0)
}

func TestRunZKExport_AllModeFailureCleansEarlierSuccessfulFiles(t *testing.T) {
	origNewService := newZKExportProverService
	t.Cleanup(func() { newZKExportProverService = origNewService })

	newZKExportProverService = func() (verifierExporter, error) {
		return &failOnSecondExporter{}, nil
	}

	dir := t.TempDir()
	var stdout, stderr bytes.Buffer

	code := runZKExport([]string{"--all", "--outdir", dir}, &stdout, &stderr)

	assert.Equal(t, 1, code)
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), "export ")
	assert.Contains(t, stderr.String(), "second export failed")

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 0)
}

func TestMain_UsesInjectedSeams(t *testing.T) {
	origArgs := zkexportArgs
	origStdout := zkexportStdout
	origStderr := zkexportStderr
	origExit := zkexportExitFn
	t.Cleanup(func() {
		zkexportArgs = origArgs
		zkexportStdout = origStdout
		zkexportStderr = origStderr
		zkexportExitFn = origExit
	})

	zkexportArgs = func() []string { return nil }
	zkexportStdout = &bytes.Buffer{}
	var stderr bytes.Buffer
	zkexportStderr = &stderr

	exitCode := -1
	zkexportExitFn = func(code int) {
		exitCode = code
	}

	main()

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr.String(), "Usage: zkexport --circuit <name> --output <path>")
}
