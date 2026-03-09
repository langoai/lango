package prompt

import (
	"fmt"
	"strings"
	"testing"
)

func BenchmarkBuilderBuild(b *testing.B) {
	tests := []struct {
		name         string
		sectionCount int
	}{
		{"Sections_3", 3},
		{"Sections_7", 7},
		{"Sections_20", 20},
	}

	for _, tt := range tests {
		builder := NewBuilder()
		for i := 0; i < tt.sectionCount; i++ {
			builder.Add(NewStaticSection(
				SectionID(fmt.Sprintf("bench_%d", i)),
				(tt.sectionCount-i)*10, // reverse priority to force sorting
				fmt.Sprintf("Section %d", i),
				strings.Repeat("This is prompt content for benchmarking. ", 10),
			))
		}

		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				builder.Build()
			}
		})
	}
}

func BenchmarkBuilderAdd(b *testing.B) {
	b.ReportAllocs()

	sections := make([]*StaticSection, 20)
	for i := range sections {
		sections[i] = NewStaticSection(
			SectionID(fmt.Sprintf("bench_%d", i)),
			i*100,
			fmt.Sprintf("Section %d", i),
			"content",
		)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewBuilder()
		for _, s := range sections {
			builder.Add(s)
		}
	}
}

func BenchmarkBuilderAddReplace(b *testing.B) {
	b.ReportAllocs()

	builder := NewBuilder()
	for i := 0; i < 10; i++ {
		builder.Add(NewStaticSection(
			SectionID(fmt.Sprintf("bench_%d", i)),
			i*100,
			fmt.Sprintf("Section %d", i),
			"original content",
		))
	}

	replacement := NewStaticSection("bench_5", 500, "Section 5", "replaced content")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.Add(replacement)
	}
}

func BenchmarkStaticSectionRender(b *testing.B) {
	tests := []struct {
		name string
		give *StaticSection
	}{
		{
			name: "WithTitle",
			give: NewStaticSection(SectionIdentity, 100, "Identity", strings.Repeat("I am an AI assistant. ", 20)),
		},
		{
			name: "WithoutTitle",
			give: NewStaticSection(SectionCustom, 600, "", strings.Repeat("Custom prompt content. ", 20)),
		},
		{
			name: "Empty",
			give: NewStaticSection(SectionCustom, 600, "Empty", ""),
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tt.give.Render()
			}
		})
	}
}
