package diff

import (
	"strings"
	"testing"

	"github.com/MagnusOpera/d1pac/internal/model"
	"github.com/MagnusOpera/d1pac/internal/projectxml"
)

func TestBuildPlanCreatesAndBlocksDrop(t *testing.T) {
	project := writableProject()
	desired := &model.SchemaModel{
		Tables: []model.TableDef{table("widgets", column("id", "INTEGER", true, 1, "id INTEGER PRIMARY KEY"))},
	}
	actual := &model.SchemaModel{
		Tables: []model.TableDef{table("legacy", column("id", "INTEGER", true, 1, "id INTEGER PRIMARY KEY"))},
	}
	plan := BuildPlan(project, desired, actual, Options{})
	if len(plan.Operations) != 2 {
		t.Fatalf("operations = %#v, want 2", plan.Operations)
	}
	if plan.Summary.Supported {
		t.Fatal("plan with a blocked drop must not be supported")
	}
	assertHasKind(t, plan, "create-table")
	assertHasKind(t, plan, "blocked-drop-table")
}

func TestBuildPlanAddsTrailingColumn(t *testing.T) {
	project := writableProject()
	desired := table(
		"widgets",
		column("id", "INTEGER", true, 1, "id INTEGER PRIMARY KEY"),
		columnWithDefault("name", "TEXT", true, "''", "name TEXT NOT NULL DEFAULT ''"),
	)
	actual := table("widgets", column("id", "INTEGER", true, 1, "id INTEGER PRIMARY KEY"))
	plan := BuildPlan(project, &model.SchemaModel{Tables: []model.TableDef{desired}}, &model.SchemaModel{Tables: []model.TableDef{actual}}, Options{})
	if len(plan.Operations) != 1 || plan.Operations[0].Kind != "alter-table-add-column" {
		t.Fatalf("unexpected operations: %#v", plan.Operations)
	}
	if !strings.Contains(plan.Operations[0].SQL, `ADD COLUMN name TEXT NOT NULL DEFAULT ''`) {
		t.Fatalf("unexpected SQL: %s", plan.Operations[0].SQL)
	}
}

func TestBuildPlanRebuildPreservesCommonColumnsAndDependentObjects(t *testing.T) {
	project := writableProject()
	desiredTable := table(
		"widgets",
		column("id", "INTEGER", true, 1, "id INTEGER PRIMARY KEY"),
		column("name", "TEXT", false, 0, "name TEXT"),
	)
	actualTable := table(
		"widgets",
		column("id", "INTEGER", true, 1, "id INTEGER PRIMARY KEY"),
		column("name", "INTEGER", false, 0, "name INTEGER"),
	)
	desired := &model.SchemaModel{
		Tables:   []model.TableDef{desiredTable},
		Indexes:  []model.IndexDef{{Name: "idx_widgets_name", TableName: "widgets", SQL: "CREATE INDEX idx_widgets_name ON widgets(name)"}},
		Triggers: []model.TriggerDef{{Name: "widgets_changed", TableName: "widgets", SQL: "CREATE TRIGGER widgets_changed AFTER UPDATE ON widgets BEGIN SELECT 1; END"}},
	}
	actual := &model.SchemaModel{
		Tables:   []model.TableDef{actualTable},
		Indexes:  desired.Indexes,
		Triggers: desired.Triggers,
	}
	plan := BuildPlan(project, desired, actual, Options{})
	if len(plan.Operations) != 1 || plan.Operations[0].Kind != "rebuild-table" {
		t.Fatalf("unexpected operations: %#v", plan.Operations)
	}
	sql := plan.Operations[0].SQL
	for _, expected := range []string{"INSERT INTO", "DROP TABLE", "RENAME TO", "CREATE INDEX", "CREATE TRIGGER"} {
		if !strings.Contains(sql, expected) {
			t.Fatalf("rebuild SQL does not contain %q: %s", expected, sql)
		}
	}
}

