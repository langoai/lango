package bootstrap

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBootstrapPhaseInventoryCommentsStayInSync(t *testing.T) {
	phasesText := readBootstrapSource(t, "phases.go")
	defaultPhasesComment := sourceCommentBefore(t, phasesText, "func DefaultPhases(")
	runText := readBootstrapSource(t, "bootstrap.go")
	runComment := sourceCommentBefore(t, runText, "func Run(")
	phaseNames := defaultPhaseNames(t)

	assert.Contains(t, defaultPhasesComment, "standard bootstrap phase sequence (12 phases)")
	assert.NotContains(t, defaultPhasesComment, "standard bootstrap phase sequence (11 phases)")

	previousIndex := -1
	for _, name := range phaseNames {
		index := strings.Index(runComment, name)
		require.NotEqual(t, -1, index, "Run comment missing phase name %q", name)
		assert.Greater(t, index, previousIndex, "Run comment lists phase name %q out of order", name)
		previousIndex = index
	}
	assert.NotContains(t, runComment, "7. Load or create configuration profile")
}

func defaultPhaseNames(t *testing.T) []string {
	t.Helper()

	phases := DefaultPhases()
	names := make([]string, 0, len(phases))
	for _, phase := range phases {
		names = append(names, phase.Name)
	}
	require.Len(t, names, 12)
	return names
}

func readBootstrapSource(t *testing.T, name string) string {
	t.Helper()

	data, err := os.ReadFile(name)
	require.NoError(t, err)
	return string(data)
}

func sourceCommentBefore(t *testing.T, source, marker string) string {
	t.Helper()

	markerIndex := strings.Index(source, marker)
	require.NotEqual(t, -1, markerIndex, "source marker %q not found", marker)
	prefix := source[:markerIndex]
	lines := strings.Split(prefix, "\n")
	commentLines := make([]string, 0)
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" && len(commentLines) == 0 {
			continue
		}
		if !strings.HasPrefix(line, "//") {
			break
		}
		commentLines = append([]string{line}, commentLines...)
	}
	require.NotEmpty(t, commentLines, "source marker %q has no leading comment", marker)
	return strings.Join(commentLines, "\n")
}
