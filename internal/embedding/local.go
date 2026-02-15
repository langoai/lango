package embedding

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// LocalProvider generates embeddings using a local Ollama instance
// via the OpenAI-compatible API endpoint.
type LocalProvider struct {
	client     *openai.Client
	model      openai.EmbeddingModel
	dimensions int
}

// NewLocalProvider creates a new local embedding provider (Ollama).
func NewLocalProvider(baseURL, model string, dimensions int) *LocalProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434/v1"
	}
	if model == "" {
		model = "nomic-embed-text"
	}
	if dimensions <= 0 {
		dimensions = 768
	}

	cfg := openai.DefaultConfig("")
	cfg.BaseURL = baseURL
	client := openai.NewClientWithConfig(cfg)

	return &LocalProvider{
		client:     client,
		model:      openai.EmbeddingModel(model),
		dimensions: dimensions,
	}
}

func (p *LocalProvider) ID() string       { return "local" }
func (p *LocalProvider) Dimensions() int  { return p.dimensions }

// Embed generates embeddings for the given texts.
func (p *LocalProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	resp, err := p.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input:      texts,
		Model:      p.model,
		Dimensions: p.dimensions,
	})
	if err != nil {
		return nil, fmt.Errorf("local embeddings: %w", err)
	}

	result := make([][]float32, len(resp.Data))
	for i, d := range resp.Data {
		result[i] = d.Embedding
	}
	return result, nil
}
