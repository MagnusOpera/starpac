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
	Strict    bool
}

type Plan = sharedplan.Plan
type Summary = sharedplan.Summary
type Operation = sharedplan.Operation

func BuildPlan(project *projectxml.Project, desired, actual *model.SchemaModel, options Options) Plan {
	operations := make([]Operation, 0)
	rebuiltTables, rebuiltTriggers := diffTables(project, desired, actual, options, &operations)
	diffIndexes(project, desired.Indexes, actual.Indexes, rebuiltTables, options, &operations)
	diffViews(project, desired.Views, actual.Views, options, &operations)
	diffTriggers(project, desired.Triggers, actual.Triggers, rebuiltTables, rebuiltTriggers, options, &operations)

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
) (map[string]bool, map[string]bool) {
	desiredByName := tablesByName(desired.Tables)
	actualByName := tablesByName(actual.Tables)
	rebuilt := map[string]bool{}
	rebuiltTriggers := map[string]bool{}
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
		if tableEqual(desiredTable, actualTable, options.Strict) {
			continue
		}
		if additions, ok := additiveColumns(desiredTable, actualTable, options.Strict); ok {
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
		rebuild, dependentTriggers := rebuildTableOperation(desiredTable, actualTable, desired, actual)
		for _, triggerName := range dependentTriggers {
			rebuiltTriggers[triggerName] = true
		}
		referencingTables := referencingRetainedTableNames(desired.Tables, actual.Tables, name)
		if len(referencingTables) > 0 {
			*operations = append(*operations, blocked(
				rebuild,
				referencedTableRebuildReason(
					desiredTable,
					actualTable,
					referencingTables,
				),
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
	return rebuilt, rebuiltTriggers
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
	rebuiltTriggers map[string]bool,
	options Options,
	operations *[]Operation,
) {
	desiredByName := triggersByName(desired)
	actualByName := triggersByName(actual)
	for name, desiredTrigger := range desiredByName {
		if rebuiltTriggers[name] {
			continue
		}
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
		if rebuiltTriggers[name] {
			continue
		}
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

func tableEqual(desired, actual model.TableDef, strict bool) bool {
	if strict {
		return columnsEqual(desired.Columns, actual.Columns) &&
			foreignKeysEqual(desired.ForeignKeys, actual.ForeignKeys) &&
			model.NormalizeDDL(desired.SQL) == model.NormalizeDDL(actual.SQL)
	}
	if !columnsEqualByName(desired.Columns, actual.Columns) {
		return false
	}
	if !foreignKeysEqualByDefinition(desired.ForeignKeys, actual.ForeignKeys) {
		return false
	}
	return stringSlicesEqual(model.TableConstraints(desired.SQL), model.TableConstraints(actual.SQL)) &&
		model.TableOptions(desired.SQL) == model.TableOptions(actual.SQL)
}

func additiveColumns(desired, actual model.TableDef, strict bool) ([]model.ColumnDef, bool) {
	if len(actual.Columns) >= len(desired.Columns) {
		return nil, false
	}
	if strict {
		for index := range actual.Columns {
			if !columnEqual(desired.Columns[index], actual.Columns[index]) ||
				!columnDefinitionsEqual(desired.Columns[index], actual.Columns[index]) {
				return nil, false
			}
		}
	}
	if !stringSlicesEqual(model.TableConstraints(desired.SQL), model.TableConstraints(actual.SQL)) {
		return nil, false
	}
	if model.TableOptions(desired.SQL) != model.TableOptions(actual.SQL) {
		return nil, false
	}
	for _, foreignKey := range actual.ForeignKeys {
		if strict && !containsForeignKey(desired.ForeignKeys, foreignKey) {
			return nil, false
		}
		if !strict && !containsForeignKeyDefinition(desired.ForeignKeys, foreignKey) {
			return nil, false
		}
	}
	if strict {
		additions := desired.Columns[len(actual.Columns):]
		for _, column := range additions {
			if !addableColumn(column) {
				return nil, false
			}
		}
		return additions, true
	}

	desiredByName := columnsByName(desired.Columns)
	actualByName := columnsByName(actual.Columns)
	for _, actualColumn := range actual.Columns {
		desiredColumn, exists := desiredByName[actualColumn.Name]
		if !exists || !columnEqualIgnoringPosition(desiredColumn, actualColumn) ||
			!columnDefinitionsEqual(desiredColumn, actualColumn) {
			return nil, false
		}
	}
	additions := make([]model.ColumnDef, 0, len(desired.Columns)-len(actual.Columns))
	for _, column := range desired.Columns {
		if _, exists := actualByName[column.Name]; exists {
			continue
		}
		additions = append(additions, column)
	}
	for _, column := range additions {
		if !addableColumn(column) {
			return nil, false
		}
	}
	return additions, true
}

func referencedTableRebuildReason(
	desired model.TableDef,
	actual model.TableDef,
	referencingTables []string,
) string {
	insertedColumns, ok := nonTrailingColumnAdditions(desired, actual)
	if ok {
		return fmt.Sprintf(
			"non-destructive migration must rebuild the table to place new column(s) %s in the declared order; automatic rebuild is unsafe because retained table(s) %s reference this table",
			quotedNames(insertedColumns),
			quotedNames(referencingTables),
		)
	}
	return fmt.Sprintf(
		"automatic rebuild is unsafe because retained table(s) %s reference this table",
		quotedNames(referencingTables),
	)
}

func nonTrailingColumnAdditions(desired, actual model.TableDef) ([]string, bool) {
	if len(actual.Columns) >= len(desired.Columns) {
		return nil, false
	}
	if !stringSlicesEqual(model.TableConstraints(desired.SQL), model.TableConstraints(actual.SQL)) {
		return nil, false
	}
	for _, foreignKey := range actual.ForeignKeys {
		if !containsForeignKey(desired.ForeignKeys, foreignKey) {
			return nil, false
		}
	}

	insertedColumns := make([]string, 0)
	desiredIndex := 0
	for _, actualColumn := range actual.Columns {
		for desiredIndex < len(desired.Columns) && desired.Columns[desiredIndex].Name != actualColumn.Name {
			if !addableColumn(desired.Columns[desiredIndex]) {
				return nil, false
			}
			insertedColumns = append(insertedColumns, desired.Columns[desiredIndex].Name)
			desiredIndex++
		}
		if desiredIndex == len(desired.Columns) {
			return nil, false
		}
		desiredColumn := desired.Columns[desiredIndex]
		if !columnEqual(desiredColumn, actualColumn) {
			return nil, false
		}
		if model.NormalizeDDL(desiredColumn.Definition) != model.NormalizeDDL(actualColumn.Definition) {
			return nil, false
		}
		desiredIndex++
	}
	for ; desiredIndex < len(desired.Columns); desiredIndex++ {
		if !addableColumn(desired.Columns[desiredIndex]) {
			return nil, false
		}
	}
	if len(insertedColumns) == 0 {
		return nil, false
	}
	return insertedColumns, true
}

func addableColumn(column model.ColumnDef) bool {
	if column.Definition == "" || column.PrimaryKey > 0 || column.Hidden != 0 {
		return false
	}
	return !column.NotNull || column.DefaultSQL != nil
}

func quotedNames(names []string) string {
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		quoted = append(quoted, fmt.Sprintf("%q", name))
	}
	return strings.Join(quoted, ", ")
}

func rebuildTableOperation(desired, actual model.TableDef, desiredSchema, actualSchema *model.SchemaModel) (Operation, []string) {
	temporaryName := "__d1pac_" + desired.Name + "_new"
	createSQL, ok := model.ReplaceCreateTableName(desired.SQL, temporaryName)
	if !ok {
		return blocked(operation(
			"rebuild-table",
			"table",
			desired.Name,
			"migration",
			"",
		), "unable to rewrite CREATE TABLE statement"), nil
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
	desiredTriggers := triggersByName(desiredSchema.Triggers)
	desiredTables := tablesByName(desiredSchema.Tables)
	dependentTriggers := make([]string, 0)
	for _, trigger := range actualSchema.Triggers {
		if trigger.TableName == desired.Name || !referencesIdentifier(trigger.SQL, desired.Name) {
			continue
		}
		if _, retained := desiredTables[trigger.TableName]; !retained {
			continue
		}
		dependentTriggers = append(dependentTriggers, trigger.Name)
	}
	sort.Strings(dependentTriggers)
	statements := make([]string, 0)
	for _, triggerName := range dependentTriggers {
		statements = append(statements, "DROP TRIGGER "+model.QuoteIdentifier(triggerName)+";")
	}
	statements = append(statements, ensureSemicolon(createSQL))
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
	for _, index := range desiredSchema.Indexes {
		if index.TableName == desired.Name {
			statements = append(statements, ensureSemicolon(index.SQL))
		}
	}
	for _, trigger := range desiredSchema.Triggers {
		if trigger.TableName == desired.Name {
			statements = append(statements, ensureSemicolon(trigger.SQL))
		}
	}
	for _, triggerName := range dependentTriggers {
		if trigger, exists := desiredTriggers[triggerName]; exists {
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
	), dependentTriggers
}

func referencesIdentifier(sql, identifier string) bool {
	for _, token := range strings.FieldsFunc(sql, func(character rune) bool {
		return !(character == '_' || character >= '0' && character <= '9' || character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z')
	}) {
		if strings.EqualFold(token, identifier) {
			return true
		}
	}
	return false
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

func columnsEqualByName(left, right []model.ColumnDef) bool {
	if len(left) != len(right) {
		return false
	}
	rightByName := columnsByName(right)
	for _, leftColumn := range left {
		rightColumn, exists := rightByName[leftColumn.Name]
		if !exists || !columnEqualIgnoringPosition(leftColumn, rightColumn) ||
			!columnDefinitionsEqual(leftColumn, rightColumn) {
			return false
		}
	}
	return true
}

func columnsByName(columns []model.ColumnDef) map[string]model.ColumnDef {
	byName := make(map[string]model.ColumnDef, len(columns))
	for _, column := range columns {
		byName[column.Name] = column
	}
	return byName
}

func columnEqual(left, right model.ColumnDef) bool {
	return left.Position == right.Position &&
		columnEqualIgnoringPosition(left, right)
}

func columnEqualIgnoringPosition(left, right model.ColumnDef) bool {
	return left.Name == right.Name &&
		strings.EqualFold(strings.TrimSpace(left.Type), strings.TrimSpace(right.Type)) &&
		left.NotNull == right.NotNull &&
		normalizeOptionalSQL(left.DefaultSQL) == normalizeOptionalSQL(right.DefaultSQL) &&
		left.PrimaryKey == right.PrimaryKey &&
		left.Hidden == right.Hidden
}

func columnDefinitionsEqual(left, right model.ColumnDef) bool {
	return model.NormalizeDDL(left.Definition) == model.NormalizeDDL(right.Definition)
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

func foreignKeysEqualByDefinition(left, right []model.ForeignKeyDef) bool {
	if len(left) != len(right) {
		return false
	}
	matched := make([]bool, len(right))
	for _, leftForeignKey := range left {
		found := false
		for index, rightForeignKey := range right {
			if !matched[index] && foreignKeyDefinitionEqual(leftForeignKey, rightForeignKey) {
				matched[index] = true
				found = true
				break
			}
		}
		if !found {
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

func containsForeignKeyDefinition(foreignKeys []model.ForeignKeyDef, candidate model.ForeignKeyDef) bool {
	for _, foreignKey := range foreignKeys {
		if foreignKeyDefinitionEqual(foreignKey, candidate) {
			return true
		}
	}
	return false
}

func foreignKeyDefinitionEqual(left, right model.ForeignKeyDef) bool {
	return left.Table == right.Table &&
		left.From == right.From &&
		left.To == right.To &&
		left.OnUpdate == right.OnUpdate &&
		left.OnDelete == right.OnDelete &&
		left.Match == right.Match
}

func referencingRetainedTableNames(
	desiredTables []model.TableDef,
	actualTables []model.TableDef,
	targetName string,
) []string {
	desiredByName := tablesByName(desiredTables)
	referencingTables := make([]string, 0)
	for _, table := range actualTables {
		if table.Name == targetName {
			continue
		}
		if _, retained := desiredByName[table.Name]; !retained {
			continue
		}
		for _, foreignKey := range table.ForeignKeys {
			if foreignKey.Table == targetName {
				referencingTables = append(referencingTables, table.Name)
				break
			}
		}
	}
	sort.Strings(referencingTables)
	return referencingTables
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
	item.Reason = reason
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
