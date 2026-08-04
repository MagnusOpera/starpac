package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/MagnusOpera/starpac/internal/pac/plan"
	"github.com/MagnusOpera/starpac/internal/pac/render"
)

func ResolvePackageOutput(outputPath, packageID, extension string, caseInsensitive bool) (string, error) {
	outputPath = filepath.Clean(outputPath)
	pathForMatch := outputPath
	extensionForMatch := extension
	if caseInsensitive {
		pathForMatch = strings.ToLower(pathForMatch)
		extensionForMatch = strings.ToLower(extensionForMatch)
	}
	if strings.HasSuffix(pathForMatch, extensionForMatch) {
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return "", err
		}
		return outputPath, nil
	}
	if err := os.MkdirAll(outputPath, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(outputPath, packageID+extension), nil
}

func WritePlan(output io.Writer, format, scriptPath string, deploymentPlan plan.Plan) error {
	if scriptPath != "" {
		if err := os.WriteFile(scriptPath, []byte(render.SQL(deploymentPlan)), 0o644); err != nil {
			return err
		}
	}
	switch strings.ToLower(format) {
	case "json":
		encoder := json.NewEncoder(output)
		encoder.SetIndent("", "  ")
		return encoder.Encode(deploymentPlan)
	case "text":
		_, err := fmt.Fprint(output, render.Text(deploymentPlan))
		return err
	default:
		return fmt.Errorf("unsupported format %q", format)
	}
}

func WriteVersion(output io.Writer, product, version, commit, buildDate string) {
	fmt.Fprintf(output, "%s %s\n", product, version)
	if commit != "" && commit != "unknown" {
		fmt.Fprintf(output, "commit: %s\n", commit)
	}
	if buildDate != "" && buildDate != "unknown" {
		fmt.Fprintf(output, "built: %s\n", buildDate)
	}
}
