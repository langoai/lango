package tooloutput

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/langoai/lango/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectContentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want ContentType
	}{
		{
			give: `{"key": "value"}`,
			want: ContentTypeJSON,
		},
		{
			give: `[{"id": 1}, {"id": 2}]`,
			want: ContentTypeJSON,
		},
		{
			give: `  { "spaced": true }`,
			want: ContentTypeJSON,
		},
		{
			give: "goroutine 1 [running]:\nmain.main()\n\t/app/main.go:10",
			want: ContentTypeStackTrace,
		},
		{
			give: "panic: runtime error: index out of range\ngoroutine 1 [running]:",
			want: ContentTypeStackTrace,
		},
		{
			give: "Exception in thread \"main\"\n\tat com.example.Main.run(Main.java:42)\n\tat com.example.Main.main(Main.java:10)",
			want: ContentTypeStackTrace,
		},
		{
			give: "Traceback (most recent call last):\n  File \"app.py\", line 10\nNameError: name 'x' is not defined",
			want: ContentTypeStackTrace,
		},
		{
			give: "2024-01-15T10:30:00 INFO  Starting server\n2024-01-15T10:30:01 ERROR Connection refused\n2024-01-15T10:30:02 WARN  Retrying",
			want: ContentTypeLog,
		},
		{
			give: "2024/01/15 10:30:00 DEBUG init\n2024/01/15 10:30:01 INFO  ready",
			want: ContentTypeLog,
		},
		{
			give: "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}",
			want: ContentTypeCode,
		},
		{
			give: "class Foo:\n    def __init__(self):\n        pass\n\n    def bar(self):\n        return 42",
			want: ContentTypeCode,
		},
		{
			give: "Hello, this is plain text with no special markers.",
			want: ContentTypeText,
		},
		{
			give: "",
			want: ContentTypeText,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.want)+"_"+truncateGive(tt.give), func(t *testing.T) {
			t.Parallel()
			got := DetectContentType(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompressJSON(t *testing.T) {
	t.Parallel()

	t.Run("under_budget_passthrough", func(t *testing.T) {
		t.Parallel()
		input := `{"name": "test"}`
		got := CompressJSON(input, 1000)
		assert.Equal(t, input, got)
	})

	t.Run("array_compression", func(t *testing.T) {
		t.Parallel()
		items := make([]map[string]string, 50)
		for i := range items {
			items[i] = map[string]string{"id": strings.Repeat("x", 20)}
		}
		data, err := json.Marshal(items)
		require.NoError(t, err)
		input := string(data)

		got := CompressJSON(input, 50)
		assert.Contains(t, got, "more items")
		assert.Less(t, types.EstimateTokens(got), types.EstimateTokens(input))
	})

	t.Run("object_value_truncation", func(t *testing.T) {
		t.Parallel()
		obj := map[string]string{
			"description": strings.Repeat("a", 1000),
			"bio":         strings.Repeat("b", 1000),
			"name":        "short",
		}
		data, err := json.Marshal(obj)
		require.NoError(t, err)

		got := CompressJSON(string(data), 500)
		assert.Contains(t, got, "name")
		assert.Contains(t, got, "description")
		assert.Less(t, types.EstimateTokens(got), types.EstimateTokens(string(data)))
	})

	t.Run("invalid_json_fallback", func(t *testing.T) {
		t.Parallel()
		input := "{this is not valid json" + strings.Repeat("\nsome line", 100)
		got := CompressJSON(input, 20)
		assert.Contains(t, got, "compressed")
	})

	t.Run("empty_string", func(t *testing.T) {
		t.Parallel()
		got := CompressJSON("", 100)
		assert.Equal(t, "", got)
	})
}

func TestCompressLog(t *testing.T) {
	t.Parallel()

	t.Run("under_budget_passthrough", func(t *testing.T) {
		t.Parallel()
		input := "2024-01-15 10:00:00 INFO start\n2024-01-15 10:00:01 ERROR fail"
		got := CompressLog(input, 1000)
		assert.Equal(t, input, got)
	})

	t.Run("extracts_error_warn_lines", func(t *testing.T) {
		t.Parallel()
		var lines []string
		for i := 0; i < 100; i++ {
			lines = append(lines, "2024-01-15 10:00:00 INFO  routine log message number something")
		}
		lines = append(lines, "2024-01-15 10:00:01 ERROR something broke badly")
		lines = append(lines, "2024-01-15 10:00:02 WARN  disk space low warning")
		input := strings.Join(lines, "\n")

		got := CompressLog(input, 50)
		assert.Contains(t, got, "ERROR")
		assert.Contains(t, got, "WARN")
		assert.Contains(t, got, "log summary")
	})

	t.Run("no_error_lines_head_tail", func(t *testing.T) {
		t.Parallel()
		var lines []string
		for i := 0; i < 100; i++ {
			lines = append(lines, "2024-01-15 10:00:00 INFO  routine log message filling space")
		}
		input := strings.Join(lines, "\n")

		got := CompressLog(input, 30)
		assert.Contains(t, got, "log summary")
		assert.Contains(t, got, "compressed")
	})

	t.Run("empty_string", func(t *testing.T) {
		t.Parallel()
		got := CompressLog("", 100)
		assert.Equal(t, "", got)
	})
}

func TestCompressCode(t *testing.T) {
	t.Parallel()

	t.Run("under_budget_passthrough", func(t *testing.T) {
		t.Parallel()
		input := "package main\n\nfunc main() {}"
		got := CompressCode(input, 1000)
		assert.Equal(t, input, got)
	})

	t.Run("extracts_signatures", func(t *testing.T) {
		t.Parallel()
		var lines []string
		lines = append(lines, "package main")
		lines = append(lines, "")
		lines = append(lines, `import "fmt"`)
		lines = append(lines, "")
		lines = append(lines, "func hello() {")
		for i := 0; i < 100; i++ {
			lines = append(lines, "\t// some long comment filling up space in the code body here")
		}
		lines = append(lines, "}")
		lines = append(lines, "")
		lines = append(lines, "func world() {")
		lines = append(lines, "\tfmt.Println()")
		lines = append(lines, "}")
		input := strings.Join(lines, "\n")

		got := CompressCode(input, 50)
		assert.Contains(t, got, "signatures")
		assert.Contains(t, got, "package main")
		assert.Contains(t, got, "func hello")
	})

	t.Run("empty_string", func(t *testing.T) {
		t.Parallel()
		got := CompressCode("", 100)
		assert.Equal(t, "", got)
	})
}

func TestCompressStackTrace(t *testing.T) {
	t.Parallel()

	t.Run("under_budget_passthrough", func(t *testing.T) {
		t.Parallel()
		input := "goroutine 1 [running]:\nmain.main()\n\t/app/main.go:10"
		got := CompressStackTrace(input, 1000)
		assert.Equal(t, input, got)
	})

	t.Run("keeps_first_goroutine", func(t *testing.T) {
		t.Parallel()
		var blocks []string
		for i := 0; i < 20; i++ {
			block := []string{
				"goroutine " + strings.Repeat("1", 1) + " [running]:",
				"main.handler()",
				"\t/app/handler.go:42",
				"\t/app/main.go:10",
				"",
			}
			blocks = append(blocks, strings.Join(block, "\n"))
		}
		input := strings.Join(blocks, "\n")

		got := CompressStackTrace(input, 20)
		assert.Contains(t, got, "goroutine")
		assert.Contains(t, got, "more goroutines/threads omitted")
	})

	t.Run("single_goroutine_head_tail", func(t *testing.T) {
		t.Parallel()
		lines := []string{"goroutine 1 [running]:"}
		for i := 0; i < 100; i++ {
			lines = append(lines, "\tmain.func"+strings.Repeat("x", 30)+"()")
		}
		input := strings.Join(lines, "\n")

		got := CompressStackTrace(input, 20)
		assert.Contains(t, got, "compressed")
	})

	t.Run("empty_string", func(t *testing.T) {
		t.Parallel()
		got := CompressStackTrace("", 100)
		assert.Equal(t, "", got)
	})
}

func TestCompressHeadTail(t *testing.T) {
	t.Parallel()

	t.Run("under_budget_passthrough", func(t *testing.T) {
		t.Parallel()
		input := "line 1\nline 2\nline 3"
		got := CompressHeadTail(input, 0.7, 0.3, 1000)
		assert.Equal(t, input, got)
	})

	t.Run("respects_head_tail_ratio", func(t *testing.T) {
		t.Parallel()
		var lines []string
		for i := 0; i < 100; i++ {
			lines = append(lines, "this is line content that has some text in it")
		}
		input := strings.Join(lines, "\n")

		got := CompressHeadTail(input, 0.7, 0.3, 200)
		parts := strings.Split(got, "... [compressed:")
		require.Len(t, parts, 2, "expected separator in output")

		headPart := parts[0]
		headLines := strings.Split(strings.TrimRight(headPart, "\n"), "\n")

		// After the separator, the tail lines
		tailPart := strings.SplitN(parts[1], "] ...\n", 2)
		tailLines := []string{}
		if len(tailPart) > 1 && tailPart[1] != "" {
			tailLines = strings.Split(tailPart[1], "\n")
		}

		// Head should be larger than tail (0.7 vs 0.3 ratio)
		assert.Greater(t, len(headLines), len(tailLines), "head should have more lines than tail")
	})

	t.Run("separator_has_correct_counts", func(t *testing.T) {
		t.Parallel()
		var lines []string
		for i := 0; i < 50; i++ {
			lines = append(lines, "a]ine of text here for testing compression behavior")
		}
		input := strings.Join(lines, "\n")

		got := CompressHeadTail(input, 0.7, 0.3, 30)
		assert.Contains(t, got, "compressed: removed")
		assert.Contains(t, got, "lines")
		assert.Contains(t, got, "tokens")
	})

	t.Run("single_line", func(t *testing.T) {
		t.Parallel()
		input := strings.Repeat("x", 500)
		got := CompressHeadTail(input, 0.7, 0.3, 10)
		assert.Contains(t, got, "truncated")
		assert.Less(t, len(got), len(input))
	})

	t.Run("empty_string", func(t *testing.T) {
		t.Parallel()
		got := CompressHeadTail("", 0.7, 0.3, 100)
		assert.Equal(t, "", got)
	})
}

func TestCompress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give        string
		contentType ContentType
		maxTokens   int
		wantContain string
	}{
		{
			give:        generateLargeJSONArray(100),
			contentType: ContentTypeJSON,
			maxTokens:   30,
			wantContain: "more items",
		},
		{
			give:        strings.Repeat("2024-01-01 10:00:00 ERROR fail\n", 50),
			contentType: ContentTypeLog,
			maxTokens:   30,
			wantContain: "log summary",
		},
		{
			give:        "package main\nfunc main() {\n" + strings.Repeat("\t// comment line\n", 100) + "}",
			contentType: ContentTypeCode,
			maxTokens:   30,
			wantContain: "signatures",
		},
		{
			give:        "goroutine 1 [running]:\nmain.main()\n" + strings.Repeat("goroutine 2 [running]:\nfoo.bar()\n", 30),
			contentType: ContentTypeStackTrace,
			maxTokens:   20,
			wantContain: "omitted",
		},
		{
			give:        strings.Repeat("some plain text line\n", 100),
			contentType: ContentTypeText,
			maxTokens:   20,
			wantContain: "compressed",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.contentType), func(t *testing.T) {
			t.Parallel()
			got := Compress(tt.give, tt.contentType, tt.maxTokens, 0.7, 0.3)
			assert.Contains(t, got, tt.wantContain)
			assert.Less(t, types.EstimateTokens(got), types.EstimateTokens(tt.give))
		})
	}

	t.Run("under_budget_passthrough", func(t *testing.T) {
		t.Parallel()
		input := "small text"
		got := Compress(input, ContentTypeText, 1000, 0.7, 0.3)
		assert.Equal(t, input, got)
	})
}

// truncateGive returns a short version of the test input for subtest naming.
func truncateGive(s string) string {
	s = strings.ReplaceAll(s, "\n", "_")
	if len(s) > 30 {
		s = s[:30]
	}
	return s
}

// generateLargeJSONArray creates a valid JSON array with n items.
func generateLargeJSONArray(n int) string {
	items := make([]map[string]string, n)
	for i := range items {
		items[i] = map[string]string{"id": strings.Repeat("x", 20)}
	}
	data, _ := json.Marshal(items)
	return string(data)
}
