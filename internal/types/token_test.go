package types

import "testing"

func TestEstimateTokens_UsesCurrentBucketFlooring(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{name: "empty", text: "", want: 0},
		{name: "ascii below one token", text: "abc", want: 0},
		{name: "ascii at one token", text: "abcd", want: 1},
		{name: "ascii floors partial token", text: "abcdefg", want: 1},
		{name: "cjk below one token", text: "你", want: 0},
		{name: "cjk at one token", text: "你好", want: 1},
		{name: "mixed counts buckets separately", text: "abcd你好", want: 2},
		{name: "non-cjk multibyte text uses non-cjk bucket", text: "🙂🙂🙂🙂", want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EstimateTokens(tt.text); got != tt.want {
				t.Fatalf("EstimateTokens(%q) = %d, want %d", tt.text, got, tt.want)
			}
		})
	}
}

func TestIsCJK_RangeBoundaries(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{name: "before extension a", r: '\u33FF', want: false},
		{name: "extension a start", r: '\u3400', want: true},
		{name: "extension a end", r: '\u4DBF', want: true},
		{name: "before unified ideographs", r: '\u4DFF', want: false},
		{name: "unified ideographs start", r: '\u4E00', want: true},
		{name: "unified ideographs end", r: '\u9FFF', want: true},
		{name: "after unified ideographs", r: '\uA000', want: false},
		{name: "before hangul syllables", r: '\uABFF', want: false},
		{name: "hangul syllables start", r: '\uAC00', want: true},
		{name: "hangul syllables end", r: '\uD7AF', want: true},
		{name: "after hangul syllables", r: '\uD7B0', want: false},
		{name: "latin letter", r: 'a', want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCJK(tt.r); got != tt.want {
				t.Fatalf("IsCJK(%U) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}
