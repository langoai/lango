package webfetch

import (
	"context"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/toolparam"
)

// BuildTools creates web fetch agent tools.
func BuildTools() []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "web_fetch",
			Description: "Fetch a web page and extract its content. Supports text, HTML, and markdown output modes.",
			SafetyLevel: agent.SafetyLevelModerate,
			Capability: agent.ToolCapability{
				Category:    "web",
				Aliases:     []string{"fetch_url", "get_page"},
				SearchHints: []string{"fetch", "download", "page", "url", "content"},
				Activity:    agent.ActivityRead,
			},
			Parameters: agent.Schema().
				Str("url", "The URL of the web page to fetch").
				Enum("mode", "Output mode for extracted content", ModeText, ModeHTML, ModeMarkdown).
				Int("max_length", "Maximum character length of returned content (default: 5000)").
				Required("url").
				Build(),
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				rawURL, err := toolparam.RequireString(params, "url")
				if err != nil {
					return nil, err
				}

				// Block internal/private network URLs in P2P context.
				if ctxkeys.IsP2PRequest(ctx) {
					if err := ValidateURLForP2P(rawURL); err != nil {
						return nil, err
					}
				}

				mode := toolparam.OptionalString(params, "mode", ModeText)
				maxLength := toolparam.OptionalInt(params, "max_length", defaultMaxLength)

				result, err := Fetch(ctx, rawURL, mode, maxLength)
				if err != nil {
					return nil, err
				}

				// Re-validate final URL after redirects in P2P context.
				if ctxkeys.IsP2PRequest(ctx) {
					if err := ValidateURLForP2P(result.URL); err != nil {
						return nil, err
					}
				}

				return result, nil
			},
		},
	}
}
