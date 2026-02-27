package tui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetVersionInfo(t *testing.T) {
	SetVersionInfo("1.2.3", "2026-01-01")
	assert.Equal(t, "1.2.3", _version)
	assert.Equal(t, "2026-01-01", _buildTime)
}

func TestSetProfile(t *testing.T) {
	SetProfile("production")
	assert.Equal(t, "production", _profile)
}

func TestBanner_ContainsLango(t *testing.T) {
	SetVersionInfo("0.4.0", "2026-01-01")
	SetProfile("default")

	banner := Banner()
	assert.True(t, strings.Contains(banner, "Lango"))
	assert.True(t, strings.Contains(banner, "0.4.0"))
}

func TestServeBanner_ContainsVersion(t *testing.T) {
	SetVersionInfo("0.5.0", "2026-02-01")

	serve := ServeBanner()
	assert.True(t, strings.Contains(serve, "0.5.0"))
	assert.True(t, strings.Contains(serve, "─"))
}

func TestBannerBox_HasBorder(t *testing.T) {
	SetVersionInfo("1.0.0", "2026-01-01")

	box := BannerBox()
	// Rounded border uses characters like ╭ ╮ ╰ ╯
	assert.True(t, strings.Contains(box, "╭") || strings.Contains(box, "│"))
}

func TestSquirrelFace(t *testing.T) {
	face := squirrelFace()
	assert.True(t, strings.Contains(face, "●.●"))
}
