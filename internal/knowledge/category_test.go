package knowledge

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
)

func TestMapKnowledgeCategory_Shared(t *testing.T) {
	t.Parallel()
	tests := []struct {
		give    string
		wantCat entknowledge.Category
		wantErr bool
	}{
		{give: "preference", wantCat: entknowledge.CategoryPreference},
		{give: "fact", wantCat: entknowledge.CategoryFact},
		{give: "rule", wantCat: entknowledge.CategoryRule},
		{give: "definition", wantCat: entknowledge.CategoryDefinition},
		{give: "pattern", wantCat: entknowledge.CategoryPattern},
		{give: "correction", wantCat: entknowledge.CategoryCorrection},
		{give: "unknown", wantErr: true},
		{give: "", wantErr: true},
		{give: "PREFERENCE", wantErr: true},
		{give: "Fact", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got, err := MapKnowledgeCategory(tt.give)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "unrecognized knowledge type")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCat, got)
			}
		})
	}
}

func TestMapLearningCategory_Shared(t *testing.T) {
	t.Parallel()
	tests := []struct {
		give    string
		wantCat entlearning.Category
		wantErr bool
	}{
		{give: "correction", wantCat: entlearning.CategoryUserCorrection},
		{give: "pattern", wantCat: entlearning.CategoryGeneral},
		{give: "tool_error", wantCat: entlearning.CategoryToolError},
		{give: "provider_error", wantCat: entlearning.CategoryProviderError},
		{give: "timeout", wantCat: entlearning.CategoryTimeout},
		{give: "permission", wantCat: entlearning.CategoryPermission},
		{give: "unknown", wantErr: true},
		{give: "", wantErr: true},
		{give: "CORRECTION", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got, err := MapLearningCategory(tt.give)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "unrecognized learning type")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCat, got)
			}
		})
	}
}
