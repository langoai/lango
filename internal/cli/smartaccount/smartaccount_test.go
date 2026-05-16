package smartaccount

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/testutil"
)

func executeSmartAccountRootCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestNewAccountCmd_HelpExamplesIncludePaymasterSurface(t *testing.T) {
	cmd := NewAccountCmd(testutil.FailBootLoader(assert.AnError))

	out, err := executeSmartAccountRootCmd(t, cmd, "--help")
	assert.NoError(t, err)
	assert.Contains(t, out, "lango account info")
	assert.Contains(t, out, "lango account deploy")
	assert.Contains(t, out, "lango account session list")
	assert.Contains(t, out, "lango account module list")
	assert.Contains(t, out, "lango account policy show")
	assert.Contains(t, out, "lango account paymaster status")
}
