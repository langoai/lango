package lineio

import (
	"bufio"
	"io"
)

// ReadLine reads a single raw line from the supplied reader.
// It preserves the behavior of bufio.Reader.ReadString('\n'), including
// returning a partial line together with an EOF error.
func ReadLine(in io.Reader) (string, error) {
	reader := bufio.NewReader(in)
	return reader.ReadString('\n')
}
