//go:build !vec

package embedding

import (
	"database/sql"
	"errors"
)

// ErrVecNotCompiled indicates sqlite-vec support was not compiled into this binary.
var ErrVecNotCompiled = errors.New("sqlite-vec support not compiled: rebuild with -tags vec")

// NewVectorStore returns an error when sqlite-vec is not compiled.
func NewVectorStore(_ *sql.DB, _ int) (VectorStore, error) {
	return nil, ErrVecNotCompiled
}
