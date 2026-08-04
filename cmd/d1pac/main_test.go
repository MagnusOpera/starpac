package main

import (
	"bytes"
	"context"
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
}

func TestRunBuildCreatesPackage(t *testing.T) {
	outputDirectory := t.TempDir()
	output := captureStdout(t, func() error {
		return run(context.Background(), []string{
			"build",
			"--project", filepath.Join("..", "..", "testdata", "sample", "sample.d1pac"),
			"--output", outputDirectory,
		})
	})
	if !strings.Contains(output, filepath.Join(outputDirectory, "SampleProject.d1pkg")) {
		t.Fatalf("unexpected build output: %q", output)
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
