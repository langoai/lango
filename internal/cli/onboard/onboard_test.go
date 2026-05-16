package onboard

import (
	"bytes"
	"strings"
	"testing"

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
