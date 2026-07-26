package telemetry

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestLocalRuntimeEmitsBoundedStructuredStartup(t *testing.T) {
	var output bytes.Buffer
	runtime, err := New(context.Background(), &output, Config{
		Version: "git-abc", Environment: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := runtime.Started(context.Background(), 2)
	runtime.Logger.InfoContext(ctx, "worker started", "jobs", 2)
	got := output.String()
	for _, expected := range []string{`"service":"obiara-worker"`, `"service_version":"git-abc"`, `"jobs":2`} {
		if !strings.Contains(got, expected) {
			t.Fatalf("log %q does not contain %q", got, expected)
		}
	}
}

func TestRuntimeRejectsCredentialsInEndpoint(t *testing.T) {
	_, err := New(context.Background(), &bytes.Buffer{}, Config{
		Endpoint: "https://user:secret@collector.example.test?token=bad",
	})
	if err == nil {
		t.Fatal("expected unsafe endpoint rejection")
	}
}
