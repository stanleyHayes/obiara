package application

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// TestCounselDependencyBoundary is an enforceable source-level architecture
// gate. Counsel production code cannot import matching, trust, or AI; those
// packages cannot import counsel. Safety leaves only through the local port.
func TestCounselDependencyBoundary(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate dependency test")
	}
	internalRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
	forbiddenConsumers := []string{
		filepath.Join(internalRoot, "matching"),
		filepath.Join(internalRoot, "trust"),
		filepath.Join(internalRoot, "ai"),
	}
	for _, root := range forbiddenConsumers {
		assertNoImport(t, root, "/services/api/internal/counsel/")
	}
	counselRoot := filepath.Join(internalRoot, "counsel")
	for _, forbidden := range []string{
		"/services/api/internal/matching/",
		"/services/api/internal/trust/",
		"/services/api/internal/ai/",
	} {
		assertNoImport(t, counselRoot, forbidden)
	}
}

func assertNoImport(t *testing.T, root, forbidden string) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return parseErr
		}
		for _, imported := range file.Imports {
			value, unquoteErr := strconv.Unquote(imported.Path.Value)
			if unquoteErr != nil {
				return unquoteErr
			}
			if strings.Contains(value, forbidden) {
				t.Errorf("%s imports forbidden boundary %s", path, value)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
