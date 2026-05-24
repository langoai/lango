package settings

import (
	"testing"
	"time"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

func TestNewObservabilityForm_WiresConfiguredValues(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Observability = config.ObservabilityConfig{
		Enabled: true,
		Tokens: config.TokenTrackingConfig{
			Enabled:        true,
			PersistHistory: true,
			RetentionDays:  14,
		},
		Health: config.HealthConfig{
			Enabled:  true,
			Interval: 45 * time.Second,
		},
		Audit: config.AuditConfig{
			Enabled:       true,
			RetentionDays: 120,
		},
		Metrics: config.MetricsExportConfig{
			Enabled: true,
			Format:  "prometheus",
		},
		TraceStore: config.TraceStoreConfig{
			MaxAge:                12 * time.Hour,
			MaxTraces:             321,
			FailedTraceMultiplier: 4,
			CleanupInterval:       15 * time.Minute,
		},
	}

	form := NewObservabilityForm(cfg)

	wantKeys := []string{
		"obs_enabled",
		"obs_tokens_enabled",
		"obs_tokens_persist",
		"obs_tokens_retention",
		"obs_health_enabled",
		"obs_health_interval",
		"obs_audit_enabled",
		"obs_audit_retention",
		"obs_metrics_enabled",
		"obs_metrics_format",
		"obs_trace_max_age",
		"obs_trace_max_traces",
		"obs_trace_failed_multiplier",
		"obs_trace_cleanup_interval",
	}
	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}
	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Fatalf("missing field %q", key)
		}
	}

	assertFieldChecked(t, form, "obs_enabled", true)
	assertFieldChecked(t, form, "obs_tokens_enabled", true)
	assertFieldChecked(t, form, "obs_tokens_persist", true)
	assertFieldValue(t, form, "obs_tokens_retention", tuicore.InputInt, "14")
	assertFieldChecked(t, form, "obs_health_enabled", true)
	assertFieldValue(t, form, "obs_health_interval", tuicore.InputText, "45s")
	assertFieldChecked(t, form, "obs_audit_enabled", true)
	assertFieldValue(t, form, "obs_audit_retention", tuicore.InputInt, "120")
	assertFieldChecked(t, form, "obs_metrics_enabled", true)
	assertFieldValue(t, form, "obs_metrics_format", tuicore.InputSelect, "prometheus")
	assertFieldValue(t, form, "obs_trace_max_age", tuicore.InputText, "12h0m0s")
	assertFieldValue(t, form, "obs_trace_max_traces", tuicore.InputInt, "321")
	assertFieldValue(t, form, "obs_trace_failed_multiplier", tuicore.InputInt, "4")
	assertFieldValue(t, form, "obs_trace_cleanup_interval", tuicore.InputText, "15m0s")

	formatField := fieldByKey(form, "obs_metrics_format")
	if got, want := formatField.Options, []string{"json", "prometheus"}; len(got) != len(want) {
		t.Fatalf("obs_metrics_format options: want %v, got %v", want, got)
	}
	for i, want := range []string{"json", "prometheus"} {
		if got := formatField.Options[i]; got != want {
			t.Errorf("obs_metrics_format option %d: want %q, got %q", i, want, got)
		}
	}
}

func TestNewObservabilityForm_UsesTraceAndMetricsDefaults(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Observability.Metrics.Format = ""
	cfg.Observability.TraceStore = config.TraceStoreConfig{}

	form := NewObservabilityForm(cfg)
	defaults := config.TraceStoreDefaults()

	assertFieldValue(t, form, "obs_metrics_format", tuicore.InputSelect, "json")
	assertFieldValue(t, form, "obs_trace_max_age", tuicore.InputText, defaults.MaxAge.String())
	assertFieldValue(t, form, "obs_trace_max_traces", tuicore.InputInt, "10000")
	assertFieldValue(t, form, "obs_trace_failed_multiplier", tuicore.InputInt, "2")
	assertFieldValue(t, form, "obs_trace_cleanup_interval", tuicore.InputText, defaults.CleanupInterval.String())
}

func TestNewObservabilityForm_TraceFieldsFollowEnabledVisibility(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Observability.Enabled = false
	form := NewObservabilityForm(cfg)

	traceField := fieldByKey(form, "obs_trace_max_age")
	if traceField == nil {
		t.Fatal("missing obs_trace_max_age field")
	}
	if traceField.IsVisible() {
		t.Fatal("trace fields should be hidden when observability is disabled")
	}

	enabledField := fieldByKey(form, "obs_enabled")
	if enabledField == nil {
		t.Fatal("missing obs_enabled field")
	}
	enabledField.Checked = true
	if !traceField.IsVisible() {
		t.Fatal("trace fields should be visible when observability is enabled")
	}
}

func TestNewObservabilityForm_Validators(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewObservabilityForm(cfg)

	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{name: "token retention accepts positive integer", key: "obs_tokens_retention", value: "1"},
		{name: "token retention rejects zero", key: "obs_tokens_retention", value: "0", wantErr: true},
		{name: "token retention rejects text", key: "obs_tokens_retention", value: "many", wantErr: true},
		{name: "audit retention accepts positive integer", key: "obs_audit_retention", value: "90"},
		{name: "audit retention rejects negative", key: "obs_audit_retention", value: "-1", wantErr: true},
		{name: "trace max age accepts duration", key: "obs_trace_max_age", value: "720h"},
		{name: "trace max age rejects invalid duration", key: "obs_trace_max_age", value: "30 days", wantErr: true},
		{name: "trace max count accepts positive integer", key: "obs_trace_max_traces", value: "10"},
		{name: "trace max count rejects zero", key: "obs_trace_max_traces", value: "0", wantErr: true},
		{name: "failed trace multiplier accepts positive integer", key: "obs_trace_failed_multiplier", value: "3"},
		{name: "failed trace multiplier rejects text", key: "obs_trace_failed_multiplier", value: "double", wantErr: true},
		{name: "cleanup interval accepts duration", key: "obs_trace_cleanup_interval", value: "30m"},
		{name: "cleanup interval rejects invalid duration", key: "obs_trace_cleanup_interval", value: "hourly", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := fieldByKey(form, tt.key)
			if field == nil {
				t.Fatalf("missing field %q", tt.key)
			}
			if field.Validate == nil {
				t.Fatalf("field %q has no validator", tt.key)
			}

			err := field.Validate(tt.value)
			if tt.wantErr && err == nil {
				t.Fatalf("expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}

func assertFieldChecked(t *testing.T, form *tuicore.FormModel, key string, want bool) {
	t.Helper()

	field := fieldByKey(form, key)
	if field == nil {
		t.Fatalf("missing field %q", key)
	}
	if field.Type != tuicore.InputBool {
		t.Fatalf("%s: want InputBool, got %d", key, field.Type)
	}
	if field.Checked != want {
		t.Fatalf("%s: want checked %v, got %v", key, want, field.Checked)
	}
}

func assertFieldValue(
	t *testing.T,
	form *tuicore.FormModel,
	key string,
	wantType tuicore.InputType,
	wantValue string,
) {
	t.Helper()

	field := fieldByKey(form, key)
	if field == nil {
		t.Fatalf("missing field %q", key)
	}
	if field.Type != wantType {
		t.Fatalf("%s: want type %d, got %d", key, wantType, field.Type)
	}
	if field.Value != wantValue {
		t.Fatalf("%s: want value %q, got %q", key, wantValue, field.Value)
	}
}
