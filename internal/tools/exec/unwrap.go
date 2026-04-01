package exec

import (
	"bytes"
	"path/filepath"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// maxUnwrapDepth is the maximum recursion depth for nested shell wrapper unwrap.
const maxUnwrapDepth = 5

// shellWrappers are command verbs that invoke a shell with -c flag.
var shellWrappers = map[string]struct{}{
	"sh":   {},
	"bash": {},
	"zsh":  {},
	"dash": {},
}

// isShellWrapper returns true if the verb (after path stripping and lowercasing)
// is a known shell wrapper binary.
func isShellWrapper(verb string) bool {
	_, ok := shellWrappers[verb]
	return ok
}

// unwrapShellWrapper detects and unwraps shell wrapper commands using AST-based
// parsing. Returns the innermost command and true if unwrapped, or the original
// command and false.
//
// Supported patterns:
//   - sh -c "cmd", bash -c 'cmd', /bin/sh -c cmd, /usr/bin/bash -c "cmd"
//   - zsh -c, dash -c
//   - sh -lc "cmd", sh -ic "cmd" (login/interactive shell with -c)
//   - /usr/bin/env sh -c "cmd", env bash -c "cmd" (env wrapper)
//   - Nested: sh -c "bash -c \"inner\"" (recursive unwrap, depth limit 5)
//
// Falls back to string-based parser on AST parse failure.
func unwrapShellWrapper(cmd string) (inner string, unwrapped bool) {
	trimmed := strings.TrimSpace(cmd)
	if trimmed == "" {
		return cmd, false
	}

	// Try AST-based parsing first.
	parser := syntax.NewParser()
	f, err := parser.Parse(strings.NewReader(trimmed), "")
	if err == nil {
		if result, ok := unwrapShellWrapperAST(f, 0); ok {
			return result, true
		}
		// AST parsed but no shell wrapper found — return original.
		return cmd, false
	}

	// Fallback: string-based parser for commands that fail AST parsing.
	return unwrapShellWrapperString(cmd)
}

// unwrapShellWrapperAST walks the AST to find and unwrap shell wrapper patterns.
// It handles recursive unwrap up to maxUnwrapDepth levels.
func unwrapShellWrapperAST(f *syntax.File, depth int) (string, bool) {
	if depth >= maxUnwrapDepth {
		return "", false
	}

	// Expect exactly one statement with a simple CallExpr.
	if len(f.Stmts) != 1 {
		return "", false
	}
	stmt := f.Stmts[0]
	call, ok := stmt.Cmd.(*syntax.CallExpr)
	if !ok || len(call.Args) < 1 {
		return "", false
	}

	args := call.Args

	// Check for env wrapper: strip "env" or "/usr/bin/env" prefix.
	verb := wordLitLower(args[0])
	if isEnvVerb(verb) {
		args = args[1:]
		if len(args) < 1 {
			return "", false
		}
		verb = wordLitLower(args[0])
	}

	// Check if the verb is a shell wrapper.
	if !isShellWrapper(verb) {
		return "", false
	}

	// Find -c flag and extract the inner command.
	innerCmd, found := extractInnerFromArgs(args[1:])
	if !found {
		return "", false
	}

	// Try recursive unwrap on the inner command.
	parser := syntax.NewParser()
	innerFile, err := parser.Parse(strings.NewReader(innerCmd), "")
	if err == nil {
		if deeper, ok := unwrapShellWrapperAST(innerFile, depth+1); ok {
			return deeper, true
		}
	}

	return innerCmd, true
}

// extractInnerFromArgs processes arguments after the shell verb to find -c flag
// and extract the inner command string. Supports:
//   - "-c" as standalone flag followed by command argument
//   - Combined flags like "-lc", "-ic" where c is the last character
func extractInnerFromArgs(args []*syntax.Word) (string, bool) {
	for i, arg := range args {
		lit := arg.Lit()
		if lit == "" {
			continue
		}

		// Check for standalone "-c" flag.
		if lit == "-c" {
			return extractCommandArg(args, i+1)
		}

		// Check for combined flags like "-lc", "-ic" where c is the last char.
		// The -c flag takes an argument, so it must be last in combined flags.
		if strings.HasPrefix(lit, "-") && len(lit) > 2 && lit[len(lit)-1] == 'c' && !strings.HasPrefix(lit, "--") {
			return extractCommandArg(args, i+1)
		}
	}
	return "", false
}

// extractCommandArg extracts the command string from the argument at position
// cmdIdx in the args slice. For quoted arguments, it extracts the content.
// For unquoted arguments, it extracts the literal text of the first word only.
func extractCommandArg(args []*syntax.Word, cmdIdx int) (string, bool) {
	if cmdIdx >= len(args) {
		return "", false
	}

	word := args[cmdIdx]

	// Use the printer to get the content as the shell would interpret it.
	// For a quoted word like "kill 1234", the AST captures it properly.
	inner := wordContent(word)
	if inner == "" {
		return "", false
	}

	return inner, true
}

// wordContent extracts the interpreted content of a Word node.
// For quoted words, it returns the content inside the quotes.
// For unquoted literal words, it returns the literal value.
func wordContent(w *syntax.Word) string {
	if len(w.Parts) == 0 {
		return ""
	}

	// Single part: extract directly for best accuracy.
	if len(w.Parts) == 1 {
		switch p := w.Parts[0].(type) {
		case *syntax.SglQuoted:
			return p.Value
		case *syntax.DblQuoted:
			return dblQuotedContent(p)
		case *syntax.Lit:
			return p.Value
		}
	}

	// Multiple parts: use the printer for accurate reconstruction.
	var buf bytes.Buffer
	printer := syntax.NewPrinter()
	if err := printer.Print(&buf, w); err != nil {
		return ""
	}
	// Strip outer quotes if present from printed output.
	return stripQuotes(buf.String())
}

// dblQuotedContent extracts the text content from a DblQuoted node.
// It unescapes shell escape sequences within double quotes (\" -> ", \\ -> \, etc.)
// so that the result can be re-parsed as a shell command for recursive unwrap.
func dblQuotedContent(dq *syntax.DblQuoted) string {
	var buf bytes.Buffer
	printer := syntax.NewPrinter()
	for _, part := range dq.Parts {
		switch p := part.(type) {
		case *syntax.Lit:
			buf.WriteString(unescapeDblQuoted(p.Value))
		default:
			// For non-literal parts (expansions etc.), use printer.
			if err := printer.Print(&buf, p); err != nil {
				return ""
			}
		}
	}
	return buf.String()
}

// unescapeDblQuoted removes POSIX double-quote escape sequences.
// Inside double quotes, \", \\, \$, \`, and \newline are escape sequences.
func unescapeDblQuoted(s string) string {
	if !strings.ContainsRune(s, '\\') {
		return s
	}
	var buf strings.Builder
	buf.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			next := s[i+1]
			switch next {
			case '"', '\\', '$', '`':
				buf.WriteByte(next)
				i++ // skip the escaped character
			default:
				buf.WriteByte(s[i])
			}
		} else {
			buf.WriteByte(s[i])
		}
	}
	return buf.String()
}

