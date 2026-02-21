// Package payment provides agent tools for blockchain payment operations.
package payment

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/langowarny/lango/internal/agent"
	"github.com/langowarny/lango/internal/payment"
	"github.com/langowarny/lango/internal/security"
	"github.com/langowarny/lango/internal/session"
	"github.com/langowarny/lango/internal/wallet"
	"github.com/langowarny/lango/internal/x402"
)

// BuildTools creates the payment agent tools.
func BuildTools(svc *payment.Service, limiter wallet.SpendingLimiter, secrets *security.SecretsStore, chainID int64, interceptor *x402.Interceptor) []*agent.Tool {
	tools := []*agent.Tool{
		buildSendTool(svc),
		buildBalanceTool(svc),
		buildHistoryTool(svc),
		buildLimitsTool(limiter),
		buildWalletInfoTool(svc),
	}
	if secrets != nil {
		tools = append(tools, buildCreateWalletTool(secrets, chainID))
	}
	if interceptor != nil && interceptor.IsEnabled() {
		tools = append(tools, buildX402FetchTool(interceptor, svc))
	}
	return tools
}

func buildSendTool(svc *payment.Service) *agent.Tool {
	return &agent.Tool{
		Name:        "payment_send",
		Description: "Send USDC payment on Base blockchain. Requires approval. Amount is in USDC (e.g. \"0.50\" for 50 cents).",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"to": map[string]interface{}{
					"type":        "string",
					"description": "Recipient wallet address (0x...)",
				},
				"amount": map[string]interface{}{
					"type":        "string",
					"description": "Amount in USDC (e.g. \"1.50\")",
				},
				"purpose": map[string]interface{}{
					"type":        "string",
					"description": "Human-readable purpose of the payment",
				},
			},
			"required": []string{"to", "amount", "purpose"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			to, _ := params["to"].(string)
			amount, _ := params["amount"].(string)
			purpose, _ := params["purpose"].(string)

			if to == "" || amount == "" || purpose == "" {
				return nil, fmt.Errorf("to, amount, and purpose are required")
			}

			sessionKey := session.SessionKeyFromContext(ctx)

			receipt, err := svc.Send(ctx, payment.PaymentRequest{
				To:         to,
				Amount:     amount,
				Purpose:    purpose,
				SessionKey: sessionKey,
			})
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"status":  "submitted",
				"txHash":  receipt.TxHash,
				"amount":  receipt.Amount,
				"from":    receipt.From,
				"to":      receipt.To,
				"chainId": receipt.ChainID,
				"network": wallet.NetworkName(receipt.ChainID),
			}, nil
		},
	}
}

func buildBalanceTool(svc *payment.Service) *agent.Tool {
	return &agent.Tool{
		Name:        "payment_balance",
		Description: "Check USDC balance of the agent wallet.",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			balance, err := svc.Balance(ctx)
			if err != nil {
				return nil, err
			}

			addr, _ := svc.WalletAddress(ctx)

			return map[string]interface{}{
				"balance":  balance,
				"currency": "USDC",
				"address":  addr,
				"chainId":  svc.ChainID(),
				"network":  wallet.NetworkName(svc.ChainID()),
			}, nil
		},
	}
}

func buildHistoryTool(svc *payment.Service) *agent.Tool {
	return &agent.Tool{
		Name:        "payment_history",
		Description: "View recent payment transaction history.",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of transactions to return (default: 20)",
				},
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			limit := 20
			if l, ok := params["limit"].(float64); ok && l > 0 {
				limit = int(l)
			}

			history, err := svc.History(ctx, limit)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"transactions": history,
				"count":        len(history),
			}, nil
		},
	}
}

func buildLimitsTool(limiter wallet.SpendingLimiter) *agent.Tool {
	return &agent.Tool{
		Name:        "payment_limits",
		Description: "View current spending limits and daily usage.",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			spent, err := limiter.DailySpent(ctx)
			if err != nil {
				return nil, fmt.Errorf("get daily spent: %w", err)
			}

			remaining, err := limiter.DailyRemaining(ctx)
			if err != nil {
				return nil, fmt.Errorf("get daily remaining: %w", err)
			}

			entLimiter, ok := limiter.(*wallet.EntSpendingLimiter)
			if !ok {
				return map[string]interface{}{
					"dailySpent":     wallet.FormatUSDC(spent),
					"dailyRemaining": wallet.FormatUSDC(remaining),
					"currency":       "USDC",
				}, nil
			}

			return map[string]interface{}{
				"maxPerTx":        wallet.FormatUSDC(entLimiter.MaxPerTx()),
				"maxDaily":        wallet.FormatUSDC(entLimiter.MaxDaily()),
				"dailySpent":      wallet.FormatUSDC(spent),
				"dailyRemaining":  wallet.FormatUSDC(remaining),
				"currency":        "USDC",
			}, nil
		},
	}
}

