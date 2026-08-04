package parser

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v6"

	"github.com/MagnusOpera/pgpac/internal/model"
	"github.com/MagnusOpera/pgpac/internal/projectxml"
)

var (
	tableCommentPattern     = regexp.MustCompile(`(?i)^COMMENT ON TABLE ([a-zA-Z0-9_".]+) IS ('(?:[^']|'')*'|NULL)$`)
	columnCommentPattern    = regexp.MustCompile(`(?i)^COMMENT ON COLUMN ([a-zA-Z0-9_".]+)\.([a-zA-Z0-9_".]+)\.([a-zA-Z0-9_".]+) IS ('(?:[^']|'')*'|NULL)$`)
	extensionVersionPattern = regexp.MustCompile(`(?i)\bVERSION\s+("?[^"\s]+"?|'[^']+')`)
)

func BuildDesiredModel(project *projectxml.Project) (*model.SchemaModel, error) {
	files, err := project.ResolveFiles()
	if err != nil {
		return nil, err
	}

	result := &model.SchemaModel{PostgresVersion: project.PostgresVersion}
	for _, file := range files {
		contentBytes, err := os.ReadFile(file.AbsPath)
		if err != nil {
			return nil, err
		}
		content := string(contentBytes)
		statements, err := pg_query.SplitWithParser(content, true)
		if err != nil {
			return nil, fmt.Errorf("%s: split sql: %w", file.RelPath, err)
		}
		for _, stmtText := range statements {
			if strings.TrimSpace(stmtText) == "" {
				continue
			}
			if err := addStatement(result, project.DefaultSchema, stmtText); err != nil {
				return nil, fmt.Errorf("%s: %w", file.RelPath, err)
			}
		}
	}

	model.Sort(result)
	return result, nil
}

func addStatement(m *model.SchemaModel, defaultSchema, stmtText string) error {
	tree, err := pg_query.Parse(stmtText)
	if err != nil {
		return err
	}
	if len(tree.Stmts) != 1 {
		return fmt.Errorf("expected exactly one statement, got %d", len(tree.Stmts))
	}

	sql, err := pg_query.Deparse(tree)
	if err != nil {
		return err
	}
	sql = model.CanonicalSQL(sql)
	stmt := tree.Stmts[0].Stmt

	switch {
	case stmt.GetCreateSchemaStmt() != nil:
		node := stmt.GetCreateSchemaStmt()
		m.Schemas = append(m.Schemas, model.SchemaDef{Name: node.GetSchemaname(), SQL: sql})
		return nil
	case stmt.GetCreateExtensionStmt() != nil:
		node := stmt.GetCreateExtensionStmt()
		version := parseCreateExtensionVersion(sql)
		m.Extensions = append(m.Extensions, model.ExtensionDef{Name: node.GetExtname(), Version: version, SQL: sql})
		return nil
	case stmt.GetCreateStmt() != nil:
		rel := stmt.GetCreateStmt().GetRelation()
		schema, name := rangeVarName(rel, defaultSchema)
		m.Tables = append(m.Tables, model.TableDef{Schema: schema, Name: name, SQL: sql})
		return nil
	case stmt.GetIndexStmt() != nil:
		node := stmt.GetIndexStmt()
		schema, tableName := rangeVarName(node.GetRelation(), defaultSchema)
		indexSchema := schema
		if idxSchema, idxName := splitQualifiedIdentifier(node.GetIdxname(), ""); idxName != "" {
			indexSchema = idxSchema
		}
		_, indexName := splitQualifiedIdentifier(node.GetIdxname(), "")
		if indexName == "" {
			indexName = node.GetIdxname()
		}
		if indexSchema == "" {
			indexSchema = defaultSchema
		}
		m.Indexes = append(m.Indexes, model.IndexDef{
			Schema:      indexSchema,
			Name:        indexName,
			TableSchema: schema,
			TableName:   tableName,
			SQL:         sql,
		})
		return nil
	case stmt.GetViewStmt() != nil:
		node := stmt.GetViewStmt()
		schema, name := rangeVarName(node.GetView(), defaultSchema)
		m.Views = append(m.Views, model.ViewDef{Schema: schema, Name: name, SQL: sql})
		return nil
	case stmt.GetCreateFunctionStmt() != nil:
		node := stmt.GetCreateFunctionStmt()
		schema, name := nameList(node.GetFuncname(), defaultSchema)
		kind := "function"
		if node.GetIsProcedure() {
			kind = "procedure"
		}
		m.Routines = append(m.Routines, model.RoutineDef{
			Schema:       schema,
			Name:         name,
			IdentityArgs: functionIdentity(node),
			Kind:         kind,
			SQL:          sql,
		})
		return nil
	case stmt.GetCreateEnumStmt() != nil:
		node := stmt.GetCreateEnumStmt()
		schema, name := nameList(node.GetTypeName(), defaultSchema)
		values := make([]string, 0, len(node.GetVals()))
		for _, value := range node.GetVals() {
			values = append(values, stringValue(value))
		}
		m.Enums = append(m.Enums, model.EnumDef{Schema: schema, Name: name, Values: values, SQL: sql})
		return nil
	case stmt.GetCreateDomainStmt() != nil:
		node := stmt.GetCreateDomainStmt()
		schema, name := nameList(node.GetDomainname(), defaultSchema)
		m.Domains = append(m.Domains, model.DomainDef{Schema: schema, Name: name, SQL: sql})
		return nil
	case stmt.GetCreateSeqStmt() != nil:
		node := stmt.GetCreateSeqStmt()
		schema, name := rangeVarName(node.GetSequence(), defaultSchema)
		m.Sequences = append(m.Sequences, model.SequenceDef{Schema: schema, Name: name, SQL: sql})
		return nil
	case stmt.GetCommentStmt() != nil:
		comment, ok := parseComment(sql)
		if !ok {
			return fmt.Errorf("unsupported comment target in %q", sql)
		}
		m.Comments = append(m.Comments, comment)
		return nil
	default:
		return fmt.Errorf("unsupported statement kind")
	}
}

