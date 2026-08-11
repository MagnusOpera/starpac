package diff

import (
	"fmt"
	"sort"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v6"

	sharedplan "github.com/MagnusOpera/starpac/internal/pac/plan"
	"github.com/MagnusOpera/starpac/internal/postgres/model"
	"github.com/MagnusOpera/starpac/internal/postgres/project"
)

type Options struct {
	AllowDrop bool
}

type Plan = sharedplan.Plan
type Summary = sharedplan.Summary
type Operation = sharedplan.Operation

type planPermission int

const (
	permissionCreate planPermission = iota
	permissionAlter
)

func BuildPlan(project *projectxml.Project, desired, actual *model.SchemaModel, options Options) Plan {
	var ops []Operation

	diffByName(
		project,
		desired.Schemas,
		actual.Schemas,
		func(item model.SchemaDef) string { return item.Name },
		func(item model.SchemaDef) Operation {
			return op("create-schema", "schema", item.Name, "safe", ensureSemicolon(item.SQL))
		},
		func(item model.SchemaDef) Operation {
			return op("drop-schema", "schema", item.Name, "destructive", fmt.Sprintf("DROP SCHEMA %s CASCADE;", quoteQName(item.Name)))
		},
		func(a, b model.SchemaDef) bool { return true },
		&ops,
		options,
		permissionCreate,
	)

	diffByName(
		project,
		desired.Extensions,
		actual.Extensions,
		func(item model.ExtensionDef) string { return item.Name },
		func(item model.ExtensionDef) Operation {
			return op("create-extension", "extension", item.Name, "safe", ensureSemicolon(item.SQL))
		},
		func(item model.ExtensionDef) Operation {
			return op("drop-extension", "extension", item.Name, "destructive", fmt.Sprintf("DROP EXTENSION %s;", quoteQName(item.Name)))
		},
		func(a, b model.ExtensionDef) bool { return extensionEqual(a, b) },
		&ops,
		options,
		permissionCreate,
	)

	diffTables(project, desired.Tables, actual.Tables, &ops, options)

	diffByName(
		project,
		desired.Indexes,
		actual.Indexes,
		func(item model.IndexDef) string { return model.QualifiedName(item.Schema, item.Name) },
		func(item model.IndexDef) Operation {
			return op("create-index", "index", model.QualifiedName(item.Schema, item.Name), "safe", ensureSemicolon(item.SQL))
		},
		func(item model.IndexDef) Operation {
			return op("drop-index", "index", model.QualifiedName(item.Schema, item.Name), "destructive", fmt.Sprintf("DROP INDEX %s;", quoteQName(item.Schema, item.Name)))
		},
		func(a, b model.IndexDef) bool { return a.SQL == b.SQL },
		&ops,
		options,
		permissionCreate,
	)

	diffByName(
		project,
		desired.Views,
		actual.Views,
		func(item model.ViewDef) string { return model.QualifiedName(item.Schema, item.Name) },
		func(item model.ViewDef) Operation {
			return op("create-view", "view", model.QualifiedName(item.Schema, item.Name), "safe", ensureSemicolon(item.SQL))
		},
		func(item model.ViewDef) Operation {
			return op("drop-view", "view", model.QualifiedName(item.Schema, item.Name), "destructive", fmt.Sprintf("DROP VIEW %s CASCADE;", quoteQName(item.Schema, item.Name)))
		},
		func(a, b model.ViewDef) bool { return viewEqual(a, b) },
		&ops,
		options,
		permissionCreate,
	)

	diffByName(
		project,
		desired.Routines,
		actual.Routines,
		func(item model.RoutineDef) string { return model.RoutineKey(item.Schema, item.Name, item.IdentityArgs) },
		func(item model.RoutineDef) Operation {
			return op("create-routine", item.Kind, model.RoutineKey(item.Schema, item.Name, item.IdentityArgs), "safe", ensureSemicolon(item.SQL))
		},
		func(item model.RoutineDef) Operation {
			return op("drop-routine", item.Kind, model.RoutineKey(item.Schema, item.Name, item.IdentityArgs), "destructive", fmt.Sprintf("DROP %s %s(%s);", strings.ToUpper(item.Kind), quoteQName(item.Schema, item.Name), item.IdentityArgs))
		},
		func(a, b model.RoutineDef) bool { return routineEqual(a, b) },
		&ops,
		options,
		permissionCreate,
	)

	diffByName(
		project,
		desired.Enums,
		actual.Enums,
		func(item model.EnumDef) string { return model.QualifiedName(item.Schema, item.Name) },
		func(item model.EnumDef) Operation {
			return op("create-enum", "enum", model.QualifiedName(item.Schema, item.Name), "safe", ensureSemicolon(item.SQL))
		},
		func(item model.EnumDef) Operation {
			return op("drop-enum", "enum", model.QualifiedName(item.Schema, item.Name), "destructive", fmt.Sprintf("DROP TYPE %s;", quoteQName(item.Schema, item.Name)))
		},
		func(a, b model.EnumDef) bool { return enumEqual(a, b) },
		&ops,
		options,
		permissionCreate,
	)

	diffByName(
		project,
		desired.Domains,
		actual.Domains,
		func(item model.DomainDef) string { return model.QualifiedName(item.Schema, item.Name) },
		func(item model.DomainDef) Operation {
			return op("create-domain", "domain", model.QualifiedName(item.Schema, item.Name), "safe", ensureSemicolon(item.SQL))
		},
		func(item model.DomainDef) Operation {
			return op("drop-domain", "domain", model.QualifiedName(item.Schema, item.Name), "destructive", fmt.Sprintf("DROP DOMAIN %s;", quoteQName(item.Schema, item.Name)))
		},
		func(a, b model.DomainDef) bool { return domainEqual(a, b) },
		&ops,
		options,
		permissionCreate,
	)

	diffByName(
		project,
		desired.Sequences,
		actual.Sequences,
		func(item model.SequenceDef) string { return model.QualifiedName(item.Schema, item.Name) },
		func(item model.SequenceDef) Operation {
			return op("create-sequence", "sequence", model.QualifiedName(item.Schema, item.Name), "safe", ensureSemicolon(item.SQL))
		},
		func(item model.SequenceDef) Operation {
			return op("drop-sequence", "sequence", model.QualifiedName(item.Schema, item.Name), "destructive", fmt.Sprintf("DROP SEQUENCE %s;", quoteQName(item.Schema, item.Name)))
		},
		func(a, b model.SequenceDef) bool { return a.SQL == b.SQL },
		&ops,
		options,
		permissionCreate,
	)

	diffByName(
		project,
		desired.Comments,
		actual.Comments,
		func(item model.CommentDef) string { return item.ObjectType + ":" + item.ObjectKey },
		func(item model.CommentDef) Operation {
			return op("set-comment", item.ObjectType, item.ObjectKey, "safe", ensureSemicolon(item.SQL))
		},
		func(item model.CommentDef) Operation {
			return op("clear-comment", item.ObjectType, item.ObjectKey, "safe", clearCommentSQL(item))
		},
		func(a, b model.CommentDef) bool { return a.Comment == b.Comment },
		&ops,
		options,
		permissionAlter,
	)

	sort.Slice(ops, func(i, j int) bool {
		if weight(ops[i].Kind) == weight(ops[j].Kind) {
			return ops[i].ObjectKey < ops[j].ObjectKey
		}
		return weight(ops[i].Kind) < weight(ops[j].Kind)
	})

	supported := true
	destructive := false
	for _, operation := range ops {
		if strings.HasPrefix(operation.Kind, "blocked-") {
			supported = false
		}
		if operation.Risk == "destructive" {
			destructive = true
		}
	}

	return Plan{
		Summary: Summary{
			Supported:      supported,
			Destructive:    destructive,
			OperationCount: len(ops),
		},
		Operations: ops,
	}
}

