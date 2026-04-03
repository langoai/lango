package settings

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExperimentalCategories_MatchesExpected verifies that the
// ExperimentalCategories map contains exactly the intended set.
// If a category is added or removed, this test will fail —
// prompting a deliberate update to the map.
func TestExperimentalCategories_MatchesExpected(t *testing.T) {
	t.Parallel()

	expected := []string{
		"a2a",
		"agent_memory",
		"alerting",
		"economy",
		"economy_escrow",
		"economy_escrow_onchain",
		"economy_negotiation",
		"economy_pricing",
		"economy_risk",
		"graph",
		"hooks",
		"librarian",
		"multi_agent",
		"observability",
		"ontology",
		"os_sandbox",
		"p2p",
		"p2p_owner",
		"p2p_pricing",
		"p2p_sandbox",
		"p2p_workspace",
		"p2p_zkp",
		"provenance",
		"runledger",
		"smartaccount",
		"smartaccount_modules",
		"smartaccount_paymaster",
		"smartaccount_session",
	}

	var got []string
	for id := range ExperimentalCategories {
		got = append(got, id)
	}
	sort.Strings(got)

	assert.Equal(t, expected, got,
		"ExperimentalCategories drift detected — update the map or this test")
}
