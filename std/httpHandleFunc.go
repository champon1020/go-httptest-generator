package std

import (
	"errors"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// analyzeHTTPHandleFunc analyze the funciton of http.HandleFunc.
// In this function, parse the handler function which is the second argument of http.HandleFunc.
// Following examples are patterns of http.HandleFunc usage:
//      - http.HandleFunc("URL", anyIndex)
//      - http.HandleFunc("URL", func(w http.ResponseWriter, r *http.Request){ ... })
// where anyIndex is the handler function whose type is func(w http.ResponseWriter, r *http.Request).
func analyzeHTTPHandleFunc(pass *analysis.Pass, h *Handler, args []ast.Expr) error {
	if len(args) != 2 || !h.SetURLFromExpr(args[0]) {
		return errors.New("arguments of http.HandleFunc is not valid")
	}

	// Examples:
	//      http.HandleFunc("url", HFunc)       // OK
	//      http.HandleFunc("url", hFunc)       // Ignore
	//      http.HandleFunc("url", HFuncLit)    // OK
	//      http.HandleFunc("url", hFuncLit)    // Ignore
	//      http.HandleFunc("url", HFuncLit2)   // OK
	//      http.HandleFunc("url", hFuncLit2)   // Ignore
	// where
	//      func HFunc(w http.ResponseWriter, r *http.Request){}
	//      func hFunc(w http.ResponseWriter, r *http.Request){}
	//      HFuncLit  := HFunc
	//      hFuncLit  := HFunc
	//      HFuncLit2 := hFunc
	//      hFuncLit2 := hFunc
	ident, ok := args[1].(*ast.Ident)
	if !ok {
		return errors.New("argument 1 of http.HandleFunc is not valid")
	}

	obj := pass.TypesInfo.ObjectOf(ident)
	if err := parseHandlerFunc(pass, h, obj); err != nil {
		return err
	}

	return nil
}

// isHTTPHandleFunc check whether the callexpr is http.HandleFunc or not.
func isHTTPHandleFunc(pass *analysis.Pass, expr *ast.CallExpr) ([]ast.Expr, string, bool) {
	return searchFuncInNetHTTP(pass, expr, "HandleFunc")
}
