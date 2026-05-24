package passphrase

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/langoai/lango/internal/lineio"
)

// ReadStdinPipe reads a single line from non-terminal stdin.
// Returns an error if the line is empty after trimming.
func ReadStdinPipe() (string, error) {
	return ReadStdinPipeFromReader(os.Stdin)
}

// ReadStdinPipeFromReader reads a single line from the provided reader.
// Returns an error if the line is empty after trimming.
func ReadStdinPipeFromReader(in io.Reader) (string, error) {
	line, err := lineio.ReadLine(in)
	if err != nil && line == "" && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read stdin: %w", err)
	}

	passphrase := strings.TrimRight(line, "\n\r")
	if passphrase == "" {
		return "", fmt.Errorf("empty passphrase from stdin")
	}

	return passphrase, nil
}
