package passphrase

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ReadStdinPipe reads a single line from non-terminal stdin.
// Returns an error if the line is empty after trimming.
func ReadStdinPipe() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("read stdin: %w", err)
	}

	passphrase := strings.TrimRight(line, "\n\r")
	if passphrase == "" {
		return "", fmt.Errorf("empty passphrase from stdin")
	}

	return passphrase, nil
}
