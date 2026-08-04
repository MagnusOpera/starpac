package apply

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/MagnusOpera/pgpac/internal/diff"
	"github.com/MagnusOpera/pgpac/internal/projectxml"
)

type Options struct {
	AllowDrop bool
	Force     bool
}

func Execute(ctx context.Context, connectionString string, project *projectxml.Project, plan diff.Plan, options Options) error {
	if plan.Summary.Destructive && !(project.Target.Plan.AllowDrop || options.AllowDrop || options.Force) {
		return fmt.Errorf("plan contains destructive operations; re-run with --allow-drop or --force")
	}

	conn, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	runInTx := func(tx pgx.Tx) error {
		if project.Target.Apply.LockTimeout != "" {
			lockTimeout, err := normalizeTimeout(project.Target.Apply.LockTimeout)
			if err != nil {
				return fmt.Errorf("invalid lock_timeout %q: %w", project.Target.Apply.LockTimeout, err)
			}
			if _, err := tx.Exec(ctx, fmt.Sprintf("SET lock_timeout = '%s'", lockTimeout)); err != nil {
				return err
			}
		}
		if project.Target.Apply.StatementTimeout != "" {
			statementTimeout, err := normalizeTimeout(project.Target.Apply.StatementTimeout)
			if err != nil {
				return fmt.Errorf("invalid statement_timeout %q: %w", project.Target.Apply.StatementTimeout, err)
			}
			if _, err := tx.Exec(ctx, fmt.Sprintf("SET statement_timeout = '%s'", statementTimeout)); err != nil {
				return err
			}
		}
		for _, operation := range plan.Operations {
			sql := strings.TrimSpace(operation.SQL)
			if sql == "" || strings.HasPrefix(sql, "--") {
				continue
			}
			if _, err := tx.Exec(ctx, sql); err != nil {
				return fmt.Errorf("%s %s: %w", operation.Kind, operation.ObjectKey, err)
			}
		}
		return nil
	}

	runNoTx := func() error {
		if project.Target.Apply.LockTimeout != "" {
			lockTimeout, err := normalizeTimeout(project.Target.Apply.LockTimeout)
			if err != nil {
				return fmt.Errorf("invalid lock_timeout %q: %w", project.Target.Apply.LockTimeout, err)
			}
			if _, err := conn.Exec(ctx, fmt.Sprintf("SET lock_timeout = '%s'", lockTimeout)); err != nil {
				return err
			}
		}
		if project.Target.Apply.StatementTimeout != "" {
			statementTimeout, err := normalizeTimeout(project.Target.Apply.StatementTimeout)
			if err != nil {
				return fmt.Errorf("invalid statement_timeout %q: %w", project.Target.Apply.StatementTimeout, err)
			}
			if _, err := conn.Exec(ctx, fmt.Sprintf("SET statement_timeout = '%s'", statementTimeout)); err != nil {
				return err
			}
		}
		for _, operation := range plan.Operations {
			sql := strings.TrimSpace(operation.SQL)
			if sql == "" || strings.HasPrefix(sql, "--") {
				continue
			}
			if _, err := conn.Exec(ctx, sql); err != nil {
				return fmt.Errorf("%s %s: %w", operation.Kind, operation.ObjectKey, err)
			}
		}
		return nil
	}

	if project.Target.Apply.UseTransaction {
		tx, err := conn.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		if err := runInTx(tx); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}

	return runNoTx()
}

func normalizeTimeout(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}

	// Support Go-style durations in project files and convert them to a PostgreSQL-compatible millisecond literal.
	if duration, err := time.ParseDuration(value); err == nil {
		return fmt.Sprintf("%dms", duration.Milliseconds()), nil
	}

	// Fall back to raw PostgreSQL-compatible timeout literals like "10min", "1h", "500ms", or "0".
	return value, nil
}
