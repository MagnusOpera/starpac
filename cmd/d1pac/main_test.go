package main

import (
	"bytes"
	"context"
	"flag"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunVersionCommand(t *testing.T) {
	output := captureStdout(t, func() error {
		return run(context.Background(), []string{"version"})
	})
	if !strings.Contains(output, "d1pac dev") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestRunHelpMentionsCommands(t *testing.T) {
	output := captureStdout(t, func() error {
		return run(context.Background(), []string{"help"})
	})
	for _, command := range []string{"d1pac build", "d1pac plan", "d1pac apply", "d1pac version"} {
		if !strings.Contains(output, command) {
			t.Fatalf("help does not mention %q: %s", command, output)
		}
	}
	if strings.Count(output, "[--strict]") != 2 {
		t.Fatalf("help must mention strict mode for plan and apply: %s", output)
	}
}

func TestRunBuildCreatesPackage(t *testing.T) {
	outputDirectory := t.TempDir()
	output := captureStdout(t, func() error {
		return run(context.Background(), []string{
			"build",
			"--project", filepath.Join("..", "..", "products", "d1pac", "testdata", "sample", "sample.d1pac"),
			"--output", outputDirectory,
		})
	})
	if !strings.Contains(output, filepath.Join(outputDirectory, "SampleProject.d1pkg")) {
		t.Fatalf("unexpected build output: %q", output)
	}
}

func TestTargetFlagsRetainExplicitValues(t *testing.T) {
	flags := flag.NewFlagSet("target", flag.ContinueOnError)
	target := addTargetFlags(flags)
	if err := flags.Parse([]string{
		"--account-id", "account-id",
		"--database", "database-id",
		"--api-token", "api-token",
		"--api-base-url", "https://api.example.test",
	}); err != nil {
		t.Fatalf("parse target flags: %v", err)
	}

	if target.accountID != "account-id" {
		t.Fatalf("unexpected account id: %q", target.accountID)
	}
	if target.database != "database-id" {
		t.Fatalf("unexpected database: %q", target.database)
	}
	if target.apiToken != "api-token" {
		t.Fatalf("unexpected API token: %q", target.apiToken)
	}
	if target.apiBaseURL != "https://api.example.test" {
		t.Fatalf("unexpected API base URL: %q", target.apiBaseURL)
	}
}

func captureStdout(t *testing.T, callback func() error) string {
	t.Helper()
	var buffer bytes.Buffer
	previous := stdout
	stdout = &buffer
	defer func() {
		stdout = previous
	}()
	if err := callback(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return buffer.String()
}
