package ontology

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type exportSchemaDigestTestRegistry struct{}

func (exportSchemaDigestTestRegistry) RegisterType(context.Context, ObjectType) error {
	return errors.New("unexpected RegisterType call")
}

func (exportSchemaDigestTestRegistry) GetType(context.Context, string) (*ObjectType, error) {
	return nil, errors.New("unexpected GetType call")
}

func (exportSchemaDigestTestRegistry) ListTypes(context.Context) ([]ObjectType, error) {
	return []ObjectType{{Name: "LocalType", Status: SchemaActive}}, nil
}

func (exportSchemaDigestTestRegistry) DeprecateType(context.Context, string) error {
	return errors.New("unexpected DeprecateType call")
}

func (exportSchemaDigestTestRegistry) RegisterPredicate(context.Context, PredicateDefinition) error {
	return errors.New("unexpected RegisterPredicate call")
}

func (exportSchemaDigestTestRegistry) GetPredicate(context.Context, string) (*PredicateDefinition, error) {
	return nil, errors.New("unexpected GetPredicate call")
}

func (exportSchemaDigestTestRegistry) ListPredicates(context.Context) ([]PredicateDefinition, error) {
	return nil, nil
}

func (exportSchemaDigestTestRegistry) DeprecatePredicate(context.Context, string) error {
	return errors.New("unexpected DeprecatePredicate call")
}

func (exportSchemaDigestTestRegistry) UpdateTypeStatus(context.Context, string, SchemaStatus) error {
	return errors.New("unexpected UpdateTypeStatus call")
}

func (exportSchemaDigestTestRegistry) UpdatePredicateStatus(context.Context, string, SchemaStatus) error {
	return errors.New("unexpected UpdatePredicateStatus call")
}

func TestExportSchema_ReturnsDigestErrorInsteadOfPanicking(t *testing.T) {
	original := marshalDigestPayload
	marshalDigestPayload = func(any) ([]byte, error) {
		return nil, errors.New("marshal unavailable")
	}
	t.Cleanup(func() { marshalDigestPayload = original })

	var gotErr error
	require.NotPanics(t, func() {
		_, gotErr = exportSchema(context.Background(), exportSchemaDigestTestRegistry{}, 1, "local")
	})
	require.Error(t, gotErr)
	require.Contains(t, gotErr.Error(), "compute schema digest")
	require.Contains(t, gotErr.Error(), "marshal digest payload")
}
