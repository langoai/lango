package exec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectOpaquePattern(t *testing.T) {
	t.Parallel()

	safeVars := map[string]struct{}{
		"HOME": {}, "PATH": {}, "USER": {}, "PWD": {},
		"SHELL": {}, "TERM": {}, "LANG": {}, "LC_ALL": {},
		"LC_CTYPE": {}, "TMPDIR": {},
	}

	tests := []struct {
		give string
		want ReasonCode
	}{
		// Command substitution: $(...)
		{give: "echo $(whoami)", want: ReasonCmdSubstitution},
		{give: "ls $(cat /etc/passwd)", want: ReasonCmdSubstitution},
		// Command substitution: backticks
		{give: "echo `whoami`", want: ReasonCmdSubstitution},
		{give: "ls `cat /etc/passwd`", want: ReasonCmdSubstitution},
		// Unsafe variable expansion: ${VAR}
		{give: "echo ${SECRET_TOKEN}", want: ReasonUnsafeVarExpand},
		{give: "echo ${API_KEY}", want: ReasonUnsafeVarExpand},
		// Unsafe variable expansion: $VAR
		{give: "echo $SECRET", want: ReasonUnsafeVarExpand},
		// Unsafe variable expansion with parameter operators
		{give: "echo ${UNKNOWN:-default}", want: ReasonUnsafeVarExpand},
		// Safe variable expansion: ${HOME}
		{give: "echo ${HOME}/bin", want: ReasonNone},
		{give: "echo $HOME/bin", want: ReasonNone},
		{give: "$PATH/mybin", want: ReasonNone},
		{give: "echo $USER", want: ReasonNone},
		{give: "cd $PWD", want: ReasonNone},
		{give: "echo $TMPDIR/cache", want: ReasonNone},
		// Eval verb
		{give: `eval "rm -rf /"`, want: ReasonEvalVerb},
		{give: `eval "echo hello"`, want: ReasonEvalVerb},
		// Encoded pipe: base64 decode to shell
		{give: "base64 -d payload | bash", want: ReasonEncodedPipe},
		{give: "base64 --decode payload | sh", want: ReasonEncodedPipe},
		{give: "cat file | base64 -d | eval", want: ReasonEncodedPipe},
		{give: "base64 -d payload | zsh", want: ReasonEncodedPipe},
		// Clean commands: no opaque patterns
		{give: "echo hello", want: ReasonNone},
		{give: "go build ./...", want: ReasonNone},
		{give: "ls -la", want: ReasonNone},
		{give: "grep -r pattern src/", want: ReasonNone},
		{give: "cat /etc/hosts", want: ReasonNone},
		// Edge: special variables are not flagged
		{give: "echo $? $$ $!", want: ReasonNone},
		// Edge: empty string
		{give: "", want: ReasonNone},
		// base64 without pipe (not encoded pipe)
		{give: "base64 -d payload > file.txt", want: ReasonNone},
		// base64 without decode flag (not encoded pipe)
		{give: "base64 file.txt | bash", want: ReasonNone},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := detectOpaquePattern(tt.give, safeVars)
			assert.Equal(t, tt.want, got, "detectOpaquePattern(%q)", tt.give)
		})
	}
}

