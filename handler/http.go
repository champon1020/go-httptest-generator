package handler

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// Search called function the name is `funcName`.
// If exists, return the following values:
//   - the url of argument 0
//   - the handler of argument 1
//   - file name in which http.HandleFunc is called
func searchFuncInNetHttp(pass *analysis.Pass, call *ast.CallExpr, funcName string) (ast.Expr, ast.Expr, string, bool) {
	var (
		arg0 ast.Expr
		arg1 ast.Expr
	)

	falseReturn := func() (ast.Expr, ast.Expr, string, bool) {
		return nil, nil, "", false
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return falseReturn()
	}
	obj, ok := pass.TypesInfo.Uses[selector.Sel]
	if !ok {
		return falseReturn()
	}

	switch obj := obj.(type) {
	case *types.Func:
		if obj.Pkg().Path() != "net/http" || obj.Name() != funcName {
			return falseReturn()
		}
	case *types.TypeName:
		if !types.Identical(obj.Type(), httpHandlerFuncObj.Type()) {
			return falseReturn()
		}
	}

	fn := pass.Fset.File(call.Lparen).Name()
	if len(call.Args) > 0 {
		arg0 = call.Args[0]
	}
	if len(call.Args) > 1 {
		arg1 = call.Args[1]
	}

	return arg0, arg1, fn, true
}
