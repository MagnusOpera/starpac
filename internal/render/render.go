package render

import (
	"strconv"
	"strings"

	"github.com/MagnusOpera/d1pac/internal/diff"
)

func Text(plan diff.Plan) string {
	var output strings.Builder
	status := "supported"
	if !plan.Summary.Supported {
		status = "blocked"
	}
	if plan.Summary.Destructive {
		status += " (contains destructive operations)"
	}
	output.WriteString("Plan status: " + status + "\n")
	output.WriteString("Operations: ")
	output.WriteString(strconv.Itoa(plan.Summary.OperationCount))
	output.WriteString("\n")
	for _, operation := range plan.Operations {
		output.WriteString("- ")
		output.WriteString(operation.Kind)
		output.WriteString(" [")
		output.WriteString(operation.Risk)
		output.WriteString("] ")
		output.WriteString(operation.ObjectKey)
		output.WriteString("\n")
	}
	return output.String()
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
