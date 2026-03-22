package automation

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/session"
	"github.com/stretchr/testify/assert"
)

func TestDetectChannelFromContext(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{
			give: "telegram:123456789:42",
			want: "telegram:123456789",
		},
		{
			give: "discord:chan-abc:user-xyz",
			want: "discord:chan-abc",
		},
		{
			give: "slack:C12345:U67890",
			want: "slack:C12345",
		},
		{
			give: "",
			want: "",
		},
		{
			give: "unknown:foo:bar",
			want: "",
		},
		{
			give: "onlyone",
			want: "",
		},
		{
			give: "telegram",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			ctx := context.Background()
			if tt.give != "" {
				ctx = session.WithSessionKey(ctx, tt.give)
			}
			got := DetectChannelFromContext(ctx)
			assert.Equal(t, tt.want, got)
		})
	}
}
