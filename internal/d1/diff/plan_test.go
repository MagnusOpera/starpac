package diff

import (
	"strings"
	"testing"

	"github.com/MagnusOpera/starpac/internal/d1/model"
	"github.com/MagnusOpera/starpac/internal/d1/project"
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

func TestBuildPlanAppendsAddableColumnRegardlessOfDeclaredOrder(t *testing.T) {
	project := writableProject()
	desired := table(
		"widgets",
		column("id", "INTEGER", true, 1, "id INTEGER PRIMARY KEY"),
		column("description", "TEXT", false, 0, "description TEXT"),
		column("created_at", "TEXT", true, 0, "created_at TEXT NOT NULL"),
	)
	actual := table(
		"widgets",
		column("id", "INTEGER", true, 1, "id INTEGER PRIMARY KEY"),
		column("created_at", "TEXT", true, 0, "created_at TEXT NOT NULL"),
	)

	plan := BuildPlan(
		project,
		&model.SchemaModel{Tables: []model.TableDef{desired}},
		&model.SchemaModel{Tables: []model.TableDef{actual}},
		Options{},
	)

	if !plan.Summary.Supported || len(plan.Operations) != 1 {
		t.Fatalf("operations = %#v, want one supported addition", plan.Operations)
	}
	operation := plan.Operations[0]
	if operation.Kind != "alter-table-add-column" || operation.Risk != "safe" {
		t.Fatalf("operation = %#v, want safe column addition", operation)
	}
	if operation.SQL != `ALTER TABLE "widgets" ADD COLUMN description TEXT;` {
		t.Fatalf("unexpected SQL: %s", operation.SQL)
	}
}

func TestBuildPlanIgnoresColumnOrderAfterAppend(t *testing.T) {
	project := writableProject()
	desired := table(
		"widgets",
		positionedColumn(0, "id", "INTEGER", true, 1, "id INTEGER PRIMARY KEY"),
		positionedColumn(1, "description", "TEXT", false, 0, "description TEXT"),
		positionedColumn(2, "created_at", "TEXT", true, 0, "created_at TEXT NOT NULL"),
	)
	actual := table(
		"widgets",
		positionedColumn(0, "id", "INTEGER", true, 1, "id INTEGER PRIMARY KEY"),
		positionedColumn(1, "created_at", "TEXT", true, 0, "created_at TEXT NOT NULL"),
		positionedColumn(2, "description", "TEXT", false, 0, "description TEXT"),
	)

	plan := BuildPlan(
		project,
		&model.SchemaModel{Tables: []model.TableDef{desired}},
		&model.SchemaModel{Tables: []model.TableDef{actual}},
		Options{},
	)

	if len(plan.Operations) != 0 {
		t.Fatalf("column order produced drift: %#v", plan.Operations)
	}

	strictPlan := BuildPlan(
		project,
		&model.SchemaModel{Tables: []model.TableDef{desired}},
		&model.SchemaModel{Tables: []model.TableDef{actual}},
		Options{Strict: true},
	)
	if len(strictPlan.Operations) != 1 || strictPlan.Operations[0].Kind != "rebuild-table" {
		t.Fatalf("strict operations = %#v, want a table rebuild", strictPlan.Operations)
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

func TestBuildPlanRebuildRefreshesTriggersThatReferenceTheTable(t *testing.T) {
	project := writableProject()
	desiredTable := table(
		"tenant_databases",
		column("id", "TEXT", true, 1, "id TEXT PRIMARY KEY"),
		column("status", "TEXT", true, 0, "status TEXT NOT NULL"),
	)
	actualTable := table(
		"tenant_databases",
		column("id", "TEXT", true, 1, "id TEXT PRIMARY KEY"),
		column("status", "INTEGER", true, 0, "status INTEGER NOT NULL"),
	)
	workspaceProfiles := table(
		"workspace_profiles",
		column("id", "TEXT", true, 1, "id TEXT PRIMARY KEY"),
	)
	dependentTrigger := model.TriggerDef{
		Name:      "require_assigned_tenant",
		TableName: "workspace_profiles",
		SQL:       "CREATE TRIGGER require_assigned_tenant BEFORE INSERT ON workspace_profiles WHEN NOT EXISTS (SELECT 1 FROM tenant_databases) BEGIN SELECT RAISE(ABORT, 'tenant required'); END",
	}
	desired := &model.SchemaModel{
		Tables:   []model.TableDef{desiredTable, workspaceProfiles},
		Triggers: []model.TriggerDef{dependentTrigger},
	}
	actual := &model.SchemaModel{
		Tables:   []model.TableDef{actualTable, workspaceProfiles},
		Triggers: []model.TriggerDef{dependentTrigger},
	}

	plan := BuildPlan(project, desired, actual, Options{})
	if len(plan.Operations) != 1 || plan.Operations[0].Kind != "rebuild-table" {
		t.Fatalf("unexpected operations: %#v", plan.Operations)
	}
	sql := plan.Operations[0].SQL
	dropTrigger := strings.Index(sql, `DROP TRIGGER "require_assigned_tenant"`)
	dropTable := strings.Index(sql, `DROP TABLE "tenant_databases"`)
	createTrigger := strings.LastIndex(sql, "CREATE TRIGGER require_assigned_tenant")
	if dropTrigger < 0 || dropTable < 0 || createTrigger < 0 || !(dropTrigger < dropTable && dropTable < createTrigger) {
		t.Fatalf("dependent trigger was not refreshed around table rebuild: %s", sql)
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

func TestBuildPlanAllowsTransactionalRebuildOfReferencedTable(t *testing.T) {
	project := writableProject()
	project.Target.Apply.UseTransaction = true
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
	if len(plan.Operations) != 1 || plan.Operations[0].Kind != "rebuild-table" {
		t.Fatalf("unexpected operations: %#v", plan.Operations)
	}
	if !plan.Summary.Supported {
		t.Fatal("transactional referenced-table rebuild must be supported")
	}
	for _, expected := range []string{
		`ALTER TABLE "widgets" RENAME TO "__d1pac_widgets_old"`,
		`ALTER TABLE "events" RENAME TO "__d1pac_events_old"`,
		`DROP TABLE "__d1pac_events_old"`,
		`DROP TABLE "__d1pac_widgets_old"`,
	} {
		if !strings.Contains(plan.Operations[0].SQL, expected) {
			t.Fatalf("transactional rebuild SQL does not contain %q: %s", expected, plan.Operations[0].SQL)
		}
	}
}

func TestBuildPlanBlocksNonTransactionalRebuildOfReferencedTable(t *testing.T) {
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
	if !strings.Contains(plan.Operations[0].Reason, `requires transactional apply because retained table(s) "events" reference this table`) {
		t.Fatalf("unexpected block reason: %s", plan.Operations[0].Reason)
	}
}

func TestBuildPlanExplainsBlockedNonTrailingColumnAddition(t *testing.T) {
	project := writableProject()
	desiredParent := table(
		"artifacts",
		column("id", "TEXT", true, 1, "id TEXT PRIMARY KEY"),
		column("build_duration_ms", "INTEGER", false, 0, "build_duration_ms INTEGER"),
		column("created_at", "TEXT", true, 0, "created_at TEXT NOT NULL"),
	)
	actualParent := table(
		"artifacts",
		column("id", "TEXT", true, 1, "id TEXT PRIMARY KEY"),
		column("created_at", "TEXT", true, 0, "created_at TEXT NOT NULL"),
	)
	child := table(
		"build_artifacts",
		column("artifact_id", "TEXT", true, 0, "artifact_id TEXT NOT NULL"),
	)
	child.ForeignKeys = []model.ForeignKeyDef{{
		Table:    "artifacts",
		From:     "artifact_id",
		To:       "id",
		OnUpdate: "NO ACTION",
		OnDelete: "CASCADE",
		Match:    "NONE",
	}}

	plan := BuildPlan(
		project,
		&model.SchemaModel{Tables: []model.TableDef{child, desiredParent}},
		&model.SchemaModel{Tables: []model.TableDef{child, actualParent}},
		Options{AllowDrop: true, Strict: true},
	)

	if len(plan.Operations) != 1 {
		t.Fatalf("operations = %#v, want one blocked rebuild", plan.Operations)
	}
	operation := plan.Operations[0]
	if operation.Kind != "blocked-rebuild-table" || operation.Risk != "migration" {
		t.Fatalf("operation = %#v, want blocked non-destructive migration", operation)
	}
	if plan.Summary.Destructive {
		t.Fatal("non-trailing column addition must not be classified as destructive")
	}
	for _, expected := range []string{
		"non-destructive migration",
		`"build_duration_ms"`,
		"declared order",
		`"build_artifacts"`,
	} {
		if !strings.Contains(operation.Reason, expected) {
			t.Fatalf("reason %q does not contain %q", operation.Reason, expected)
		}
	}
}

func TestBuildPlanRebuildsTableAfterDroppingReferencingTable(t *testing.T) {
	project := writableProject()
	desiredParent := table("widgets", column("id", "TEXT", true, 1, "id TEXT PRIMARY KEY"))
	actualParent := table("widgets", column("id", "INTEGER", true, 1, "id INTEGER PRIMARY KEY"))
	legacyChild := table("legacy_events", column("widget_id", "INTEGER", true, 0, "widget_id INTEGER NOT NULL"))
	legacyChild.ForeignKeys = []model.ForeignKeyDef{{
		Table:    "widgets",
		From:     "widget_id",
		To:       "id",
		OnUpdate: "NO ACTION",
		OnDelete: "CASCADE",
		Match:    "NONE",
	}}
	legacyTrigger := model.TriggerDef{
		Name:      "legacy_widget_cleanup",
		TableName: "legacy_events",
		SQL:       "CREATE TRIGGER legacy_widget_cleanup AFTER DELETE ON legacy_events BEGIN SELECT id FROM widgets; END",
	}

	plan := BuildPlan(
		project,
		&model.SchemaModel{Tables: []model.TableDef{desiredParent}},
		&model.SchemaModel{
			Tables:   []model.TableDef{actualParent, legacyChild},
			Triggers: []model.TriggerDef{legacyTrigger},
		},
		Options{AllowDrop: true},
	)

	if !plan.Summary.Supported {
		t.Fatalf("plan must be supported when the referencing table is dropped: %#v", plan.Operations)
	}
	if len(plan.Operations) != 3 {
		t.Fatalf("operations = %#v, want trigger drop, table drop, and rebuild", plan.Operations)
	}
	if plan.Operations[0].Kind != "drop-trigger" || plan.Operations[0].ObjectKey != "legacy_widget_cleanup" {
		t.Fatalf("obsolete trigger must be dropped first: %#v", plan.Operations)
	}
	if plan.Operations[1].Kind != "drop-table" || plan.Operations[1].ObjectKey != "legacy_events" {
		t.Fatalf("referencing table must be dropped first: %#v", plan.Operations)
	}
	if plan.Operations[2].Kind != "rebuild-table" || plan.Operations[2].ObjectKey != "widgets" {
		t.Fatalf("referenced table must be rebuilt after the drop: %#v", plan.Operations)
	}
	if strings.Contains(plan.Operations[2].SQL, "legacy_widget_cleanup") {
		t.Fatalf("rebuild must not refresh a trigger owned by a dropped table: %s", plan.Operations[2].SQL)
	}
}

func TestBuildPlanEquivalentModelsHaveNoOperations(t *testing.T) {
	project := writableProject()
	desired := table("widgets", column("id", "INTEGER", true, 1, "id INTEGER PRIMARY KEY"))
	actual := desired
	actual.SQL = `CREATE TABLE "widgets" ( id integer primary key )`
	plan := BuildPlan(project, &model.SchemaModel{Tables: []model.TableDef{desired}}, &model.SchemaModel{Tables: []model.TableDef{actual}}, Options{})
	if plan.Operations == nil {
		t.Fatal("operations must serialize as an empty array for a no-op plan")
	}
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

func positionedColumn(position int, name, dataType string, notNull bool, primaryKey int, definition string) model.ColumnDef {
	column := column(name, dataType, notNull, primaryKey, definition)
	column.Position = position
	return column
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