func diffByName[T any](project *projectxml.Project, desired, actual []T, key func(T) string, createOp func(T) Operation, dropOp func(T) Operation, equal func(T, T) bool, ops *[]Operation, options Options, createPermission planPermission) {
	desiredMap := make(map[string]T, len(desired))
	for _, item := range desired {
		desiredMap[key(item)] = item
	}
	actualMap := make(map[string]T, len(actual))
	for _, item := range actual {
		actualMap[key(item)] = item
	}

	for name, desiredItem := range desiredMap {
		actualItem, exists := actualMap[name]
		if !exists {
			*ops = append(*ops, requirePermission(createOp(desiredItem), project, createPermission))
			continue
		}
		if !equal(desiredItem, actualItem) {
			*ops = append(*ops, recreateOps(createOp(desiredItem), dropOp(actualItem), project, options)...)
		}
	}

	for name, actualItem := range actualMap {
		if _, exists := desiredMap[name]; exists {
			continue
		}
		operation := dropOp(actualItem)
		if operation.Risk != "destructive" {
			*ops = append(*ops, requirePermission(operation, project, permissionAlter))
			continue
		}
		if !(project.Target.Plan.AllowDrop || options.AllowDrop) {
			operation = blockOperation(operation, "drops are disabled by the project and invocation")
		}
		*ops = append(*ops, operation)
	}
}

func diffTables(project *projectxml.Project, desired, actual []model.TableDef, ops *[]Operation, options Options) {
	desiredMap := make(map[string]model.TableDef, len(desired))
	for _, item := range desired {
		desiredMap[model.QualifiedName(item.Schema, item.Name)] = item
	}
	actualMap := make(map[string]model.TableDef, len(actual))
	for _, item := range actual {
		actualMap[model.QualifiedName(item.Schema, item.Name)] = item
	}

	for name, desiredItem := range desiredMap {
		actualItem, exists := actualMap[name]
		if !exists {
			create := op("create-table", "table", name, "safe", ensureSemicolon(desiredItem.SQL))
			*ops = append(*ops, requirePermission(create, project, permissionCreate))
			continue
		}
		if tableEqual(desiredItem, actualItem) {
			continue
		}
		if alterOps, ok := alterTableOps(desiredItem, actualItem); ok {
			for index := range alterOps {
				alterOps[index] = requirePermission(alterOps[index], project, permissionAlter)
				if alterOps[index].Risk == "destructive" && !(project.Target.Plan.AllowDrop || options.AllowDrop) {
					alterOps[index] = blockOperation(alterOps[index], "requires --allow-drop because a column would be removed")
				}
			}
			*ops = append(*ops, alterOps...)
			continue
		}
		create := op("create-table", "table", name, "safe", ensureSemicolon(desiredItem.SQL))
		drop := op("drop-table", "table", name, "destructive", fmt.Sprintf("DROP TABLE %s CASCADE;", quoteQName(actualItem.Schema, actualItem.Name)))
		*ops = append(*ops, recreateOps(create, drop, project, options)...)
	}

	for name, actualItem := range actualMap {
		if _, exists := desiredMap[name]; exists {
			continue
		}
		drop := op("drop-table", "table", name, "destructive", fmt.Sprintf("DROP TABLE %s CASCADE;", quoteQName(actualItem.Schema, actualItem.Name)))
		if project.Target.Plan.AllowDrop || options.AllowDrop {
			*ops = append(*ops, drop)
			continue
		}
		drop.Kind = "blocked-" + drop.Kind
		drop.SQL = "-- " + drop.SQL
		*ops = append(*ops, drop)
	}
}

