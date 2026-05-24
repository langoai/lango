//go:build darwin && cgo

package keyring

import "testing"

func TestOSStatusDescription(t *testing.T) {
	tests := []struct {
		code int
		want string
	}{
		{code: -34018, want: "errSecMissingEntitlement: binary needs Apple Developer signing"},
		{code: -25308, want: "errSecInteractionNotAllowed: cannot present Touch ID UI"},
		{code: -128, want: "errSecUserCanceled: user cancelled biometric prompt"},
		{code: -25293, want: "errSecAuthFailed: authentication failed or biometric enrollment changed"},
		{code: -25300, want: "errSecItemNotFound: item not found"},
		{code: -25291, want: "errSecInvalidOwnerEdit: device passcode may not be set"},
		{code: 12345, want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := osStatusDescription(tt.code); got != tt.want {
				t.Fatalf("osStatusDescription(%d) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}
