package clitypes

import (
	"reflect"
	"testing"
)

func TestGetProviderMetadataKnownProviders(t *testing.T) {
	tests := []struct {
		id           string
		name         string
		envVar       string
		defaultModel string
	}{
		{id: "gemini", name: "Google Gemini", envVar: "GOOGLE_API_KEY", defaultModel: "gemini-2.0-flash-exp"},
		{id: "openai", name: "OpenAI", envVar: "OPENAI_API_KEY", defaultModel: "gpt-4o"},
		{id: "anthropic", name: "Anthropic Claude", envVar: "ANTHROPIC_API_KEY", defaultModel: "claude-3-5-sonnet-20241022"},
		{id: "ollama", name: "Ollama", envVar: "", defaultModel: "llama3"},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			meta, ok := GetProviderMetadata(tt.id)
			if !ok {
				t.Fatalf("GetProviderMetadata(%q) returned ok=false", tt.id)
			}
			if meta.ID != tt.id {
				t.Fatalf("ID = %q, want %q", meta.ID, tt.id)
			}
			if meta.Name != tt.name {
				t.Fatalf("Name = %q, want %q", meta.Name, tt.name)
			}
			if meta.EnvVar != tt.envVar {
				t.Fatalf("EnvVar = %q, want %q", meta.EnvVar, tt.envVar)
			}
			if meta.DefaultModel != tt.defaultModel {
				t.Fatalf("DefaultModel = %q, want %q", meta.DefaultModel, tt.defaultModel)
			}
			if meta.Description == "" {
				t.Fatal("Description is empty")
			}
		})
	}
}

func TestGetProviderMetadataUnknownProvider(t *testing.T) {
	meta, ok := GetProviderMetadata("unknown")
	if ok {
		t.Fatalf("GetProviderMetadata returned ok=true for unknown provider: %+v", meta)
	}
	if meta != (ProviderMetadata{}) {
		t.Fatalf("unknown provider metadata = %+v, want zero value", meta)
	}
}

func TestGetSupportedProvidersReturnsMetadataInProviderOrder(t *testing.T) {
	metas := GetSupportedProviders()
	var gotIDs []string
	for _, meta := range metas {
		gotIDs = append(gotIDs, meta.ID)
		if meta.Name == "" {
			t.Fatalf("provider %q has empty Name", meta.ID)
		}
		if meta.Description == "" {
			t.Fatalf("provider %q has empty Description", meta.ID)
		}
		if meta.DefaultModel == "" {
			t.Fatalf("provider %q has empty DefaultModel", meta.ID)
		}
	}

	wantIDs := []string{"gemini", "openai", "anthropic", "ollama"}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("supported provider IDs = %#v, want %#v", gotIDs, wantIDs)
	}
}

func TestGetAPIKeyReadsKnownProviderEnvironment(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-openai-key")

	if got := GetAPIKey("openai"); got != "test-openai-key" {
		t.Fatalf("GetAPIKey(openai) = %q, want %q", got, "test-openai-key")
	}
	if got := GetAPIKey("unknown"); got != "" {
		t.Fatalf("GetAPIKey(unknown) = %q, want empty string", got)
	}
}
