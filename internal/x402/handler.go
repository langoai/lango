package x402

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// ParseChallenge extracts an X402 payment challenge from HTTP 402 response headers.
func ParseChallenge(url string, resp *http.Response) (*Challenge, error) {
	if resp.StatusCode != http.StatusPaymentRequired {
		return nil, fmt.Errorf("expected HTTP 402, got %d", resp.StatusCode)
	}

	amount := resp.Header.Get(HeaderPaymentAmount)
	if amount == "" {
		return nil, fmt.Errorf("missing %s header", HeaderPaymentAmount)
	}

	recipient := resp.Header.Get(HeaderPaymentRecipient)
	if recipient == "" {
		return nil, fmt.Errorf("missing %s header", HeaderPaymentRecipient)
	}

	token := resp.Header.Get(HeaderPaymentToken)
	network := resp.Header.Get(HeaderPaymentNetwork)

	var chainID int64
	if cidStr := resp.Header.Get(HeaderPaymentChainID); cidStr != "" {
		var err error
		chainID, err = strconv.ParseInt(cidStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid %s header: %w", HeaderPaymentChainID, err)
		}
	}

	return &Challenge{
		PaymentURL:       url,
		Amount:           amount,
		TokenAddress:     token,
		RecipientAddress: recipient,
		Network:          network,
		ChainID:          chainID,
	}, nil
}

// BuildPaymentHeader creates the X-PAYMENT header value from a payment payload.
func BuildPaymentHeader(payload PaymentPayload) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payment payload: %w", err)
	}
	return string(data), nil
}
