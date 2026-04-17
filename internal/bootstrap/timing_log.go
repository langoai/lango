package bootstrap

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	timingLogDir      = "diagnostics"
	timingLogFile     = "bootstrap-timing.jsonl"
	maxTimingEntries  = 50
	timingFilePerm    = 0o644
	timingDirPerm     = 0o755
)

type TimingLogEntry struct {
	Timestamp string              `json:"ts"`
	Version   string              `json:"version"`
	Phases    []PhaseTimingRecord `json:"phases"`
}

type PhaseTimingRecord struct {
	Name       string `json:"name"`
	DurationMs int64  `json:"durationMs"`
}

// timingLogPathFn resolves the JSONL path. Override in tests.
var timingLogPathFn = defaultTimingLogPath

func defaultTimingLogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".lango", timingLogDir, timingLogFile), nil
}

// AppendTimingLog persists phase timing to the diagnostics JSONL file.
// Errors are returned to the caller for logging; they should never be fatal.
func AppendTimingLog(entries []PhaseTimingEntry, version string) error {
	path, err := timingLogPathFn()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), timingDirPerm); err != nil {
		return err
	}

	existing, _ := readTimingLog(path)

	record := TimingLogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   version,
		Phases:    make([]PhaseTimingRecord, len(entries)),
	}
	for i, e := range entries {
		record.Phases[i] = PhaseTimingRecord{
			Name:       e.Phase,
			DurationMs: e.Duration.Milliseconds(),
		}
	}
	existing = append(existing, record)

	if len(existing) > maxTimingEntries {
		existing = existing[len(existing)-maxTimingEntries:]
	}

	return writeTimingLog(path, existing)
}

// ReadTimingLog reads all valid entries from the JSONL file.
// Corrupted lines are silently skipped.
func ReadTimingLog() ([]TimingLogEntry, error) {
	path, err := timingLogPathFn()
	if err != nil {
		return nil, err
	}
	return readTimingLog(path)
}

func readTimingLog(path string) ([]TimingLogEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []TimingLogEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var e TimingLogEntry
		if json.Unmarshal(scanner.Bytes(), &e) == nil && e.Timestamp != "" {
			entries = append(entries, e)
		}
	}
	return entries, scanner.Err()
}

func writeTimingLog(path string, entries []TimingLogEntry) error {
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, timingFilePerm)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			f.Close()
			os.Remove(tmp)
			return err
		}
	}

	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}
