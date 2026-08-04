package model

import "testing"

func TestNormalizeDDLIgnoresFormattingAndIdentifierQuotes(t *testing.T) {
	left := "CREATE TABLE widgets (id INTEGER PRIMARY KEY)"
	right := `create  table "widgets"(
      "id" integer primary key
    );`
	if NormalizeDDL(left) != NormalizeDDL(right) {
		t.Fatalf("normalized DDL differs: %q != %q", NormalizeDDL(left), NormalizeDDL(right))
	}
}

func TestTableConstraintsFindsOnlyTableConstraints(t *testing.T) {
	constraints := TableConstraints(`
CREATE TABLE widgets (
  id INTEGER PRIMARY KEY,
  status TEXT CHECK(status <> ''),
  CONSTRAINT widgets_status UNIQUE(status),
  FOREIGN KEY(id) REFERENCES parents(id)
) STRICT`)
	if len(constraints) != 2 {
		t.Fatalf("constraints = %#v, want 2", constraints)
	}
}

func TestReplaceCreateTableNamePreservesDefinition(t *testing.T) {
	rewritten, ok := ReplaceCreateTableName(`CREATE TABLE "widgets" (id INTEGER) STRICT`, "__new_widgets")
	if !ok {
		t.Fatal("ReplaceCreateTableName returned false")
	}
	want := `CREATE TABLE "__new_widgets" (id INTEGER) STRICT`
	if rewritten != want {
		t.Fatalf("rewritten = %q, want %q", rewritten, want)
	}
}
