// Package firewall implements the Knowledge Firewall for P2P queries.
package firewall

import (
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// OwnerProtectionConfig configures owner data protection.
type OwnerProtectionConfig struct {
	OwnerName          string   `json:"ownerName"`
	OwnerEmail         string   `json:"ownerEmail"`
	OwnerPhone         string   `json:"ownerPhone"`
	ExtraTerms         []string `json:"extraTerms,omitempty"`
	BlockConversations bool     `json:"blockConversations"`
}

// OwnerShield prevents owner personal data from leaking via P2P responses.
// No amount of USDC can bypass this layer.
type OwnerShield struct {
	exactTerms    []string
	regexPatterns []*regexp.Regexp
	blockConvKeys bool
	logger        *zap.SugaredLogger
}

// conversationKeys are substrings that identify conversation-related fields.
var conversationKeys = []string{
	"conversation",
	"message_history",
	"chat_log",
	"session_history",
	"chat_history",
}

const redactedPlaceholder = "[owner-data-redacted]"

// NewOwnerShield creates a new OwnerShield from the given config.
func NewOwnerShield(cfg OwnerProtectionConfig, logger *zap.SugaredLogger) *OwnerShield {
	var exactTerms []string
	for _, term := range []string{cfg.OwnerName, cfg.OwnerEmail, cfg.OwnerPhone} {
		if term != "" {
			exactTerms = append(exactTerms, strings.ToLower(term))
		}
	}
	for _, term := range cfg.ExtraTerms {
		if term != "" {
			exactTerms = append(exactTerms, strings.ToLower(term))
		}
	}

	regexPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\b[\w.+-]+@[\w.-]+\.\w{2,}\b`),
		regexp.MustCompile(`\b\d{2,4}[-.]?\d{3,4}[-.]?\d{4}\b`),
	}

	return &OwnerShield{
		exactTerms:    exactTerms,
		regexPatterns: regexPatterns,
		blockConvKeys: cfg.BlockConversations,
		logger:        logger,
	}
}

// ScanAndRedact recursively walks the response map and redacts owner data.
// It returns the redacted map and a list of redacted field paths.
func (s *OwnerShield) ScanAndRedact(response map[string]interface{}) (map[string]interface{}, []string) {
	result := make(map[string]interface{}, len(response))
	var blocked []string
	s.scanMap(response, result, "", &blocked)
	return result, blocked
}

// ContainsOwnerData checks if the text contains any owner data.
func (s *OwnerShield) ContainsOwnerData(text string) bool {
	lower := strings.ToLower(text)
	for _, term := range s.exactTerms {
		if strings.Contains(lower, term) {
			return true
		}
	}
	for _, re := range s.regexPatterns {
		if re.MatchString(text) {
			return true
		}
	}
	return false
}

func (s *OwnerShield) scanMap(src, dst map[string]interface{}, prefix string, blocked *[]string) {
	for k, v := range src {
		path := joinPath(prefix, k)

		// Block conversation-related keys entirely.
		if s.blockConvKeys && isConversationKey(k) {
			dst[k] = redactedPlaceholder
			*blocked = append(*blocked, path)
			continue
		}

		switch val := v.(type) {
		case string:
			if s.containsOwnerMatch(val) {
				dst[k] = redactedPlaceholder
				*blocked = append(*blocked, path)
			} else {
				dst[k] = val
			}
		case map[string]interface{}:
			nested := make(map[string]interface{}, len(val))
			s.scanMap(val, nested, path, blocked)
			dst[k] = nested
		case []interface{}:
			dst[k] = s.scanSlice(val, path, blocked)
		default:
			dst[k] = v
		}
	}
}

func (s *OwnerShield) scanSlice(src []interface{}, prefix string, blocked *[]string) []interface{} {
	result := make([]interface{}, len(src))
	for i, elem := range src {
		path := fmt.Sprintf("%s[%d]", prefix, i)
		switch val := elem.(type) {
		case string:
			if s.containsOwnerMatch(val) {
				result[i] = redactedPlaceholder
				*blocked = append(*blocked, path)
			} else {
				result[i] = val
			}
		case map[string]interface{}:
			nested := make(map[string]interface{}, len(val))
			s.scanMap(val, nested, path, blocked)
			result[i] = nested
		case []interface{}:
			result[i] = s.scanSlice(val, path, blocked)
		default:
			result[i] = elem
		}
	}
	return result
}

func (s *OwnerShield) containsOwnerMatch(text string) bool {
	lower := strings.ToLower(text)
	for _, term := range s.exactTerms {
		if strings.Contains(lower, term) {
			return true
		}
	}
	for _, re := range s.regexPatterns {
		if re.MatchString(text) {
			return true
		}
	}
	return false
}

func isConversationKey(key string) bool {
	lower := strings.ToLower(key)
	for _, ck := range conversationKeys {
		if strings.Contains(lower, ck) {
			return true
		}
	}
	return false
}

func joinPath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}
