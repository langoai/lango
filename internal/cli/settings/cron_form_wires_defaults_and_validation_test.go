package settings

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

func TestCronFormWiresDefaultsAndValidation(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Cron.Enabled = true
	cfg.Cron.Timezone = "Asia/Seoul"
	cfg.Cron.MaxConcurrentJobs = 7
	cfg.Cron.DefaultSessionMode = ""
	cfg.Cron.HistoryRetention = "14d"
	cfg.Cron.DefaultDeliverTo = []string{"telegram", "discord"}

	form := NewCronForm(cfg)

	assert.Equal(t, "Cron Scheduler Configuration", form.Title)
	assert.Equal(t, []string{
		"cron_enabled",
		"cron_timezone",
		"cron_max_jobs",
		"cron_session_mode",
		"cron_history_retention",
		"cron_default_deliver",
	}, formKeys(form))

	enabled := fieldByKey(form, "cron_enabled")
	require.NotNil(t, enabled)
	assert.Equal(t, tuicore.InputBool, enabled.Type)
	assert.True(t, enabled.Checked)

	assert.Equal(t, "Asia/Seoul", fieldByKey(form, "cron_timezone").Value)
	assert.Equal(t, "14d", fieldByKey(form, "cron_history_retention").Value)
	assert.Equal(t, "telegram,discord", fieldByKey(form, "cron_default_deliver").Value)

	maxJobs := fieldByKey(form, "cron_max_jobs")
	require.NotNil(t, maxJobs)
	assert.Equal(t, tuicore.InputInt, maxJobs.Type)
	assert.Equal(t, "7", maxJobs.Value)
	require.NoError(t, maxJobs.Validate("1"))
	assert.EqualError(t, maxJobs.Validate("0"), "must be a positive integer")
	assert.EqualError(t, maxJobs.Validate("not-a-number"), "must be a positive integer")

	sessionMode := fieldByKey(form, "cron_session_mode")
	require.NotNil(t, sessionMode)
	assert.Equal(t, tuicore.InputSelect, sessionMode.Type)
	assert.Equal(t, "isolated", sessionMode.Value)
	assert.Equal(t, []string{"isolated", "main"}, sessionMode.Options)
}

func TestBackgroundFormWiresFieldsAndValidators(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Background.Enabled = true
	cfg.Background.YieldMs = 0
	cfg.Background.MaxConcurrentTasks = 9
	cfg.Background.DefaultDeliverTo = []string{"slack", "telegram"}

	form := NewBackgroundForm(cfg)

	assert.Equal(t, "Background Tasks Configuration", form.Title)
	assert.Equal(t, []string{
		"bg_enabled",
		"bg_yield_ms",
		"bg_max_tasks",
		"bg_default_deliver",
	}, formKeys(form))

	assert.True(t, fieldByKey(form, "bg_enabled").Checked)
	assert.Equal(t, "slack,telegram", fieldByKey(form, "bg_default_deliver").Value)

	yield := fieldByKey(form, "bg_yield_ms")
	require.NotNil(t, yield)
	assert.Equal(t, tuicore.InputInt, yield.Type)
	assert.Equal(t, "0", yield.Value)
	require.NoError(t, yield.Validate("0"))
	require.NoError(t, yield.Validate("250"))
	assert.EqualError(t, yield.Validate("-1"), "must be a non-negative integer")
	assert.EqualError(t, yield.Validate("bad"), "must be a non-negative integer")

	maxTasks := fieldByKey(form, "bg_max_tasks")
	require.NotNil(t, maxTasks)
	assert.Equal(t, "9", maxTasks.Value)
	require.NoError(t, maxTasks.Validate("1"))
	assert.EqualError(t, maxTasks.Validate("0"), "must be a positive integer")
}

func TestWorkflowFormWiresFieldsAndValidation(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Workflow.Enabled = true
	cfg.Workflow.MaxConcurrentSteps = 6
	cfg.Workflow.DefaultTimeout = 45 * time.Second
	cfg.Workflow.StateDir = "/tmp/lango-workflows"
	cfg.Workflow.DefaultDeliverTo = []string{"discord"}

	form := NewWorkflowForm(cfg)

	assert.Equal(t, "Workflow Engine Configuration", form.Title)
	assert.Equal(t, []string{
		"wf_enabled",
		"wf_max_steps",
		"wf_timeout",
		"wf_state_dir",
		"wf_default_deliver",
	}, formKeys(form))

	assert.True(t, fieldByKey(form, "wf_enabled").Checked)
	assert.Equal(t, "45s", fieldByKey(form, "wf_timeout").Value)
	assert.Equal(t, "10m", fieldByKey(form, "wf_timeout").Placeholder)
	assert.Equal(t, "/tmp/lango-workflows", fieldByKey(form, "wf_state_dir").Value)
	assert.Equal(t, "discord", fieldByKey(form, "wf_default_deliver").Value)

	maxSteps := fieldByKey(form, "wf_max_steps")
	require.NotNil(t, maxSteps)
	assert.Equal(t, tuicore.InputInt, maxSteps.Type)
	assert.Equal(t, "6", maxSteps.Value)
	require.NoError(t, maxSteps.Validate("1"))
	assert.EqualError(t, maxSteps.Validate("0"), "must be a positive integer")
	assert.EqualError(t, maxSteps.Validate("bad"), "must be a positive integer")
}

