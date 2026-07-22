// Package architecture holds executable checks for the layering rules the
// README describes. It contains no production code.
package architecture

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/husari/hube"

// layerRules lists, per layer, the layers it must never import. The direction
// of dependency is domain <- application <- infrastructure: inner layers must
// not know about outer ones.
var layerRules = map[string][]string{
	"internal/domain":      {"internal/application", "internal/infrastructure"},
	"internal/application": {"internal/infrastructure"},
}

// TestLayerDependencies walks the source tree and fails when a package imports
// a layer it is not allowed to depend on.
//
// The usual way this breaks is convenience: a service takes a concrete
// *sqlite.Repo or *external.Client because it is quicker than declaring a
// port. The cost shows up later as a package that cannot be unit tested
// without a database or a network call.
func TestLayerDependencies(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("resolve module root: %v", err)
	}

	var checked int
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		forbidden, ok := ruleFor(rel)
		if !ok {
			return nil
		}
		checked++

		imports, err := importsOf(path)
		if err != nil {
			return err
		}
		for _, imp := range imports {
			for _, bad := range forbidden {
				if strings.HasPrefix(imp, modulePath+"/"+bad) {
					t.Errorf("%s imports %s\n\tthis layer must not depend on %s; declare a port in the consuming package and let the concrete type satisfy it", rel, imp, bad)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	// Guard against the check silently passing because the paths moved.
	if checked == 0 {
		t.Fatal("no files were checked — the layer paths in layerRules are probably stale")
	}
	t.Logf("checked %d files across %d layers", checked, len(layerRules))
}

// ruleFor returns the forbidden prefixes for the layer that owns rel.
func ruleFor(rel string) ([]string, bool) {
	for layer, forbidden := range layerRules {
		if strings.HasPrefix(rel, layer+"/") {
			return forbidden, true
		}
	}
	return nil, false
}

// importsOf returns the import paths of a single Go file.
func importsOf(path string) ([]string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(f.Imports))
	for _, spec := range f.Imports {
		p, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}
