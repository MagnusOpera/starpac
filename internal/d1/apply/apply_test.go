package apply

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/MagnusOpera/starpac/internal/d1/diff"
	"github.com/MagnusOpera/starpac/internal/d1/project"
)

type sqliteExecutor struct {
	database *sql.DB
}

func (executor *sqliteExecutor) Query(ctx context.Context, statement string) ([]map[string]any, error) {
	_, err := executor.database.ExecContext(ctx, statement)
	return nil, err
}

func (executor *sqliteExecutor) Batch(ctx context.Context, statements []string) error {
	transaction, err := executor.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	for _, statement := range statements {
		if _, err := transaction.ExecContext(ctx, statement); err != nil {
			_ = transaction.Rollback()
			return err
		}
	}
	return transaction.Commit()
}

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

func TestExecuteRebuildsReferencedTableTransactionally(t *testing.T) {
	ctx := context.Background()
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	defer database.Close()

	setup := []string{
		"PRAGMA foreign_keys = ON",
		"CREATE TABLE widgets (id INTEGER PRIMARY KEY, obsolete TEXT)",
		"CREATE TABLE events (id INTEGER PRIMARY KEY, widget_id INTEGER NOT NULL REFERENCES widgets(id))",
		"INSERT INTO widgets (id, obsolete) VALUES (1, 'legacy')",
		"INSERT INTO events (id, widget_id) VALUES (1, 1)",
	}
	for _, statement := range setup {
		if _, err := database.ExecContext(ctx, statement); err != nil {
			t.Fatalf("setup %q: %v", statement, err)
		}
	}

	project := &projectxml.Project{Target: projectxml.TargetConfig{
		Apply: projectxml.ApplyConfig{UseTransaction: true},
	}}
	plan := diff.Plan{
		Summary: diff.Summary{Destructive: true, Supported: true},
		Operations: []diff.Operation{{
			Kind: "rebuild-table",
			Risk: "destructive",
			SQL: strings.Join([]string{
				"ALTER TABLE widgets RENAME TO __d1pac_old_widgets",
				"CREATE TABLE widgets (id INTEGER PRIMARY KEY)",
				"INSERT INTO widgets (id) SELECT id FROM __d1pac_old_widgets",
				"ALTER TABLE events RENAME TO __d1pac_old_events",
				"CREATE TABLE events (id INTEGER PRIMARY KEY, widget_id INTEGER NOT NULL REFERENCES widgets(id))",
				"INSERT INTO events (id, widget_id) SELECT id, widget_id FROM __d1pac_old_events",
				"DROP TABLE __d1pac_old_events",
				"DROP TABLE __d1pac_old_widgets",
			}, ";\n") + ";",
		}},
	}
	executor := &sqliteExecutor{database: database}
	if err := Execute(ctx, executor, project, plan, Options{AllowDrop: true, Force: true}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	var childCount int
	if err := database.QueryRowContext(ctx, "SELECT COUNT(*) FROM events WHERE widget_id = 1").Scan(&childCount); err != nil {
		t.Fatalf("query retained child: %v", err)
	}
	if childCount != 1 {
		t.Fatalf("retained child count = %d, want 1", childCount)
	}
	var violations int
	if err := database.QueryRowContext(ctx, "SELECT COUNT(*) FROM pragma_foreign_key_check").Scan(&violations); err != nil {
		t.Fatalf("check foreign keys: %v", err)
	}
	if violations != 0 {
		t.Fatalf("foreign key violations = %d, want 0", violations)
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
