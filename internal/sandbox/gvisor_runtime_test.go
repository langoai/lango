package sandbox

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGVisorRuntime_Stub(t *testing.T) {
	tests := []struct {
		give        string
		wantName    string
		wantAvail   bool
		wantRunErr  error
		wantCleanup bool // true means Cleanup should return nil
	}{
		{
			give:        "default stub runtime",
			wantName:    "gvisor",
			wantAvail:   false,
			wantRunErr:  ErrRuntimeUnavailable,
			wantCleanup: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			rt := NewGVisorRuntime()

			assert.Equal(t, tt.wantName, rt.Name())
			assert.Equal(t, tt.wantAvail, rt.IsAvailable(context.Background()))

			result, err := rt.Run(context.Background(), ContainerConfig{})
			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantRunErr)
			assert.Nil(t, result)

			if tt.wantCleanup {
				assert.NoError(t, rt.Cleanup(context.Background(), "any-id"))
			}
		})
	}
}

func TestGVisorRuntime_IsAvailable(t *testing.T) {
	tests := []struct {
		give string
		want bool
	}{
		{
			give: "background context",
			want: false,
		},
		{
			give: "cancelled context",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			rt := NewGVisorRuntime()

			var ctx context.Context
			switch tt.give {
			case "cancelled context":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			default:
				ctx = context.Background()
			}

			assert.Equal(t, tt.want, rt.IsAvailable(ctx))
		})
	}
}

func TestGVisorRuntime_Run(t *testing.T) {
	tests := []struct {
		give    ContainerConfig
		wantErr error
	}{
		{
			give:    ContainerConfig{},
			wantErr: ErrRuntimeUnavailable,
		},
		{
			give: ContainerConfig{
				Image:         "alpine:latest",
				ToolName:      "echo",
				NetworkMode:   "none",
				MemoryLimitMB: 128,
			},
			wantErr: ErrRuntimeUnavailable,
		},
	}

	for _, tt := range tests {
		name := "empty config"
		if tt.give.ToolName != "" {
			name = tt.give.ToolName
		}
		t.Run(name, func(t *testing.T) {
			rt := NewGVisorRuntime()

			result, err := rt.Run(context.Background(), tt.give)
			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Nil(t, result)
		})
	}
}

func TestGVisorRuntime_Name(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{
			give: "new instance",
			want: "gvisor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			rt := NewGVisorRuntime()
			assert.Equal(t, tt.want, rt.Name())
		})
	}
}

func TestGVisorRuntime_ImplementsContainerRuntime(t *testing.T) {
	// Verify the compile-time interface check by instantiating the type.
	var rt ContainerRuntime = NewGVisorRuntime()
	assert.NotNil(t, rt)
}
