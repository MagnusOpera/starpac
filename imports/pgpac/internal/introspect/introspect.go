package introspect

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/MagnusOpera/pgpac/internal/model"
)

func LoadActualModel(ctx context.Context, connectionString string, ownedSchemas []string, managedExtensions []string, expectedVersion int) (*model.SchemaModel, error) {
	conn, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	var versionNum int
	if err := conn.QueryRow(ctx, `select current_setting('server_version_num')::int`).Scan(&versionNum); err != nil {
		return nil, err
	}
	targetVersion := versionNum / 10000
	if targetVersion < 17 {
		return nil, fmt.Errorf("target database is PostgreSQL %d, but pgpac requires PostgreSQL 17 or newer", targetVersion)
	}

	m := &model.SchemaModel{PostgresVersion: targetVersion}
	if err := loadSchemas(ctx, conn, ownedSchemas, m); err != nil {
		return nil, err
	}
	if err := loadExtensions(ctx, conn, managedExtensions, m); err != nil {
		return nil, err
	}
	if err := loadTables(ctx, conn, ownedSchemas, m); err != nil {
		return nil, err
	}
	if err := loadIndexes(ctx, conn, ownedSchemas, m); err != nil {
		return nil, err
	}
	if err := loadViews(ctx, conn, ownedSchemas, m); err != nil {
		return nil, err
	}
	if err := loadRoutines(ctx, conn, ownedSchemas, m); err != nil {
		return nil, err
	}
	if err := loadEnums(ctx, conn, ownedSchemas, m); err != nil {
		return nil, err
	}
	if err := loadDomains(ctx, conn, ownedSchemas, m); err != nil {
		return nil, err
	}
	if err := loadSequences(ctx, conn, ownedSchemas, m); err != nil {
		return nil, err
	}
	if err := loadComments(ctx, conn, ownedSchemas, m); err != nil {
		return nil, err
	}

	model.Sort(m)
	return m, nil
}

func loadSchemas(ctx context.Context, conn *pgx.Conn, ownedSchemas []string, m *model.SchemaModel) error {
	rows, err := conn.Query(ctx, `select nspname from pg_namespace where nspname = any($1::text[])`, ownedSchemas)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return err
		}
		m.Schemas = append(m.Schemas, model.SchemaDef{Name: schema, SQL: model.CanonicalSQL(fmt.Sprintf("CREATE SCHEMA %s", quoteIdent(schema)))})
	}
	return rows.Err()
}

func loadExtensions(ctx context.Context, conn *pgx.Conn, managedExtensions []string, m *model.SchemaModel) error {
	if len(managedExtensions) == 0 {
		return nil
	}

	rows, err := conn.Query(ctx, `select extname, extversion from pg_extension where extname = any($1::text[])`, managedExtensions)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var name, version string
		if err := rows.Scan(&name, &version); err != nil {
			return err
		}
		m.Extensions = append(m.Extensions, model.ExtensionDef{
			Name:    name,
			Version: version,
			SQL:     model.CanonicalSQL(fmt.Sprintf("CREATE EXTENSION %s VERSION '%s'", quoteIdent(name), version)),
		})
	}
	return rows.Err()
}

