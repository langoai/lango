package learning

import (
	"testing"

	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapKnowledgeCategory(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, err := mapKnowledgeCategory(tt.give)
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