func recreateOps(create Operation, drop Operation, project *projectxml.Project, options Options) []Operation {
	if create.ObjectType == "view" {
		create.Kind = "replace-view"
		create.SQL = ensureOrReplace(create.SQL, "CREATE VIEW", "CREATE OR REPLACE VIEW")
		return []Operation{requirePermission(create, project, permissionAlter)}
	}
	if create.ObjectType == "function" || create.ObjectType == "procedure" {
		needle := "CREATE " + strings.ToUpper(create.ObjectType)
		replacement := "CREATE OR REPLACE " + strings.ToUpper(create.ObjectType)
		create.Kind = "replace-routine"
		create.SQL = ensureOrReplace(create.SQL, needle, replacement)
		return []Operation{requirePermission(create, project, permissionAlter)}
	}
	if !project.Target.Plan.AllowAlter {
		return []Operation{
			blockOperation(drop, "alters are disabled by the project"),
			blockOperation(create, "alters are disabled by the project"),
		}
	}

	if project.Target.Plan.AllowDrop || options.AllowDrop {
		create.Risk = "destructive"
		return []Operation{drop, create}
	}
	drop.Kind = "blocked-" + drop.Kind
	drop.SQL = "-- " + drop.SQL
	create.Kind = "blocked-recreate-" + create.Kind
	create.Risk = "destructive"
	create.SQL = "-- requires destructive recreate\n-- " + create.SQL
	return []Operation{drop, create}
}

func requirePermission(operation Operation, project *projectxml.Project, permission planPermission) Operation {
	allowed := project.Target.Plan.AllowCreate
	reason := "creates are disabled by the project"
	if permission == permissionAlter {
		allowed = project.Target.Plan.AllowAlter
		reason = "alters are disabled by the project"
	}
	if allowed {
		return operation
	}
	return blockOperation(operation, reason)
}

func blockOperation(operation Operation, reason string) Operation {
	if !strings.HasPrefix(operation.Kind, "blocked-") {
		operation.Kind = "blocked-" + operation.Kind
	}
	if operation.SQL != "" && !strings.HasPrefix(strings.TrimSpace(operation.SQL), "--") {
		operation.SQL = "-- " + reason + "\n-- " + strings.ReplaceAll(operation.SQL, "\n", "\n-- ")
	}
	return operation
}

func ensureSemicolon(sql string) string {
	sql = strings.TrimSpace(sql)
	if strings.HasSuffix(sql, ";") {
		return sql
	}
	return sql + ";"
}

func ensureOrReplace(sql, needle, replacement string) string {
	sql = ensureSemicolon(sql)
	up := strings.ToUpper(sql)
	if strings.HasPrefix(up, replacement) {
		return sql
	}
	if strings.HasPrefix(up, needle) {
		return replacement + sql[len(needle):]
	}
	return sql
}

func clearCommentSQL(comment model.CommentDef) string {
	switch comment.ObjectType {
	case "table":
		return fmt.Sprintf("COMMENT ON TABLE %s IS NULL;", quoteQName(strings.Split(comment.ObjectKey, ".")...))
	case "column":
		parts := strings.Split(comment.ObjectKey, ".")
		if len(parts) == 3 {
			return fmt.Sprintf("COMMENT ON COLUMN %s.%s.%s IS NULL;", quoteQName(parts[0]), quoteQName(parts[1]), quoteQName(parts[2]))
		}
	}
	return "-- unsupported comment clear;"
}

func op(kind, objectType, objectKey, risk, sql string) Operation {
	return Operation{Kind: kind, ObjectType: objectType, ObjectKey: objectKey, Risk: risk, SQL: sql}
}

func quoteQName(parts ...string) string {
	quoted := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		quoted = append(quoted, `"`+strings.ReplaceAll(part, `"`, `""`)+`"`)
	}
	return strings.Join(quoted, ".")
}

func weight(kind string) int {
	switch {
	case strings.Contains(kind, "create-schema"):
		return 10
	case strings.Contains(kind, "create-extension"):
		return 20
	case strings.Contains(kind, "create-enum"), strings.Contains(kind, "create-domain"), strings.Contains(kind, "create-sequence"):
		return 30
	case strings.Contains(kind, "drop-table"):
		return 40
	case strings.Contains(kind, "drop-constraint"):
		return 41
	case strings.Contains(kind, "drop-primary-key"):
		return 41
	case strings.Contains(kind, "drop-column-default"):
		return 42
	case strings.Contains(kind, "alter-column-type"):
		return 43
	case strings.Contains(kind, "set-column-default"):
		return 44
	case strings.Contains(kind, "column-not-null"):
		return 45
	case strings.Contains(kind, "add-column"):
		return 46
	case strings.Contains(kind, "add-primary-key"):
		return 47
	case strings.Contains(kind, "add-constraint"):
		return 47
	case strings.Contains(kind, "drop-column"):
		return 48
	case strings.Contains(kind, "create-table"):
		return 50
	case strings.Contains(kind, "replace-view"), strings.Contains(kind, "create-view"):
		return 60
	case strings.Contains(kind, "replace-routine"), strings.Contains(kind, "create-routine"):
		return 70
	case strings.Contains(kind, "create-index"):
		return 80
	case strings.Contains(kind, "set-comment"), strings.Contains(kind, "clear-comment"):
		return 90
	default:
		return 100
	}
}