func loadTables(ctx context.Context, conn *pgx.Conn, ownedSchemas []string, m *model.SchemaModel) error {
	rows, err := conn.Query(ctx, `
select
  n.nspname,
  c.relname,
  pg_get_expr(d.adbin, d.adrelid),
  col_description(c.oid, 0),
  a.attname,
  format_type(a.atttypid, a.atttypmod),
  a.attnotnull,
  pg_get_expr(ad.adbin, ad.adrelid),
  col_description(c.oid, a.attnum)
from pg_class c
join pg_namespace n on n.oid = c.relnamespace
join pg_attribute a on a.attrelid = c.oid
left join pg_attrdef ad on ad.adrelid = a.attrelid and ad.adnum = a.attnum
left join pg_attrdef d on d.adrelid = c.oid and d.adnum = 0
where c.relkind in ('r','p')
  and a.attnum > 0
  and not a.attisdropped
  and n.nspname = any($1::text[])
order by n.nspname, c.relname, a.attnum`, ownedSchemas)
	if err != nil {
		return err
	}
	defer rows.Close()

	type column struct {
		name       string
		dataType   string
		notNull    bool
		defaultSQL string
		comment    string
	}
	type tableAgg struct {
		schema  string
		name    string
		comment string
		cols    []column
	}
	tables := map[string]*tableAgg{}
	for rows.Next() {
		var schema, name string
		var tableDefault *string
		var tableComment *string
		var colName, dataType string
		var notNull bool
		var defaultSQL *string
		var colComment *string
		if err := rows.Scan(&schema, &name, &tableDefault, &tableComment, &colName, &dataType, &notNull, &defaultSQL, &colComment); err != nil {
			return err
		}
		key := model.QualifiedName(schema, name)
		agg, ok := tables[key]
		if !ok {
			agg = &tableAgg{schema: schema, name: name}
			if tableComment != nil {
				agg.comment = *tableComment
			}
			tables[key] = agg
		}
		agg.cols = append(agg.cols, column{
			name:       colName,
			dataType:   dataType,
			notNull:    notNull,
			defaultSQL: deref(defaultSQL),
			comment:    deref(colComment),
		})
	}
	if rows.Err() != nil {
		return rows.Err()
	}

	constraintRows, err := conn.Query(ctx, `
select
  n.nspname,
  c.relname,
  co.conname,
  pg_get_constraintdef(co.oid, true)
from pg_constraint co
join pg_class c on c.oid = co.conrelid
join pg_namespace n on n.oid = c.relnamespace
where n.nspname = any($1::text[])
order by n.nspname, c.relname, co.conname`, ownedSchemas)
	if err != nil {
		return err
	}
	defer constraintRows.Close()
	constraints := map[string][]string{}
	for constraintRows.Next() {
		var schema, name, constraintName, definition string
		if err := constraintRows.Scan(&schema, &name, &constraintName, &definition); err != nil {
			return err
		}
		key := model.QualifiedName(schema, name)
		constraints[key] = append(constraints[key], fmt.Sprintf("CONSTRAINT %s %s", quoteIdent(constraintName), definition))
	}
	if constraintRows.Err() != nil {
		return constraintRows.Err()
	}

	for key, table := range tables {
		var parts []string
		for _, col := range table.cols {
			line := fmt.Sprintf("%s %s", quoteIdent(col.name), col.dataType)
			if col.defaultSQL != "" {
				line += " DEFAULT " + col.defaultSQL
			}
			if col.notNull {
				line += " NOT NULL"
			}
			parts = append(parts, line)
		}
		parts = append(parts, constraints[key]...)
		sql := fmt.Sprintf("CREATE TABLE %s (%s)", model.QualifiedName(quoteIdent(table.schema), quoteIdent(table.name)), strings.Join(parts, ", "))
		m.Tables = append(m.Tables, model.TableDef{Schema: table.schema, Name: table.name, SQL: model.CanonicalSQL(sql)})
		if table.comment != "" {
			m.Comments = append(m.Comments, model.CommentDef{
				ObjectType: "table",
				ObjectKey:  key,
				Comment:    table.comment,
				SQL:        model.CanonicalSQL(fmt.Sprintf("COMMENT ON TABLE %s IS '%s'", key, escapeLiteral(table.comment))),
			})
		}
		for _, col := range table.cols {
			if col.comment == "" {
				continue
			}
			objectKey := key + "." + col.name
			m.Comments = append(m.Comments, model.CommentDef{
				ObjectType: "column",
				ObjectKey:  objectKey,
				Comment:    col.comment,
				SQL:        model.CanonicalSQL(fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '%s'", key, quoteIdent(col.name), escapeLiteral(col.comment))),
			})
		}
	}
	return nil
}

func loadIndexes(ctx context.Context, conn *pgx.Conn, ownedSchemas []string, m *model.SchemaModel) error {
	rows, err := conn.Query(ctx, `
select
  ni.nspname,
  i.relname,
  nt.nspname,
  t.relname,
  pg_get_indexdef(i.oid)
from pg_index x
join pg_class i on i.oid = x.indexrelid
join pg_class t on t.oid = x.indrelid
join pg_namespace ni on ni.oid = i.relnamespace
join pg_namespace nt on nt.oid = t.relnamespace
where nt.nspname = any($1::text[])
  and not x.indisprimary
order by ni.nspname, i.relname`, ownedSchemas)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var schema, name, tableSchema, tableName, sql string
		if err := rows.Scan(&schema, &name, &tableSchema, &tableName, &sql); err != nil {
			return err
		}
		m.Indexes = append(m.Indexes, model.IndexDef{
			Schema:      schema,
			Name:        name,
			TableSchema: tableSchema,
			TableName:   tableName,
			SQL:         model.CanonicalSQL(sql),
		})
	}
	return rows.Err()
}

