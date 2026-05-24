package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/paymenttx"
)

// TxRecord describes a persisted payment transaction row.
type TxRecord struct {
	ID            uuid.UUID
	TxHash        string
	FromAddress   string
	ToAddress     string
	Amount        string
	ChainID       int64
	Status        paymenttx.Status
	SessionKey    string
	Purpose       string
	X402URL       string
	PaymentMethod paymenttx.PaymentMethod
	ErrorMessage  string
	CreatedAt     time.Time
}

// TxStore persists payment transaction lifecycle records.
type TxStore interface {
	Create(ctx context.Context, record TxRecord) (TxRecord, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status paymenttx.Status, txHash, errMsg string) error
	List(ctx context.Context, limit int) ([]TxRecord, error)
	DailySpendSince(ctx context.Context, since time.Time) ([]string, error)
}

// EntTxStore is the default Ent-backed payment transaction store.
type EntTxStore struct {
	client *ent.Client
}

func NewEntTxStore(client *ent.Client) *EntTxStore {
	if client == nil {
		return nil
	}
	return &EntTxStore{client: client}
}

func (s *EntTxStore) Create(ctx context.Context, record TxRecord) (TxRecord, error) {
	if s == nil || s.client == nil {
		return TxRecord{}, fmt.Errorf("payment tx store unavailable")
	}
	builder := s.client.PaymentTx.Create().
		SetID(record.ID).
		SetFromAddress(record.FromAddress).
		SetToAddress(record.ToAddress).
		SetAmount(record.Amount).
		SetChainID(record.ChainID).
		SetStatus(record.Status).
		SetPaymentMethod(record.PaymentMethod)
	if v := nilIfEmpty(record.SessionKey); v != nil {
		builder = builder.SetSessionKey(*v)
	}
	if v := nilIfEmpty(record.Purpose); v != nil {
		builder = builder.SetPurpose(*v)
	}
	if v := nilIfEmpty(record.X402URL); v != nil {
		builder = builder.SetX402URL(*v)
	}
	if v := nilIfEmpty(record.ErrorMessage); v != nil {
		builder = builder.SetErrorMessage(*v)
	}
	row, err := builder.Save(ctx)
	if err != nil {
		return TxRecord{}, err
	}
	return txRecordFromEnt(row), nil
}

func (s *EntTxStore) UpdateStatus(ctx context.Context, id uuid.UUID, status paymenttx.Status, txHash, errMsg string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("payment tx store unavailable")
	}
	update := s.client.PaymentTx.UpdateOneID(id).SetStatus(status)
	if txHash != "" {
		update = update.SetTxHash(txHash)
	}
	if errMsg != "" {
		update = update.SetErrorMessage(errMsg)
	}
	return update.Exec(ctx)
}

func (s *EntTxStore) List(ctx context.Context, limit int) ([]TxRecord, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("payment tx store unavailable")
	}
	if limit <= 0 {
		limit = DefaultHistoryLimit
	}
	rows, err := s.client.PaymentTx.Query().
		Order(ent.Desc(paymenttx.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]TxRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, txRecordFromEnt(row))
	}
	return out, nil
}

func (s *EntTxStore) DailySpendSince(ctx context.Context, since time.Time) ([]string, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("payment tx store unavailable")
	}
	rows, err := s.client.PaymentTx.Query().
		Where(
			paymenttx.CreatedAtGTE(since),
			paymenttx.StatusIn(paymenttx.StatusPending, paymenttx.StatusSubmitted, paymenttx.StatusConfirmed),
		).
		Select(paymenttx.FieldAmount).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Amount)
	}
	return out, nil
}

func txRecordFromEnt(row *ent.PaymentTx) TxRecord {
	if row == nil {
		return TxRecord{}
	}
	return TxRecord{
		ID:            row.ID,
		TxHash:        row.TxHash,
		FromAddress:   row.FromAddress,
		ToAddress:     row.ToAddress,
		Amount:        row.Amount,
		ChainID:       row.ChainID,
		Status:        row.Status,
		SessionKey:    row.SessionKey,
		Purpose:       row.Purpose,
		X402URL:       row.X402URL,
		PaymentMethod: row.PaymentMethod,
		ErrorMessage:  row.ErrorMessage,
		CreatedAt:     row.CreatedAt,
	}
}
