package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	paymentcore "github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/storage"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dummyBootLoader returns a boot loader that always errors.
func dummyBootLoader() func() (*bootstrap.Result, error) {
	return func() (*bootstrap.Result, error) {
		return nil, assert.AnError
	}
}

func executePaymentCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func executePaymentCommandWithInput(t *testing.T, cmd *cobra.Command, input string, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(bytes.NewBufferString(input))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

type stubWalletProvider struct {
	address string
}

func (s stubWalletProvider) Address(_ context.Context) (string, error) { return s.address, nil }
func (s stubWalletProvider) Balance(_ context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}
func (s stubWalletProvider) SignTransaction(_ context.Context, _ []byte) ([]byte, error) {
	return nil, nil
}
func (s stubWalletProvider) SignMessage(_ context.Context, _ []byte) ([]byte, error) {
	return nil, nil
}
func (s stubWalletProvider) PublicKey(_ context.Context) ([]byte, error) { return nil, nil }

func TestNewPaymentCmd_Structure(t *testing.T) {
	cmd := NewPaymentCmd(dummyBootLoader())
	require.NotNil(t, cmd)

	assert.Equal(t, "payment", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}

func TestNewPaymentCmd_Subcommands(t *testing.T) {
	cmd := NewPaymentCmd(dummyBootLoader())

	expected := []string{"balance", "history", "limits", "info", "send", "x402"}

	subCmds := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subCmds[sub.Use] = true
	}

	for _, name := range expected {
		assert.True(t, subCmds[name], "missing subcommand: %s", name)
	}
}

func TestNewPaymentCmd_SubcommandCount(t *testing.T) {
	cmd := NewPaymentCmd(dummyBootLoader())
	assert.Equal(t, 6, len(cmd.Commands()), "expected 6 payment subcommands")
}

func TestBalanceCmd_HasOutputFlag(t *testing.T) {
	cmd := NewPaymentCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		if sub.Use == "balance" {
			outputFlag := sub.Flags().Lookup("output")
			assert.NotNil(t, outputFlag, "balance command should have --output flag")
			return
		}
	}
	t.Fatal("balance subcommand not found")
}

func TestSendCmd_HasRequiredFlags(t *testing.T) {
	cmd := NewPaymentCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		if sub.Use == "send" {
			assert.NotNil(t, sub.Flags().Lookup("to"), "send should have --to flag")
			assert.NotNil(t, sub.Flags().Lookup("amount"), "send should have --amount flag")
			assert.NotNil(t, sub.Flags().Lookup("purpose"), "send should have --purpose flag")
			assert.NotNil(t, sub.Flags().Lookup("force"), "send should have --force flag")
			return
		}
	}
	t.Fatal("send subcommand not found")
}

func TestHistoryCmd_HasLimitFlag(t *testing.T) {
	cmd := NewPaymentCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		if sub.Use == "history" {
			limitFlag := sub.Flags().Lookup("limit")
			assert.NotNil(t, limitFlag, "history command should have --limit flag")
			return
		}
	}
	t.Fatal("history subcommand not found")
}

func TestSubcommands_HaveShortDescription(t *testing.T) {
	cmd := NewPaymentCmd(dummyBootLoader())
	for _, sub := range cmd.Commands() {
		assert.NotEmpty(t, sub.Short, "subcommand %q should have a Short description", sub.Use)
	}
}

func TestX402Cmd_WritesTableOutputToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = true
	cfg.Payment.X402.AutoIntercept = true
	cfg.Payment.X402.MaxAutoPayAmount = "1.25"
	cfg.Payment.Limits.MaxPerTx = "2.00"
	cfg.Payment.Limits.MaxDaily = "10.00"
	cmd := NewPaymentCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executePaymentCommand(t, cmd, "x402")

	require.NoError(t, err)
	assert.Contains(t, out, "X402 Protocol Configuration")
	assert.Contains(t, out, "Auto-Intercept:      enabled")
	assert.Contains(t, out, "1.25 USDC")
}

func TestX402Cmd_WritesJSONOutputToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = true
	cfg.Payment.X402.MaxAutoPayAmount = "3.50"
	cmd := NewPaymentCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executePaymentCommand(t, cmd, "x402", "--output", "json")

	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.Equal(t, true, payload["payment_enabled"])
	assert.Equal(t, "3.50", payload["max_auto_pay_usdc"])
}

