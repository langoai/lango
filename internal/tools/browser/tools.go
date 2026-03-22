package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/langoai/lango/internal/agent"
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
			Description: "Navigate the browser to a URL and return the page title, URL, and a text snippet",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: agent.Schema().
				Str("url", "The URL to navigate to").
				Required("url").
				Build(),
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				url, err := toolparam.RequireString(params, "url")
				if err != nil {
					return nil, err
				}

				sessionID, err := sm.EnsureSession()
				if err != nil {
					return nil, err
				}

				if err := sm.Tool().Navigate(ctx, sessionID, url); err != nil {
					return nil, err
				}

				return sm.Tool().GetSnapshot(sessionID)
			},
		},
		{
			Name:        "browser_action",
			Description: "Perform an action on the current browser page: click, type, eval, get_text, get_element_info, or wait",
			SafetyLevel: agent.SafetyLevelDangerous,
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
