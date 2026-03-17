package supervisor

import (
	"context"
	"errors"
	"iter"
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/provider"
)

// mockProvider records calls and returns pre-configured results.
type mockProvider struct {
	id        string
	calls     []mockCall
	generateFn func(ctx context.Context, params provider.GenerateParams) (iter.Seq2[provider.StreamEvent, error], error)
}

type mockCall struct {
	Model string
}

func (m *mockProvider) ID() string { return m.id }

func (m *mockProvider) Generate(ctx context.Context, params provider.GenerateParams) (iter.Seq2[provider.StreamEvent, error], error) {
	m.calls = append(m.calls, mockCall{Model: params.Model})
	return m.generateFn(ctx, params)
}

func (m *mockProvider) ListModels(ctx context.Context) ([]provider.ModelInfo, error) {
	return nil, nil
}

func emptyStream() (iter.Seq2[provider.StreamEvent, error], error) {
	return func(yield func(provider.StreamEvent, error) bool) {
		yield(provider.StreamEvent{Type: provider.StreamEventDone}, nil)
	}, nil
}

func failStream() (iter.Seq2[provider.StreamEvent, error], error) {
	return nil, errors.New("provider unavailable")
}

func TestProxyFallback(t *testing.T) {
	tests := []struct {
		give               string
		primaryFail        bool
		fallbackFail       bool
		wantFallbackCalled bool
		wantErr            bool
		wantFallbackModel  string // model arg passed to fallback in Supervisor.Generate
	}{
		{
			give:               "primary succeeds",
			primaryFail:        false,
			wantFallbackCalled: false,
			wantErr:            false,
		},
		{
			give:               "primary fails, fallback succeeds with correct model",
			primaryFail:        true,
			fallbackFail:       false,
			wantFallbackCalled: true,
			wantErr:            false,
			wantFallbackModel:  "gemini-3-flash-preview",
		},
		{
			give:               "both fail",
			primaryFail:        true,
			fallbackFail:       true,
			wantFallbackCalled: true,
			wantErr:            true,
			wantFallbackModel:  "gemini-3-flash-preview",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			primary := &mockProvider{
				id: "openai-1",
				generateFn: func(_ context.Context, _ provider.GenerateParams) (iter.Seq2[provider.StreamEvent, error], error) {
					if tt.primaryFail {
						return failStream()
					}
					return emptyStream()
				},
			}
			fallback := &mockProvider{
				id: "gemini-1",
				generateFn: func(_ context.Context, _ provider.GenerateParams) (iter.Seq2[provider.StreamEvent, error], error) {
					if tt.fallbackFail {
						return failStream()
					}
					return emptyStream()
				},
			}

			reg := provider.NewRegistry()
			reg.Register(primary)
			reg.Register(fallback)

			sv := &Supervisor{
				Config:   &config.Config{},
				registry: reg,
			}

			proxy := NewProviderProxy(sv, "openai-1", "gpt-5.3-codex",
				WithFallback("gemini-1", "gemini-3-flash-preview"),
			)

			params := provider.GenerateParams{
				Model:    "gpt-5.3-codex",
				Messages: []provider.Message{{Role: "user", Content: "hello"}},
			}

			_, err := proxy.Generate(context.Background(), params)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantFallbackCalled && len(fallback.calls) == 0 {
				t.Fatal("expected fallback to be called, but it was not")
			}
			if !tt.wantFallbackCalled && len(fallback.calls) > 0 {
				t.Fatal("expected fallback not to be called, but it was")
			}

			// Critical: verify that the fallback's params.Model was reset so
			// Supervisor.Generate applies the fallback model instead of carrying
			// over the primary model.
			if tt.wantFallbackCalled && len(fallback.calls) > 0 {
				gotModel := fallback.calls[0].Model
				if gotModel != tt.wantFallbackModel {
					t.Errorf("fallback received model %q, want %q", gotModel, tt.wantFallbackModel)
				}
			}

			// Verify original params were not mutated.
			if params.Model != "gpt-5.3-codex" {
				t.Errorf("original params.Model mutated to %q", params.Model)
			}
		})
	}
}