func TestRunLedgerFormWiresFieldsAndValidators(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RunLedger.Enabled = true
	cfg.RunLedger.Shadow = true
	cfg.RunLedger.WriteThrough = true
	cfg.RunLedger.AuthoritativeRead = true
	cfg.RunLedger.WorkspaceIsolation = true
	cfg.RunLedger.StaleTTL = 90 * time.Minute
	cfg.RunLedger.MaxRunHistory = 0
	cfg.RunLedger.ValidatorTimeout = 3 * time.Minute
	cfg.RunLedger.PlannerMaxRetries = 4

	form := NewRunLedgerForm(cfg)

	assert.Equal(t, "RunLedger Configuration", form.Title)
	assert.Equal(t, []string{
		"runledger_enabled",
		"runledger_shadow",
		"runledger_write_through",
		"runledger_authoritative_read",
		"runledger_workspace_isolation",
		"runledger_stale_ttl",
		"runledger_max_history",
		"runledger_validator_timeout",
		"runledger_planner_retries",
	}, formKeys(form))

	for _, key := range []string{
		"runledger_enabled",
		"runledger_shadow",
		"runledger_write_through",
		"runledger_authoritative_read",
		"runledger_workspace_isolation",
	} {
		field := fieldByKey(form, key)
		require.NotNil(t, field)
		assert.Equal(t, tuicore.InputBool, field.Type)
		assert.True(t, field.Checked)
	}

	staleTTL := fieldByKey(form, "runledger_stale_ttl")
	require.NotNil(t, staleTTL)
	assert.Equal(t, "1h30m0s", staleTTL.Value)
	require.NoError(t, staleTTL.Validate("1ns"))
	assert.EqualError(t, staleTTL.Validate("0s"), "must be greater than 0")
	assert.EqualError(t, staleTTL.Validate("not-duration"), "must be a valid duration")

	maxHistory := fieldByKey(form, "runledger_max_history")
	require.NotNil(t, maxHistory)
	assert.Equal(t, tuicore.InputInt, maxHistory.Type)
	assert.Equal(t, "0", maxHistory.Value)
	require.NoError(t, maxHistory.Validate("0"))
	assert.EqualError(t, maxHistory.Validate("-1"), "must be 0 or greater")
	assert.EqualError(t, maxHistory.Validate("bad"), "must be an integer")

	validatorTimeout := fieldByKey(form, "runledger_validator_timeout")
	require.NotNil(t, validatorTimeout)
	assert.Equal(t, "3m0s", validatorTimeout.Value)
	require.NoError(t, validatorTimeout.Validate("250ms"))
	assert.EqualError(t, validatorTimeout.Validate("-1s"), "must be greater than 0")
	assert.EqualError(t, validatorTimeout.Validate("bad"), "must be a valid duration")

	plannerRetries := fieldByKey(form, "runledger_planner_retries")
	require.NotNil(t, plannerRetries)
	assert.Equal(t, "4", plannerRetries.Value)
	require.NoError(t, plannerRetries.Validate("0"))
	assert.EqualError(t, plannerRetries.Validate("-1"), "must be 0 or greater")
	assert.EqualError(t, plannerRetries.Validate("bad"), "must be an integer")
}

func TestProvenanceFormWiresFieldsAndValidators(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Provenance.Enabled = true
	cfg.Provenance.Checkpoints.AutoOnStepComplete = false
	cfg.Provenance.Checkpoints.AutoOnPolicy = true
	cfg.Provenance.Checkpoints.MaxPerSession = 0
	cfg.Provenance.Checkpoints.RetentionDays = 14

	form := NewProvenanceForm(cfg)

	assert.Equal(t, "Provenance Configuration", form.Title)
	assert.Equal(t, []string{
		"provenance_enabled",
		"provenance_auto_on_step_complete",
		"provenance_auto_on_policy",
		"provenance_max_per_session",
		"provenance_retention_days",
	}, formKeys(form))

	assert.True(t, fieldByKey(form, "provenance_enabled").Checked)
	assert.False(t, fieldByKey(form, "provenance_auto_on_step_complete").Checked)
	assert.True(t, fieldByKey(form, "provenance_auto_on_policy").Checked)

	maxPerSession := fieldByKey(form, "provenance_max_per_session")
	require.NotNil(t, maxPerSession)
	assert.Equal(t, tuicore.InputInt, maxPerSession.Type)
	assert.Equal(t, "0", maxPerSession.Value)
	require.NoError(t, maxPerSession.Validate("0"))
	assert.EqualError(t, maxPerSession.Validate("-1"), "must be 0 or greater")
	assert.EqualError(t, maxPerSession.Validate("bad"), "must be an integer")

	retentionDays := fieldByKey(form, "provenance_retention_days")
	require.NotNil(t, retentionDays)
	assert.Equal(t, "14", retentionDays.Value)
	require.NoError(t, retentionDays.Validate("30"))
	assert.EqualError(t, retentionDays.Validate("-1"), "must be 0 or greater")
	assert.EqualError(t, retentionDays.Validate("bad"), "must be an integer")
}
