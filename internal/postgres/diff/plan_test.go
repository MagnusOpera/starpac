package diff

import (
	"strings"
	"testing"

	"github.com/MagnusOpera/starpac/internal/postgres/model"
	"github.com/MagnusOpera/starpac/internal/postgres/project"
)

func TestBuildPlanCreateAndDrop(t *testing.T) {
	project := &projectxml.Project{
		Target: projectxml.TargetConfig{
			Plan: projectxml.PlanConfig{AllowDrop: false},
		},
	}
	desired := &model.SchemaModel{
		Tables: []model.TableDef{{Schema: "app", Name: "widgets", SQL: "CREATE TABLE app.widgets (id uuid)"}},
	}
	actual := &model.SchemaModel{
		Tables: []model.TableDef{{Schema: "app", Name: "legacy", SQL: "CREATE TABLE app.legacy (id uuid)"}},
	}
	plan := BuildPlan(project, desired, actual, Options{})
	if len(plan.Operations) != 2 {
		t.Fatalf("expected 2 operations, got %d", len(plan.Operations))
	}
	if plan.Operations[0].Kind != "blocked-drop-table" && plan.Operations[1].Kind != "blocked-drop-table" {
		t.Fatalf("expected blocked drop operation, got %#v", plan.Operations)
	}
}

func TestBuildPlanTreatsEquivalentTableDefinitionsAsEqual(t *testing.T) {
	project := &projectxml.Project{}
	desired := &model.SchemaModel{
		Tables: []model.TableDef{{
			Schema: "app",
			Name:   "widgets",
			SQL:    "CREATE TABLE app.widgets (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), name app.widget_name, status app.widget_status NOT NULL DEFAULT 'new')",
		}},
	}
	actual := &model.SchemaModel{
		Tables: []model.TableDef{{
			Schema: "app",
			Name:   "widgets",
			SQL:    `CREATE TABLE "app"."widgets" ("id" uuid DEFAULT gen_random_uuid() NOT NULL, "name" app.widget_name, "status" app.widget_status DEFAULT 'new'::app.widget_status NOT NULL, CONSTRAINT "widgets_pkey" PRIMARY KEY (id))`,
		}},
	}

	plan := BuildPlan(project, desired, actual, Options{})
	if len(plan.Operations) != 0 {
		t.Fatalf("expected no operations, got %#v", plan.Operations)
	}
}

func TestBuildPlanAddsColumnIncrementally(t *testing.T) {
	project := &projectxml.Project{}
	desired := &model.SchemaModel{
		Tables: []model.TableDef{{
			Schema: "app",
			Name:   "widgets",
			SQL:    "CREATE TABLE app.widgets (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), name app.widget_name, status app.widget_status NOT NULL DEFAULT 'new', version integer NOT NULL DEFAULT 1)",
		}},
	}
	actual := &model.SchemaModel{
		Tables: []model.TableDef{{
			Schema: "app",
			Name:   "widgets",
			SQL:    `CREATE TABLE "app"."widgets" ("id" uuid DEFAULT gen_random_uuid() NOT NULL, "name" app.widget_name, "status" app.widget_status DEFAULT 'new'::app.widget_status NOT NULL, CONSTRAINT "widgets_pkey" PRIMARY KEY (id))`,
		}},
	}

	plan := BuildPlan(project, desired, actual, Options{})
	if len(plan.Operations) != 1 {
		t.Fatalf("expected 1 operation, got %#v", plan.Operations)
	}
	if got, want := plan.Operations[0].Kind, "alter-table-add-column"; got != want {
		t.Fatalf("operation kind = %q, want %q", got, want)
	}
	if got, want := plan.Operations[0].SQL, `ALTER TABLE "app"."widgets" ADD COLUMN version integer NOT NULL DEFAULT 1;`; got != want {
		t.Fatalf("operation SQL = %q, want %q", got, want)
	}
}

