package settings

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/langoai/lango/internal/bootstrap"
)

func TestPrintNextStepsIncludesServeDoctorAndProfileManagement(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	printNextSteps(&out, "default")

	wantSnippets := []string{
		"lango serve",
		"lango doctor",
		"lango config list",
		"lango config use",
	}

	for _, want := range wantSnippets {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("next steps should mention %q; output:\n%s", want, out.String())
		}
	}
}

func TestNewCommand_NonInteractiveGuardRunsBeforeSettings(t *testing.T) {
	guardErr := errors.New("settings requires an interactive terminal")
	prevGuard := requireInteractiveTerminal
	prevRun := runSettingsFn
	defer func() {
		requireInteractiveTerminal = prevGuard
		runSettingsFn = prevRun
	}()

	requireInteractiveTerminal = func(message string) error {
		if !strings.Contains(message, "lango config import") {
			t.Fatalf("guard message should include scripted configuration guidance, got %q", message)
		}
		return guardErr
	}
	runCalled := false
	runSettingsFn = func(out io.Writer, profileName string) error {
		runCalled = true
		return nil
	}

	cmd := NewCommand()
	err := cmd.Execute()
	if !errors.Is(err, guardErr) {
		t.Fatalf("expected guard error, got %v", err)
	}
	if runCalled {
		t.Fatal("settings run path should not start after non-interactive guard failure")
	}
}

func TestRunSettingsUsesBootLoader(t *testing.T) {
	bootErr := errors.New("bootstrap unavailable")
	prevBoot := settingsBootResult
	defer func() {
		settingsBootResult = prevBoot
	}()

	called := false
	settingsBootResult = func() (*bootstrap.Result, error) {
		called = true
		return nil, bootErr
	}

	err := runSettings(io.Discard, "default")

	if !called {
		t.Fatal("runSettings should use the shared boot loader seam")
	}
	if !errors.Is(err, bootErr) {
		t.Fatalf("expected bootstrap error, got %v", err)
	}
}
