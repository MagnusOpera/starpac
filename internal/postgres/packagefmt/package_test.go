package packagefmt

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/MagnusOpera/starpac/internal/postgres/parser"
	"github.com/MagnusOpera/starpac/internal/postgres/project"
)

func TestReadLegacyPgpkgContract(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.pgpkg")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	archive := zip.NewWriter(file)
	entries := map[string]string{
		"manifest.json": `{"formatVersion":"1","packageId":"Legacy","packageVersion":"0.5.1","postgresVersion":18,"builtAtUtc":"2026-01-01T00:00:00Z","projectFile":"legacy.pgpac","files":[]}`,
		"model.json":    `{"postgresVersion":18,"tables":[{"schema":"public","name":"widgets","sql":"CREATE TABLE public.widgets (id integer PRIMARY KEY)"}]}`,
		"project.xml":   `<PgPac ProjectVersion="1"><PropertyGroup><PackageId>Legacy</PackageId><Version>0.5.1</Version><PostgresVersion>18</PostgresVersion><DefaultSchema>public</DefaultSchema></PropertyGroup><ItemGroup><Table Include="Tables/**/*.sql" /></ItemGroup><Target><OwnedSchemas><Schema Name="public" /></OwnedSchemas><Comparison MatchPrivileges="false" MatchOwners="false" MatchComments="true" /><Plan AllowCreate="true" AllowAlter="true" AllowDrop="false" /><Apply UseTransaction="true" LockTimeout="5s" StatementTimeout="10m" StopOnDataLossRisk="true" /></Target></PgPac>`,
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
		t.Fatalf("Read() legacy pgpkg error = %v", err)
	}
	if pkg.Manifest.PackageID != "Legacy" || len(pkg.Model.Tables) != 1 {
		t.Fatalf("legacy pgpkg changed: manifest=%#v model=%#v", pkg.Manifest, pkg.Model)
	}
}

func TestWriteReadPreservesTargetConfig(t *testing.T) {
	projectPath := filepath.Join("..", "..", "..", "products", "pgpac", "testdata", "sample", "sample.pgpac")
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
	project, raw, err := projectxml.Load("../../../products/pgpac/testdata/sample/sample.pgpac")
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
