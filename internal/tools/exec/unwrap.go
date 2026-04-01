package exec

import (
	"bytes"
	"path/filepath"
	"strings"
	"unicode"

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

	// Check for env wrapper: strip "env" or "/usr/bin/env" prefix and skip
	// env-specific flags (-i, -0, -u NAME, -C DIR), variable assignments
	// (NAME=value), and the -- terminator to find the actual command verb.
	verb := wordLitLower(args[0])
	if isEnvVerb(verb) {
		args = skipEnvArgs(args)
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

// skipEnvArgs skips env-specific arguments (flags, flag arguments, variable
// assignments, and the -- terminator) and returns the remaining args starting
// from the actual command verb.
//
// env command syntax (POSIX + GNU coreutils):
//
//	env [-i] [-0] [-u name] [-C dir] [-S string] [--] [name=value]... [command [args...]]
func skipEnvArgs(args []*syntax.Word) []*syntax.Word {
	i := 1 // skip 'env' itself
loop:
	for i < len(args) {
		lit := args[i].Lit()
		switch {
		case lit == "--":
			i++
			break loop // next is command
		case lit == "-i" || lit == "-0":
			i++ // standalone flag
		case lit == "-u" || lit == "-C" || lit == "-S":
			i += 2 // flag + its argument (-u NAME, -C DIR, -S STRING)
		case strings.HasPrefix(lit, "-") && len(lit) > 1:
			i++ // other unknown flags
		case looksLikeEnvAssignment(lit):
			i++ // NAME=value
		default:
			break loop // found command verb
		}
	}
	if i >= len(args) {
		return nil
	}
	return args[i:]
}

// looksLikeEnvAssignment returns true if s matches the pattern NAME=value where
// NAME is a valid shell variable name (first char letter/_, rest alnum/_).
// This prevents paths like ./foo=bar or flags like --flag=val from being
// misidentified as env variable assignments.
func looksLikeEnvAssignment(s string) bool {
	idx := strings.IndexByte(s, '=')
	if idx <= 0 {
		return false
	}
	name := s[:idx]
	for i, r := range name {
		if i == 0 && !unicode.IsLetter(r) && r != '_' {
			return false
		}
		if i > 0 && !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
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

// extractXargsVerb extracts the inner command verb from an xargs invocation.
// Pattern: xargs [-flags] [-I repl] [-n num] [--] cmd [args...]
// Returns the extracted verb and true if found, or ("", false) if extraction fails.
func extractXargsVerb(cmd string) (string, bool) {
	fields := strings.Fields(strings.TrimSpace(cmd))
	if len(fields) == 0 {
		return "", false
	}

	// First field must be "xargs".
	if strings.ToLower(filepath.Base(fields[0])) != "xargs" {
		return "", false
	}

	// Skip xargs flags to find the inner command verb.
	// xargs flags that take an argument: -I, -L, -n, -P, -s, -E, -R
	flagsWithArg := map[string]struct{}{
		"-I": {}, "-L": {}, "-n": {}, "-P": {}, "-s": {}, "-E": {}, "-R": {},
		"--max-args": {}, "--max-procs": {}, "--replace": {},
	}

	i := 1 // skip "xargs"
	for i < len(fields) {
		f := fields[i]
		if f == "--" {
			i++
			break
		}
		if !strings.HasPrefix(f, "-") {
			break // found the command verb
		}
		// Check if this flag takes an argument.
		if _, hasArg := flagsWithArg[f]; hasArg {
			i += 2 // skip flag + its argument
		} else {
			i++ // standalone flag (e.g., -r, -0, -t, -p, --no-run-if-empty)
		}
	}

	if i >= len(fields) {
		return "", false
	}

	verb := filepath.Base(fields[i])
	return strings.ToLower(verb), true
}

// extractFindExecVerb extracts the inner command verb from a find -exec invocation.
// Pattern: find [path...] -exec cmd {} \; or find [path...] -exec cmd {} +
// Also supports -execdir variant.
// Returns the extracted verb and true if found, or ("", false) if extraction fails.
func extractFindExecVerb(cmd string) (string, bool) {
	fields := strings.Fields(strings.TrimSpace(cmd))
	if len(fields) < 3 {
		return "", false
	}

	// First field must be "find".
	if strings.ToLower(filepath.Base(fields[0])) != "find" {
		return "", false
	}

	// Scan for -exec or -execdir flag.
	for i := 1; i < len(fields); i++ {
		if fields[i] == "-exec" || fields[i] == "-execdir" {
			// The next field is the command verb.
			if i+1 < len(fields) {
				verb := filepath.Base(fields[i+1])
				return strings.ToLower(verb), true
			}
			return "", false
		}
	}
	return "", false
}

// unwrapEnvPrefix strips leading VAR=val assignments from a command string
// (bare env prefix without explicit "env" command) and returns the remaining
// command. Uses AST parsing to detect CallExpr.Assigns.
//
// Pattern: VAR=val [VAR2=val2 ...] cmd [args...]
// Returns the command after stripping assignments, and true if assignments were found.
func unwrapEnvPrefix(cmd string) (string, bool) {
	trimmed := strings.TrimSpace(cmd)
	if trimmed == "" {
		return cmd, false
	}

	parser := syntax.NewParser()
	f, err := parser.Parse(strings.NewReader(trimmed), "")
	if err != nil {
		return cmd, false
	}

	if len(f.Stmts) != 1 {
		return cmd, false
	}

	call, ok := f.Stmts[0].Cmd.(*syntax.CallExpr)
	if !ok {
		return cmd, false
	}

	// If there are assignments and args, the assignments are env prefixes.
	if len(call.Assigns) > 0 && len(call.Args) > 0 {
		// Reconstruct the command from Args only (without assignments).
		var parts []string
		printer := syntax.NewPrinter()
		for _, arg := range call.Args {
			var buf bytes.Buffer
			if err := printer.Print(&buf, arg); err != nil {
				return cmd, false
			}
			parts = append(parts, buf.String())
		}
		return strings.Join(parts, " "), true
	}

	return cmd, false
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
