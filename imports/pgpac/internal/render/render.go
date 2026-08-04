package render

import (
	"strconv"
	"strings"

	"github.com/MagnusOpera/pgpac/internal/diff"
)

func Text(plan diff.Plan) string {
	var b strings.Builder
	status := "supported"
	if plan.Summary.Destructive {
		status += " (contains destructive operations)"
	}
	b.WriteString("Plan status: " + status + "\n")
	b.WriteString("Operations: ")
	b.WriteString(strconv.Itoa(plan.Summary.OperationCount))
	b.WriteString("\n")
	for _, operation := range plan.Operations {
		b.WriteString("- ")
		b.WriteString(operation.Kind)
		b.WriteString(" [")
		b.WriteString(operation.Risk)
		b.WriteString("] ")
		b.WriteString(operation.ObjectKey)
		b.WriteString("\n")
	}
	return b.String()
}

func SQL(plan diff.Plan) string {
	var lines []string
	for _, operation := range plan.Operations {
		lines = append(lines, "-- "+operation.Kind+" "+operation.ObjectKey)
		lines = append(lines, strings.TrimSpace(operation.SQL))
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}
