package smartaccount

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/storage"

	_ "github.com/mattn/go-sqlite3"
)

func TestPolicyHelpersWrapBootstrapErrors(t *testing.T) {
	t.Parallel()

	bootErr := errors.New("boot failed")
	loader := func() (*bootstrap.Result, error) {
		return nil, bootErr
	}

	showResult, showCleanup, showErr := loadPolicyShowInfo(loader)
	require.Error(t, showErr)
	assert.Contains(t, showErr.Error(), "bootstrap: boot failed")
	assert.Equal(t, policyShowInfo{}, showResult)
	assert.Nil(t, showCleanup)

	setResult, setCleanup, setErr := updatePolicyLimits(loader, "100", "", "")
	require.Error(t, setErr)
	assert.Contains(t, setErr.Error(), "bootstrap: boot failed")
	assert.Equal(t, policySetResult{}, setResult)
	assert.Nil(t, setCleanup)
}

func TestPolicyHelpersCloseBootOnInitFailure(t *testing.T) {
	tests := []struct {
		name string
		run  func(BootLoader) error
	}{
		{
			name: "show",
			run: func(loader BootLoader) error {
				_, _, err := loadPolicyShowInfo(loader)
				return err
			},
		},
		{
			name: "set",
			run: func(loader BootLoader) error {
				_, _, err := updatePolicyLimits(loader, "100", "", "")
				return err
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, err := sql.Open("sqlite3", ":memory:")
			require.NoError(t, err)

			boot := &bootstrap.Result{
				Config:  config.DefaultConfig(),
				Storage: storage.NewFacade(nil, nil, storage.WithRawDB(db)),
			}
			err = tt.run(func() (*bootstrap.Result, error) {
				return boot, nil
			})

			require.Error(t, err)
			assert.Contains(t, err.Error(), "smart account not enabled")
			assert.Error(t, db.Ping(), "bootstrap result storage should be closed on init failure")
		})
	}
}

func TestParsePolicyLimitInputsReturnsOnlyProvidedLimits(t *testing.T) {
	t.Parallel()

	maxTx, daily, monthly, err := parsePolicyLimitInputs("100", "", "300")
	require.NoError(t, err)
	require.NotNil(t, maxTx)
	require.NotNil(t, monthly)
	assert.Equal(t, "100", maxTx.String())
	assert.Nil(t, daily)
	assert.Equal(t, "300", monthly.String())

	maxTx, daily, monthly, err = parsePolicyLimitInputs("", "200", "")
	require.NoError(t, err)
	assert.Nil(t, maxTx)
	require.NotNil(t, daily)
	assert.Equal(t, "200", daily.String())
	assert.Nil(t, monthly)
}
