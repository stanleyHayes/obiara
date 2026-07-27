package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/stanleyHayes/obiara/internal/quality/providerdiligence"
)

func main() {
	var path, at string
	flag.StringVar(&path, "registry", "", "provider diligence registry JSON")
	flag.StringVar(&at, "at", "", "evaluation instant (RFC3339)")
	flag.Parse()
	now := time.Now().UTC()
	if at != "" {
		var err error
		now, err = time.Parse(time.RFC3339, at)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid --at")
			os.Exit(64)
		}
	}
	_, decision, err := providerdiligence.Load(path, now)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	_ = json.NewEncoder(os.Stdout).Encode(decision)
	if !decision.Ready {
		os.Exit(2)
	}
}
