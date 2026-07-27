package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/stanleyHayes/obiara/internal/quality/launchgates"
	"os"
	"time"
)

func main() {
	at := flag.String("at", "", "RFC3339 evaluation time (required)")
	flag.Parse()
	if flag.NArg() != 1 || *at == "" {
		fmt.Fprintln(os.Stderr, "usage: launchgates --at RFC3339 registry.json")
		os.Exit(1)
	}
	now, e := time.Parse(time.RFC3339, *at)
	if e != nil {
		fmt.Fprintln(os.Stderr, "invalid evaluation time")
		os.Exit(1)
	}
	_, decision, e := launchgates.Load(flag.Arg(0), now)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	encoded, _ := json.MarshalIndent(decision, "", "  ")
	fmt.Println(string(encoded))
	if !decision.Ready {
		os.Exit(2)
	}
}
