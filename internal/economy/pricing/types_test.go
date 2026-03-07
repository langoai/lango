package pricing

import (
	"testing"
)

func TestPriceModifierType_StringValues(t *testing.T) {
	tests := []struct {
		give string
		want PriceModifierType
	}{
		{give: "trust_discount", want: ModifierTrustDiscount},
		{give: "volume_discount", want: ModifierVolumeDiscount},
		{give: "surge", want: ModifierSurge},
		{give: "custom", want: ModifierCustom},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			if string(tt.want) != tt.give {
				t.Errorf("PriceModifierType %q: got %q, want %q", tt.give, string(tt.want), tt.give)
			}
		})
	}
}

func TestQuote_IsFreeDefault(t *testing.T) {
	var q Quote
	if q.IsFree {
		t.Error("zero-value Quote should not be free")
	}
}
