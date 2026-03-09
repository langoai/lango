package types

import (
	"strings"
	"testing"
)

func BenchmarkEstimateTokens(b *testing.B) {
	tests := []struct {
		name string
		give string
	}{
		{
			name: "Short_ASCII",
			give: "Hello, world!",
		},
		{
			name: "Medium_ASCII",
			give: strings.Repeat("The quick brown fox jumps over the lazy dog. ", 20),
		},
		{
			name: "Long_ASCII",
			give: strings.Repeat("The quick brown fox jumps over the lazy dog. ", 200),
		},
		{
			name: "Short_CJK",
			give: "안녕하세요 세계",
		},
		{
			name: "Medium_CJK",
			give: strings.Repeat("이것은 한국어 테스트 문장입니다. ", 20),
		},
		{
			name: "Long_CJK",
			give: strings.Repeat("이것은 한국어 테스트 문장입니다. ", 200),
		},
		{
			name: "Mixed_ASCII_CJK",
			give: strings.Repeat("Hello 안녕 World 세계 ", 50),
		},
		{
			name: "Empty",
			give: "",
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				EstimateTokens(tt.give)
			}
		})
	}
}

func BenchmarkIsCJK(b *testing.B) {
	tests := []struct {
		name string
		give rune
	}{
		{"ASCII", 'A'},
		{"CJK_Unified", '中'},
		{"Korean_Hangul", '한'},
		{"CJK_ExtA", '\u3500'},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				IsCJK(tt.give)
			}
		})
	}
}
