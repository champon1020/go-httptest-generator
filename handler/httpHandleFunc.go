package handler

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyze when using http.HandleFunc.
func analyzeHttpHandleFunc(pass *analysis.Pass, handlerInfo *HandlerInfo, arg0 ast.Expr, arg1 ast.Expr) bool {
	if !handlerInfo.SetURLFromExpr(arg0) {
		return false
	}

	// Examples:
	// http.HandleFunc("url", Index) // OK
	// http.HandleFunc("url", index) // Ignore
	// http.HandleFunc("url", IndexVar) // OK
	// http.HandleFunc("url", IndexVar2) // OK
	// http.HandleFunc("url", IndexVar3) // Ignore
	// http.HandleFunc("url", IndexVar4) // OK
	// http.HandleFunc("url", IndexVar5) // Ignore
	ident, ok := arg1.(*ast.Ident)
	if !ok {
		return false
	}
	obj := pass.TypesInfo.ObjectOf(ident)
	if !parseHandlerFunc(obj, handlerInfo, pass) {
		return false
	}

	return true
}

// The CallExpr is whether `http.HandleFunc` or not.
func isHttpHandleFunc(pass *analysis.Pass, call *ast.CallExpr) (ast.Expr, ast.Expr, string, bool) {
	return searchFuncInNetHttp(pass, call, "HandleFunc")
}
