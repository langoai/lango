package configcmd

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/langoai/lango/internal/config"
)

type configGetTestJSONMarshaler struct {
	Value string
}

func (m configGetTestJSONMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(`"` + m.Value + `"`), nil
}

func TestFormatPlainScalarAndReflectionValues(t *testing.T) {
	type namedString string

	values := map[string]interface{}{
		"bool":         false,
		"float":        1.25,
		"named string": namedString("named"),
		"array":        [2]int{2, 1},
		"sorted map": map[string]int{
			"z": 26,
			"a": 1,
		},
	}

	tests := []struct {
		name string
		give interface{}
		want string
	}{
		{name: "bool", give: values["bool"], want: "false"},
		{name: "float", give: values["float"], want: "1.25"},
		{name: "named string", give: values["named string"], want: "named"},
		{name: "array default formatting", give: values["array"], want: "[2 1]"},
		{name: "map keys sorted", give: values["sorted map"], want: "a=1,z=26"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatPlain(tt.give); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestRedactConfigGetValueCoversReflectPointersMapsAndSlices(t *testing.T) {
	type nested struct {
		Token string `mapstructure:"token" json:"token"`
		Name  string `mapstructure:"name" json:"name"`
	}
	type reflectedConfig struct {
		APIKey     *string           `mapstructure:"apiKey" json:"apiKey"`
		NilSecret  *string           `mapstructure:"clientSecret" json:"clientSecret"`
		Children   []nested          `mapstructure:"children" json:"children"`
		Headers    map[string]string `mapstructure:"headers" json:"headers"`
		Unredacted string            `mapstructure:"label" json:"label"`
	}

	secret := "raw-api-key"
	value := reflectedConfig{
		APIKey: &secret,
		Children: []nested{
			{Token: "child-token", Name: "child-name"},
		},
		Headers: map[string]string{
			"Authorization": "Bearer secret",
			"Trace":         "trace-1",
		},
		Unredacted: "visible",
	}

	redacted, ok := redactConfigGetValue("custom", value, "plain").(map[string]interface{})
	if !ok {
		t.Fatalf("expected reflected struct map, got %T", redacted)
	}
	if redacted["apiKey"] != "<redacted>" {
		t.Fatalf("expected pointer apiKey to be redacted, got %v", redacted["apiKey"])
	}
	if redacted["clientSecret"] != "<redacted>" {
		t.Fatalf("expected nil secret pointer path to be redacted, got %v", redacted["clientSecret"])
	}
	if redacted["label"] != "visible" {
		t.Fatalf("expected non-sensitive scalar to remain visible, got %v", redacted["label"])
	}

	children, ok := redacted["children"].([]interface{})
	if !ok {
		t.Fatalf("expected reflected slice, got %T", redacted["children"])
	}
	if len(children) != 1 {
		t.Fatalf("expected one reflected child, got %d", len(children))
	}
	child, ok := children[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected reflected child map, got %T", children[0])
	}
	if child["token"] != "<redacted>" {
		t.Fatalf("expected nested token to be redacted, got %v", child["token"])
	}
	if child["name"] != "child-name" {
		t.Fatalf("expected nested name to remain visible, got %v", child["name"])
	}

	headers, ok := redacted["headers"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected reflected headers map, got %T", redacted["headers"])
	}
	if headers["Authorization"] != "<redacted>" {
		t.Fatalf("expected Authorization header to be redacted, got %v", headers["Authorization"])
	}
	if headers["Trace"] != "trace-1" {
		t.Fatalf("expected Trace header to remain visible, got %v", headers["Trace"])
	}
}

func TestRedactConfigGetValueJSONRedactsNestedGenericValues(t *testing.T) {
	value := map[string]interface{}{
		"service": map[string]interface{}{
			"webhookURL": "https://secret.example/hook",
			"name":       "alerts",
		},
		"items": []interface{}{
			map[string]interface{}{
				"accessToken": "raw-token",
				"id":          "item-1",
			},
		},
	}

	var out bytes.Buffer
	if err := printValue(&out, redactConfigGetValue("root", value, "json"), "json"); err != nil {
		t.Fatalf("print redacted JSON: %v", err)
	}
	if strings.Contains(out.String(), "secret.example") || strings.Contains(out.String(), "raw-token") {
		t.Fatalf("redacted JSON must not contain raw secrets, got %q", out.String())
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("decode redacted JSON: %v", err)
	}
	service, ok := decoded["service"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected service object, got %T", decoded["service"])
	}
	if service["webhookURL"] != "<redacted>" {
		t.Fatalf("expected webhookURL redaction, got %v", service["webhookURL"])
	}
	items, ok := decoded["items"].([]interface{})
	if !ok {
		t.Fatalf("expected items array, got %T", decoded["items"])
	}
	if len(items) != 1 {
		t.Fatalf("expected one item, got %d", len(items))
	}
	item, ok := items[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected first item object, got %T", items[0])
	}
	if item["accessToken"] != "<redacted>" {
		t.Fatalf("expected accessToken redaction, got %v", item["accessToken"])
	}
	if item["id"] != "item-1" {
		t.Fatalf("expected non-sensitive item id to remain visible, got %v", item["id"])
	}
}

func TestResolveAndSetInvalidPathsReturnDiscoveryHints(t *testing.T) {
	cfg := config.DefaultConfig()

	tests := []struct {
		name     string
		run      func() error
		contains []string
	}{
		{
			name: "resolve map key missing",
			run: func() error {
				_, err := resolveConfigPath(cfg, "providers.missing.apiKey")
				return err
			},
			contains: []string{`key "missing" not found in map`},
		},
		{
			name: "set through scalar leaf",
			run: func() error {
				return setConfigPath(cfg, "agent.provider.name", "openai")
			},
			contains: []string{`"name" is not a struct`, "did you mean: agent.provider", "lango config keys agent"},
		},
		{
			name: "set whole struct map entry",
			run: func() error {
				return setConfigPath(cfg, "providers.openai", "raw")
			},
			contains: []string{"unsupported map value type config.ProviderConfig"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run()
			if err == nil {
				t.Fatal("expected error")
			}
			for _, fragment := range tt.contains {
				if !strings.Contains(err.Error(), fragment) {
					t.Fatalf("expected error %q to contain %q", err.Error(), fragment)
				}
			}
		})
	}
}

func TestSetFieldRejectsInvalidScalarInputs(t *testing.T) {
	type unsupportedFields struct {
		Strings []int
		Delay   time.Duration
	}

	tests := []struct {
		name     string
		field    reflect.Value
		rawValue string
		contains string
	}{
		{
			name:     "invalid bool",
			field:    reflect.ValueOf(new(bool)).Elem(),
			rawValue: "truthy",
			contains: `invalid bool "truthy"`,
		},
		{
			name:     "invalid integer",
			field:    reflect.ValueOf(new(int)).Elem(),
			rawValue: "12.5",
			contains: `invalid integer "12.5"`,
		},
		{
			name:     "invalid unsigned integer",
			field:    reflect.ValueOf(new(uint64)).Elem(),
			rawValue: "-1",
			contains: `invalid unsigned integer "-1"`,
		},
		{
			name:     "invalid float",
			field:    reflect.ValueOf(new(float64)).Elem(),
			rawValue: "warm",
			contains: `invalid float "warm"`,
		},
		{
			name:     "unsupported slice type",
			field:    reflect.ValueOf(&unsupportedFields{}).Elem().FieldByName("Strings"),
			rawValue: "1,2",
			contains: "unsupported slice type",
		},
		{
			name:     "duration guidance",
			field:    reflect.ValueOf(&unsupportedFields{}).Elem().FieldByName("Delay"),
			rawValue: "5s",
			contains: "duration fields should use 'lango settings' TUI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setField(tt.field, tt.rawValue, "viewSkillRejectsSymlinkEscape8.path")
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.contains) {
				t.Fatalf("expected error %q to contain %q", err.Error(), tt.contains)
			}
		})
	}
}

func TestConfigSetRejectsInvalidScalarBeforeSave(t *testing.T) {
	cfg := config.DefaultConfig()
	saveCalled := false
	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			saveCalled = true
			return nil
		},
	)

	out, err := executeConfigCommand(t, cmd, "agent.maxTokens", "not-an-int")
	if err == nil {
		t.Fatal("expected invalid integer error")
	}
	if !strings.Contains(err.Error(), `invalid integer "not-an-int"`) {
		t.Fatalf("expected invalid integer error, got %v", err)
	}
	if out != "" {
		t.Fatalf("expected no command output, got %q", out)
	}
	if saveCalled {
		t.Fatal("cfgSaver must not run after invalid scalar parsing")
	}
}

