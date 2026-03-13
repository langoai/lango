package toolparam

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireString(t *testing.T) {
	tests := []struct {
		give     map[string]interface{}
		giveKey  string
		wantVal  string
		wantErr  bool
		wantName string
	}{
		{
			give:    map[string]interface{}{"name": "alice"},
			giveKey: "name",
			wantVal: "alice",
		},
		{
			give:     map[string]interface{}{"name": ""},
			giveKey:  "name",
			wantErr:  true,
			wantName: "name",
		},
		{
			give:     map[string]interface{}{},
			giveKey:  "name",
			wantErr:  true,
			wantName: "name",
		},
		{
			give:     map[string]interface{}{"name": 123},
			giveKey:  "name",
			wantErr:  true,
			wantName: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.giveKey, func(t *testing.T) {
			val, err := RequireString(tt.give, tt.giveKey)
			if tt.wantErr {
				require.Error(t, err)
				var missing *ErrMissingParam
				require.True(t, errors.As(err, &missing))
				assert.Equal(t, tt.wantName, missing.Name)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantVal, val)
			}
		})
	}
}

func TestOptionalString(t *testing.T) {
	tests := []struct {
		give    map[string]interface{}
		giveKey string
		giveFB  string
		wantVal string
	}{
		{
			give:    map[string]interface{}{"key": "val"},
			giveKey: "key",
			giveFB:  "default",
			wantVal: "val",
		},
		{
			give:    map[string]interface{}{"key": ""},
			giveKey: "key",
			giveFB:  "default",
			wantVal: "default",
		},
		{
			give:    map[string]interface{}{},
			giveKey: "key",
			giveFB:  "default",
			wantVal: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.giveKey, func(t *testing.T) {
			assert.Equal(t, tt.wantVal, OptionalString(tt.give, tt.giveKey, tt.giveFB))
		})
	}
}

func TestOptionalInt(t *testing.T) {
	tests := []struct {
		give    map[string]interface{}
		giveKey string
		giveFB  int
		wantVal int
	}{
		{
			give:    map[string]interface{}{"limit": float64(42)},
			giveKey: "limit",
			giveFB:  20,
			wantVal: 42,
		},
		{
			give:    map[string]interface{}{},
			giveKey: "limit",
			giveFB:  20,
			wantVal: 20,
		},
		{
			give:    map[string]interface{}{"limit": "not a number"},
			giveKey: "limit",
			giveFB:  20,
			wantVal: 20,
		},
		{
			give:    map[string]interface{}{"limit": float64(0)},
			giveKey: "limit",
			giveFB:  20,
			wantVal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.giveKey, func(t *testing.T) {
			assert.Equal(t, tt.wantVal, OptionalInt(tt.give, tt.giveKey, tt.giveFB))
		})
	}
}

func TestOptionalBool(t *testing.T) {
	tests := []struct {
		give    map[string]interface{}
		giveKey string
		giveFB  bool
		wantVal bool
	}{
		{
			give:    map[string]interface{}{"flag": true},
			giveKey: "flag",
			giveFB:  false,
			wantVal: true,
		},
		{
			give:    map[string]interface{}{"flag": false},
			giveKey: "flag",
			giveFB:  true,
			wantVal: false,
		},
		{
			give:    map[string]interface{}{},
			giveKey: "flag",
			giveFB:  true,
			wantVal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.giveKey, func(t *testing.T) {
			assert.Equal(t, tt.wantVal, OptionalBool(tt.give, tt.giveKey, tt.giveFB))
		})
	}
}

func TestStringSlice(t *testing.T) {
	tests := []struct {
		give    map[string]interface{}
		giveKey string
		wantVal []string
	}{
		{
			give:    map[string]interface{}{"tags": []interface{}{"a", "b", "c"}},
			giveKey: "tags",
			wantVal: []string{"a", "b", "c"},
		},
		{
			give:    map[string]interface{}{},
			giveKey: "tags",
			wantVal: nil,
		},
		{
			give:    map[string]interface{}{"tags": []interface{}{"a", 42, "c"}},
			giveKey: "tags",
			wantVal: []string{"a", "c"},
		},
		{
			give:    map[string]interface{}{"tags": []interface{}{}},
			giveKey: "tags",
			wantVal: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.giveKey, func(t *testing.T) {
			result := StringSlice(tt.give, tt.giveKey)
			assert.Equal(t, tt.wantVal, result)
		})
	}
}

func TestErrMissingParam_ErrorsAs(t *testing.T) {
	_, err := RequireString(map[string]interface{}{}, "workspaceId")
	require.Error(t, err)

	var target *ErrMissingParam
	require.True(t, errors.As(err, &target))
	assert.Equal(t, "workspaceId", target.Name)
	assert.Equal(t, "missing workspaceId parameter", err.Error())
}
