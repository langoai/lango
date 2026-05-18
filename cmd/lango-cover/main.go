package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/langoai/lango/internal/coverreport"
)

var (
	langoCoverArgs             = func() []string { return os.Args[1:] }
	langoCoverStdout io.Writer = os.Stdout
	langoCoverStderr io.Writer = os.Stderr
	langoCoverExitFn           = os.Exit
)

func main() {
	langoCoverExitFn(run(langoCoverArgs(), langoCoverStdout, langoCoverStderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("lango-cover", flag.ContinueOnError)
	flags.SetOutput(stderr)

	profilePath := flags.String("profile", ".coverage/non-generated.out", "Go coverage profile to analyze")
	rootDir := flags.String("root", ".", "repository root used to resolve source files")
	modulePath := flags.String("module", "", "Go module path; defaults to module in root go.mod")
	topLimit := flags.Int("top", 10, "number of files to show by uncovered statements")
	threshold := flags.Float64("threshold", 0, "minimum coverage percent; 0 disables gate")

	if err := flags.Parse(args); err != nil {
		return 2
	}

	module := *modulePath
	if module == "" {
		module = readModulePath(*rootDir)
	}
	report, err := coverreport.AnalyzeProfile(context.Background(), coverreport.Options{
		ProfilePath: *profilePath,
		RootDir:     *rootDir,
		ModulePath:  module,
		TopLimit:    *topLimit,
	})
	if err != nil {
		fmt.Fprintf(stderr, "coverage report failed: %v\n", err)
		return 1
	}

	printReport(stdout, report, *threshold)
	if *threshold > 0 {
		if err := coverreport.CheckThreshold(report, *threshold); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	}
	return 0
}

func printReport(w io.Writer, report coverreport.Report, threshold float64) {
	fmt.Fprintf(w, "Non-generated coverage: %.2f%%\n", report.Percent())
	fmt.Fprintf(w, "Covered statements: %d\n", report.CoveredStatements)
	fmt.Fprintf(w, "Total statements: %d\n", report.TotalStatements)
	fmt.Fprintf(w, "Uncovered statements: %d\n", report.UncoveredStatements)
	if threshold > 0 {
		fmt.Fprintf(w, "Threshold: %.2f%%\n", threshold)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Top uncovered files:")
	for i, file := range report.TopUncovered {
		fmt.Fprintf(w, "%2d. %-64s %7d uncovered / %7d total (%6.2f%%)\n",
			i+1,
			file.Path,
			file.UncoveredStatements,
			file.TotalStatements,
			file.Percent(),
		)
	}
}

func readModulePath(rootDir string) string {
	content, err := os.ReadFile(filepath.Join(rootDir, "go.mod"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(content), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "module" {
			return fields[1]
		}
	}
	return ""
}