func TestConfigGetScalarAndJSONCompatibilityResidualBranches(t *testing.T) {
	var nilPointer *string
	if !isConfigGetScalarValue(reflect.ValueOf(nilPointer)) {
		t.Fatal("expected nil pointer to be treated as scalar")
	}

	value := "visible"
	interfaceValue := interface{}(&value)
	if !isConfigGetScalarValue(reflect.ValueOf(&interfaceValue)) {
		t.Fatal("expected pointer through interface to scalar string to be scalar")
	}

	if _, ok := configGetJSONCompatibleValue(func() {}); ok {
		t.Fatal("expected non-marshalable function value to be rejected")
	}
}

func TestRedactConfigGetReflectValuePreservesJSONMarshalers(t *testing.T) {
	value := configGetTestJSONMarshaler{Value: "custom"}
	got := redactConfigGetReflectValue([]string{"custom"}, reflect.ValueOf(value))
	if got != value {
		t.Fatalf("expected struct JSON marshaler to be preserved, got %#v", got)
	}

	pointer := &value
	got = redactConfigGetReflectValue([]string{"custom"}, reflect.ValueOf(pointer))
	if got != pointer {
		t.Fatalf("expected pointer JSON marshaler to be preserved, got %#v", got)
	}

	if got := redactConfigGetReflectValue([]string{"apiKey"}, reflect.ValueOf("raw")); got != "<redacted>" {
		t.Fatalf("expected sensitive reflect path to redact before scalar handling, got %#v", got)
	}
	if got := redactConfigGetReflectValue([]string{"missing"}, reflect.Value{}); got != nil {
		t.Fatalf("expected invalid reflect value to redact to nil, got %#v", got)
	}
}