// wordLitLower returns the lowercased base name of a word's literal value.
// Used to identify command verbs like "sh", "bash", "/usr/bin/env".
func wordLitLower(w *syntax.Word) string {
	lit := w.Lit()
	if lit == "" {
		return ""
	}
	return strings.ToLower(filepath.Base(lit))
}

// isEnvVerb returns true if the verb is "env" (the env command used to wrap
// shell invocations like /usr/bin/env sh -c "cmd").
func isEnvVerb(verb string) bool {
	return verb == "env"
}

// unwrapShellWrapperString is the original string-based unwrap implementation.
// Used as fallback when AST parsing fails.
func unwrapShellWrapperString(cmd string) (inner string, unwrapped bool) {
	trimmed := strings.TrimSpace(cmd)
	if trimmed == "" {
		return cmd, false
	}

	// Split into whitespace-separated fields for robust parsing.
	fields := strings.Fields(trimmed)
	if len(fields) < 3 {
		// Need at least: <shell> -c <cmd>
		return cmd, false
	}

	// Check if the first field (verb) is a shell wrapper.
	verb := strings.ToLower(filepath.Base(fields[0]))
	if !isShellWrapper(verb) {
		return cmd, false
	}

	// Second field must be exactly "-c".
	if fields[1] != "-c" {
		return cmd, false
	}

	// Extract only the command_string (first argument after -c).
	// POSIX: sh -c command_string [command_name [argument...]]
	// Only command_string is executed; remaining args are positional parameters.

	// Find the position of "-c" after the first word.
	shellEnd := strings.IndexFunc(trimmed, func(r rune) bool { return r == ' ' || r == '\t' })
	if shellEnd < 0 {
		return cmd, false
	}
	flagIdx := strings.Index(trimmed[shellEnd:], "-c")
	if flagIdx < 0 {
		return cmd, false
	}
	flagIdx += shellEnd // adjust to absolute position
	afterFlag := strings.TrimSpace(trimmed[flagIdx+2:])
	if afterFlag == "" {
		return cmd, false
	}

	// If the command_string is quoted, extract just the quoted portion.
	if afterFlag[0] == '"' || afterFlag[0] == '\'' {
		quote := afterFlag[0]
		end := strings.IndexByte(afterFlag[1:], quote)
		if end < 0 {
			// Unmatched quote — cannot determine command boundary.
			return cmd, false
		}
		inner = afterFlag[1 : end+1] // content inside quotes
		if inner == "" {
			return cmd, false
		}
		return inner, true
	}

	// Unquoted: first whitespace-delimited token is the command_string.
	if spaceIdx := strings.IndexFunc(afterFlag, func(r rune) bool { return r == ' ' || r == '\t' }); spaceIdx >= 0 {
		inner = afterFlag[:spaceIdx]
	} else {
		inner = afterFlag
	}
	if inner == "" {
		return cmd, false
	}
	return inner, true
}

// stripQuotes removes matching outer single or double quotes from a string.
// Only strips if the first and last characters are the same quote character.
func stripQuotes(s string) string {
	if len(s) < 2 {
		return s
	}
	first := s[0]
	last := s[len(s)-1]
	if (first == '"' || first == '\'') && first == last {
		return s[1 : len(s)-1]
	}
	return s
}
