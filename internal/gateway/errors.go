package gateway

import (
	"errors"
	"fmt"
)

var (
	ErrNoCompanion     = errors.New("no companion connected")
	ErrApprovalTimeout = errors.New("approval timeout")
	ErrAgentNotReady   = errors.New("agent not ready")
)

// Error implements the error interface for RPCError.
func (e RPCError) Error() string {
	return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message)
}
