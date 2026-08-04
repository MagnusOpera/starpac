package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	sharedcli "github.com/MagnusOpera/starpac/internal/pac/cli"
	"github.com/MagnusOpera/starpac/internal/postgres/apply"
	"github.com/MagnusOpera/starpac/internal/postgres/diff"
	"github.com/MagnusOpera/starpac/internal/postgres/introspect"
	"github.com/MagnusOpera/starpac/internal/postgres/packagefmt"
	"github.com/MagnusOpera/starpac/internal/postgres/parser"
	"github.com/MagnusOpera/starpac/internal/postgres/project"
)

var (
	version             = "dev"
	commit              = "unknown"
	buildDate           = "unknown"
	stdout    io.Writer = os.Stdout
	stderr    io.Writer = os.Stderr
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		printUsage()
		return errors.New("missing command")
	}

	switch args[0] {
	case "build":
		return runBuild(ctx, args[1:])
	case "plan":
		return runPlan(ctx, args[1:])
	case "apply":
		return runApply(ctx, args[1:])
	case "version", "--version":
		printVersion()
		return nil
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runBuild(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	projectPath := fs.String("project", "", "Path to a .pgpac project file")
	outputPath := fs.String("output", "", "Output file or directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *projectPath == "" || *outputPath == "" {
		return errors.New("build requires --project and --output")
	}

	project, rawXML, err := projectxml.Load(*projectPath)
	if err != nil {
		return err
	}

	desired, err := parser.BuildDesiredModel(project)
	if err != nil {
		return err
	}

	manifest := packagefmt.NewManifest(project, desired)
	targetOutput, err := resolvePackageOutput(*outputPath, project.PackageID)
	if err != nil {
		return err
	}

	if err := packagefmt.Write(targetOutput, manifest, desired, rawXML, project); err != nil {
		return err
	}

	fmt.Fprintln(stdout, targetOutput)
	_ = ctx
	return nil
}

func runPlan(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("plan", flag.ContinueOnError)
	packagePath := fs.String("package", "", "Path to a .pgpkg package")
	connectionString := fs.String("connection", "", "PostgreSQL connection string")
	format := fs.String("format", "text", "Output format: text or json")
	scriptPath := fs.String("script", "", "Optional path to write the SQL preview")
	allowDrop := fs.Bool("allow-drop", false, "Allow destructive operations to be rendered as executable SQL")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *packagePath == "" || *connectionString == "" {
		return errors.New("plan requires --package and --connection")
	}

	pkg, err := packagefmt.Read(*packagePath)
	if err != nil {
		return err
	}

	actual, err := introspect.LoadActualModel(
		ctx,
		*connectionString,
		pkg.Project.Target.OwnedSchemaNames(),
		pkg.Project.Target.ExtensionNames(),
		pkg.Project.PostgresVersion,
	)
	if err != nil {
		return err
	}

	plan := diff.BuildPlan(pkg.Project, pkg.Model, actual, diff.Options{AllowDrop: *allowDrop})
	return sharedcli.WritePlan(os.Stdout, *format, *scriptPath, plan)
}

func runApply(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	packagePath := fs.String("package", "", "Path to a .pgpkg package")
	connectionString := fs.String("connection", "", "PostgreSQL connection string")
	allowDrop := fs.Bool("allow-drop", false, "Allow destructive operations")
	force := fs.Bool("force", false, "Bypass destructive operation protection")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *packagePath == "" || *connectionString == "" {
		return errors.New("apply requires --package and --connection")
	}

	pkg, err := packagefmt.Read(*packagePath)
	if err != nil {
		return err
	}

	actual, err := introspect.LoadActualModel(
		ctx,
		*connectionString,
		pkg.Project.Target.OwnedSchemaNames(),
		pkg.Project.Target.ExtensionNames(),
		pkg.Project.PostgresVersion,
	)
	if err != nil {
		return err
	}

	plan := diff.BuildPlan(pkg.Project, pkg.Model, actual, diff.Options{AllowDrop: *allowDrop || *force})
	if err := apply.Execute(ctx, *connectionString, pkg.Project, plan, apply.Options{AllowDrop: *allowDrop, Force: *force}); err != nil {
		return err
	}

	fmt.Fprintln(stdout, "Applied package.")
	return nil
}

func resolvePackageOutput(outputPath, packageID string) (string, error) {
	return sharedcli.ResolvePackageOutput(outputPath, packageID, ".pgpkg", false)
}

func printUsage() {
	fmt.Fprintln(stdout, "pgpac build --project <file.pgpac> --output <dir-or-file>")
	fmt.Fprintln(stdout, "pgpac plan --package <file.pgpkg> --connection <postgres-uri> [--format text|json] [--script <file>] [--allow-drop]")
	fmt.Fprintln(stdout, "pgpac apply --package <file.pgpkg> --connection <postgres-uri> [--allow-drop] [--force]")
	fmt.Fprintln(stdout, "pgpac version")
	fmt.Fprintln(stdout, "pgpac --version")
}

func printVersion() {
	sharedcli.WriteVersion(stdout, "pgpac", version, commit, buildDate)
}
