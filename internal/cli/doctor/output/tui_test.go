package output

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/cli/doctor/checks"
)

func TestTUIRenderer_RenderResult_SanitizesDisplayText(t *testing.T) {
	t.Parallel()

	renderer := &TUIRenderer{}
	out := renderer.RenderResult(checks.Result{
		Name:    "Multi-\x1b[31mAgent\n",
		Status:  checks.StatusWarn,
		Message: "Recent\x1b[31m failures\nfound",
		Details: "Inspect\x1b[31m traces\nnow",
	})

	assert.Contains(t, out, "Multi-Agent")
	assert.Contains(t, out, "Recent failures found")
	assert.Contains(t, out, "Inspect traces now")
	assert.NotContains(t, out, "\x1b")
}
