// Package archtest provides automated architecture boundary verification.
// These tests parse the Go import graph and enforce cross-domain dependency
// rules — any violation fails the test (enforced mode).
package archtest

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

// goListPackage represents the JSON output of `go list -json`.
type goListPackage struct {
	ImportPath string   `json:"ImportPath"`
	Imports    []string `json:"Imports"`
}

// boundaryRule defines a forbidden cross-domain dependency.
type boundaryRule struct {
	// name is a human-readable description of the rule.
	name string
	// sourceMatch returns true if the given import path is a source subject to this rule.
	sourceMatch func(importPath string) bool
	// forbiddenMatch returns true if the given dependency is forbidden for matched sources.
	forbiddenMatch func(dep string) bool
}

const modulePath = "github.com/langoai/lango/"

// p2pNetworkingPrefixes are the p2p sub-packages that must not reach into economy/payment/wallet.
var p2pNetworkingPrefixes = []string{
	modulePath + "internal/p2p/discovery",
	modulePath + "internal/p2p/handshake",
	modulePath + "internal/p2p/identity",
	modulePath + "internal/p2p/firewall",
	modulePath + "internal/p2p/protocol",
	modulePath + "internal/p2p/agentpool",
}

var forbiddenForP2P = []string{
	modulePath + "internal/economy",
	modulePath + "internal/payment",
	modulePath + "internal/wallet",
}

var rules = []boundaryRule{
	{
		name: "p2p networking must not import economy/payment/wallet",
		sourceMatch: func(importPath string) bool {
			for _, prefix := range p2pNetworkingPrefixes {
				if importPath == prefix || strings.HasPrefix(importPath, prefix+"/") {
					return true
				}
			}
			return false
		},
		forbiddenMatch: func(dep string) bool {
			for _, prefix := range forbiddenForP2P {
				if dep == prefix || strings.HasPrefix(dep, prefix+"/") {
					return true
				}
			}
			return false
		},
	},
	{
		name: "economy must not import p2p",
		sourceMatch: func(importPath string) bool {
			return strings.HasPrefix(importPath, modulePath+"internal/economy/") ||
				importPath == modulePath+"internal/economy"
		},
		forbiddenMatch: func(dep string) bool {
			return strings.HasPrefix(dep, modulePath+"internal/p2p/") ||
				dep == modulePath+"internal/p2p"
		},
	},
	{
		name: "provenance must not import p2p/identity",
		sourceMatch: func(importPath string) bool {
			return importPath == modulePath+"internal/provenance" ||
				strings.HasPrefix(importPath, modulePath+"internal/provenance/")
		},
		forbiddenMatch: func(dep string) bool {
			return dep == modulePath+"internal/p2p/identity" ||
				strings.HasPrefix(dep, modulePath+"internal/p2p/identity/")
		},
	},
}

func TestBoundaryViolations(t *testing.T) {
	pkgs := loadPackages(t)

	violations := 0
	for _, pkg := range pkgs {
		for _, rule := range rules {
			if !rule.sourceMatch(pkg.ImportPath) {
				continue
			}
			shortSource := strings.TrimPrefix(pkg.ImportPath, modulePath)
			for _, dep := range pkg.Imports {
				if rule.forbiddenMatch(dep) {
					shortDep := strings.TrimPrefix(dep, modulePath)
					t.Errorf("BOUNDARY VIOLATION: %s imports %s (rule: %s)", shortSource, shortDep, rule.name)
					violations++
				}
			}
		}
	}

	t.Logf("Boundary check complete: %d violation(s) found", violations)
}

// loadPackages runs `go list -json ./internal/...` and parses the output
// into a slice of goListPackage. It calls t.Fatalf on setup errors.
func loadPackages(t *testing.T) []goListPackage {
	t.Helper()

	cmd := exec.Command("go", "list", "-json", "./internal/...")
	cmd.Dir = findModuleRoot(t)

	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}

	// go list -json outputs concatenated JSON objects (not an array),
	// so we use a streaming decoder.
	dec := json.NewDecoder(strings.NewReader(string(out)))
	var pkgs []goListPackage
	for dec.More() {
		var pkg goListPackage
		if err := dec.Decode(&pkg); err != nil {
			t.Fatalf("decode go list output: %v", err)
		}
		pkgs = append(pkgs, pkg)
	}

	if len(pkgs) == 0 {
		t.Fatalf("go list returned no packages")
	}

	return pkgs
}

// findModuleRoot locates the module root by running `go env GOMOD` and
// returning its parent directory.
func findModuleRoot(t *testing.T) string {
	t.Helper()

	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		t.Fatalf("go env GOMOD: %v", err)
	}

	gomod := strings.TrimSpace(string(out))
	if gomod == "" || gomod == "/dev/null" {
		t.Fatalf("not inside a Go module")
	}

	// gomod is e.g. /path/to/lango/go.mod — return the directory.
	idx := strings.LastIndex(gomod, "/")
	if idx < 0 {
		t.Fatalf("unexpected GOMOD path: %s", gomod)
	}
	return gomod[:idx]
}
