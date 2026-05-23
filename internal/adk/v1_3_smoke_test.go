package adk

import (
	"testing"

	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool/functiontool"
	"google.golang.org/adk/tool/skilltoolset"
)

// TestADKv13SurfaceReachable asserts that the v1.3 API surface we plan to
// adopt in Tracks A-F compiles and is importable. It does NOT exercise
// behavior — it is a tripwire against accidental downgrade.
func TestADKv13SurfaceReachable(t *testing.T) {
	// v1.1+: Agent interface includes FindAgent/FindSubAgent (compile-time check).
	var _ interface {
		FindAgent(name string) adkagent.Agent
		FindSubAgent(name string) adkagent.Agent
	} = (adkagent.Agent)(nil)

	// v1.3: model.LLMResponse has transcription fields (compile-time field access).
	var resp model.LLMResponse
	_ = resp.InputTranscription
	_ = resp.OutputTranscription
	_ = resp.SessionResumptionHandle

	// v1.3: streaming function tools package is importable.
	_ = functiontool.Config{}

	// v1.2+: skill toolset package is importable.
	_ = skilltoolset.Config{}

	// v1.3: live session interface is reachable (used in Track E).
	var _ adkagent.LiveSession = (adkagent.LiveSession)(nil)
}
