package projectxml

import "testing"

func TestLoadFromBytesParsesTargetAttributes(t *testing.T) {
	raw := []byte(`
<PgPac ProjectVersion="1">
  <PropertyGroup>
    <PackageId>SampleProject</PackageId>
    <Version>0.1.0</Version>
    <PostgresVersion>18</PostgresVersion>
    <DefaultSchema>app</DefaultSchema>
  </PropertyGroup>
  <ItemGroup>
    <Schema Include="Schemas/**/*.sql" />
  </ItemGroup>
  <Target>
    <OwnedSchemas>
      <Schema Name="app" />
    </OwnedSchemas>
    <Extensions>
      <Extension Name="pgcrypto" Version="1.3" />
    </Extensions>
  </Target>
</PgPac>`)

	project, _, err := LoadFromBytes(raw, "sample.pgpac")
	if err != nil {
		t.Fatalf("LoadFromBytes returned error: %v", err)
	}

	if got, want := project.Target.OwnedSchemaNames(), []string{"app"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("OwnedSchemaNames() = %v, want %v", got, want)
	}

	if got, want := project.Target.ExtensionNames(), []string{"pgcrypto"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("ExtensionNames() = %v, want %v", got, want)
	}

	if got, want := project.Target.Extensions[0].Version, "1.3"; got != want {
		t.Fatalf("extension version = %q, want %q", got, want)
	}
}

func TestLoadAndResolve(t *testing.T) {
	project, _, err := Load("../../testdata/sample/sample.pgpac")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	files, err := project.ResolveFiles()
	if err != nil {
		t.Fatalf("ResolveFiles() error = %v", err)
	}
	if len(files) != 9 {
		t.Fatalf("expected 9 sql files, got %d", len(files))
	}
}
