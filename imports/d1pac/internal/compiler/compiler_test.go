package compiler

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/MagnusOpera/d1pac/internal/projectxml"
)

func TestBuildDesiredModel(t *testing.T) {
	project, _, err := projectxml.Load(filepath.Join("..", "..", "testdata", "sample", "sample.d1pac"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	schema, err := BuildDesiredModel(context.Background(), project)
	if err != nil {
		t.Fatalf("BuildDesiredModel returned error: %v", err)
	}
	if len(schema.Tables) != 2 {
		t.Fatalf("tables = %d, want 2", len(schema.Tables))
	}
	if len(schema.Indexes) != 1 || len(schema.Views) != 1 || len(schema.Triggers) != 1 {
		t.Fatalf("unexpected schema model: %#v", schema)
	}
	if schema.SQLiteVersion == "" {
		t.Fatal("SQLiteVersion is empty")
	}
	events := schema.Tables[0]
	if events.Name != "widget_events" || len(events.ForeignKeys) != 1 {
		t.Fatalf("foreign-key introspection failed: %#v", events)
	}
	if events.Columns[0].Definition == "" {
		t.Fatal("column definition was not preserved")
	}
}
