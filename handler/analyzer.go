package handler

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
)

var newObj = types.Universe.Lookup("new")

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

	handlersInfo := []*HandlerInfo{}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		handlerInfo := NewHandlerInfo(pass.Pkg.Name(), pass.Pkg.Path())
		call, _ := n.(*ast.CallExpr)

		// http.Handle
		if arg0, arg1, fn, ok := isHttpHandle(pass, call); ok {
			handlerInfo.File = fn
			if analyzeHttpHandle(pass, handlerInfo, arg0, arg1) {
				handlersInfo = append(handlersInfo, handlerInfo)
				pass.Reportf(n.Pos(), "Handle %s %s %s", handlerInfo.URL, handlerInfo.Method, handlerInfo.Name)
			}
			return
		}

		// http.HandleFunc
		if arg0, arg1, fn, ok := isHttpHandleFunc(pass, call); ok {
			handlerInfo.File = fn
			if analyzeHttpHandleFunc(pass, handlerInfo, arg0, arg1) {
				handlersInfo = append(handlersInfo, handlerInfo)
				pass.Reportf(n.Pos(), "HandleFunc %s %s %s", handlerInfo.URL, handlerInfo.Method, handlerInfo.Name)
			}
			return
		}
	})

	/*
		// If this is not test, generate test files.
		if flag.Lookup("test.v") == nil {
			GenerateAllTests(handlersInfo)
		}
	*/

	return nil, nil
}
