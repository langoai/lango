package x402

import (
	"testing"
)

func TestCAIP2Network(t *testing.T) {
	tests := []struct {
		give    int64
		want    string
	}{
		{give: 1, want: "eip155:1"},
		{give: 8453, want: "eip155:8453"},
		{give: 84532, want: "eip155:84532"},
		{give: 11155111, want: "eip155:11155111"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := CAIP2Network(tt.give)
			if got != tt.want {
				t.Errorf("CAIP2Network(%d) = %q, want %q", tt.give, got, tt.want)
			}
		})
	}
}

func TestConfig(t *testing.T) {
	cfg := Config{
		Enabled:          true,
		ChainID:          84532,
		MaxAutoPayAmount: "1.00",
	}

	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if cfg.ChainID != 84532 {
		t.Errorf("ChainID = %d, want 84532", cfg.ChainID)
	}
	if cfg.MaxAutoPayAmount != "1.00" {
		t.Errorf("MaxAutoPayAmount = %q, want %q", cfg.MaxAutoPayAmount, "1.00")
	}
}
