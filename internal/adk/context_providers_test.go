package adk

import (
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/knowledge"
)

func TestToolRegistryAdapter_ListTools(t *testing.T) {
	tools := []*agent.Tool{
		{Name: "exec", Description: "Execute commands"},
		{Name: "read", Description: "Read files"},
	}
	adapter := NewToolRegistryAdapter(tools)

	got := adapter.ListTools()
	if len(got) != 2 {
		t.Fatalf("want 2 tools, got %d", len(got))
	}
	if got[0].Name != "exec" {
		t.Errorf("want exec, got %s", got[0].Name)
	}
	if got[1].Name != "read" {
		t.Errorf("want read, got %s", got[1].Name)
	}
}

func TestToolRegistryAdapter_SearchTools(t *testing.T) {
	adapter := NewToolRegistryAdapter([]*agent.Tool{
		{Name: "exec_command", Description: "Execute shell commands"},
		{Name: "read_file", Description: "Read file contents"},
		{Name: "write_file", Description: "Write file contents"},
		{Name: "web_search", Description: "Search the web"},
	})

	tests := []struct {
		give      string
		giveLimit int
		wantCount int
		wantFirst string
	}{
		{give: "exec", giveLimit: 10, wantCount: 1, wantFirst: "exec_command"},
		{give: "file", giveLimit: 10, wantCount: 2, wantFirst: "read_file"},
		{give: "file", giveLimit: 1, wantCount: 1, wantFirst: "read_file"},
		{give: "SEARCH", giveLimit: 10, wantCount: 1, wantFirst: "web_search"},
		{give: "nonexistent", giveLimit: 10, wantCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := adapter.SearchTools(tt.give, tt.giveLimit)
			if len(got) != tt.wantCount {
				t.Fatalf("want %d results, got %d", tt.wantCount, len(got))
			}
			if tt.wantCount > 0 && got[0].Name != tt.wantFirst {
				t.Errorf("want first %s, got %s", tt.wantFirst, got[0].Name)
			}
		})
	}
}

func TestToolRegistryAdapter_BoundaryCopy(t *testing.T) {
	tools := []*agent.Tool{
		{Name: "original", Description: "Original tool"},
	}
	adapter := NewToolRegistryAdapter(tools)

	// Mutate original slice
	tools[0].Name = "mutated"

	got := adapter.ListTools()
	if got[0].Name != "original" {
		t.Errorf("boundary copy violated: want original, got %s", got[0].Name)
	}
}

func TestRuntimeContextAdapter(t *testing.T) {
	adapter := NewRuntimeContextAdapter(5, true, true, false)

	t.Run("defaults", func(t *testing.T) {
		rc := adapter.GetRuntimeContext()
		if rc.ActiveToolCount != 5 {
			t.Errorf("want 5 tools, got %d", rc.ActiveToolCount)
		}
		if !rc.EncryptionEnabled {
			t.Error("want encryption enabled")
		}
		if !rc.KnowledgeEnabled {
			t.Error("want knowledge enabled")
		}
		if rc.MemoryEnabled {
			t.Error("want memory disabled")
		}
		if rc.ChannelType != "direct" {
			t.Errorf("want direct channel, got %s", rc.ChannelType)
		}
		if rc.SessionKey != "" {
			t.Errorf("want empty session key, got %s", rc.SessionKey)
		}
	})

	t.Run("SetSession updates state", func(t *testing.T) {
		adapter.SetSession("telegram:123:456")
		rc := adapter.GetRuntimeContext()
		if rc.SessionKey != "telegram:123:456" {
			t.Errorf("want telegram:123:456, got %s", rc.SessionKey)
		}
		if rc.ChannelType != "telegram" {
			t.Errorf("want telegram, got %s", rc.ChannelType)
		}
	})
}

func TestDeriveChannelType(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{give: "", want: "direct"},
		{give: "noseparator", want: "direct"},
		{give: "telegram:123:456", want: "telegram"},
		{give: "discord:guild:channel", want: "discord"},
		{give: "slack:team:channel", want: "slack"},
		{give: "unknown:123:456", want: "direct"},
		{give: "http:something", want: "direct"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := deriveChannelType(tt.give)
			if got != tt.want {
				t.Errorf("deriveChannelType(%q): want %q, got %q", tt.give, tt.want, got)
			}
		})
	}
}

// Verify interface compliance at compile time.
var _ knowledge.ToolRegistryProvider = (*ToolRegistryAdapter)(nil)
var _ knowledge.RuntimeContextProvider = (*RuntimeContextAdapter)(nil)
