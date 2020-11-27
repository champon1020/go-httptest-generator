package std

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// searchFuncInNetHTTP search the function at net.http package whose name is `funcName`.
func searchFuncInNetHTTP(pass *analysis.Pass, callExpr *ast.CallExpr, funcName string) ([]ast.Expr, string, bool) {
	var args []ast.Expr

	returnFalse := func() ([]ast.Expr, string, bool) {
		return args, "", false
	}

	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return returnFalse()
	}

	obj, ok := pass.TypesInfo.Uses[selExpr.Sel]
	if !ok {
		return returnFalse()
	}

	switch obj := obj.(type) {
	case *types.Func:
		if obj.Pkg().Path() != "net/http" || obj.Name() != funcName {
			return returnFalse()
		}
	case *types.TypeName:
		if !types.Identical(obj.Type(), httpHandlerFuncObj.Type()) {
			return returnFalse()
		}
	}

	// Get file name which has searching function.
	fn := pass.Fset.File(callExpr.Lparen).Name()

	// Get arguments of searching function.
	for _, arg := range callExpr.Args {
		args = append(args, arg)
	}

	return args, fn, true
}
