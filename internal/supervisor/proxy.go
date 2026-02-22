package supervisor

import (
	"context"
	"fmt"
	"iter"

	"github.com/langoai/lango/internal/provider"
)

// ProxyOption configures optional parameters for ProviderProxy.
type ProxyOption interface {
	apply(*proxyOptions)
}

type proxyOptions struct {
	temperature        float64
	maxTokens          int
	fallbackProviderID string
	fallbackModel      string
}

type temperatureOption float64

func (o temperatureOption) apply(opts *proxyOptions) { opts.temperature = float64(o) }

// WithTemperature sets the default temperature for generation requests.
func WithTemperature(t float64) ProxyOption { return temperatureOption(t) }

type maxTokensOption int

func (o maxTokensOption) apply(opts *proxyOptions) { opts.maxTokens = int(o) }

// WithMaxTokens sets the default max tokens for generation requests.
func WithMaxTokens(n int) ProxyOption { return maxTokensOption(n) }

type fallbackOption struct {
	providerID string
	model      string
}

func (o fallbackOption) apply(opts *proxyOptions) {
	opts.fallbackProviderID = o.providerID
	opts.fallbackModel = o.model
}

// WithFallback configures a fallback provider and model used when the primary fails.
func WithFallback(providerID, model string) ProxyOption {
	return fallbackOption{providerID: providerID, model: model}
}

// ProviderProxy implements provider.Provider but forwards requests to the Supervisor.
type ProviderProxy struct {
	supervisor         *Supervisor
	providerID         string
	defaultModel       string
	temperature        float64
	maxTokens          int
	fallbackProviderID string
	fallbackModel      string
}

// NewProviderProxy creates a new proxy for a specific provider.
func NewProviderProxy(sv *Supervisor, providerID, defaultModel string, opts ...ProxyOption) *ProviderProxy {
	var options proxyOptions
	for _, o := range opts {
		o.apply(&options)
	}

	return &ProviderProxy{
		supervisor:         sv,
		providerID:         providerID,
		defaultModel:       defaultModel,
		temperature:        options.temperature,
		maxTokens:          options.maxTokens,
		fallbackProviderID: options.fallbackProviderID,
		fallbackModel:      options.fallbackModel,
	}
}

// ID returns the provider ID.
func (p *ProviderProxy) ID() string {
	return p.providerID
}

// Generate forwards the request to the Supervisor.
func (p *ProviderProxy) Generate(ctx context.Context, params provider.GenerateParams) (iter.Seq2[provider.StreamEvent, error], error) {
	if params.Temperature == 0 && p.temperature != 0 {
		params.Temperature = p.temperature
	}
	if params.MaxTokens == 0 && p.maxTokens != 0 {
		params.MaxTokens = p.maxTokens
	}

	stream, err := p.supervisor.Generate(ctx, p.providerID, p.defaultModel, params)
	if err != nil && p.fallbackProviderID != "" {
		logger.Warnw("primary provider failed, trying fallback", "provider", p.providerID, "fallback", p.fallbackProviderID, "error", err)

		stream, err = p.supervisor.Generate(ctx, p.fallbackProviderID, p.fallbackModel, params)
		if err != nil {
			return nil, fmt.Errorf("fallback provider %q: %w", p.fallbackProviderID, err)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("provider %q: %w", p.providerID, err)
	}

	return stream, nil
}

// ListModels is not proxied yet; returns empty list. Implement when model listing via proxy is needed.
func (p *ProviderProxy) ListModels(ctx context.Context) ([]provider.ModelInfo, error) {
	return []provider.ModelInfo{}, nil
}
