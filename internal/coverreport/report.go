package coverreport

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Options struct {
	ProfilePath string
	RootDir     string
	ModulePath  string
	TopLimit    int
}

type Report struct {
	CoveredStatements   int
	TotalStatements     int
	UncoveredStatements int
	Files               []FileCoverage
	TopUncovered        []FileCoverage
}

type FileCoverage struct {
	Path                string
	CoveredStatements   int
	TotalStatements     int
	UncoveredStatements int
}

type ThresholdError struct {
	Measured  float64
	Threshold float64
}

func AnalyzeProfile(ctx context.Context, opts Options) (Report, error) {
	if opts.ProfilePath == "" {
		return Report{}, fmt.Errorf("coverage profile path is required")
	}
	root := opts.RootDir
	if root == "" {
		root = "."
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return Report{}, fmt.Errorf("resolve root directory: %w", err)
	}

	file, err := os.Open(opts.ProfilePath)
	if err != nil {
		return Report{}, fmt.Errorf("open coverage profile: %w", err)
	}
	defer file.Close()

	files := make(map[string]*FileCoverage)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return Report{}, ctx.Err()
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}
		block, err := parseProfileLine(line)
		if err != nil {
			return Report{}, err
		}
		relPath, sourcePath, err := resolveProfilePath(rootAbs, opts.ModulePath, block.Path)
		if err != nil {
			return Report{}, err
		}
		generated, err := isGeneratedSource(relPath, sourcePath)
		if err != nil {
			return Report{}, err
		}
		if generated {
			continue
		}

		coverage := files[relPath]
		if coverage == nil {
			coverage = &FileCoverage{Path: relPath}
			files[relPath] = coverage
		}
		coverage.TotalStatements += block.Statements
		if block.Count > 0 {
			coverage.CoveredStatements += block.Statements
		} else {
			coverage.UncoveredStatements += block.Statements
		}
	}
	if err := scanner.Err(); err != nil {
		return Report{}, fmt.Errorf("read coverage profile: %w", err)
	}

	report := buildReport(files, opts.TopLimit)
	return report, nil
}

func (r Report) Percent() float64 {
	if r.TotalStatements == 0 {
		return 0
	}
	return float64(r.CoveredStatements) * 100 / float64(r.TotalStatements)
}

func (f FileCoverage) Percent() float64 {
	if f.TotalStatements == 0 {
		return 0
	}
	return float64(f.CoveredStatements) * 100 / float64(f.TotalStatements)
}

func CheckThreshold(report Report, threshold float64) error {
	measured := report.Percent()
	if measured >= threshold {
		return nil
	}
	return ThresholdError{Measured: measured, Threshold: threshold}
}

func (e ThresholdError) Error() string {
	return fmt.Sprintf("non-generated coverage %.2f%% is below required %.2f%%", e.Measured, e.Threshold)
}

type profileBlock struct {
	Path       string
	Statements int
	Count      int
}

func parseProfileLine(line string) (profileBlock, error) {
	fields := strings.Fields(line)
	if len(fields) != 3 {
		return profileBlock{}, fmt.Errorf("invalid coverage profile line %q", line)
	}
	colon := strings.LastIndex(fields[0], ":")
	if colon <= 0 {
		return profileBlock{}, fmt.Errorf("invalid coverage profile location %q", fields[0])
	}
	statements, err := strconv.Atoi(fields[1])
	if err != nil {
		return profileBlock{}, fmt.Errorf("invalid statement count %q: %w", fields[1], err)
	}
	count, err := strconv.Atoi(fields[2])
	if err != nil {
		return profileBlock{}, fmt.Errorf("invalid coverage count %q: %w", fields[2], err)
	}
	return profileBlock{
		Path:       fields[0][:colon],
		Statements: statements,
		Count:      count,
	}, nil
}

func resolveProfilePath(rootAbs, modulePath, profilePath string) (string, string, error) {
	slashed := filepath.ToSlash(profilePath)
	if modulePath != "" {
		modulePrefix := strings.TrimSuffix(modulePath, "/") + "/"
		slashed = strings.TrimPrefix(slashed, modulePrefix)
	}
	slashed = strings.TrimPrefix(slashed, "./")

	var sourcePath string
	if filepath.IsAbs(profilePath) {
		sourcePath = profilePath
		rel, err := filepath.Rel(rootAbs, sourcePath)
		if err == nil && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".." {
			slashed = filepath.ToSlash(rel)
		}
	} else {
		sourcePath = filepath.Join(rootAbs, filepath.FromSlash(slashed))
	}
	return slashed, sourcePath, nil
}

func isGeneratedSource(relPath, sourcePath string) (bool, error) {
	slashed := filepath.ToSlash(relPath)
	if strings.HasPrefix(slashed, "internal/ent/") {
		return true, nil
	}

	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return false, fmt.Errorf("read source file %q: %w", relPath, err)
	}
	return hasGeneratedMarker(string(content)), nil
}

func hasGeneratedMarker(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "package ") {
			break
		}
		if strings.HasPrefix(trimmed, "// Code generated ") && strings.Contains(trimmed, " DO NOT EDIT.") {
			return true
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		break
	}
	return false
}

func buildReport(files map[string]*FileCoverage, topLimit int) Report {
	report := Report{
		Files: make([]FileCoverage, 0, len(files)),
	}
	for _, file := range files {
		report.CoveredStatements += file.CoveredStatements
		report.TotalStatements += file.TotalStatements
		report.UncoveredStatements += file.UncoveredStatements
		report.Files = append(report.Files, *file)
	}
	sort.Slice(report.Files, func(i, j int) bool {
		return report.Files[i].Path < report.Files[j].Path
	})

	report.TopUncovered = append(report.TopUncovered, report.Files...)
	sort.Slice(report.TopUncovered, func(i, j int) bool {
		if report.TopUncovered[i].UncoveredStatements != report.TopUncovered[j].UncoveredStatements {
			return report.TopUncovered[i].UncoveredStatements > report.TopUncovered[j].UncoveredStatements
		}
		return report.TopUncovered[i].Path < report.TopUncovered[j].Path
	})
	if topLimit > 0 && len(report.TopUncovered) > topLimit {
		report.TopUncovered = report.TopUncovered[:topLimit]
	}
	return report
}
