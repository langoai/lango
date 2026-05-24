package config

import "testing"

func TestEvaluateAgentSetup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  *Config
		want AgentSetupStatus
	}{
		{
			name: "nil config",
			cfg:  nil,
			want: AgentSetupStatus{
				MissingProvider: true,
				MissingModel:    true,
			},
		},
		{
			name: "missing provider",
			cfg: &Config{
				Agent: AgentConfig{Provider: "", Model: "gpt-5"},
			},
			want: AgentSetupStatus{
				Model:           "gpt-5",
				MissingProvider: true,
			},
		},
		{
			name: "missing model",
			cfg: &Config{
				Agent: AgentConfig{Provider: "openai", Model: ""},
			},
			want: AgentSetupStatus{
				ProviderID:         "openai",
				MissingModel:       true,
				MissingProviderMap: true,
			},
		},
		{
			name: "missing provider map entry",
			cfg: &Config{
				Agent:     AgentConfig{Provider: "openai", Model: "gpt-5"},
				Providers: map[string]ProviderConfig{},
			},
			want: AgentSetupStatus{
				ProviderID:         "openai",
				Model:              "gpt-5",
				MissingProviderMap: true,
			},
		},
		{
			name: "missing provider type",
			cfg: &Config{
				Agent: AgentConfig{Provider: "openai", Model: "gpt-5"},
				Providers: map[string]ProviderConfig{
					"openai": {},
				},
			},
			want: AgentSetupStatus{
				ProviderID:          "openai",
				Model:               "gpt-5",
				MissingProviderType: true,
			},
		},
		{
			name: "missing api key for remote provider",
			cfg: &Config{
				Agent: AgentConfig{Provider: "openai", Model: "gpt-5"},
				Providers: map[string]ProviderConfig{
					"openai": {Type: "openai"},
				},
			},
			want: AgentSetupStatus{
				ProviderID:    "openai",
				Model:         "gpt-5",
				ProviderType:  "openai",
				MissingAPIKey: true,
			},
		},
		{
			name: "ollama requires no api key",
			cfg: &Config{
				Agent: AgentConfig{Provider: "local-ollama", Model: "llama3.1"},
				Providers: map[string]ProviderConfig{
					"local-ollama": {Type: "ollama"},
				},
			},
			want: AgentSetupStatus{
				ProviderID:   "local-ollama",
				Model:        "llama3.1",
				ProviderType: "ollama",
			},
		},
		{
			name: "remote provider ready",
			cfg: &Config{
				Agent: AgentConfig{Provider: "openai", Model: "gpt-5"},
				Providers: map[string]ProviderConfig{
					"openai": {Type: "openai", APIKey: "sk-test"},
				},
			},
			want: AgentSetupStatus{
				ProviderID:   "openai",
				Model:        "gpt-5",
				ProviderType: "openai",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvaluateAgentSetup(tt.cfg)
			if got != tt.want {
				t.Fatalf("EvaluateAgentSetup() mismatch\nwant: %#v\ngot:  %#v", tt.want, got)
			}
		})
	}
}

func TestAgentSetupStatusReady(t *testing.T) {
	t.Parallel()

	if !EvaluateAgentSetup(&Config{
		Agent: AgentConfig{Provider: "openai", Model: "gpt-5"},
		Providers: map[string]ProviderConfig{
			"openai": {Type: "openai", APIKey: "sk-test"},
		},
	}).Ready() {
		t.Fatal("expected ready remote provider config")
	}

	if EvaluateAgentSetup(DefaultConfig()).Ready() {
		t.Fatal("default config should not be ready")
	}
}
