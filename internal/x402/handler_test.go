package x402

import (
	"net/http"
	"testing"
)

func TestParseChallenge(t *testing.T) {
	tests := []struct {
		give       string
		giveStatus int
		headers    map[string]string
		wantErr    bool
		wantAmount string
	}{
		{
			give:       "basic challenge",
			giveStatus: 402,
			headers: map[string]string{
				HeaderPaymentAmount:    "0.01",
				HeaderPaymentRecipient: "0x1234567890abcdef1234567890abcdef12345678",
				HeaderPaymentToken:     "0xUSDC",
				HeaderPaymentNetwork:   "base-sepolia",
				HeaderPaymentChainID:   "84532",
			},
			wantAmount: "0.01",
		},
		{
			give:       "non-402 status",
			giveStatus: 200,
			headers:    map[string]string{},
			wantErr:    true,
		},
		{
			give:       "missing amount header",
			giveStatus: 402,
			headers: map[string]string{
				HeaderPaymentRecipient: "0x1234",
			},
			wantErr: true,
		},
		{
			give:       "missing recipient header",
			giveStatus: 402,
			headers: map[string]string{
				HeaderPaymentAmount: "0.01",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.giveStatus,
				Header:     http.Header{},
			}
			for k, v := range tt.headers {
				resp.Header.Set(k, v)
			}

			challenge, err := ParseChallenge("https://example.com/api", resp)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if challenge.Amount != tt.wantAmount {
				t.Errorf("Amount = %q, want %q", challenge.Amount, tt.wantAmount)
			}
		})
	}
}

func TestBuildPaymentHeader(t *testing.T) {
	payload := PaymentPayload{
		TxHash:  "0xabc123",
		From:    "0xsender",
		ChainID: 84532,
	}

	header, err := BuildPaymentHeader(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if header == "" {
		t.Error("expected non-empty header")
	}
}
