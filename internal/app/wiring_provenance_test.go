package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/provenance"
)

func TestWireProvenanceRuntimeShortCircuitsWhenProvenanceUnavailable(t *testing.T) {
	tests := []struct {
		name     string
		resolved interface{}
	}{
		{
			name:     "missing provenance values",
			resolved: nil,
		},
		{
			name:     "missing attribution service",
			resolved: &provenanceValues{bundle: validProvenanceRuntimeValues().bundle},
		},
		{
			name:     "missing bundle service",
			resolved: &provenanceValues{attribution: validProvenanceRuntimeValues().attribution},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &recordingAppResolver{
				values: map[appinit.Provides]interface{}{
					appinit.ProvidesProvenance: tt.resolved,
				},
			}

			require.NotPanics(t, func() {
				wireProvenanceRuntime(&App{}, resolver)
			})
			assert.Equal(t, []appinit.Provides{appinit.ProvidesProvenance}, resolver.keys)
		})
	}
}

func TestWireProvenanceRuntimeResolvesOptionalWorkspaceAndP2PAfterValidProvenance(t *testing.T) {
	resolver := &recordingAppResolver{
		values: map[appinit.Provides]interface{}{
			appinit.ProvidesProvenance: validProvenanceRuntimeValues(),
			appinit.ProvidesWorkspace:  "not workspace components",
		},
	}

	require.NotPanics(t, func() {
		wireProvenanceRuntime(&App{}, resolver)
	})

	assert.Equal(t, []appinit.Provides{
		appinit.ProvidesProvenance,
		appinit.ProvidesWorkspace,
		appinit.ProvidesP2P,
	}, resolver.keys)
}

func validProvenanceRuntimeValues() *provenanceValues {
	checkpoints := provenance.NewMemoryStore()
	treeStore := provenance.NewMemoryTreeStore()
	attributionStore := provenance.NewMemoryAttributionStore()
	attribution := provenance.NewAttributionService(attributionStore, checkpoints, nil)

	return &provenanceValues{
		attribution: attribution,
		bundle:      provenance.NewBundleService(checkpoints, treeStore, attributionStore, attribution, nil),
	}
}

type recordingAppResolver struct {
	values map[appinit.Provides]interface{}
	keys   []appinit.Provides
}

func (r *recordingAppResolver) Resolve(key appinit.Provides) interface{} {
	r.keys = append(r.keys, key)
	return r.values[key]
}
