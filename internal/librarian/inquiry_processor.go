package librarian

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
)

const answerDetectionPrompt = `You are a knowledge librarian. Analyze the recent conversation messages to detect if the user has answered any of the pending knowledge inquiries.

For each inquiry, determine if a recent message provides an answer (direct or indirect).

Output JSON array of matches:
[
  {
    "inquiry_id": "uuid of the matched inquiry",
    "answer": "the extracted answer from conversation",
    "confidence": "high|medium|low",
    "knowledge": {
      "key": "unique_snake_case_key",
      "category": "preference|fact|rule|definition",
      "content": "structured knowledge to save"
    }
  }
]

Rules:
- Only match if the user clearly answered the question (high/medium confidence)
- Extract the answer as clean, reusable knowledge
- If no matches found, return an empty array []
- Do NOT force matches — empty is better than wrong`

// InquiryProcessor detects user answers to pending inquiries and saves knowledge.
type InquiryProcessor struct {
	generator      TextGenerator
	inquiryStore   *InquiryStore
	knowledgeStore *knowledge.Store
	logger         *zap.SugaredLogger
}

// NewInquiryProcessor creates a new inquiry processor.
func NewInquiryProcessor(
	generator TextGenerator,
	inquiryStore *InquiryStore,
	knowledgeStore *knowledge.Store,
	logger *zap.SugaredLogger,
) *InquiryProcessor {
	return &InquiryProcessor{
		generator:      generator,
		inquiryStore:   inquiryStore,
		knowledgeStore: knowledgeStore,
		logger:         logger,
	}
}

// ProcessAnswers checks recent messages against pending inquiries for answer matches.
func (p *InquiryProcessor) ProcessAnswers(ctx context.Context, sessionKey string, messages []session.Message) error {
	pending, err := p.inquiryStore.ListPendingInquiries(ctx, sessionKey, 10)
	if err != nil {
		return fmt.Errorf("list pending inquiries: %w", err)
	}
	if len(pending) == 0 {
		return nil
	}

	// Take only the last few messages for matching.
	recentMessages := messages
	if len(recentMessages) > 6 {
		recentMessages = recentMessages[len(recentMessages)-6:]
	}

	prompt := p.buildMatchPrompt(pending, recentMessages)
	raw, err := p.generator.GenerateText(ctx, answerDetectionPrompt, prompt)
	if err != nil {
		return fmt.Errorf("detect answers: %w", err)
	}

	matches, err := parseAnswerMatches(raw)
	if err != nil {
		p.logger.Debugw("parse answer matches", "error", err, "raw", raw)
		return nil
	}

	for _, match := range matches {
		if match.Confidence == types.ConfidenceLow {
			continue
		}

		inquiryID, err := uuid.Parse(match.InquiryID)
		if err != nil {
			p.logger.Warnw("parse inquiry ID", "id", match.InquiryID, "error", err)
			continue
		}

		var knowledgeKey string
		if match.Knowledge != nil && match.Knowledge.Key != "" {
			cat, err := mapCategory(match.Knowledge.Category)
			if err != nil {
				p.logger.Warnw("skip knowledge: invalid category",
					"key", match.Knowledge.Key, "category", match.Knowledge.Category, "error", err)
			} else {
				entry := knowledge.KnowledgeEntry{
					Key:      match.Knowledge.Key,
					Category: cat,
					Content:  match.Knowledge.Content,
					Source:   "proactive_librarian",
				}
				if err := p.knowledgeStore.SaveKnowledge(ctx, sessionKey, entry); err != nil {
					p.logger.Warnw("save matched knowledge", "key", entry.Key, "error", err)
				} else {
					knowledgeKey = entry.Key
					p.logger.Infow("knowledge saved from inquiry answer",
						"key", entry.Key, "inquiryID", match.InquiryID)
				}
			}
		}

		if err := p.inquiryStore.ResolveInquiry(ctx, inquiryID, match.Answer, knowledgeKey); err != nil {
			p.logger.Warnw("resolve inquiry", "id", match.InquiryID, "error", err)
		}
	}

	return nil
}

func (p *InquiryProcessor) buildMatchPrompt(pending []Inquiry, messages []session.Message) string {
	var b strings.Builder

	b.WriteString("## Pending Inquiries\n")
	for _, inq := range pending {
		b.WriteString(fmt.Sprintf("- ID: %s | Topic: %s | Question: %s\n", inq.ID.String(), inq.Topic, inq.Question))
	}

	b.WriteString("\n## Recent Messages\n")
	for _, msg := range messages {
		b.WriteString(fmt.Sprintf("[%s]: %s\n", msg.Role, msg.Content))
	}

	return b.String()
}
