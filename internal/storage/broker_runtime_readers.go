package storage

import (
	"context"
	"time"

	"github.com/langoai/lango/internal/p2p/reputation"
	"github.com/langoai/lango/internal/storagebroker"
	"go.uber.org/zap"
)

func WithBrokerRuntimeReaders(broker storagebroker.API) Option {
	return func(f *Facade) {
		if broker == nil || f == nil {
			return
		}
		f.learningHistory = func(ctx context.Context, limit int) ([]LearningHistoryRecord, error) {
			result, err := broker.LearningHistory(ctx, limit)
			if err != nil {
				return nil, err
			}
			out := make([]LearningHistoryRecord, 0, len(result.Entries))
			for _, row := range result.Entries {
				out = append(out, LearningHistoryRecord{
					ID:         row.ID,
					Trigger:    row.Trigger,
					Category:   row.Category,
					Diagnosis:  row.Diagnosis,
					Fix:        row.Fix,
					Confidence: row.Confidence,
					CreatedAt:  row.CreatedAt,
				})
			}
			return out, nil
		}
		f.pendingInquiries = func(ctx context.Context, limit int) ([]InquiryRecord, error) {
			result, err := broker.PendingInquiries(ctx, limit)
			if err != nil {
				return nil, err
			}
			out := make([]InquiryRecord, 0, len(result.Entries))
			for _, row := range result.Entries {
				out = append(out, InquiryRecord{
					ID:       row.ID,
					Topic:    row.Topic,
					Question: row.Question,
					Priority: row.Priority,
					Created:  row.Created,
				})
			}
			return out, nil
		}
		f.alerts = func(ctx context.Context, from time.Time) ([]AlertRecord, error) {
			result, err := broker.Alerts(ctx, from)
			if err != nil {
				return nil, err
			}
			out := make([]AlertRecord, 0, len(result.Alerts))
			for _, row := range result.Alerts {
				out = append(out, AlertRecord{
					ID:        row.ID,
					Type:      row.Type,
					Actor:     row.Actor,
					Details:   row.Details,
					Timestamp: row.Timestamp,
				})
			}
			return out, nil
		}
		f.paymentHistory = func(ctx context.Context, limit int) ([]PaymentHistoryRecord, error) {
			result, err := broker.PaymentHistory(ctx, limit)
			if err != nil {
				return nil, err
			}
			out := make([]PaymentHistoryRecord, 0, len(result.Entries))
			for _, row := range result.Entries {
				out = append(out, PaymentHistoryRecord{
					TxHash:        row.TxHash,
					Status:        row.Status,
					Amount:        row.Amount,
					From:          row.From,
					To:            row.To,
					ChainID:       row.ChainID,
					Purpose:       row.Purpose,
					X402URL:       row.X402URL,
					PaymentMethod: row.PaymentMethod,
					ErrorMessage:  row.ErrorMessage,
					CreatedAt:     row.CreatedAt,
				})
			}
			return out, nil
		}
		f.paymentUsage = func(ctx context.Context) (PaymentUsageSummary, error) {
			result, err := broker.PaymentUsage(ctx)
			if err != nil {
				return PaymentUsageSummary{}, err
			}
			return PaymentUsageSummary{DailySpent: result.DailySpent}, nil
		}
		f.reputationLookup = func(ctx context.Context, peerDID string) (*reputation.PeerDetails, error) {
			return (&brokerReputationStore{broker: broker}).GetDetails(ctx, peerDID)
		}
		f.workflowState = func(logger *zap.SugaredLogger) WorkflowRunReader {
			return &brokerWorkflowRunStore{broker: broker}
		}
	}
}
