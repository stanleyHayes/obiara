package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/stanleyHayes/obiara/internal/quality/fieldtest"
)

func main() {
	manifest := flag.String("manifest", "", "field-test manifest JSON")
	candidate := flag.String("candidate-sha", "", "exact lowercase 40-character candidate SHA")
	atRaw := flag.String("at", "", "RFC3339 validation time")
	flag.Parse()
	at, err := time.Parse(time.RFC3339, *atRaw)
	if err != nil {
		exitf("invalid --at: %v", err)
	}
	if _, err := fieldtest.Load(*manifest, *candidate, at); err != nil {
		exitf("invalid field-test evidence: %v", err)
	}
}

func exitf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