func TestConfigPathAndSuggestionResidualBranches(t *testing.T) {
	if segments := splitConfigPath(""); segments != nil {
		t.Fatalf("expected empty path to split to nil, got %#v", segments)
	}

	if got := configPathDiscoveryError("bad path", "", []string{"agent.provider", "", "agent.provider"}).Error(); !strings.Contains(got, "did you mean: agent.provider") {
		t.Fatalf("expected duplicate suggestions to be collapsed, got %q", got)
	}

	if prefix := parentConfigPrefix("agent"); prefix != "" {
		t.Fatalf("expected top-level parent prefix to be empty, got %q", prefix)
	}

	if got := nearestConfigKeys("zzzz", []string{"agent.provider", "p2p.enabled"}); len(got) != 0 {
		t.Fatalf("expected no nearby config keys, got %#v", got)
	}
	if got := nearestConfigKeys("abc", []string{"abc", "abd", "abe", "abf", "abg"}); len(got) != 3 {
		t.Fatalf("expected nearest config keys to be capped at three, got %#v", got)
	}

	if got := editDistance("", "abc"); got != 3 {
		t.Fatalf("expected insertion distance 3, got %d", got)
	}
	if got := editDistance("abc", ""); got != 3 {
		t.Fatalf("expected deletion distance 3, got %d", got)
	}
}

