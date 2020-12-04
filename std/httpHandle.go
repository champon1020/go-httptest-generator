package std

import (
	"errors"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// analyzeHTTPHandle analyze the function of http.Handle.
// In this function, parse the handler which is the second argument of http.Handle.
// Following examples are patterns of http.Handle usage:
//      - http.Handle("URL", http.HandlerFunc(HFunc))
//      - http.Handle("URL", new(AnyHandler))
//      - http.Handle("URL", H)
// where HFunc is the function whose type is func(w ResponseWriter, r *Request)
// and AnyHandler is the handler struct
// and H is the constructed handler object.
func analyzeHTTPHandle(pass *analysis.Pass, h *Handler, args []ast.Expr) error {
	if len(args) != 2 || !h.SetURLFromExpr(args[0]) {
		return errors.New("arguments of http.Handle is not valid")
	}

	switch arg1 := args[1].(type) {
	case *ast.CallExpr:
		// Parse handler function which is the argument of http.HandlerFunc().
		// Examples:
		//      http.Handle("url", http.HandlerFunc(HFunc))  // OK
		//      http.Handle("url", http.HandlerFunc(hFunc))  // Ignore
		// where
		//      func HFunc(w http.ResponseWriter, r *http.Request){}
		//      func hFunc(w http.ResponseWriter, r *http.Request){}
		if selExpr, ok := arg1.Fun.(*ast.SelectorExpr); ok {
			obj := pass.TypesInfo.ObjectOf(selExpr.Sel)
			if types.Identical(httpHandlerFuncObj.Type(), obj.Type()) {
				ident, ok := arg1.Args[0].(*ast.Ident)
				if !ok {
					return errors.New("argument of http.HandlerFunc is not valid")
				}

				if parseHandlerFunc(pass, h, pass.TypesInfo.ObjectOf(ident)) {
					break
				}
			}
		}

		// Parse handler object with new builtin.
		// Examples:
		//      http.Handle("url", new(AnyHandler)) // OK
		//      http.Handle("url", new(anyHandler)) // Ignore
		// where
		//      type AnyHandler struct{}
		//      type anyHandler struct{}
		if ident, ok := arg1.Fun.(*ast.Ident); ok {
			obj := pass.TypesInfo.Uses[ident]
			if types.Identical(newObj.Type(), obj.Type()) &&
				parseHandlerWithNewBuiltin(pass, h, arg1.Args[0]) {
				ident, _ := arg1.Args[0].(*ast.Ident)
				h.Name = ident.Name
				h.TypeFlg |= (1 << NewBuiltinH)
				break
			}
		}

		return errors.New("not found")
	case *ast.Ident:
		obj := pass.TypesInfo.Uses[arg1]

		// Parse handler which is created by http.HandlerFunc().
		// Examples:
		//      http.Handle("url", HFuncLit)  // OK
		//      http.Handle("url", hFuncLit)  // Ignore
		// where
		//      HFuncLit  := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request))
		//      hFuncLit  := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request))
		if types.Identical(obj.Type(), httpHandlerFuncObj.Type()) &&
			analyzeHTTPHandlerFunc(pass, h, obj) {
			break
		}

		// Parse constructed handler object.
		// Examples:
		//      http.Handle("url", H)  // OK
		//      http.Handle("url", h)  // Ignore
		//      http.Handle("url", H2) // OK
		//      http.Handle("url", h2) // Ignore
		// where
		//      H  := new(AnyHandler)
		//      h  := new(AnyHandler)
		//      H2 := http.HandlerFunc(HFunc)
		//      h2 := http.HandlerFunc(HFunc)
		if parseHandler(pass, h, obj) {
			break
		}

		return errors.New("not found")
	}

	return nil
}

// isHTTPHandle check whether the type of callexpr is http.Handle or not.
func isHTTPHandle(pass *analysis.Pass, expr *ast.CallExpr) ([]ast.Expr, string, bool) {
	return searchFuncInNetHTTP(pass, expr, "Handle")
}
