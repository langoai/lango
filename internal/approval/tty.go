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

// RequestApproval prompts the user on stderr and reads y/N from stdin.
func (t *TTYProvider) RequestApproval(_ context.Context, req ApprovalRequest) (bool, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false, nil
	}

	fmt.Fprintf(os.Stderr, "\n⚠ Sensitive tool '%s' requires approval.\n", req.ToolName)
	if req.Summary != "" {
		fmt.Fprintf(os.Stderr, "  %s\n", req.Summary)
	}
	fmt.Fprint(os.Stderr, "  Allow? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("read approval input: %w", err)
	}

	answer := strings.TrimSpace(strings.ToLower(input))
	return answer == "y" || answer == "yes", nil
}

// CanHandle always returns false. TTY is used as a fallback only.
func (t *TTYProvider) CanHandle(_ string) bool {
	return false
}
