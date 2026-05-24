package schemautil

import (
	"context"
	"errors"
	"testing"

	"github.com/langoai/lango/internal/ent"
	_ "github.com/mattn/go-sqlite3"
)

func TestRunExclusivePropagatesCallbackError(t *testing.T) {
	wantErr := errors.New("schema unavailable")

	if err := RunExclusive(func() error { return wantErr }); !errors.Is(err, wantErr) {
		t.Fatalf("RunExclusive() error = %v, want %v", err, wantErr)
	}
}

func TestCreateSchemaNilClientIsNoop(t *testing.T) {
	if err := CreateSchema(context.Background(), nil); err != nil {
		t.Fatalf("CreateSchema(nil) returned error: %v", err)
	}
}

func TestCreateSchemaCreatesEntSchema(t *testing.T) {
	client, err := ent.Open("sqlite3", "file:schemautil?mode=memory&cache=shared&_fk=1")
	if err != nil {
		t.Fatalf("open ent client: %v", err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Fatalf("close ent client: %v", err)
		}
	})

	if err := CreateSchema(context.Background(), client); err != nil {
		t.Fatalf("CreateSchema() returned error: %v", err)
	}
}
