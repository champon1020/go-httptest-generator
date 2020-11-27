package std

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// analyzeHTTPHandle analyze the function of http.Handle.
// In this function, parse the handler which is the second argument of http.Handle.
// Following examples are patterns of http.Handle usage:
//      - http.Handle("URL", http.HandlerFunc(anyIndex))
//      - http.Handle("URL", new(anyHandler))
//      - http.Handle("URL", anyHandlerFunc)
//      - http.Handle("URL", anyHandlerInstance)
// where anyIndex is the function whose type is func(w ResponseWriter, r *Request)
// and anyHandler is the handler struct
// and anyHandlerFunc is the handler variable which is created by http.HandlerFunc()
// and anyHandlerInstance is the constructed handler object.
func analyzeHTTPHandle(pass *analysis.Pass, h *Handler, args []ast.Expr) bool {
	if len(args) != 2 || !h.SetURLFromExpr(args[0]) {
		return false
	}

	switch arg1 := args[1].(type) {
	case *ast.CallExpr:
		// Parse handler function which is the argument of http.HandlerFunc().
		// Examples:
		//      http.Hanle("url", http.HandlerFunc(AnyIndex))  // OK
		//      http.Hanle("url", http.HandlerFunc(anyIndex))  // Ignore
		// where
		//      AnyIndex := func(w http.ResponseWriter, r *http.Request){}
		//      anyIndex := func(w http.ResponseWriter, r *http.Request){}
		if selExpr, ok := arg1.Fun.(*ast.SelectorExpr); ok {
			obj := pass.TypesInfo.ObjectOf(selExpr.Sel)
			if types.Identical(httpHandlerFuncObj.Type(), obj.Type()) {
				ident, ok := arg1.Args[0].(*ast.Ident)
				if !ok {
					return false
				}
				if parseHandlerFunc(pass, h, pass.TypesInfo.ObjectOf(ident)) {
					break
				}
			}
		}
		if ident, ok := arg1.Fun.(*ast.Ident); ok {
			// Parse handler object with new builtin.
			// Examples:
			//      http.Handle("url", new(AnyHandler)) // OK
			//      http.Handle("url", new(anyHandler)) // Ignore
			// where
			//      type AnyHandler struct {}
			//      type anyHandler struct {}
			obj := pass.TypesInfo.Uses[ident]
			if types.Identical(newObj.Type(), obj.Type()) &&
				parseHandlerWithNewBuiltin(pass, h, arg1.Args[0]) {
				ident, _ := arg1.Args[0].(*ast.Ident)
				h.Name = ident.Name
				h.TypeFlg |= (1 << NewBuiltinH)
				break
			}
		}
		return false
	case *ast.Ident:
		obj := pass.TypesInfo.Uses[arg1]
		// Parse handler which is created by http.HandlerFunc().
		// Examples:
		//      http.Handle("url", AnyHandlerFunc)  // OK
		//      http.Handle("url", anyHandlerFunc)  // Ignore
		//      http.Handle("url", AnyHandlerFunc2) // OK
		//      http.Handle("url", anyHandlerFunc2) // Ignore
		// where
		//      AnyHandlerFunc  := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request))
		//      anyHandlerFunc  := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request))
		//      AnyHandlerFunc2 := http.HandlerFunc(anyindex)
		//      anyHandlerFunc2 := http.HandlerFunc(AnyIndex)
		if types.Identical(obj.Type(), httpHandlerFuncObj.Type()) &&
			analyzeHTTPHandlerFunc(pass, h, obj) {
			break
		}

		// Parse constructed handler object.
		// Examples:
		//      http.Handle("url", AnyHandlerInstance)  // OK
		//      http.Handle("url", anyHandlerInstance)  // OK
		//      http.Handle("url", AnyHandlerInstance2) // OK
		//      http.Handle("url", anyHandlerInstance2) // Ignore
		// where
		//      AnyHandlerInstance  := AnyHandler
		//      anyHandlerInstance  := AnyHandler
		//      AnyHandlerInstance2 := anyHandler
		//      anyHandlerInstance2 := anyHandler
		if parseHandler(pass, h, obj) {
			break
		}
		return false
	}

	return true
}

// isHTTPHandle check whether the type of callexpr is http.Handle or not.
func isHTTPHandle(pass *analysis.Pass, expr *ast.CallExpr) ([]ast.Expr, string, bool) {
	return searchFuncInNetHTTP(pass, expr, "Handle")
}
