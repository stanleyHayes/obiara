package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/stanleyHayes/obiara/internal/quality/residencydecision"
)

func main() {
	at := flag.String("at", "", "RFC3339 evaluation time (required)")
	flag.Parse()
	if *at == "" || flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: residencydecision --at RFC3339 decision.json")
		os.Exit(1)
	}
	now, err := time.Parse(time.RFC3339, *at)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid evaluation time")
		os.Exit(1)
	}
	_, decision, err := residencydecision.Load(flag.Arg(0), now)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoded, _ := json.MarshalIndent(decision, "", "  ")
	fmt.Println(string(encoded))
	if !decision.Eligible {
		os.Exit(2)
	}
}
