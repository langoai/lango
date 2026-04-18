package librarian

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/testutil"
)

type stubPayloadProtector struct{}

func (stubPayloadProtector) EncryptPayload(plaintext []byte) ([]byte, []byte, int, error) {
	return append([]byte("enc:"), plaintext...), []byte("123456789012"), security.PayloadKeyVersionV1, nil
}

func (stubPayloadProtector) DecryptPayload(ciphertext []byte, nonce []byte, keyVersion int) ([]byte, error) {
	if keyVersion != security.PayloadKeyVersionV1 {
		return nil, errors.New("unexpected key version")
	}
	if string(nonce) != "123456789012" {
		return nil, errors.New("unexpected nonce")
	}
	return []byte(strings.TrimPrefix(string(ciphertext), "enc:")), nil
}

func TestInquiryPayloadProtection_PreservesQuestionContextAndAnswer(t *testing.T) {
	client := testutil.TestEntClient(t)
	store := NewInquiryStore(client, testutil.NopLogger())
	store.SetPayloadProtector(stubPayloadProtector{})
	ctx := context.Background()

	id := uuid.New()
	require.NoError(t, store.SaveInquiry(ctx, Inquiry{
		ID:         id,
		SessionKey: "sess-1",
		Topic:      "runtime",
		Question:   "email alice@example.com?",
		Context:    "token SECRETSECRETSECRETSECRETSECRETSECRET",
		Priority:   "high",
	}))
	require.NoError(t, store.ResolveInquiry(ctx, id, "number 123456789", "knowledge-key"))

	row := client.Inquiry.GetX(ctx, id)
	require.NotNil(t, row.PayloadCiphertext)
	require.NotContains(t, row.Question, "alice@example.com")
	require.NotContains(t, *row.Context, "SECRETSECRETSECRETSECRETSECRETSECRET")
	require.NotContains(t, *row.Answer, "123456789")

	decoded, err := store.entToInquiry(row)
	require.NoError(t, err)
	require.Equal(t, "email alice@example.com?", decoded.Question)
	require.Equal(t, "token SECRETSECRETSECRETSECRETSECRETSECRET", decoded.Context)
	require.Equal(t, "number 123456789", decoded.Answer)
}
