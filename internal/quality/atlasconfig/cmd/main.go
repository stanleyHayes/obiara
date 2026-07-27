package main

import (
	"fmt"
	"github.com/stanleyHayes/obiara/internal/quality/atlasconfig"
	"os"
	"path/filepath"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	for _, name := range []string{"staging.yaml", "production.yaml"} {
		path := filepath.Join(root, "deploy", "atlas", name)
		if err := atlasconfig.ValidateFile(root, path); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
			os.Exit(1)
		}
	}
	fmt.Println("Atlas staging and production configuration valid")
}
