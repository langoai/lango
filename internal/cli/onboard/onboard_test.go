package onboard

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
)

func TestPrintNextStepsIncludesPrimaryEntryPoints(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	var out bytes.Buffer
	printNextSteps(&out, "default", cfg)

	wantSnippets := []string{
		"lango",
		"lango serve",
		"lango doctor",
		"lango settings",
	}

	for _, want := range wantSnippets {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("next steps should mention %q; output:\n%s", want, out.String())
		}
	}
}

func TestNewCommand_NonInteractiveGuardRunsBeforeOnboard(t *testing.T) {
	guardErr := errors.New("onboard requires an interactive terminal")
	prevGuard := requireInteractiveTerminal
	prevRun := runOnboardFn
	defer func() {
		requireInteractiveTerminal = prevGuard
		runOnboardFn = prevRun
	}()

	requireInteractiveTerminal = func(in io.Reader, message string) error {
		if in == nil {
			t.Fatal("guard should receive command input stream")
		}
		if !strings.Contains(message, "lango config import") {
			t.Fatalf("guard message should include scripted setup guidance, got %q", message)
		}
		return guardErr
	}
	runCalled := false
	runOnboardFn = func(out io.Writer, profileName, preset string) error {
		runCalled = true
		return nil
	}

	cmd := NewCommand()
	err := cmd.Execute()
	if !errors.Is(err, guardErr) {
		t.Fatalf("expected guard error, got %v", err)
	}
	if runCalled {
		t.Fatal("onboard run path should not start after non-interactive guard failure")
	}
}

func TestRunOnboardUsesBootLoader(t *testing.T) {
	bootErr := errors.New("bootstrap unavailable")
	prevBoot := onboardBootResult
	defer func() {
		onboardBootResult = prevBoot
	}()

	called := false
	onboardBootResult = func() (*bootstrap.Result, error) {
		called = true
		return nil, bootErr
	}

	err := runOnboard(io.Discard, "default", "")

	if !called {
		t.Fatal("runOnboard should use the shared boot loader seam")
	}
	if !errors.Is(err, bootErr) {
		t.Fatalf("expected bootstrap error, got %v", err)
	}
}
