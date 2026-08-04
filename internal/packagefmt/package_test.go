package packagefmt

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/MagnusOpera/d1pac/internal/compiler"
	"github.com/MagnusOpera/d1pac/internal/projectxml"
)

func TestWriteReadRoundTrip(t *testing.T) {
	project, rawProject, err := projectxml.Load(filepath.Join("..", "..", "testdata", "sample", "sample.d1pac"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	schema, err := compiler.BuildDesiredModel(context.Background(), project)
	if err != nil {
		t.Fatalf("BuildDesiredModel returned error: %v", err)
	}
	path := filepath.Join(t.TempDir(), "sample.d1pkg")
	if err := Write(path, NewManifest(project, schema), schema, rawProject, project); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	pkg, err := Read(path)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if pkg.Manifest.Engine != "cloudflare-d1" || pkg.Manifest.PackageID != "SampleProject" {
		t.Fatalf("unexpected manifest: %#v", pkg.Manifest)
	}
	if len(pkg.Model.Tables) != 2 || len(pkg.Scripts) != 4 {
		t.Fatalf("unexpected package contents: tables=%d scripts=%d", len(pkg.Model.Tables), len(pkg.Scripts))
	}
	if !pkg.Project.Target.Apply.UseTransaction {
		t.Fatal("target configuration was not preserved")
	}
}
