package adk

import (
	"context"
	"iter"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/logging"
	"google.golang.org/adk/model"
)

var piiModelLogger = logging.SubsystemSugar("pii-model")

// PIIRedactingModelAdapter wraps an LLM and redacts PII from user messages
// before forwarding them to the underlying model. It also scans tool results
// and model responses for known secret values.
type PIIRedactingModelAdapter struct {
	inner    model.LLM
	redactor *agent.PIIRedactor
	scanner  *agent.SecretScanner
}

// NewPIIRedactingModelAdapter creates a PII-redacting model adapter.
// If scanner is nil, output scanning is disabled.
func NewPIIRedactingModelAdapter(
	inner model.LLM,
	redactor *agent.PIIRedactor,
	scanner *agent.SecretScanner,
) *PIIRedactingModelAdapter {
	return &PIIRedactingModelAdapter{
		inner:    inner,
		redactor: redactor,
		scanner:  scanner,
	}
}

// Name delegates to the inner adapter.
func (m *PIIRedactingModelAdapter) Name() string {
	return m.inner.Name()
}

// GenerateContent redacts PII from user messages, scans tool results for
// secrets, and wraps the response iterator to scan model output.
func (m *PIIRedactingModelAdapter) GenerateContent(
	ctx context.Context,
	req *model.LLMRequest,
	stream bool,
) iter.Seq2[*model.LLMResponse, error] {
	// Redact PII from user messages (input side)
	for _, content := range req.Contents {
		if content.Role != "user" {
			continue
		}
		for _, part := range content.Parts {
			if part.Text == "" {
				continue
			}
			redacted := m.redactor.RedactInput(part.Text)
			if redacted != part.Text {
				piiModelLogger.Debugw("redacted PII from user message")
				part.Text = redacted
			}
		}
	}

	// Scan tool results for secret values (input side)
	if m.scanner != nil && m.scanner.HasSecrets() {
		for _, content := range req.Contents {
			if content.Role != "tool" {
				continue
			}
			for _, part := range content.Parts {
				if part.Text == "" {
					continue
				}
				scanned := m.scanner.Scan(part.Text)
				if scanned != part.Text {
					piiModelLogger.Debugw("redacted secrets from tool result")
					part.Text = scanned
				}
			}
		}
	}

	inner := m.inner.GenerateContent(ctx, req, stream)

	// If no scanner, return inner iterator directly
	if m.scanner == nil || !m.scanner.HasSecrets() {
		return inner
	}

	// Wrap iterator to scan model responses for secrets (output side)
	return func(yield func(*model.LLMResponse, error) bool) {
		for resp, err := range inner {
			if err == nil && resp != nil && resp.Content != nil {
				for _, part := range resp.Content.Parts {
					if part.Text == "" {
						continue
					}
					scanned := m.scanner.Scan(part.Text)
					if scanned != part.Text {
						piiModelLogger.Debugw("redacted secrets from model response")
						part.Text = scanned
					}
				}
			}
			if !yield(resp, err) {
				return
			}
		}
	}
}
