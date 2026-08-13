package diff

import (
	"fmt"
	"sort"
	"strings"

	"github.com/MagnusOpera/starpac/internal/d1/model"
	"github.com/MagnusOpera/starpac/internal/d1/project"
	sharedplan "github.com/MagnusOpera/starpac/internal/pac/plan"
)

type Options struct {
	AllowDrop bool
}

type Plan = sharedplan.Plan
type Summary = sharedplan.Summary
type Operation = sharedplan.Operation

func BuildPlan(project *projectxml.Project, desired, actual *model.SchemaModel, options Options) Plan {
	operations := make([]Operation, 0)
	rebuiltTables := diffTables(project, desired, actual, options, &operations)
	diffIndexes(project, desired.Indexes, actual.Indexes, rebuiltTables, options, &operations)
	diffViews(project, desired.Views, actual.Views, options, &operations)
	diffTriggers(project, desired.Triggers, actual.Triggers, rebuiltTables, options, &operations)

	sort.SliceStable(operations, func(left, right int) bool {
		leftWeight := operationWeight(operations[left].Kind)
		rightWeight := operationWeight(operations[right].Kind)
		if leftWeight == rightWeight {
			return operations[left].ObjectKey < operations[right].ObjectKey
		}
		return leftWeight < rightWeight
	})
	summary := Summary{
		Supported:      true,
		OperationCount: len(operations),
	}
	for _, operation := range operations {
		if operation.Risk == "destructive" {
			summary.Destructive = true
		}
		if strings.HasPrefix(operation.Kind, "blocked-") {
			summary.Supported = false
		}
	}
	return Plan{
		Summary:    summary,
		Operations: operations,
	}
}

func diffTables(
	project *projectxml.Project,
	desired *model.SchemaModel,
	actual *model.SchemaModel,
	options Options,
	operations *[]Operation,
) map[string]bool {
	desiredByName := tablesByName(desired.Tables)
	actualByName := tablesByName(actual.Tables)
	rebuilt := map[string]bool{}
	for name, desiredTable := range desiredByName {
		actualTable, exists := actualByName[name]
		if !exists {
			appendCreate(project, operations, operation(
				"create-table",
				"table",
				name,
				"safe",
				ensureSemicolon(desiredTable.SQL),
			))
			continue
		}
		if tableEqual(desiredTable, actualTable) {
			continue
		}
		if additions, ok := additiveColumns(desiredTable, actualTable); ok {
			for _, column := range additions {
				appendAlter(project, operations, operation(
					"alter-table-add-column",
					"table",
					name,
					"safe",
					fmt.Sprintf(
						"ALTER TABLE %s ADD COLUMN %s;",
						model.QuoteIdentifier(name),
						column.Definition,
					),
				))
			}
			continue
		}
		rebuilt[name] = true
		rebuild := rebuildTableOperation(desiredTable, actualTable, desired)
		if referencingTables(actual.Tables, name) {
			*operations = append(*operations, blocked(
				rebuild,
				"automatic rebuild is unsafe because another table references this table",
			))
			continue
		}
		if rebuild.Risk == "destructive" && !(project.Target.Plan.AllowDrop || options.AllowDrop) {
			*operations = append(*operations, blocked(rebuild, "requires --allow-drop because columns would be removed"))
			continue
		}
		appendAlter(project, operations, rebuild)
	}
	for name := range actualByName {
		if _, exists := desiredByName[name]; exists {
			continue
		}
		drop := operation(
			"drop-table",
			"table",
			name,
			"destructive",
			"DROP TABLE "+model.QuoteIdentifier(name)+";",
		)
		appendDrop(project, options, operations, drop)
	}
	return rebuilt
}