func extensionEqual(a, b model.ExtensionDef) bool {
	if a.Name != b.Name {
		return false
	}
	if a.Version == "" || b.Version == "" {
		return true
	}
	return a.Version == b.Version
}

func viewEqual(a, b model.ViewDef) bool {
	return normalizeQuotedDDL(a.SQL) == normalizeQuotedDDL(b.SQL)
}

func routineEqual(a, b model.RoutineDef) bool {
	return normalizeRoutineSQL(a.SQL) == normalizeRoutineSQL(b.SQL)
}

func enumEqual(a, b model.EnumDef) bool {
	if a.Schema != b.Schema || a.Name != b.Name {
		return false
	}
	if len(a.Values) == 0 || len(b.Values) == 0 {
		return a.SQL == b.SQL
	}
	if len(a.Values) != len(b.Values) {
		return false
	}
	for i := range a.Values {
		if a.Values[i] != b.Values[i] {
			return false
		}
	}
	return true
}

func domainEqual(a, b model.DomainDef) bool {
	return normalizeDomainSQL(a.SQL) == normalizeDomainSQL(b.SQL)
}

func normalizeDomainSQL(sql string) string {
	sql = normalizeQuotedDDL(sql)
	sql = strings.Join(strings.Fields(sql), " ")
	for strings.Contains(sql, " NOT NULL NOT NULL") {
		sql = strings.ReplaceAll(sql, " NOT NULL NOT NULL", " NOT NULL")
	}
	return sql
}

func normalizeQuotedDDL(sql string) string {
	sql = strings.ReplaceAll(sql, `"`, "")
	return strings.Join(strings.Fields(sql), " ")
}

func normalizeRoutineSQL(sql string) string {
	sql = normalizeQuotedDDL(sql)
	sql = strings.Replace(sql, "CREATE OR REPLACE FUNCTION", "CREATE FUNCTION", 1)
	sql = strings.Replace(sql, "CREATE OR REPLACE PROCEDURE", "CREATE PROCEDURE", 1)
	sql = normalizeDollarQuotes(sql)
	return sql
}

func normalizeDollarQuotes(sql string) string {
	var out strings.Builder
	inTag := false
	for i := 0; i < len(sql); i++ {
		if sql[i] != '$' {
			out.WriteByte(sql[i])
			continue
		}
		j := i + 1
		for j < len(sql) && sql[j] != '$' {
			j++
		}
		if j >= len(sql) {
			out.WriteByte(sql[i])
			continue
		}
		tag := sql[i : j+1]
		if len(tag) >= 2 {
			out.WriteString("$$")
			i = j
			inTag = !inTag
			_ = inTag
			continue
		}
		out.WriteByte(sql[i])
	}
	return out.String()
}

type tableColumnShape struct {
	Name       string
	DataType   string
	DefaultSQL string
	NotNull    bool
	Fragment   string
}

type tableShape struct {
	Columns        []tableColumnShape
	PrimaryKey     []string
	PrimaryKeyName string
	Constraints    []tableConstraintShape
}

type tableConstraintShape struct {
	Name       string
	Kind       string
	Definition string
	Semantic   string
}

func tableEqual(a, b model.TableDef) bool {
	desired, ok := parseCreateTableShape(a.SQL, a.Schema)
	if !ok {
		return a.SQL == b.SQL
	}
	actual, ok := parseCreateTableShape(b.SQL, b.Schema)
	if !ok {
		return a.SQL == b.SQL
	}
	return tableShapeEqual(desired, actual)
}

