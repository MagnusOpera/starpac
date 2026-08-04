package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/MagnusOpera/starpac/internal/pac/plan"
)

func TestWriteVersion(t *testing.T) {
	var output bytes.Buffer
	WriteVersion(&output, "d1pac", "1.2.3", "abc123", "2026-08-04T00:00:00Z")
	want := "d1pac 1.2.3\ncommit: abc123\nbuilt: 2026-08-04T00:00:00Z\n"
	if output.String() != want {
		t.Fatalf("WriteVersion() = %q, want %q", output.String(), want)
	}
}

func TestWritePlanJSONAndScript(t *testing.T) {
	deploymentPlan := plan.Plan{
		Summary: plan.Summary{
			Supported:      true,
			OperationCount: 1,
		},
		Operations: []plan.Operation{{
			Kind:       "create-table",
			ObjectType: "table",
			ObjectKey:  "widgets",
			Risk:       plan.RiskSafe,
			SQL:        "CREATE TABLE widgets (id integer);",
		}},
	}
	var output bytes.Buffer
	scriptPath := filepath.Join(t.TempDir(), "plan.sql")
	if err := WritePlan(&output, "json", scriptPath, deploymentPlan); err != nil {
		t.Fatalf("WritePlan returned error: %v", err)
	}
	wantJSON := `{
  "summary": {
    "supported": true,
    "destructive": false,
    "operationCount": 1
  },
  "operations": [
    {
      "kind": "create-table",
      "objectType": "table",
      "objectKey": "widgets",
      "risk": "safe",
      "sql": "CREATE TABLE widgets (id integer);"
    }
  ]
}
`
	if output.String() != wantJSON {
		t.Fatalf("JSON contract changed:\n%s", output.String())
	}

	script, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	wantScript := "-- create-table widgets\nCREATE TABLE widgets (id integer);\n"
	if string(script) != wantScript {
		t.Fatalf("SQL contract changed: %q", script)
	}
}
