package openapi

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// TestEveryServedRouteIsInTheContract compares what the service registers
// against what the document describes, in both directions.
//
// The operation-id count guard next to this cannot catch either failure. It
// counts ids inside the document, so a route served and never written down is
// invisible to it, and so is a path written down that nothing serves. Three
// seed-stage routes had been live and undocumented for exactly that reason:
// the guard was doing its job and its job was the wrong shape.
//
// A generated client is built from this document. A route missing from it is
// unreachable to every client that trusts it; a path in it that nothing serves
// gives clients a method that 404s.
func TestEveryServedRouteIsInTheContract(t *testing.T) {
	served, err := registeredRoutes()
	if err != nil {
		t.Fatalf("read routes: %v", err)
	}
	documented, err := documentedRoutes()
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}
	if len(served) == 0 || len(documented) == 0 {
		t.Fatalf("extraction found nothing: %d served, %d documented", len(served), len(documented))
	}

	for _, route := range difference(served, documented) {
		t.Errorf("served but absent from the contract: %s", route)
	}
	for _, route := range difference(documented, served) {
		t.Errorf("in the contract but served by nothing: %s", route)
	}
}

var handlePattern = regexp.MustCompile(
	`mux\.Handle(?:Func)?\("(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS) (/[^"]*)"`)

// registeredRoutes reads the registrations out of the source.
//
// Go's ServeMux does not expose the patterns it holds, so there is nothing to
// ask at runtime; reading the calls is the only way to know what is served.
func registeredRoutes() (map[string]struct{}, error) {
	routes := map[string]struct{}{}
	roots := []string{
		filepath.Join("..", "internal", "platform", "http"),
		filepath.Join(".."),
	}
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() || !strings.HasSuffix(name, ".go") ||
				strings.HasSuffix(name, "_test.go") {
				continue
			}
			source, err := os.ReadFile(filepath.Join(root, name))
			if err != nil {
				return nil, err
			}
			for _, match := range handlePattern.FindAllStringSubmatch(string(source), -1) {
				routes[match[1]+" "+match[2]] = struct{}{}
			}
		}
	}
	return routes, nil
}

var (
	pathPattern   = regexp.MustCompile(`^  (/\S*):\s*$`)
	methodPattern = regexp.MustCompile(`^    (get|post|put|patch|delete|head|options):\s*$`)
)

// documentedRoutes walks the paths block line by line.
//
// Matching a path and then scanning to the next one would let a path that is a
// prefix of another claim its methods — /v1/blocks would swallow the delete on
// /v1/blocks/{blockerId}/{blockedId}. Tracking the current path as the lines go
// past cannot make that mistake.
func documentedRoutes() (map[string]struct{}, error) {
	document, err := os.ReadFile(filepath.Join("openapi.yaml"))
	if err != nil {
		return nil, err
	}
	routes := map[string]struct{}{}
	inPaths, current := false, ""
	for _, line := range strings.Split(string(document), "\n") {
		if strings.TrimRight(line, " ") == "paths:" {
			inPaths = true
			continue
		}
		if !inPaths {
			continue
		}
		// A top-level key ends the paths block.
		if line != "" && !strings.HasPrefix(line, " ") {
			break
		}
		if match := pathPattern.FindStringSubmatch(line); match != nil {
			current = match[1]
			continue
		}
		if match := methodPattern.FindStringSubmatch(line); match != nil && current != "" {
			routes[strings.ToUpper(match[1])+" "+current] = struct{}{}
		}
	}
	return routes, nil
}

func difference(from, minus map[string]struct{}) []string {
	var only []string
	for route := range from {
		if _, found := minus[route]; !found {
			only = append(only, route)
		}
	}
	sort.Strings(only)
	return only
}
