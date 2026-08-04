package compiler

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"

	"github.com/MagnusOpera/d1pac/internal/introspect"
	"github.com/MagnusOpera/d1pac/internal/model"
	"github.com/MagnusOpera/d1pac/internal/projectxml"
)

func BuildDesiredModel(ctx context.Context, project *projectxml.Project) (*model.SchemaModel, error) {
	files, err := project.ResolveFiles()
	if err != nil {
		return nil, err
	}
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	defer database.Close()
	if _, err := database.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}
	for _, file := range files {
		content, err := os.ReadFile(file.AbsPath)
		if err != nil {
			return nil, err
		}
		if _, err := database.ExecContext(ctx, string(content)); err != nil {
			return nil, fmt.Errorf("%s: compile schema: %w", file.RelPath, err)
		}
	}
	return introspect.LoadLocal(ctx, database, project.IsIgnored)
}
