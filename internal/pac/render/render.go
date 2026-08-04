package render

import (
	"strconv"
	"strings"

	"github.com/MagnusOpera/starpac/internal/pac/plan"
)

func Text(deploymentPlan plan.Plan) string {
	var output strings.Builder
	status := "supported"
	if !deploymentPlan.Summary.Supported {
		status = "blocked"
	}
	if deploymentPlan.Summary.Destructive {
		status += " (contains destructive operations)"
	}
	output.WriteString("Plan status: " + status + "\n")
	output.WriteString("Operations: ")
	output.WriteString(strconv.Itoa(deploymentPlan.Summary.OperationCount))
	output.WriteString("\n")
	for _, operation := range deploymentPlan.Operations {
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

func SQL(deploymentPlan plan.Plan) string {
	var lines []string
	for _, operation := range deploymentPlan.Operations {
		lines = append(lines, "-- "+operation.Kind+" "+operation.ObjectKey)
		lines = append(lines, strings.TrimSpace(operation.SQL))
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}
