package adk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/knowledge"
)

func TestToolRegistryAdapter_ListTools(t *testing.T) {
	t.Parallel()

	tools := []*agent.Tool{
		{Name: "exec", Description: "Execute commands"},
		{Name: "read", Description: "Read files"},
	}
	adapter := NewToolRegistryAdapter(tools)

	got := adapter.ListTools()
	require.Len(t, got, 2)
	assert.Equal(t, "exec", got[0].Name)
	assert.Equal(t, "read", got[1].Name)
}

func TestToolRegistryAdapter_SearchTools(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			got := adapter.SearchTools(tt.give, tt.giveLimit)
			require.Len(t, got, tt.wantCount)
			if tt.wantCount > 0 {
				assert.Equal(t, tt.wantFirst, got[0].Name)
			}
		})
	}
}

func TestToolRegistryAdapter_BoundaryCopy(t *testing.T) {
	t.Parallel()

	tools := []*agent.Tool{
		{Name: "original", Description: "Original tool"},
	}
	adapter := NewToolRegistryAdapter(tools)

	// Mutate original slice
	tools[0].Name = "mutated"

	got := adapter.ListTools()
	assert.Equal(t, "original", got[0].Name, "boundary copy violated")
}

func TestRuntimeContextAdapter(t *testing.T) {
	t.Parallel()

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()
		adapter := NewRuntimeContextAdapter(5, true, true, false)
		rc := adapter.GetRuntimeContext()
		assert.Equal(t, 5, rc.ActiveToolCount)
		assert.True(t, rc.EncryptionEnabled, "want encryption enabled")
		assert.True(t, rc.KnowledgeEnabled, "want knowledge enabled")
		assert.False(t, rc.MemoryEnabled, "want memory disabled")
		assert.Equal(t, "direct", rc.ChannelType)
		assert.Empty(t, rc.SessionKey)
	})

	t.Run("SetSession updates state", func(t *testing.T) {
		t.Parallel()
		adapter := NewRuntimeContextAdapter(5, true, true, false)
		adapter.SetSession("telegram:123:456")
		rc := adapter.GetRuntimeContext()
		assert.Equal(t, "telegram:123:456", rc.SessionKey)
		assert.Equal(t, "telegram", rc.ChannelType)
	})
}

func TestDeriveChannelType(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			got := deriveChannelType(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToolRegistryAdapter_SearchTools_Aliases(t *testing.T) {
	t.Parallel()

	adapter := NewToolRegistryAdapter([]*agent.Tool{
		{
			Name:        "fs_list",
			Description: "List directory contents",
			Capability: agent.ToolCapability{
				Aliases: []string{"ls", "dir"},
			},
		},
		{
			Name:        "fs_read",
			Description: "Read file contents",
			Capability: agent.ToolCapability{
				Aliases: []string{"cat"},
			},
		},
	})

	tests := []struct {
		give      string
		wantCount int
		wantFirst string
	}{
		{give: "ls", wantCount: 1, wantFirst: "fs_list"},
		{give: "dir", wantCount: 1, wantFirst: "fs_list"},
		{give: "cat", wantCount: 1, wantFirst: "fs_read"},
		{give: "LS", wantCount: 1, wantFirst: "fs_list"},  // case insensitive
		{give: "nope", wantCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := adapter.SearchTools(tt.give, 10)
			require.Len(t, got, tt.wantCount)
			if tt.wantCount > 0 {
				assert.Equal(t, tt.wantFirst, got[0].Name)
			}
		})
	}
}

func TestToolRegistryAdapter_SearchTools_CategoryAndHints(t *testing.T) {
	t.Parallel()

	adapter := NewToolRegistryAdapter([]*agent.Tool{
		{
			Name:        "encrypt_data",
			Description: "Encrypt data using AES",
			Capability: agent.ToolCapability{
				Category:    "crypto",
				SearchHints: []string{"security", "aes256"},
			},
		},
		{
			Name:        "hash_file",
			Description: "Compute file hash",
			Capability: agent.ToolCapability{
				Category:    "crypto",
				SearchHints: []string{"sha256", "checksum"},
			},
		},
		{
			Name:        "web_search",
			Description: "Search the web",
			Capability: agent.ToolCapability{
				Category:    "network",
				SearchHints: []string{"google", "query"},
			},
		},
	})

	tests := []struct {
		give      string
		wantCount int
		wantFirst string
	}{
		{give: "crypto", wantCount: 2, wantFirst: "encrypt_data"}, // category match
		{give: "security", wantCount: 1, wantFirst: "encrypt_data"}, // hint match
		{give: "checksum", wantCount: 1, wantFirst: "hash_file"},    // hint match
		{give: "network", wantCount: 1, wantFirst: "web_search"},    // category match
		{give: "CRYPTO", wantCount: 2, wantFirst: "encrypt_data"},   // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := adapter.SearchTools(tt.give, 10)
			require.Len(t, got, tt.wantCount)
			if tt.wantCount > 0 {
				assert.Equal(t, tt.wantFirst, got[0].Name)
			}
		})
	}
}

func TestToolRegistryAdapter_SearchTools_EmptyCapability(t *testing.T) {
	t.Parallel()

	// Tools with zero-value Capability — backward compatibility
	adapter := NewToolRegistryAdapter([]*agent.Tool{
		{Name: "exec", Description: "Execute commands"},
		{Name: "read", Description: "Read files"},
	})

	tests := []struct {
		give      string
		wantCount int
		wantFirst string
	}{
		{give: "exec", wantCount: 1, wantFirst: "exec"},
		{give: "Read", wantCount: 1, wantFirst: "read"},
		{give: "missing", wantCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := adapter.SearchTools(tt.give, 10)
			require.Len(t, got, tt.wantCount)
			if tt.wantCount > 0 {
				assert.Equal(t, tt.wantFirst, got[0].Name)
			}
		})
	}
}

func TestToolRegistryAdapter_NewPopulatesFields(t *testing.T) {
	t.Parallel()

	tools := []*agent.Tool{
		{
			Name:        "fs_list",
			Description: "List files",
			Capability: agent.ToolCapability{
				Aliases:     []string{"ls", "dir"},
				Category:    "filesystem",
				SearchHints: []string{"directory", "listing"},
			},
		},
	}
	adapter := NewToolRegistryAdapter(tools)

	got := adapter.ListTools()
	require.Len(t, got, 1)
	assert.Equal(t, "fs_list", got[0].Name)
	assert.Equal(t, []string{"ls", "dir"}, got[0].Aliases)
	assert.Equal(t, "filesystem", got[0].Category)
	assert.Equal(t, []string{"directory", "listing"}, got[0].SearchHints)
}

// Verify interface compliance at compile time.
var _ knowledge.ToolRegistryProvider = (*ToolRegistryAdapter)(nil)
var _ knowledge.RuntimeContextProvider = (*RuntimeContextAdapter)(nil)
