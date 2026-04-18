package archtest

import (
	"os/exec"
	"strings"
	"testing"
)

func TestNoBootstrapDBCloseCallsRemain(t *testing.T) {
	root := findModuleRoot(t)
	cmd := exec.Command("rg", "-n", `boot\.DBClient\.Close\(|result\.DBClient\.Close\(`, "./internal", "./cmd")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("found direct bootstrap DB close calls:\n%s", string(out))
	}
	// rg exits 1 when no matches are found.
	if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
		return
	}
	t.Fatalf("rg failed: %v\n%s", err, string(out))
}

func TestBootstrapResultNoDirectDBFields(t *testing.T) {
	root := findModuleRoot(t)
	cmd := exec.Command("rg", "-n", `^\s*DBClient\s+\*ent\.Client|^\s*RawDB\s+\*sql\.DB`, "./internal/bootstrap/bootstrap.go")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("found direct DB fields on bootstrap.Result:\n%s", string(out))
	}
	if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
		return
	}
	t.Fatalf("rg failed: %v\n%s", err, string(out))
}

func TestOpenDatabaseReadOnlyOnlyUsedInApprovedPackages(t *testing.T) {
	root := findModuleRoot(t)
	cmd := exec.Command("rg", "-n", `OpenDatabaseReadOnly`, "./internal")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return
		}
		t.Fatalf("rg failed: %v\n%s", err, string(out))
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		if strings.Contains(line, "internal/bootstrap/bootstrap.go") ||
			strings.Contains(line, "internal/storagebroker/server.go") ||
			strings.Contains(line, "internal/archtest/bootstrap_boundary_test.go") {
			continue
		}
		t.Fatalf("unexpected OpenDatabaseReadOnly usage: %s", line)
	}
}

func TestBootstrapDBHandleUsageOnlyInApprovedFiles(t *testing.T) {
	root := findModuleRoot(t)
	cmd := exec.Command("rg", "-n", `boot\.DBClient|boot\.RawDB|m\.boot\.DBClient|m\.boot\.RawDB|result\.DBClient|result\.RawDB`, "./internal", "./cmd")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return
		}
		t.Fatalf("rg failed: %v\n%s", err, string(out))
	}

	allowed := []string{
		"internal/cli/agent/trace.go",
		"internal/cli/cron/cron.go",
		"internal/cli/doctor/checks/multi_agent.go",
		"internal/cli/librarian/inquiries.go",
		"internal/cli/learning/history.go",
		"internal/cli/p2p/p2p.go",
		"internal/cli/p2p/reputation.go",
		"internal/cli/payment/payment.go",
		"internal/cli/run/run.go",
		"internal/cli/smartaccount/deps.go",
		"internal/cli/workflow/workflow.go",
		"internal/archtest/bootstrap_boundary_test.go",
	}

	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		ok := false
		for _, allow := range allowed {
			if strings.Contains(line, allow) {
				ok = true
				break
			}
		}
		if !ok {
			t.Fatalf("unexpected bootstrap DB handle usage: %s", line)
		}
	}
}

func TestExtraFilesOnlyUsedInBrokerClient(t *testing.T) {
	root := findModuleRoot(t)
	cmd := exec.Command("rg", "-n", `ExtraFiles`, "./internal", "./cmd")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return
		}
		t.Fatalf("rg failed: %v\n%s", err, string(out))
	}

	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		if strings.Contains(line, "internal/storagebroker/client.go") ||
			strings.Contains(line, "internal/archtest/bootstrap_boundary_test.go") {
			continue
		}
		t.Fatalf("unexpected ExtraFiles usage: %s", line)
	}
}

func TestNoBootstrapDBHandlesRemainInAppAndMigratedCLI(t *testing.T) {
	root := findModuleRoot(t)
	cmd := exec.Command(
		"rg",
		"-n",
		`boot\.DBClient|boot\.RawDB`,
		"./internal/app",
		"./internal/cli/security",
		"./internal/cli/sandbox",
		"./internal/cli/doctor",
		"./internal/cli/configcmd",
		"./internal/cli/settings",
		"./internal/cli/onboard",
		"./internal/cli/provenance",
	)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("found direct bootstrap DB handle usage in migrated packages:\n%s", string(out))
	}
	if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
		return
	}
	t.Fatalf("rg failed: %v\n%s", err, string(out))
}

func TestExtraFilesOnlyConfiguredByBroker(t *testing.T) {
	root := findModuleRoot(t)
	cmd := exec.Command("rg", "-n", `ExtraFiles`, "./internal")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return
		}
		t.Fatalf("rg failed: %v\n%s", err, string(out))
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		if strings.Contains(line, "internal/storagebroker/client.go") ||
			strings.Contains(line, "internal/archtest/bootstrap_boundary_test.go") {
			continue
		}
		t.Fatalf("unexpected ExtraFiles usage: %s", line)
	}
}

func TestNoRemovedStorageFacadeRawAccessorsInProductionPackages(t *testing.T) {
	root := findModuleRoot(t)
	cmd := exec.Command(
		"rg",
		"-n",
		"-g",
		"!**/*_test.go",
		`EntClient\(|RawDB\(|FTSDB\(|PaymentClient\(`,
		"./internal/app",
		"./internal/cli",
		"./internal/knowledge",
		"./internal/librarian",
		"./internal/agentmemory",
		"./internal/payment",
		"./internal/p2p",
		"./internal/observability",
		"./internal/workflow",
	)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("found removed raw storage accessor usage in production packages:\n%s", string(out))
	}
	if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
		return
	}
	t.Fatalf("rg failed: %v\n%s", err, string(out))
}

func TestNoTestOnlyStorageWiringHelpersInProductionPackages(t *testing.T) {
	root := findModuleRoot(t)
	cmd := exec.Command(
		"rg",
		"-n",
		"-g",
		"!**/*_test.go",
		`WithEntClient\(|WithRawDB\(`,
		"./internal/app",
		"./internal/cli",
		"./internal/knowledge",
		"./internal/librarian",
		"./internal/agentmemory",
		"./internal/payment",
		"./internal/p2p",
		"./internal/observability",
		"./internal/workflow",
	)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("found test-only storage wiring helper usage in production packages:\n%s", string(out))
	}
	if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
		return
	}
	t.Fatalf("rg failed: %v\n%s", err, string(out))
}

func TestNoPaymentEntClientExtractionInProductionPaths(t *testing.T) {
	root := findModuleRoot(t)
	cmd := exec.Command(
		"rg",
		"-n",
		"-g",
		"!**/*_test.go",
		`entStore\.Client\(`,
		"./internal/cli/payment",
		"./internal/app/wiring_payment.go",
		"./internal/app/wiring_p2p.go",
	)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("found payment-side ent client extraction in production paths:\n%s", string(out))
	}
	if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
		return
	}
	t.Fatalf("rg failed: %v\n%s", err, string(out))
}
