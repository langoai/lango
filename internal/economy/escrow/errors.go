package escrow

import "errors"

var (
	ErrNotFunded    = errors.New("escrow not funded")
	ErrInvalidStatus = errors.New("invalid escrow status for operation")
)
