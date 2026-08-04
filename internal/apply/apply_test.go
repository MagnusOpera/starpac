package apply

import (
	"context"
	"strings"
	"testing"

	"github.com/MagnusOpera/d1pac/internal/diff"
	"github.com/MagnusOpera/d1pac/internal/projectxml"
)

type recordingExecutor struct {
	queries []string
	batches [][]string
}

func (executor *recordingExecutor) Query(_ context.Context, sql string) ([]map[string]any, error) {
	executor.queries = append(executor.queries, sql)
	return nil, nil
}

func (executor *recordingExecutor) Batch(_ context.Context, statements []string) error {
	executor.batches = append(executor.batches, statements)
	return nil
}

func TestExecuteUsesBatchForTransactionalApply(t *testing.T) {
	executor := &recordingExecutor{}
	project := &projectxml.Project{Target: projectxml.TargetConfig{
		Apply: projectxml.ApplyConfig{UseTransaction: true},
	}}
	plan := diff.Plan{Operations: []diff.Operation{{
		Kind: "create-table",
		SQL:  "CREATE TABLE widgets(id INTEGER);",
	}}}
	if err := Execute(context.Background(), executor, project, plan, Options{}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(executor.batches) != 1 || len(executor.batches[0]) != 2 {
		t.Fatalf("unexpected batches: %#v", executor.batches)
	}
	if !strings.HasPrefix(executor.batches[0][0], "PRAGMA defer_foreign_keys") {
		t.Fatalf("foreign keys were not deferred: %#v", executor.batches[0])
	}
}

func TestExecuteRejectsDestructivePlan(t *testing.T) {
	executor := &recordingExecutor{}
	project := &projectxml.Project{}
	plan := diff.Plan{Summary: diff.Summary{Destructive: true}}
	err := Execute(context.Background(), executor, project, plan, Options{})
	if err == nil || !strings.Contains(err.Error(), "--allow-drop") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteSkipsBlockedOperations(t *testing.T) {
	executor := &recordingExecutor{}
	project := &projectxml.Project{Target: projectxml.TargetConfig{
		Apply: projectxml.ApplyConfig{UseTransaction: true},
	}}
	plan := diff.Plan{Operations: []diff.Operation{{
		Kind: "blocked-drop-table",
		SQL:  "-- drops are disabled\n-- DROP TABLE widgets;",
	}}}
	if err := Execute(context.Background(), executor, project, plan, Options{}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(executor.batches) != 0 {
		t.Fatalf("blocked operation was executed: %#v", executor.batches)
	}
}