func alterTableOps(desiredItem, actualItem model.TableDef) ([]Operation, bool) {
	desired, ok := parseCreateTableShape(desiredItem.SQL, desiredItem.Schema)
	if !ok {
		return nil, false
	}
	actual, ok := parseCreateTableShape(actualItem.SQL, actualItem.Schema)
	if !ok {
		return nil, false
	}
	tableKey := model.QualifiedName(desiredItem.Schema, desiredItem.Name)
	tableSQL := quoteQName(desiredItem.Schema, desiredItem.Name)
	desiredColumns := make(map[string]tableColumnShape, len(desired.Columns))
	actualColumns := make(map[string]tableColumnShape, len(actual.Columns))
	for _, column := range desired.Columns {
		desiredColumns[column.Name] = column
	}
	for _, column := range actual.Columns {
		actualColumns[column.Name] = column
	}

	var operations []Operation
	constraintOperations, ok := alterConstraintOps(tableKey, tableSQL, desired.Constraints, actual.Constraints)
	if !ok {
		return nil, false
	}
	operations = append(operations, constraintOperations...)
	if !primaryKeyEqual(desired, actual) && len(actual.PrimaryKey) > 0 {
		constraintName := actual.PrimaryKeyName
		if constraintName == "" {
			constraintName = desiredItem.Name + "_pkey"
		}
		operations = append(operations, op(
			"alter-table-drop-primary-key",
			"constraint",
			tableKey+"."+constraintName,
			"migration",
			fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", tableSQL, quoteQName(constraintName)),
		))
	}

	for _, desiredColumn := range desired.Columns {
		actualColumn, exists := actualColumns[desiredColumn.Name]
		if !exists {
			if containsIdentifier(desired.PrimaryKey, desiredColumn.Name) && strings.Contains(strings.ToUpper(desiredColumn.Fragment), " PRIMARY KEY") {
				return nil, false
			}
			operations = append(operations, op(
				"alter-table-add-column",
				"column",
				tableKey+"."+desiredColumn.Name,
				"safe",
				fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;", tableSQL, desiredColumn.Fragment),
			))
			continue
		}
		operations = append(operations, alterColumnOps(tableKey, tableSQL, desiredColumn, actualColumn)...)
	}

	if !primaryKeyEqual(desired, actual) && len(desired.PrimaryKey) > 0 {
		constraintName := desired.PrimaryKeyName
		if constraintName == "" {
			constraintName = actual.PrimaryKeyName
		}
		if constraintName == "" {
			constraintName = desiredItem.Name + "_pkey"
		}
		columns := make([]string, 0, len(desired.PrimaryKey))
		for _, column := range desired.PrimaryKey {
			columns = append(columns, quoteQName(column))
		}
		operations = append(operations, op(
			"alter-table-add-primary-key",
			"constraint",
			tableKey+"."+constraintName,
			"migration",
			fmt.Sprintf(
				"ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s);",
				tableSQL,
				quoteQName(constraintName),
				strings.Join(columns, ", "),
			),
		))
	}

	for _, actualColumn := range actual.Columns {
		if _, exists := desiredColumns[actualColumn.Name]; exists {
			continue
		}
		operations = append(operations, op(
			"alter-table-drop-column",
			"column",
			tableKey+"."+actualColumn.Name,
			"destructive",
			fmt.Sprintf(
				"ALTER TABLE %s DROP COLUMN %s;",
				tableSQL,
				quoteQName(actualColumn.Name),
			),
		))
	}
	return operations, true
}

func alterConstraintOps(tableKey, tableSQL string, desired, actual []tableConstraintShape) ([]Operation, bool) {
	matchedActual := make([]bool, len(actual))
	matchedDesired := make([]bool, len(desired))
	for desiredIndex, desiredConstraint := range desired {
		for actualIndex, actualConstraint := range actual {
			if matchedActual[actualIndex] || desiredConstraint.Kind != actualConstraint.Kind || desiredConstraint.Semantic != actualConstraint.Semantic {
				continue
			}
			if desiredConstraint.Name != "" && desiredConstraint.Name != actualConstraint.Name {
				continue
			}
			matchedDesired[desiredIndex] = true
			matchedActual[actualIndex] = true
			break
		}
	}

	var operations []Operation
	for index, constraint := range actual {
		if matchedActual[index] {
			continue
		}
		if constraint.Name == "" {
			return nil, false
		}
		operations = append(operations, op(
			"alter-table-drop-constraint",
			"constraint",
			tableKey+"."+constraint.Name,
			"migration",
			fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", tableSQL, quoteQName(constraint.Name)),
		))
	}
	for index, constraint := range desired {
		if matchedDesired[index] {
			continue
		}
		definition := constraint.Definition
		if constraint.Name != "" {
			definition = "CONSTRAINT " + quoteQName(constraint.Name) + " " + definition
		}
		objectName := constraint.Name
		if objectName == "" {
			objectName = constraint.Kind + ":" + constraint.Semantic
		}
		operations = append(operations, op(
			"alter-table-add-constraint",
			"constraint",
			tableKey+"."+objectName,
			"migration",
			fmt.Sprintf("ALTER TABLE %s ADD %s;", tableSQL, definition),
		))
	}
	return operations, true
}

func alterColumnOps(tableKey, tableSQL string, desired, actual tableColumnShape) []Operation {
	columnKey := tableKey + "." + desired.Name
	columnSQL := quoteQName(desired.Name)
	typeChanged := normalizeComparableType(desired.DataType) != normalizeComparableType(actual.DataType)
	defaultChanged := normalizeComparableExpr(desired.DefaultSQL) != normalizeComparableExpr(actual.DefaultSQL)
	var operations []Operation

	if typeChanged && actual.DefaultSQL != "" {
		operations = append(operations, op(
			"alter-table-drop-column-default",
			"column",
			columnKey,
			"migration",
			fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT;", tableSQL, columnSQL),
		))
	}
	if typeChanged {
		operations = append(operations, op(
			"alter-table-alter-column-type",
			"column",
			columnKey,
			"migration",
			fmt.Sprintf(
				"ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s::%s;",
				tableSQL,
				columnSQL,
				desired.DataType,
				columnSQL,
				desired.DataType,
			),
		))
	}
	if typeChanged || defaultChanged {
		if desired.DefaultSQL == "" {
			if !typeChanged {
				operations = append(operations, op(
					"alter-table-drop-column-default",
					"column",
					columnKey,
					"safe",
					fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT;", tableSQL, columnSQL),
				))
			}
		} else {
			operations = append(operations, op(
				"alter-table-set-column-default",
				"column",
				columnKey,
				"safe",
				fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s;", tableSQL, columnSQL, desired.DefaultSQL),
			))
		}
	}
	if desired.NotNull != actual.NotNull {
		kind := "alter-table-drop-column-not-null"
		action := "DROP NOT NULL"
		risk := "safe"
		if desired.NotNull {
			kind = "alter-table-set-column-not-null"
			action = "SET NOT NULL"
			risk = "migration"
		}
		operations = append(operations, op(
			kind,
			"column",
			columnKey,
			risk,
			fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s %s;", tableSQL, columnSQL, action),
		))
	}
	return operations
}

