package diff

import (
	"fmt"
	"sort"
	"strings"

	"github.com/MagnusOpera/pgpac/internal/model"
	"github.com/MagnusOpera/pgpac/internal/projectxml"
)

type Options struct {
	AllowDrop bool
}

type Plan struct {
	Summary    Summary     `json:"summary"`
	Operations []Operation `json:"operations"`
}

type Summary struct {
	Supported      bool `json:"supported"`
	Destructive    bool `json:"destructive"`
	OperationCount int  `json:"operationCount"`
}

type Operation struct {
	Kind       string `json:"kind"`
	ObjectType string `json:"objectType"`
	ObjectKey  string `json:"objectKey"`
	Risk       string `json:"risk"`
	SQL        string `json:"sql"`
}

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
	)

	sort.Slice(ops, func(i, j int) bool {
		if weight(ops[i].Kind) == weight(ops[j].Kind) {
			return ops[i].ObjectKey < ops[j].ObjectKey
		}
		return weight(ops[i].Kind) < weight(ops[j].Kind)
	})

	destructive := false
	for _, operation := range ops {
		if operation.Risk == "destructive" {
			destructive = true
			break
		}
	}

	return Plan{
		Summary: Summary{
			Supported:      true,
			Destructive:    destructive,
			OperationCount: len(ops),
		},
		Operations: ops,
	}
}

func diffByName[T any](project *projectxml.Project, desired, actual []T, key func(T) string, createOp func(T) Operation, dropOp func(T) Operation, equal func(T, T) bool, ops *[]Operation, options Options) {
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
			*ops = append(*ops, createOp(desiredItem))
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
		if project.Target.Plan.AllowDrop || options.AllowDrop {
			*ops = append(*ops, dropOp(actualItem))
		} else {
			operation := dropOp(actualItem)
			operation.Kind = "blocked-" + operation.Kind
			operation.SQL = "-- " + operation.SQL
			*ops = append(*ops, operation)
		}
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
			*ops = append(*ops, op("create-table", "table", name, "safe", ensureSemicolon(desiredItem.SQL)))
			continue
		}
		if tableEqual(desiredItem, actualItem) {
			continue
		}
		if alterOps, ok := alterTableOps(desiredItem, actualItem); ok {
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
		return []Operation{create}
	}
	if create.ObjectType == "function" || create.ObjectType == "procedure" {
		needle := "CREATE " + strings.ToUpper(create.ObjectType)
		replacement := "CREATE OR REPLACE " + strings.ToUpper(create.ObjectType)
		create.Kind = "replace-routine"
		create.SQL = ensureOrReplace(create.SQL, needle, replacement)
		return []Operation{create}
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
	Columns    []tableColumnShape
	PrimaryKey []string
}

func tableEqual(a, b model.TableDef) bool {
	desired, ok := parseCreateTableShape(a.SQL)
	if !ok {
		return a.SQL == b.SQL
	}
	actual, ok := parseCreateTableShape(b.SQL)
	if !ok {
		return a.SQL == b.SQL
	}
	return tableShapeEqual(desired, actual)
}

func alterTableOps(desiredItem, actualItem model.TableDef) ([]Operation, bool) {
	desired, ok := parseCreateTableShape(desiredItem.SQL)
	if !ok {
		return nil, false
	}
	actual, ok := parseCreateTableShape(actualItem.SQL)
	if !ok {
		return nil, false
	}
	if len(actual.Columns) > len(desired.Columns) {
		return nil, false
	}
	if !stringSlicesEqual(desired.PrimaryKey, actual.PrimaryKey) {
		return nil, false
	}
	for i := range actual.Columns {
		if !tableColumnEqual(desired.Columns[i], actual.Columns[i]) {
			return nil, false
		}
	}
	var ops []Operation
	objectKey := model.QualifiedName(desiredItem.Schema, desiredItem.Name)
	for _, col := range desired.Columns[len(actual.Columns):] {
		ops = append(ops, op(
			"alter-table-add-column",
			"table",
			objectKey,
			"safe",
			fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;", quoteQName(desiredItem.Schema, desiredItem.Name), col.Fragment),
		))
	}
	return ops, len(ops) > 0
}

func tableShapeEqual(a, b tableShape) bool {
	if !stringSlicesEqual(a.PrimaryKey, b.PrimaryKey) || len(a.Columns) != len(b.Columns) {
		return false
	}
	for i := range a.Columns {
		if !tableColumnEqual(a.Columns[i], b.Columns[i]) {
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

func parseCreateTableShape(sql string) (tableShape, bool) {
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
		if pk, ok := parsePrimaryKeyConstraint(item); ok {
			shape.PrimaryKey = pk
		}
	}
	return shape, true
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

	endType := len(rest)
	for _, idx := range []int{defaultIdx, notNullIdx, primaryKeyIdx} {
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
		for _, idx := range []int{notNullIdx, primaryKeyIdx} {
			if idx >= 0 && idx > defaultIdx && idx < endDefault {
				endDefault = idx
			}
		}
		col.DefaultSQL = strings.TrimSpace(rest[defaultIdx+len(" DEFAULT ") : endDefault])
	}
	return col, primaryKeyIdx >= 0, true
}

func parsePrimaryKeyConstraint(fragment string) ([]string, bool) {
	upper := strings.ToUpper(fragment)
	idx := strings.Index(upper, "PRIMARY KEY")
	if idx == -1 {
		return nil, false
	}
	open := strings.Index(fragment[idx:], "(")
	if open == -1 {
		return nil, false
	}
	open += idx
	close := findMatchingParen(fragment, open)
	if close == -1 {
		return nil, false
	}
	parts := splitTopLevelCommaList(fragment[open+1 : close])
	keys := make([]string, 0, len(parts))
	for _, part := range parts {
		keys = append(keys, normalizeIdentifier(strings.TrimSpace(part)))
	}
	return keys, true
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
	switch strings.ToLower(s) {
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
	return s
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
	var parts []string
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
				parts = append(parts, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	parts = append(parts, strings.TrimSpace(s[start:]))
	return parts
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
