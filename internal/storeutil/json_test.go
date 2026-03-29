package storeutil

import (
	"math"
	"testing"
)

func TestMarshalField(t *testing.T) {
	tests := []struct {
		give     interface{}
		wantJSON string
		wantErr  bool
	}{
		{give: map[string]string{"key": "val"}, wantJSON: `{"key":"val"}`},
		{give: nil, wantJSON: "null"},
		{give: "hello", wantJSON: `"hello"`},
		{give: math.NaN(), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.wantJSON, func(t *testing.T) {
			got, err := MarshalField(tt.give)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error for unmarshalable value")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.wantJSON {
				t.Errorf("MarshalField(%v) = %s, want %s", tt.give, got, tt.wantJSON)
			}
		})
	}
}

func TestUnmarshalField(t *testing.T) {
	var target map[string]string
	err := UnmarshalField([]byte(`{"key":"val"}`), &target, "test payload")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target["key"] != "val" {
		t.Errorf("got %v, want key=val", target)
	}

	err = UnmarshalField([]byte(`invalid`), &target, "bad data")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
