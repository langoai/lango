package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/toolparam"
)

// Browser action constants.
const (
	actionClick   = "click"
	actionType    = "type"
	actionEval    = "eval"
	actionGetText = "get_text"
	actionGetInfo = "get_element_info"
	actionWait    = "wait"
)

// BuildTools creates browser agent tools backed by the given SessionManager.
func BuildTools(sm *SessionManager) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "browser_navigate",
			Description: "Navigate the browser to a URL and return a structured page snapshot",
			SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Category:    "browser",
				Activity:    agent.ActivityExecute,
				Aliases:     []string{"goto", "open_url"},
				SearchHints: []string{"url", "navigate", "page"},
			},
			Parameters: agent.Schema().
				Str("url", "The URL to navigate to").
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

				sessionID, err := sm.EnsureSession()
				if err != nil {
					return nil, err
				}

				if err := sm.Tool().Navigate(ctx, sessionID, rawURL); err != nil {
					return nil, err
				}

				// Re-validate final URL after navigation in P2P context.
				// Always re-validate regardless of URL string equality to
				// prevent DNS rebinding attacks where the same hostname
				// resolves to a different (private) IP at navigation time.
				if ctxkeys.IsP2PRequest(ctx) {
					finalURL, err := sm.Tool().CurrentURL(sessionID)
					if err == nil {
						if err := ValidateURLForP2P(finalURL); err != nil {
							// Navigate away from blocked destination.
							_ = sm.Tool().Navigate(ctx, sessionID, "about:blank")
							return nil, fmt.Errorf("redirect to blocked URL: %w", err)
						}
					}
				}

				return sm.Tool().Snapshot(sessionID, defaultLinkLimit, defaultActionLimit)
			},
		},
		{
			Name:        "browser_search",
			Description: "[Deprecated: prefer web_search] Search the web using the browser and return structured search results.",
			SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Category:    "browser",
				Activity:    agent.ActivityQuery,
				Aliases:     []string{"search", "web_search_browser"},
				SearchHints: []string{"search", "web", "query"},
			},
			Parameters: agent.Schema().
				Str("query", "The search query to run").
				Int("limit", "Maximum number of results to return (default: 5)").
				Required("query").
				Build(),
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				query, err := toolparam.RequireString(params, "query")
				if err != nil {
					return nil, err
				}

				sessionID, err := sm.EnsureSession()
				if err != nil {
					return nil, err
				}

				limit := toolparam.OptionalInt(params, "limit", defaultSearchResultsLimit)
				return sm.Tool().Search(ctx, sessionID, query, limit)
			},
		},
		{
			Name:        "browser_observe",
			Description: "Return actionable elements from the current browser page with stable selectors",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:    "browser",
				Activity:    agent.ActivityRead,
				ReadOnly:    true,
				Aliases:     []string{"observe_page", "inspect"},
				SearchHints: []string{"elements", "selectors", "page"},
			},
			Parameters: agent.Schema().
				Int("limit", "Maximum number of actionable elements to return (default: 10)").
				Build(),
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				_ = ctx
				sessionID, err := sm.EnsureSession()
				if err != nil {
					return nil, err
				}

				limit := toolparam.OptionalInt(params, "limit", defaultObservationLimit)
				return sm.Tool().Observe(sessionID, limit)
			},
		},
		{
			Name:        "browser_extract",
			Description: "Extract structured data from the current page: summary, links, article, or search_results",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "browser",
				Activity:        agent.ActivityRead,
				ReadOnly:        true,
				ConcurrencySafe: true,
				Aliases:         []string{"extract_content", "scrape"},
				SearchHints:     []string{"content", "article", "links"},
			},
			Parameters: agent.Schema().
				Enum("mode", "The extraction mode", "summary", "links", "article", "search_results").
				Int("limit", "Maximum number of extracted items where applicable").
				Build(),
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				_ = ctx
				sessionID, err := sm.EnsureSession()
				if err != nil {
					return nil, err
				}

				mode := toolparam.OptionalString(params, "mode", "summary")
				limit := toolparam.OptionalInt(params, "limit", defaultLinkLimit)
				return sm.Tool().Extract(sessionID, mode, limit)
			},
		},
		{
			Name:        "browser_action",
			Description: "Perform an action on the current browser page: click, type, eval, get_text, get_element_info, or wait",
			SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Category:    "browser",
				Activity:    agent.ActivityExecute,
				Aliases:     []string{"click", "type", "interact"},
				SearchHints: []string{"click", "type", "scroll"},
			},
			Parameters: agent.Schema().
				Enum("action", "The action to perform", actionClick, actionType, actionEval, actionGetText, actionGetInfo, actionWait).
				Str("selector", "CSS selector for the target element (required for click, type, get_text, get_element_info, wait)").
				Str("text", "Text to type (required for type action) or JavaScript to evaluate (required for eval action)").
				Int("timeout", "Timeout in seconds for wait action (default: 10)").
				Required("action").
				Build(),
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				action, err := toolparam.RequireString(params, "action")
				if err != nil {
					return nil, err
				}

				// Block eval action for P2P requests before session creation.
				if action == actionEval && ctxkeys.IsP2PRequest(ctx) {
					return nil, ErrEvalBlockedP2P
				}

				sessionID, err := sm.EnsureSession()
				if err != nil {
					return nil, err
				}

				selector := toolparam.OptionalString(params, "selector", "")
				text := toolparam.OptionalString(params, "text", "")

				switch action {
				case actionClick:
					if selector == "" {
						return nil, fmt.Errorf("selector required for click action")
					}
					return nil, sm.Tool().Click(ctx, sessionID, selector)

				case actionType:
					if selector == "" {
						return nil, fmt.Errorf("selector required for type action")
					}
					if text == "" {
						return nil, fmt.Errorf("text required for type action")
					}
					return nil, sm.Tool().Type(ctx, sessionID, selector, text)

				case actionEval:
					if text == "" {
						return nil, fmt.Errorf("text (JavaScript) required for eval action")
					}
					return sm.Tool().Eval(sessionID, text)

				case actionGetText:
					if selector == "" {
						return nil, fmt.Errorf("selector required for get_text action")
					}
					return sm.Tool().GetText(sessionID, selector)

				case actionGetInfo:
					if selector == "" {
						return nil, fmt.Errorf("selector required for get_element_info action")
					}
					return sm.Tool().GetElementInfo(sessionID, selector)

				case actionWait:
					if selector == "" {
						return nil, fmt.Errorf("selector required for wait action")
					}
					timeout := time.Duration(toolparam.OptionalInt(params, "timeout", 10)) * time.Second
					return nil, sm.Tool().WaitForSelector(ctx, sessionID, selector, timeout)

				default:
					return nil, fmt.Errorf("unknown action: %s", action)
				}
			},
		},
		{
			Name:        "browser_screenshot",
			Description: "Capture a screenshot of the current browser page as base64 PNG",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:    "browser",
				Activity:    agent.ActivityRead,
				ReadOnly:    true,
				Aliases:     []string{"screenshot", "capture"},
				SearchHints: []string{"screenshot", "image", "capture"},
			},
			Parameters: agent.Schema().
				Bool("fullPage", "Capture the full scrollable page (default: false)").
				Build(),
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				sessionID, err := sm.EnsureSession()
				if err != nil {
					return nil, err
				}

				fullPage := toolparam.OptionalBool(params, "fullPage", false)
				return sm.Tool().Screenshot(sessionID, fullPage)
			},
		},
	}
}
