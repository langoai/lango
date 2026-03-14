package tooloutput

import (
	"regexp"
	"strings"
)

// ContentType represents the detected content type of tool output.
type ContentType string

const (
	ContentTypeJSON       ContentType = "json"
	ContentTypeLog        ContentType = "log"
	ContentTypeCode       ContentType = "code"
	ContentTypeStackTrace ContentType = "stacktrace"
	ContentTypeText       ContentType = "text"
)

var (
	stackTracePatterns = []*regexp.Regexp{
		regexp.MustCompile(`goroutine \d+`),
		regexp.MustCompile(`^panic:`),
		regexp.MustCompile(`at .*:\d+`),
		regexp.MustCompile(`Traceback`),
		regexp.MustCompile(`\.java:\d+`),
	}

	logTimestampPattern = regexp.MustCompile(
		`(?m)^\d{4}[-/]\d{2}[-/]\d{2}[T ]\d{2}:\d{2}`,
	)

	logLevelPattern = regexp.MustCompile(`(?i)\b(ERROR|WARN|INFO|DEBUG)\b`)

	codeKeywords = []string{
		"func ", "class ", "def ", "import ", "package ",
		"var ", "const ", "type ", "interface ", "struct ",
	}
)

// DetectContentType analyzes text and returns the most likely content type.
func DetectContentType(text string) ContentType {
	if text == "" {
		return ContentTypeText
	}

	trimmed := strings.TrimSpace(text)

	// 1. JSON: starts with { or [
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		return ContentTypeJSON
	}

	// 2. StackTrace: contains goroutine/panic/traceback patterns
	for _, re := range stackTracePatterns {
		if re.MatchString(text) {
			return ContentTypeStackTrace
		}
	}

	// 3. Log: multiple lines with timestamps and log levels
	lines := strings.Split(text, "\n")
	if len(lines) >= 2 {
		timestampCount := len(logTimestampPattern.FindAllStringIndex(text, 2))
		levelCount := len(logLevelPattern.FindAllStringIndex(text, 2))
		if timestampCount >= 2 && levelCount >= 2 {
			return ContentTypeLog
		}
	}

	// 4. Code: contains syntax keywords
	keywordCount := 0
	for _, kw := range codeKeywords {
		if strings.Contains(text, kw) {
			keywordCount++
		}
	}
	if keywordCount >= 2 {
		return ContentTypeCode
	}

	// 5. Text: default
	return ContentTypeText
}