func TestInfoCmd_WritesTableAndJSONToCommandWriter(t *testing.T) {
	orig := paymentDepsLoader
	t.Cleanup(func() { paymentDepsLoader = orig })

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = true
	cfg.Payment.WalletProvider = "local"
	cfg.Payment.Network.USDCContract = "0xabc"
	cfg.Payment.Network.RPCURL = "https://rpc.example"
	cfg.Payment.X402.AutoIntercept = true
	cfg.Payment.X402.MaxAutoPayAmount = "4.20"

	paymentDepsLoader = func(boot *bootstrap.Result) (*paymentDeps, error) {
		return &paymentDeps{
			service: paymentcore.NewService(stubWalletProvider{address: "0x1234"}, nil, nil, nil, nil, 84532),
			config:  &cfg.Payment,
			cleanup: func() {},
		}, nil
	}

	cmd := NewPaymentCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})

	tableOut, err := executePaymentCommand(t, cmd, "info")
	require.NoError(t, err)
	assert.Contains(t, tableOut, "Payment System Info")
	assert.Contains(t, tableOut, "0x1234")

	jsonOut, err := executePaymentCommand(t, cmd, "info", "--output", "json")
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(jsonOut), &payload))
	assert.Equal(t, "0x1234", payload["address"])
	assert.EqualValues(t, 84532, payload["chainId"])
}

func TestHistoryCmd_WritesTableAndJSONToCommandWriter(t *testing.T) {
	orig := paymentHistoryLoader
	t.Cleanup(func() { paymentHistoryLoader = orig })

	paymentHistoryLoader = func(ctx context.Context, boot *bootstrap.Result, limit int) ([]storage.PaymentHistoryRecord, error) {
		return []storage.PaymentHistoryRecord{{
			TxHash:        "0xaabbccddeeff",
			Status:        "confirmed",
			Amount:        "1.50",
			To:            "0x5678abcd",
			Purpose:       "API access fee",
			PaymentMethod: "direct",
			CreatedAt:     time.Date(2026, 5, 14, 9, 30, 0, 0, time.UTC),
		}}, nil
	}

	cfg := config.DefaultConfig()
	cmd := NewPaymentCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})

	tableOut, err := executePaymentCommand(t, cmd, "history")
	require.NoError(t, err)
	assert.Contains(t, tableOut, "STATUS")
	assert.Contains(t, tableOut, "API access fee")

	jsonOut, err := executePaymentCommand(t, cmd, "history", "--output", "json")
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(jsonOut), &payload))
	assert.EqualValues(t, 1, payload["count"])
}

func TestLimitsCmd_WritesTableAndJSONToCommandWriter(t *testing.T) {
	orig := paymentUsageLoader
	t.Cleanup(func() { paymentUsageLoader = orig })

	paymentUsageLoader = func(ctx context.Context, boot *bootstrap.Result) (storage.PaymentUsageSummary, error) {
		return storage.PaymentUsageSummary{DailySpent: "3.50"}, nil
	}

	cfg := config.DefaultConfig()
	cfg.Payment.Limits.MaxPerTx = "1.00"
	cfg.Payment.Limits.MaxDaily = "10.00"
	cmd := NewPaymentCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})

	tableOut, err := executePaymentCommand(t, cmd, "limits")
	require.NoError(t, err)
	assert.Contains(t, tableOut, "Spending Limits")
	assert.Contains(t, tableOut, "Remaining Today:      6.50 USDC")

	jsonOut, err := executePaymentCommand(t, cmd, "limits", "--output", "json")
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(jsonOut), &payload))
	assert.Equal(t, "3.50", payload["dailySpent"])
	assert.Equal(t, "6.50", payload["dailyRemaining"])
}

func TestSendCmd_WritesPromptAndSuccessOutputToCommandWriter(t *testing.T) {
	origDeps := paymentDepsLoader
	origSend := paymentSendExecutor
	t.Cleanup(func() {
		paymentDepsLoader = origDeps
		paymentSendExecutor = origSend
	})

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = true
	cfg.Payment.X402.AutoIntercept = false
	paymentDepsLoader = func(boot *bootstrap.Result) (*paymentDeps, error) {
		return &paymentDeps{
			service: paymentcore.NewService(stubWalletProvider{address: "0xfrom"}, nil, nil, nil, nil, 84532),
			config:  &cfg.Payment,
			cleanup: func() {},
		}, nil
	}
	paymentSendExecutor = func(ctx context.Context, deps *paymentDeps, req paymentcore.PaymentRequest) (*paymentcore.PaymentReceipt, error) {
		return &paymentcore.PaymentReceipt{
			Status:  "pending",
			TxHash:  "0xaabb",
			Amount:  req.Amount,
			From:    "0xfrom",
			To:      req.To,
			ChainID: 84532,
		}, nil
	}

	cmd := NewPaymentCmd(func() (*bootstrap.Result, error) { return &bootstrap.Result{Config: cfg}, nil })

	out, err := executePaymentCommandWithInput(t, cmd, "n\n", "send", "--to", "0xto", "--amount", "1.50", "--purpose", "API access")
	require.NoError(t, err)
	assert.Contains(t, out, "Send 1.50 USDC to 0xto on Base Sepolia?")
	assert.Contains(t, out, "Confirm [y/N]:")
	assert.Contains(t, out, "Aborted.")

	out, err = executePaymentCommand(t, cmd, "send", "--to", "0xto", "--amount", "1.50", "--purpose", "API access", "--force")
	require.NoError(t, err)
	assert.Contains(t, out, "Payment Submitted")
	assert.Contains(t, out, "0xaabb")

	out, err = executePaymentCommand(t, cmd, "send", "--to", "0xto", "--amount", "1.50", "--purpose", "API access", "--force", "--output", "json")
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.Equal(t, "0xaabb", payload["txHash"])
}

