package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// workerFlag is the CLI flag that triggers sandbox worker mode.
const workerFlag = "--sandbox-worker"

// IsWorkerMode returns true if the process was launched as a sandbox worker.
func IsWorkerMode() bool {
	for _, arg := range os.Args[1:] {
		if arg == workerFlag {
			return true
		}
	}
	return false
}

// ToolHandler is a function that executes a named tool with parameters.
type ToolHandler func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// ToolRegistry maps tool names to their handlers for the worker process.
type ToolRegistry map[string]ToolHandler

// RunWorker is the entry point for the sandbox worker subprocess.
// It reads an ExecutionRequest from stdin, executes the named tool
// from the registry, and writes an ExecutionResult to stdout.
// The worker exits with code 0 on success, 1 on failure.
func RunWorker(registry ToolRegistry) {
	var req ExecutionRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		writeResult(ExecutionResult{Error: fmt.Sprintf("decode request: %v", err)})
		os.Exit(1)
	}

	handler, ok := registry[req.ToolName]
	if !ok {
		writeResult(ExecutionResult{Error: fmt.Sprintf("tool %q not registered in worker", req.ToolName)})
		os.Exit(1)
	}

	ctx := context.Background()
	result, err := handler(ctx, req.Params)
	if err != nil {
		writeResult(ExecutionResult{Error: err.Error()})
		os.Exit(0) // exit 0 — error is communicated via JSON
	}

	// Coerce result to map[string]interface{}.
	var output map[string]interface{}
	switch v := result.(type) {
	case map[string]interface{}:
		output = v
	default:
		output = map[string]interface{}{"result": v}
	}

	writeResult(ExecutionResult{Output: output})
}

// writeResult encodes an ExecutionResult to stdout.
func writeResult(r ExecutionResult) {
	_ = json.NewEncoder(os.Stdout).Encode(r)
}
