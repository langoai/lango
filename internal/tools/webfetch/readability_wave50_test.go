package webfetch

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestWave50FindContentRootPrefersRoleMainDiv(t *testing.T) {
	t.Parallel()

	doc := parseWave50HTML(t, `<html><body>
		<div id="small">tiny</div>
		<section><div role="main" id="target"><p>Main role content</p></div></section>
		<div id="large">This div has more text but should not win over role main.</div>
	</body></html>`)

	root := findContentRoot(doc)
	require.NotNil(t, root)
	assert.Equal(t, atom.Div, root.DataAtom)
	assert.Equal(t, "target", attrVal(root, "id"))
	assert.Same(t, root, findByAttr(doc, atom.Div, "role", "main"))
}

func TestWave50FindLargestDivChoosesMostTextWhenSemanticRootsAreAbsent(t *testing.T) {
	t.Parallel()

	doc := parseWave50HTML(t, `<html><body>
		<section><div id="short">short</div></section>
		<div id="largest"><p>This is the longest useful content block.</p></div>
		<div id="middle">medium content</div>
	</body></html>`)

	root := findContentRoot(doc)
	require.NotNil(t, root)
	assert.Equal(t, "largest", attrVal(root, "id"))

	largest := findLargestDiv(doc)
	require.NotNil(t, largest)
	assert.Equal(t, "largest", attrVal(largest, "id"))
}

func TestWave50FindLargestDivReturnsNilWithoutDivs(t *testing.T) {
	t.Parallel()

	doc := parseWave50HTML(t, `<html><body><section>content</section></body></html>`)

	assert.Nil(t, findLargestDiv(doc))
	assert.Nil(t, findByAttr(doc, atom.Div, "role", "main"))
}

func TestWave50ExtractMarkdownCoversInlineAndBlockBranches(t *testing.T) {
	t.Parallel()

	title, body, err := extractMarkdown(strings.NewReader(`<html><head><title>Branches</title></head><body>
		<main>
			<h4>Small Heading</h4>
			<p>Line<br>Break with <code>inline</code> and <a>plain link</a>.</p>
			<blockquote>quoted text</blockquote>
			<ol><li>First ordered</li><li>Second ordered</li></ol>
		</main>
	</body></html>`))
	require.NoError(t, err)

	assert.Equal(t, "Branches", title)
	assert.Contains(t, body, "#### Small Heading")
	assert.Contains(t, body, "Line\nBreak")
	assert.Contains(t, body, "`inline`")
	assert.Contains(t, body, "plain link")
	assert.Contains(t, body, "> quoted text")
	assert.Contains(t, body, "- First ordered")
	assert.Contains(t, body, "- Second ordered")
}

func parseWave50HTML(t *testing.T, input string) *html.Node {
	t.Helper()

	doc, err := html.Parse(strings.NewReader(input))
	require.NoError(t, err)
	return doc
}
