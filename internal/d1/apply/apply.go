package apply

import (
	"context"
	"fmt"
	"strings"

	"github.com/MagnusOpera/starpac/internal/d1/diff"
	"github.com/MagnusOpera/starpac/internal/d1/project"
	"github.com/MagnusOpera/starpac/internal/pac/safety"
)

type Executor interface {
	Query(context.Context, string) ([]map[string]any, error)
	Batch(context.Context, []string) error
}

type Options struct {
	AllowDrop bool
	Force     bool
}

func Execute(ctx context.Context, executor Executor, project *projectxml.Project, plan diff.Plan, options Options) error {
	if err := safety.ValidateDestructivePlan(
		plan.Summary.Destructive,
		project.Target.Plan.AllowDrop,
		options.AllowDrop,
		options.Force,
	); err != nil {
		return err
	}
	statements := make([]string, 0, len(plan.Operations)+1)
	for _, operation := range plan.Operations {
		sql := strings.TrimSpace(operation.SQL)
		if sql == "" || strings.HasPrefix(sql, "--") {
			continue
		}
		statements = append(statements, sql)
	}
	if len(statements) == 0 {
		return nil
	}
	statements = append([]string{"PRAGMA defer_foreign_keys = ON"}, statements...)
	if project.Target.Apply.UseTransaction {
		if err := executor.Batch(ctx, statements); err != nil {
			return fmt.Errorf("apply transaction: %w", err)
		}
		return nil
	}
	for index, statement := range statements {
		if _, err := executor.Query(ctx, statement); err != nil {
			return fmt.Errorf("apply statement %d: %w", index+1, err)
		}
	}
	return nil
}
