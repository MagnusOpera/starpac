package model

import (
	"sort"
	"strings"
)

type SchemaModel struct {
	PostgresVersion int            `json:"postgresVersion"`
	Schemas         []SchemaDef    `json:"schemas,omitempty"`
	Extensions      []ExtensionDef `json:"extensions,omitempty"`
	Tables          []TableDef     `json:"tables,omitempty"`
	Indexes         []IndexDef     `json:"indexes,omitempty"`
	Views           []ViewDef      `json:"views,omitempty"`
	Routines        []RoutineDef   `json:"routines,omitempty"`
	Enums           []EnumDef      `json:"enums,omitempty"`
	Domains         []DomainDef    `json:"domains,omitempty"`
	Sequences       []SequenceDef  `json:"sequences,omitempty"`
	Comments        []CommentDef   `json:"comments,omitempty"`
}

type SchemaDef struct {
	Name string `json:"name"`
	SQL  string `json:"sql,omitempty"`
}

type ExtensionDef struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	SQL     string `json:"sql,omitempty"`
}

type TableDef struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
	SQL    string `json:"sql"`
}

type IndexDef struct {
	Schema      string `json:"schema"`
	Name        string `json:"name"`
	TableSchema string `json:"tableSchema,omitempty"`
	TableName   string `json:"tableName,omitempty"`
	SQL         string `json:"sql"`
}

type ViewDef struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
	SQL    string `json:"sql"`
}

type RoutineDef struct {
	Schema       string `json:"schema"`
	Name         string `json:"name"`
	IdentityArgs string `json:"identityArgs"`
	Kind         string `json:"kind"`
	SQL          string `json:"sql"`
}

type EnumDef struct {
	Schema string   `json:"schema"`
	Name   string   `json:"name"`
	Values []string `json:"values,omitempty"`
	SQL    string   `json:"sql"`
}

type DomainDef struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
	SQL    string `json:"sql"`
}

type SequenceDef struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
	SQL    string `json:"sql"`
}

type CommentDef struct {
	ObjectType string `json:"objectType"`
	ObjectKey  string `json:"objectKey"`
	Comment    string `json:"comment"`
	SQL        string `json:"sql"`
}

func CanonicalSQL(sql string) string {
	sql = strings.TrimSpace(sql)
	sql = strings.TrimSuffix(sql, ";")
	sql = strings.Join(strings.Fields(sql), " ")
	return strings.TrimSpace(sql)
}

func QualifiedName(schema, name string) string {
	if schema == "" {
		return name
	}
	return schema + "." + name
}

func RoutineKey(schema, name, identityArgs string) string {
	return QualifiedName(schema, name) + "(" + strings.TrimSpace(identityArgs) + ")"
}

func Sort(m *SchemaModel) {
	sort.Slice(m.Schemas, func(i, j int) bool { return m.Schemas[i].Name < m.Schemas[j].Name })
	sort.Slice(m.Extensions, func(i, j int) bool { return m.Extensions[i].Name < m.Extensions[j].Name })
	sort.Slice(m.Tables, func(i, j int) bool {
		return QualifiedName(m.Tables[i].Schema, m.Tables[i].Name) < QualifiedName(m.Tables[j].Schema, m.Tables[j].Name)
	})
	sort.Slice(m.Indexes, func(i, j int) bool {
		return QualifiedName(m.Indexes[i].Schema, m.Indexes[i].Name) < QualifiedName(m.Indexes[j].Schema, m.Indexes[j].Name)
	})
	sort.Slice(m.Views, func(i, j int) bool {
		return QualifiedName(m.Views[i].Schema, m.Views[i].Name) < QualifiedName(m.Views[j].Schema, m.Views[j].Name)
	})
	sort.Slice(m.Routines, func(i, j int) bool {
		return RoutineKey(m.Routines[i].Schema, m.Routines[i].Name, m.Routines[i].IdentityArgs) < RoutineKey(m.Routines[j].Schema, m.Routines[j].Name, m.Routines[j].IdentityArgs)
	})
	sort.Slice(m.Enums, func(i, j int) bool {
		return QualifiedName(m.Enums[i].Schema, m.Enums[i].Name) < QualifiedName(m.Enums[j].Schema, m.Enums[j].Name)
	})
	sort.Slice(m.Domains, func(i, j int) bool {
		return QualifiedName(m.Domains[i].Schema, m.Domains[i].Name) < QualifiedName(m.Domains[j].Schema, m.Domains[j].Name)
	})
	sort.Slice(m.Sequences, func(i, j int) bool {
		return QualifiedName(m.Sequences[i].Schema, m.Sequences[i].Name) < QualifiedName(m.Sequences[j].Schema, m.Sequences[j].Name)
	})
	sort.Slice(m.Comments, func(i, j int) bool {
		if m.Comments[i].ObjectType == m.Comments[j].ObjectType {
			return m.Comments[i].ObjectKey < m.Comments[j].ObjectKey
		}
		return m.Comments[i].ObjectType < m.Comments[j].ObjectType
	})
}
