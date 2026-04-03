// Package webfetch provides an HTTP fetch tool with content extraction for agent use.
package webfetch

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// stripTags are elements removed entirely (including children) during extraction.
var stripTags = map[atom.Atom]bool{
	atom.Nav:      true,
	atom.Footer:   true,
	atom.Aside:    true,
	atom.Script:   true,
	atom.Style:    true,
	atom.Noscript: true,
	atom.Iframe:   true,
	atom.Svg:      true,
}

// extractText parses HTML and returns cleaned plain text.
func extractText(r io.Reader) (title string, body string, err error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", "", fmt.Errorf("parse HTML: %w", err)
	}

	title = findTitle(doc)
	content := findContentRoot(doc)
	if content == nil {
		content = doc
	}

	var buf strings.Builder
	renderText(content, &buf)
	return title, strings.TrimSpace(buf.String()), nil
}

// extractMarkdown parses HTML and returns simplified markdown.
func extractMarkdown(r io.Reader) (title string, body string, err error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", "", fmt.Errorf("parse HTML: %w", err)
	}

	title = findTitle(doc)
	content := findContentRoot(doc)
	if content == nil {
		content = doc
	}

	var buf strings.Builder
	renderMarkdown(content, &buf)
	return title, strings.TrimSpace(buf.String()), nil
}

// findTitle extracts the document title from <title> or first <h1>.
func findTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.DataAtom == atom.Title {
		return strings.TrimSpace(textContent(n))
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if t := findTitle(c); t != "" {
			return t
		}
	}
	return ""
}

// findContentRoot locates the best content root: <article>, <main>, or largest <div>.
func findContentRoot(doc *html.Node) *html.Node {
	// Priority 1: <article>
	if n := findFirst(doc, atom.Article); n != nil {
		return n
	}
	// Priority 2: <main>
	if n := findFirst(doc, atom.Main); n != nil {
		return n
	}
	// Priority 3: <div> with role="main"
	if n := findByAttr(doc, atom.Div, "role", "main"); n != nil {
		return n
	}
	// Priority 4: largest <div> by text content length
	return findLargestDiv(doc)
}

// findFirst returns the first element matching the given atom.
func findFirst(n *html.Node, a atom.Atom) *html.Node {
	if n.Type == html.ElementNode && n.DataAtom == a {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findFirst(c, a); found != nil {
			return found
		}
	}
	return nil
}

// findByAttr returns the first element matching atom with a specific attribute value.
func findByAttr(n *html.Node, a atom.Atom, key, value string) *html.Node {
	if n.Type == html.ElementNode && n.DataAtom == a {
		for _, attr := range n.Attr {
			if attr.Key == key && attr.Val == value {
				return n
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findByAttr(c, a, key, value); found != nil {
			return found
		}
	}
	return nil
}

// findLargestDiv finds the <div> with the most text content.
func findLargestDiv(n *html.Node) *html.Node {
	var best *html.Node
	var bestLen int
	findLargestDivHelper(n, &best, &bestLen)
	return best
}

func findLargestDivHelper(n *html.Node, best **html.Node, bestLen *int) {
	if n.Type == html.ElementNode && n.DataAtom == atom.Div {
		text := textContent(n)
		if l := len(text); l > *bestLen {
			*bestLen = l
			*best = n
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findLargestDivHelper(c, best, bestLen)
	}
}

// textContent returns all text content from a node (recursive).
func textContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var buf strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		buf.WriteString(textContent(c))
	}
	return buf.String()
}

// shouldStrip reports whether the node should be stripped.
func shouldStrip(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	return stripTags[n.DataAtom]
}

// renderText extracts clean text from a node tree.
func renderText(n *html.Node, buf *strings.Builder) {
	if shouldStrip(n) {
		return
	}
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			if buf.Len() > 0 {
				buf.WriteByte(' ')
			}
			buf.WriteString(text)
		}
		return
	}

	// Add paragraph/heading breaks.
	if n.Type == html.ElementNode {
		switch n.DataAtom {
		case atom.P, atom.Div, atom.Br, atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6, atom.Li, atom.Blockquote:
			if buf.Len() > 0 {
				buf.WriteString("\n\n")
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		renderText(c, buf)
	}
}

// renderMarkdown converts HTML to simplified markdown.
func renderMarkdown(n *html.Node, buf *strings.Builder) {
	if shouldStrip(n) {
		return
	}
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			if buf.Len() > 0 {
				last := buf.String()[buf.Len()-1]
				if last != '\n' && last != '#' && last != ' ' {
					buf.WriteByte(' ')
				}
			}
			buf.WriteString(text)
		}
		return
	}

	if n.Type == html.ElementNode {
		switch n.DataAtom {
		case atom.H1:
			if buf.Len() > 0 {
				buf.WriteString("\n\n")
			}
			buf.WriteString("# ")
			renderMarkdownChildren(n, buf)
			buf.WriteByte('\n')
			return
		case atom.H2:
			if buf.Len() > 0 {
				buf.WriteString("\n\n")
			}
			buf.WriteString("## ")
			renderMarkdownChildren(n, buf)
			buf.WriteByte('\n')
			return
		case atom.H3:
			if buf.Len() > 0 {
				buf.WriteString("\n\n")
			}
			buf.WriteString("### ")
			renderMarkdownChildren(n, buf)
			buf.WriteByte('\n')
			return
		case atom.H4, atom.H5, atom.H6:
			if buf.Len() > 0 {
				buf.WriteString("\n\n")
			}
			buf.WriteString("#### ")
			renderMarkdownChildren(n, buf)
			buf.WriteByte('\n')
			return
		case atom.A:
			href := attrVal(n, "href")
			text := strings.TrimSpace(textContent(n))
			if href != "" && text != "" {
				buf.WriteString(fmt.Sprintf("[%s](%s)", text, href))
			} else if text != "" {
				buf.WriteString(text)
			}
			return
		case atom.Li:
			if buf.Len() > 0 {
				buf.WriteByte('\n')
			}
			buf.WriteString("- ")
			renderMarkdownChildren(n, buf)
			return
		case atom.P, atom.Div:
			if buf.Len() > 0 {
				buf.WriteString("\n\n")
			}
			renderMarkdownChildren(n, buf)
			return
		case atom.Br:
			buf.WriteByte('\n')
			return
		case atom.Strong, atom.B:
			text := strings.TrimSpace(textContent(n))
			buf.WriteString("**" + text + "**")
			return
		case atom.Em, atom.I:
			text := strings.TrimSpace(textContent(n))
			buf.WriteString("*" + text + "*")
			return
		case atom.Code:
			buf.WriteString("`")
			renderMarkdownChildren(n, buf)
			buf.WriteString("`")
			return
		case atom.Pre:
			if buf.Len() > 0 {
				buf.WriteString("\n\n")
			}
			buf.WriteString("```\n")
			buf.WriteString(strings.TrimSpace(textContent(n)))
			buf.WriteString("\n```")
			return
		case atom.Blockquote:
			if buf.Len() > 0 {
				buf.WriteString("\n\n")
			}
			buf.WriteString("> ")
			renderMarkdownChildren(n, buf)
			return
		case atom.Ul, atom.Ol:
			if buf.Len() > 0 {
				buf.WriteByte('\n')
			}
			renderMarkdownChildren(n, buf)
			return
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		renderMarkdown(c, buf)
	}
}

// renderMarkdownChildren renders all children of a node.
func renderMarkdownChildren(n *html.Node, buf *strings.Builder) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		renderMarkdown(c, buf)
	}
}

// attrVal returns the value of the named attribute, or empty string.
func attrVal(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}
