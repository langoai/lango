package configcmd

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/langoai/lango/internal/config"
)

func TestWave52ConfigSetFromEnvArgValidationBranches(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{
			name:     "blank env name",
			args:     []string{"agent.provider", "--from-env", " \t "},
			contains: "--from-env requires an environment variable name",
		},
		{
			name:     "missing path",
			args:     []string{"--from-env", "LANGO_WAVE52_PROVIDER"},
			contains: "accepts 1 arg with --from-env: <dot.path>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loadCalled := false
			saveCalled := false
			cmd := NewSetCmd(
				func() (*config.Config, map[string]bool, func(), error) {
					loadCalled = true
					return config.DefaultConfig(), nil, func() {}, nil
				},
				func(*config.Config, map[string]bool) error {
					saveCalled = true
					return nil
				},
			)

			out, err := executeConfigCommand(t, cmd, tt.args...)
			if err == nil {
				t.Fatal("expected argument validation error")
			}
			if !strings.Contains(err.Error(), tt.contains) {
				t.Fatalf("expected error %q to contain %q", err.Error(), tt.contains)
			}
			if out != "" {
				t.Fatalf("expected no command output, got %q", out)
			}
			if loadCalled {
				t.Fatal("cfgLoader must not run after argument validation failure")
			}
			if saveCalled {
				t.Fatal("cfgSaver must not run after argument validation failure")
			}
		})
	}
}

func TestWave52ConfigPathHelpersCoverEmptyAndFallbackBranches(t *testing.T) {
	if got := splitConfigPath(""); got != nil {
		t.Fatalf("expected empty path to split to nil, got %#v", got)
	}
	if got := parentConfigPrefix("agent"); got != "" {
		t.Fatalf("expected top-level parent prefix to be empty, got %q", got)
	}
	if got := joinConfigKey("", "agent"); got != "agent" {
		t.Fatalf("expected empty-prefix join to return segment, got %q", got)
	}

	type jsonTagged struct {
		Field string `json:"fieldName"`
	}
	field := reflect.TypeOf(jsonTagged{}).Field(0)
	if got := configFieldPathName(field); got != "fieldName" {
		t.Fatalf("expected json tag fallback, got %q", got)
	}

	if got := uniqueNonEmptyStrings([]string{"agent.provider", "", "agent.provider"}); !reflect.DeepEqual(got, []string{"agent.provider"}) {
		t.Fatalf("expected unique non-empty values, got %#v", got)
	}
}

func TestWave52SensitiveSegmentVariantsRemainRedacted(t *testing.T) {
	sensitive := []string{
		"webhookURL",
		"credential",
		"credentials",
		"userCredential",
		"clientCredentials",
		"privateKey",
		"accessKey",
	}

	for _, segment := range sensitive {
		t.Run(segment, func(t *testing.T) {
			if !configSetSegmentIsSensitive(normalizeConfigPathSegment(segment)) {
				t.Fatalf("expected %q to be classified as sensitive", segment)
			}
		})
	}
}

func TestWave52RedactConfigGetValueNilAndJSONMarshalFallback(t *testing.T) {
	if got := redactConfigGetValue("agent.provider", nil, "plain"); got != nil {
		t.Fatalf("expected nil value to remain nil, got %#v", got)
	}

	if decoded, ok := configGetJSONCompatibleValue(func() {}); ok || decoded != nil {
		t.Fatalf("expected non-marshalable value to fail JSON normalization, got %#v ok=%v", decoded, ok)
	}
}

func TestWave52RedactReflectValuePreservesJSONMarshalers(t *testing.T) {
	value := wave52JSONMarshaler{Value: "raw-secret"}
	got := redactConfigGetReflectValue([]string{"custom"}, reflect.ValueOf(value))

	if _, ok := got.(wave52JSONMarshaler); !ok {
		t.Fatalf("expected json.Marshaler struct to be preserved, got %T", got)
	}

	got = redactConfigGetReflectValue([]string{"custom"}, reflect.ValueOf(&value))
	if _, ok := got.(*wave52JSONMarshaler); !ok {
		t.Fatalf("expected json.Marshaler pointer to be preserved, got %T", got)
	}
}

type wave52JSONMarshaler struct {
	Value string
}

func (m wave52JSONMarshaler) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"value": m.Value})
}
