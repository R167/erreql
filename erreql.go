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
var DefaultAnalyzer = Build(Config{})

func Build(config Config) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "erreql",
		Doc:      "Check for usages of error ==/!= to non-nil values and suggest errors.Is instead.",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runner(config.compileConfig()),
	}
}

func runner(c config) func(pass *analysis.Pass) (interface{}, error) {

	return func(pass *analysis.Pass) (interface{}, error) {
		if c.skipPackage(pass.Pkg.Path()) {
			return nil, nil
		}

		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

		nodeFilter := []ast.Node{
			(*ast.BinaryExpr)(nil),
		}
		if !c.SkipSwitches {
			nodeFilter = append(nodeFilter, (*ast.SwitchStmt)(nil))
		}
		inspect.Preorder(nodeFilter, func(n ast.Node) {
			// libraries that use a sentinel error value to indicate success.
			switch n := n.(type) {
			case *ast.BinaryExpr:
				checkBin(pass, n)
			case *ast.SwitchStmt:
				checkSwitch(pass, n)
			}

		})

		return nil, nil
	}
}

func checkSwitch(pass *analysis.Pass, sw *ast.SwitchStmt) {
	if sw.Tag == nil || !isErr(pass, sw.Tag) {
		return
	}
	for _, cc := range sw.Body.List {
		cc := cc.(*ast.CaseClause)
		if len(cc.List) == 0 {
			continue
		}
		for _, cond := range cc.List {
			if isNil(pass, cond) {
				continue
			}
			if !isErr(pass, cond) {
				continue
			}
			if isLiteralNonError(pass, cond) {
				continue
			}
			pass.Reportf(cond.Pos(), "switch does not handle wrapped errors")
		}
	}
}

func checkBin(pass *analysis.Pass, bin *ast.BinaryExpr) {
	if bin.Op != token.EQL && bin.Op != token.NEQ {
		return
	}

	if isNil(pass, bin.X) || isNil(pass, bin.Y) {
		// checking against untyped nil is fine
		return
	}

	if !isErr(pass, bin.X) || !isErr(pass, bin.Y) {
		// checking against non-error types is fine
		return
	}

	if isLiteralNonError(pass, bin.X) || isLiteralNonError(pass, bin.Y) {
		// Literal non-Error types are fine. This is a common pattern for
		return
	}

	pass.Reportf(bin.Pos(), "use errors.Is or errors.As instead of %s", bin.Op)
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
