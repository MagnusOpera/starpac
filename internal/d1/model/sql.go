package model

import "strings"

func ColumnDefinitions(createTableSQL string) map[string]string {
	definitions := map[string]string{}
	open := strings.Index(createTableSQL, "(")
	if open == -1 {
		return definitions
	}
	close := findMatchingParenthesis(createTableSQL, open)
	if close == -1 {
		return definitions
	}
	for _, fragment := range splitTopLevel(createTableSQL[open+1 : close]) {
		name, _, ok := cutIdentifier(fragment)
		if !ok || isTableConstraint(name) {
			continue
		}
		definitions[normalizeIdentifier(name)] = strings.TrimSpace(fragment)
	}
	return definitions
}

func TableConstraints(createTableSQL string) []string {
	open := strings.Index(createTableSQL, "(")
	if open == -1 {
		return nil
	}
	close := findMatchingParenthesis(createTableSQL, open)
	if close == -1 {
		return nil
	}
	var constraints []string
	for _, fragment := range splitTopLevel(createTableSQL[open+1 : close]) {
		name, _, ok := cutIdentifier(fragment)
		if ok && isTableConstraint(name) {
			constraints = append(constraints, NormalizeDDL(fragment))
		}
	}
	return constraints
}

func TableOptions(createTableSQL string) string {
	open := strings.Index(createTableSQL, "(")
	if open == -1 {
		return ""
	}
	close := findMatchingParenthesis(createTableSQL, open)
	if close == -1 {
		return ""
	}
	return NormalizeDDL(createTableSQL[close+1:])
}

func ReplaceCreateTableName(createTableSQL, replacement string) (string, bool) {
	upper := strings.ToUpper(createTableSQL)
	position := strings.Index(upper, "CREATE TABLE")
	if position == -1 {
		return "", false
	}
	restStart := position + len("CREATE TABLE")
	rest := strings.TrimSpace(createTableSQL[restStart:])
	if strings.HasPrefix(strings.ToUpper(rest), "IF NOT EXISTS") {
		rest = strings.TrimSpace(rest[len("IF NOT EXISTS"):])
	}
	_, afterName, ok := cutIdentifier(rest)
	if !ok {
		return "", false
	}
	return "CREATE TABLE " + QuoteIdentifier(replacement) + " " + afterName, true
}

func QuoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func cutIdentifier(input string) (string, string, bool) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", false
	}
	if input[0] == '"' || input[0] == '`' || input[0] == '[' {
		terminator := input[0]
		if terminator == '[' {
			terminator = ']'
		}
		for index := 1; index < len(input); index++ {
			if input[index] != terminator {
				continue
			}
			if terminator != ']' && index+1 < len(input) && input[index+1] == terminator {
				index++
				continue
			}
			return input[:index+1], strings.TrimSpace(input[index+1:]), true
		}
		return "", "", false
	}
	for index, character := range input {
		if character == ' ' || character == '\t' || character == '\n' || character == '(' {
			return input[:index], strings.TrimSpace(input[index:]), true
		}
	}
	return input, "", true
}

func normalizeIdentifier(identifier string) string {
	identifier = strings.TrimSpace(identifier)
	if len(identifier) >= 2 {
		switch {
		case identifier[0] == '"' && identifier[len(identifier)-1] == '"':
			identifier = strings.ReplaceAll(identifier[1:len(identifier)-1], `""`, `"`)
		case identifier[0] == '`' && identifier[len(identifier)-1] == '`':
			identifier = strings.ReplaceAll(identifier[1:len(identifier)-1], "``", "`")
		case identifier[0] == '[' && identifier[len(identifier)-1] == ']':
			identifier = identifier[1 : len(identifier)-1]
		}
	}
	return identifier
}

func isTableConstraint(firstToken string) bool {
	switch strings.ToUpper(normalizeIdentifier(firstToken)) {
	case "CONSTRAINT", "PRIMARY", "UNIQUE", "CHECK", "FOREIGN":
		return true
	default:
		return false
	}
}

func findMatchingParenthesis(input string, open int) int {
	depth := 0
	inSingleQuote := false
	inDoubleQuote := false
	for index := open; index < len(input); index++ {
		switch input[index] {
		case '\'':
			if !inDoubleQuote {
				if inSingleQuote && index+1 < len(input) && input[index+1] == '\'' {
					index++
					continue
				}
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
		case '(':
			if !inSingleQuote && !inDoubleQuote {
				depth++
			}
		case ')':
			if !inSingleQuote && !inDoubleQuote {
				depth--
				if depth == 0 {
					return index
				}
			}
		}
	}
	return -1
}

func splitTopLevel(input string) []string {
	var parts []string
	start := 0
	depth := 0
	inSingleQuote := false
	inDoubleQuote := false
	for index := 0; index < len(input); index++ {
		switch input[index] {
		case '\'':
			if !inDoubleQuote {
				if inSingleQuote && index+1 < len(input) && input[index+1] == '\'' {
					index++
					continue
				}
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
		case '(':
			if !inSingleQuote && !inDoubleQuote {
				depth++
			}
		case ')':
			if !inSingleQuote && !inDoubleQuote && depth > 0 {
				depth--
			}
		case ',':
			if !inSingleQuote && !inDoubleQuote && depth == 0 {
				parts = append(parts, strings.TrimSpace(input[start:index]))
				start = index + 1
			}
		}
	}
	parts = append(parts, strings.TrimSpace(input[start:]))
	return parts
}