func diffIndexes(
	project *projectxml.Project,
	desired []model.IndexDef,
	actual []model.IndexDef,
	rebuiltTables map[string]bool,
	options Options,
	operations *[]Operation,
) {
	desiredByName := indexesByName(desired)
	actualByName := indexesByName(actual)
	for name, desiredIndex := range desiredByName {
		if rebuiltTables[desiredIndex.TableName] {
			continue
		}
		actualIndex, exists := actualByName[name]
		if !exists {
			appendCreate(project, operations, operation("create-index", "index", name, "safe", ensureSemicolon(desiredIndex.SQL)))
			continue
		}
		if model.NormalizeDDL(desiredIndex.SQL) != model.NormalizeDDL(actualIndex.SQL) {
			appendAlter(project, operations, operation(
				"replace-index",
				"index",
				name,
				"safe",
				"DROP INDEX "+model.QuoteIdentifier(name)+";\n"+ensureSemicolon(desiredIndex.SQL),
			))
		}
	}
	for name, actualIndex := range actualByName {
		if rebuiltTables[actualIndex.TableName] {
			continue
		}
		if _, exists := desiredByName[name]; exists {
			continue
		}
		appendDrop(project, options, operations, operation(
			"drop-index",
			"index",
			name,
			"destructive",
			"DROP INDEX "+model.QuoteIdentifier(name)+";",
		))
	}
}

func diffViews(
	project *projectxml.Project,
	desired []model.ViewDef,
	actual []model.ViewDef,
	options Options,
	operations *[]Operation,
) {
	desiredByName := viewsByName(desired)
	actualByName := viewsByName(actual)
	for name, desiredView := range desiredByName {
		actualView, exists := actualByName[name]
		if !exists {
			appendCreate(project, operations, operation("create-view", "view", name, "safe", ensureSemicolon(desiredView.SQL)))
			continue
		}
		if model.NormalizeDDL(desiredView.SQL) != model.NormalizeDDL(actualView.SQL) {
			appendAlter(project, operations, operation(
				"replace-view",
				"view",
				name,
				"safe",
				"DROP VIEW "+model.QuoteIdentifier(name)+";\n"+ensureSemicolon(desiredView.SQL),
			))
		}
	}
	for name := range actualByName {
		if _, exists := desiredByName[name]; exists {
			continue
		}
		appendDrop(project, options, operations, operation(
			"drop-view",
			"view",
			name,
			"destructive",
			"DROP VIEW "+model.QuoteIdentifier(name)+";",
		))
	}
}

func diffTriggers(
	project *projectxml.Project,
	desired []model.TriggerDef,
	actual []model.TriggerDef,
	rebuiltTables map[string]bool,
	options Options,
	operations *[]Operation,
) {
	desiredByName := triggersByName(desired)
	actualByName := triggersByName(actual)
	for name, desiredTrigger := range desiredByName {
		if rebuiltTables[desiredTrigger.TableName] {
			continue
		}
		actualTrigger, exists := actualByName[name]
		if !exists {
			appendCreate(project, operations, operation("create-trigger", "trigger", name, "safe", ensureSemicolon(desiredTrigger.SQL)))
			continue
		}
		if model.NormalizeDDL(desiredTrigger.SQL) != model.NormalizeDDL(actualTrigger.SQL) {
			appendAlter(project, operations, operation(
				"replace-trigger",
				"trigger",
				name,
				"safe",
				"DROP TRIGGER "+model.QuoteIdentifier(name)+";\n"+ensureSemicolon(desiredTrigger.SQL),
			))
		}
	}
	for name, actualTrigger := range actualByName {
		if rebuiltTables[actualTrigger.TableName] {
			continue
		}
		if _, exists := desiredByName[name]; exists {
			continue
		}
		appendDrop(project, options, operations, operation(
			"drop-trigger",
			"trigger",
			name,
			"destructive",
			"DROP TRIGGER "+model.QuoteIdentifier(name)+";",
		))
	}
}

func tableEqual(desired, actual model.TableDef) bool {
	if !columnsEqual(desired.Columns, actual.Columns) {
		return false
	}
	if !foreignKeysEqual(desired.ForeignKeys, actual.ForeignKeys) {
		return false
	}
	return model.NormalizeDDL(desired.SQL) == model.NormalizeDDL(actual.SQL)
}

