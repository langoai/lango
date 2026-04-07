package os

import (
	"context"
	"os/exec"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackendMode_String(t *testing.T) {
	tests := []struct {
		give BackendMode
		want string
	}{
		{give: BackendAuto, want: "auto"},
		{give: BackendSeatbelt, want: "seatbelt"},
		{give: BackendBwrap, want: "bwrap"},
		{give: BackendNative, want: "native"},
		{give: BackendNone, want: "none"},
		{give: BackendMode(99), want: "BackendMode(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.give.String())
		})
	}
}

func TestParseBackendMode(t *testing.T) {
	tests := []struct {
		give    string
		want    BackendMode
		wantErr bool
	}{
		{give: "", want: BackendAuto},
		{give: "auto", want: BackendAuto},
		{give: "AUTO", want: BackendAuto},
		{give: "  auto  ", want: BackendAuto},
		{give: "seatbelt", want: BackendSeatbelt},
		{give: "Seatbelt", want: BackendSeatbelt},
		{give: "bwrap", want: BackendBwrap},
		{give: "BWRAP", want: BackendBwrap},
		{give: "native", want: BackendNative},
		{give: "none", want: BackendNone},
		{give: "NONE", want: BackendNone},
		{give: "unknown", wantErr: true},
		{give: "docker", wantErr: true},
		{give: "landlock", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, err := ParseBackendMode(tt.give)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "unknown sandbox backend")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// fakeIsolator is a test helper that implements OSIsolator with configurable behavior.
type fakeIsolator struct {
	name      string
	available bool
	reason    string
}

func (f *fakeIsolator) Apply(_ context.Context, _ *exec.Cmd, _ Policy) error {
	if !f.available {
		return ErrIsolatorUnavailable
	}
	return nil
}

func (f *fakeIsolator) Available() bool { return f.available }
func (f *fakeIsolator) Name() string    { return f.name }

func (f *fakeIsolator) Reason() string {
	if f.available {
		return ""
	}
	return f.reason
}

func TestSelectBackend(t *testing.T) {
	availableSeatbelt := &fakeIsolator{name: "seatbelt", available: true}
	unavailableSeatbelt := &fakeIsolator{name: "seatbelt", available: false, reason: "sandbox-exec not found"}
	unavailableBwrap := &fakeIsolator{name: "bwrap", available: false, reason: "bwrap not installed"}
	availableBwrap := &fakeIsolator{name: "bwrap", available: true}
	unavailableNative := &fakeIsolator{name: "native", available: false, reason: "kernel too old"}

	tests := []struct {
		give           string
		giveMode       BackendMode
		giveCandidates []BackendCandidate
		wantName       string
		wantAvailable  bool
		wantMode       BackendMode
		wantIsNoop     bool
	}{
		{
			give:     "auto with mixed candidates picks first available",
			giveMode: BackendAuto,
			giveCandidates: []BackendCandidate{
				{Mode: BackendSeatbelt, Isolator: unavailableSeatbelt},
				{Mode: BackendBwrap, Isolator: availableBwrap},
				{Mode: BackendNative, Isolator: unavailableNative},
			},
			wantName:      "bwrap",
			wantAvailable: true,
			wantMode:      BackendBwrap,
		},
		{
			give:     "auto with all unavailable returns noop",
			giveMode: BackendAuto,
			giveCandidates: []BackendCandidate{
				{Mode: BackendSeatbelt, Isolator: unavailableSeatbelt},
				{Mode: BackendBwrap, Isolator: unavailableBwrap},
				{Mode: BackendNative, Isolator: unavailableNative},
			},
			wantName:      "noop",
			wantAvailable: false,
			wantMode:      BackendAuto,
			wantIsNoop:    true,
		},
		{
			give:     "auto with empty candidates returns noop",
			giveMode: BackendAuto,
			wantName:      "noop",
			wantAvailable: false,
			wantMode:      BackendAuto,
			wantIsNoop:    true,
		},
		{
			give:     "explicit seatbelt with available returns it",
			giveMode: BackendSeatbelt,
			giveCandidates: []BackendCandidate{
				{Mode: BackendSeatbelt, Isolator: availableSeatbelt},
				{Mode: BackendBwrap, Isolator: availableBwrap},
			},
			wantName:      "seatbelt",
			wantAvailable: true,
			wantMode:      BackendSeatbelt,
		},
		{
			give:     "explicit bwrap with unavailable returns bwrap stub not noop",
			giveMode: BackendBwrap,
			giveCandidates: []BackendCandidate{
				{Mode: BackendSeatbelt, Isolator: availableSeatbelt},
				{Mode: BackendBwrap, Isolator: unavailableBwrap},
			},
			wantName:      "bwrap",
			wantAvailable: false,
			wantMode:      BackendBwrap,
		},
		{
			give:     "explicit mode not in candidates returns noop",
			giveMode: BackendNative,
			giveCandidates: []BackendCandidate{
				{Mode: BackendSeatbelt, Isolator: availableSeatbelt},
			},
			wantName:      "noop",
			wantAvailable: false,
			wantMode:      BackendNative,
			wantIsNoop:    true,
		},
		{
			give:          "none always returns noop",
			giveMode:      BackendNone,
			giveCandidates: []BackendCandidate{
				{Mode: BackendSeatbelt, Isolator: availableSeatbelt},
				{Mode: BackendBwrap, Isolator: availableBwrap},
			},
			wantName:      "noop",
			wantAvailable: false,
			wantMode:      BackendNone,
			wantIsNoop:    true,
		},
		{
			give:          "none with empty candidates returns noop",
			giveMode:      BackendNone,
			wantName:      "noop",
			wantAvailable: false,
			wantMode:      BackendNone,
			wantIsNoop:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			iso, info := SelectBackend(tt.giveMode, tt.giveCandidates)

			assert.Equal(t, tt.wantName, info.Name)
			assert.Equal(t, tt.wantAvailable, info.Available)
			assert.Equal(t, tt.wantMode, info.Mode)

			// Verify the isolator is actually usable.
			assert.Equal(t, tt.wantName, iso.Name())
			assert.Equal(t, tt.wantAvailable, iso.Available())

			if tt.wantIsNoop {
				_, ok := iso.(*noopIsolator)
				assert.True(t, ok, "expected noopIsolator, got %T", iso)
			}
		})
	}
}

func TestSelectBackend_NoneReason(t *testing.T) {
	iso, info := SelectBackend(BackendNone, nil)
	assert.Equal(t, "backend explicitly set to none", iso.Reason())
	assert.Equal(t, "backend explicitly set to none", info.Reason)
}

func TestSelectBackend_AutoNoCandidatesReason(t *testing.T) {
	iso, info := SelectBackend(BackendAuto, nil)
	assert.Equal(t, "no backends configured for this platform", iso.Reason())
	assert.Equal(t, "no backends configured for this platform", info.Reason)
}

func TestSelectBackend_AutoAggregatesCandidateReasons(t *testing.T) {
	candidates := []BackendCandidate{
		{Mode: BackendBwrap, Isolator: NewBwrapStub()},
		{Mode: BackendNative, Isolator: NewNativeStub()},
	}
	iso, _ := SelectBackend(BackendAuto, candidates)
	reason := iso.Reason()
	assert.Contains(t, reason, "bwrap: bwrap backend not yet implemented")
	assert.Contains(t, reason, "native: native backend not yet implemented")
}

func TestSelectBackend_ExplicitNotFoundReason(t *testing.T) {
	iso, info := SelectBackend(BackendBwrap, nil)
	assert.Contains(t, iso.Reason(), "backend bwrap not available on this platform")
	assert.Contains(t, info.Reason, "backend bwrap not available on this platform")
}

func TestSelectBackend_ExplicitPreservesIdentity(t *testing.T) {
	// When an explicit mode is requested and the candidate exists (even if unavailable),
	// the returned isolator must be the candidate's isolator, not a noop.
	stub := NewBwrapStub()
	candidates := []BackendCandidate{
		{Mode: BackendBwrap, Isolator: stub},
	}

	iso, _ := SelectBackend(BackendBwrap, candidates)
	assert.Same(t, stub, iso, "explicit mode must return the candidate isolator, not a noop")
	assert.False(t, iso.Available())
	assert.Equal(t, "bwrap", iso.Name())
}

func TestListBackends(t *testing.T) {
	candidates := []BackendCandidate{
		{Mode: BackendSeatbelt, Isolator: &fakeIsolator{name: "seatbelt", available: true}},
		{Mode: BackendBwrap, Isolator: &fakeIsolator{name: "bwrap", available: false, reason: "not installed"}},
		{Mode: BackendNative, Isolator: &fakeIsolator{name: "native", available: false, reason: "kernel too old"}},
	}

	infos := ListBackends(candidates)
	require.Len(t, infos, 3)

	assert.Equal(t, "seatbelt", infos[0].Name)
	assert.Equal(t, BackendSeatbelt, infos[0].Mode)
	assert.True(t, infos[0].Available)
	assert.Empty(t, infos[0].Reason)

	assert.Equal(t, "bwrap", infos[1].Name)
	assert.Equal(t, BackendBwrap, infos[1].Mode)
	assert.False(t, infos[1].Available)
	assert.Equal(t, "not installed", infos[1].Reason)

	assert.Equal(t, "native", infos[2].Name)
	assert.Equal(t, BackendNative, infos[2].Mode)
	assert.False(t, infos[2].Available)
	assert.Equal(t, "kernel too old", infos[2].Reason)
}

func TestListBackends_Empty(t *testing.T) {
	infos := ListBackends(nil)
	assert.Empty(t, infos)
}

func TestBwrapStub(t *testing.T) {
	stub := NewBwrapStub()
	assert.False(t, stub.Available())
	assert.Equal(t, "bwrap", stub.Name())
	assert.Equal(t, "bwrap backend not yet implemented", stub.Reason())

	err := stub.Apply(context.Background(), &exec.Cmd{}, Policy{})
	assert.ErrorIs(t, err, ErrIsolatorUnavailable)
}

func TestNativeStub(t *testing.T) {
	stub := NewNativeStub()
	assert.False(t, stub.Available())
	assert.Equal(t, "native", stub.Name())
	assert.Equal(t, "native backend not yet implemented", stub.Reason())

	err := stub.Apply(context.Background(), &exec.Cmd{}, Policy{})
	assert.ErrorIs(t, err, ErrIsolatorUnavailable)
}

func TestPlatformBackendCandidates(t *testing.T) {
	candidates := PlatformBackendCandidates()

	switch runtime.GOOS {
	case "darwin":
		require.Len(t, candidates, 3)
		assert.Equal(t, BackendSeatbelt, candidates[0].Mode)
		assert.Equal(t, "seatbelt", candidates[0].Isolator.Name())
		assert.Equal(t, BackendBwrap, candidates[1].Mode)
		assert.Equal(t, "bwrap", candidates[1].Isolator.Name())
		assert.Equal(t, BackendNative, candidates[2].Mode)
		assert.Equal(t, "native", candidates[2].Isolator.Name())

	case "linux":
		require.Len(t, candidates, 2)
		assert.Equal(t, BackendBwrap, candidates[0].Mode)
		assert.Equal(t, "bwrap", candidates[0].Isolator.Name())
		assert.Equal(t, BackendNative, candidates[1].Mode)
		assert.Equal(t, "native", candidates[1].Isolator.Name())

	default:
		assert.Empty(t, candidates)
	}
}