func TestBuildPlanBlocksDestructiveRebuild(t *testing.T) {
	project := writableProject()
	desired := table("widgets", column("id", "INTEGER", true, 1, "id INTEGER PRIMARY KEY"))
	actual := table(
		"widgets",
		column("id", "INTEGER", true, 1, "id INTEGER PRIMARY KEY"),
		column("obsolete", "TEXT", false, 0, "obsolete TEXT"),
	)
	plan := BuildPlan(project, &model.SchemaModel{Tables: []model.TableDef{desired}}, &model.SchemaModel{Tables: []model.TableDef{actual}}, Options{})
	if len(plan.Operations) != 1 || plan.Operations[0].Kind != "blocked-rebuild-table" {
		t.Fatalf("unexpected operations: %#v", plan.Operations)
	}
	if !plan.Summary.Destructive {
		t.Fatal("destructive rebuild was not reported")
	}
}

func TestBuildPlanBlocksRebuildOfReferencedTable(t *testing.T) {
	project := writableProject()
	desiredParent := table("widgets", column("id", "TEXT", true, 1, "id TEXT PRIMARY KEY"))
	actualParent := table("widgets", column("id", "INTEGER", true, 1, "id INTEGER PRIMARY KEY"))
	child := table("events", column("widget_id", "INTEGER", true, 0, "widget_id INTEGER NOT NULL"))
	child.ForeignKeys = []model.ForeignKeyDef{{
		Table:    "widgets",
		From:     "widget_id",
		To:       "id",
		OnUpdate: "NO ACTION",
		OnDelete: "CASCADE",
		Match:    "NONE",
	}}
	plan := BuildPlan(
		project,
		&model.SchemaModel{Tables: []model.TableDef{child, desiredParent}},
		&model.SchemaModel{Tables: []model.TableDef{child, actualParent}},
		Options{AllowDrop: true},
	)
	if len(plan.Operations) != 1 || plan.Operations[0].Kind != "blocked-rebuild-table" {
		t.Fatalf("unexpected operations: %#v", plan.Operations)
	}
	if !strings.Contains(plan.Operations[0].SQL, "another table references") {
		t.Fatalf("unexpected blocked SQL: %s", plan.Operations[0].SQL)
	}
}

func TestBuildPlanEquivalentModelsHaveNoOperations(t *testing.T) {
	project := writableProject()
	desired := table("widgets", column("id", "INTEGER", true, 1, "id INTEGER PRIMARY KEY"))
	actual := desired
	actual.SQL = `CREATE TABLE "widgets" ( id integer primary key )`
	plan := BuildPlan(project, &model.SchemaModel{Tables: []model.TableDef{desired}}, &model.SchemaModel{Tables: []model.TableDef{actual}}, Options{})
	if len(plan.Operations) != 0 {
		t.Fatalf("unexpected operations: %#v", plan.Operations)
	}
}

func writableProject() *projectxml.Project {
	return &projectxml.Project{Target: projectxml.TargetConfig{
		Plan: projectxml.PlanConfig{AllowCreate: true, AllowAlter: true},
	}}
}

func table(name string, columns ...model.ColumnDef) model.TableDef {
	definitions := make([]string, 0, len(columns))
	for _, item := range columns {
		definitions = append(definitions, item.Definition)
	}
	return model.TableDef{
		Name:    name,
		SQL:     "CREATE TABLE " + name + " (" + strings.Join(definitions, ", ") + ")",
		Columns: columns,
	}
}

func column(name, dataType string, notNull bool, primaryKey int, definition string) model.ColumnDef {
	return model.ColumnDef{
		Position:   0,
		Name:       name,
		Type:       dataType,
		NotNull:    notNull,
		PrimaryKey: primaryKey,
		Definition: definition,
	}
}

func columnWithDefault(name, dataType string, notNull bool, defaultSQL, definition string) model.ColumnDef {
	column := column(name, dataType, notNull, 0, definition)
	column.DefaultSQL = &defaultSQL
	column.Position = 1
	return column
}

func assertHasKind(t *testing.T, plan Plan, kind string) {
	t.Helper()
	for _, operation := range plan.Operations {
		if operation.Kind == kind {
			return
		}
	}
	t.Fatalf("plan does not contain %s: %#v", kind, plan.Operations)
}
