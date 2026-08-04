package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunVersionCommand(t *testing.T) {
	output := captureStdout(t, func() error {
		return run(context.Background(), []string{"version"})
	})

	if !strings.Contains(output, "pgpac dev") {
		t.Fatalf("expected version output, got %q", output)
	}
}

func TestRunVersionFlag(t *testing.T) {
	output := captureStdout(t, func() error {
		return run(context.Background(), []string{"--version"})
	})

	if !strings.Contains(output, "pgpac dev") {
		t.Fatalf("expected version output, got %q", output)
	}
}

func TestRunHelpMentionsVersion(t *testing.T) {
	output := captureStdout(t, func() error {
		return run(context.Background(), []string{"help"})
	})

	if !strings.Contains(output, "pgpac version") {
		t.Fatalf("expected help output to mention version command, got %q", output)
	}
}

func captureStdout(t *testing.T, fn func() error) string {
	t.Helper()

	var buffer bytes.Buffer
	oldStdout := stdout
	stdout = &buffer
	defer func() {
		stdout = oldStdout
	}()

	if err := fn(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return buffer.String()
}
