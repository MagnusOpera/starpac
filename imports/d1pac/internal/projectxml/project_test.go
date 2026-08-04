package projectxml

import (
	"path/filepath"
	"testing"
)

func TestLoadAndResolveFiles(t *testing.T) {
	project, _, err := Load(filepath.Join("..", "..", "testdata", "sample", "sample.d1pac"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if project.PackageID != "SampleProject" {
		t.Fatalf("PackageID = %q, want SampleProject", project.PackageID)
	}
	if !project.Target.Plan.AllowCreate || !project.Target.Plan.AllowAlter || project.Target.Plan.AllowDrop {
		t.Fatalf("unexpected plan configuration: %#v", project.Target.Plan)
	}
	files, err := project.ResolveFiles()
	if err != nil {
		t.Fatalf("ResolveFiles returned error: %v", err)
	}
	if len(files) != 4 {
		t.Fatalf("ResolveFiles returned %d files, want 4", len(files))
	}
	if files[0].Kind != "table" || files[len(files)-1].Kind != "trigger" {
		t.Fatalf("unexpected dependency order: %#v", files)
	}
	if !project.IsIgnored("table", "d1_migrations") {
		t.Fatal("expected built-in migration table to be ignored")
	}
	if !project.IsIgnored("table", "application_metadata") {
		t.Fatal("expected configured table to be ignored")
	}
}

func TestLoadRejectsWrongRootElement(t *testing.T) {
	_, _, err := LoadFromBytes([]byte(`<PgPac ProjectVersion="1"></PgPac>`), "bad.d1pac")
	if err == nil {
		t.Fatal("LoadFromBytes returned nil error")
	}
}