func additiveColumns(desired, actual model.TableDef) ([]model.ColumnDef, bool) {
	if len(actual.Columns) >= len(desired.Columns) {
		return nil, false
	}
	for index := range actual.Columns {
		if !columnEqual(desired.Columns[index], actual.Columns[index]) {
			return nil, false
		}
		if model.NormalizeDDL(desired.Columns[index].Definition) != model.NormalizeDDL(actual.Columns[index].Definition) {
			return nil, false
		}
	}
	if !stringSlicesEqual(model.TableConstraints(desired.SQL), model.TableConstraints(actual.SQL)) {
		return nil, false
	}
	for _, foreignKey := range actual.ForeignKeys {
		if !containsForeignKey(desired.ForeignKeys, foreignKey) {
			return nil, false
		}
	}
	additions := desired.Columns[len(actual.Columns):]
	for _, column := range additions {
		if column.Definition == "" || column.PrimaryKey > 0 || column.Hidden != 0 {
			return nil, false
		}
		if column.NotNull && column.DefaultSQL == nil {
			return nil, false
		}
	}
	return additions, true
}

func rebuildTableOperation(desired, actual model.TableDef, schema *model.SchemaModel) Operation {
	temporaryName := "__d1pac_" + desired.Name + "_new"
	createSQL, ok := model.ReplaceCreateTableName(desired.SQL, temporaryName)
	if !ok {
		return blocked(operation(
			"rebuild-table",
			"table",
			desired.Name,
			"migration",
			"",
		), "unable to rewrite CREATE TABLE statement")
	}
	actualColumns := map[string]bool{}
	for _, column := range actual.Columns {
		actualColumns[column.Name] = true
	}
	var commonColumns []string
	for _, column := range desired.Columns {
		if actualColumns[column.Name] {
			commonColumns = append(commonColumns, model.QuoteIdentifier(column.Name))
		}
	}
	statements := []string{ensureSemicolon(createSQL)}
	if len(commonColumns) > 0 {
		columns := strings.Join(commonColumns, ", ")
		statements = append(statements, fmt.Sprintf(
			"INSERT INTO %s (%s) SELECT %s FROM %s;",
			model.QuoteIdentifier(temporaryName),
			columns,
			columns,
			model.QuoteIdentifier(actual.Name),
		))
	}
	statements = append(statements,
		"DROP TABLE "+model.QuoteIdentifier(actual.Name)+";",
		"ALTER TABLE "+model.QuoteIdentifier(temporaryName)+" RENAME TO "+model.QuoteIdentifier(desired.Name)+";",
	)
	for _, index := range schema.Indexes {
		if index.TableName == desired.Name {
			statements = append(statements, ensureSemicolon(index.SQL))
		}
	}
	for _, trigger := range schema.Triggers {
		if trigger.TableName == desired.Name {
			statements = append(statements, ensureSemicolon(trigger.SQL))
		}
	}
	risk := "migration"
	desiredColumns := map[string]bool{}
	for _, column := range desired.Columns {
		desiredColumns[column.Name] = true
	}
	for _, column := range actual.Columns {
		if !desiredColumns[column.Name] {
			risk = "destructive"
			break
		}
	}
	return operation(
		"rebuild-table",
		"table",
		desired.Name,
		risk,
		strings.Join(statements, "\n"),
	)
}

func columnsEqual(left, right []model.ColumnDef) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if !columnEqual(left[index], right[index]) {
			return false
		}
	}
	return true
}

func columnEqual(left, right model.ColumnDef) bool {
	return left.Position == right.Position &&
		left.Name == right.Name &&
		strings.EqualFold(strings.TrimSpace(left.Type), strings.TrimSpace(right.Type)) &&
		left.NotNull == right.NotNull &&
		normalizeOptionalSQL(left.DefaultSQL) == normalizeOptionalSQL(right.DefaultSQL) &&
		left.PrimaryKey == right.PrimaryKey &&
		left.Hidden == right.Hidden
}

