package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
// from the registry, writes an ExecutionResult to stdout, and returns
// the intended process exit code.
func RunWorker(registry ToolRegistry) int {
	return RunWorkerWithIO(registry, os.Stdin, os.Stdout)
}

// RunWorkerWithIO executes one sandbox worker request using injected IO.
// It writes exactly one ExecutionResult and returns the intended process exit code.
func RunWorkerWithIO(registry ToolRegistry, in io.Reader, out io.Writer) int {
	var req ExecutionRequest
	if err := json.NewDecoder(in).Decode(&req); err != nil {
		writeResult(out, ExecutionResult{Error: fmt.Sprintf("decode request: %v", err)})
		return 1
	}

	handler, ok := registry[req.ToolName]
	if !ok {
		writeResult(out, ExecutionResult{Error: fmt.Sprintf("tool %q not registered in worker", req.ToolName)})
		return 1
	}

	ctx := context.Background()
	result, err := handler(ctx, req.Params)
	if err != nil {
		writeResult(out, ExecutionResult{Error: err.Error()})
		return 0
	}

	// Coerce result to map[string]interface{}.
	var output map[string]interface{}
	switch v := result.(type) {
	case map[string]interface{}:
		output = v
	default:
		output = map[string]interface{}{"result": v}
	}

	writeResult(out, ExecutionResult{Output: output})
	return 0
}

// writeResult encodes an ExecutionResult to out.
func writeResult(out io.Writer, r ExecutionResult) {
	_ = json.NewEncoder(out).Encode(r)
}