func TestExtractSingleCommandFromConstruct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give           string
		wantInner      string
		wantReason     ReasonCode
		wantExtractable bool
	}{
		// Subshell with single command → extractable.
		{
			give:            "(kill 1)",
			wantInner:       "kill 1",
			wantReason:      ReasonGroupedSubshell,
			wantExtractable: true,
		},
		{
			give:            "(echo hello)",
			wantInner:       "echo hello",
			wantReason:      ReasonGroupedSubshell,
			wantExtractable: true,
		},
		// Block with single command → extractable.
		{
			give:            "{ lango security; }",
			wantInner:       "lango security",
			wantReason:      ReasonGroupedSubshell,
			wantExtractable: true,
		},
		{
			give:            "{ echo hello; }",
			wantInner:       "echo hello",
			wantReason:      ReasonGroupedSubshell,
			wantExtractable: true,
		},
		// Block with rm -rf / → extractable.
		{
			give:            "{ rm -rf /; }",
			wantInner:       "rm -rf /",
			wantReason:      ReasonGroupedSubshell,
			wantExtractable: true,
		},
		// Multi-statement subshell → not extractable.
		{
			give:            "(kill 1; echo done)",
			wantInner:       "",
			wantReason:      ReasonGroupedSubshell,
			wantExtractable: false,
		},
		// Multi-statement block → not extractable.
		{
			give:            "{ echo hello; echo world; }",
			wantInner:       "",
			wantReason:      ReasonGroupedSubshell,
			wantExtractable: false,
		},
		// FuncDecl with single command body (standalone decl, no invocation) → extractable.
		{
			give:            "f() { kill 1; }",
			wantInner:       "kill 1",
			wantReason:      ReasonShellFunction,
			wantExtractable: true,
		},
		// FuncDecl + invocation = 2 top-level stmts → not extractable.
		{
			give:            "f() { kill 1; }; f",
			wantInner:       "",
			wantReason:      ReasonNone,
			wantExtractable: false,
		},
		// Plain command → not a construct, not extractable.
		{
			give:            "echo hello",
			wantInner:       "",
			wantReason:      ReasonNone,
			wantExtractable: false,
		},
		// Empty → not extractable.
		{
			give:            "",
			wantInner:       "",
			wantReason:      ReasonNone,
			wantExtractable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			inner, reason, extractable := extractSingleCommandFromConstruct(tt.give)
			assert.Equal(t, tt.wantExtractable, extractable, "extractable for %q", tt.give)
			assert.Equal(t, tt.wantReason, reason, "reason for %q", tt.give)
			if extractable {
				assert.Equal(t, tt.wantInner, inner, "inner for %q", tt.give)
			}
		})
	}
}

func TestDetectShellConstruct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want ReasonCode
	}{
		// Heredoc: << operator
		{give: "cat << 'EOF'\nhello\nEOF", want: ReasonHeredoc},
		{give: "sh << 'EOF'\necho hello\nEOF", want: ReasonHeredoc},
		// Heredoc: <<- (dash heredoc, strips leading tabs)
		{give: "cat <<- EOF\n\thello\n\tEOF", want: ReasonHeredoc},
		// Here-string: <<<
		{give: `cat <<< "hello world"`, want: ReasonHeredoc},
		// Process substitution: <(cmd)
		{give: "diff <(ls dir1) <(ls dir2)", want: ReasonProcessSubst},
		{give: "cat <(echo hello)", want: ReasonProcessSubst},
		// Process substitution: >(cmd)
		{give: "tee >(grep error > errors.log)", want: ReasonProcessSubst},
		// Grouped subshell: (cmd; cmd)
		{give: "(echo hello; echo world)", want: ReasonGroupedSubshell},
		{give: "(cd /tmp; ls)", want: ReasonGroupedSubshell},
		// Brace group: { cmd; }
		{give: "{ echo hello; echo world; }", want: ReasonGroupedSubshell},
		// Shell function definition
		{give: "f() { echo hello; }; f", want: ReasonShellFunction},
		{give: "myfunc() { rm -rf /; }; myfunc", want: ReasonShellFunction},
		// Clean commands: no shell constructs
		{give: "echo hello", want: ReasonNone},
		{give: "ls -la", want: ReasonNone},
		{give: "go build ./...", want: ReasonNone},
		{give: "grep -r pattern src/", want: ReasonNone},
		// Edge: empty string
		{give: "", want: ReasonNone},
		// Edge: single command with pipe (not a grouped subshell)
		{give: "ls | grep test", want: ReasonNone},
		// Edge: find without -exec (not a construct)
		{give: "find . -name '*.go'", want: ReasonNone},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := detectShellConstruct(tt.give)
			assert.Equal(t, tt.want, got, "detectShellConstruct(%q)", tt.give)
		})
	}
}