func TestConfigCollectionResidualBranches(t *testing.T) {
	type unsupportedLeaf struct {
		Delay time.Duration `mapstructure:"delay"`
		Blob  []int         `mapstructure:"blob"`
		Name  string        `mapstructure:"name"`
	}
	type nestedMaps struct {
		ByInt    map[int]string              `mapstructure:"byInt"`
		ByName   map[string]string           `mapstructure:"byName"`
		Nested   map[string]map[string]int   `mapstructure:"nested"`
		Profiles map[string]*unsupportedLeaf `mapstructure:"profiles"`
	}

	if got := collectMapKeys(reflect.TypeOf(map[int]string{}), "bad"); got != nil {
		t.Fatalf("expected non-string map keys to be skipped, got %#v", got)
	}

	keys := collectKeys(reflect.TypeOf(nestedMaps{}), "")
	for _, want := range []string{
		"byInt",
		"byName",
		"byName.<key>",
		"nested.<name>.<key>",
		"profiles.<name>.name",
	} {
		if !containsString(keys, want) {
			t.Fatalf("expected collected keys to include %q, got %#v", want, keys)
		}
	}
	if containsString(keys, "profiles.<name>.delay") || containsString(keys, "profiles.<name>.blob") {
		t.Fatalf("expected unsupported settable fields to be omitted, got %#v", keys)
	}

	if !configSetFieldTypeIsSupported(reflect.TypeOf((*string)(nil))) {
		t.Fatal("expected pointer to supported scalar type to be settable")
	}
	if configSetFieldTypeIsSupported(reflect.TypeOf(time.Second)) {
		t.Fatal("expected time.Duration to be rejected as a settable config field")
	}

	if got := collectSettableConfigKeys(reflect.TypeOf(map[string]string{}), "labels"); !reflect.DeepEqual(got, []string{"labels.<key>"}) {
		t.Fatalf("expected settable map keys, got %#v", got)
	}
	if got := collectSettableConfigKeys(reflect.TypeOf(""), "name"); !reflect.DeepEqual(got, []string{"name"}) {
		t.Fatalf("expected direct scalar settable key, got %#v", got)
	}
	if got := collectSettableConfigKeys(reflect.TypeOf([]int{}), "numbers"); got != nil {
		t.Fatalf("expected unsupported direct leaf to be skipped, got %#v", got)
	}
}

func TestConfigFieldNamesAndSensitiveSegmentsResidualBranches(t *testing.T) {
	type tagged struct {
		EmptyMapstructure string `mapstructure:",squash" json:"jsonName,omitempty"`
		EmptyJSON         string `json:",omitempty"`
		Plain             string
	}

	typ := reflect.TypeOf(tagged{})
	if got := configFieldPathName(typ.Field(0)); got != "jsonName" {
		t.Fatalf("expected json fallback name, got %q", got)
	}
	if got := configFieldPathName(typ.Field(1)); got != "EmptyJSON" {
		t.Fatalf("expected field name fallback, got %q", got)
	}
	if got := configFieldPathName(typ.Field(2)); got != "Plain" {
		t.Fatalf("expected plain field name fallback, got %q", got)
	}

	for _, segment := range []string{"wallet_private-key", "aws_access-key", "session.credentials"} {
		if !configSetSegmentIsSensitive(normalizeConfigPathSegment(segment)) {
			t.Fatalf("expected %q to be sensitive", segment)
		}
	}
}

func TestSetConfigValueAndPrintResidualErrors(t *testing.T) {
	err := setConfigValue(reflect.ValueOf(&struct{}{}).Elem(), nil, "value", "empty", nil)
	if err == nil || !strings.Contains(err.Error(), "empty path") {
		t.Fatalf("expected empty path error, got %v", err)
	}

	type customMap struct {
		Values map[int]string `mapstructure:"values"`
	}
	holder := customMap{}
	err = setConfigValue(reflect.ValueOf(&holder).Elem(), []string{"values", "one"}, "value", "values.one", nil)
	if err == nil || !strings.Contains(err.Error(), "unsupported map key type int") {
		t.Fatalf("expected unsupported map key error, got %v", err)
	}

	var out bytes.Buffer
	err = printValue(&out, func() {}, "json")
	if err == nil || !strings.Contains(err.Error(), "marshal value") {
		t.Fatalf("expected JSON marshal error, got %v", err)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
