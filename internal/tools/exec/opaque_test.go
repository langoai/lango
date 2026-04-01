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
