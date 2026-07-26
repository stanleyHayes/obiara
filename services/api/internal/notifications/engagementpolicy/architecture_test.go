package engagementpolicy

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPurePolicyArchitecture(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	forbiddenImports := []string{
		"mongo", "database/sql", "net/http", "kafka", "resend", "openai",
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" ||
			strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		source, err := os.ReadFile(entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		file, err := parser.ParseFile(token.NewFileSet(), entry.Name(), source, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, imported := range file.Imports {
			path := strings.Trim(imported.Path.Value, `"`)
			for _, forbidden := range forbiddenImports {
				if strings.Contains(path, forbidden) {
					t.Fatalf("pure policy imports %q", path)
				}
			}
		}
		for _, declaration := range file.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.TYPE {
				continue
			}
			for _, spec := range general.Specs {
				name := spec.(*ast.TypeSpec).Name.Name
				for _, forbidden := range []string{
					"Dispatcher", "Repository", "Member", "Score",
					"Generator", "DeliveryCapBypass",
				} {
					if strings.Contains(name, forbidden) {
						t.Fatalf("forbidden architecture type %q", name)
					}
				}
			}
		}
	}
}