func parseCreateExtensionVersion(sql string) string {
	matches := extensionVersionPattern.FindStringSubmatch(sql)
	if len(matches) != 2 {
		return ""
	}
	return strings.Trim(matches[1], `"'`)
}

func rangeVarName(rel interface {
	GetSchemaname() string
	GetRelname() string
}, defaultSchema string) (string, string) {
	schema := rel.GetSchemaname()
	if schema == "" {
		schema = defaultSchema
	}
	return schema, rel.GetRelname()
}

func nameList(nodes []*pg_query.Node, defaultSchema string) (string, string) {
	parts := make([]string, 0, len(nodes))
	for _, node := range nodes {
		parts = append(parts, stringValue(node))
	}
	if len(parts) == 1 {
		return defaultSchema, parts[0]
	}
	if len(parts) >= 2 {
		return parts[len(parts)-2], parts[len(parts)-1]
	}
	return defaultSchema, ""
}

func functionIdentity(fn *pg_query.CreateFunctionStmt) string {
	args := make([]string, 0, len(fn.GetParameters()))
	for _, param := range fn.GetParameters() {
		if p := param.GetFunctionParameter(); p != nil {
			typeName := typeNameString(p.GetArgType())
			if typeName != "" {
				args = append(args, typeName)
			}
		}
	}
	return strings.Join(args, ", ")
}

func typeNameString(t *pg_query.TypeName) string {
	if t == nil {
		return ""
	}
	parts := make([]string, 0, len(t.GetNames()))
	for _, node := range t.GetNames() {
		parts = append(parts, stringValue(node))
	}
	return strings.Join(parts, ".")
}

func stringValue(node *pg_query.Node) string {
	switch {
	case node == nil:
		return ""
	case node.GetString_() != nil:
		return node.GetString_().GetSval()
	case node.GetAConst() != nil && node.GetAConst().GetSval() != nil:
		return node.GetAConst().GetSval().GetSval()
	default:
		return ""
	}
}

func splitQualifiedIdentifier(input string, defaultSchema string) (string, string) {
	input = strings.TrimSpace(input)
	input = strings.Trim(input, `"`)
	parts := strings.Split(input, ".")
	if len(parts) == 1 {
		return defaultSchema, parts[0]
	}
	return parts[len(parts)-2], parts[len(parts)-1]
}

func parseComment(sql string) (model.CommentDef, bool) {
	if matches := tableCommentPattern.FindStringSubmatch(sql); len(matches) == 3 {
		key := strings.Trim(matches[1], `"`)
		comment := strings.Trim(matches[2], "'")
		comment = strings.ReplaceAll(comment, "''", "'")
		return model.CommentDef{
			ObjectType: "table",
			ObjectKey:  key,
			Comment:    comment,
			SQL:        sql,
		}, true
	}
	if matches := columnCommentPattern.FindStringSubmatch(sql); len(matches) == 5 {
		key := strings.Trim(matches[1], `"`) + "." + strings.Trim(matches[2], `"`) + "." + strings.Trim(matches[3], `"`)
		comment := strings.Trim(matches[4], "'")
		comment = strings.ReplaceAll(comment, "''", "'")
		return model.CommentDef{
			ObjectType: "column",
			ObjectKey:  key,
			Comment:    comment,
			SQL:        sql,
		}, true
	}
	return model.CommentDef{}, false
}
