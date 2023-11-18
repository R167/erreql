package erreql

import (
	"go/ast"
	"go/token"
	"go/types"
	"regexp"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer is the main entry point for the analyzer.
var Analyzer = &analysis.Analyzer{
	Name:     "erreql",
	Doc:      "Check for usages of `err ==` and `err !=` to non-nil values and suggest [errors.Is] or [errors.As] instead.",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	switch pass.Pkg.Path() {
	case "errors", "errors_test":
		// These packages know how to use their own APIs.
		// Sometimes they are testing what happens to incorrect programs.
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.BinaryExpr)(nil),
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		bin := n.(*ast.BinaryExpr)
		if bin.Op != token.EQL && bin.Op != token.NEQ {
			return
		}
		// checking against untyped nil is fine
		if isNil(pass, bin.X) || isNil(pass, bin.Y) {
			return
		}
		// checking against non-error types is fine
		if !isErr(pass, bin.X) || !isErr(pass, bin.Y) {
			return
		}
		// Literal non-Error types are fine. This is a common pattern for
		// libraries that use a sentinel error value to indicate success.
		if isLiteralNonError(pass, bin.X) || isLiteralNonError(pass, bin.Y) {
			return
		}

		pass.Reportf(bin.Pos(), "use errors.Is or errors.As instead of %s", bin.Op)
	})

	return nil, nil
}

var errName = regexp.MustCompile(`^err.|Err|Error|Exception`)

// Check if the expression is a sentinel error value like io.EOF.
// A value is considered a sentinel error value when it's defined as a variable
// of type error and its name matches none of the following patterns:
//   - err.+
//   - .*Err.*
//   - .*Exception.*
func isLiteralNonError(pass *analysis.Pass, expr ast.Expr) bool {
	var id *ast.Ident
	switch t := expr.(type) {
	case *ast.Ident:
		id = t
	case *ast.SelectorExpr:
		id = t.Sel
	default:
		return false
	}
	if id.Name == "err" {
		return false
	}

	return !errName.MatchString(id.Name)
}

// The error interface type.
var errType = types.Universe.Lookup("error").Type().Underlying().(*types.Interface)

// Check if the expression's type implements the error interface.
func isErr(pass *analysis.Pass, e ast.Expr) bool {
	typ := pass.TypesInfo.TypeOf(e)
	if typ == nil {
		return false
	}
	if it, ok := typ.Underlying().(*types.Interface); ok && it.NumMethods() == 0 {
		// skip interface{} since it could be anything
		return false
	}

	return types.Implements(typ, errType)
}

// Check if the expression is nil.
func isNil(pass *analysis.Pass, e ast.Expr) bool {
	typ := pass.TypesInfo.TypeOf(e)
	if typ == nil {
		return false
	}
	return typ == types.Typ[types.UntypedNil]
}
