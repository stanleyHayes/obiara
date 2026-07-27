package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/stanleyHayes/obiara/internal/quality/stagingqualification"
)

func main() {
	mode := flag.String("mode", "validate", "generate or validate")
	candidate := flag.String("candidate-sha", "", "exact lowercase 40-character candidate SHA")
	atRaw := flag.String("at", "", "RFC3339 generation or validation time")
	output := flag.String("output", "", "generated qualification output path")
	qualification := flag.String("qualification", "", "qualification path to validate")
	releaseEvidence := flag.String("release-evidence", "", "release evidence JSON path")
	releaseBundle := flag.String("release-bundle", "", "release bundle JSON path")
	drEvidence := flag.String("dr-evidence", "", "DR evidence JSON path")
	flag.Parse()

	at, err := time.Parse(time.RFC3339, *atRaw)
	if err != nil {
		exitf("invalid --at: %v", err)
	}
	sources := stagingqualification.Sources{
		ReleaseEvidencePath: *releaseEvidence,
		ReleaseBundlePath:   *releaseBundle,
		DRRehearsalPath:     *drEvidence,
	}
	switch *mode {
	case "generate":
		if *output == "" {
			exitf("--output is required")
		}
		result, err := stagingqualification.Generate(*candidate, at, sources)
		if err != nil {
			exitf("generate: %v", err)
		}
		raw, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			exitf("encode: %v", err)
		}
		raw = append(raw, '\n')
		if err := os.WriteFile(*output, raw, 0o644); err != nil {
			exitf("write: %v", err)
		}
	case "validate":
		if *qualification == "" {
			exitf("--qualification is required")
		}
		if _, err := stagingqualification.Load(*qualification, *candidate, at, sources); err != nil {
			exitf("validate: %v", err)
		}
	default:
		exitf("unsupported --mode")
	}
}

func exitf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
