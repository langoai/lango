package pricing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriceModifierType_StringValues(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			assert.Equal(t, tt.give, string(tt.want))
		})
	}
}

func TestQuote_IsFreeDefault(t *testing.T) {
	t.Parallel()

	var q Quote
	assert.False(t, q.IsFree)
}