func TestPaymentCommands_InvalidOutputFailFast(t *testing.T) {
	cmd := NewPaymentCmd(func() (*bootstrap.Result, error) {
		t.Fatal("boot loader should not be called for invalid output")
		return nil, nil
	})

	cases := [][]string{
		{"balance", "--output", "yaml"},
		{"history", "--output", "yaml"},
		{"limits", "--output", "yaml"},
		{"info", "--output", "yaml"},
		{"x402", "--output", "yaml"},
		{"send", "--to", "0xto", "--amount", "1.50", "--purpose", "API access", "--output", "yaml"},
	}

	for _, args := range cases {
		_, err := executePaymentCommand(t, cmd, args...)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `unknown output format "yaml"`)
	}
}

func TestSendCmd_NonInteractiveRequiresForce(t *testing.T) {
	origDeps := paymentDepsLoader
	t.Cleanup(func() {
		paymentDepsLoader = origDeps
	})

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = true
	cfg.Payment.X402.AutoIntercept = false
	paymentDepsLoader = func(boot *bootstrap.Result) (*paymentDeps, error) {
		return &paymentDeps{
			service: paymentcore.NewService(stubWalletProvider{address: "0xfrom"}, nil, nil, nil, nil, 84532),
			config:  &cfg.Payment,
			cleanup: func() {},
		}, nil
	}
	cmd := NewPaymentCmd(func() (*bootstrap.Result, error) { return &bootstrap.Result{Config: cfg}, nil })
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { _ = reader.Close() })
	t.Cleanup(func() { _ = writer.Close() })
	cmd.SetIn(reader)
	cmd.SetArgs([]string{"send", "--to", "0xto", "--amount", "1.50", "--purpose", "API access"})

	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "use --force for non-interactive mode")
	assert.Empty(t, out.String())
}

func TestSendCmd_EOFAbortsWithoutSubmission(t *testing.T) {
	origDeps := paymentDepsLoader
	origSend := paymentSendExecutor
	t.Cleanup(func() {
		paymentDepsLoader = origDeps
		paymentSendExecutor = origSend
	})

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = true
	cfg.Payment.X402.AutoIntercept = false
	paymentDepsLoader = func(boot *bootstrap.Result) (*paymentDeps, error) {
		return &paymentDeps{
			service: paymentcore.NewService(stubWalletProvider{address: "0xfrom"}, nil, nil, nil, nil, 84532),
			config:  &cfg.Payment,
			cleanup: func() {},
		}, nil
	}
	paymentSendExecutor = func(ctx context.Context, deps *paymentDeps, req paymentcore.PaymentRequest) (*paymentcore.PaymentReceipt, error) {
		t.Fatal("paymentSendExecutor should not be called on EOF denial")
		return nil, nil
	}

	cmd := NewPaymentCmd(func() (*bootstrap.Result, error) { return &bootstrap.Result{Config: cfg}, nil })
	out, err := executePaymentCommandWithInput(t, cmd, "", "send", "--to", "0xto", "--amount", "1.50", "--purpose", "API access")
	require.NoError(t, err)
	assert.Contains(t, out, "Confirm [y/N]:")
	assert.Contains(t, out, "Aborted.")
}

func TestBalanceCmd_WritesTableOutputToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = true
	cfg.Payment.Network.ChainID = 84532
	cmd := NewPaymentCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	out, err := executePaymentCommand(t, cmd, "balance")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "payment")
	assert.NotContains(t, out, "Wallet Balance")
}

func TestBalanceCmd_WritesJSONOutputToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = true
	cmd := NewPaymentCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	out, err := executePaymentCommand(t, cmd, "balance", "--output", "json")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "payment")
	assert.Equal(t, "", out)
}