func TestBuildPlanDropsColumnIncrementallyWhenAuthorized(t *testing.T) {
	project := &projectxml.Project{}
	desired := &model.SchemaModel{Tables: []model.TableDef{{
		Schema: "app",
		Name:   "widgets",
		SQL:    "CREATE TABLE app.widgets (id uuid PRIMARY KEY, name text)",
	}}}
	actual := &model.SchemaModel{Tables: []model.TableDef{{
		Schema: "app",
		Name:   "widgets",
		SQL:    `CREATE TABLE "app"."widgets" ("id" uuid NOT NULL, "obsolete" text, "name" text, CONSTRAINT "widgets_pkey" PRIMARY KEY (id))`,
	}}}

	plan := BuildPlan(project, desired, actual, Options{AllowDrop: true})
	if len(plan.Operations) != 1 {
		t.Fatalf("expected 1 operation, got %#v", plan.Operations)
	}
	operation := plan.Operations[0]
	if got, want := operation.Kind, "alter-table-drop-column"; got != want {
		t.Fatalf("operation kind = %q, want %q", got, want)
	}
	if got, want := operation.ObjectKey, "app.widgets.obsolete"; got != want {
		t.Fatalf("object key = %q, want %q", got, want)
	}
	if got, want := operation.SQL, `ALTER TABLE "app"."widgets" DROP COLUMN "obsolete";`; got != want {
		t.Fatalf("operation SQL = %q, want %q", got, want)
	}
	if operation.Risk != "destructive" || !plan.Summary.Destructive || !plan.Summary.Supported {
		t.Fatalf("unexpected plan safety metadata: %#v", plan)
	}
}

func TestBuildPlanBlocksColumnDropWithoutAuthorization(t *testing.T) {
	project := &projectxml.Project{}
	desired := &model.SchemaModel{Tables: []model.TableDef{{
		Schema: "app",
		Name:   "widgets",
		SQL:    "CREATE TABLE app.widgets (id uuid PRIMARY KEY)",
	}}}
	actual := &model.SchemaModel{Tables: []model.TableDef{{
		Schema: "app",
		Name:   "widgets",
		SQL:    `CREATE TABLE "app"."widgets" ("id" uuid NOT NULL, "obsolete" text, CONSTRAINT "widgets_pkey" PRIMARY KEY (id))`,
	}}}

	plan := BuildPlan(project, desired, actual, Options{})
	if len(plan.Operations) != 1 || plan.Operations[0].Kind != "blocked-alter-table-drop-column" {
		t.Fatalf("unexpected operations: %#v", plan.Operations)
	}
	if plan.Summary.Supported || !plan.Summary.Destructive {
		t.Fatalf("unexpected plan safety metadata: %#v", plan.Summary)
	}
	if !strings.HasPrefix(plan.Operations[0].SQL, "-- requires --allow-drop") {
		t.Fatalf("drop SQL was not blocked: %q", plan.Operations[0].SQL)
	}
}

func TestBuildPlanAltersColumnTypeDefaultAndNullabilityWithoutRecreatingTable(t *testing.T) {
	project := &projectxml.Project{}
	desired := &model.SchemaModel{Tables: []model.TableDef{{
		Schema: "app",
		Name:   "widgets",
		SQL:    "CREATE TABLE app.widgets (id uuid PRIMARY KEY, quantity bigint NOT NULL DEFAULT 1)",
	}}}
	actual := &model.SchemaModel{Tables: []model.TableDef{{
		Schema: "app",
		Name:   "widgets",
		SQL:    `CREATE TABLE "app"."widgets" ("id" uuid NOT NULL, "quantity" integer DEFAULT 0, CONSTRAINT "widgets_pkey" PRIMARY KEY (id))`,
	}}}

	plan := BuildPlan(project, desired, actual, Options{})
	wantKinds := []string{
		"alter-table-drop-column-default",
		"alter-table-alter-column-type",
		"alter-table-set-column-default",
		"alter-table-set-column-not-null",
	}
	if len(plan.Operations) != len(wantKinds) {
		t.Fatalf("operations = %#v, want %d native alterations", plan.Operations, len(wantKinds))
	}
	for index, wantKind := range wantKinds {
		if plan.Operations[index].Kind != wantKind {
			t.Fatalf("operation %d kind = %q, want %q: %#v", index, plan.Operations[index].Kind, wantKind, plan.Operations)
		}
		if strings.Contains(plan.Operations[index].SQL, "DROP TABLE") {
			t.Fatalf("column alteration recreates table: %s", plan.Operations[index].SQL)
		}
	}
	if got, want := plan.Operations[1].SQL, `ALTER TABLE "app"."widgets" ALTER COLUMN "quantity" TYPE bigint USING "quantity"::bigint;`; got != want {
		t.Fatalf("type alteration SQL = %q, want %q", got, want)
	}
	if plan.Summary.Destructive || !plan.Summary.Supported {
		t.Fatalf("unexpected plan safety metadata: %#v", plan.Summary)
	}
}

