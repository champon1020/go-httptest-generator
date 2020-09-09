package handler

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// Analyze when using http.Handle.
func analyzeHttpHandle(ctx *Context, arg0 ast.Expr, arg1 ast.Expr) bool {
	pass := ctx.pass

	if !ctx.SetURLFromExpr(arg0) {
		return false
	}

	switch arg1 := arg1.(type) {
	case *ast.CallExpr:
		// Examples:
		// http.Hanle("url", http.HandlerFunc(Index))  // OK
		// http.Hanle("url", http.HandlerFunc(index))  // Ignore
		if selector, ok := arg1.Fun.(*ast.SelectorExpr); ok {
			obj := pass.TypesInfo.ObjectOf(selector.Sel)
			if types.Identical(httpHandlerFuncObj.Type(), obj.Type()) {
				ident, ok := arg1.Args[0].(*ast.Ident)
				if !ok {
					return false
				}
				if parseHandlerFunc(ctx, pass.TypesInfo.ObjectOf(ident)) {
					break
				}
			}
		}
		if ident, ok := arg1.Fun.(*ast.Ident); ok {
			// Examples:
			// http.Handle("url", new(AnyHandler)) // OK
			// http.Handle("url", new(anyHandler)) // Ignore
			obj := pass.TypesInfo.Uses[ident]
			if types.Identical(newObj.Type(), obj.Type()) &&
				parseHandlerWithNew(ctx, arg1.Args[0]) {
				hIdent, _ := arg1.Args[0].(*ast.Ident)
				ctx.Name = hIdent.Name
				ctx.IsNew = true
				break
			}
		}
		return false
	case *ast.Ident:
		obj := pass.TypesInfo.Uses[arg1]
		// Examples:
		// http.Handle("url", H) // OK
		// http.Handle("url", h) // Ignore
		// http.Handle("url", H2) // OK
		// http.Handle("url", h2) // Ignore
		if types.Identical(obj.Type(), httpHandlerFuncObj.Type()) &&
			parseHttpHandlerFunc(ctx, obj) {
			break
		}

		// Examples:
		// http.Handle("url", A)  // OK
		// http.Handle("url", AA) // Ignore
		// http.Handle("url", a)  // OK
		// http.Handle("url", aa) // Ignore
		if parseHandler(ctx, obj) {
			break
		}
		return false
	}

	return true
}

// The CallExpr is whether `http.Handle` or not.
func isHttpHandle(pass *analysis.Pass, call *ast.CallExpr) (ast.Expr, ast.Expr, string, bool) {
	return searchFuncInNetHttp(pass, call, "Handle")
}
