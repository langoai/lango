package policy

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/smartaccount/bindings"
)

// Syncer synchronizes Go-side harness policies with on-chain SpendingHook limits.
type Syncer struct {
	engine *Engine
	hook   *bindings.SpendingHookClient
}

// NewSyncer creates a policy syncer.
func NewSyncer(engine *Engine, hook *bindings.SpendingHookClient) *Syncer {
	return &Syncer{engine: engine, hook: hook}
}

// PushToChain pushes the current Go-side policy to the SpendingHook contract.
// It converts HarnessPolicy limits to SpendingHook setLimits format:
//
//	MaxTxAmount   -> perTxLimit
//	DailyLimit    -> dailyLimit
//	MonthlyLimit  -> cumulativeLimit
func (s *Syncer) PushToChain(ctx context.Context, account common.Address) (string, error) {
	policy, ok := s.engine.GetPolicy(account)
	if !ok {
		return "", fmt.Errorf("no policy for account %s", account.Hex())
	}

	perTx := policy.MaxTxAmount
	if perTx == nil {
		perTx = new(big.Int) // 0 = unlimited
	}
	daily := policy.DailyLimit
	if daily == nil {
		daily = new(big.Int)
	}
	cumulative := policy.MonthlyLimit
	if cumulative == nil {
		cumulative = new(big.Int)
	}

	return s.hook.SetLimits(ctx, perTx, daily, cumulative)
}

// PullFromChain reads the on-chain SpendingHook config and updates the
// Go-side policy. Returns the fetched config for inspection.
func (s *Syncer) PullFromChain(ctx context.Context, account common.Address) (*bindings.SpendingConfig, error) {
	cfg, err := s.hook.GetConfig(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("get on-chain config: %w", err)
	}

	// Update Go-side policy with on-chain limits.
	policy, _ := s.engine.GetPolicy(account)
	if policy == nil {
		policy = &HarnessPolicy{}
	}

	updated := *policy
	if cfg.PerTxLimit != nil && cfg.PerTxLimit.Sign() > 0 {
		updated.MaxTxAmount = new(big.Int).Set(cfg.PerTxLimit)
	}
	if cfg.DailyLimit != nil && cfg.DailyLimit.Sign() > 0 {
		updated.DailyLimit = new(big.Int).Set(cfg.DailyLimit)
	}
	if cfg.CumulativeLimit != nil && cfg.CumulativeLimit.Sign() > 0 {
		updated.MonthlyLimit = new(big.Int).Set(cfg.CumulativeLimit)
	}

	s.engine.SetPolicy(account, &updated)
	return cfg, nil
}

// DriftReport describes differences between Go-side and on-chain policy.
type DriftReport struct {
	Account       common.Address
	HasDrift      bool
	GoPolicy      *HarnessPolicy
	OnChainConfig *bindings.SpendingConfig
	Differences   []string
}

// DetectDrift compares Go-side and on-chain policies and reports differences.
func (s *Syncer) DetectDrift(ctx context.Context, account common.Address) (*DriftReport, error) {
	goPolicy, ok := s.engine.GetPolicy(account)
	if !ok {
		return nil, fmt.Errorf("no Go-side policy for account %s", account.Hex())
	}

	onChain, err := s.hook.GetConfig(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("get on-chain config: %w", err)
	}

	report := &DriftReport{
		Account:       account,
		GoPolicy:      goPolicy,
		OnChainConfig: onChain,
	}

	if !bigIntEqual(goPolicy.MaxTxAmount, onChain.PerTxLimit) {
		report.HasDrift = true
		report.Differences = append(report.Differences,
			fmt.Sprintf("perTxLimit: go=%v on-chain=%v", goPolicy.MaxTxAmount, onChain.PerTxLimit))
	}
	if !bigIntEqual(goPolicy.DailyLimit, onChain.DailyLimit) {
		report.HasDrift = true
		report.Differences = append(report.Differences,
			fmt.Sprintf("dailyLimit: go=%v on-chain=%v", goPolicy.DailyLimit, onChain.DailyLimit))
	}
	if !bigIntEqual(goPolicy.MonthlyLimit, onChain.CumulativeLimit) {
		report.HasDrift = true
		report.Differences = append(report.Differences,
			fmt.Sprintf("cumulativeLimit: go=%v on-chain=%v", goPolicy.MonthlyLimit, onChain.CumulativeLimit))
	}

	return report, nil
}

// bigIntEqual compares two *big.Int values, treating nil as zero.
func bigIntEqual(a, b *big.Int) bool {
	if a == nil {
		a = new(big.Int)
	}
	if b == nil {
		b = new(big.Int)
	}
	return a.Cmp(b) == 0
}
