package embedding

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIProvider generates embeddings using OpenAI's API.
type OpenAIProvider struct {
	client     *openai.Client
	model      openai.EmbeddingModel
	dimensions int
}

// NewOpenAIProvider creates a new OpenAI embedding provider.
func NewOpenAIProvider(apiKey string, model string, dimensions int) *OpenAIProvider {
	if model == "" {
		model = "text-embedding-3-small"
	}
	if dimensions <= 0 {
		dimensions = 1536
	}
	client := openai.NewClient(apiKey)
	return &OpenAIProvider{
		client:     client,
		model:      openai.EmbeddingModel(model),
		dimensions: dimensions,
	}
}

func (p *OpenAIProvider) ID() string       { return "openai" }
func (p *OpenAIProvider) Dimensions() int  { return p.dimensions }

// Embed generates embeddings for the given texts.
func (p *OpenAIProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	resp, err := p.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input:      texts,
		Model:      p.model,
		Dimensions: p.dimensions,
	})
	if err != nil {
		return nil, fmt.Errorf("openai embeddings: %w", err)
	}

	result := make([][]float32, len(resp.Data))
	for i, d := range resp.Data {
		result[i] = d.Embedding
	}
	return result, nil
}
