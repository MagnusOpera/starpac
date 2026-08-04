package diff

import (
	"testing"

	"github.com/MagnusOpera/pgpac/internal/model"
	"github.com/MagnusOpera/pgpac/internal/projectxml"
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
