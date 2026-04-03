## Tasks

- [x] Create `internal/tools/webfetch/readability.go` with HTML parsing and content extraction
  - extractText: parse HTML, find content root (article > main > largest div), strip nav/footer/aside/script/style, return clean text
  - extractMarkdown: same content root logic, render headings as `#`, links as `[text](url)`, lists as `- item`, bold/italic markers
  - findTitle: extract from `<title>` element
  - findContentRoot: priority chain article > main > div[role=main] > largest div

- [x] Create `internal/tools/webfetch/fetch.go` with HTTP fetch and orchestration
  - Fetch function: HTTP GET with 30s timeout, max 5 redirects, Lango/1.0 User-Agent
  - FetchResult struct: URL, Title, Content, ContentLength, Truncated
  - Mode switching: text/html/markdown with content truncation to max_length
  - 5MB body read limit to prevent memory exhaustion
  - URL scheme auto-prepend (https://) for bare hostnames
  - ValidateURLForP2P: block file://, localhost, loopback, private network ranges

- [x] Create `internal/tools/webfetch/tools.go` with BuildTools registration
  - Single web_fetch tool with SafetyLevelModerate
  - ToolCapability: category "web", aliases, search hints, ActivityRead
  - Handler: param extraction, P2P validation before fetch and after redirect, mode/max_length defaults

- [x] Create `internal/tools/webfetch/fetch_test.go` with comprehensive tests
  - TestFetch_TextMode: verify title, content extraction, nav/footer/script stripping
  - TestFetch_HTMLMode: verify raw HTML returned with title
  - TestFetch_MarkdownMode: verify heading/link/list markdown formatting
  - TestFetch_MaxLengthTruncation: verify truncation and flag
  - TestFetch_Non200Error / ServerError: verify HTTP error handling
  - TestFetch_InvalidMode / EmptyURL: verify input validation
  - TestFetch_Redirect: verify redirect following and final URL tracking
  - TestValidateURLForP2P: table-driven tests for all blocked/allowed URL patterns
  - TestExtractText_Readability: verify article and main element priority
  - TestExtractMarkdown: verify markdown rendering of all element types
  - TestBuildTools_ToolDefinition: verify tool metadata
  - TestBuildTools_Handler: verify end-to-end handler invocation

- [x] Verify build: `go build ./...` passes
- [x] Verify tests: `go test ./internal/tools/webfetch/...` all 18 tests pass
- [x] Verify vet: `go vet ./internal/tools/webfetch/...` clean
