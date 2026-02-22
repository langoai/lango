package learning

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	entlearning "github.com/langoai/lango/internal/ent/learning"
)

var (
	_uuidRegex      = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	_timestampRegex = regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}`)
	_pathRegex      = regexp.MustCompile(`/[^\s:]+/`)
	_portRegex      = regexp.MustCompile(`:\d{4,5}`)
)

func extractErrorPattern(err error) string {
	msg := err.Error()
	msg = _uuidRegex.ReplaceAllString(msg, "")
	msg = _timestampRegex.ReplaceAllString(msg, "")
	msg = _pathRegex.ReplaceAllString(msg, "<path>")
	msg = _portRegex.ReplaceAllString(msg, ":<port>")
	return strings.TrimSpace(msg)
}

func categorizeError(toolName string, err error) entlearning.Category {
	msg := strings.ToLower(err.Error())

	switch {
	case isDeadlineExceeded(err) || strings.Contains(msg, "deadline exceeded") || strings.Contains(msg, "timeout"):
		return entlearning.CategoryTimeout
	case strings.Contains(msg, "permission denied") || strings.Contains(msg, "access denied") || strings.Contains(msg, "forbidden"):
		return entlearning.CategoryPermission
	case strings.Contains(msg, "api") || strings.Contains(msg, "model") || strings.Contains(msg, "provider") || strings.Contains(msg, "rate limit"):
		return entlearning.CategoryProviderError
	default:
		if toolName != "" {
			return entlearning.CategoryToolError
		}
		return entlearning.CategoryGeneral
	}
}

func isDeadlineExceeded(err error) bool {
	return errors.Is(err, context.DeadlineExceeded)
}

func summarizeParams(params map[string]interface{}) map[string]interface{} {
	if params == nil {
		return nil
	}

	result := make(map[string]interface{}, len(params))
	for k, v := range params {
		switch val := v.(type) {
		case string:
			if len(val) > 200 {
				result[k] = val[:200] + "..."
			} else {
				result[k] = val
			}
		case []interface{}:
			result[k] = fmt.Sprintf("[%d items]", len(val))
		default:
			result[k] = v
		}
	}
	return result
}
