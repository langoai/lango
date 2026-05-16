package lineio

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestReadLine_ReturnsFullLineWithNewline(t *testing.T) {
	line, err := ReadLine(bytes.NewBufferString("alpha\nbeta\n"))
	if err != nil {
		t.Fatalf("ReadLine returned error: %v", err)
	}
	if line != "alpha\n" {
		t.Fatalf("expected alpha\\n, got %q", line)
	}
}

func TestReadLine_ReturnsPartialLineOnEOF(t *testing.T) {
	line, err := ReadLine(bytes.NewBufferString("alpha"))
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF, got %v", err)
	}
	if line != "alpha" {
		t.Fatalf("expected partial line alpha, got %q", line)
	}
}
