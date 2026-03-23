package approval

import (
	"errors"
	"fmt"
)

var (
	ErrDenied      = errors.New("approval denied")
	ErrTimeout     = errors.New("approval timeout")
	ErrUnavailable = errors.New("approval unavailable")
)

// Error carries structured approval failure metadata while preserving a sentinel cause.
type Error struct {
	Kind      error
	Provider  string
	RequestID string
	Message   string
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Kind != nil {
		return e.Kind.Error()
	}
	return "approval error"
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Kind
}

func WrapError(kind error, provider, requestID, message string) error {
	return &Error{
		Kind:      kind,
		Provider:  provider,
		RequestID: requestID,
		Message:   message,
	}
}

func ProviderFromError(err error) string {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Provider
	}
	return ""
}

func RequestIDFromError(err error) string {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.RequestID
	}
	return ""
}

func FormatToolExecutionError(toolName string, kind error) error {
	switch {
	case errors.Is(kind, ErrDenied):
		return fmt.Errorf("tool '%s' execution denied by user approval: %w", toolName, ErrDenied)
	case errors.Is(kind, ErrTimeout):
		return fmt.Errorf("tool '%s' approval expired: %w", toolName, ErrTimeout)
	case errors.Is(kind, ErrUnavailable):
		return fmt.Errorf("tool '%s' execution denied: no approval channel available: %w", toolName, ErrUnavailable)
	default:
		return fmt.Errorf("tool '%s' approval failed: %w", toolName, kind)
	}
}
