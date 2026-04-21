// Package payment provides agent tools for blockchain payment operations.
package payment

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/paymentgate"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolparam"
	"github.com/langoai/lango/internal/wallet"
	"github.com/langoai/lango/internal/x402"
)

// ServiceAPI captures the payment service methods used by the tool surface.
type ServiceAPI interface {
	Send(context.Context, payment.PaymentRequest) (*payment.PaymentReceipt, error)
	Balance(context.Context) (string, error)
	History(context.Context, int) ([]payment.TransactionInfo, error)
	WalletAddress(context.Context) (string, error)
	ChainID() int64
	RecordX402Payment(context.Context, payment.X402PaymentRecord) error
}

// PaymentExecutionGate evaluates whether a direct payment may proceed.
type PaymentExecutionGate interface {
	EvaluateDirectPayment(context.Context, paymentgate.Request) (paymentgate.Result, error)
}

// PaymentExecutionAuditEntry records a payment execution decision for audit logging.
type PaymentExecutionAuditEntry struct {
	ToolName             string
	SessionKey           string
	TransactionReceiptID string
	Outcome              string
	Reason               string
}

// PaymentExecutionAuditor records payment execution decisions to audit logs.
type PaymentExecutionAuditor interface {
	RecordPaymentExecution(context.Context, PaymentExecutionAuditEntry) error
}

// PaymentExecutionTrail records payment execution decisions in receipt trails.
type PaymentExecutionTrail interface {
	AppendPaymentExecutionAuthorized(context.Context, string) error
	AppendPaymentExecutionDenied(context.Context, string, string) error
}

type PaymentExecutionDeniedResult struct {
	Status               string `json:"status"`
	ToolName             string `json:"tool_name"`
	TransactionReceiptID string `json:"transaction_receipt_id,omitempty"`
	Reason               string `json:"reason"`
	Message              string `json:"message"`
}

