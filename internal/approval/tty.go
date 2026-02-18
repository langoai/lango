package approval

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// TTYProvider prompts the user via the terminal (stdin) for approval.
// CanHandle always returns false because TTY is a special fallback,
// not prefix-matched by session key.
type TTYProvider struct{}

var _ Provider = (*TTYProvider)(nil)

// RequestApproval prompts the user on stderr and reads y/a/N from stdin.
// "y" or "yes" approves once; "a" or "always" approves and grants persistent
// access for the tool in this session.
func (t *TTYProvider) RequestApproval(_ context.Context, req ApprovalRequest) (ApprovalResponse, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return ApprovalResponse{}, fmt.Errorf("TTY approval unavailable: stdin is not a terminal")
	}

	fmt.Fprintf(os.Stderr, "\n⚠ Sensitive tool '%s' requires approval.\n", req.ToolName)
	if req.Summary != "" {
		fmt.Fprintf(os.Stderr, "  %s\n", req.Summary)
	}
	fmt.Fprint(os.Stderr, "  Allow? [y/a/N] (a=always): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return ApprovalResponse{}, fmt.Errorf("read approval input: %w", err)
	}

	answer := strings.TrimSpace(strings.ToLower(input))
	switch answer {
	case "a", "always":
		return ApprovalResponse{Approved: true, AlwaysAllow: true}, nil
	case "y", "yes":
		return ApprovalResponse{Approved: true}, nil
	default:
		return ApprovalResponse{}, nil
	}
}

// CanHandle always returns false. TTY is used as a fallback only.
func (t *TTYProvider) CanHandle(_ string) bool {
	return false
}
