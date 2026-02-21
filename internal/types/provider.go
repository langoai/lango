package types

// ProviderType represents an LLM provider type.
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
	ProviderGemini    ProviderType = "gemini"
	ProviderGoogle    ProviderType = "google"
	ProviderOllama    ProviderType = "ollama"
	ProviderGitHub    ProviderType = "github"
)

// Valid reports whether p is a known provider type.
func (p ProviderType) Valid() bool {
	switch p {
	case ProviderOpenAI, ProviderAnthropic, ProviderGemini, ProviderGoogle, ProviderOllama, ProviderGitHub:
		return true
	}
	return false
}

// Values returns all known provider types.
func (p ProviderType) Values() []ProviderType {
	return []ProviderType{ProviderOpenAI, ProviderAnthropic, ProviderGemini, ProviderGoogle, ProviderOllama, ProviderGitHub}
}
