package postadjudicationreplay

import (
	"context"
	"errors"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/receipts"
)

var (
	ErrTransactionReceiptNotFound   = errors.New("transaction receipt not found")
	ErrCurrentSubmissionMissing     = errors.New("current submission missing")
	ErrDeadLetterEvidenceMissing    = errors.New("dead-letter evidence missing")
	ErrCanonicalAdjudicationMissing = errors.New("canonical adjudication missing")
	ErrActorUnresolved              = errors.New("actor_unresolved")
	ErrReplayNotAllowed             = errors.New("replay_not_allowed")
)

type Request struct {
	TransactionReceiptID string `json:"transaction_receipt_id"`
}

type Result struct {
	CanonicalAdjudication     CanonicalAdjudicationSnapshot `json:"canonical_adjudication"`
	BackgroundDispatchReceipt *BackgroundDispatchReceipt    `json:"background_dispatch_receipt,omitempty"`
}

type CanonicalAdjudicationSnapshot struct {
	TransactionReceipt receipts.TransactionReceipt `json:"transaction_receipt"`
	SubmissionReceipt  receipts.SubmissionReceipt  `json:"submission_receipt"`
	SubmissionEvents   []receipts.ReceiptEvent     `json:"submission_events,omitempty"`
}

type BackgroundDispatchRequest struct {
	TransactionReceiptID string                              `json:"transaction_receipt_id"`
	SubmissionReceiptID  string                              `json:"submission_receipt_id"`
	EscrowReference      string                              `json:"escrow_reference,omitempty"`
	Outcome              receipts.EscrowAdjudicationDecision `json:"outcome,omitempty"`
	Prompt               string                              `json:"prompt"`
}

type BackgroundDispatchReceipt struct {
	Status               string `json:"status"`
	TransactionReceiptID string `json:"transaction_receipt_id"`
	SubmissionReceiptID  string `json:"submission_receipt_id,omitempty"`
	EscrowReference      string `json:"escrow_reference,omitempty"`
	Outcome              string `json:"outcome,omitempty"`
	DispatchReference    string `json:"dispatch_reference,omitempty"`
}

type receiptStore interface {
	GetTransactionReceipt(context.Context, string) (receipts.TransactionReceipt, error)
	GetSubmissionReceipt(context.Context, string) (receipts.SubmissionReceipt, []receipts.ReceiptEvent, error)
	RecordManualRetryRequested(context.Context, receipts.ManualRetryRequestedRequest) error
}

type dispatcher interface {
	Dispatch(context.Context, BackgroundDispatchRequest) (BackgroundDispatchReceipt, error)
}

type ReplayPolicy struct {
	AllowedActors        []string
	ReleaseAllowedActors []string
	RefundAllowedActors  []string
}

func ReplayPolicyFromConfig(cfg *config.Config) ReplayPolicy {
	if cfg == nil {
		return ReplayPolicy{}
	}
	return ReplayPolicy{
		AllowedActors:        append([]string(nil), cfg.Replay.AllowedActors...),
		ReleaseAllowedActors: append([]string(nil), cfg.Replay.ReleaseAllowedActors...),
		RefundAllowedActors:  append([]string(nil), cfg.Replay.RefundAllowedActors...),
	}
}
