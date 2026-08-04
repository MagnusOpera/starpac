package introspect

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/MagnusOpera/d1pac/internal/model"
)

type RemoteQueryer interface {
	Query(context.Context, string) ([]map[string]any, error)
}

type IgnoreFunc func(objectType, name string) bool

func LoadLocal(ctx context.Context, database *sql.DB, ignore IgnoreFunc) (*model.SchemaModel, error) {
	schema := &model.SchemaModel{}
	if err := database.QueryRowContext(ctx, "SELECT sqlite_version()").Scan(&schema.SQLiteVersion); err != nil {
		return nil, err
	}
	rows, err := database.QueryContext(ctx, `
SELECT type, name, tbl_name, sql
FROM sqlite_schema
WHERE type IN ('table', 'index', 'view', 'trigger')
  AND sql IS NOT NULL
ORDER BY type, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var objectType string
		var name string
		var tableName string
		var ddl string
		if err := rows.Scan(&objectType, &name, &tableName, &ddl); err != nil {
			return nil, err
		}
		if ignore != nil && ignore(objectType, name) {
			continue
		}
		addObject(schema, objectType, name, tableName, ddl)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for tableIndex := range schema.Tables {
		if err := loadLocalTableDetails(ctx, database, &schema.Tables[tableIndex]); err != nil {
			return nil, err
		}
	}
	model.Sort(schema)
	return schema, nil
}

func LoadRemote(ctx context.Context, queryer RemoteQueryer, ignore IgnoreFunc) (*model.SchemaModel, error) {
	versionRows, err := queryer.Query(ctx, "SELECT sqlite_version() AS sqlite_version")
	if err != nil {
		return nil, err
	}
	schema := &model.SchemaModel{}
	if len(versionRows) > 0 {
		schema.SQLiteVersion = stringValue(versionRows[0]["sqlite_version"])
	}
	rows, err := queryer.Query(ctx, `
SELECT type, name, tbl_name, sql
FROM sqlite_schema
WHERE type IN ('table', 'index', 'view', 'trigger')
  AND sql IS NOT NULL
ORDER BY type, name`)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		objectType := stringValue(row["type"])
		name := stringValue(row["name"])
		if ignore != nil && ignore(objectType, name) {
			continue
		}
		addObject(schema, objectType, name, stringValue(row["tbl_name"]), stringValue(row["sql"]))
	}
	for tableIndex := range schema.Tables {
		if err := loadRemoteTableDetails(ctx, queryer, &schema.Tables[tableIndex]); err != nil {
			return nil, err
		}
	}
	model.Sort(schema)
	return schema, nil
}

func addObject(schema *model.SchemaModel, objectType, name, tableName, ddl string) {
	ddl = model.CanonicalSQL(ddl)
	switch objectType {
	case "table":
		schema.Tables = append(schema.Tables, model.TableDef{
			Name: name,
			SQL:  ddl,
		})
	case "index":
		schema.Indexes = append(schema.Indexes, model.IndexDef{
			Name:      name,
			TableName: tableName,
			SQL:       ddl,
		})
	case "view":
		schema.Views = append(schema.Views, model.ViewDef{
			Name: name,
			SQL:  ddl,
		})
	case "trigger":
		schema.Triggers = append(schema.Triggers, model.TriggerDef{
			Name:      name,
			TableName: tableName,
			SQL:       ddl,
		})
	}
}

func loadLocalTableDetails(ctx context.Context, database *sql.DB, table *model.TableDef) error {
	definitions := model.ColumnDefinitions(table.SQL)
	columnRows, err := database.QueryContext(ctx, "PRAGMA table_xinfo("+model.QuoteIdentifier(table.Name)+")")
	if err != nil {
		return err
	}
	for columnRows.Next() {
		var column model.ColumnDef
		var notNull int
		var defaultSQL sql.NullString
		if err := columnRows.Scan(
			&column.Position,
			&column.Name,
			&column.Type,
			&notNull,
			&defaultSQL,
			&column.PrimaryKey,
			&column.Hidden,
		); err != nil {
			columnRows.Close()
			return err
		}
		column.NotNull = notNull != 0
		if defaultSQL.Valid {
			value := defaultSQL.String
			column.DefaultSQL = &value
		}
		column.Definition = definitions[column.Name]
		table.Columns = append(table.Columns, column)
	}
	if err := columnRows.Close(); err != nil {
		return err
	}
	foreignKeyRows, err := database.QueryContext(ctx, "PRAGMA foreign_key_list("+model.QuoteIdentifier(table.Name)+")")
	if err != nil {
		return err
	}
	defer foreignKeyRows.Close()
	for foreignKeyRows.Next() {
		var foreignKey model.ForeignKeyDef
		if err := foreignKeyRows.Scan(
			&foreignKey.ID,
			&foreignKey.Sequence,
			&foreignKey.Table,
			&foreignKey.From,
			&foreignKey.To,
			&foreignKey.OnUpdate,
			&foreignKey.OnDelete,
			&foreignKey.Match,
		); err != nil {
			return err
		}
		table.ForeignKeys = append(table.ForeignKeys, foreignKey)
	}
	return foreignKeyRows.Err()
}

func loadRemoteTableDetails(ctx context.Context, queryer RemoteQueryer, table *model.TableDef) error {
	definitions := model.ColumnDefinitions(table.SQL)
	columnRows, err := queryer.Query(ctx, "PRAGMA table_xinfo("+model.QuoteIdentifier(table.Name)+")")
	if err != nil {
		return fmt.Errorf("inspect columns for %s: %w", table.Name, err)
	}
	for _, row := range columnRows {
		column := model.ColumnDef{
			Position:   intValue(row["cid"]),
			Name:       stringValue(row["name"]),
			Type:       stringValue(row["type"]),
			NotNull:    intValue(row["notnull"]) != 0,
			PrimaryKey: intValue(row["pk"]),
			Hidden:     intValue(row["hidden"]),
		}
		if value, exists := row["dflt_value"]; exists && value != nil {
			defaultSQL := stringValue(value)
			column.DefaultSQL = &defaultSQL
		}
		column.Definition = definitions[column.Name]
		table.Columns = append(table.Columns, column)
	}
	foreignKeyRows, err := queryer.Query(ctx, "PRAGMA foreign_key_list("+model.QuoteIdentifier(table.Name)+")")
	if err != nil {
		return fmt.Errorf("inspect foreign keys for %s: %w", table.Name, err)
	}
	for _, row := range foreignKeyRows {
		table.ForeignKeys = append(table.ForeignKeys, model.ForeignKeyDef{
			ID:       intValue(row["id"]),
			Sequence: intValue(row["seq"]),
			Table:    stringValue(row["table"]),
			From:     stringValue(row["from"]),
			To:       stringValue(row["to"]),
			OnUpdate: stringValue(row["on_update"]),
			OnDelete: stringValue(row["on_delete"]),
			Match:    stringValue(row["match"]),
		})
	}
	return nil
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return fmt.Sprint(typed)
	}
}

func intValue(value any) int {
	text := strings.TrimSpace(stringValue(value))
	integer, _ := strconv.Atoi(text)
	return integer
}
