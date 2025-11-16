package nplusone

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer is a go/analysis analyzer that detects potential SQL N+1 issues.
// It reports database/sql method calls (Query*/Exec*/Prepare*) that are invoked inside loops.
var Analyzer = &analysis.Analyzer{
	Name: "nplusone",
	Doc:  "detect potential SQL N+1 query patterns for database/sql calls inside loops",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Track loop nesting depth and detect database/sql calls made inside loops.
	nodeFilter := []ast.Node{
		(*ast.ForStmt)(nil),
		(*ast.RangeStmt)(nil),
		(*ast.CallExpr)(nil),
	}

	// Candidate method names (only those that belong to database/sql are considered)
	suspectNames := map[string]struct{}{
		"Query":           {},
		"QueryContext":    {},
		"QueryRow":        {},
		"QueryRowContext": {},
		"Exec":            {},
		"ExecContext":     {},
		"Prepare":         {},
		"PrepareContext":  {},
	}

	loopDepth := 0
	insp.Nodes(nodeFilter, func(n ast.Node, push bool) bool {
		switch n := n.(type) {
		case *ast.ForStmt, *ast.RangeStmt:
			if push {
				loopDepth++
			} else {
				loopDepth--
			}
			// Keep traversing inside the loop body
			return true
		case *ast.CallExpr:
			if !push {
				// Do nothing on exit phase
				return false
			}
			if loopDepth <= 0 {
				return false
			}
			if isDBSQLMethodCall(pass, n, suspectNames) {
				// Warn if a database/sql call is made inside a loop.
				// Point the diagnostic at the selector identifier (method name).
				if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
					method := sel.Sel.Name
					// Use a dedicated message for Prepare* methods.
					if method == "Prepare" || method == "PrepareContext" {
						pass.Reportf(sel.Sel.Pos(), "prepare inside loop; consider preparing once and reusing the statement")
					} else {
						pass.Reportf(sel.Sel.Pos(), "potential N+1: database/sql method %s called inside a loop", method)
					}
				} else {
					pass.Reportf(n.Lparen, "potential N+1: database/sql call inside a loop")
				}
			}
			// No need to analyze children of CallExpr
			return false
		default:
			return true
		}
	})

	return nil, nil
}

// isDBSQLMethodCall reports whether the given call is a database/sql method call
// and the method name is contained in the provided set.
func isDBSQLMethodCall(pass *analysis.Pass, call *ast.CallExpr, names map[string]struct{}) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// If Selections has an entry, it's a method call (x.M()).
	// If nil, it's most likely a package function (pkg.Func).
	selInfo := pass.TypesInfo.Selections[sel]
	if selInfo == nil {
		// Package functions (e.g., sql.Open) are excluded
		return false
	}

	// Determine by the package the method belongs to; only database/sql is targeted.
	if fn, ok := selInfo.Obj().(*types.Func); ok {
		if fn.Pkg() != nil && fn.Pkg().Path() == "database/sql" {
			if _, ok := names[fn.Name()]; ok {
				return true
			}
		}
	}

	// As a fallback, also check the receiver type's package (interfaces, wrappers, etc.).
	recv := selInfo.Recv()
	if recv != nil {
		if pkgPathOfType(recv) == "database/sql" {
			if _, ok := names[sel.Sel.Name]; ok {
				return true
			}
		}
	}

	return false
}

// pkgPathOfType returns the package path of the given type. It recursively dereferences
// pointers until reaching a non-pointer type, then extracts the package path if the type
// is a named type. Returns an empty string if the type has no associated package path.
func pkgPathOfType(t types.Type) string {
	for {
		switch u := t.(type) {
		case *types.Pointer:
			t = u.Elem()
			continue
		default:
		}
		break
	}
	if n, ok := t.(*types.Named); ok {
		if n.Obj() != nil && n.Obj().Pkg() != nil {
			return n.Obj().Pkg().Path()
		}
	}
	return ""
}
