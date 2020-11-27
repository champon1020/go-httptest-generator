package std

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// analyzeHTTPHandleFunc analyze the funciton of http.HandleFunc.
// In this function, parse the handler function which is the second argument of http.HandleFunc.
// Following examples are patterns of http.HandleFunc usage:
//      - http.HandleFunc("URL", anyIndex)
//      - http.HandleFunc("URL", func(w http.ResponseWriter, r *http.Request){ ... })
// where anyIndex is the handler function whose type is func(w http.ResponseWriter, r *http.Request).
func analyzeHTTPHandleFunc(pass *analysis.Pass, h *Handler, args []ast.Expr) bool {
	if len(args) != 2 || !h.SetURLFromExpr(args[0]) {
		return false
	}

	// Examples:
	//      http.HandleFunc("url", Index)       // OK
	//      http.HandleFunc("url", index)       // Ignore
	//      http.HandleFunc("url", IndexVar)    // OK
	//      http.HandleFunc("url", IndexVar2)   // OK
	//      http.HandleFunc("url", indexVar)    // OK
	//      http.HandleFunc("url", indexVar2)   // Ignore
	//      http.HandleFunc("url", IndexVar3)   // OK
	//      http.HandleFunc("url", IndexVar4)   // Ignore
	// where
	//      func Index(w http.ResponseWriter, r *http.Request) { ... }
	//      func index(w http.ResponseWriter, r *http.Request) { ... }
	//      IndexVar  := Index
	//      IndexVar2 := index
	//      indexVar  := Index
	//      indexVar2 := index
	//      {
	//          IndexVar3 := Index
	//          IndexVar4 := index
	//      }
	ident, ok := args[1].(*ast.Ident)
	if !ok {
		return false
	}

	obj := pass.TypesInfo.ObjectOf(ident)
	if !parseHandlerFunc(pass, h, obj) {
		return false
	}

	return true
}

// isHTTPHandleFunc check whether the callexpr is http.HandleFunc or not.
func isHTTPHandleFunc(pass *analysis.Pass, expr *ast.CallExpr) ([]ast.Expr, string, bool) {
	return searchFuncInNetHTTP(pass, expr, "HandleFunc")
}
