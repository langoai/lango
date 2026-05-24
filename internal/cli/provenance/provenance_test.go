package provenance

import (
	"bytes"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/storage"
)

func disabledBootLoader(t *testing.T) func() (*bootstrap.Result, error) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	cfg := config.DefaultConfig()
	cfg.Provenance.Enabled = false
	return func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config:  cfg,
			Storage: storage.NewFacade(nil, nil, storage.WithEntClient(client)),
		}, nil
	}
}

func TestProvenanceDisabled_SubcommandsPrintNotice(t *testing.T) {
	tests := []struct {
		give string
		args []string
	}{
		{give: "checkpoint list", args: []string{"checkpoint", "list", "--run", "r1"}},
		{give: "checkpoint create", args: []string{"checkpoint", "create", "label", "--run", "r1"}},
		{give: "checkpoint show", args: []string{"checkpoint", "show", "cp-1"}},
		{give: "session tree", args: []string{"session", "tree", "sk-1"}},
		{give: "session list", args: []string{"session", "list"}},
		{give: "attribution show", args: []string{"attribution", "show", "sk-1"}},
		{give: "attribution report", args: []string{"attribution", "report", "sk-1"}},
		{give: "bundle export", args: []string{"bundle", "export", "sk-1"}},
		{give: "bundle import", args: []string{"bundle", "import", "/dev/null"}},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			cmd := NewProvenanceCmd(disabledBootLoader(t))
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)
			err := cmd.Execute()
			assert.NoError(t, err)
			assert.Contains(t, buf.String(), "Provenance is disabled")
		})
	}
}

func TestProvenanceDisabled_StatusShowsConfigAndNotice(t *testing.T) {
	cmd := NewProvenanceCmd(disabledBootLoader(t))
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"status"})
	err := cmd.Execute()
	assert.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "Enabled:")
	assert.Contains(t, out, "Provenance is disabled")
}
