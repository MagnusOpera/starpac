package render

import (
	"testing"

	"github.com/MagnusOpera/starpac/internal/pac/plan"
)

func TestTextAndSQLAreDeterministic(t *testing.T) {
	deploymentPlan := plan.Plan{
		Summary: plan.Summary{
			Supported:      false,
			Destructive:    true,
			OperationCount: 1,
		},
		Operations: []plan.Operation{{
			Kind:      "blocked-drop-table",
			ObjectKey: "widgets",
			Risk:      plan.RiskDestructive,
			SQL:       " -- DROP TABLE widgets; ",
		}},
	}

	wantText := "Plan status: blocked (contains destructive operations)\nOperations: 1\n- blocked-drop-table [destructive] widgets\n"
	if got := Text(deploymentPlan); got != wantText {
		t.Fatalf("Text() = %q, want %q", got, wantText)
	}

	wantSQL := "-- blocked-drop-table widgets\n-- DROP TABLE widgets;\n"
	if got := SQL(deploymentPlan); got != wantSQL {
		t.Fatalf("SQL() = %q, want %q", got, wantSQL)
	}
}