func TestBuildPlanDropsColumnDefaultAndNotNullWithoutRecreatingTable(t *testing.T) {
	project := &projectxml.Project{}
	desired := &model.SchemaModel{Tables: []model.TableDef{{
		Schema: "app",
		Name:   "widgets",
		SQL:    "CREATE TABLE app.widgets (id uuid PRIMARY KEY, name text)",
	}}}
	actual := &model.SchemaModel{Tables: []model.TableDef{{
		Schema: "app",
		Name:   "widgets",
		SQL:    `CREATE TABLE "app"."widgets" ("id" uuid NOT NULL, "name" text DEFAULT 'unknown' NOT NULL, CONSTRAINT "widgets_pkey" PRIMARY KEY (id))`,
	}}}

	plan := BuildPlan(project, desired, actual, Options{})
	if len(plan.Operations) != 2 {
		t.Fatalf("unexpected operations: %#v", plan.Operations)
	}
	if plan.Operations[0].Kind != "alter-table-drop-column-default" || plan.Operations[1].Kind != "alter-table-drop-column-not-null" {
		t.Fatalf("unexpected operation order: %#v", plan.Operations)
	}
	for _, operation := range plan.Operations {
		if operation.Risk != "safe" || strings.Contains(operation.SQL, "DROP TABLE") {
			t.Fatalf("unexpected native alteration: %#v", operation)
		}
	}
}

func TestBuildPlanReplacesPrimaryKeyWithoutRecreatingTable(t *testing.T) {
	project := &projectxml.Project{}
	desired := &model.SchemaModel{Tables: []model.TableDef{{
		Schema: "app",
		Name:   "widgets",
		SQL:    "CREATE TABLE app.widgets (tenant_id uuid NOT NULL, id uuid NOT NULL, CONSTRAINT widgets_pkey PRIMARY KEY (tenant_id, id))",
	}}}
	actual := &model.SchemaModel{Tables: []model.TableDef{{
		Schema: "app",
		Name:   "widgets",
		SQL:    `CREATE TABLE "app"."widgets" ("tenant_id" uuid NOT NULL, "id" uuid NOT NULL, CONSTRAINT "widgets_pkey" PRIMARY KEY (id))`,
	}}}

	plan := BuildPlan(project, desired, actual, Options{})
	if len(plan.Operations) != 2 {
		t.Fatalf("unexpected operations: %#v", plan.Operations)
	}
	if plan.Operations[0].Kind != "alter-table-drop-primary-key" || plan.Operations[1].Kind != "alter-table-add-primary-key" {
		t.Fatalf("unexpected primary-key operations: %#v", plan.Operations)
	}
	if got, want := plan.Operations[1].SQL, `ALTER TABLE "app"."widgets" ADD CONSTRAINT "widgets_pkey" PRIMARY KEY ("tenant_id", "id");`; got != want {
		t.Fatalf("add primary key SQL = %q, want %q", got, want)
	}
	for _, operation := range plan.Operations {
		if operation.Risk != "migration" || strings.Contains(operation.SQL, "DROP TABLE") {
			t.Fatalf("unexpected primary-key alteration: %#v", operation)
		}
	}
}
