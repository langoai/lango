package app

import (
	"testing"

	"github.com/langoai/lango/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestParseDeliveryTarget(t *testing.T) {
	tests := []struct {
		give       string
		wantType   types.ChannelType
		wantTarget string
	}{
		{
			give:       "telegram:123456789",
			wantType:   types.ChannelTelegram,
			wantTarget: "123456789",
		},
		{
			give:       "discord:channel-id-here",
			wantType:   types.ChannelDiscord,
			wantTarget: "channel-id-here",
		},
		{
			give:       "slack:C12345",
			wantType:   types.ChannelSlack,
			wantTarget: "C12345",
		},
		{
			give:       "telegram",
			wantType:   types.ChannelTelegram,
			wantTarget: "",
		},
		{
			give:       "discord",
			wantType:   types.ChannelDiscord,
			wantTarget: "",
		},
		{
			give:       "slack",
			wantType:   types.ChannelSlack,
			wantTarget: "",
		},
		{
			give:       "  TELEGRAM:999  ",
			wantType:   types.ChannelTelegram,
			wantTarget: "999",
		},
		{
			give:       "  Discord  ",
			wantType:   types.ChannelDiscord,
			wantTarget: "",
		},
		{
			give:       "unknown:abc",
			wantType:   types.ChannelType("unknown"),
			wantTarget: "abc",
		},
		{
			give:       "",
			wantType:   types.ChannelType(""),
			wantTarget: "",
		},
		{
			give:       "telegram:chat:extra",
			wantType:   types.ChannelTelegram,
			wantTarget: "chat:extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			gotType, gotTarget := parseDeliveryTarget(tt.give)
			assert.Equal(t, tt.wantType, gotType, "channel type")
			assert.Equal(t, tt.wantTarget, gotTarget, "target ID")
		})
	}
}