// BuildTools creates the payment agent tools.
func BuildTools(svc ServiceAPI, limiter wallet.SpendingLimiter, secrets *security.SecretsStore, chainID int64, interceptor *x402.Interceptor, receiptStore *receipts.Store, auditor PaymentExecutionAuditor) []*agent.Tool {
	var gate PaymentExecutionGate
	var trail PaymentExecutionTrail
	if receiptStore != nil {
		gate = paymentgate.NewService(receiptStore)
		trail = receiptStore
	} else {
		gate = DenyAllPaymentExecutionGate{}
	}

	tools := []*agent.Tool{
		buildSendTool(svc, gate, trail, auditor),
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

// DenyAllPaymentExecutionGate is a fail-closed gate for unavailable receipt-backed execution.
type DenyAllPaymentExecutionGate struct{}

func (DenyAllPaymentExecutionGate) EvaluateDirectPayment(context.Context, paymentgate.Request) (paymentgate.Result, error) {
	return paymentgate.Result{Decision: paymentgate.Deny, Reason: paymentgate.ReasonMissingReceipt}, nil
}

func buildSendTool(svc ServiceAPI, gate PaymentExecutionGate, trail PaymentExecutionTrail, auditor PaymentExecutionAuditor) *agent.Tool {
	return &agent.Tool{
		Name:        "payment_send",
		Description: "Send USDC payment on Base blockchain. Requires approval. Amount is in USDC (e.g. \"0.50\" for 50 cents).",
		SafetyLevel: agent.SafetyLevelDangerous,
		Capability: agent.ToolCapability{
			Category:             "payment",
			Activity:             agent.ActivityExecute,
			RequiredCapabilities: []string{"payment"},
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"to": map[string]interface{}{
					"type":        "string",
					"description": "Recipient wallet address (0x...)",
				},
				"transaction_receipt_id": map[string]interface{}{
					"type":        "string",
					"description": "Linked transaction receipt identifier that must be approved for direct payment execution",
				},
				"submission_receipt_id": map[string]interface{}{
					"type":        "string",
					"description": "Explicit submission receipt identifier that should receive the execution evidence",
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
			"required": []string{"to", "transaction_receipt_id", "submission_receipt_id", "amount", "purpose"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			to := toolparam.OptionalString(params, "to", "")
			transactionReceiptID := toolparam.OptionalString(params, "transaction_receipt_id", "")
			submissionReceiptID := toolparam.OptionalString(params, "submission_receipt_id", "")
			amount := toolparam.OptionalString(params, "amount", "")
			purpose := toolparam.OptionalString(params, "purpose", "")

			if to == "" || amount == "" || purpose == "" {
				return nil, fmt.Errorf("to, amount, and purpose are required")
			}

			allowed, denied, err := CheckDirectPaymentExecution(ctx, "payment_send", transactionReceiptID, submissionReceiptID, gate, trail, auditor)
			if err != nil {
				return nil, err
			}
			if !allowed {
				return denied, nil
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

			result := map[string]interface{}{
				"status":  receipt.Status,
				"txHash":  receipt.TxHash,
				"amount":  receipt.Amount,
				"from":    receipt.From,
				"to":      receipt.To,
				"chainId": receipt.ChainID,
				"network": wallet.NetworkName(receipt.ChainID),
			}
			if receipt.GasUsed > 0 {
				result["gasUsed"] = receipt.GasUsed
			}
			if receipt.BlockNumber > 0 {
				result["blockNumber"] = receipt.BlockNumber
			}
			return result, nil
		},
	}
}

func buildBalanceTool(svc ServiceAPI) *agent.Tool {
	return &agent.Tool{
		Name:        "payment_balance",
		Description: "Check USDC balance of the agent wallet.",
		SafetyLevel: agent.SafetyLevelSafe,
		Capability: agent.ToolCapability{
			Category:             "payment",
			Activity:             agent.ActivityQuery,
			ReadOnly:             true,
			ConcurrencySafe:      true,
			RequiredCapabilities: []string{"payment"},
		},
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
				"currency": wallet.CurrencyUSDC,
				"address":  addr,
				"chainId":  svc.ChainID(),
				"network":  wallet.NetworkName(svc.ChainID()),
			}, nil
		},
	}
}

func buildHistoryTool(svc ServiceAPI) *agent.Tool {
	return &agent.Tool{
		Name:        "payment_history",
		Description: "View recent payment transaction history.",
		SafetyLevel: agent.SafetyLevelSafe,
		Capability: agent.ToolCapability{
			Category:             "payment",
			Activity:             agent.ActivityQuery,
			ReadOnly:             true,
			ConcurrencySafe:      true,
			RequiredCapabilities: []string{"payment"},
		},
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
			limit := toolparam.OptionalInt(params, "limit", 20)

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
		Capability: agent.ToolCapability{
			Category:             "payment",
			Activity:             agent.ActivityQuery,
			ReadOnly:             true,
			ConcurrencySafe:      true,
			RequiredCapabilities: []string{"payment"},
		},
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
					"currency":       wallet.CurrencyUSDC,
				}, nil
			}

			return map[string]interface{}{
				"maxPerTx":       wallet.FormatUSDC(entLimiter.MaxPerTx()),
				"maxDaily":       wallet.FormatUSDC(entLimiter.MaxDaily()),
				"dailySpent":     wallet.FormatUSDC(spent),
				"dailyRemaining": wallet.FormatUSDC(remaining),
				"currency":       wallet.CurrencyUSDC,
			}, nil
		},
	}
}

