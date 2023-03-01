package utils_test

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"reflect"
	"strings"
	"testing"

	"github.com/github/git-bundle-server/cmd/utils"
	. "github.com/github/git-bundle-server/internal/testhelpers"
	typeutils "github.com/github/git-bundle-server/internal/utils"
	"github.com/stretchr/testify/assert"
)

func findAllGetDependencyTypesInDir(relativePathToDir string) (*token.FileSet, []ast.Expr) {
	fset := token.NewFileSet() // positions are relative to fset
	pkgs, err := parser.ParseDir(fset, relativePathToDir, nil, 0)
	if err != nil {
		panic("could not read directory")
	}

	typeNodes := []ast.Expr{}
	for _, pkg := range pkgs {
		ast.Inspect(pkg, func(n ast.Node) bool {
			switch x := n.(type) {
			// Might be an invocation of GetDependency
			case *ast.IndexExpr:
				fnSelector, ok := x.X.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				if fnSelector.Sel.Name == "GetDependency" {
					// Now, get the identifier (or selector) for the type
					typeNodes = append(typeNodes, x.Index)
				}

				return true
			}

			// Keep recursing!
			return true
		})
	}

	return fset, typeNodes
}

// Test that all GetDependency invocations in the 'git-bundle-server' 'main'
// package are in the container built by 'BuildGitBundleServerContainer'.
//
// This test is somewhat fragile, and it isn't comprehensive. Its main utility
// is to cover easy-to-miss runtime issues that could arise from the dependency
// provider.
//
// Scenarios that can cause issues:
//   - 'utils' (not to be confused with 'internal/utils') is imported with a
//     dot-import in the tested file (don't do this!).
//   - types requested by 'GetDependency()' belong to an explicitly aliased or
//     dot-imported package.
//   - invocations of the dependency container in files outside the tested
//     packages (also don't do this!).
func TestDependencyContainer(t *testing.T) {
	logger := &MockTraceLogger{}
	ctx := context.Background()

	t.Run("Container is created successfully", func(t *testing.T) {
		assert.NotPanics(t, func() { utils.BuildGitBundleServerContainer(logger) })
	})

	t.Run("Verify container is internally consistent", func(t *testing.T) {
		container := utils.BuildGitBundleServerContainer(logger)
		assert.NotPanics(t, func() { container.InvokeAll(ctx) })
	})

	t.Run("Verify all external invocations are registered", func(t *testing.T) {
		container := utils.BuildGitBundleServerContainer(logger)
		registeredTypes := typeutils.Map(container.ListRegisteredTypes(),
			func(t reflect.Type) string {
				return t.String()
			},
		)

		fset, typeNodes := findAllGetDependencyTypesInDir("../git-bundle-server")

		// We expect at least one registered dependency, otherwise get rid of
		// this test.
		assert.NotEmpty(t, typeNodes)

		// Ensure each node is found in the container
		for _, node := range typeNodes {
			var name string
			if ident, ok := node.(*ast.Ident); ok {
				// No package identified with the type -
				pkgPath := reflect.TypeOf(*container).PkgPath()
				pkgComponents := strings.Split(pkgPath, "/")
				name = fmt.Sprintf("%s.%s", pkgComponents[len(pkgComponents)-1], ident.Name)
			} else {
				var nameBuf bytes.Buffer
				err := printer.Fprint(&nameBuf, fset, node)
				if err != nil {
					assert.Fail(t, err.Error())
				}
				name = nameBuf.String()
			}

			// check if name is in registered types list
			callLocation := fset.Position(node.Pos())
			assert.Contains(t, registeredTypes, name,
				"Type %s was not registered; see call to 'GetDependency()' in file '%s', line %d",
				name, callLocation.Filename, callLocation.Line)
		}
	})
}
