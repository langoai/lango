package onboard

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/storage"
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

func TestRunOnboardPreservesLoadedExplicitKeysForExistingProfile(t *testing.T) {
	store := &onboardProfileStore{
		loadCfg: config.DefaultConfig(),
		loadExplicitKeys: map[string]bool{
			"knowledge.enabled": true,
		},
	}
	prevBoot := onboardBootResult
	prevRunWizard := runOnboardWizard
	defer func() {
		onboardBootResult = prevBoot
		runOnboardWizard = prevRunWizard
	}()

	onboardBootResult = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Storage: storage.NewFacade(store, nil)}, nil
	}
	runOnboardWizard = func(cfg *config.Config) (*Wizard, error) {
		wizard := NewWizard(cfg)
		wizard.Completed = true
		return wizard, nil
	}

	if err := runOnboard(io.Discard, "existing", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.saveExplicitKeys == nil {
		t.Fatal("expected explicit keys to be preserved on save")
	}
	if !store.saveExplicitKeys["knowledge.enabled"] {
		t.Fatalf("expected saved explicit keys to include knowledge.enabled, got %#v", store.saveExplicitKeys)
	}
	if store.setActiveCalled {
		t.Fatal("existing profile should not be activated as new")
	}
}

func TestRunOnboardSavesPresetExplicitKeysForNewProfile(t *testing.T) {
	store := &onboardProfileStore{loadErr: configstore.ErrProfileNotFound}
	prevBoot := onboardBootResult
	prevRunWizard := runOnboardWizard
	defer func() {
		onboardBootResult = prevBoot
		runOnboardWizard = prevRunWizard
	}()

	onboardBootResult = func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Storage: storage.NewFacade(store, nil)}, nil
	}
	runOnboardWizard = func(cfg *config.Config) (*Wizard, error) {
		wizard := NewWizard(cfg)
		wizard.Completed = true
		return wizard, nil
	}

	if err := runOnboard(io.Discard, "research", "researcher"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := config.PresetExplicitKeys("researcher")
	for key := range want {
		if !store.saveExplicitKeys[key] {
			t.Fatalf("expected saved preset explicit keys to include %q, got %#v", key, store.saveExplicitKeys)
		}
	}
	if !store.setActiveCalled {
		t.Fatal("new preset-backed profile should be activated")
	}
}

type onboardProfileStore struct {
	loadCfg          *config.Config
	loadExplicitKeys map[string]bool
	loadErr          error
	saveExplicitKeys map[string]bool
	setActiveCalled  bool
}

func (s *onboardProfileStore) Save(ctx context.Context, name string, cfg *config.Config, explicitKeys map[string]bool) error {
	s.saveExplicitKeys = explicitKeys
	return nil
}

func (s *onboardProfileStore) Load(ctx context.Context, name string) (*config.Config, map[string]bool, error) {
	if s.loadErr != nil {
		return nil, nil, s.loadErr
	}
	return s.loadCfg, s.loadExplicitKeys, nil
}

func (s *onboardProfileStore) LoadActive(ctx context.Context) (string, *config.Config, map[string]bool, error) {
	return "", nil, nil, errors.New("not implemented")
}

func (s *onboardProfileStore) SetActive(ctx context.Context, name string) error {
	s.setActiveCalled = true
	return nil
}

func (s *onboardProfileStore) List(ctx context.Context) ([]configstore.ProfileInfo, error) {
	return nil, errors.New("not implemented")
}

func (s *onboardProfileStore) Delete(ctx context.Context, name string) error {
	return errors.New("not implemented")
}

func (s *onboardProfileStore) Exists(ctx context.Context, name string) (bool, error) {
	return false, errors.New("not implemented")
}
