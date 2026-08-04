package introspect

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

type fixtureQueryer struct{}

func (fixtureQueryer) Query(_ context.Context, sql string) ([]map[string]any, error) {
	switch {
	case strings.Contains(sql, "FROM sqlite_schema"):
		return []map[string]any{
			{"type": "table", "name": "widgets", "tbl_name": "widgets", "sql": "CREATE TABLE widgets (id INTEGER PRIMARY KEY, parent_id INTEGER REFERENCES widgets(id))"},
			{"type": "table", "name": "d1_migrations", "tbl_name": "d1_migrations", "sql": "CREATE TABLE d1_migrations(id INTEGER)"},
			{"type": "index", "name": "idx_widgets_parent", "tbl_name": "widgets", "sql": "CREATE INDEX idx_widgets_parent ON widgets(parent_id)"},
			{"type": "view", "name": "widget_ids", "tbl_name": "widget_ids", "sql": "CREATE VIEW widget_ids AS SELECT id FROM widgets"},
			{"type": "trigger", "name": "widgets_changed", "tbl_name": "widgets", "sql": "CREATE TRIGGER widgets_changed AFTER UPDATE ON widgets BEGIN SELECT 1; END"},
		}, nil
	case strings.Contains(sql, "table_xinfo"):
		return []map[string]any{
			{"cid": float64(0), "name": "id", "type": "INTEGER", "notnull": float64(0), "dflt_value": nil, "pk": float64(1), "hidden": float64(0)},
			{"cid": float64(1), "name": "parent_id", "type": "INTEGER", "notnull": float64(0), "dflt_value": nil, "pk": float64(0), "hidden": float64(0)},
		}, nil
	case strings.Contains(sql, "foreign_key_list"):
		return []map[string]any{{
			"id": float64(0), "seq": float64(0), "table": "widgets", "from": "parent_id", "to": "id",
			"on_update": "NO ACTION", "on_delete": "NO ACTION", "match": "NONE",
		}}, nil
	default:
		return nil, fmt.Errorf("unexpected query %q", sql)
	}
}

func TestLoadRemoteBuildsSemanticSchema(t *testing.T) {
	schema, err := LoadRemote(context.Background(), fixtureQueryer{}, func(objectType, name string) bool {
		return objectType == "table" && name == "d1_migrations"
	})
	if err != nil {
		t.Fatalf("LoadRemote returned error: %v", err)
	}
	if schema.SQLiteVersion != "" {
		t.Fatalf("SQLiteVersion = %q", schema.SQLiteVersion)
	}
	if len(schema.Tables) != 1 || len(schema.Tables[0].Columns) != 2 || len(schema.Tables[0].ForeignKeys) != 1 {
		t.Fatalf("unexpected table model: %#v", schema.Tables)
	}
	if len(schema.Indexes) != 1 || len(schema.Views) != 1 || len(schema.Triggers) != 1 {
		t.Fatalf("unexpected secondary objects: %#v", schema)
	}
	if schema.Tables[0].Columns[1].Definition != "parent_id INTEGER REFERENCES widgets(id)" {
		t.Fatalf("unexpected column definition: %q", schema.Tables[0].Columns[1].Definition)
	}
}