func loadViews(ctx context.Context, conn *pgx.Conn, ownedSchemas []string, m *model.SchemaModel) error {
	rows, err := conn.Query(ctx, `
select
  n.nspname,
  c.relname,
  pg_get_viewdef(c.oid, true)
from pg_class c
join pg_namespace n on n.oid = c.relnamespace
where c.relkind = 'v'
  and n.nspname = any($1::text[])
order by n.nspname, c.relname`, ownedSchemas)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var schema, name, definition string
		if err := rows.Scan(&schema, &name, &definition); err != nil {
			return err
		}
		sql := fmt.Sprintf("CREATE VIEW %s AS %s", model.QualifiedName(quoteIdent(schema), quoteIdent(name)), definition)
		m.Views = append(m.Views, model.ViewDef{Schema: schema, Name: name, SQL: model.CanonicalSQL(sql)})
	}
	return rows.Err()
}

func loadRoutines(ctx context.Context, conn *pgx.Conn, ownedSchemas []string, m *model.SchemaModel) error {
	rows, err := conn.Query(ctx, `
select
  n.nspname,
  p.proname,
  pg_get_function_identity_arguments(p.oid),
  case p.prokind when 'p' then 'procedure' else 'function' end,
  pg_get_functiondef(p.oid)
from pg_proc p
join pg_namespace n on n.oid = p.pronamespace
where n.nspname = any($1::text[])
order by n.nspname, p.proname, pg_get_function_identity_arguments(p.oid)`, ownedSchemas)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var schema, name, identityArgs, kind, sql string
		if err := rows.Scan(&schema, &name, &identityArgs, &kind, &sql); err != nil {
			return err
		}
		m.Routines = append(m.Routines, model.RoutineDef{
			Schema:       schema,
			Name:         name,
			IdentityArgs: identityArgs,
			Kind:         kind,
			SQL:          model.CanonicalSQL(sql),
		})
	}
	return rows.Err()
}

func loadEnums(ctx context.Context, conn *pgx.Conn, ownedSchemas []string, m *model.SchemaModel) error {
	rows, err := conn.Query(ctx, `
select
  n.nspname,
  t.typname,
  string_agg(quote_literal(e.enumlabel), ', ' order by e.enumsortorder)
from pg_type t
join pg_namespace n on n.oid = t.typnamespace
join pg_enum e on e.enumtypid = t.oid
where n.nspname = any($1::text[])
group by n.nspname, t.typname
order by n.nspname, t.typname`, ownedSchemas)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var schema, name, labels string
		if err := rows.Scan(&schema, &name, &labels); err != nil {
			return err
		}
		sql := fmt.Sprintf("CREATE TYPE %s AS ENUM (%s)", model.QualifiedName(quoteIdent(schema), quoteIdent(name)), labels)
		m.Enums = append(m.Enums, model.EnumDef{
			Schema: schema,
			Name:   name,
			Values: parseEnumLabels(labels),
			SQL:    model.CanonicalSQL(sql),
		})
	}
	return rows.Err()
}