func primaryKeyEqual(desired, actual tableShape) bool {
	if !stringSlicesEqual(desired.PrimaryKey, actual.PrimaryKey) {
		return false
	}
	return desired.PrimaryKeyName == "" || desired.PrimaryKeyName == actual.PrimaryKeyName
}

func containsIdentifier(items []string, candidate string) bool {
	for _, item := range items {
		if item == candidate {
			return true
		}
	}
	return false
}

func tableShapeEqual(a, b tableShape) bool {
	if !primaryKeyEqual(a, b) || !constraintsEqual(a.Constraints, b.Constraints) || len(a.Columns) != len(b.Columns) {
		return false
	}
	for i := range a.Columns {
		if !tableColumnEqual(a.Columns[i], b.Columns[i]) {
			return false
		}
	}
	return true
}

func constraintsEqual(desired, actual []tableConstraintShape) bool {
	if len(desired) != len(actual) {
		return false
	}
	matched := make([]bool, len(actual))
	for _, desiredConstraint := range desired {
		found := false
		for index, actualConstraint := range actual {
			if matched[index] || desiredConstraint.Kind != actualConstraint.Kind || desiredConstraint.Semantic != actualConstraint.Semantic {
				continue
			}
			if desiredConstraint.Name != "" && desiredConstraint.Name != actualConstraint.Name {
				continue
			}
			matched[index] = true
			found = true
			break
		}
		if !found {
			return false
		}
	}
	return true
}

func tableColumnEqual(a, b tableColumnShape) bool {
	return a.Name == b.Name &&
		normalizeComparableType(a.DataType) == normalizeComparableType(b.DataType) &&
		normalizeComparableExpr(a.DefaultSQL) == normalizeComparableExpr(b.DefaultSQL) &&
		a.NotNull == b.NotNull
}

func parseCreateTableShape(sql, schema string) (tableShape, bool) {
	up := strings.ToUpper(sql)
	start := strings.Index(up, "CREATE TABLE")
	if start == -1 {
		return tableShape{}, false
	}
	open := strings.Index(sql[start:], "(")
	if open == -1 {
		return tableShape{}, false
	}
	open += start
	close := findMatchingParen(sql, open)
	if close == -1 {
		return tableShape{}, false
	}
	items := splitTopLevelCommaList(sql[open+1 : close])
	shape := tableShape{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		col, inlinePK, ok := parseTableColumn(item)
		if ok {
			shape.Columns = append(shape.Columns, col)
			if inlinePK {
				shape.PrimaryKey = []string{col.Name}
			}
			continue
		}
		if pk, name, ok := parsePrimaryKeyConstraint(item); ok {
			shape.PrimaryKey = pk
			shape.PrimaryKeyName = name
		}
	}
	constraints, ok := parseNonPrimaryConstraints(sql, schema, open+1, close)
	if !ok {
		return tableShape{}, false
	}
	shape.Constraints = constraints
	return shape, true
}

func parseNonPrimaryConstraints(sql, schema string, bodyStart, bodyEnd int) ([]tableConstraintShape, bool) {
	tree, err := pg_query.Parse(sql)
	if err != nil || len(tree.Stmts) != 1 || tree.Stmts[0].Stmt.GetCreateStmt() == nil {
		return nil, false
	}
	elements := tree.Stmts[0].Stmt.GetCreateStmt().GetTableElts()
	ranges := splitTopLevelCommaRanges(sql[bodyStart:bodyEnd])
	if len(elements) != len(ranges) {
		return nil, false
	}

	var result []tableConstraintShape
	for elementIndex, element := range elements {
		bounds := ranges[elementIndex]
		bounds[0] += bodyStart
		bounds[1] += bodyStart
		columnName := ""
		var nodes []*pg_query.Node
		if column := element.GetColumnDef(); column != nil {
			columnName = column.GetColname()
			nodes = column.GetConstraints()
		} else if constraint := element.GetConstraint(); constraint != nil {
			nodes = []*pg_query.Node{element}
		}
		for nodeIndex, node := range nodes {
			constraint := node.GetConstraint()
			kind := constraintKind(constraint)
			if kind == "" || kind == "primary-key" {
				continue
			}
			start := int(constraint.GetLocation())
			if start < bounds[0] || start >= bounds[1] {
				return nil, false
			}
			end := bounds[1]
			for _, laterNode := range nodes[nodeIndex+1:] {
				later := laterNode.GetConstraint()
				if later == nil {
					continue
				}
				location := int(later.GetLocation())
				if location > start && location < end {
					end = location
					break
				}
			}
			_, definition := stripConstraintName(strings.TrimSpace(sql[start:end]))
			definition = tableConstraintDefinition(kind, definition, columnName, constraint)
			if definition == "" {
				return nil, false
			}
			result = append(result, tableConstraintShape{
				Name:       constraint.GetConname(),
				Kind:       kind,
				Definition: definition,
				Semantic:   normalizeConstraintSemantic(kind, definition, schema),
			})
		}
	}
	return result, true
}

