package packagefmt

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/MagnusOpera/starpac/internal/d1/compiler"
	"github.com/MagnusOpera/starpac/internal/d1/project"
)

func TestReadLegacyD1pkgContract(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.d1pkg")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	archive := zip.NewWriter(file)
	entries := map[string]string{
		"manifest.json": `{"formatVersion":"1","engine":"cloudflare-d1","packageId":"Legacy","packageVersion":"0.0.3","sqliteVersion":"3.50.4","builtAtUtc":"2026-01-01T00:00:00Z","projectFile":"legacy.d1pac","files":[]}`,
		"model.json":    `{"sqliteVersion":"3.50.4","tables":[{"name":"widgets","sql":"CREATE TABLE widgets (id integer PRIMARY KEY)","columns":[{"position":0,"name":"id","type":"INTEGER","notNull":false,"primaryKey":1}]}]}`,
		"project.xml":   `<D1Pac ProjectVersion="1"><PropertyGroup><PackageId>Legacy</PackageId><Version>0.0.3</Version></PropertyGroup><ItemGroup><Table Include="Tables/**/*.sql" /></ItemGroup><Target><Plan AllowCreate="true" AllowAlter="true" AllowDrop="false" /><Apply UseTransaction="true" StopOnDataLossRisk="true" /></Target></D1Pac>`,
	}
	for name, content := range entries {
		entry, createErr := archive.Create(name)
		if createErr != nil {
			t.Fatal(createErr)
		}
		if _, writeErr := entry.Write([]byte(content)); writeErr != nil {
			t.Fatal(writeErr)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	pkg, err := Read(path)
	if err != nil {
		t.Fatalf("Read() legacy d1pkg error = %v", err)
	}
	if pkg.Manifest.PackageID != "Legacy" || len(pkg.Model.Tables) != 1 {
		t.Fatalf("legacy d1pkg changed: manifest=%#v model=%#v", pkg.Manifest, pkg.Model)
	}
}

func TestWriteReadRoundTrip(t *testing.T) {
	project, rawProject, err := projectxml.Load(filepath.Join("..", "..", "..", "products", "d1pac", "testdata", "sample", "sample.d1pac"))
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
