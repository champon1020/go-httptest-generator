package std

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "handlerAnalyzer analyzes handlers and get handler information."

// Analyzer analyzes handlers and get handler information.
var Analyzer = &analysis.Analyzer{
	Name:     "hanlderAnalyzer",
	Doc:      doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

var (
	httpHandlerObj     types.Object
	httpHandlerFuncObj types.Object
	httpRequestObj     types.Object
	newObj             = types.Universe.Lookup("new")
)

func initObj(imports []*types.Package) bool {
	var flg bool
	for _, p := range imports {
		if p.Path() == "net/http" {
			httpHandlerObj = p.Scope().Lookup("Handler")
			httpHandlerFuncObj = p.Scope().Lookup("HandlerFunc")
			httpRequestObj = p.Scope().Lookup("Request")
			flg = true
			break
		}
	}
	return flg
}

func run(pass *analysis.Pass) (interface{}, error) {
	if ok := initObj(pass.Pkg.Imports()); !ok {
		/* handle error */
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	hs := []*Handler{}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		h := NewHandler(pass)
		cExpr, _ := n.(*ast.CallExpr)

		// Check if the expr is http.Handle and if true, analyze it.
		if args, fn, ok := isHTTPHandle(pass, cExpr); ok {
			h.File = fn
			if analyzeHTTPHandle(pass, h, args) {
				hs = append(hs, h)
				pass.Reportf(n.Pos(), "Handle %s %s %s", h.URL, h.Method, h.Name)
			}
			return
		}

		// Check if the expr is http.HandleFunc and if true, analyze it.
		if args, fn, ok := isHTTPHandleFunc(pass, cExpr); ok {
			h.File = fn
			if analyzeHTTPHandleFunc(pass, h, args) {
				hs = append(hs, h)
				pass.Reportf(n.Pos(), "HandleFunc %s %s %s", h.URL, h.Method, h.Name)
			}
			return
		}
	})

	/*
		// If this is not test, generate test files.
		if flag.Lookup("test.v") == nil {
			GenerateAllTests(contexs)
		}
	*/

	return nil, nil
}