func buildWalletInfoTool(svc ServiceAPI) *agent.Tool {
	return &agent.Tool{
		Name:        "payment_wallet_info",
		Description: "Show wallet address and network information.",
		SafetyLevel: agent.SafetyLevelSafe,
		Capability: agent.ToolCapability{
			Category:             "payment",
			Activity:             agent.ActivityQuery,
			ReadOnly:             true,
			ConcurrencySafe:      true,
			RequiredCapabilities: []string{"payment"},
		},
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
		Capability: agent.ToolCapability{
			Category:             "payment",
			Activity:             agent.ActivityExecute,
			RequiredCapabilities: []string{"payment"},
		},
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
func buildX402FetchTool(interceptor *x402.Interceptor, svc ServiceAPI) *agent.Tool {
	return &agent.Tool{
		Name:        "payment_x402_fetch",
		Description: "Make an HTTP request with automatic X402 payment. If the server responds with HTTP 402, the agent wallet automatically signs an EIP-3009 authorization and retries. Requires approval.",
		SafetyLevel: agent.SafetyLevelDangerous,
		Capability: agent.ToolCapability{
			Category:             "payment",
			Activity:             agent.ActivityExecute,
			RequiredCapabilities: []string{"payment"},
		},
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
			url := toolparam.OptionalString(params, "url", "")
			if url == "" {
				return nil, fmt.Errorf("url is required")
			}

			method := toolparam.OptionalString(params, "method", "")
			if method == "" {
				method = "GET"
			}

			body := toolparam.OptionalString(params, "body", "")

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

func CheckDirectPaymentExecution(ctx context.Context, toolName, transactionReceiptID, submissionReceiptID string, gate PaymentExecutionGate, trail PaymentExecutionTrail, auditor PaymentExecutionAuditor) (bool, *PaymentExecutionDeniedResult, error) {
	result, err := gate.EvaluateDirectPayment(ctx, paymentgate.Request{
		TransactionReceiptID: transactionReceiptID,
		SubmissionReceiptID:  submissionReceiptID,
		ToolName:             toolName,
		Context: map[string]interface{}{
			"session_key": session.SessionKeyFromContext(ctx),
		},
	})
	if err != nil {
		return false, nil, fmt.Errorf("evaluate payment execution gate: %w", err)
	}

	entry := PaymentExecutionAuditEntry{
		ToolName:             toolName,
		SessionKey:           session.SessionKeyFromContext(ctx),
		TransactionReceiptID: transactionReceiptID,
	}

	if result.Decision == paymentgate.Deny {
		entry.Outcome = "denied"
		entry.Reason = string(result.Reason)
		if auditor != nil {
			if err := auditor.RecordPaymentExecution(ctx, entry); err != nil {
				return false, nil, fmt.Errorf("record payment execution audit: %w", err)
			}
		}
		if trail != nil {
			if strings.TrimSpace(submissionReceiptID) == "" {
				return false, nil, fmt.Errorf("record payment execution receipt trail: submission_receipt_id is required")
			}
			if err := trail.AppendPaymentExecutionDenied(ctx, submissionReceiptID, string(result.Reason)); err != nil {
				return false, nil, fmt.Errorf("record payment execution receipt trail: %w", err)
			}
		}
		return false, &PaymentExecutionDeniedResult{
			Status:               "denied",
			ToolName:             toolName,
			TransactionReceiptID: transactionReceiptID,
			Reason:               string(result.Reason),
			Message:              PaymentExecutionDeniedMessage(result.Reason, transactionReceiptID, submissionReceiptID),
		}, nil
	}

	entry.Outcome = "authorized"
	if auditor != nil {
		if err := auditor.RecordPaymentExecution(ctx, entry); err != nil {
			return false, nil, fmt.Errorf("record payment execution audit: %w", err)
		}
	}
	if trail != nil {
		if strings.TrimSpace(submissionReceiptID) == "" {
			return false, nil, fmt.Errorf("record payment execution receipt trail: submission_receipt_id is required")
		}
		if err := trail.AppendPaymentExecutionAuthorized(ctx, submissionReceiptID); err != nil {
			return false, nil, fmt.Errorf("record payment execution receipt trail: %w", err)
		}
	}
	return true, nil, nil
}

func PaymentExecutionDeniedMessage(reason paymentgate.DenyReason, transactionReceiptID, submissionReceiptID string) string {
	switch reason {
	case paymentgate.ReasonMissingReceipt:
		if strings.TrimSpace(transactionReceiptID) == "" {
			return "direct payment execution denied: transaction_receipt_id is required"
		}
		if strings.TrimSpace(submissionReceiptID) == "" {
			return "direct payment execution denied: submission_receipt_id is required"
		}
		return "direct payment execution denied: transaction_receipt_id was not found"
	case paymentgate.ReasonApprovalNotApproved:
		return "direct payment execution denied: canonical payment approval is not approved"
	case paymentgate.ReasonExecutionModeMismatch:
		return "direct payment execution denied: canonical settlement hint must be prepay"
	case paymentgate.ReasonStaleState:
		return "direct payment execution denied: canonical payment state is stale"
	default:
		return "direct payment execution denied"
	}
}
