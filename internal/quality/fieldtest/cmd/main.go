package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/stanleyHayes/obiara/internal/quality/fieldtest"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

type decision struct {
	SchemaVersion string   `json:"schemaVersion"`
	Valid         bool     `json:"valid"`
	Disposition   string   `json:"disposition"`
	Qualified     bool     `json:"qualified"`
	CandidateSHA  string   `json:"candidateSha,omitempty"`
	Blockers      []string `json:"blockers"`
	Error         string   `json:"error,omitempty"`
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("fieldtest", flag.ContinueOnError)
	flags.SetOutput(stderr)
	manifestPath := flags.String("manifest", "", "field-test manifest JSON")
	candidate := flags.String("candidate-sha", "", "exact lowercase 40-character candidate SHA")
	atRaw := flags.String("at", "", "RFC3339 validation time")
	if err := flags.Parse(args); err != nil {
		return 64
	}
	if *manifestPath == "" || *candidate == "" || *atRaw == "" || flags.NArg() != 0 {
		_, _ = fmt.Fprintln(stderr, "--manifest, --candidate-sha and --at are required; positional arguments are not accepted")
		return 64
	}
	at, err := time.Parse(time.RFC3339, *atRaw)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "invalid --at: %v\n", err)
		return 64
	}
	manifest, err := fieldtest.Load(*manifestPath, *candidate, at)
	if err != nil {
		writeDecision(stdout, decision{
			SchemaVersion: "obiara.field-test-decision.v1",
			Valid:         false, Disposition: "invalid", Qualified: false,
			Blockers: []string{}, Error: err.Error(),
		})
		return 1
	}
	result := decision{
		SchemaVersion: "obiara.field-test-decision.v1",
		Valid:         true, Disposition: manifest.Disposition,
		Qualified:    manifest.Disposition == "qualified-field-evidence",
		CandidateSHA: manifest.CandidateSHA,
		Blockers:     append([]string(nil), manifest.Blockers...),
	}
	writeDecision(stdout, result)
	if manifest.Disposition == "blocked" {
		return 2
	}
	if result.Qualified {
		return 0
	}
	return 1
}

func writeDecision(output io.Writer, result decision) {
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(result)
}
