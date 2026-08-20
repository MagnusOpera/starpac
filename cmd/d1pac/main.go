package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/MagnusOpera/starpac/internal/d1/apply"
	"github.com/MagnusOpera/starpac/internal/d1/cloudflare"
	"github.com/MagnusOpera/starpac/internal/d1/compiler"
	"github.com/MagnusOpera/starpac/internal/d1/diff"
	"github.com/MagnusOpera/starpac/internal/d1/introspect"
	"github.com/MagnusOpera/starpac/internal/d1/packagefmt"
	"github.com/MagnusOpera/starpac/internal/d1/project"
	sharedcli "github.com/MagnusOpera/starpac/internal/pac/cli"
)

var (
	version             = "dev"
	commit              = "unknown"
	buildDate           = "unknown"
	stdout    io.Writer = os.Stdout
	stderr    io.Writer = os.Stderr
)

type targetFlags struct {
	accountID  string
	database   string
	apiToken   string
	apiBaseURL string
}

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, arguments []string) error {
	if len(arguments) == 0 {
		printUsage()
		return errors.New("missing command")
	}
	switch arguments[0] {
	case "build":
		return runBuild(ctx, arguments[1:])
	case "plan":
		return runPlan(ctx, arguments[1:])
	case "apply":
		return runApply(ctx, arguments[1:])
	case "version", "--version":
		printVersion()
		return nil
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command %q", arguments[0])
	}
}

func runBuild(ctx context.Context, arguments []string) error {
	flags := flag.NewFlagSet("build", flag.ContinueOnError)
	projectPath := flags.String("project", "", "Path to a .d1pac project file")
	outputPath := flags.String("output", "", "Output file or directory")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if *projectPath == "" || *outputPath == "" {
		return errors.New("build requires --project and --output")
	}
	project, rawProject, err := projectxml.Load(*projectPath)
	if err != nil {
		return err
	}
	desired, err := compiler.BuildDesiredModel(ctx, project)
	if err != nil {
		return err
	}
	manifest := packagefmt.NewManifest(project, desired)
	targetOutput, err := resolvePackageOutput(*outputPath, project.PackageID)
	if err != nil {
		return err
	}
	if err := packagefmt.Write(targetOutput, manifest, desired, rawProject, project); err != nil {
		return err
	}
	fmt.Fprintln(stdout, targetOutput)
	return nil
}

func runPlan(ctx context.Context, arguments []string) error {
	flags := flag.NewFlagSet("plan", flag.ContinueOnError)
	packagePath := flags.String("package", "", "Path to a .d1pkg package")
	format := flags.String("format", "text", "Output format: text or json")
	scriptPath := flags.String("script", "", "Optional path to write the SQL preview")
	allowDrop := flags.Bool("allow-drop", false, "Allow destructive operations to be rendered as executable SQL")
	strict := flags.Bool("strict", false, "Require declared table column order to match the live schema")
	target := addTargetFlags(flags)
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if *packagePath == "" {
		return errors.New("plan requires --package")
	}
	pkg, err := packagefmt.Read(*packagePath)
	if err != nil {
		return err
	}
	client, err := newTargetClient(target)
	if err != nil {
		return err
	}
	actual, err := introspect.LoadRemote(ctx, client, pkg.Project.IsIgnored)
	if err != nil {
		return err
	}
	plan := diff.BuildPlan(pkg.Project, pkg.Model, actual, diff.Options{
		AllowDrop: *allowDrop,
		Strict:    *strict,
	})
	return sharedcli.WritePlan(stdout, *format, *scriptPath, plan)
}

func runApply(ctx context.Context, arguments []string) error {
	flags := flag.NewFlagSet("apply", flag.ContinueOnError)
	packagePath := flags.String("package", "", "Path to a .d1pkg package")
	allowDrop := flags.Bool("allow-drop", false, "Allow destructive operations")
	force := flags.Bool("force", false, "Bypass destructive-operation protection")
	strict := flags.Bool("strict", false, "Require declared table column order to match the live schema")
	target := addTargetFlags(flags)
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if *packagePath == "" {
		return errors.New("apply requires --package")
	}
	pkg, err := packagefmt.Read(*packagePath)
	if err != nil {
		return err
	}
	client, err := newTargetClient(target)
	if err != nil {
		return err
	}
	actual, err := introspect.LoadRemote(ctx, client, pkg.Project.IsIgnored)
	if err != nil {
		return err
	}
	plan := diff.BuildPlan(pkg.Project, pkg.Model, actual, diff.Options{
		AllowDrop: *allowDrop || *force,
		Strict:    *strict,
	})
	if err := apply.Execute(ctx, client, pkg.Project, plan, apply.Options{
		AllowDrop: *allowDrop,
		Force:     *force,
	}); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "Applied package.")
	return nil
}

func addTargetFlags(flags *flag.FlagSet) *targetFlags {
	target := &targetFlags{}
	flags.StringVar(&target.accountID, "account-id", os.Getenv("CLOUDFLARE_ACCOUNT_ID"), "Cloudflare account id; defaults to CLOUDFLARE_ACCOUNT_ID")
	flags.StringVar(&target.database, "database", "", "D1 database name or UUID")
	flags.StringVar(&target.apiToken, "api-token", os.Getenv("CLOUDFLARE_API_TOKEN"), "Cloudflare API token; defaults to CLOUDFLARE_API_TOKEN")
	flags.StringVar(&target.apiBaseURL, "api-base-url", "", "Cloudflare API base URL")
	return target
}

func newTargetClient(target *targetFlags) (*cloudflare.Client, error) {
	return cloudflare.New(cloudflare.Config{
		AccountID: target.accountID,
		Database:  target.database,
		APIToken:  target.apiToken,
		BaseURL:   target.apiBaseURL,
	})
}

func resolvePackageOutput(outputPath, packageID string) (string, error) {
	return sharedcli.ResolvePackageOutput(outputPath, packageID, ".d1pkg", true)
}

func printUsage() {
	fmt.Fprintln(stdout, "d1pac build --project <file.d1pac> --output <dir-or-file>")
	fmt.Fprintln(stdout, "d1pac plan --package <file.d1pkg> --account-id <id> --database <name-or-id> [--format text|json] [--script <file>] [--allow-drop] [--strict]")
	fmt.Fprintln(stdout, "d1pac apply --package <file.d1pkg> --account-id <id> --database <name-or-id> [--allow-drop] [--force] [--strict]")
	fmt.Fprintln(stdout, "d1pac version")
	fmt.Fprintln(stdout, "d1pac --version")
}

func printVersion() {
	sharedcli.WriteVersion(stdout, "d1pac", version, commit, buildDate)
}