func constraintKind(constraint *pg_query.Constraint) string {
	if constraint == nil {
		return ""
	}
	switch constraint.GetContype() {
	case pg_query.ConstrType_CONSTR_PRIMARY:
		return "primary-key"
	case pg_query.ConstrType_CONSTR_FOREIGN:
		return "foreign-key"
	case pg_query.ConstrType_CONSTR_UNIQUE:
		return "unique"
	case pg_query.ConstrType_CONSTR_CHECK:
		return "check"
	default:
		return ""
	}
}

func stripConstraintName(definition string) (string, string) {
	definition = strings.TrimSpace(definition)
	if !strings.HasPrefix(strings.ToUpper(definition), "CONSTRAINT ") {
		return "", definition
	}
	name, rest, ok := cutIdentifier(strings.TrimSpace(definition[len("CONSTRAINT "):]))
	if !ok {
		return "", definition
	}
	return normalizeIdentifier(name), strings.TrimSpace(rest)
}

func tableConstraintDefinition(kind, definition, columnName string, constraint *pg_query.Constraint) string {
	definition = strings.TrimSpace(definition)
	if columnName == "" {
		return definition
	}
	switch kind {
	case "unique":
		result := "UNIQUE"
		if constraint.GetNullsNotDistinct() {
			result += " NULLS NOT DISTINCT"
		}
		result += " (" + quoteQName(columnName) + ")"
		return result + constraintTiming(constraint)
	case "foreign-key":
		return "FOREIGN KEY (" + quoteQName(columnName) + ") " + definition
	case "check":
		return definition
	default:
		return ""
	}
}

func constraintTiming(constraint *pg_query.Constraint) string {
	if !constraint.GetDeferrable() {
		return ""
	}
	if constraint.GetInitdeferred() {
		return " DEFERRABLE INITIALLY DEFERRED"
	}
	return " DEFERRABLE INITIALLY IMMEDIATE"
}

func normalizeConstraintSemantic(kind, definition, schema string) string {
	definition = strings.TrimSpace(definition)
	if kind == "check" {
		upper := strings.ToUpper(definition)
		if index := strings.Index(upper, "CHECK"); index >= 0 {
			expression := strings.TrimSpace(definition[index+len("CHECK"):])
			expression = stripOuterParentheses(expression)
			definition = "CHECK(" + expression + ")"
		}
	}
	result := compactConstraintSQL(definition)
	if kind == "foreign-key" && schema != "" {
		result = strings.Replace(result, "references"+strings.ToLower(schema)+".", "references", 1)
	}
	return result
}

func stripOuterParentheses(expression string) string {
	expression = strings.TrimSpace(expression)
	for len(expression) >= 2 && expression[0] == '(' {
		close := findMatchingParen(expression, 0)
		if close != len(expression)-1 {
			break
		}
		expression = strings.TrimSpace(expression[1:close])
	}
	return expression
}

func compactConstraintSQL(sql string) string {
	var result strings.Builder
	inSingle := false
	for index := 0; index < len(sql); index++ {
		character := sql[index]
		if inSingle {
			result.WriteByte(character)
			if character == '\'' {
				if index+1 < len(sql) && sql[index+1] == '\'' {
					index++
					result.WriteByte(sql[index])
					continue
				}
				inSingle = false
			}
			continue
		}
		if character == '\'' {
			inSingle = true
			result.WriteByte(character)
			continue
		}
		if character == '"' || character == ' ' || character == '\t' || character == '\n' || character == '\r' {
			continue
		}
		if character >= 'A' && character <= 'Z' {
			character += 'a' - 'A'
		}
		result.WriteByte(character)
	}
	return result.String()
}

func parseTableColumn(fragment string) (tableColumnShape, bool, bool) {
	name, rest, ok := cutIdentifier(fragment)
	if !ok {
		return tableColumnShape{}, false, false
	}
	upperRest := strings.ToUpper(rest)
	if strings.HasPrefix(strings.ToUpper(name), "CONSTRAINT") || strings.HasPrefix(upperRest, "PRIMARY KEY") {
		return tableColumnShape{}, false, false
	}

	defaultIdx := findTopLevelKeyword(upperRest, " DEFAULT ")
	notNullIdx := findTopLevelKeyword(upperRest, " NOT NULL")
	primaryKeyIdx := findTopLevelKeyword(upperRest, " PRIMARY KEY")
	uniqueIdx := findTopLevelKeyword(upperRest, " UNIQUE")
	referencesIdx := findTopLevelKeyword(upperRest, " REFERENCES")
	checkIdx := findTopLevelKeyword(upperRest, " CHECK")
	constraintIdx := findTopLevelKeyword(upperRest, " CONSTRAINT")

	endType := len(rest)
	for _, idx := range []int{defaultIdx, notNullIdx, primaryKeyIdx, uniqueIdx, referencesIdx, checkIdx, constraintIdx} {
		if idx >= 0 && idx < endType {
			endType = idx
		}
	}

	col := tableColumnShape{
		Name:     normalizeIdentifier(name),
		DataType: strings.TrimSpace(rest[:endType]),
		NotNull:  notNullIdx >= 0 || primaryKeyIdx >= 0,
		Fragment: fragment,
	}
	if defaultIdx >= 0 {
		endDefault := len(rest)
		for _, idx := range []int{notNullIdx, primaryKeyIdx, uniqueIdx, referencesIdx, checkIdx, constraintIdx} {
			if idx >= 0 && idx > defaultIdx && idx < endDefault {
				endDefault = idx
			}
		}
		col.DefaultSQL = strings.TrimSpace(rest[defaultIdx+len(" DEFAULT ") : endDefault])
	}
	return col, primaryKeyIdx >= 0, true
}

