package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveFilesFiltersAndOrdersMatches(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"b.sql", "a.sql", "ignored.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("SELECT 1;"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	files, err := ResolveFiles(root, []Item{{Kind: "table", Include: "*"}}, nil)
	if err != nil {
		t.Fatalf("ResolveFiles returned error: %v", err)
	}
	if len(files) != 2 || files[0].RelPath != "a.sql" || files[1].RelPath != "b.sql" {
		t.Fatalf("unexpected files: %#v", files)
	}
}

func TestValidateBasePreservesProjectErrors(t *testing.T) {
	if err := ValidateBase("D1Pac", "", "package", "1.0.0", []Item{{}}); err == nil || err.Error() != "D1Pac ProjectVersion is required" {
		t.Fatalf("unexpected validation error: %v", err)
	}
}
