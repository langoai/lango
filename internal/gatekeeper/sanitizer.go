package gatekeeper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/langoai/lango/internal/config"
)

// Sanitizer removes internal content from model responses.
type Sanitizer struct {
	cfg            config.GatekeeperConfig
	thoughtPattern *regexp.Regexp
	markerPattern  *regexp.Regexp
	jsonPattern    *regexp.Regexp
	customPatterns []*regexp.Regexp
	blankLine      *regexp.Regexp
}

// NewSanitizer creates a new Sanitizer. Returns error if custom patterns are invalid.
func NewSanitizer(cfg config.GatekeeperConfig) (*Sanitizer, error) {
	s := &Sanitizer{
		cfg:            cfg,
		thoughtPattern: regexp.MustCompile(`(?s)<(thought|thinking)>.*?</(thought|thinking)>`),
		markerPattern:  regexp.MustCompile(`(?m)^\s*\[(INTERNAL|DEBUG|SYSTEM|OBSERVATION)\].*$`),
		jsonPattern:    regexp.MustCompile("(?s)```(?:json)?\\s*\\n(.*?)```"),
		blankLine:      regexp.MustCompile(`\n{3,}`),
	}
	for _, p := range cfg.CustomPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("compile custom pattern %q: %w", p, err)
		}
		s.customPatterns = append(s.customPatterns, re)
	}
	return s, nil
}

// Enabled reports whether the sanitizer is active.
func (s *Sanitizer) Enabled() bool {
	return boolDefault(s.cfg.Enabled, true)
}

// Sanitize applies all sanitization rules to the text.
func (s *Sanitizer) Sanitize(text string) string {
	if !s.Enabled() {
		return text
	}

	// 1. Strip thought/thinking tags (but preserve code blocks)
	if boolDefault(s.cfg.StripThoughtTags, true) {
		text = s.stripThoughtTags(text)
	}

	// 2. Strip internal markers
	if boolDefault(s.cfg.StripInternalMarkers, true) {
		text = s.markerPattern.ReplaceAllString(text, "")
	}

	// 3. Replace large JSON code blocks
	if boolDefault(s.cfg.StripRawJSON, true) {
		threshold := s.cfg.RawJSONThreshold
		if threshold <= 0 {
			threshold = 500
		}
		text = s.jsonPattern.ReplaceAllStringFunc(text, func(match string) string {
			inner := s.jsonPattern.FindStringSubmatch(match)
			if len(inner) > 1 && len(inner[1]) > threshold {
				return "[Large data block omitted]"
			}
			return match
		})
	}

	// 4. Custom patterns
	for _, re := range s.customPatterns {
		text = re.ReplaceAllString(text, "")
	}

	// 5. Collapse multiple blank lines
	text = s.blankLine.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}

// stripThoughtTags removes thought/thinking tags while preserving code blocks.
func (s *Sanitizer) stripThoughtTags(text string) string {
	codeBlockRe := regexp.MustCompile("(?s)```.*?```")
	var codeBlocks []string
	placeholder := "\x00CODEBLOCK_%d\x00"

	protected := codeBlockRe.ReplaceAllStringFunc(text, func(match string) string {
		idx := len(codeBlocks)
		codeBlocks = append(codeBlocks, match)
		return fmt.Sprintf(placeholder, idx)
	})

	protected = s.thoughtPattern.ReplaceAllString(protected, "")

	for i, cb := range codeBlocks {
		protected = strings.Replace(protected, fmt.Sprintf(placeholder, i), cb, 1)
	}

	return protected
}

func boolDefault(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}