func parsePrimaryKeyConstraint(fragment string) ([]string, string, bool) {
	upper := strings.ToUpper(fragment)
	idx := strings.Index(upper, "PRIMARY KEY")
	if idx == -1 {
		return nil, "", false
	}
	constraintName := ""
	prefix := strings.TrimSpace(fragment[:idx])
	if strings.HasPrefix(strings.ToUpper(prefix), "CONSTRAINT") {
		name, _, ok := cutIdentifier(strings.TrimSpace(prefix[len("CONSTRAINT"):]))
		if ok {
			constraintName = normalizeIdentifier(name)
		}
	}
	open := strings.Index(fragment[idx:], "(")
	if open == -1 {
		return nil, "", false
	}
	open += idx
	close := findMatchingParen(fragment, open)
	if close == -1 {
		return nil, "", false
	}
	parts := splitTopLevelCommaList(fragment[open+1 : close])
	keys := make([]string, 0, len(parts))
	for _, part := range parts {
		keys = append(keys, normalizeIdentifier(strings.TrimSpace(part)))
	}
	return keys, constraintName, true
}

func cutIdentifier(input string) (string, string, bool) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", false
	}
	if input[0] == '"' {
		end := 1
		for end < len(input) {
			if input[end] == '"' {
				if end+1 < len(input) && input[end+1] == '"' {
					end += 2
					continue
				}
				break
			}
			end++
		}
		if end >= len(input) {
			return "", "", false
		}
		return input[:end+1], strings.TrimSpace(input[end+1:]), true
	}
	for i := 0; i < len(input); i++ {
		if input[i] == ' ' || input[i] == '\t' {
			return input[:i], strings.TrimSpace(input[i+1:]), true
		}
	}
	return input, "", true
}

func normalizeIdentifier(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) && len(s) >= 2 {
		s = s[1 : len(s)-1]
	}
	return strings.ReplaceAll(s, `""`, `"`)
}

func normalizeComparableType(s string) string {
	s = strings.ReplaceAll(strings.TrimSpace(s), `"`, "")
	normalized := strings.ToLower(s)
	switch normalized {
	case "integer":
		return "int"
	case "character varying":
		return "varchar"
	case "boolean":
		return "bool"
	case "double precision":
		return "float8"
	case "real":
		return "float4"
	}
	return normalized
}

func normalizeComparableExpr(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, `"`, "")
	if idx := strings.Index(s, "::"); idx != -1 {
		s = s[:idx]
	}
	return strings.Join(strings.Fields(s), " ")
}

func findMatchingParen(s string, open int) int {
	depth := 0
	inSingle := false
	inDouble := false
	for i := open; i < len(s); i++ {
		switch s[i] {
		case '\'':
			if !inDouble {
				if inSingle && i+1 < len(s) && s[i+1] == '\'' {
					i++
					continue
				}
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '(':
			if !inSingle && !inDouble {
				depth++
			}
		case ')':
			if !inSingle && !inDouble {
				depth--
				if depth == 0 {
					return i
				}
			}
		}
	}
	return -1
}

func splitTopLevelCommaList(s string) []string {
	ranges := splitTopLevelCommaRanges(s)
	parts := make([]string, 0, len(ranges))
	for _, bounds := range ranges {
		parts = append(parts, strings.TrimSpace(s[bounds[0]:bounds[1]]))
	}
	return parts
}

func splitTopLevelCommaRanges(s string) [][2]int {
	var ranges [][2]int
	start := 0
	depth := 0
	inSingle := false
	inDouble := false
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\'':
			if !inDouble {
				if inSingle && i+1 < len(s) && s[i+1] == '\'' {
					i++
					continue
				}
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '(':
			if !inSingle && !inDouble {
				depth++
			}
		case ')':
			if !inSingle && !inDouble && depth > 0 {
				depth--
			}
		case ',':
			if !inSingle && !inDouble && depth == 0 {
				ranges = append(ranges, [2]int{start, i})
				start = i + 1
			}
		}
	}
	ranges = append(ranges, [2]int{start, len(s)})
	return ranges
}

func findTopLevelKeyword(upper, keyword string) int {
	depth := 0
	inSingle := false
	inDouble := false
	for i := 0; i <= len(upper)-len(keyword); i++ {
		switch upper[i] {
		case '\'':
			if !inDouble {
				if inSingle && i+1 < len(upper) && upper[i+1] == '\'' {
					i++
					continue
				}
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '(':
			if !inSingle && !inDouble {
				depth++
			}
		case ')':
			if !inSingle && !inDouble && depth > 0 {
				depth--
			}
		}
		if !inSingle && !inDouble && depth == 0 && strings.HasPrefix(upper[i:], keyword) {
			return i
		}
	}
	return -1
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
