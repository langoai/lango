package types

import (
	"reflect"
	"testing"
)

func TestChannelType_ValidAndValues(t *testing.T) {
	var enum ChannelType
	wantValues := []ChannelType{ChannelTelegram, ChannelDiscord, ChannelSlack}

	if got := enum.Values(); !reflect.DeepEqual(got, wantValues) {
		t.Fatalf("Values() = %#v, want %#v", got, wantValues)
	}

	for _, value := range wantValues {
		if !value.Valid() {
			t.Fatalf("%q should be valid", value)
		}
	}

	for _, value := range []ChannelType{"", "TELEGRAM", "matrix"} {
		if value.Valid() {
			t.Fatalf("%q should be invalid", value)
		}
	}
}

func TestProviderType_ValidAndValues(t *testing.T) {
	var enum ProviderType
	wantValues := []ProviderType{
		ProviderOpenAI,
		ProviderAnthropic,
		ProviderGemini,
		ProviderGoogle,
		ProviderOllama,
		ProviderGitHub,
	}

	if got := enum.Values(); !reflect.DeepEqual(got, wantValues) {
		t.Fatalf("Values() = %#v, want %#v", got, wantValues)
	}

	for _, value := range wantValues {
		if !value.Valid() {
			t.Fatalf("%q should be valid", value)
		}
	}

	for _, value := range []ProviderType{"", "OpenAI", "azure"} {
		if value.Valid() {
			t.Fatalf("%q should be invalid", value)
		}
	}
}

func TestConfidence_ValidAndValues(t *testing.T) {
	var enum Confidence
	wantValues := []Confidence{ConfidenceHigh, ConfidenceMedium, ConfidenceLow}

	if got := enum.Values(); !reflect.DeepEqual(got, wantValues) {
		t.Fatalf("Values() = %#v, want %#v", got, wantValues)
	}

	for _, value := range wantValues {
		if !value.Valid() {
			t.Fatalf("%q should be valid", value)
		}
	}

	for _, value := range []Confidence{"", "HIGH", "unknown"} {
		if value.Valid() {
			t.Fatalf("%q should be invalid", value)
		}
	}
}

func TestMessageRole_ValidValuesAndNormalize(t *testing.T) {
	var enum MessageRole
	wantValues := []MessageRole{RoleUser, RoleAssistant, RoleTool, RoleFunction, RoleModel}

	if got := enum.Values(); !reflect.DeepEqual(got, wantValues) {
		t.Fatalf("Values() = %#v, want %#v", got, wantValues)
	}

	for _, value := range wantValues {
		if !value.Valid() {
			t.Fatalf("%q should be valid", value)
		}
	}

	for _, value := range []MessageRole{"", "USER", "system"} {
		if value.Valid() {
			t.Fatalf("%q should be invalid", value)
		}
	}

	tests := []struct {
		name string
		role MessageRole
		want MessageRole
	}{
		{name: "model becomes assistant", role: RoleModel, want: RoleAssistant},
		{name: "function becomes tool", role: RoleFunction, want: RoleTool},
		{name: "user remains user", role: RoleUser, want: RoleUser},
		{name: "invalid remains unchanged", role: MessageRole("system"), want: MessageRole("system")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.Normalize(); got != tt.want {
				t.Fatalf("Normalize() = %q, want %q", got, tt.want)
			}
		})
	}
}
