package settings

import (
	"bytes"
	"strings"
	"testing"
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
