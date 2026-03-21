package gitbundle

import "context"

// FileStat is a per-file diff summary.
type FileStat struct {
	FilePath     string `json:"filePath"`
	LinesAdded   int    `json:"linesAdded"`
	LinesRemoved int    `json:"linesRemoved"`
}

// MergeHookEvent captures a task-branch merge for external observers.
type MergeHookEvent struct {
	WorkspaceID    string
	TaskID         string
	TargetBranch   string
	MergeCommit    string
	SourceCommit   string
	PreviousTarget string
	Files          []FileStat
}

// BundleCreatedHook observes bundle creation.
type BundleCreatedHook func(ctx context.Context, workspaceID, headCommit string, bundleSize int)

// MergeHook observes task branch merges.
type MergeHook func(ctx context.Context, event MergeHookEvent)
