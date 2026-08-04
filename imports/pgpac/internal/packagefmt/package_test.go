package packagefmt

import (
	"path/filepath"
	"testing"

	"github.com/MagnusOpera/pgpac/internal/parser"
	"github.com/MagnusOpera/pgpac/internal/projectxml"
)

func TestWriteReadPreservesTargetConfig(t *testing.T) {
	projectPath := filepath.Join("..", "..", "testdata", "sample", "sample.pgpac")
	project, rawXML, err := projectxml.Load(projectPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	model, err := parser.BuildDesiredModel(project)
	if err != nil {
		t.Fatalf("BuildDesiredModel returned error: %v", err)
	}

	path := filepath.Join(t.TempDir(), "sample.pgpkg")
	if err := Write(path, NewManifest(project, model), model, rawXML, project); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	pkg, err := Read(path)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}

	if got, want := pkg.Project.Target.OwnedSchemaNames(), []string{"app"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("OwnedSchemaNames() = %v, want %v", got, want)
	}

	if got, want := pkg.Project.Target.ExtensionNames(), []string{"pgcrypto"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("ExtensionNames() = %v, want %v", got, want)
	}

	if got, want := pkg.Project.Target.Extensions[0].Version, "1.3"; got != want {
		t.Fatalf("extension version = %q, want %q", got, want)
	}
}

func TestWriteReadRoundTrip(t *testing.T) {
	project, raw, err := projectxml.Load("../../testdata/sample/sample.pgpac")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	model, err := parser.BuildDesiredModel(project)
	if err != nil {
		t.Fatalf("BuildDesiredModel() error = %v", err)
	}
	path := filepath.Join(t.TempDir(), "sample.pgpkg")
	if err := Write(path, NewManifest(project, model), model, raw, project); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	pkg, err := Read(path)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if pkg.Manifest.PackageID != "SampleProject" {
		t.Fatalf("unexpected package id %q", pkg.Manifest.PackageID)
	}
	if len(pkg.Model.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(pkg.Model.Tables))
	}
}
