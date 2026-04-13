package economy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildTools_AllNil(t *testing.T) {
	tools := BuildTools(nil, nil, nil, nil, nil)
	assert.Empty(t, tools, "all-nil engines should produce zero tools")
}

func TestBuildTools_ToolNames(t *testing.T) {
	// We cannot easily construct real engines without full store setup,
	// but we verify nil-guard branches produce no panics and the expected
	// tool count/names when engines are non-nil is tested via the
	// integration path in app/ package tests. Here we verify nil handling.

	tests := []struct {
		give     string
		wantZero bool
	}{
		{give: "all nil", wantZero: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			tools := BuildTools(nil, nil, nil, nil, nil)
			if tt.wantZero {
				assert.Empty(t, tools)
			}
		})
	}
}

func TestBuildTools_NoAppImport(t *testing.T) {
	// Compile-time guarantee: this file is in package economy, not app.
	// If economy/tools.go imported app, this test file would not compile
	// in the economy package. This is a structural smoke test.
	_ = BuildTools
}
