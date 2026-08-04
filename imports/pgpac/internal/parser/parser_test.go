package parser

import (
	"testing"

	"github.com/MagnusOpera/pgpac/internal/projectxml"
)

func TestBuildDesiredModel(t *testing.T) {
	project, _, err := projectxml.Load("../../testdata/sample/sample.pgpac")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	model, err := BuildDesiredModel(project)
	if err != nil {
		t.Fatalf("BuildDesiredModel() error = %v", err)
	}
	if len(model.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(model.Tables))
	}
	if len(model.Views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(model.Views))
	}
	if len(model.Routines) != 1 {
		t.Fatalf("expected 1 routine, got %d", len(model.Routines))
	}
}