func loadDomains(ctx context.Context, conn *pgx.Conn, ownedSchemas []string, m *model.SchemaModel) error {
	rows, err := conn.Query(ctx, `
select
  n.nspname,
  t.typname,
  format_type(t.typbasetype, t.typtypmod),
  t.typnotnull,
  pg_get_expr(d.adbin, d.adrelid),
  (
    select string_agg(pg_get_constraintdef(c.oid, true), ' ')
    from pg_constraint c
    where c.contypid = t.oid
  )
from pg_type t
join pg_namespace n on n.oid = t.typnamespace
left join pg_attrdef d on d.adrelid = 0 and d.adnum = 0
where t.typtype = 'd'
  and n.nspname = any($1::text[])
order by n.nspname, t.typname`, ownedSchemas)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var schema, name, baseType string
		var notNull bool
		var defaultSQL, constraints *string
		if err := rows.Scan(&schema, &name, &baseType, &notNull, &defaultSQL, &constraints); err != nil {
			return err
		}
		parts := []string{fmt.Sprintf("CREATE DOMAIN %s AS %s", model.QualifiedName(quoteIdent(schema), quoteIdent(name)), baseType)}
		if defaultSQL != nil && *defaultSQL != "" {
			parts = append(parts, "DEFAULT "+*defaultSQL)
		}
		if notNull {
			parts = append(parts, "NOT NULL")
		}
		if constraints != nil && *constraints != "" {
			parts = append(parts, *constraints)
		}
		m.Domains = append(m.Domains, model.DomainDef{Schema: schema, Name: name, SQL: model.CanonicalSQL(strings.Join(parts, " "))})
	}
	return rows.Err()
}

func loadSequences(ctx context.Context, conn *pgx.Conn, ownedSchemas []string, m *model.SchemaModel) error {
	rows, err := conn.Query(ctx, `
select
  schemaname,
  sequencename,
  data_type,
  start_value,
  increment_by,
  min_value,
  max_value,
  cycle,
  cache_size
from pg_sequences
where schemaname = any($1::text[])
order by schemaname, sequencename`, ownedSchemas)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var schema, name, dataType string
		var startValue, incrementBy, minValue, maxValue, cacheSize int64
		var cycle bool
		if err := rows.Scan(&schema, &name, &dataType, &startValue, &incrementBy, &minValue, &maxValue, &cycle, &cacheSize); err != nil {
			return err
		}
		sql := fmt.Sprintf(
			"CREATE SEQUENCE %s AS %s INCREMENT BY %d MINVALUE %d MAXVALUE %d START WITH %d CACHE %d %s",
			model.QualifiedName(quoteIdent(schema), quoteIdent(name)),
			dataType,
			incrementBy,
			minValue,
			maxValue,
			startValue,
			cacheSize,
			map[bool]string{true: "CYCLE", false: "NO CYCLE"}[cycle],
		)
		m.Sequences = append(m.Sequences, model.SequenceDef{Schema: schema, Name: name, SQL: model.CanonicalSQL(sql)})
	}
	return rows.Err()
}

func loadComments(ctx context.Context, conn *pgx.Conn, ownedSchemas []string, m *model.SchemaModel) error {
	_ = ctx
	_ = conn
	_ = ownedSchemas
	return nil
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func escapeLiteral(s string) string {
	return strings.ReplaceAll(s, `'`, `''`)
}

func deref(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func parseEnumLabels(labels string) []string {
	if labels == "" {
		return nil
	}
	parts := strings.Split(labels, ", ")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		values = append(values, strings.Trim(part, `'`))
	}
	return values
}
