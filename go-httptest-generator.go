package generator

import (
	"go/ast"
	"strconv"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "go-httptest-generator is ..."

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "go-httptest-generator",
	Doc:  doc,
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := inspector.New(pass.Files)
	nodeFilter := []ast.Node{
		new(ast.CallExpr),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		if url, ok := httpHandleFunc(n); ok {
			pass.Reportf(n.Pos(), "http.HandleFunc with %s", url)
		}
	})

	return nil, nil
}

func httpHandleFunc(n ast.Node) (string, bool) {
	callExpr, ok := n.(*ast.CallExpr)
	if !ok {
		return "", false
	}

	selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok || selectorExpr.Sel.Name != "HandleFunc" {
		return "", false
	}

	ident, ok := selectorExpr.X.(*ast.Ident)
	if !ok || ident.Name != "http" {
		return "", false
	}

	lit, _ := callExpr.Args[0].(*ast.BasicLit)
	url, err := strconv.Unquote(lit.Value)
	if err != nil {
		// handle error
		return "", false
	}

	return url, true
}
