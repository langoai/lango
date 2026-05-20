package approval

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestApprovalTargetContextRoundtripAndEmpty(t *testing.T) {
	if got := ApprovalTargetFromContext(context.Background()); got != "" {
		t.Fatalf("ApprovalTargetFromContext(empty) = %q, want empty string", got)
	}

	ctx := WithApprovalTarget(context.Background(), "telegram:chat:user")
	if got := ApprovalTargetFromContext(ctx); got != "telegram:chat:user" {
		t.Fatalf("ApprovalTargetFromContext() = %q, want %q", got, "telegram:chat:user")
	}
}

func TestApprovalErrorNilAndFallbackMessages(t *testing.T) {
	var appErr *Error
	if got := appErr.Error(); got != "" {
		t.Fatalf("nil Error.Error() = %q, want empty string", got)
	}
	if got := appErr.Unwrap(); got != nil {
		t.Fatalf("nil Error.Unwrap() = %v, want nil", got)
	}

	if got := (&Error{Kind: ErrTimeout}).Error(); got != ErrTimeout.Error() {
		t.Fatalf("kind Error.Error() = %q, want %q", got, ErrTimeout.Error())
	}
	if got := (&Error{}).Error(); got != "approval error" {
		t.Fatalf("empty Error.Error() = %q, want approval error", got)
	}
}

func TestWrapErrorMetadataAndSentinel(t *testing.T) {
	err := WrapError(ErrDenied, "telegram", "req-123", "custom denial")

	if !errors.Is(err, ErrDenied) {
		t.Fatalf("errors.Is(err, ErrDenied) = false, want true")
	}
	if got := err.Error(); got != "custom denial" {
		t.Fatalf("Error() = %q, want custom denial", got)
	}
	if got := ProviderFromError(err); got != "telegram" {
		t.Fatalf("ProviderFromError() = %q, want telegram", got)
	}
	if got := RequestIDFromError(err); got != "req-123" {
		t.Fatalf("RequestIDFromError() = %q, want req-123", got)
	}
}

func TestApprovalMetadataFromPlainError(t *testing.T) {
	err := errors.New("plain")

	if got := ProviderFromError(err); got != "" {
		t.Fatalf("ProviderFromError(plain) = %q, want empty string", got)
	}
	if got := RequestIDFromError(err); got != "" {
		t.Fatalf("RequestIDFromError(plain) = %q, want empty string", got)
	}
}

func TestFormatToolExecutionErrorVariants(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
		is   error
	}{
		{name: "denied", err: ErrDenied, want: "execution denied by user approval", is: ErrDenied},
		{name: "timeout", err: ErrTimeout, want: "approval expired", is: ErrTimeout},
		{name: "unavailable", err: ErrUnavailable, want: "no approval channel available", is: ErrUnavailable},
		{name: "other", err: errors.New("backend down"), want: "approval failed", is: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatToolExecutionError("filesystem.write", tt.err)
			if !strings.Contains(got.Error(), "filesystem.write") {
				t.Fatalf("error %q does not include tool name", got.Error())
			}
			if !strings.Contains(got.Error(), tt.want) {
				t.Fatalf("error %q does not contain %q", got.Error(), tt.want)
			}
			if tt.is != nil && !errors.Is(got, tt.is) {
				t.Fatalf("errors.Is(%v) = false, want true", tt.is)
			}
		})
	}
}