func normalizeOptionalSQL(value *string) string {
	if value == nil {
		return ""
	}
	return model.NormalizeDDL(*value)
}

func foreignKeysEqual(left, right []model.ForeignKeyDef) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func containsForeignKey(foreignKeys []model.ForeignKeyDef, candidate model.ForeignKeyDef) bool {
	for _, foreignKey := range foreignKeys {
		if foreignKey == candidate {
			return true
		}
	}
	return false
}

func referencingTables(tables []model.TableDef, targetName string) bool {
	for _, table := range tables {
		if table.Name == targetName {
			continue
		}
		for _, foreignKey := range table.ForeignKeys {
			if foreignKey.Table == targetName {
				return true
			}
		}
	}
	return false
}

func stringSlicesEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func appendCreate(project *projectxml.Project, operations *[]Operation, item Operation) {
	if project.Target.Plan.AllowCreate {
		*operations = append(*operations, item)
		return
	}
	*operations = append(*operations, blocked(item, "creates are disabled by the project"))
}

func appendAlter(project *projectxml.Project, operations *[]Operation, item Operation) {
	if project.Target.Plan.AllowAlter {
		*operations = append(*operations, item)
		return
	}
	*operations = append(*operations, blocked(item, "alters are disabled by the project"))
}

func appendDrop(project *projectxml.Project, options Options, operations *[]Operation, item Operation) {
	if project.Target.Plan.AllowDrop || options.AllowDrop {
		*operations = append(*operations, item)
		return
	}
	*operations = append(*operations, blocked(item, "drops are disabled by the project"))
}

func blocked(item Operation, reason string) Operation {
	item.Kind = "blocked-" + item.Kind
	if item.SQL != "" {
		item.SQL = "-- " + reason + "\n-- " + strings.ReplaceAll(item.SQL, "\n", "\n-- ")
	} else {
		item.SQL = "-- " + reason
	}
	return item
}

func operation(kind, objectType, objectKey, risk, sql string) Operation {
	return Operation{
		Kind:       kind,
		ObjectType: objectType,
		ObjectKey:  objectKey,
		Risk:       risk,
		SQL:        sql,
	}
}

func ensureSemicolon(sql string) string {
	sql = strings.TrimSpace(sql)
	if strings.HasSuffix(sql, ";") {
		return sql
	}
	return sql + ";"
}

func operationWeight(kind string) int {
	switch {
	case strings.Contains(kind, "drop-trigger"):
		return 10
	case strings.Contains(kind, "drop-view"):
		return 20
	case strings.Contains(kind, "drop-index"):
		return 30
	case strings.Contains(kind, "drop-table"):
		return 40
	case strings.Contains(kind, "create-table"), strings.Contains(kind, "rebuild-table"), strings.Contains(kind, "alter-table"):
		return 50
	case strings.Contains(kind, "create-index"), strings.Contains(kind, "replace-index"):
		return 60
	case strings.Contains(kind, "create-view"), strings.Contains(kind, "replace-view"):
		return 70
	case strings.Contains(kind, "create-trigger"), strings.Contains(kind, "replace-trigger"):
		return 80
	default:
		return 100
	}
}

func tablesByName(items []model.TableDef) map[string]model.TableDef {
	result := make(map[string]model.TableDef, len(items))
	for _, item := range items {
		result[item.Name] = item
	}
	return result
}

func indexesByName(items []model.IndexDef) map[string]model.IndexDef {
	result := make(map[string]model.IndexDef, len(items))
	for _, item := range items {
		result[item.Name] = item
	}
	return result
}

func viewsByName(items []model.ViewDef) map[string]model.ViewDef {
	result := make(map[string]model.ViewDef, len(items))
	for _, item := range items {
		result[item.Name] = item
	}
	return result
}

func triggersByName(items []model.TriggerDef) map[string]model.TriggerDef {
	result := make(map[string]model.TriggerDef, len(items))
	for _, item := range items {
		result[item.Name] = item
	}
	return result
}
