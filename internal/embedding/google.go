package embedding

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// GoogleProvider generates embeddings using Google's Generative AI API.
type GoogleProvider struct {
	client     *genai.Client
	model      string
	dimensions int
}

// NewGoogleProvider creates a new Google embedding provider.
func NewGoogleProvider(apiKey string, model string, dimensions int) (*GoogleProvider, error) {
	if model == "" {
		model = "text-embedding-004"
	}
	if dimensions <= 0 {
		dimensions = 768
	}

	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create google client: %w", err)
	}

	return &GoogleProvider{
		client:     client,
		model:      model,
		dimensions: dimensions,
	}, nil
}

func (p *GoogleProvider) ID() string       { return "google" }
func (p *GoogleProvider) Dimensions() int  { return p.dimensions }

// Embed generates embeddings for the given texts.
func (p *GoogleProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	contents := make([]*genai.Content, len(texts))
	for i, t := range texts {
		contents[i] = genai.NewContentFromText(t, genai.RoleUser)
	}

	dim := int32(p.dimensions)
	cfg := &genai.EmbedContentConfig{
		OutputDimensionality: &dim,
	}
	resp, err := p.client.Models.EmbedContent(ctx, p.model, contents, cfg)
	if err != nil {
		return nil, fmt.Errorf("google embeddings: %w", err)
	}

	result := make([][]float32, len(resp.Embeddings))
	for i, emb := range resp.Embeddings {
		result[i] = emb.Values
	}
	return result, nil
}