func buildWalletInfoTool(svc *payment.Service) *agent.Tool {
	return &agent.Tool{
		Name:        "payment_wallet_info",
		Description: "Show wallet address and network information.",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			addr, err := svc.WalletAddress(ctx)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"address": addr,
				"chainId": svc.ChainID(),
				"network": wallet.NetworkName(svc.ChainID()),
			}, nil
		},
	}
}

func buildCreateWalletTool(secrets *security.SecretsStore, chainID int64) *agent.Tool {
	return &agent.Tool{
		Name:        "payment_create_wallet",
		Description: "Create a new blockchain wallet. Generates a private key stored securely — only the public address is returned. Requires approval.",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			addr, err := wallet.CreateWallet(ctx, secrets)
			if err != nil {
				if errors.Is(err, wallet.ErrWalletExists) {
					return map[string]interface{}{
						"status":  "exists",
						"address": addr,
						"chainId": chainID,
						"network": wallet.NetworkName(chainID),
						"message": "Wallet already exists. Use payment_wallet_info to view details.",
					}, nil
				}
				return nil, err
			}

			return map[string]interface{}{
				"status":  "created",
				"address": addr,
				"chainId": chainID,
				"network": wallet.NetworkName(chainID),
			}, nil
		},
	}
}

// buildX402FetchTool creates the payment_x402_fetch tool for HTTP requests with automatic X402 payment.
func buildX402FetchTool(interceptor *x402.Interceptor, svc *payment.Service) *agent.Tool {
	return &agent.Tool{
		Name:        "payment_x402_fetch",
		Description: "Make an HTTP request with automatic X402 payment. If the server responds with HTTP 402, the agent wallet automatically signs an EIP-3009 authorization and retries. Requires approval.",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "The URL to request",
				},
				"method": map[string]interface{}{
					"type":        "string",
					"description": "HTTP method (default: GET)",
					"enum":        []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
				},
				"body": map[string]interface{}{
					"type":        "string",
					"description": "Request body (for POST/PUT/PATCH)",
				},
				"headers": map[string]interface{}{
					"type":        "object",
					"description": "Additional HTTP headers as key-value pairs",
				},
			},
			"required": []string{"url"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			url, _ := params["url"].(string)
			if url == "" {
				return nil, fmt.Errorf("url is required")
			}

			method, _ := params["method"].(string)
			if method == "" {
				method = "GET"
			}

			body, _ := params["body"].(string)

			httpClient, err := interceptor.HTTPClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("create X402 HTTP client: %w", err)
			}

			var bodyReader io.Reader
			if body != "" {
				bodyReader = strings.NewReader(body)
			}

			req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
			if err != nil {
				return nil, fmt.Errorf("create request: %w", err)
			}

			// Add custom headers.
			if hdrs, ok := params["headers"].(map[string]interface{}); ok {
				for k, v := range hdrs {
					if s, ok := v.(string); ok {
						req.Header.Set(k, s)
					}
				}
			}

			resp, err := httpClient.Do(req)
			if err != nil {
				return nil, fmt.Errorf("X402 request: %w", err)
			}
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("read response body: %w", err)
			}

			// Truncate large responses for agent context.
			bodyStr := string(respBody)
			const maxBodyLen = 8192
			truncated := false
			if len(bodyStr) > maxBodyLen {
				bodyStr = bodyStr[:maxBodyLen]
				truncated = true
			}

			respHeaders := make(map[string]string, len(resp.Header))
			for k, v := range resp.Header {
				if len(v) > 0 {
					respHeaders[k] = v[0]
				}
			}

			result := map[string]interface{}{
				"statusCode": resp.StatusCode,
				"body":       bodyStr,
				"headers":    respHeaders,
			}
			if truncated {
				result["truncated"] = true
			}

			// If payment was made (non-402 response after retry), record it for audit.
			if svc != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
				if paymentResp := resp.Header.Get("Payment-Response"); paymentResp != "" {
					addr, _ := interceptor.SignerAddress(ctx)
					_ = svc.RecordX402Payment(ctx, payment.X402PaymentRecord{
						URL:     url,
						From:    addr,
						ChainID: 0, // Set from config at wiring level if needed.
					})
				}
			}

			return result, nil
		},
	}
}
