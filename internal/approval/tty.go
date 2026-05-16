package approval

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/langoai/lango/internal/lineio"
)

// TTYProvider prompts the user via the terminal (stdin) for approval.
// CanHandle always returns false because TTY is a special fallback,
// not prefix-matched by session key.
type TTYProvider struct{}

var _ Provider = (*TTYProvider)(nil)

var (
	ttyIsTerminal           = func() bool { return term.IsTerminal(int(os.Stdin.Fd())) }
	ttyInput      io.Reader = os.Stdin
	ttyError      io.Writer = os.Stderr
)

// RequestApproval prompts the user on stderr and reads y/a/N from stdin.
// "y" or "yes" approves once; "a" or "always" approves and grants persistent
// access for the tool in this session.
func (t *TTYProvider) RequestApproval(_ context.Context, req ApprovalRequest) (ApprovalResponse, error) {
	if !ttyIsTerminal() {
		return ApprovalResponse{}, WrapError(
			ErrUnavailable,
			"tty",
			req.ID,
			"TTY approval unavailable: stdin is not a terminal",
		)
	}

	fmt.Fprintf(ttyError, "\n⚠ Sensitive tool '%s' requires approval.\n", req.ToolName)
	if req.Summary != "" {
		fmt.Fprintf(ttyError, "  %s\n", req.Summary)
	}
	fmt.Fprint(ttyError, "  Allow? [y/a/N] (a=always): ")

	input, err := lineio.ReadLine(ttyInput)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return ApprovalResponse{Provider: "tty"}, nil
		}
		return ApprovalResponse{}, fmt.Errorf("read approval input: %w", err)
	}

	answer := strings.TrimSpace(strings.ToLower(input))
	switch answer {
	case "a", "always":
		return ApprovalResponse{Approved: true, AlwaysAllow: true, Provider: "tty"}, nil
	case "y", "yes":
		return ApprovalResponse{Approved: true, Provider: "tty"}, nil
	default:
		return ApprovalResponse{Provider: "tty"}, nil
	}
}

// CanHandle always returns false. TTY is used as a fallback only.
func (t *TTYProvider) CanHandle(_ string) bool {
	return false
}

func (t *TTYProvider) Name() string {
	return "tty"
}
