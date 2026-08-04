package model

import (
	"sort"
	"strings"
)

type SchemaModel struct {
	SQLiteVersion string       `json:"sqliteVersion,omitempty"`
	Tables        []TableDef   `json:"tables,omitempty"`
	Indexes       []IndexDef   `json:"indexes,omitempty"`
	Views         []ViewDef    `json:"views,omitempty"`
	Triggers      []TriggerDef `json:"triggers,omitempty"`
}

type TableDef struct {
	Name        string          `json:"name"`
	SQL         string          `json:"sql"`
	Columns     []ColumnDef     `json:"columns,omitempty"`
	ForeignKeys []ForeignKeyDef `json:"foreignKeys,omitempty"`
}

type ColumnDef struct {
	Position   int     `json:"position"`
	Name       string  `json:"name"`
	Type       string  `json:"type,omitempty"`
	NotNull    bool    `json:"notNull"`
	DefaultSQL *string `json:"defaultSql,omitempty"`
	PrimaryKey int     `json:"primaryKey"`
	Hidden     int     `json:"hidden,omitempty"`
	Definition string  `json:"definition,omitempty"`
}

type ForeignKeyDef struct {
	ID       int    `json:"id"`
	Sequence int    `json:"sequence"`
	Table    string `json:"table"`
	From     string `json:"from"`
	To       string `json:"to"`
	OnUpdate string `json:"onUpdate"`
	OnDelete string `json:"onDelete"`
	Match    string `json:"match"`
}

type IndexDef struct {
	Name      string `json:"name"`
	TableName string `json:"tableName"`
	SQL       string `json:"sql"`
}

type ViewDef struct {
	Name string `json:"name"`
	SQL  string `json:"sql"`
}

type TriggerDef struct {
	Name      string `json:"name"`
	TableName string `json:"tableName"`
	SQL       string `json:"sql"`
}

func CanonicalSQL(sql string) string {
	sql = strings.TrimSpace(sql)
	sql = strings.TrimSuffix(sql, ";")
	return strings.Join(strings.Fields(sql), " ")
}

func NormalizeDDL(sql string) string {
	sql = CanonicalSQL(sql)
	var result strings.Builder
	var token strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	flushToken := func() {
		if token.Len() == 0 {
			return
		}
		result.WriteString(strings.ToLower(token.String()))
		token.Reset()
	}

	for index := 0; index < len(sql); index++ {
		character := sql[index]
		switch {
		case inSingleQuote:
			result.WriteByte(character)
			if character == '\'' {
				if index+1 < len(sql) && sql[index+1] == '\'' {
					index++
					result.WriteByte(sql[index])
					continue
				}
				inSingleQuote = false
			}
		case inDoubleQuote:
			if character == '"' {
				if index+1 < len(sql) && sql[index+1] == '"' {
					result.WriteByte('"')
					index++
					continue
				}
				inDoubleQuote = false
				continue
			}
			result.WriteByte(character)
		case character == '\'':
			flushToken()
			inSingleQuote = true
			result.WriteByte(character)
		case character == '"':
			flushToken()
			inDoubleQuote = true
		case isIdentifierCharacter(character):
			token.WriteByte(character)
		case character == ' ' || character == '\t' || character == '\n' || character == '\r':
			flushToken()
		default:
			flushToken()
			result.WriteByte(character)
		}
	}
	flushToken()
	return result.String()
}

func Sort(schema *SchemaModel) {
	sort.Slice(schema.Tables, func(left, right int) bool {
		return schema.Tables[left].Name < schema.Tables[right].Name
	})
	sort.Slice(schema.Indexes, func(left, right int) bool {
		return schema.Indexes[left].Name < schema.Indexes[right].Name
	})
	sort.Slice(schema.Views, func(left, right int) bool {
		return schema.Views[left].Name < schema.Views[right].Name
	})
	sort.Slice(schema.Triggers, func(left, right int) bool {
		return schema.Triggers[left].Name < schema.Triggers[right].Name
	})
	for tableIndex := range schema.Tables {
		table := &schema.Tables[tableIndex]
		sort.Slice(table.Columns, func(left, right int) bool {
			return table.Columns[left].Position < table.Columns[right].Position
		})
		sort.Slice(table.ForeignKeys, func(left, right int) bool {
			if table.ForeignKeys[left].ID == table.ForeignKeys[right].ID {
				return table.ForeignKeys[left].Sequence < table.ForeignKeys[right].Sequence
			}
			return table.ForeignKeys[left].ID < table.ForeignKeys[right].ID
		})
	}
}

func isIdentifierCharacter(character byte) bool {
	return character >= 'a' && character <= 'z' ||
		character >= 'A' && character <= 'Z' ||
		character >= '0' && character <= '9' ||
		character == '_'
}
